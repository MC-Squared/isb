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

//Song represents an invididual song read from a .song file
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

//ParseSongFile attempts to read a Song from the given filename.
//Applying any transposition to the Chords.
//Returns the newly created Song, or error on failure.
func ParseSongFile(filename string, transpose int) (*Song, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	filename = filepath.Base(filename)

	var (
		//Song variables
		stanzas       []Stanza
		stanzaCount   = 1
		title         = ""
		section       = ""
		scanner       = bufio.NewScanner(file)
		songStanzaNum = true
		//Stanza variables
		lines         []Line
		isChorus      = false
		stanzaShowNum = true
		useLibFont    = false
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

	stanzaBeforeComments := make([]string, 0)
	stanzaAfterComments := make([]string, 0)
	songBeforeComments := make([]string, 0)
	songAfterComments := make([]string, 0)

	chordRegex := regexp.MustCompile("\\[.*?\\]")
	badCommandRegex := regexp.MustCompile("\\{|\\}")
	songStarted := false

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
				isChorus = true
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
				if !songStarted {
					songBeforeComments = append(songBeforeComments, parseCommand(line))
				} else {
					if len(lines) > 0 {
						stanzaAfterComments = append(stanzaAfterComments, parseCommand(line))
					} else {
						stanzaBeforeComments = append(stanzaBeforeComments, parseCommand(line))
					}
				}
				continue
			} else if strings.HasPrefix(command, "{no_number") {
				if !songStarted {
					songStanzaNum = false
				} else {
					stanzaShowNum = false
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
			songStarted = true

			if len(lines) > 0 {
				stanzas = append(stanzas, *&Stanza{
					Lines:          lines,
					Number:         stanzaCount,
					IsChorus:       isChorus,
					ShowNumber:     stanzaShowNum,
					BeforeComments: stanzaBeforeComments,
					AfterComments:  stanzaAfterComments})

				//Choruses do not get stanza numbers
				if !isChorus {
					stanzaCount++
				}

				isChorus = false
				stanzaShowNum = true
				lines = make([]Line, 0)
				stanzaBeforeComments = make([]string, 0)
				stanzaAfterComments = make([]string, 0)
			}
		} else {
			songStarted = true
			//check for echo marker
			if i := strings.Index(line, "{echo:"); i >= 0 {
				end := strings.Index(line, "}")

				if end < 1 {
					fmt.Printf("Bad echo tag: %s\n", line)
				} else {
					//to work out the index we have to remove the chords
					clean := chordRegex.ReplaceAllString(line, "")
					echo = strings.Index(clean, "{echo:")

					echoTxt := line[i+len("{echo:") : end]
					echoTxt = strings.TrimSpace(echoTxt)

					//remove command from text
					tmp := line[0:i]
					tmp += echoTxt

					line = tmp
				}
			}

			chordsPos := chordRegex.FindAllStringIndex(line, -1)
			chordLen := 0
			chords := make([]Chord, 0)

			for _, pos := range chordsPos {
				chordText := line[pos[0]+1 : pos[1]-1]
				chordLen += pos[1] - pos[0]
				position := pos[1] - chordLen

				chords = append(chords, Chord{text: chordText, Position: position, Transpose: transpose})
			}

			//remove all chord markers
			line = chordRegex.ReplaceAllString(line, "")
			lines = append(lines, Line{Text: line, Chords: chords, EchoIndex: echo})

			//check for bad commands
			for _, pos := range badCommandRegex.FindAllStringIndex(line, -1) {
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
				startReg := regexp.MustCompile("^[a-zA-Z0-9]")
				endReg := regexp.MustCompile("[a-zA-Z0-9]$")
				startDone := false
				endDone := false

				for {
					if len(title) == 1 {
						break
					}

					pos := startReg.FindAllStringIndex(title, 1)

					//No match, title has punctuation at the start
					if !startDone && len(pos) == 0 {
						title = title[1:]
					} else {
						startDone = true
					}

					pos = endReg.FindAllStringIndex(title, 1)

					if !endDone && len(pos) == 0 {
						title = title[0 : len(title)-1]
					} else {
						endDone = true
					}

					if startDone && endDone {
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
			Number:         stanzaCount,
			IsChorus:       isChorus,
			ShowNumber:     stanzaShowNum,
			BeforeComments: stanzaBeforeComments,
			AfterComments:  stanzaAfterComments})
	} else if len(stanzaBeforeComments) > 0 {
		songAfterComments = stanzaBeforeComments
	}

	return &Song{
			Filename:          filename,
			Title:             title,
			Section:           section,
			StanzaCount:       0,
			SongNumber:        -1,
			ShowStanzaNumbers: songStanzaNum,
			Stanzas:           stanzas,
			BeforeComments:    songBeforeComments,
			AfterComments:     songAfterComments,
			UseLiberationFont: useLibFont,
			transpose:         transpose},
		nil
}

//parseCommand parses a given command string and strips off the framing characters.
//i.e. given "{command: setting}", it will return "setting"
func parseCommand(command string) string {
	return strings.TrimSpace(command[strings.Index(command, ":")+1 : strings.Index(command, "}")])
}

//GetTranspose returns the current chord transposition setting for this Song.
//The transposition is an integer representing the half-notes up or down (+/-)
//that the Chords are being adjusted. All Chords contained in this song should
//have the same Transpose setting.
func (song Song) GetTranspose() int {
	return song.transpose
}

//Transpose will iterate through all Chords contained in this song and call
//Chord.Transpose(changeBy) on each one.
func (song *Song) Transpose(changeBy int) {
	song.transpose = changeBy

	for _, s := range song.Stanzas {
		for _, l := range s.Lines {
			//index range here because we are modifying the Chord
			for i := range l.Chords {
				l.Chords[i].Transpose = changeBy
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

//HasBeforeComments returns true if this Song has any BeforeComments set,
//false otherwise.
func (song Song) HasBeforeComments() bool {
	return len(song.BeforeComments) > 0
}

//HasAfterComments returns true if this Song has any BeforeComments set,
//false otherwise.
func (song Song) HasAfterComments() bool {
	return len(song.AfterComments) > 0
}

//Link provides a substring of this Song's Filename as a way to easily
//provide HTML links.
//i.e. if Filename is ".../song name.song" Link will return "song name"
func (song Song) Link() string {
	return song.Filename[0 : len(song.Filename)-5]
}
