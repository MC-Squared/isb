package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

//Songbook represents a single songbook, including any options
//used when displaying/printing
type Songbook struct {
	FixedOrder    bool
	UseSection    bool
	IndexChorus   bool
	IndexPosition int
	Filename      string
	Title         string
	Songs         map[int]Song
}

//IndexAtStart returns true if this songbook should have an index at the start
func (sb Songbook) IndexAtStart() bool {
	return sb.IndexPosition == IndexStart
}

//IndexAtEnd returns true if this songbook should have an index at the end
func (sb Songbook) IndexAtEnd() bool {
	return sb.IndexPosition == IndexEnd
}

//NoIndex returns true if this songbook should not have an index
func (sb Songbook) NoIndex() bool {
	return sb.IndexPosition == IndexNone
}

const (
	//IndexNone indicates no index should be created when printing/displaying.
	//This is the default.
	IndexNone = 0
	//IndexStart indicates an index should be placed at the start when printing/displaying
	IndexStart = 1
	//IndexEnd indicates an index should be placed at the end when printing/displaying
	IndexEnd = 2
)

//ParseSongbookFile parses a songbook from filename, expecting
//any song files to be located in a songsRoot directory.
//Returns the created Songbook or any errors that occurred.
func ParseSongbookFile(filename string, songsRoot string) (*Songbook, error) {
	file, err := os.Open(filename)

	filename = filepath.Base(filename)
	title := filename[0 : len(filename)-len(".songlist")]

	if err != nil {
		return nil, err
	}
	defer file.Close()

	var (
		scanner       = bufio.NewScanner(file)
		fixedOrder    = false
		useSection    = false
		useChorus     = false
		indexPosition = IndexNone
		songs         = make(map[int]Song)
	)

	//bad_command_regex := regexp.MustCompile("\\{|\\}")

	for scanner.Scan() {
		line := scanner.Text()

		//is this a command
		if strings.HasPrefix(line, "{") {
			command := strings.ToLower(line)
			if strings.HasPrefix(command, "{title:") {
				title = parseCommand(line)
				continue
			} else if strings.HasPrefix(command, "{fixedOrder}") {
				fixedOrder = true
				continue
			} else if strings.HasPrefix(command, "{index_useSections}") {
				useSection = true
				continue
			} else if strings.HasPrefix(command, "{index_useChorus}") {
				useChorus = true
				continue
			} else if strings.HasPrefix(command, "{indexPositionition:") {
				p := parseCommand(line)

				switch p {
				case "start":
					indexPosition = IndexStart
					break
				case "end":
					indexPosition = IndexEnd
					break
				case "none":
				default:
					indexPosition = IndexNone
				}

				continue
			} else {
				fmt.Printf("Unknown tag: %s\n", line)
				continue
			}
		}

		line = strings.TrimSpace(line)

		//ignore blank lines
		if len(line) > 0 {
			num := -1
			//check for fixed numbering
			if strings.Index(line, ",") > 0 {
				numStr := line[0:strings.Index(line, ",")]
				num, err = strconv.Atoi(numStr)
				if err == nil {
					line = line[len(numStr)+1 : len(line)]
				} else {
					num = -1
				}
			}

			//including '.song' extension is optional
			if !strings.HasSuffix(line, ".song") {
				line += ".song"
			}

			song, err := ParseSongFile(songsRoot+"/"+line, 0)

			if err != nil {
				fmt.Println(num, ":", err)
			} else {
				if num < 0 {
					num = len(songs) + 1
				}
				song.SongNumber = num
				songs[num] = *song
			}
		}
	}

	return &Songbook{
			Title:         title,
			FixedOrder:    fixedOrder,
			Filename:      filename,
			UseSection:    useSection,
			IndexChorus:   useChorus,
			IndexPosition: indexPosition,
			Songs:         songs},
		nil
}

//GetSongOrder gets the song numbers of the Songs in this Songbook.
//Note that Song numbers may not be sequential.
func GetSongOrder(sb *Songbook) (keys []int) {
	keys = make([]int, len(sb.Songs))
	i := 0
	for k := range sb.Songs {
		keys[i] = k
		i++
	}
	sort.Sort(sort.IntSlice(keys))

	return keys
}

//GetSongSlice creates a []Song including all the Songs included in this
//Songbook, following the order from GetSongOrder.
func GetSongSlice(sb *Songbook) (songs []Song) {
	keys := GetSongOrder(sb)
	songs = make([]Song, len(sb.Songs))

	ind := 0
	for _, k := range keys {
		songs[ind] = sb.Songs[k]
		ind++
	}

	return songs
}

//Link provides a substring of this Songbook's Filename as a way to easily
//provide HTML links.
//i.e. if Filename is ".../song book.songlist" Link will return "song book"
func (sb Songbook) Link() string {
	return sb.Filename[0 : len(sb.Filename)-9]
}
