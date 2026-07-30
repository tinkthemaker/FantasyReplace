package main

// tone.go - dark-comedy flourishes via sentiment-keyed reversal
// (Pratchett-style bathos): gallows humor after misfortune, outsized
// pomp after small triumphs, deadpan parenthetical asides.

import (
	"math/rand"
	"regexp"
	"strings"
)

// keywords are matched against the ALREADY-WIZARDIFIED sentence.
// Pruned of words that over-fire on mundane corporate/tech prose
// ("unveil", "launch", "concluded" are routine product-announcement
// verbs, not triumphs; "hour/hours/stuck/wrong" are too common to
// signal genuine misfortune).
var misfortuneRe = regexp.MustCompile(`(?i)\b(curse|cursed|dark omen|unravel\w*|fail\w*|falter\w*|broke|sunder\w*|accursed|dead|died|perish\w*|amiss|wasted|woe|blunder|ghastly|dread|vexation|affliction|banish\w*|ruin\w*|lost|forfeit\w*)\b`)

var triumphRe = regexp.MustCompile(`(?i)\b(remedied|prevail\w*|triumph\w*|flawless|shipped|finished|at long last|merry|gladness|wondrous|glorious|magnificent|succeeded|works now|alive)\b`)

// reSentenceSpan matches a sentence: a run of non-terminator, non-newline
// characters ending in a sentence terminator. (\n is a regex newline escape
// inside the class, NOT a literal backslash-n — using a single backslash.)
var reSentenceSpan = regexp.MustCompile(`[^.!?\n]+[.!?]`)

// spec-sentence signals: a dry product/technical spec (numbers, acronyms,
// hyphenated compounds, percentages) is not a narrative beat and should not
// receive a flourish no matter what keyword it trips.
var reHasDigit = regexp.MustCompile(`\d`)
var reHasAcronym = regexp.MustCompile(`\b[A-Z]{2,}\b`)
var reHasHyphenCompound = regexp.MustCompile(`[A-Za-z]+-[A-Za-z]+-[A-Za-z]+`)

// isSpecSentence reports whether a sentence looks like a dry technical spec
// rather than a narrative beat. Flourishes read as non-sequiturs on specs
// ("550B-parameter effigy. Songs will be written."), so we suppress them.
func isSpecSentence(s string) bool {
	n := 0
	if reHasDigit.MatchString(s) {
		n++
	}
	if reHasAcronym.MatchString(s) {
		n++
	}
	if strings.Contains(s, "%") {
		n++
	}
	if reHasHyphenCompound.MatchString(s) {
		n++
	}
	return n >= 2
}

// addToneFlourishes appends mood-matched exclamations after sentences
// and occasionally slips a deadpan aside before the final period.
func addToneFlourishes(text string, lex *Lexicon, rng *rand.Rand) string {
	return reSentenceSpan.ReplaceAllStringFunc(text, func(s string) string {
		// skip headings, list markers, table rows, placeholders
		trimmed := strings.TrimSpace(s)
		if strings.HasPrefix(trimmed, "#") || strings.HasPrefix(trimmed, "|") ||
			strings.Contains(s, "\x00") {
			return s
		}
		// suppress flourishes on dry technical specs
		if isSpecSentence(s) {
			return s
		}
		dark := misfortuneRe.MatchString(s)
		bright := !dark && triumphRe.MatchString(s)

		// deadpan aside, tucked inside the sentence: "... a Saturday (do not ask what it cost)."
		if dark && len(lex.Asides) > 0 && rng.Float64() < 0.15 {
			body, punct := s[:len(s)-1], s[len(s)-1:]
			return body + " " + lex.Asides[rng.Intn(len(lex.Asides))] + punct
		}

		var pool []string
		var rate float64
		switch {
		case dark && len(lex.Misfortunes) > 0:
			pool, rate = lex.Misfortunes, 0.35
		case bright && len(lex.Triumphs) > 0:
			pool, rate = lex.Triumphs, 0.30
		case len(lex.Exclamations) > 0:
			pool, rate = lex.Exclamations, 0.06
		}
		if pool != nil && rng.Float64() < rate {
			return s + " " + pool[rng.Intn(len(pool))]
		}
		return s
	})
}
