package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

//Structs used when parsing a song file
type Songbook struct {
	FixedOrder    bool
	UseSection    bool
	IndexChorus   bool
	IndexPosition int
	Filename      string
	Songs         []Song
}

const (
	IndexNone  = 0
	IndexStart = 1
	IndexEnd   = 2
)

func ParseSongbookFile(filename string) (*Songbook, error) {
	file, err := os.Open(filename)

	filename = filepath.Base(filename)

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
		songs       = make([]Song, 0)
	)

	//bad_command_regex := regexp.MustCompile("\\{|\\}")

	for scanner.Scan() {
		line := scanner.Text()

		//is this a command
		if strings.HasPrefix(line, "{") {
			command := strings.ToLower(line)
			if strings.HasPrefix(command, "{fixed_order}") {
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
			song, err := ParseSongFile("./songs_master/"+line, 0)

			if err != nil {
				fmt.Println(err)
			} else {
				song.SongNumber = len(songs) + 1
				songs = append(songs, *song)
			}
		}
	}

	return &Songbook{
			FixedOrder:    fixed_order,
			Filename:      filename,
			UseSection:    use_section,
			IndexChorus:   use_chorus,
			IndexPosition: index_pos,
			Songs:         songs},
		nil
}
