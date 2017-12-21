package main

import "testing"

var transposeTests = []struct {
	in       Chord
	expected string
}{
	{
		Chord{text: "A", Position: 1, Transpose: 0},
		"A",
	},
	{
		Chord{text: "A", Position: 1, Transpose: 2},
		"B",
	},
	{
		Chord{text: "G-C-G", Position: 1, Transpose: 4},
		"B-E-B",
	},
	{
		Chord{text: "G-C-G", Position: 1, Transpose: 12},
		"G-C-G",
	},
	{
		Chord{text: "G-C-G", Position: 1, Transpose: 3},
		"Bb-D#-Bb",
	},
	{
		Chord{text: "Gsus2", Position: 1, Transpose: 2},
		"Asus2",
	},
	{
		Chord{text: "G#", Position: 1, Transpose: 2},
		"Bb",
	},
}

func TestChordTransposition(t *testing.T) {
	for _, ct := range transposeTests {
		actual := ct.in.GetText()

		if actual != ct.expected {
			t.Errorf("Chord(%s:%d), expected %v, actual %v", ct.in.text, ct.in.Transpose, ct.expected, actual)
		}
	}
}
