package main

import "testing"

var stanzaChordTests = []struct {
	in       Stanza
	expected bool
}{
	{
		Stanza{},
		false,
	},
	{
		Stanza{Lines: []Line{
			Line{
				Text:   "Line2",
				Chords: []Chord{},
			},
				}},
		false,
	},
	{
		Stanza{Lines: []Line{
			Line{
				Text:   "Line2",
				Chords: []Chord{Chord{}},
			},
				}},
		true,
	},
}



func TestStanzaHasChords(t *testing.T) {
	for i, ct := range stanzaChordTests {
		actual := ct.in.HasChords()

		if actual != ct.expected {
			t.Errorf("Stanza(%d), expected %v, actual %v", i, ct.expected, actual)
		}
	}
}
