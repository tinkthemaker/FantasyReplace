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

var (
	reNegDecl = regexp.MustCompile(`\b(do|does|did|Do|Does|Did) not ([a-z]+)\b`)
	reNegImp  = regexp.MustCompile(`(^|[.!?]\s+|\n)(Do not|do not) ([a-z]+)\b`)
)

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
		switch aux {
		case "did":
			verb = conjugate(verb, "VBD")
		case "does":
			verb = conjugate(verb, "VBZ")
		}
		return matchCase(g[1], verb+" not")
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
