package main

import (
    "os"
    "bufio"
    "strings"
    "bytes"
    "fmt"
    "path/filepath"
)

//Structs used when parsing a song file
type Song struct {
    Title string
    Section string
    StanzaCount int
    SongNumber int
    Stanzas []Stanza
    BeforeComments []string
    AfterComments []string
}

type Stanza struct {
    ShowNumber bool
    HasChords bool
    IsChorus bool
    Number int
    BeforeComments []string
    AfterComments []string
    Lines []Line
}

type Line struct {
    HasChords bool
    Text string
    Chords []Chord
}

type Chord struct {
    Text string
    Position int
}

func ParseSongFile(filename string) (*Song, error) {
    file, err := os.Open(filename)
    if err != nil {
        return nil, err
    }
    defer file.Close()

    var (
        lines []Line
        stanzas []Stanza
        stanza_count = 1
        is_chorus = false
        title = filepath.Base(filename)[0:len(filepath.Base(filename))-5]
        section = ""
        scanner = bufio.NewScanner(file)
    )

    stanza_before_comments := make([]string, 0)
    stanza_after_comments := make([]string, 0)
    song_before_comments := make([]string, 0)
    //song_after_comments := make([]string, 0)
    
    for scanner.Scan() {
        line := scanner.Text()

        //is this a command
        if strings.HasPrefix(line, "{") {
            if (strings.HasPrefix(line, "{start_of_chorus}")) {
                is_chorus = true
            } else if (strings.HasPrefix(line,"{end_of_chorus}")) {

            } else if (strings.HasPrefix(line, "{title:")) {
                title = parseCommand(line)
            } else if (strings.HasPrefix(line, "{section:")) {
                section = parseCommand(line)
            } else if (strings.HasPrefix(line, "{comments:")) {
                if (len(stanzas) == 0) {
                    song_before_comments = append(song_before_comments, parseCommand(line))
                } else {
                    if (len(lines) > 0) {
                        stanza_after_comments = append(stanza_after_comments, parseCommand(line))
                    } else {
                        stanza_before_comments = append(stanza_before_comments, parseCommand(line))
                    }
                }
            }
        //blank line separates stanzas
        } else if len(line) == 0 {
            if len(lines) > 0 {
                stanzas = append(stanzas, *&Stanza{
                    Lines: lines,
                    Number: stanza_count,
                    IsChorus: is_chorus,
                    BeforeComments: stanza_before_comments,
                    AfterComments: stanza_after_comments})

                stanza_count++
                is_chorus = false
                lines = make([]Line, 0)
                stanza_before_comments = make([]string, 0)
                stanza_after_comments = make([]string, 0)

                if (is_chorus) {
                    stanza_count++
                }
            }
        } else {
            lines = append(lines, *&Line{HasChords: false, Text: line})
        }
    }

    //check for last stanza
    if len(lines) > 0 {
        stanzas = append(stanzas, *&Stanza{
            Lines: lines,
            Number: stanza_count,
            IsChorus: is_chorus})
    }
    
    return &Song{
        Title: title,
        Section: section,
        StanzaCount: 0,
        SongNumber: -1,
        Stanzas: stanzas,
        BeforeComments: song_before_comments},
        nil
}

func parseCommand(command string) string {
    return strings.TrimSpace(command[strings.Index(command, ":")+1:strings.Index(command, "}")])
}

func (song Song) String() string {
    var buffer bytes.Buffer

    if len(song.Section) > 0 {
        buffer.WriteString(fmt.Sprintf("Section: %s\n", song.Section))
    }

    if len(song.BeforeComments) > 0 {
        buffer.WriteString(fmt.Sprintf("/%s/\n", song.BeforeComments))
    }

    for _,s := range song.Stanzas {
        if s.IsChorus {
            buffer.WriteString("---CHORUS--\n")
        }
        buffer.WriteString(fmt.Sprintf("STANZA: %d\n", s.Number))
        for _,l := range s.Lines {
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
