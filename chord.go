package main

type Chord struct {
	Text     string
	Position int
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

func transposeKey(key string, change int) string {
	if change == 0 {
		return key
	}

	new_key := ""

	for i := 0; i < len(key); i++ {
		k := key[i:]

		//check 2 characters first,
		//In case there are sharps or flats
		if len(k) > 1 {
			k = key[i : i+2]
		}

		scale_ind, ok := scales[k]
		if !ok && len(k) > 1 {
			k = k[0:1]
			scale_ind, ok = scales[k]
		}

		if !ok {
			new_key += k
		} else {
			scale_ind = (scale_ind + change) % (len(scales))
			for k := range scales {
				if scales[k] == scale_ind {
					new_key += k
				}
			}
		}

		if len(k) > 1 {
			i += len(k) - 1
		}
	}

	return new_key
}
