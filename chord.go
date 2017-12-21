package main

//Chord represents an individual chord in a Song.
//Including the text displayed, its position on a lyric,
//and any key transpositions to apply.
type Chord struct {
	text      string
	Position  int
	Transpose int
}

//GetText returns the text to be displayed for this Chord.
//This will apply any needed key transposition.
func (chord Chord) GetText() string {
	return transposeKey(chord.text, chord.Transpose)
}

var scales = map[string]int{
	"A":  0,
	"Bb": 1,
	"B":  2,
	"C":  3,
	"C#": 4,
	"D":  5,
	"D#": 6,
	"E":  7,
	"F":  8,
	"F#": 9,
	"G":  10,
	"G#": 11,
}

//transposeKey transposes the given <key> by <change> half-notes.
//Any unknown characters are skipped, as this allows for Chord text such as
//"G-C-G" to be transposed.
func transposeKey(key string, change int) string {
	if change == 0 {
		return key
	}

	newKey := ""

	for i := 0; i < len(key); i++ {
		k := key[i:]

		//check 2 characters first,
		//In case there are sharps or flats
		if len(k) > 1 {
			k = key[i : i+2]
		}

		scaleInd, ok := scales[k]
		if !ok && len(k) > 1 {
			k = k[0:1]
			scaleInd, ok = scales[k]
		}

		if !ok {
			newKey += k
		} else {
			scaleInd = (scaleInd + change) % (len(scales))
			for k := range scales {
				if scales[k] == scaleInd {
					newKey += k
				}
			}
		}

		if len(k) > 1 {
			i += len(k) - 1
		}
	}

	return newKey
}
