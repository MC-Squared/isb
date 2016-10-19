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

var (
	stanzaFont     = PDFFont{"Times", "", 12}
	songNumberFont = PDFFont{"Helvetica", "B", stanzaFont.Size * 1.5}
	chordFont      = PDFFont{"Helvetica", "B", stanzaFont.Size * 0.85}
	commentFont    = PDFFont{"Times", "I", chordFont.Size}
	titleFont      = PDFFont{"Helvetica", "B", stanzaFont.Size * 1.5}
	sectionFont    = commentFont
	//tocFont = stanzaFont
	indexFont = PDFFont{"Helvetica", "", stanzaFont.Size * 1.25}
)
var stanzaIndent = stanzaFont.Size * 0.75
var stanzaNumberIndent = stanzaIndent / 2.0
var chorusIndent = stanzaIndent * 2

func WriteBookPDF(sbook *Songbook) (*bytes.Buffer, error) {
	//Set up PDF object
	pdf := initPDF(sbook.Title)
	tr := pdf.UnicodeTranslatorFromDescriptor("")

	//Print title
	setFont(pdf, titleFont)
	pdf.WriteAligned(0, titleFont.Height(pdf), sbook.Title, "C")
	pdf.Ln(titleFont.Height(pdf) + stanzaFont.Height(pdf))

	//Print index
	songLinks := make(map[int]int, 0)
	setFont(pdf, indexFont)
	pdf.SetTextColor(0, 0, 255)
	lineHt := indexFont.Height(pdf) * 1.5
	for _, song := range GetSongSlice(sbook) {
		link := pdf.AddLink()
		pdf.SetFont("", "U", 0)

		pdf.WriteLinkID(lineHt, strconv.Itoa(song.SongNumber), link)
		pdf.SetFont("", "", 0)
		pdf.Write(lineHt, "   ")
		songLinks[song.SongNumber] = link
	}
	pdf.SetTextColor(0, 0, 0)
	newPage(pdf)
	for _, song := range GetSongSlice(sbook) {
		y := pdf.GetY()
		//two-column songs must start on col 0
		if song.getHeight(pdf) > height && y > getSongStartY(pdf, false) {
			newPage(pdf)
		} else if y+song.getHeight(pdf) > height {
			nextCol(pdf)
		}

		pdf.SetLink(songLinks[song.SongNumber], 0, -1)
		setXAndMargin(pdf, xMargin)
		println(pdf, tr, strconv.Itoa(song.SongNumber), songNumberFont)
		y = printSong(pdf, &song)
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

	leftMargin, _, _, _ = pdf.GetMargins()
	xMargin = leftMargin

	width, height = pdf.GetPageSize()
	_, top, right, bot := pdf.GetMargins()
	//Subtract margins to get usable area
	height -= (top + bot)
	width -= (leftMargin + right)

	crrntCol = 0

	return pdf
}

var (
	crrntCol   = 0
	xMargin    = 0.0
	leftMargin = 0.0
	width      = 0.0
	height     = 0.0
)

func printSong(pdf *gofpdf.Fpdf, song *Song) float64 {
	tr := pdf.UnicodeTranslatorFromDescriptor("")

	//Print pre-song comments
	printlnSlice(
		pdf,
		tr,
		song.BeforeComments,
		commentFont)
	pdf.Ln(commentFont.Height(pdf))

	//Print stanzas
	for _, stanza := range song.Stanzas {
		if (pdf.GetY() + stanza.getHeight(pdf)) >= height {
			nextCol(pdf)
		}

		setXAndMargin(pdf, xMargin+stanzaIndent)

		if stanza.IsChorus {
			setXAndMargin(pdf, xMargin+chorusIndent)
		}

		//Print pre-stanza comments
		printlnSlice(
			pdf,
			tr,
			stanza.BeforeComments,
			commentFont)

		//Print stanza lines
		for ind, line := range stanza.Lines {

			//List of lines to print, may grow if the
			//line contents are too long to fit in a column
			to_print := make([]Line, 0)
			to_print = append(to_print, line)
			//w := pdf.GetStringWidth(line.Text)
			print_stanza_number := song.ShowStanzaNumbers && ind == 0 && stanza.ShowNumber && !stanza.IsChorus

			//check width
			for len(to_print) > 0 {
				l := to_print[0]
				to_print = to_print[1:]

				if pdf.GetStringWidth(l.Text) > (width / 2) {
					new_lines := l.SplitLine()
					to_print = append(new_lines, to_print...)
				} else {
					printLine(pdf, tr, stanza, l, print_stanza_number)
					print_stanza_number = false
				}
			}
		}

		//print post-stanza comments
		printlnSlice(
			pdf,
			tr,
			stanza.AfterComments,
			commentFont)
		pdf.Ln(stanzaFont.Height(pdf))
	}

	setXAndMargin(pdf, xMargin+stanzaIndent)

	//Print post-song comments
	printlnSlice(
		pdf,
		tr,
		song.AfterComments,
		commentFont)

	setXAndMargin(pdf, xMargin)

	return pdf.GetY()
}

func printLine(pdf *gofpdf.Fpdf, tr func(string) string, stanza Stanza, line Line, stanza_number bool) {
	last_pos := 0
	last_chord_w := 0.0
	last_chord_x := -1.0
	adjust_w := 0.0
	var w float64

	for _, chord := range line.Chords {

		//Put blank padding
		//to position chords correctly
		if chord.Position > 0 {
			setFont(pdf, stanzaFont)
			w = pdf.GetStringWidth(tr(line.Text[last_pos:chord.Position]))
			w -= last_chord_w
			w -= adjust_w
			pdf.Cell(w, chordFont.Height(pdf), "")
		}

		//Prevent chords from crowding each other
		if last_chord_x+last_chord_w > pdf.GetX() {
			setFont(pdf, stanzaFont)
			w = (last_chord_x + pdf.GetStringWidth("-")) - pdf.GetX()
			adjust_w += w

			pdf.Cell(w, chordFont.Height(pdf), "")
		}

		last_chord_w, _ = print(pdf, tr, chord.Text, chordFont)
		last_pos = chord.Position
		last_chord_x = pdf.GetX()
	}
	if stanza.HasChords() {
		pdf.Ln(chordFont.Height(pdf))
	}

	setFont(pdf, stanzaFont)
	//Print stanza number on the first line
	if stanza_number {
		num := strconv.Itoa(stanza.Number)
		pdf.SetX(xMargin + stanzaNumberIndent)
		pdf.Cell(stanzaIndent-stanzaNumberIndent, stanzaFont.Height(pdf), tr(num))
	}

	//Print echo
	if line.HasEcho() {
		if line.EchoIndex > 0 {
			str := line.Text[0:line.EchoIndex]
			w = pdf.GetStringWidth(str)
			pdf.Cell(w, stanzaFont.Height(pdf), tr(str))
		}

		pdf.SetTextColor(128, 128, 128)
		str := line.Text[line.EchoIndex:len(line.Text)]

		w = pdf.GetStringWidth(str)
		pdf.Cell(w, stanzaFont.Height(pdf), tr(str))
		pdf.SetTextColor(0, 0, 0)
	} else {
		w = pdf.GetStringWidth(line.Text)
		pdf.Cell(w, stanzaFont.Height(pdf), tr(line.Text))
	}

	pdf.Ln(stanzaFont.Height(pdf))
}

func WriteSongPDF(song *Song) (*bytes.Buffer, error) {
	//Set up PDF object
	pdf := initPDF(song.Title)

	//Print title
	setFont(pdf, titleFont)
	pdf.WriteAligned(0, titleFont.Height(pdf), song.Title, "C")
	pdf.Ln(titleFont.Height(pdf))

	//Print section
	if len(song.Section) > 0 {
		setFont(pdf, sectionFont)
		pdf.WriteAligned(0, sectionFont.Height(pdf), song.Section, "C")
		pdf.Ln(sectionFont.Height(pdf))
	}

	printSong(pdf, song)

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

func (song Song) getHeight(pdf *gofpdf.Fpdf) float64 {
	//title and section
	h := titleFont.Height(pdf) + stanzaFont.Height(pdf)
	if len(song.Section) > 0 {
		h += sectionFont.Height(pdf)
	}

	h += commentFont.Height(pdf) * (float64)(len(song.BeforeComments)+len(song.AfterComments))
	//before comments also have a blank line after
	if song.HasBeforeComments() {
		h += commentFont.Height(pdf)
	}

	for _, stanza := range song.Stanzas {
		h += stanza.getHeight(pdf)

		//blank line between stanzas
		h += stanzaFont.Height(pdf)
	}

	return h
}

func (stanza Stanza) getHeight(pdf *gofpdf.Fpdf) float64 {
	h := commentFont.Height(pdf) * (float64)(len(stanza.BeforeComments)+len(stanza.AfterComments))
	if stanza.HasChords() {
		h += chordFont.Height(pdf) * (float64)(len(stanza.Lines))
	}
	h += stanzaFont.Height(pdf) * (float64)(len(stanza.Lines))

	return h
}

func (font PDFFont) Height(pdf *gofpdf.Fpdf) float64 {
	return pdf.PointConvert(font.Size)
}

func gotoSongStartY(pdf *gofpdf.Fpdf, section bool) {
	pdf.SetY(getSongStartY(pdf, section))
}

func getSongStartY(pdf *gofpdf.Fpdf, section bool) float64 {
	_, top, _, _ := pdf.GetMargins()

	y := top + titleFont.Height(pdf) + stanzaFont.Height(pdf)
	if section {
		y += sectionFont.Height(pdf)
	}

	return y
}

func nextCol(pdf *gofpdf.Fpdf) {
	crrntCol++
	if crrntCol > 1 {
		newPage(pdf)
	}

	updateXMargin()
	gotoSongStartY(pdf, false)
}

func newPage(pdf *gofpdf.Fpdf) {
	crrntCol = 0
	pdf.AddPage()
	updateXMargin()
	gotoSongStartY(pdf, false)
}

func updateXMargin() {
	xMargin = leftMargin + float64(crrntCol)*(width/2)
}
