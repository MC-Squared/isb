package main

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

//Structs used when parsing a song file
type Song struct {
	Filename          string
	Title             string
	Section           string
	StanzaCount       int
	SongNumber        int
	Stanzas           []Stanza
	ShowStanzaNumbers bool
	BeforeComments    []string
	AfterComments     []string
	UseLiberationFont bool
	transpose         int
}

func ParseSongFile(filename string, transpose int) (*Song, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	filename = filepath.Base(filename)

	var (
		//Song variables
		stanzas         []Stanza
		stanza_count    = 1
		title           = ""
		section         = ""
		scanner         = bufio.NewScanner(file)
		song_stanza_num = true
		//Stanza variables
		lines           []Line
		is_chorus       = false
		stanza_show_num = true
		useLibFont      = false
	)

	//We need to handle /r only as Mac OS <= 9 uses this as end-of-line marker
	//This is based on bufio/scan.go ScanLines function
	split := func(data []byte, atEOF bool) (advance int, token []byte, err error) {
		if atEOF && len(data) == 0 {
			return 0, nil, nil
		}
		i := bytes.IndexByte(data, '\n')

		if i < 0 {
			i = bytes.IndexByte(data, '\r')
		}

		ind := 0
		if i > 0 && data[i-1] == '\r' {
			ind = -1
		}

		if i >= 0 {

			// We have a full newline-terminated line.
			return i + 1, data[0 : i+ind], nil
		}
		// If we're at EOF, we have a final, non-terminated line. Return it.
		if atEOF {
			return len(data), data[0 : len(data)+ind], nil
		}
		// Request more data.
		return 0, nil, nil
	}
	scanner.Split(split)

	stanza_before_comments := make([]string, 0)
	stanza_after_comments := make([]string, 0)
	song_before_comments := make([]string, 0)
	song_after_comments := make([]string, 0)

	chord_regex := regexp.MustCompile("\\[.*?\\]")
	bad_command_regex := regexp.MustCompile("\\{|\\}")
	song_started := false

	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, "ā") {
			useLibFont = true
		}
		echo := -1

		//is this a command
		if strings.HasPrefix(line, "{") {
			command := strings.ToLower(line)
			if strings.HasPrefix(command, "{start_of_chorus}") {
				is_chorus = true
				continue
			} else if strings.HasPrefix(command, "{end_of_chorus}") {
				continue
			} else if strings.HasPrefix(command, "{title:") {
				title = parseCommand(line)
				continue
			} else if strings.HasPrefix(command, "{section:") {
				section = parseCommand(line)
				continue
			} else if strings.HasPrefix(command, "{comments:") {
				if !song_started {
					song_before_comments = append(song_before_comments, parseCommand(line))
				} else {
					if len(lines) > 0 {
						stanza_after_comments = append(stanza_after_comments, parseCommand(line))
					} else {
						stanza_before_comments = append(stanza_before_comments, parseCommand(line))
					}
				}
				continue
			} else if strings.HasPrefix(command, "{no_number") {
				if !song_started {
					song_stanza_num = false
				} else {
					stanza_show_num = false
				}
				continue
			} else if i := strings.Index(command, "{echo"); i >= 0 {
				//fall through
			} else {
				fmt.Println(filename)
				fmt.Printf("Unknown tag: %s\n", line)
				continue
			}
		}

		//blank line separates stanzas
		if len(line) == 0 {
			song_started = true

			if len(lines) > 0 {
				stanzas = append(stanzas, *&Stanza{
					Lines:          lines,
					Number:         stanza_count,
					IsChorus:       is_chorus,
					ShowNumber:     stanza_show_num,
					BeforeComments: stanza_before_comments,
					AfterComments:  stanza_after_comments})

				//Choruses do not get stanza numbers
				if !is_chorus {
					stanza_count++
				}

				is_chorus = false
				stanza_show_num = true
				lines = make([]Line, 0)
				stanza_before_comments = make([]string, 0)
				stanza_after_comments = make([]string, 0)
			}
		} else {
			song_started = true
			//check for echo marker
			if i := strings.Index(line, "{echo:"); i >= 0 {
				end := strings.Index(line, "}")

				if end < 1 {
					fmt.Printf("Bad echo tag: %s\n", line)
				} else {
					//to work out the index we have to remove the chords
					clean := chord_regex.ReplaceAllString(line, "")
					echo = strings.Index(clean, "{echo:")

					echo_txt := line[i+len("{echo:") : end]
					echo_txt = strings.TrimSpace(echo_txt)

					//remove command from text
					tmp := line[0:i]
					tmp += echo_txt

					line = tmp
				}
			}

			chords_pos := chord_regex.FindAllStringIndex(line, -1)
			chord_len := 0
			chords := make([]Chord, 0)

			for _, pos := range chords_pos {
				chord_text := line[pos[0]+1 : pos[1]-1]
				chord_len += pos[1] - pos[0]
				position := pos[1] - chord_len

				chords = append(chords, Chord{text: chord_text, Position: position, Transpose: transpose})
			}

			//remove all chord markers
			line = chord_regex.ReplaceAllString(line, "")
			lines = append(lines, Line{Text: line, Chords: chords, EchoIndex: echo})

			//check for bad commands
			for _, pos := range bad_command_regex.FindAllStringIndex(line, -1) {
				fmt.Println(filename)
				fmt.Println(line)
				for i := 0; i < pos[0]; i++ {
					fmt.Print(" ")
				}
				fmt.Println("^")
			}

			//Default title is first line text
			if len(title) == 0 {
				//Replace all quotation marks in title
				title = line
				re := regexp.MustCompile("[\"“”]")
				title = re.ReplaceAllString(title, "")
				title = strings.TrimSpace(title)

				//trim any trailing/leading puncutation
				start_reg := regexp.MustCompile("^[a-zA-Z0-9]")
				end_reg := regexp.MustCompile("[a-zA-Z0-9]$")
				start_done := false
				end_done := false

				for {
					if len(title) == 1 {
						break
					}

					pos := start_reg.FindAllStringIndex(title, 1)

					//No match, title has punctuation at the start
					if !start_done && len(pos) == 0 {
						title = title[1:]
					} else {
						start_done = true
					}

					pos = end_reg.FindAllStringIndex(title, 1)

					if !end_done && len(pos) == 0 {
						title = title[0 : len(title)-1]
					} else {
						end_done = true
					}

					if start_done && end_done {
						break
					}
				}
			}
		}
	}

	//check for last stanza
	if len(lines) > 0 {
		stanzas = append(stanzas, Stanza{
			Lines:          lines,
			Number:         stanza_count,
			IsChorus:       is_chorus,
			ShowNumber:     stanza_show_num,
			BeforeComments: stanza_before_comments,
			AfterComments:  stanza_after_comments})
	} else if len(stanza_before_comments) > 0 {
		song_after_comments = stanza_before_comments
	}

	return &Song{
			Filename:          filename,
			Title:             title,
			Section:           section,
			StanzaCount:       0,
			SongNumber:        -1,
			ShowStanzaNumbers: song_stanza_num,
			Stanzas:           stanzas,
			BeforeComments:    song_before_comments,
			AfterComments:     song_after_comments,
			UseLiberationFont: useLibFont,
			transpose:         transpose},
		nil
}

