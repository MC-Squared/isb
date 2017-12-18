package main

import (
	"bytes"
	"log"
	"strconv"

	"github.com/jung-kurt/gofpdf"
)

type PDFFont struct {
	Family string
	Style  string
	Size   float64
}

type BookFonts struct {
	Stanza     PDFFont
	SongNumber PDFFont
	Chord      PDFFont
	Comment    PDFFont
	Title      PDFFont
	Section    PDFFont
	TOC        PDFFont
	Index      PDFFont

	StanzaIndent       float64
	StanzaNumberIndent float64
	ChorusIndent       float64
}

//Basic number for electronic and printed sizes
var eStanzaSize = 24.0
var pStanzaSize = 12.0

var electronicFonts = BookFonts{
	SongNumber: PDFFont{"Helvetica", "B", eStanzaSize * 1.5},
	Title:      PDFFont{"Helvetica", "B", eStanzaSize * 1.0},
	Index:      PDFFont{"Helvetica", "", eStanzaSize * 1.0},

	Stanza:  PDFFont{"Times", "", eStanzaSize},
	Chord:   PDFFont{"Helvetica", "B", eStanzaSize * 0.85},
	Comment: PDFFont{"Times", "I", eStanzaSize * 0.85},
	Section: PDFFont{"Times", "I", eStanzaSize * 0.85},
	TOC:     PDFFont{"Times", "", eStanzaSize},

	StanzaIndent:       eStanzaSize * 0.75,
	StanzaNumberIndent: (eStanzaSize * 0.75) / 2.0,
	ChorusIndent:       (eStanzaSize * 0.75) * 2,
}

var printFonts = BookFonts{
	Stanza:     PDFFont{"Times", "", pStanzaSize},
	SongNumber: PDFFont{"Helvetica", "B", pStanzaSize * 1.5},
	Chord:      PDFFont{"Helvetica", "B", pStanzaSize * 0.85},
	Comment:    PDFFont{"Times", "I", pStanzaSize * 0.85},
	Title:      PDFFont{"Helvetica", "B", pStanzaSize * 1.5},
	Section:    PDFFont{"Times", "I", pStanzaSize * 0.85},
	TOC:        PDFFont{"Times", "", pStanzaSize},
	Index:      PDFFont{"Helvetica", "", pStanzaSize * 1.25},

	StanzaIndent:       pStanzaSize * 0.75,
	StanzaNumberIndent: (pStanzaSize * 0.75) / 2.0,
	ChorusIndent:       (pStanzaSize * 0.75) * 2,
}

//WriteBookPDF and WriteBookPDFElectronic are similar, the difference is that
//The electronic verion includes a hyper-linked index page
//And prints one song per-page
//Whereas the normal (print) version does not have a numeric index page
//and prints two-columns of songs per-page
func WriteBookPDFElectronic(sbook *Songbook) (*bytes.Buffer, error) {
	//Set up PDF object
	pdf := initPDF(sbook.Title)

	printTitle(pdf, sbook.Title, electronicFonts)

	//Print index
	songLinks := make(map[int]int, 0)
	setFont(pdf, electronicFonts.Index)
	pdf.SetTextColor(0, 0, 255)
	lineHt := electronicFonts.Index.Height(pdf) * 1.5

	for _, song := range GetSongSlice(sbook) {
		link := pdf.AddLink()
		pdf.SetFont("", "U", 0)

		pdf.WriteLinkID(lineHt, strconv.Itoa(song.SongNumber), link)
		pdf.SetFont("", "", 0)
		pdf.Write(lineHt, "   ")
		songLinks[song.SongNumber] = link
	}

	//Print the songs
	for _, song := range GetSongSlice(sbook) {
		newPage(pdf, electronicFonts)

		pdf.SetLink(songLinks[song.SongNumber], 0, -1)

		link := pdf.AddLink()
		setXAndMargin(pdf, (width - rightMargin - pdf.GetStringWidth("Index")))
		pdf.SetFont("", "U", 0)
		pdf.SetTextColor(0, 0, 255)
		pdf.WriteLinkID(lineHt, "Index", link)
		pdf.SetTextColor(0, 0, 0)

		_ = printSong(pdf, &song, electronicFonts, false)
	}

	buf := new(bytes.Buffer)
	err := pdf.Output(buf)

	if err != nil {
		log.Println(err)
	}

	return buf, nil
}

