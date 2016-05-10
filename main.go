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
	"strings"

	"github.com/julienschmidt/httprouter"
)

var (
	songs_root = "./songs"
	books_root = "./books"
)

func main() {
	filepath.Walk(songs_root, loadSongs)
	filepath.Walk(books_root, loadBooks)

	fmt.Printf("%d Songs loaded.\n", len(loadedSongs))
	fmt.Printf("%d Books loaded.\n", len(loadedBooks))

	r := httprouter.New()

	r.GET("/", indexHandler)
	r.GET("/index.php", indexHandler)
	r.GET("/index.html", indexHandler)
	r.GET("/song/:song", songHandler)
	r.GET("/pdf/:song", pdfHandler)
	r.GET("/book/:book", bookIndexHandler)
	r.GET("/book/:book/:number", bookHandler)
	r.ServeFiles("/css/*filepath", http.Dir("css"))
	r.ServeFiles("/js/*filepath", http.Dir("js"))

	log.Fatal(http.ListenAndServe(":8090", r))

}

// indexTemplate is the main site template.
// The default template includes two template blocks ("sidebar" and "content")
// that may be replaced in templates derived from this one.
//var indexTemplate = template.Must(template.ParseFiles("templates/index.tmpl"))

type DisplayList struct {
	Link  string
	Title string
}

func (song DisplayList) MatchTitle(title string) bool {
	return song.Title == title
}

type IndexPage struct {
	Title        string
	Songs        []DisplayList
	Books        []DisplayList
	ShowIndigo   bool
	SelectedSong string
	SelectedBook string
}

type SongPage struct {
	Song Song
	IndexPage
}

type BookPage struct {
	Songbook Songbook
	BookLink string
	IndexPage
}

var loadedSongs = make([]DisplayList, 0)
var loadedBooks = make([]DisplayList, 0)

// indexHandler is an HTTP handler that serves the index page.
func indexHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	t, err := template.ParseFiles("templates/index.tmpl")
	if err != nil {
		panic(err)
	}

	data := getBasicIndexData()
	data.ShowIndigo = true

	if err := t.ExecuteTemplate(w, "index.tmpl", data); err != nil {
		log.Println(err)
	}
}

// imageTemplate is a clone of indexTemplate that provides
// alternate "sidebar" and "content" templates.
//var songTemplate = template.Must(template.Must(indexTemplate.Clone()).ParseFiles("templates/song.tmpl"))

func loadSongFile(title string, transpose int) (*Song, error) {
	filename := songs_root + "/" + title + ".song"

	song, err := ParseSongFile(filename, transpose)

	if err != nil {
		return nil, err
	}

	return song, nil
}

func getBasicIndexData() IndexPage {
	index_data := IndexPage{
		Title:        "Indigo Song Book",
		Songs:        loadedSongs,
		Books:        loadedBooks,
		ShowIndigo:   false,
		SelectedSong: "",
		SelectedBook: ""}

	return index_data
}

func bookIndexHandler(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	sbook, err := ParseSongbookFile(books_root+"/"+p.ByName("book")+".songlist", songs_root)

	if err != nil {
		fmt.Println(err)
	}

	index := getBasicIndexData()
	index.SelectedBook = sbook.Title

	book_data := &BookPage{
		Songbook:  *sbook,
		BookLink:  sbook.Filename[0 : len(sbook.Filename)-len(".songlist")],
		IndexPage: index}

	temp, err := template.ParseFiles("templates/index.tmpl", "templates/book.tmpl")
	if err != nil {
		panic(err)
	}

	if err := temp.ExecuteTemplate(w, "book.tmpl", book_data); err != nil {
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

	sbook, err := ParseSongbookFile(books_root+"/"+p.ByName("book")+".songlist", songs_root)

	if err != nil {
		fmt.Println(err)
	}

	song := sbook.Songs[n]
	song.Transpose = t

	index := getBasicIndexData()
	index.SelectedSong = song.Title
	index.SelectedBook = sbook.Title

	page_data := &SongPage{
		Song:      song,
		IndexPage: index}

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

	index := getBasicIndexData()
	index.SelectedSong = data.Title

	page_data := &SongPage{
		Song:      *data,
		IndexPage: index}

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

func loadSongs(path string, f os.FileInfo, err error) error {
	if f.IsDir() && f.Name() != songs_root[2:] {
		return filepath.SkipDir
	} else if f.IsDir() {
		return nil
	}

	if strings.HasSuffix(strings.ToLower(f.Name()), ".song") {
		song, err := ParseSongFile(songs_root+"/"+f.Name(), 0)
		link := song.Filename[0 : len(song.Filename)-len(".song")]
		loadedSongs = append(loadedSongs, DisplayList{Link: link, Title: song.Title})

		return err
	}

	return nil
}

func loadBooks(path string, f os.FileInfo, err error) error {
	if f.IsDir() && f.Name() != books_root[2:] {
		return filepath.SkipDir
	} else if f.IsDir() {
		return nil
	}

	if strings.HasSuffix(strings.ToLower(f.Name()), ".songlist") {
		book, err := ParseSongbookFile(books_root+"/"+f.Name(), songs_root)
		link := book.Filename[0 : len(book.Filename)-len(".songlist")]
		loadedBooks = append(loadedBooks, DisplayList{Link: link, Title: book.Title})

		return err
	}

	return nil
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
