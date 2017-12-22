package main

import "testing"
import "reflect"

var linkTests = []struct {
	in       Songbook
	expected string
}{
	{
		Songbook{Filename: "test-filEname.songbook"},
		"test-filEname",
	},
}

func TestSongbookLink(t *testing.T) {
	for _, ct := range linkTests {
		actual := ct.in.Link()

		if actual != ct.expected {
			t.Errorf("Songbook Link, expected %v, actual %v", ct.expected, actual)
		}
	}
}


var orderTests = []struct {
	in       Songbook
	expected []int
}{
	{
		Songbook{Songs: map[int]Song{}},
		make([]int, 0),
	},
  {
		Songbook{Songs: map[int]Song{0: Song{}}},
		make([]int, 0),
	},
}

func TestSongbookSongOrder(t *testing.T) {
	for _, ct := range orderTests {
		actual := GetSongOrder(&ct.in)

      if !reflect.DeepEqual(actual, ct.expected) {
        t.Errorf("Songbook Order, expected %v, actual %v", ct.expected, actual)
      }

    // for i, ex := range ct.expected {
    //   if actual[i] != ex {
    //     t.Errorf("Songbook Order, expected %v, actual %v", ex, actual[i])
    //   }


		// if actual != ct.expected {
		// 	t.Errorf("Songbook Link, expected %v, actual %v", ct.expected, actual)
		// }
	}
}
