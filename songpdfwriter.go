package main

import (
	"bytes"
	"fmt"
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
	stanzaFont  = PDFFont{"Times", "", 12}
	chordFont   = PDFFont{"Helvetica", "B", stanzaFont.Size * 0.85}
	commentFont = PDFFont{"Times", "I", chordFont.Size}
	titleFont   = PDFFont{"Helvetica", "B", stanzaFont.Size * 1.5}
	sectionFont = commentFont
)
var stanzaIndent = stanzaFont.Size
var stanzaNumberIndent = stanzaIndent / 2.0
var chorusIndent = stanzaIndent * 2

//var songNumberFont = PDFFont{"Helvetica", "B", 15}
//var tocFont = PDFFont{"Times", "", 12}
//var tocSectionFont = PDFFont{"Times", "I", 12}

func WriteSongPDF(song *Song) (*bytes.Buffer, error) {
	//Set up PDF object
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()
	pdf.SetDisplayMode("fullpage", "TwoColumnLeft")
	pdf.SetTitle(song.Title, true)
	pdf.SetAuthor("Indigo Song Book", false)

	//Get font heights
	stanzaHt := pdf.PointConvert(stanzaFont.Size)
	chordHt := pdf.PointConvert(chordFont.Size)
	commentHt := pdf.PointConvert(commentFont.Size)
	titleHt := pdf.PointConvert(titleFont.Size)
	sectionHt := pdf.PointConvert(sectionFont.Size)

	tr := pdf.UnicodeTranslatorFromDescriptor("")
	width, height := pdf.GetPageSize()
	left, top, right, bot := pdf.GetMargins()
	//Subtract margins to get usable area
	height -= (top + bot)
	width -= (left + right)

	crrntCol := 0
	xMargin := left

	//Column handling
	//(Taken from fpdf_test.go:ExampleFpdf_SetLeftMargin())
	setCol := func(col int) {
		// Set position at a given column
		crrntCol = col
		xMargin = left + float64(col)*(width/2)
	}

	pdf.SetAcceptPageBreakFunc(func() bool {
		// Method accepting or not automatic page break
		if crrntCol < 1 {
			// Go to next column
			setCol(crrntCol + 1)
			// Set ordinate to top
			pdf.SetY(top)
			// Keep on page
			return false
		}
		// Go back to first column
		setCol(0)
		// Page break
		return true
	})

	//Print title
	setFont(pdf, titleFont)
	pdf.WriteAligned(0, titleHt, song.Title, "C")
	pdf.Ln(titleHt)

	//Print section
	if len(song.Section) > 0 {
		setFont(pdf, sectionFont)
		pdf.WriteAligned(0, sectionHt, song.Section, "C")
		pdf.Ln(sectionHt)
	}
	gotoSongStartY := func() {
		y := top + titleHt + stanzaHt
		if len(song.Section) > 0 {
			y += sectionHt
		}

		pdf.SetY(y)
	}

	gotoSongStartY()
	var w float64
	//Print pre-song comments
	printlnSlice(
		pdf,
		tr,
		song.BeforeComments,
		commentFont)
	pdf.Ln(commentHt)

	//Print stanzas
	for _, stanza := range song.Stanzas {
		if (pdf.GetY() + stanza.getHeight(pdf)) >= height {
			crrntCol++
			if crrntCol > 1 {
				crrntCol = 0
				pdf.AddPage()
			} else {

			}
			gotoSongStartY()
			xMargin = left + float64(crrntCol)*(width/2)
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
			last_pos := 0
			last_chord_w := 0.0
			last_chord_x := -1.0
			adjust_w := 0.0

			for _, chord := range line.Chords {

				//Put blank padding
				//to position chords correctly
				if chord.Position > 0 {
					setFont(pdf, stanzaFont)
					w = pdf.GetStringWidth(tr(line.Text[last_pos:chord.Position]))
					w -= last_chord_w
					w -= adjust_w
					pdf.Cell(w, chordHt, "")
				}

				//Prevent chords from crowding each other
				if last_chord_x+last_chord_w > pdf.GetX() {
					setFont(pdf, stanzaFont)
					w = (last_chord_x + pdf.GetStringWidth("-")) - pdf.GetX()
					adjust_w += w

					pdf.Cell(w, chordHt, "")
				}

				last_chord_w, _ = print(pdf, tr, chord.Text, chordFont)
				last_pos = chord.Position
				last_chord_x = pdf.GetX()
			}
			if stanza.HasChords() {
				pdf.Ln(chordHt)
			}

			setFont(pdf, stanzaFont)
			//Print stanza number on the first line
			if song.ShowStanzaNumbers &&
				ind == 0 &&
				stanza.ShowNumber &&
				!stanza.IsChorus {

				num := strconv.Itoa(stanza.Number)
				pdf.SetX(xMargin + stanzaNumberIndent)
				pdf.Cell(stanzaIndent-stanzaNumberIndent, stanzaHt, tr(num))
			}

			//Print echo
			if line.HasEcho() {
				if line.EchoIndex > 0 {
					str := line.Text[0:line.EchoIndex]
					w = pdf.GetStringWidth(str)
					pdf.Cell(w, stanzaHt, tr(str))
				}

				pdf.SetTextColor(128, 128, 128)
				str := line.Text[line.EchoIndex:len(line.Text)]

				fmt.Println(str)
				w = pdf.GetStringWidth(str)
				pdf.Cell(w, stanzaHt, tr(str))
				pdf.SetTextColor(0, 0, 0)
			} else {
				w = pdf.GetStringWidth(line.Text)
				pdf.Cell(w, stanzaHt, tr(line.Text))
			}
			pdf.Ln(stanzaHt)
		}

		//print post-stanza comments
		printlnSlice(
			pdf,
			tr,
			stanza.AfterComments,
			commentFont)
		pdf.Ln(stanzaHt)
	}

	setXAndMargin(pdf, xMargin+stanzaIndent)

	//Print post-song comments
	printlnSlice(
		pdf,
		tr,
		song.AfterComments,
		commentFont)

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

	h := pdf.PointConvert(font.Size)
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
	commentHt := pdf.PointConvert(commentFont.Size)
	stanzaHt := pdf.PointConvert(stanzaFont.Size)
	titleHt := pdf.PointConvert(titleFont.Size)
	sectionHt := pdf.PointConvert(sectionFont.Size)

	//title and section
	h := titleHt + stanzaHt
	if len(song.Section) > 0 {
		h += sectionHt
	}

	h += commentHt * (float64)(len(song.BeforeComments)+len(song.AfterComments))
	//before comments also have a blank line after
	if song.HasBeforeComments() {
		h += commentHt
	}

	for _, stanza := range song.Stanzas {
		h += stanza.getHeight(pdf)

		//blank line between stanzas
		h += stanzaHt
	}

	return h
}

func (stanza Stanza) getHeight(pdf *gofpdf.Fpdf) float64 {
	//get font heights
	commentHt := pdf.PointConvert(commentFont.Size)
	chordHt := pdf.PointConvert(chordFont.Size)
	stanzaHt := pdf.PointConvert(stanzaFont.Size)

	h := commentHt * (float64)(len(stanza.BeforeComments)+len(stanza.AfterComments))
	if stanza.HasChords() {
		h += chordHt * (float64)(len(stanza.Lines))
	}
	h += stanzaHt * (float64)(len(stanza.Lines))

	return h
}