func parseCommand(command string) string {
	return strings.TrimSpace(command[strings.Index(command, ":")+1 : strings.Index(command, "}")])
}

func (song Song) GetTranspose() int {
	return song.transpose
}

func (song *Song) Transpose(change_by int) {
	song.transpose = change_by

	for _, s := range song.Stanzas {
		for _, l := range s.Lines {
			//index range here because we are modifying the Chord
			for i := range l.Chords {
				l.Chords[i].Transpose = change_by
			}
		}
	}
}

func (song Song) String() string {
	var buffer bytes.Buffer

	if len(song.Section) > 0 {
		buffer.WriteString(fmt.Sprintf("Section: %s\n", song.Section))
	}

	if len(song.BeforeComments) > 0 {
		buffer.WriteString(fmt.Sprintf("/%s/\n", song.BeforeComments))
	}

	for _, s := range song.Stanzas {
		if s.IsChorus {
			buffer.WriteString("---CHORUS--\n")
		} else {
			buffer.WriteString(fmt.Sprintf("STANZA: %d\n", s.Number))
		}
		for _, l := range s.Lines {
			buffer.WriteString(fmt.Sprintf("%s\n", l.Text))
		}
		if s.IsChorus {
			buffer.WriteString(fmt.Sprintln("---END CHORUS--"))
		}

		buffer.WriteString("\n")
	}

	return buffer.String()
}

func (song Song) HasBeforeComments() bool {
	return len(song.BeforeComments) > 0
}

func (song Song) Link() string {
	return song.Filename[0 : len(song.Filename)-5]
}
