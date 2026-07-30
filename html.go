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

var newsProtectedTags = map[string]bool{
	"a": true, "blockquote": true, "q": true, "cite": true, "time": true,
	"footer": true, "h1": true, "h2": true, "h3": true,
	"h4": true, "h5": true, "h6": true,
}

var blockTags = map[string]bool{
	"body": true, "article": true, "section": true, "main": true, "aside": true,
	"p": true, "li": true, "blockquote": true, "figcaption": true,
	"dt": true, "dd": true, "td": true, "th": true,
	"h1": true, "h2": true, "h3": true, "h4": true, "h5": true, "h6": true,
}

func walkHTML(n *html.Node, fn func(string) string, newsSafe bool) {
	if n.Type == html.ElementNode && (skipTags[n.Data] || newsSafe && newsProtectedTags[n.Data]) {
		return
	}
	if n.Type == html.ElementNode && blockTags[n.Data] {
		transformInlineText(n, fn)
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			if c.Type == html.ElementNode && blockTags[c.Data] {
				walkHTML(c, fn, newsSafe)
			} else {
				walkNestedBlocks(c, fn, newsSafe)
			}
		}
		return
	}
	if n.Type == html.TextNode {
		n.Data = fn(n.Data)
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		walkHTML(c, fn, newsSafe)
	}
}

func walkNestedBlocks(n *html.Node, fn func(string) string, newsSafe bool) {
	if n.Type == html.ElementNode && (skipTags[n.Data] || newsSafe && newsProtectedTags[n.Data] || blockTags[n.Data]) {
		if blockTags[n.Data] && !skipTags[n.Data] && !(newsSafe && newsProtectedTags[n.Data]) {
			walkHTML(n, fn, newsSafe)
		}
		return
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		walkNestedBlocks(c, fn, newsSafe)
	}
}

// transformInlineText gives the tagger a whole block of prose instead of
// isolated text nodes. Stable sentinels let us put transformed text back into
// the original inline DOM nodes without changing tags or attributes.
func transformInlineText(root *html.Node, fn func(string) string) {
	var nodes []*html.Node
	var collect func(*html.Node)
	collect = func(n *html.Node) {
		if n != root && n.Type == html.ElementNode && (skipTags[n.Data] || blockTags[n.Data]) {
			return
		}
		if n.Type == html.TextNode {
			nodes = append(nodes, n)
			return
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			collect(c)
		}
	}
	collect(root)
	if len(nodes) == 0 {
		return
	}

	// Surround the sentinel with disposable spaces so the tokenizer cannot
	// merge it with a word at an inline-element boundary.
	marker := " \x00§\x00 "
	for strings.Contains(joinedText(nodes), marker) {
		marker = "§" + marker + "§"
	}
	markers := make([]string, len(nodes)-1)
	var joined strings.Builder
	for i, node := range nodes {
		joined.WriteString(node.Data)
		if i < len(markers) {
			markers[i] = marker
			joined.WriteString(markers[i])
		}
	}
	transformed := fn(joined.String())
	parts := []string{transformed}
	for _, marker := range markers {
		last := parts[len(parts)-1]
		at := strings.Index(last, marker)
		if at < 0 {
			// Conservative fallback: retain the original DOM rather than risk
			// assigning prose to the wrong inline element.
			return
		}
		parts[len(parts)-1] = last[:at]
		parts = append(parts, last[at+len(marker):])
	}
	if len(parts) != len(nodes) {
		return
	}
	for i, node := range nodes {
		node.Data = parts[i]
	}
}

func joinedText(nodes []*html.Node) string {
	var b strings.Builder
	for _, node := range nodes {
		b.WriteString(node.Data)
	}
	return b.String()
}

// WizardifyHTML transmutes the text content of an HTML document or
// fragment. Full documents (containing <html>) keep their structure;
// anything else is treated as a body fragment.
func WizardifyHTML(src string, lex *Lexicon, intensity int, seed int64) (string, error) {
	w, err := NewWizardifier(lex)
	if err != nil {
		return "", err
	}
	return w.HTML(src, DefaultTransformOptions(intensity, seed))
}

func (w *Wizardifier) HTML(src string, opts TransformOptions) (string, error) {
	rs, err := w.rulesetFor(opts)
	if err != nil {
		return "", err
	}
	rng := rand.New(rand.NewSource(opts.Seed))
	flourishSeed := opts.FlourishSeed
	if flourishSeed == 0 {
		flourishSeed = opts.Seed
	}
	flourishRNG := rand.New(rand.NewSource(flourishSeed))
	transform := func(s string) string {
		if strings.TrimSpace(s) == "" {
			return s
		}
		if !opts.NewsSafe {
			return transformText(s, rs, w.lex, opts, rng, flourishRNG)
		}
		var stash []string
		protected := protectNewsFacts(s, &stash)
		result := restoreStash(transformText(protected, rs, w.lex, opts, rng, flourishRNG), stash)
		if opts.NewsFlair > 0 {
			result = applyNewsFlair(result, opts.Seed, opts.NewsFlair)
		}
		return result
	}

	if strings.Contains(strings.ToLower(src), "<html") {
		doc, err := html.Parse(strings.NewReader(src))
		if err != nil {
			return "", err
		}
		walkHTML(doc, transform, opts.NewsSafe)
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
		walkHTML(n, transform, opts.NewsSafe)
		if err := html.Render(&b, n); err != nil {
			return "", err
		}
	}
	return b.String(), nil
}
