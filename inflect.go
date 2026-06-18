package main

// inflect.go - lemmatization and inflection so one lexicon entry
// covers every surface form (crash -> crashes/crashed/crashing).

import "strings"

// irregular verb: lemma -> [VBD, VBN, VBZ, VBG]
var irregularVerbs = map[string][4]string{
	"abide":  {"abode", "abode", "abides", "abiding"},
	"be":     {"was", "been", "is", "being"},
	"break":  {"broke", "broken", "breaks", "breaking"},
	"bring":  {"brought", "brought", "brings", "bringing"},
	"build":  {"built", "built", "builds", "building"},
	"buy":    {"bought", "bought", "buys", "buying"},
	"cast":   {"cast", "cast", "casts", "casting"},
	"come":   {"came", "come", "comes", "coming"},
	"do":     {"did", "done", "does", "doing"},
	"find":   {"found", "found", "finds", "finding"},
	"forge":  {"forged", "forged", "forges", "forging"},
	"get":    {"got", "gotten", "gets", "getting"},
	"give":   {"gave", "given", "gives", "giving"},
	"go":     {"went", "gone", "goes", "going"},
	"have":   {"had", "had", "has", "having"},
	"hold":   {"held", "held", "holds", "holding"},
	"keep":   {"kept", "kept", "keeps", "keeping"},
	"know":   {"knew", "known", "knows", "knowing"},
	"learn":  {"learnt", "learnt", "learns", "learning"},
	"leave":  {"left", "left", "leaves", "leaving"},
	"make":   {"made", "made", "makes", "making"},
	"put":    {"put", "put", "puts", "putting"},
	"read":   {"read", "read", "reads", "reading"},
	"run":    {"ran", "run", "runs", "running"},
	"say":    {"said", "said", "says", "saying"},
	"see":    {"saw", "seen", "sees", "seeing"},
	"send":   {"sent", "sent", "sends", "sending"},
	"set":    {"set", "set", "sets", "setting"},
	"slay":   {"slew", "slain", "slays", "slaying"},
	"spend":  {"spent", "spent", "spends", "spending"},
	"sunder": {"sundered", "sundered", "sunders", "sundering"},
	"take":   {"took", "taken", "takes", "taking"},
	"think":  {"thought", "thought", "thinks", "thinking"},
	"weave":  {"wove", "woven", "weaves", "weaving"},
	"write":  {"wrote", "written", "writes", "writing"},
}

// surface form -> lemma, derived from irregularVerbs plus extras
var irregularLemmas = func() map[string]string {
	m := map[string]string{
		"was": "be", "were": "be", "been": "be", "am": "be", "are": "be", "is": "be",
		"went": "go", "gone": "go", "did": "do", "done": "do",
	}
	for lemma, forms := range irregularVerbs {
		for _, f := range forms {
			if _, ok := m[f]; !ok {
				m[f] = lemma
			}
		}
	}
	return m
}()

var irregularPlurals = map[string]string{
	"child": "children", "foot": "feet", "tooth": "teeth", "man": "men",
	"woman": "women", "mouse": "mice", "person": "people", "datum": "data",
}

var consonants = "bcdfghjklmnpqrstvwxz"

func isVowel(c byte) bool { return strings.IndexByte("aeiou", c) >= 0 }

// lemmatize reduces a lowercase word to its dictionary form using its POS tag.
func lemmatize(word, tag string) string {
	if l, ok := irregularLemmas[word]; ok && strings.HasPrefix(tag, "V") {
		return l
	}
	switch {
	case tag == "NNS" || tag == "NNPS":
		return singularize(word)
	case tag == "VBZ":
		return singularize(word)
	case tag == "VBD" || tag == "VBN":
		return unEd(word)
	case tag == "VBG":
		return unIng(word)
	}
	return word
}

func singularize(w string) string {
	switch {
	case strings.HasSuffix(w, "ies") && len(w) > 4:
		return w[:len(w)-3] + "y"
	case strings.HasSuffix(w, "sses"), strings.HasSuffix(w, "shes"),
		strings.HasSuffix(w, "ches"), strings.HasSuffix(w, "xes"), strings.HasSuffix(w, "zes"):
		return w[:len(w)-2]
	case strings.HasSuffix(w, "s") && !strings.HasSuffix(w, "ss") && len(w) > 3:
		return w[:len(w)-1]
	}
	return w
}

func unEd(w string) string {
	if !strings.HasSuffix(w, "ed") || len(w) < 4 {
		return w
	}
	stem := w[:len(w)-2]
	switch {
	case strings.HasSuffix(stem, "i"):
		return stem[:len(stem)-1] + "y" // tried -> try
	case len(stem) > 2 && stem[len(stem)-1] == stem[len(stem)-2] &&
		strings.IndexByte(consonants, stem[len(stem)-1]) >= 0 &&
		!strings.HasSuffix(stem, "ll") && !strings.HasSuffix(stem, "ss"):
		return stem[:len(stem)-1] // stopped -> stop
	case strings.HasSuffix(w, "ced"), strings.HasSuffix(w, "sed"),
		strings.HasSuffix(w, "ved"), strings.HasSuffix(w, "ged"),
		strings.HasSuffix(w, "zed"), strings.HasSuffix(w, "ked"),
		strings.HasSuffix(w, "ted") && isVowel(stem[len(stem)-2]):
		return stem + "e" // raced -> race, noted -> note
	}
	return stem
}