func WriteBookPDF(sbook *Songbook) (*bytes.Buffer, error) {
	//Set up PDF object
	pdf := initPDF(sbook.Title)

	printTitle(pdf, sbook.Title, printFonts)

	for _, song := range GetSongSlice(sbook) {
		y := pdf.GetY()

		//two-column songs must start on col 0
		if y > getSongStartY(pdf, false, printFonts) {
			if song.getHeight(pdf, printFonts) > height {
				newPage(pdf, printFonts)
			} else if crrntCol > 0 && y+song.getHeight(pdf, printFonts) > height {
				nextCol(pdf, printFonts)
			}
		}

		y = printSong(pdf, &song, printFonts, true)
	}

	buf := new(bytes.Buffer)
	err := pdf.Output(buf)

	if err != nil {
		log.Println(err)
	}

	return buf, nil
}

func initPDF(title string) *gofpdf.Fpdf {
	//Set up PDF object
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()
	pdf.SetDisplayMode("fullpage", "TwoColumnLeft")
	pdf.SetTitle(title, true)
	pdf.SetAuthor("Indigo Song Book", false)

	leftMargin, rightMargin, _, _ = pdf.GetMargins()
	xMargin = leftMargin

	width, height = pdf.GetPageSize()
	_, top, right, bot := pdf.GetMargins()
	//Subtract margins to get usable area
	height -= (top + bot)
	width -= (leftMargin + right)

	crrntCol = 0

	return pdf
}

func printTitle(pdf *gofpdf.Fpdf, title string, fonts BookFonts) {
	//Print title
	setFont(pdf, fonts.Title)
	pdf.WriteAligned(0, fonts.Title.Height(pdf), title, "C")
	pdf.Ln(fonts.Title.Height(pdf) + fonts.Stanza.Height(pdf))
}

func printSongNumber(pdf *gofpdf.Fpdf, tr func(string) string, number int, fonts BookFonts) {
	setXAndMargin(pdf, xMargin)
	println(pdf, tr, strconv.Itoa(number), fonts.SongNumber)
}

var (
	crrntCol    = 0
	xMargin     = 0.0
	leftMargin  = 0.0
	rightMargin = 0.0
	width       = 0.0
	height      = 0.0
)

func printSong(pdf *gofpdf.Fpdf, song *Song, fonts BookFonts, two_columns bool) float64 {
	tr := pdf.UnicodeTranslatorFromDescriptor("")

	//flag so we don't print song number on the wrong page/column
	var song_started = false

	//Print stanzas
	for _, stanza := range song.Stanzas {
		if (pdf.GetY() + stanza.getHeight(pdf, fonts)) >= height {
			if two_columns {
				nextCol(pdf, fonts)
			} else {
				newPage(pdf, fonts)
			}
		}

		if !song_started {
			printSongNumber(pdf, tr, song.SongNumber, fonts)

			//Print pre-song comments
			printlnSlice(
				pdf,
				tr,
				song.BeforeComments,
				fonts.Comment)
			pdf.Ln(fonts.Comment.Height(pdf))

			song_started = true
		}

		setXAndMargin(pdf, xMargin+fonts.StanzaIndent)

		if stanza.IsChorus {
			setXAndMargin(pdf, xMargin+fonts.ChorusIndent)
		}

		//Print pre-stanza comments
		printlnSlice(
			pdf,
			tr,
			stanza.BeforeComments,
			fonts.Comment)

		//Print stanza lines
		for ind, line := range stanza.Lines {

			//List of lines to print, may grow if the
			//line contents are too long to fit in a column
			to_print := make([]Line, 0)
			to_print = append(to_print, line)
			offset := fonts.StanzaIndent
			if stanza.IsChorus {
				offset = fonts.ChorusIndent
			}

			print_stanza_number := song.ShowStanzaNumbers && ind == 0 && stanza.ShowNumber && !stanza.IsChorus

			//check width
			for len(to_print) > 0 {
				l := to_print[0]
				to_print = to_print[1:]

				if false && (pdf.GetStringWidth(l.Text)+offset) > (width/2) {
					new_lines := l.SplitLine()
					to_print = append(new_lines, to_print...)
				} else {
					printLine(pdf, tr, stanza, l, print_stanza_number, fonts)
					print_stanza_number = false
				}
			}
		}

		//print post-stanza comments
		printlnSlice(
			pdf,
			tr,
			stanza.AfterComments,
			fonts.Comment)
		pdf.Ln(fonts.Stanza.Height(pdf))
	}

	setXAndMargin(pdf, xMargin+fonts.StanzaIndent)

	//Print post-song comments
	printlnSlice(
		pdf,
		tr,
		song.AfterComments,
		fonts.Comment)

	setXAndMargin(pdf, xMargin)

	return pdf.GetY()
}

