package main

import "strings"

type Line struct {
	Text      string
	Chords    []Chord
	EchoIndex int
}

func (line Line) HasChords() bool {
	return len(line.Chords) > 0
}

func (line Line) HasEcho() bool {
	return line.EchoIndex >= 0
}

func (line Line) PreEchoText() string {
	if line.EchoIndex < 0 || line.EchoIndex >= len(line.Text) {
		return line.Text
	}

	return line.Text[0:line.EchoIndex]
}

func (line Line) EchoText() string {
	if line.EchoIndex < 0 || line.EchoIndex >= len(line.Text) {
		return ""
	}

	return line.Text[line.EchoIndex:len(line.Text)]
}

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

	pos := line.Chords[ind].Position + len(line.Chords[ind].GetText())

	//if the chord text is long,
	//We might have a problem here
	if pos > chord.Position {
		return ""
	}

	return line.Text[pos:chord.Position]
}

//Splits a single Line into two lines.
//If the line contains an Echo, the echo is the split point,
//otherwise if there is a comma near the center it is used,
//otherwise if there is a space near the center it is used,
//if all else fails, it will split at the middle character
func (line Line) SplitLine() []Line {
	split := -1
	//If the line has an echo, split at the echo
	if line.HasEcho() && line.EchoIndex > 0 {
		split = line.EchoIndex
	} else {
		center := len(line.Text) / 2
		max_window := len(line.Text) / 4

		if max_window < 1 {
			max_window = 1
		}

		//check for comma
		window := 1
		split = center
		for window <= max_window {
			split = strings.Index(line.Text[center-window:center+window], ",")
			if split > 0 {
				break
			}
			window++
		}

		//check for space
		if split < 0 {
			window = 1
			for window <= max_window {
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

		if line.Text[split] == ' ' {
			split++
		}
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
			if c.Position <= len(text1) {
				chords1 = append(chords1, c)
			} else {
				chords2 = append(chords2, Chord{c.GetText(), c.Position - len(text1), c.Transpose})
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
