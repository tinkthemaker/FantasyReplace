package main

// archaic.go - generic archaization: works on ANY verb or pronoun,
// no lexicon entry needed. Intensity 3 only.

import "strings"

// auxiliaries conjugated for "thou"
var thouAux = map[string]string{
	"are": "art", "were": "wert", "have": "hast", "do": "dost",
	"did": "didst", "will": "wilt", "would": "wouldst", "can": "canst",
	"could": "couldst", "shall": "shalt", "should": "shouldst",
	"may": "mayst", "might": "mightst", "must": "must",
}

var estSpecial = map[string]string{
	"edit": "editest", // two syllables; the final consonant is not doubled
}

var ethSpecial = map[string]string{
	"has": "hath", "does": "doth", "says": "saith", "is": "is", "was": "was",
}

// ethify converts a lemma to its -eth form: sync -> synceth, run -> runneth.
func ethify(lemma string) string {
	switch {
	case strings.HasSuffix(lemma, "e"):
		return lemma + "th"
	case strings.HasSuffix(lemma, "y") && len(lemma) > 1 &&
		strings.IndexByte(consonants, lemma[len(lemma)-2]) >= 0:
		return lemma[:len(lemma)-1] + "ieth"
	case shouldDouble(lemma):
		return lemma + string(lemma[len(lemma)-1]) + "eth"
	default:
		return lemma + "eth"
	}
}

// ethifyPhrase converts the first word of a VBZ-form phrase to -eth:
// "unravels" -> "unravelleth" (well, "unraveleth"), "weaves its magic" -> "weaveth its magic".
func ethifyPhrase(phrase string) string {
	words := strings.Fields(phrase)
	if len(words) == 0 {
		return phrase
	}
	w := strings.ToLower(words[0])
	if special, ok := ethSpecial[w]; ok {
		words[0] = special
	} else {
		words[0] = ethify(singularize(w))
	}
	return strings.Join(words, " ")
}

// estify conjugates a verb for "thou": edit -> editest, have -> hast.
func estify(verb string) string {
	if aux, ok := thouAux[verb]; ok {
		return aux
	}
	if special, ok := estSpecial[verb]; ok {
		return special
	}
	switch {
	case strings.HasSuffix(verb, "e"):
		return verb + "st"
	case strings.HasSuffix(verb, "y") && len(verb) > 1 &&
		strings.IndexByte(consonants, verb[len(verb)-2]) >= 0:
		return verb[:len(verb)-1] + "iest"
	case shouldDouble(verb):
		return verb + string(verb[len(verb)-1]) + "est"
	default:
		return verb + "est"
	}
}

var auxSet = map[string]bool{
	"are": true, "were": true, "have": true, "do": true, "did": true,
	"will": true, "would": true, "can": true, "could": true,
	"shall": true, "should": true, "may": true, "might": true, "must": true,
}

var thouContractions = map[string]string{
	"'ll": " wilt", "'re": " art", "'ve": " hast", "'d": " wouldst",
}

// whole tokens, for when the tokenizer doesn't split the contraction
var youContractions = map[string]string{
	"you'll": "thou wilt", "you're": "thou art",
	"you've": "thou hast", "you'd": "thou wouldst",
}

func isAlphabetic(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if !(r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z') {
			return false
		}
	}
	return true
}

// isTitleCaseContext returns true if the token at index i is part of a
// title-case sequence — the token itself is capitalized AND at least one
// neighboring word token is also capitalized (excluding "I"). This detects
// headings and proper-noun phrases in body text, preventing -eth on words
// like "Dominates" in "Google Dominates the Market".
func isTitleCaseContext(toks []token, i int) bool {
	if len(toks[i].Text) == 0 {
		return false
	}
	self := toks[i].Text[0]
	if !(self >= 'A' && self <= 'Z') {
		return false // not capitalized, can't be title-case
	}
	if toks[i].Text == "I" {
		return false // "I" is always capitalized
	}
	capNeighbor := func(j int) bool {
		if j < 0 || j >= len(toks) {
			return false
		}
		t := toks[j].Text
		return len(t) > 0 && t[0] >= 'A' && t[0] <= 'Z' && t != "I"
	}
	return capNeighbor(prevWord(toks, i)) || capNeighbor(nextWord(toks, i))
}

