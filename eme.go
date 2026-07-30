package main

// eme.go - Early Modern English grammar transforms (intensity 3).
// Research notes: do-support displaced "I know not" style negation only
// by ~1700; 'tis/'twas were everyday contractions; mine/thine appeared
// before vowels and h-; questions inverted without do ("What say you?").

import (
	"regexp"
	"strings"
)

// ---- contraction expansion (precondition for negation inversion)

var contractionMap = map[string]string{
	"don't": "do not", "doesn't": "does not", "didn't": "did not",
	"isn't": "is not", "wasn't": "was not", "aren't": "are not",
	"weren't": "were not", "haven't": "have not", "hasn't": "has not",
	"hadn't": "had not", "couldn't": "could not", "shouldn't": "should not",
	"wouldn't": "would not", "can't": "cannot", "won't": "will not",
	"mustn't": "must not", "needn't": "need not", "ain't": "is not",
}

var reContraction = regexp.MustCompile(`(?i)\b(don|doesn|didn|isn|wasn|aren|weren|haven|hasn|hadn|couldn|shouldn|wouldn|can|won|mustn|needn|ain)['’]t\b`)

func expandContractions(text string) string {
	return reContraction.ReplaceAllStringFunc(text, func(m string) string {
		key := strings.ToLower(strings.ReplaceAll(m, "’", "'"))
		full, ok := contractionMap[key]
		if !ok {
			return m
		}
		return matchCase(m, full)
	})
}

// ---- negation without do-support: "I do not know" -> "I know not"
// But NOT for participle constructions: "do not get stuck" should stay,
// because "get not stuck" is ungrammatical. We detect this by checking
// if the word after the main verb looks like a past participle.

var (
	reNegDecl = regexp.MustCompile(`\b(do|does|did|Do|Does|Did) not ([a-z]+)(?:\s+([a-z]+))?`)
	reNegImp  = regexp.MustCompile(`(^|[.!?]\s+|\n)(Do not|do not) ([a-z]+)\b`)
)

// looksLikeParticiple checks if a word is likely a past participle: ends in
// -ed, -en, or is a known irregular VBN form. Conservative — better to
// skip a valid transformation than produce "get not stuck".
func looksLikeParticiple(word string) bool {
	w := strings.ToLower(word)
	if strings.HasSuffix(w, "ed") || strings.HasSuffix(w, "en") {
		return true
	}
	for _, forms := range irregularVerbs {
		if forms[1] == w { // VBN form
			return true
		}
	}
	// Common irregular past participles not in the conjugation table
	// (the table only covers verbs that have lexicon entries).
	return irregularParticiples[w]
}

// irregularParticiples supplements irregularVerbs for the participle check.
// These are common English past participles that aren't needed for
// conjugation (no lexicon entry uses them) but must be recognized by
// looksLikeParticiple to guard do-support removal.
var irregularParticiples = map[string]bool{
	"stuck": true, "struck": true, "hung": true, "swung": true,
	"flung": true, "slung": true, "wrung": true, "sung": true,
	"rung": true, "sunk": true, "begun": true, "drunk": true,
	"shrunk": true, "slain": true, "shown": true, "blown": true,
	"drawn": true, "grown": true, "known": true, "thrown": true,
	"flown": true, "sown": true, "torn": true, "worn": true,
	"born": true, "borne": true, "frozen": true, "chosen": true,
	"stolen": true, "woken": true, "bidden": true, "forbidden": true,
	"hidden": true, "ridden": true, "risen": true, "arisen": true,
	"woven": true, "overdone": true, "undone": true, "misdone": true,
	"redone": true, "foreseen": true, "overseen": true,
}

func removeDoSupport(text string) string {
	// imperatives first: "Do not touch the altar" -> "Touch not the altar"
	text = reNegImp.ReplaceAllStringFunc(text, func(m string) string {
		g := reNegImp.FindStringSubmatch(m)
		verb := g[3]
		if strings.HasPrefix(g[2], "D") {
			verb = strings.ToUpper(verb[:1]) + verb[1:]
		}
		return g[1] + verb + " not"
	})
	// declaratives: "I did not see" -> "I saw not"
	text = reNegDecl.ReplaceAllStringFunc(text, func(m string) string {
		g := reNegDecl.FindStringSubmatch(m)
		aux, verb := strings.ToLower(g[1]), g[2]
		// Skip participle constructions: "don't get stuck" should NOT become
		// "get not stuck". If the word after the verb is a past participle,
		// the verb + participle form a unit that can't absorb negation.
		if g[3] != "" && looksLikeParticiple(g[3]) {
			return m
		}
		switch aux {
		case "did":
			verb = conjugate(verb, "VBD")
		case "does":
			verb = conjugate(verb, "VBZ")
		}
		out := matchCase(g[1], verb+" not")
		// Preserve the following word if the regex captured it (g[3]).
		if g[3] != "" {
			out += " " + g[3]
		}
		return out
	})
	return text
}

// ---- question inversion: "What do you mean?" -> "What mean you?"

var reQuestion = regexp.MustCompile(`\b(What|Why|How|Where|When|Whither|Wherefore|what|why|how|where|when) (do|does) (you|I|we|they|he|she|it) ([a-z]+)`)

func invertQuestions(text string) string {
	return reQuestion.ReplaceAllStringFunc(text, func(m string) string {
		g := reQuestion.FindStringSubmatch(m)
		wh, aux, subj, verb := g[1], strings.ToLower(g[2]), g[3], g[4]
		if subj == "you" {
			// archaize here: the tagger cannot be trusted with inverted clauses
			return wh + " " + estify(verb) + " thou"
		}
		if aux == "does" {
			verb = conjugate(verb, "VBZ")
		}
		return wh + " " + verb + " " + subj
	})
}

// ---- 'tis / 'twas / 'twill / 'twould

var reItIs = regexp.MustCompile(`\b(It|it) (is|was|will|would)\b`)

var tisMap = map[string]string{
	"is": "'tis", "was": "'twas", "will": "'twill", "would": "'twould",
}

func applyTis(text string, rng interface{ Float64() float64 }, rate float64) string {
	return reItIs.ReplaceAllStringFunc(text, func(m string) string {
		if rng.Float64() >= rate {
			return m
		}
		g := reItIs.FindStringSubmatch(m)
		out := tisMap[g[2]]
		if g[1] == "It" {
			return "'" + strings.ToUpper(out[1:2]) + out[2:]
		}
		return out
	})
}

// ---- mine/thine before vowels and h-

var reMyThy = regexp.MustCompile(`\b(My|my|Thy|thy) ([aeiouAEIOUh])`)

func applyMineThine(text string) string {
	return reMyThy.ReplaceAllStringFunc(text, func(m string) string {
		g := reMyThy.FindStringSubmatch(m)
		poss := map[string]string{"my": "mine", "thy": "thine", "My": "Mine", "Thy": "Thine"}[g[1]]
		return poss + " " + g[2]
	})
}
