package main

// tone.go - dark-comedy flourishes via sentiment-keyed reversal
// (Pratchett-style bathos): gallows humor after misfortune, outsized
// pomp after small triumphs, deadpan parenthetical asides.

import (
	"math/rand"
	"regexp"
	"strings"
)

// keywords are matched against the ALREADY-WIZARDIFIED sentence
var misfortuneRe = regexp.MustCompile(`(?i)\b(curse|cursed|dark omen|unravel\w*|fail\w*|falter\w*|broke|sunder\w*|accursed|dead|died|perish\w*|wrong|amiss|stuck|wasted|woe|blunder|ghastly|dread|vexation|affliction|banish\w*|hour|hours|ruin\w*|lost|forfeit\w*)\b`)

var triumphRe = regexp.MustCompile(`(?i)\b(remedied|prevail\w*|triumph\w*|concluded|unveil\w*|flawless|shipped|finished|at long last|merry|gladness|wondrous|glorious|magnificent|succeeded|launch\w*|works now|alive)\b`)

var reSentenceSpan = regexp.MustCompile(`[^.!?\n]+[.!?]`)

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