func unIng(w string) string {
	if !strings.HasSuffix(w, "ing") || len(w) < 5 {
		return w
	}
	stem := w[:len(w)-3]
	switch {
	case len(stem) > 2 && stem[len(stem)-1] == stem[len(stem)-2] &&
		strings.IndexByte(consonants, stem[len(stem)-1]) >= 0 &&
		!strings.HasSuffix(stem, "ll") && !strings.HasSuffix(stem, "ss"):
		return stem[:len(stem)-1] // running -> run
	case strings.HasSuffix(stem, "v"), strings.HasSuffix(stem, "c"),
		strings.HasSuffix(stem, "s") && !strings.HasSuffix(stem, "ss"),
		strings.HasSuffix(stem, "z"),
		len(stem) > 2 && len(stem) <= 4 && strings.IndexByte(consonants, stem[len(stem)-1]) >= 0 &&
			isVowel(stem[len(stem)-2]) && !isVowel(stem[len(stem)-3]) &&
			!strings.HasSuffix(stem, "w") && !strings.HasSuffix(stem, "x") && !strings.HasSuffix(stem, "y"):
		return stem + "e" // weaving -> weave, making -> make (short stems only)
	}
	return stem
}

func pluralize(w string) string {
	if p, ok := irregularPlurals[w]; ok {
		return p
	}
	switch {
	case strings.HasSuffix(w, "y") && len(w) > 1 &&
		strings.IndexByte(consonants, w[len(w)-2]) >= 0:
		return w[:len(w)-1] + "ies"
	case strings.HasSuffix(w, "s"), strings.HasSuffix(w, "sh"),
		strings.HasSuffix(w, "ch"), strings.HasSuffix(w, "x"), strings.HasSuffix(w, "z"):
		return w + "es"
	}
	return w + "s"
}

// conjugate produces the requested form (VBZ/VBD/VBN/VBG) of a lemma.
func conjugate(lemma, tag string) string {
	if forms, ok := irregularVerbs[lemma]; ok {
		switch tag {
		case "VBD":
			return forms[0]
		case "VBN":
			return forms[1]
		case "VBZ":
			return forms[2]
		case "VBG":
			return forms[3]
		}
	}
	switch tag {
	case "VBZ":
		return pluralize(lemma)
	case "VBD", "VBN":
		switch {
		case strings.HasSuffix(lemma, "e"):
			return lemma + "d"
		case strings.HasSuffix(lemma, "y") && len(lemma) > 1 &&
			strings.IndexByte(consonants, lemma[len(lemma)-2]) >= 0:
			return lemma[:len(lemma)-1] + "ied"
		case shouldDouble(lemma):
			return lemma + string(lemma[len(lemma)-1]) + "ed"
		default:
			return lemma + "ed"
		}
	case "VBG":
		switch {
		case strings.HasSuffix(lemma, "ie"):
			return lemma[:len(lemma)-2] + "ying"
		case strings.HasSuffix(lemma, "e") && !strings.HasSuffix(lemma, "ee"):
			return lemma[:len(lemma)-1] + "ing"
		case shouldDouble(lemma):
			return lemma + string(lemma[len(lemma)-1]) + "ing"
		default:
			return lemma + "ing"
		}
	}
	return lemma
}

// shouldDouble: short verb ending consonant-vowel-consonant (stop -> stopped)
func shouldDouble(w string) bool {
	n := len(w)
	if n < 3 {
		return false
	}
	last, mid, first := w[n-1], w[n-2], w[n-3]
	return strings.IndexByte(consonants, last) >= 0 && last != 'w' && last != 'x' && last != 'y' &&
		isVowel(mid) && !isVowel(first) && n <= 5
}

// inflectNounPhrase pluralizes the head noun: the word before " of "
// if present, otherwise the last word.
func inflectNounPhrase(phrase string, plural bool) string {
	if !plural {
		return phrase
	}
	words := strings.Fields(phrase)
	if len(words) == 0 {
		return phrase
	}
	head := len(words) - 1
	for i, w := range words {
		if w == "of" && i > 0 {
			head = i - 1
			break
		}
	}
	words[head] = pluralize(words[head])
	return strings.Join(words, " ")
}

// inflectVerbPhrase conjugates the first word: "conjure down" -> "conjured down".
func inflectVerbPhrase(phrase, tag string) string {
	words := strings.Fields(phrase)
	if len(words) == 0 {
		return phrase
	}
	words[0] = conjugate(words[0], tag)
	return strings.Join(words, " ")
}
