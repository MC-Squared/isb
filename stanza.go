package main

//Stanza represents an individual stanza of a song, including the chorus.
//Essentially a Stanza is a collection of Lines, along with
//Comments that appear before and after the Stanza.
type Stanza struct {
	ShowNumber     bool
	IsChorus       bool
	Number         int
	BeforeComments []string
	AfterComments  []string
	Lines          []Line
}

//HasChords returns true if any Lines in this Stanza have any Chords,
//false otherwise.
func (stanza Stanza) HasChords() bool {
	for _, l := range stanza.Lines {
		if l.HasChords() {
			return true
		}
	}

	return false
}
