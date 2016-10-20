package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

//Structs used when parsing a song file
type Songbook struct {
	FixedOrder    bool
	UseSection    bool
	IndexChorus   bool
	IndexPosition int
	Filename      string
	Title         string
	Songs         map[int]Song
}

const (
	IndexNone  = 0
	IndexStart = 1
	IndexEnd   = 2
)

func ParseSongbookFile(filename string, songs_root string) (*Songbook, error) {
	file, err := os.Open(filename)

	filename = filepath.Base(filename)
	title := filename[0 : len(filename)-len(".songlist")]

	if err != nil {
		return nil, err
	}
	defer file.Close()

	var (
		scanner     = bufio.NewScanner(file)
		fixed_order = false
		use_section = false
		use_chorus  = false
		index_pos   = IndexNone
		songs       = make(map[int]Song)
	)

	//bad_command_regex := regexp.MustCompile("\\{|\\}")

	for scanner.Scan() {
		line := scanner.Text()

		//is this a command
		if strings.HasPrefix(line, "{") {
			command := strings.ToLower(line)
			if strings.HasPrefix(command, "{title:") {
				title = parseCommand(line)
				continue
			} else if strings.HasPrefix(command, "{fixed_order}") {
				fixed_order = true
				continue
			} else if strings.HasPrefix(command, "{index_use_sections}") {
				use_section = true
				continue
			} else if strings.HasPrefix(command, "{index_use_chorus}") {
				use_chorus = true
				continue
			} else if strings.HasPrefix(command, "{index_position:") {
				p := parseCommand(line)

				switch p {
				case "start":
					index_pos = IndexStart
					break
				case "end":
					index_pos = IndexEnd
					break
				case "none":
				default:
					index_pos = IndexNone
				}

				continue
			} else {
				fmt.Printf("Unknown tag: %s\n", line)
				continue
			}
		}

		//ignore blank lines
		if len(line) > 0 {
			num := -1
			//check for fixed numbering
			if strings.Index(line, ",") > 0 {
				num_str := line[0:strings.Index(line, ",")]
				num, err = strconv.Atoi(num_str)
				if err == nil {
					line = line[len(num_str)+1 : len(line)]
				} else {
					num = -1
				}
			}

			line = strings.TrimSpace(line)

			//including '.song' extension is optional
			if !strings.HasSuffix(line, ".song") {
				line += ".song"
			}

			song, err := ParseSongFile(songs_root+"/"+line, 0)

			if err != nil {
				fmt.Println(num, ":", err)
			} else {
				if num < 0 {
					num = len(songs) + 1
				}
				song.SongNumber = num
				songs[num] = *song
			}
		}
	}

	return &Songbook{
			Title:         title,
			FixedOrder:    fixed_order,
			Filename:      filename,
			UseSection:    use_section,
			IndexChorus:   use_chorus,
			IndexPosition: index_pos,
			Songs:         songs},
		nil
}

func GetSongOrder(sbook *Songbook) (keys []int) {
	keys = make([]int, len(sbook.Songs))
	i := 0
	for k := range sbook.Songs {
		keys[i] = k
		i++
	}
	sort.Sort(sort.IntSlice(keys))

	return keys
}

func GetSongSlice(sbook *Songbook) (songs []Song) {
	keys := GetSongOrder(sbook)
	songs = make([]Song, len(sbook.Songs))

	ind := 0
	for _, k := range keys {
		songs[ind] = sbook.Songs[k]
		ind++
	}

	return songs
}

func (sbook Songbook) Link() string {
	return sbook.Filename[0 : len(sbook.Filename)-9]
}
