package main

// invert.go - occasional archaic sentence inversions (intensity 3).

import (
	"math/rand"
	"regexp"
	"strings"
)

var (
	reSpeechInv = regexp.MustCompile(`\b(I|he|she|we|they|He|She|We|They) (said|quoth|replied|thought|declared|exclaimed|wrote|reckoned)\b`)
	reNeverInv  = regexp.MustCompile(`\b(I|he|she|we|they|He|She|We|They) (never|rarely|seldom) ([a-z]+ed|saw|knew|thought|found|made|went|did|had|was|were|got|took|came|ran|built|spent|wrote|read|kept|held|left|sent|set|put)\b`)
	reItWasInv  = regexp.MustCompile(`(^|[.!?]\s+)It was (an? )?([a-z][^,.;:!?\n]{2,28})([.!])`)
)

// applyInversions reorders a fraction of matching sentences:
//   "I thought"      -> "thought I"
//   "I never saw"    -> "never did I see"
//   "It was a fine forging." -> "A fine forging it was."
func applyInversions(text string, rng *rand.Rand) string {
	text = reSpeechInv.ReplaceAllStringFunc(text, func(m string) string {
		if rng.Float64() >= 0.5 {
			return m
		}
		g := reSpeechInv.FindStringSubmatch(m)
		pron, verb := g[1], g[2]
		capital := pron[0] >= 'A' && pron[0] <= 'Z' && pron != "I"
		if pron != "I" {
			pron = strings.ToLower(pron)
		}
		if capital {
			verb = strings.ToUpper(verb[:1]) + verb[1:]
		}
		return verb + " " + pron
	})

	text = reNeverInv.ReplaceAllStringFunc(text, func(m string) string {
		if rng.Float64() >= 0.5 {
			return m
		}
		g := reNeverInv.FindStringSubmatch(m)
		pron, adv, verb := g[1], g[2], g[3]
		capital := pron[0] >= 'A' && pron[0] <= 'Z' && pron != "I"
		if pron != "I" {
			pron = strings.ToLower(pron)
		}
		var out string
		if verb == "was" || verb == "were" {
			out = adv + " " + verb + " " + pron
		} else {
			lemma := verb
			if l, ok := irregularLemmas[verb]; ok {
				lemma = l
			} else {
				lemma = unEd(verb)
			}
			out = adv + " did " + pron + " " + lemma
		}
		if capital || m[0] >= 'A' && m[0] <= 'Z' && g[1] == "I" {
			out = strings.ToUpper(out[:1]) + out[1:]
		}
		return out
	})

	text = reItWasInv.ReplaceAllStringFunc(text, func(m string) string {
		if rng.Float64() >= 0.35 {
			return m
		}
		g := reItWasInv.FindStringSubmatch(m)
		lead, art, pred, punct := g[1], g[2], g[3], g[4]
		out := art + pred + " it was" + punct
		out = strings.ToUpper(out[:1]) + out[1:]
		return lead + out
	})
	return text
}
