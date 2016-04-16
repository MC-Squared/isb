package main

import (
    "os"
    "bufio"
    "strings"
    "bytes"
    "fmt"
)

//Structs used when parsing a song file
type Song struct {
    Title string
    Section string
    StanzaCount int
    SongNumber int
    Stanzas []Stanza
}

type Stanza struct {
    ShowNumber bool
    HasChords bool
    IsChorus bool
    Number int
    StartComments []string
    EndComments []string
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

    //var text []string
    var lines []Line
    var stanzas []Stanza
    stanza_count := 1
    is_chorus := false

    scanner := bufio.NewScanner(file)
    for scanner.Scan() {
        line := scanner.Text()

        //is this a command
        if strings.HasPrefix(line, "{") {
            switch line {
                case "{start_of_chorus}":
                    is_chorus = true
                    break
                case "{end_of_chorus}":
                    //is_chorus = false
                    break
            }

        } else {
            //blank line separates stanzas
            if len(line) == 0 {
                if len(lines) > 0 {
                    stanzas = append(stanzas, *&Stanza{
                        Lines: lines,
                        Number: stanza_count,
                        IsChorus: is_chorus})

                    stanza_count++
                    is_chorus = false
                    lines = make([]Line, 0)

                    if (is_chorus) {
                        stanza_count++
                    }
                }
            } else {
                lines = append(lines, *&Line{HasChords: false, Text: line})
            }
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
        Title: "Title",
        Section: "",
        StanzaCount: 0,
        SongNumber: -1,
        Stanzas: stanzas},
        nil
}

func (song Song) String() string {
    var buffer bytes.Buffer

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