func printLine(pdf *gofpdf.Fpdf, tr func(string) string, stanza Stanza, line Line, stanza_number bool, fonts BookFonts) {
	last_pos := 0
	last_chord_w := 0.0
	last_chord_x := -1.0
	adjust_w := 0.0
	var w float64

	for _, chord := range line.Chords {

		//Put blank padding
		//to position chords correctly
		if chord.Position > 0 {
			setFont(pdf, fonts.Stanza)
			w = pdf.GetStringWidth(tr(line.Text[last_pos:chord.Position]))
			w -= last_chord_w
			w -= adjust_w
			pdf.Cell(w, fonts.Chord.Height(pdf), "")
		}

		//Prevent chords from crowding each other
		if last_chord_x+last_chord_w > pdf.GetX() {
			setFont(pdf, fonts.Stanza)
			w = (last_chord_x + pdf.GetStringWidth("-")) - pdf.GetX()
			adjust_w += w

			pdf.Cell(w, fonts.Chord.Height(pdf), "")
		}

		last_chord_w, _ = print(pdf, tr, chord.GetText(), fonts.Chord)
		last_pos = chord.Position
		last_chord_x = pdf.GetX()
	}
	if stanza.HasChords() {
		pdf.Ln(fonts.Chord.Height(pdf))
	}

	setFont(pdf, fonts.Stanza)
	//Print stanza number on the first line
	if stanza_number {
		num := strconv.Itoa(stanza.Number)
		pdf.SetX(xMargin + fonts.StanzaNumberIndent)
		pdf.Cell(fonts.StanzaIndent-fonts.StanzaNumberIndent, fonts.Stanza.Height(pdf), tr(num))
	}

	//Print echo
	if line.HasEcho() {
		if line.EchoIndex > 0 {
			str := line.Text[0:line.EchoIndex]
			w = pdf.GetStringWidth(str)
			pdf.Cell(w, fonts.Stanza.Height(pdf), tr(str))
		}

		pdf.SetTextColor(128, 128, 128)
		str := line.Text[line.EchoIndex:len(line.Text)]

		w = pdf.GetStringWidth(str)
		pdf.Cell(w, fonts.Stanza.Height(pdf), tr(str))
		pdf.SetTextColor(0, 0, 0)
	} else {
		w = pdf.GetStringWidth(line.Text)
		pdf.Cell(w, fonts.Stanza.Height(pdf), tr(line.Text))
	}

	pdf.Ln(fonts.Stanza.Height(pdf))
}

