package main

import "testing"

var chord_tests = []struct {
	in       Line
	expected bool
}{
	{
		Line{
			Text:   "Line1",
			Chords: []Chord{Chord{}},
		},
		true,
	},

	{
		Line{
			Text:   "Line2",
			Chords: []Chord{},
		},
		false,
	},
}

func TestHasChords(t *testing.T) {
	for _, ct := range chord_tests {
		actual := ct.in.HasChords()

		if actual != ct.expected {
			t.Errorf("Line(%s), expected %v, actual %v", ct.in.Text, ct.expected, actual)
		}
	}
}

var echo_test = []struct {
	in       Line
	expected bool
}{
	{
		Line{
			Text:      "Line1",
			EchoIndex: 0,
		},
		true,
	},
	{
		Line{
			Text:      "Line2",
			EchoIndex: -1,
		},
		false,
	},
	{
		Line{
			Text:      "Line3",
			EchoIndex: 100,
		},
		true,
	},
}

func TestHasEcho(t *testing.T) {
	for _, ct := range echo_test {
		actual := ct.in.HasEcho()

		if actual != ct.expected {
			t.Errorf("Line(%s), expected %v, actual %v", ct.in.Text, ct.expected, actual)
		}
	}
}

var pre_echo_test = []struct {
	in            Line
	expected_pre  string
	expected_post string
}{
	{
		Line{
			Text:      "Test string with no echo",
			EchoIndex: -1,
		},
		"Test string with no echo",
		"",
	},
	{
		Line{
			Text:      "Test string with echo index = 10",
			EchoIndex: 10,
		},
		"Test strin",
		"g with echo index = 10",
	},
	{
		Line{
			Text:      "Test string with echo index = 0",
			EchoIndex: 0,
		},
		"",
		"Test string with echo index = 0",
	},
	{
		Line{
			Text:      "Test string with echo index = 32",
			EchoIndex: 32,
		},
		"Test string with echo index = 32",
		"",
	},
}

func TestPreEchoText(t *testing.T) {
	for _, ct := range pre_echo_test {
		actual := ct.in.PreEchoText()

		if actual != ct.expected_pre {
			t.Errorf("Line(%s), expected '%v', actual '%v'", ct.in.Text, ct.expected_pre, actual)
		}
	}
}

func TestEchoText(t *testing.T) {
	for _, ct := range pre_echo_test {
		actual := ct.in.EchoText()

		if actual != ct.expected_post {
			t.Errorf("Line(%s), expected '%v', actual '%v'", ct.in.Text, ct.expected_post, actual)
		}
	}
}

var pre_chord_test = []struct {
	line     Line
	chord    Chord
	expected string
}{
	{
		Line{
			Text: "Test string with no chords",
		},
		Chord{
			Position: 5,
		},
		"",
	},
	{
		Line{
			Text:   "Test string with one chord at position 7",
			Chords: []Chord{Chord{Position: 7}},
		},
		Chord{
			Position: 7,
		},
		"Test st",
	},
	{
		Line{
			Text:   "Test string with two chords at position 10 and 20",
			Chords: []Chord{Chord{Position: 10}, Chord{Position: 20}},
		},
		Chord{
			Position: 20,
		},
		"g with two",
	},
	{
		Line{
			Text:   "Test string with non-existing chord",
			Chords: []Chord{Chord{Position: 10}, Chord{Position: 20}},
		},
		Chord{
			Position: 15,
		},
		"",
	},
	{
		Line{
			Text:   "Test string with long chord text chord",
			Chords: []Chord{Chord{Position: 10, Text: "Text-that-is-too-long"}, Chord{Position: 20}},
		},
		Chord{
			Position: 20,
		},
		"",
	},
}

func TestPreChordText(t *testing.T) {
	for _, ct := range pre_chord_test {
		actual := ct.line.PreChordText(ct.chord)

		if actual != ct.expected {
			t.Errorf("Line(%s), expected '%v', actual '%v'", ct.line.Text, ct.expected, actual)
		}
	}
}

var split_line_test = []struct {
	line           Line
	expected1      string
	expected2      string
	expected_echo1 int
	expected_echo2 int
}{
	{
		Line{
			Text:      "Test string with comma towards the, end",
			EchoIndex: -1,
		},
		"Test string with ",
		"comma towards the, end",
		-1,
		-1,
	},
	{
		Line{
			Text:      "Test string with no comma anywhere",
			EchoIndex: -1,
		},
		"Test string with ",
		"no comma anywhere",
		-1,
		-1,
	},
	{
		Line{
			Text:      "Test string with, comma towards the, middle",
			EchoIndex: -1,
		},
		"Test string with, ",
		"comma towards the, middle",
		-1,
		-1,
	},
	{
		Line{
			Text:      "Test string with echo text at the end",
			EchoIndex: 32,
		},
		"Test string with echo text at th",
		"e end",
		-1,
		0,
	},
	{
		Line{
			Text:      "Test string with echo text from the start",
			EchoIndex: 0,
		},
		"Test string with echo ",
		"text from the start",
		0,
		0,
	},
	{
		Line{
			Text:      "Test_string_with_no_spaces_or_commas_at_all",
			EchoIndex: -1,
		},
		"Test_string_with_no_s",
		"paces_or_commas_at_all",
		-1,
		-1,
	},
}

func TestSplitLine(t *testing.T) {
	for _, ct := range split_line_test {
		actual := ct.line.SplitLine()

		if actual[0].Text != ct.expected1 {
			t.Errorf("Line(%s), expected1 '%v', actual '%v'", ct.line.Text, ct.expected1, actual[0].Text)
		} else if actual[1].Text != ct.expected2 {
			t.Errorf("Line(%s), expected2 '%v', actual '%v'", ct.line.Text, ct.expected2, actual[1].Text)
		}
	}
}

func TestSplitLineEchoIndex(t *testing.T) {
	for _, ct := range split_line_test {
		actual := ct.line.SplitLine()

		if actual[0].EchoIndex != ct.expected_echo1 {
			t.Errorf("Line(%s), expected_echo1 '%v', actual '%v'", ct.line.Text, ct.expected_echo1, actual[0].EchoIndex)
		}

		if actual[1].EchoIndex != ct.expected_echo2 {
			t.Errorf("Line(%s), expected_echo2 '%v', actual '%v'", ct.line.Text, ct.expected_echo2, actual[1].EchoIndex)
		}
	}
}
