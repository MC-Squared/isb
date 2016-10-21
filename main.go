package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
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

	//Follow symlinks if applicable
	link, _ := os.Readlink(songs_root)
	if link != "" {
		songs_root = link
	}
	link, _ = os.Readlink(books_root)
	if link != "" {
		books_root = link
	}

	//basic sanity check
	_, err := os.Stat(songs_root)
	if os.IsNotExist(err) {
		fmt.Printf("Error: Songs path (%s) does not exist.\n", songs_root)
		return
	}

	_, err = os.Stat(books_root)
	if os.IsNotExist(err) {
		fmt.Printf("Error: Books path (%s) does not exist.\n", books_root)
		return
	}

	//Load songs and books
	err = loadSongs(songs_root)
	if err != nil {
		fmt.Println("Error loading songs")
		fmt.Println(err)
		return
	}

	err = loadBooks(books_root)
	if err != nil {
		fmt.Println("Error loading books")
		fmt.Println(err)
		return
	}

	fmt.Printf("%d Songs loaded.\n", len(loadedSongs))

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
	r.GET("/new-book/", newBookHandler)
	r.GET("/book/:book/index", bookIndexHandler)
	r.GET("/book/:book/song/:number", bookHandler)
	r.ServeFiles("/css/*filepath", http.Dir("css"))
	r.ServeFiles("/js/*filepath", http.Dir("js"))

	r.POST("/new-book/", newBookPostHandler)
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

func newBookPostHandler(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	raw_settings := r.PostFormValue("settings")
	var settings map[string]string
	_ = json.Unmarshal([]byte(raw_settings), &settings)

	raw_songs := r.PostFormValue("songs")
	var songs []string

	_ = json.Unmarshal([]byte(raw_songs), &songs)

	//open songbook file
	name := strings.TrimSpace(settings["name"])
	file, err := os.Create(books_root + "/" + name + ".songlist")

	defer file.Close()
	defer loadBooks(books_root)

	if err != nil {
		fmt.Println(err)
		return
	}

	//write out the settings
	for i, k := range settings {
		switch i {
		case "name":
			continue
		case "fixed-order":
			file.WriteString("{fixed_order}\n")
			break
		case "index-pos":
			file.WriteString("{index_position: ")
			file.WriteString(k)
			file.WriteString("}\n")
			break
		case "use-chorus":
			file.WriteString("{index_use_chorus}\n")
			break
		case "use-sections":
			file.WriteString("{index_use_sections}\n")
			break
		default:
			fmt.Println("Unknown book setting: ", i, " -> ", k)
			break
		}
	}

	for _, s := range songs {
		_, err = file.WriteString(strings.TrimSpace(s))
		if err == nil {
			_, err = file.WriteString("\n")
		}

		if err != nil {
			fmt.Println(err)
			return
		}
	}
}

func newBookHandler(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	temp, err := template.ParseFiles(
		"templates/index.tmpl",
		"templates/book_new.tmpl")
	if err != nil {
		panic(err)
	}

	if err := temp.ExecuteTemplate(w, "book_new.tmpl", getBasicIndexData()); err != nil {
		log.Println(err)
	}
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

	pdf, err := WriteBookPDFElectronic(sbook)

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

func loadSongs(path string) error {
	song_files, err := ioutil.ReadDir(songs_root)
	if err != nil {
		return err
	}

	for _, f := range song_files {
		if !f.IsDir() && strings.HasSuffix(strings.ToLower(f.Name()), ".song") {
			song, err := ParseSongFile(songs_root+"/"+f.Name(), 0)
			if err != nil {
				return err
			}

			link := song.Filename[0 : len(song.Filename)-len(".song")]
			loadedSongs = append(loadedSongs, DisplayList{Link: link, Title: song.Title})
		}
	}

	return err
}

func loadBooks(path string) error {
	book_files, err := ioutil.ReadDir(books_root)
	if err != nil {
		return err
	}

	loadedBooks = make([]DisplayList, 0)

	for _, f := range book_files {
		if !f.IsDir() && strings.HasSuffix(strings.ToLower(f.Name()), ".songlist") {
			book, err := ParseSongbookFile(books_root+"/"+f.Name(), songs_root)
			if err != nil {
				return err
			}
			link := book.Filename[0:len(book.Filename)-len(".songlist")] + "/index"
			loadedBooks = append(loadedBooks, DisplayList{Link: link, Title: book.Title})
		}
	}

	fmt.Printf("%d Books loaded.\n", len(loadedBooks))

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
