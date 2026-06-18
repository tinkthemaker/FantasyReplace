package main

// html.go - wizardify for HTML sources: transforms text nodes only,
// leaving tags, attributes, and code-like content untouched.

import (
	"math/rand"
	"strings"

	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

var skipTags = map[string]bool{
	"script": true, "style": true, "code": true, "pre": true,
	"textarea": true, "kbd": true, "samp": true, "var": true,
}

func walkHTML(n *html.Node, fn func(string) string) {
	if n.Type == html.ElementNode && skipTags[n.Data] {
		return
	}
	if n.Type == html.TextNode {
		n.Data = fn(n.Data)
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		walkHTML(c, fn)
	}
}

// WizardifyHTML transmutes the text content of an HTML document or
// fragment. Full documents (containing <html>) keep their structure;
// anything else is treated as a body fragment.
func WizardifyHTML(src string, lex *Lexicon, intensity int, seed int64) (string, error) {
	rng := rand.New(rand.NewSource(seed))
	rs := buildRuleset(lex, intensity)
	transform := func(s string) string {
		if strings.TrimSpace(s) == "" {
			return s
		}
		return transformText(s, rs, lex, intensity, rng)
	}

	if strings.Contains(strings.ToLower(src), "<html") {
		doc, err := html.Parse(strings.NewReader(src))
		if err != nil {
			return "", err
		}
		walkHTML(doc, transform)
		var b strings.Builder
		if err := html.Render(&b, doc); err != nil {
			return "", err
		}
		return b.String(), nil
	}

	ctx := &html.Node{Type: html.ElementNode, Data: "body", DataAtom: atom.Body}
	nodes, err := html.ParseFragment(strings.NewReader(src), ctx)
	if err != nil {
		return "", err
	}
	var b strings.Builder
	for _, n := range nodes {
		walkHTML(n, transform)
		if err := html.Render(&b, n); err != nil {
			return "", err
		}
	}
	return b.String(), nil
}
