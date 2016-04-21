package main

type Stanza struct {
    ShowNumber bool
    IsChorus bool
    Number int
    BeforeComments []string
    AfterComments []string
    Lines []Line
}

func (stanza Stanza) HasChords() bool {
    for _,l := range stanza.Lines {
        if l.HasChords() {
            return true
        }
    }

    return false
}