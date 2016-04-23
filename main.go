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
	"sort"
	"strconv"
	"strings"
)

func main() {
	songs_root := "./songs_master" // 1st argument is the directory location
	filepath.Walk(songs_root, walkpath)

	fmt.Printf("%d Songs loaded.\n", len(filenames))

	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/song/", songHandler)
	http.Handle("/css/", http.StripPrefix("/css/", http.FileServer(http.Dir("css"))))
	log.Fatal(http.ListenAndServe("localhost:8090", nil))
}

// indexTemplate is the main site template.
// The default template includes two template blocks ("sidebar" and "content")
// that may be replaced in templates derived from this one.
var indexTemplate = template.Must(template.ParseFiles("templates/index.tmpl"))

// Index is a data structure used to populate an indexTemplate.
type Index struct {
	Title string
	Songs []string
}

type SongPage struct {
	Title   string
	Songs   []string
	Song    Song
	HasSong bool
}

// indexHandler is an HTTP handler that serves the index page.
func indexHandler(w http.ResponseWriter, r *http.Request) {
	data := &SongPage{
		Title:   "Indigo Song Book",
		HasSong: false,
		Songs:   filenames}

	if err := songTemplate.Execute(w, data); err != nil {
		log.Println(err)
	}
}

// imageTemplate is a clone of indexTemplate that provides
// alternate "sidebar" and "content" templates.
var songTemplate = template.Must(template.Must(indexTemplate.Clone()).ParseFiles("templates/song.tmpl"))

// Image is a data structure used to populate an imageTemplate.
type SongFile struct {
	Title string
	Lines []string
}

func loadSongFile(title string, transpose int) (*Song, error) {
	filename := "songs_master/" + title + ".song"

	song, err := ParseSongFile(filename, transpose)

	if err != nil {
		return nil, err
	}

	return song, nil
}

// imageHandler is an HTTP handler that serves the image pages.
func songHandler(w http.ResponseWriter, r *http.Request) {
	target := strings.TrimPrefix(r.URL.Path, "/song/")

	ind := sort.SearchStrings(filenames, target)
	if ind > len(filenames) && filenames[ind] != target {
		http.NotFound(w, r)
		return
	}

	var transpose = r.FormValue("transpose")
	if len(transpose) == 0 {
		transpose = "0"
	}
	t, err := strconv.Atoi(transpose)

	data, err := loadSongFile(target, t)
	if err != nil {
		log.Println(err)
		return
	}

	page_data := &SongPage{
		Title:   "Indigo Song Book",
		Song:    *data,
		HasSong: true,
		Songs:   filenames}

	if err := songTemplate.Execute(w, page_data); err != nil {
		log.Println(err)
	}
}

var filenames = make([]string, 0)

func walkpath(path string, f os.FileInfo, err error) error {
	if f.IsDir() && f.Name() != "songs_master" {
		return filepath.SkipDir
	} else if f.IsDir() {
		return nil
	}

	file := f.Name()[0 : len(f.Name())-5]
	filenames = append(filenames, file)

	fmt.Printf("%s\n", file)
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
