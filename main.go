package main

import (
	"bufio"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"sort"
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

	//Sort lists by title
	sort.Sort(ByTitle(loadedSongs))
	sort.Sort(ByTitle(loadedBooks))

	r := httprouter.New()

	r.GET("/", indexHandler)
	r.GET("/index.php", indexHandler)
	r.GET("/index.html", indexHandler)
	r.GET("/song/:song", songHandler)
	r.GET("/pdf/song/:song", songPdfHandler)
	r.GET("/pdf/book/:book", bookPdfHandler)
	r.GET("/book/:book/index", bookIndexHandler)
	r.GET("/book/:book/song/:number", bookHandler)
	r.ServeFiles("/css/*filepath", http.Dir("css"))
	r.ServeFiles("/js/*filepath", http.Dir("js"))

	log.Fatal(http.ListenAndServe(":8090", r))

}

type DisplayList struct {
	Link  string
	Title string
}

// ByTitle implements sort.Interface for []DisplayList based on
// the Title field.
type ByTitle []DisplayList

func (a ByTitle) Len() int      { return len(a) }
func (a ByTitle) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a ByTitle) Less(i, j int) bool {
	num_regex := regexp.MustCompile("^[0-9]")

	i_pos := num_regex.FindAllStringIndex(a[i].Title, 1)
	j_pos := num_regex.FindAllStringIndex(a[j].Title, 1)

	i_num := (len(i_pos) > 0)
	j_num := (len(j_pos) > 0)

	//neither title begins with a number,
	//or both begin with a number, normal sort
	if i_num == j_num {
		return a[i].Title < a[j].Title
	}

	//I is a number, put it at the end
	if i_num {
		return false
	}

	return true
}

func (song DisplayList) MatchTitle(title string) bool {
	return song.Title == title
}

func (song Song) MatchNumber(num int) bool {
	return song.SongNumber == num
}

type IndexPage struct {
	Title        string
	Recent       []DisplayList
	Songs        []DisplayList
	Books        []DisplayList
	ShowIndigo   bool
	SelectedSong string
	SelectedBook string
	Error        string
}

func (i IndexPage) HasSong() bool {
	return len(i.SelectedSong) > 0
}

func (i IndexPage) HasBook() bool {
	return len(i.SelectedBook) > 0
}

type SongPage struct {
	Song     Song
	Songbook Songbook
	NextSong int
	PrevSong int
	Selected int
	IndexPage
}

type BookPage struct {
	Songbook Songbook
	Selected int
	IndexPage
}

var loadedSongs = make([]DisplayList, 0)
var loadedBooks = make([]DisplayList, 0)
var recent = make([]DisplayList, 0)
var last_err = ""

// indexHandler is an HTTP handler that serves the index page.
func indexHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	t, err := template.ParseFiles("templates/index.tmpl", "templates/_song_select.tmpl", "templates/_book_select.tmpl")
	if err != nil {
		panic(err)
	}

	data := getBasicIndexData()
	data.ShowIndigo = true

	if err := t.ExecuteTemplate(w, "index.tmpl", data); err != nil {
		log.Println(err)
	}
}

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
		Recent:       recent,
		Songs:        loadedSongs,
		Books:        loadedBooks,
		ShowIndigo:   false,
		SelectedSong: "",
		SelectedBook: "",
		Error:        last_err}

	last_err = ""

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
		IndexPage: index}

	temp, err := template.ParseFiles(
		"templates/index.tmpl",
		"templates/_book_navigation.tmpl",
		"templates/book_index.tmpl")
	if err != nil {
		panic(err)
	}

	if err := temp.ExecuteTemplate(w, "book_index.tmpl", book_data); err != nil {
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
		http.NotFound(w, r)
	}

	keys := GetSongOrder(sbook)

	prev := n
	next := n
	for _, i := range keys {
		if i < n {
			prev = i
		}

		if i > n {
			next = i
			break
		}
	}

	if next == n {
		next = 0
	}

	if prev == n {
		prev = 0
	}

	song, ok := sbook.Songs[n]

	//Number does not exist in songbook
	if !ok {
		index := httprouter.CleanPath(r.URL.String() + "/../../index")
		last_err = "Song '" + strconv.Itoa(n) + "' not found in songbook."
		http.Redirect(w, r, index, 302)
		return
	}
	song.Transpose = t

	index := getBasicIndexData()
	index.SelectedSong = song.Title
	index.SelectedBook = sbook.Title

	updateRecent(song.Link(), song.Title)

	page_data := &SongPage{
		Song:      song,
		Songbook:  *sbook,
		PrevSong:  prev,
		NextSong:  next,
		Selected:  song.SongNumber,
		IndexPage: index}

	temp, err := template.ParseFiles(
		"templates/index.tmpl",
		"templates/_song_select.tmpl",
		"templates/_book_select.tmpl",
		"templates/_display_song.tmpl",
		"templates/_book_navigation.tmpl",
		"templates/book_song.tmpl")
	if err != nil {
		panic(err)
	}

	if err := temp.ExecuteTemplate(w, "book_song.tmpl", page_data); err != nil {
		log.Println(err)
	}
}

func updateRecent(link string, title string) {
	//if the song is already in the list, move it to the top
	for i, dl := range recent {
		if dl.Title == title {
			//remove element
			recent = append(recent[:i], recent[i+1:]...)
			break
		}
	}

	recent = append([]DisplayList{DisplayList{Link: link, Title: title}}, recent...)
	if len(recent) > 5 {
		recent = recent[0:5]
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

		home := httprouter.CleanPath(r.URL.String() + "/../..")
		last_err = "Song '" + p.ByName("song") + "' not found."
		http.Redirect(w, r, home, 302)
		return
	}

	updateRecent(p.ByName("song"), data.Title)

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

func songPdfHandler(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
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

func bookPdfHandler(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	sbook, err := ParseSongbookFile(books_root+"/"+p.ByName("book")+".songlist", songs_root)

	if err != nil {
		log.Println(err)
		http.NotFound(w, r)
		return
	}

	pdf, err := WriteBookPDF(sbook)

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	if pdf == nil {
		return
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
		link := book.Filename[0:len(book.Filename)-len(".songlist")] + "/index"
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
