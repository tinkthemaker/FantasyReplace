package main

// pos.go - POS tagging via prose, with source offsets and a correction
// layer for tech vocabulary the perceptron mis-tags.

import (
	"strings"

	prose "github.com/jdkato/prose/v2"
)

type token struct {
	Text  string
	Tag   string
	Start int // byte offset into the segment
	End   int
}

// forceTags pins words the tagger gets wrong in tech prose.
var forceTags = map[string]string{
	"git": "NN", "repo": "NN", "config": "NN", "firmware": "NN",
	"wifi": "NN", "api": "NN", "cli": "NN", "pcb": "NN", "qmk": "NNP",
	"backend": "NN", "frontend": "NN", "linux": "NNP", "windows": "NNP",
}

// verbish words often mis-tagged NNS when they are 3rd-person verbs
// ("the server crashes").
var verbish = map[string]bool{
	"crash": true, "sync": true, "run": true, "work": true, "fail": true,
	"break": true, "build": true, "boot": true, "load": true, "compile": true,
	"deploy": true, "render": true, "print": true, "flash": true, "hold": true,
}

func tagSegment(text string) []token {
	doc, err := prose.NewDocument(text,
		prose.WithExtraction(false), prose.WithSegmentation(false))
	if err != nil {
		return nil
	}
	var toks []token
	cur := 0
	for _, t := range doc.Tokens() {
		idx := strings.Index(text[cur:], t.Text)
		if idx < 0 {
			continue // tokenizer normalized something; skip it
		}
		start := cur + idx
		toks = append(toks, token{Text: t.Text, Tag: t.Tag, Start: start, End: start + len(t.Text)})
		cur = start + len(t.Text)
	}
	correctTags(toks)
	return toks
}

func correctTags(toks []token) {
	for i := range toks {
		lower := strings.ToLower(toks[i].Text)
		if tag, ok := forceTags[lower]; ok {
			toks[i].Tag = tag
			continue
		}
		// "the server crashes" - NNS after a noun/pronoun, where the word
		// is a known verb, is really VBZ
		if toks[i].Tag == "NNS" && i > 0 && verbish[singularize(lower)] {
			prev := toks[i-1].Tag
			if prev == "NN" || prev == "NNP" || prev == "PRP" ||
				prev == "WDT" || prev == "WP" || prev == "IN" {
				toks[i].Tag = "VBZ"
			}
		}
		// "I build a lot" - noun tag right after a subject pronoun is a verb
		if (toks[i].Tag == "NN" || toks[i].Tag == "NNS") && i > 0 {
			prev := strings.ToLower(toks[i-1].Text)
			if toks[i-1].Tag == "PRP" && (prev == "i" || prev == "we" || prev == "they" || prev == "you") {
				if toks[i].Tag == "NNS" {
					toks[i].Tag = "VBZ"
				} else {
					toks[i].Tag = "VBP"
				}
			}
		}
		// sentence-lead "Today" etc. tagged NN is fine; leave it
	}
}

func isWordTag(tag string) bool {
	if tag == "" {
		return false
	}
	c := tag[0]
	return (c >= 'A' && c <= 'Z') && tag != "SYM"
}
