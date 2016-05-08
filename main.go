/*
Copyright 2016 Google Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

     http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Command template is a trivial web server that uses the text/template (and
// html/template) package's "block" feature to implement a kind of template
// inheritance.
//
// It should be executed from the directory in which the source resides,
// as it will look for its template files in the current directory.
package main

import (
	"bufio"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"github.com/julienschmidt/httprouter"
)

func main() {
	songs_root := "./songs_master" // 1st argument is the directory location
	filepath.Walk(songs_root, walkpath)

	fmt.Printf("%d Songs loaded.\n", len(loadedSongs))

	r := httprouter.New()

	r.GET("/", indexHandler)
	r.GET("/song/:song", songHandler)
	r.GET("/pdf/:song", pdfHandler)
	r.GET("/book/:book", bookIndexHandler)
	r.GET("/book/:book/:number", bookHandler)
	r.ServeFiles("/css/*filepath", http.Dir("css"))
	r.ServeFiles("/js/*filepath", http.Dir("js"))
	log.Fatal(http.ListenAndServe("localhost:8090", r))

}

// indexTemplate is the main site template.
// The default template includes two template blocks ("sidebar" and "content")
// that may be replaced in templates derived from this one.
//var indexTemplate = template.Must(template.ParseFiles("templates/index.tmpl"))

type DisplaySong struct {
	Link  string
	Title string
}

func (song DisplaySong) MatchTitle(title string) bool {
	return song.Title == title
}

type SongPage struct {
	Title        string
	Songs        []DisplaySong
	Song         Song
	HasSong      bool
	SelectedSong string
}

type BookPage struct {
	Title        string
	Songs        []DisplaySong
	Songbook     Songbook
	HasSong      bool
	SelectedSong string
	BookLink     string
}

var loadedSongs = make([]DisplaySong, 0)

// indexHandler is an HTTP handler that serves the index page.
func indexHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	t, err := template.ParseGlob("templates/*.tmpl")
	if err != nil {
		panic(err)
	}

	data := &SongPage{
		Title:   "Indigo Song Book",
		HasSong: false,
		Songs:   loadedSongs}

	if err := t.ExecuteTemplate(w, "index.tmpl", data); err != nil {
		log.Println(err)
	}
}

// imageTemplate is a clone of indexTemplate that provides
// alternate "sidebar" and "content" templates.
//var songTemplate = template.Must(template.Must(indexTemplate.Clone()).ParseFiles("templates/song.tmpl"))

func loadSongFile(title string, transpose int) (*Song, error) {
	filename := "songs_master/" + title + ".song"

	song, err := ParseSongFile(filename, transpose)

	if err != nil {
		return nil, err
	}

	return song, nil
}

func bookIndexHandler(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	sbook, err := ParseSongbookFile("./songs_master/Songbook/" + p.ByName("book") + ".songlist")

	if err != nil {
		fmt.Println(err)
	}

	page_data := &BookPage{
		Title:    "Indigo Song Book",
		HasSong:  false,
		Songbook: *sbook,
		Songs:    loadedSongs,
		BookLink: sbook.Filename[0 : len(sbook.Filename)-len(".songlist")]}

	temp, err := template.ParseFiles("templates/index.tmpl", "templates/book.tmpl")
	if err != nil {
		panic(err)
	}

	if err := temp.ExecuteTemplate(w, "book.tmpl", page_data); err != nil {
		log.Println(err)
	}
}

func bookHandler(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	num := p.ByName("number")
	n, _ := strconv.Atoi(num)

	var transpose = r.FormValue("transpose")
	if len(transpose) == 0 {
		transpose = "0"
	}
	t, err := strconv.Atoi(transpose)

	sbook, err := ParseSongbookFile("./songs_master/Songbook/" + p.ByName("book") + ".songlist")

	if err != nil {
		fmt.Println(err)
	}

	song := sbook.Songs[n-1]
	song.Transpose = t

	page_data := &SongPage{
		Title:   "Indigo Song Book",
		Song:    song,
		HasSong: true,
		Songs:   loadedSongs}

	temp, err := template.ParseGlob("templates/*.tmpl")
	if err != nil {
		panic(err)
	}

	if err := temp.ExecuteTemplate(w, "song.tmpl", page_data); err != nil {
		log.Println(err)
	}
}

func songHandler(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	var transpose = r.FormValue("transpose")
	if len(transpose) == 0 {
		transpose = "0"
	}
	t, err := strconv.Atoi(transpose)

	data, err := loadSongFile(p.ByName("song"), t)
	if err != nil {
		log.Println(err)
		http.NotFound(w, r)
		return
	}

	page_data := &SongPage{
		Title:        "Indigo Song Book",
		Song:         *data,
		HasSong:      true,
		Songs:        loadedSongs,
		SelectedSong: data.Title}

	temp, err := template.ParseGlob("templates/*.tmpl")
	if err != nil {
		panic(err)
	}

	if err := temp.ExecuteTemplate(w, "song.tmpl", page_data); err != nil {
		log.Println(err)
	}
}

func pdfHandler(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	var transpose = r.FormValue("transpose")
	if len(transpose) == 0 {
		transpose = "0"
	}
	t, err := strconv.Atoi(transpose)

	song, err := loadSongFile(p.ByName("song"), t)
	if err != nil {
		log.Println(err)
		http.NotFound(w, r)
		return
	}

	pdf, err := WriteSongPDF(song)

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// stream straight to client(browser)
	w.Header().Set("Content-type", "application/pdf")

	if _, err := pdf.WriteTo(w); err != nil {
		fmt.Fprintf(w, "%s", err)
	}

	w.Write([]byte("PDF Generated"))
}

func walkpath(path string, f os.FileInfo, err error) error {
	if f.IsDir() && f.Name() != "songs_master" {
		return filepath.SkipDir
	} else if f.IsDir() {
		return nil
	}

	song, err := ParseSongFile("songs_master/"+f.Name(), 0)
	link := song.Filename[0 : len(song.Filename)-len(".song")]
	loadedSongs = append(loadedSongs, DisplaySong{Link: link, Title: song.Title})

	return err
}

// readLines reads a whole file into memory
// and returns a slice of its lines.
func readLines(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines, scanner.Err()
}
