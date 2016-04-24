package main

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
	if line.EchoIndex < 0 {
		return line.Text
	}

	return line.Text[0:line.EchoIndex]
}

func (line Line) EchoText() string {
	if line.EchoIndex < 0 {
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

	pos := line.Chords[ind].Position + len(line.Chords[ind].Text)

	return line.Text[pos:chord.Position]
}
