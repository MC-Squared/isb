package main

import (
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
	stanzaFont  = PDFFont{"Times", "", 24}
	chordFont   = PDFFont{"Helvetica", "B", stanzaFont.Size * 0.85}
	commentFont = PDFFont{"Times", "I", chordFont.Size}
)

var stanzaIndent = stanzaFont.Size
var stanzaNumberIndent = stanzaIndent / 2.0
var chorusIndent = stanzaIndent * 2.0

//var echoFont = PDFFont{stanzaFont.Family, "", stanzaFont.Size} DARK GREY
//var songNumberFont = PDFFont{"Helvetica", "B", 15}
//var tocFont = PDFFont{"Times", "", 12}
//var tocSectionFont = PDFFont{"Times", "I", 12}

func WriteSongPDF(song *Song) {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()

	stanzaHt := pdf.PointConvert(stanzaFont.Size)
	chordHt := pdf.PointConvert(chordFont.Size)
	commentHt := pdf.PointConvert(commentFont.Size)

	tr := pdf.UnicodeTranslatorFromDescriptor("")

	var w float64
	//Print pre-song comments
	printComments(
		pdf,
		tr,
		song.BeforeComments,
		commentFont)
	pdf.Ln(commentHt)

	for _, stanza := range song.Stanzas {
		if stanza.IsChorus {
			pdf.SetLeftMargin(chorusIndent)
			pdf.SetX(chorusIndent)
		}

		//Print pre-stanza comments
		printComments(
			pdf,
			tr,
			stanza.BeforeComments,
			commentFont)

		for ind, line := range stanza.Lines {
			last_pos := 0
			last_chord_w := 0.0
			for _, chord := range line.Chords {

				//Put blank padding
				//to position chords correctly
				if chord.Position > 0 {
					setFont(pdf, stanzaFont)
					w = pdf.GetStringWidth(line.Text[last_pos:chord.Position])
					w -= last_chord_w
					pdf.Cell(w, chordHt, "")
				}

				setFont(pdf, chordFont)
				w = pdf.GetStringWidth((chord.Text))
				pdf.Cell(w, chordHt, tr(chord.Text))

				last_chord_w = w
				last_pos = chord.Position

			}
			if stanza.HasChords() {
				pdf.Ln(chordHt)
			}

			setFont(pdf, stanzaFont)
			//Print stanza number on the first line
			if ind == 0 && stanza.ShowNumber && !stanza.IsChorus {
				num := strconv.Itoa(stanza.Number)
				pdf.SetX(stanzaNumberIndent)
				pdf.Cell(stanzaNumberIndent, stanzaHt, tr(num))
			}

			w = pdf.GetStringWidth(line.Text)
			pdf.Cell(w, stanzaHt, tr(line.Text))
			pdf.Ln(stanzaHt)
		}

		//print post-stanza comments
		printComments(
			pdf,
			tr,
			stanza.AfterComments,
			commentFont)
		pdf.Ln(stanzaHt)

		pdf.SetLeftMargin(stanzaIndent)
		pdf.SetX(stanzaIndent)
	}

	//Print post-song comments
	printComments(
		pdf,
		tr,
		song.AfterComments,
		commentFont)
	pdf.Ln(commentHt)

	err := pdf.OutputFileAndClose("hello.pdf")

	if err != nil {
		log.Println(err)
	}
}

func setFont(pdf *gofpdf.Fpdf, font PDFFont) {
	pdf.SetFont(font.Family, font.Style, font.Size)
}

func print(pdf *gofpdf.Fpdf, tr func(string) string, str string, font PDFFont) {
	setFont(pdf, font)

	ht := pdf.PointConvert(font.Size)
	w := pdf.GetStringWidth(str)
	pdf.Cell(w, ht, tr(str))
}

func printComments(pdf *gofpdf.Fpdf, tr func(string) string, comments []string, font PDFFont) {
	commentHt := pdf.PointConvert(font.Size)
	setFont(pdf, font)

	for _, comment := range comments {
		w := pdf.GetStringWidth(comment)
		pdf.Cell(w, commentHt, tr(comment))
		pdf.Ln(commentHt)
	}
}