// archaizeToken decides the archaic form for token i, using context.
// Returns "" if unchanged.
func archaizeToken(toks []token, i int) string {
	t := toks[i]
	lower := strings.ToLower(t.Text)

	// contraction fragments: "you'll" -> "thou wilt"; otherwise hands off
	if !isAlphabetic(t.Text) {
		if full, ok := youContractions[lower]; ok {
			return full
		}
		if full, ok := thouContractions[lower]; ok {
			j := prevWord(toks, i)
			if j >= 0 && strings.ToLower(toks[j].Text) == "you" && youIsSubject(toks, j) {
				return full
			}
		}
		return ""
	}

	switch lower {
	case "said":
		if t.Tag == "VBD" {
			return "quoth"
		}
	case "you":
		if youIsSubject(toks, i) {
			return "thou"
		}
		return "thee"
	case "yourself":
		return "thyself"
	}

	if strings.HasPrefix(t.Tag, "VB") || t.Tag == "MD" {
		// aux before "you" in a question: "Do you want" -> "Dost thou want"
		if auxSet[lower] {
			n := nextWord(toks, i)
			if n >= 0 && strings.ToLower(toks[n].Text) == "you" && youIsSubject(toks, n) {
				return estify(lower)
			}
		}
		// inverted question: verb followed by subject-"you" ("What mean you")
		if n := nextWord(toks, i); n >= 0 && strings.ToLower(toks[n].Text) == "you" && youIsSubject(toks, n) {
			p := prevWord(toks, i)
			if p < 0 || strings.HasPrefix(toks[p].Tag, "W") {
				return estify(lower)
			}
		}
		// verb directly after subject-"you": give it -est ("you edit" -> "thou editest"),
		// unless the clause is a question whose aux already carries the -est
		j := prevWord(toks, i)
		for j >= 0 && toks[j].Tag == "RB" { // skip adverbs: "you just edit"
			j = prevWord(toks, j)
		}
		if j >= 0 && strings.ToLower(toks[j].Text) == "you" && youIsSubject(toks, j) {
			p := prevWord(toks, j)
			if p >= 0 && auxSet[strings.ToLower(toks[p].Text)] {
				return "" // "Dost thou want" - leave main verb in base form
			}
			return estify(lower)
		}
	}

	// any 3rd-person singular verb: -eth
	if t.Tag == "VBZ" {
		// Skip -eth in title-case context (proper nouns, headings):
		// "Google Dominates" should not become "Google Dominateth".
		if isTitleCaseContext(toks, i) {
			return ""
		}
		if special, ok := ethSpecial[lower]; ok {
			if special == lower {
				return ""
			}
			return special
		}
		return ethify(lemmatize(lower, "VBZ"))
	}
	return ""
}

func prevWord(toks []token, i int) int {
	for j := i - 1; j >= 0; j-- {
		if isWordTag(toks[j].Tag) {
			return j
		}
	}
	return -1
}

func nextWord(toks []token, i int) int {
	for j := i + 1; j < len(toks); j++ {
		if isWordTag(toks[j].Tag) {
			return j
		}
	}
	return -1
}

// youIsSubject: "you edit" / "Are you ready" -> subject (thou);
// "thank you" / "to you" -> object (thee).
func youIsSubject(toks []token, i int) bool {
	j := nextWord(toks, i)
	for j >= 0 && toks[j].Tag == "RB" {
		j = nextWord(toks, j)
	}
	if j >= 0 && (strings.HasPrefix(toks[j].Tag, "VB") || toks[j].Tag == "MD") {
		return true
	}
	// question form: aux before you at clause start ("Are you ready")
	p := prevWord(toks, i)
	if p >= 0 && auxSet[strings.ToLower(toks[p].Text)] {
		pp := prevWord(toks, p)
		if pp < 0 || strings.HasPrefix(toks[pp].Tag, "W") {
			return true
		}
	}
	// inverted question without aux: "What mean you?" -> subject
	if p >= 0 && strings.HasPrefix(toks[p].Tag, "VB") {
		pp := prevWord(toks, p)
		if pp < 0 || strings.HasPrefix(toks[pp].Tag, "W") {
			return true
		}
	}
	return false
}