func WriteSongPDF(song *Song) (*bytes.Buffer, error) {
	//Set up PDF object
	pdf := initPDF(song.Title)
	fonts := printFonts

	//Print title
	setFont(pdf, fonts.Title)
	pdf.WriteAligned(0, fonts.Title.Height(pdf), song.Title, "C")
	pdf.Ln(fonts.Title.Height(pdf))

	//Print section
	if len(song.Section) > 0 {
		setFont(pdf, fonts.Section)
		pdf.WriteAligned(0, fonts.Section.Height(pdf), song.Section, "C")
		pdf.Ln(fonts.Section.Height(pdf))
	}

	printSong(pdf, song, fonts, false)

	buf := new(bytes.Buffer)
	err := pdf.Output(buf)

	if err != nil {
		log.Println(err)
	}

	return buf, nil
}

func setXAndMargin(pdf *gofpdf.Fpdf, x float64) {
	pdf.SetX(x)
	pdf.SetLeftMargin(x)
}

func setFont(pdf *gofpdf.Fpdf, font PDFFont) {
	pdf.SetFont(font.Family, font.Style, font.Size)
}

func print(pdf *gofpdf.Fpdf, tr func(string) string, str string, font PDFFont) (width, height float64) {
	setFont(pdf, font)

	h := font.Height(pdf)
	w := pdf.GetStringWidth(str)
	pdf.Cell(w, h, tr(str))

	return w, h
}

func println(pdf *gofpdf.Fpdf, tr func(string) string, str string, font PDFFont) (width, height float64) {
	w, h := print(pdf, tr, str, font)
	pdf.Ln(h)
	return w, h
}

func printlnSlice(pdf *gofpdf.Fpdf, tr func(string) string, slice []string, font PDFFont) {
	for _, s := range slice {
		println(pdf, tr, s, font)
	}
}

func (song Song) getHeight(pdf *gofpdf.Fpdf, fonts BookFonts) float64 {
	//title and section
	h := fonts.Title.Height(pdf) + fonts.Stanza.Height(pdf)
	if len(song.Section) > 0 {
		h += fonts.Section.Height(pdf)
	}

	h += fonts.SongNumber.Height(pdf)
	h += fonts.Comment.Height(pdf) * (float64)(len(song.BeforeComments)+len(song.AfterComments))
	//before comments also have a blank line after
	if song.HasBeforeComments() {
		h += fonts.Comment.Height(pdf)
	}

	for _, stanza := range song.Stanzas {
		h += stanza.getHeight(pdf, fonts)

		//blank line between stanzas
		h += fonts.Stanza.Height(pdf)
	}

	return h
}

func (stanza Stanza) getHeight(pdf *gofpdf.Fpdf, fonts BookFonts) float64 {
	h := fonts.Comment.Height(pdf) * (float64)(len(stanza.BeforeComments)+len(stanza.AfterComments))
	if stanza.HasChords() {
		h += fonts.Chord.Height(pdf) * (float64)(len(stanza.Lines))
	}
	h += fonts.Stanza.Height(pdf) * (float64)(len(stanza.Lines)+1)

	return h
}

func (font PDFFont) Height(pdf *gofpdf.Fpdf) float64 {
	return pdf.PointConvert(font.Size)
}

func gotoSongStartY(pdf *gofpdf.Fpdf, section bool, fonts BookFonts) {
	pdf.SetY(getSongStartY(pdf, section, fonts))
}

func getSongStartY(pdf *gofpdf.Fpdf, section bool, fonts BookFonts) float64 {
	_, top, _, _ := pdf.GetMargins()

	y := top + fonts.Title.Height(pdf) + fonts.Stanza.Height(pdf)
	if section {
		y += fonts.Section.Height(pdf)
	}

	return y
}

func nextCol(pdf *gofpdf.Fpdf, fonts BookFonts) {
	crrntCol++
	if crrntCol > 1 {
		newPage(pdf, fonts)
	}

	updateXMargin()
	gotoSongStartY(pdf, false, fonts)
}

func newPage(pdf *gofpdf.Fpdf, fonts BookFonts) {
	crrntCol = 0
	pdf.AddPage()
	updateXMargin()
	gotoSongStartY(pdf, false, fonts)
}

func updateXMargin() {
	xMargin = leftMargin + float64(crrntCol)*(width/2)
}
