package main

type Chord struct {
    Text string
    Position int
}


var scales = map[string]int{
    "A": 0,
    "Bb": 1,
    "B": 2,
    "C": 3,
    "C#": 4,
    "D": 5,
    "D#": 6,
    "E": 7,
    "F": 8,
    "F#": 9,
    "G": 10,
    "G#": 11,
}

func transposeKey(key string, change int) string {
    if change == 0 {
        return key
    }

    //check first two letters for match
    var scale_ind = -1
    var ok = false
    if len(key) > 1 {
        scale_ind, ok = scales[key[0:2]]
        if !ok {
            scale_ind = -1
        }
    }

    //check for single key match
    if scale_ind < 0 && len(key) > 0 {
        scale_ind, ok = scales[key[0:1]]
        if !ok {
            scale_ind = -1
        }
    }

    if scale_ind < 0 {
        return key
    }

    scale_ind = (scale_ind + change) % (len(scales))
    for k := range scales {
        if scales[k] == scale_ind {
            return k
        }
    }

    return key
}
