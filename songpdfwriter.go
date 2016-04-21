package main

import (
    "github.com/jung-kurt/gofpdf"
    "log"
)

type PDFFont struct {
    Family string
    Style string
    Size float64
}

var stanzaFont = PDFFont{"Times", "", 12}
var chordFont = PDFFont{"Helvetica", "B", 10}
var commentFont = PDFFont{"Times", "I", 10}

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
    pdf.Ln(commentHt);

    for _, stanza := range song.Stanzas {
        if stanza.IsChorus {
            pdf.SetLeftMargin(20)
            pdf.SetX(20)
        } else {
            pdf.SetLeftMargin(10)
            pdf.SetX(10)
        }

        //Print pre-stanza comments
        printComments(
            pdf,
            tr,
            stanza.BeforeComments,
            commentFont)

        for _, line := range stanza.Lines {
            setFont(pdf, chordFont)
            //blank space before first chord (if any)
            if line.HasChords() {
                w = pdf.GetStringWidth(line.PreChordText(line.Chords[0]))
                if w > 0 {
                    pdf.Cell(w, chordHt, "")
                }
            }
            for _, chord := range line.Chords {
                w = pdf.GetStringWidth(chord.Text) + pdf.GetStringWidth(line.PreChordText(chord))
                pdf.Cell(w, chordHt, tr(chord.Text))
            }
            if stanza.HasChords() {
                pdf.Ln(chordHt)
            }

            setFont(pdf, stanzaFont)
            w = pdf.GetStringWidth(line.Text)
            pdf.Cell(w, stanzaHt, tr(line.Text))
            //y += 10
            pdf.Ln(stanzaHt);
        }

        //print post-stanza comments
        printComments(
            pdf,
            tr,
            stanza.AfterComments,
            commentFont)
        pdf.Ln(stanzaHt);
    }

    err := pdf.OutputFileAndClose("hello.pdf")

    if err != nil {
        log.Println(err)
    }
}

func setFont(pdf *gofpdf.Fpdf, font PDFFont) {
    pdf.SetFont(font.Family, font.Style, font.Size)
}

func printComments(pdf *gofpdf.Fpdf, tr func(string) string, comments []string, font PDFFont) {
    commentHt := pdf.PointConvert(font.Size)
    setFont(pdf, font)
    
    for _, comment := range comments {
        w := pdf.GetStringWidth(comment)
        pdf.Cell(w, commentHt, tr(comment))
        pdf.Ln(commentHt);
    }
}