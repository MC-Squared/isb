package main

import (
	"strings"
	"unicode/utf8"
)

//Line represents the line of a Stanza.
//A Line includes the Text, which is a string (i.e. the lyrics)
//A series of Chords that appear in this Line (may be empty)
//And an EchoIndex, indicating the portion of the Line that is an echo
// (so it can be printed differently, i.e. italic/lighter/...)
type Line struct {
	Text      string
	Chords    []Chord
	EchoIndex int
}

//HasChords returns true if this Line contains any Chord objects,
//false otherwise.
func (line Line) HasChords() bool {
	return len(line.Chords) > 0
}

//HasEcho returns true if this Line has an EchoIndex assigned,
//false otherwise.
func (line Line) HasEcho() bool {
	return line.EchoIndex >= 0
}

//PreEchoText returns this Line's Text up to EchoIndex if it is set.
//If EchoIndex is not set, then the full Text is returned.
func (line Line) PreEchoText() string {
	if line.EchoIndex < 0 || line.EchoIndex >= utf8.RuneCountInString(line.Text) {
		return line.Text
	}

	return line.Text[0:line.EchoIndex]
}

//EchoText returns this Line's text from EchoIndex to the end of the Text if
//EchoIndex is set. If EchoIndex is not set, an empty string is returned.
func (line Line) EchoText() string {
	if line.EchoIndex < 0 || line.EchoIndex >= utf8.RuneCountInString(line.Text) {
		return ""
	}

	return line.Text[line.EchoIndex:utf8.RuneCountInString(line.Text)]
}

//PreChordText returns the substring of Text that occurs before the given Chord
//and after the previous Chord (or the start of the Text if there is no previous
// Chord).
func (line Line) PreChordText(chord Chord) string {
	//first, find the chord
	ind := -1
	for i, ch := range line.Chords {
		if ch == chord {
			ind = i
			break
		}
	}

	if ind < 0 {
		return ""
	}

	//We need the text from the previous chord up to this chord
	ind--

	//chord is the first chord
	if ind < 0 {
		return line.Text[0:chord.Position]
	}

	pos := line.Chords[ind].Position + utf8.RuneCountInString(line.Chords[ind].GetText())

	//if the chord text is long,
	//We might have a problem here
	if pos > chord.Position {
		return ""
	}

	return line.Text[pos:chord.Position]
}

//SplitLine splits a single Line into two lines.
//If the line contains an Echo, the echo is the split point,
//otherwise if there is a comma near the center it is used,
//otherwise if there is a space near the center it is used,
//if all else fails, it will split at the middle character
func (line Line) SplitLine() []Line {
	var split int

	//special case
	// if utf8.RuneCountInString(line.Text) < 2 {
	// 	res := make([]Line, 0)
	// 	return append(res, line)
	// }

	//If the line has an echo, split at the echo
	if line.HasEcho() && line.EchoIndex > 0 {
		split = line.EchoIndex
	} else {
		center := utf8.RuneCountInString(line.Text) / 2
		maxWindow := utf8.RuneCountInString(line.Text) / 4

		if maxWindow < 1 {
			maxWindow = 1
		}

		//check for comma
		window := 1
		split = center
		for center > 0 && window <= maxWindow {
			split = strings.Index(line.Text[center-window:center+window], ",")
			if split > 0 {
				break
			}
			window++
		}

		//check for space
		if split < 0 {
			window = 1
			for window <= maxWindow {
				split = strings.Index(line.Text[center-window:center+window], " ")
				if split > 0 {
					break
				}
				window++
			}
		}

		if split < 0 {
			split = center
		} else {
			split += center - window + 1 //+1 so we don't include the comma/space on the next line
		}

		for line.Text[split] == ' ' || line.Text[split] == ',' {
			split++
		}
	}

	if split == 0 {
		res := make([]Line, 1)
		res[0] = line
		return res
	}

	return line.splitLineAt(split)
}

func (line Line) splitLineAt(split int) []Line {
	text1 := line.Text[0:split]
	text2 := line.Text[split:]

	chords1 := make([]Chord, 0)
	chords2 := make([]Chord, 0)

	//now split up the chords
	if line.HasChords() {
		for _, c := range line.Chords {
			if c.Position <= utf8.RuneCountInString(text1) {
				chords1 = append(chords1, c)
			} else {
				chords2 = append(chords2, Chord{c.GetText(), c.Position - utf8.RuneCountInString(text1), c.Transpose})
			}
		}
	}

	res := make([]Line, 0)
	res = append(res, Line{Text: text1, Chords: chords1, EchoIndex: -1})
	res = append(res, Line{Text: text2, Chords: chords2, EchoIndex: -1})

	if line.EchoIndex >= 0 && line.EchoIndex < split {
		res[0].EchoIndex = line.EchoIndex
		res[1].EchoIndex = 0
	} else if line.EchoIndex >= split {
		res[1].EchoIndex = line.EchoIndex - split
	}

	return res
}
