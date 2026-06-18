package main

// wizard.go - the wizardify engine, v2: POS-tagged, lemma-based,
// auto-inflecting replacement.

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

// ----------------------------------------------------------------- lexicon

// Entry is either a plain replacement (any POS) or per-POS senses.
type Entry struct {
	Any  string
	Noun string
	Verb string
	Adj  string
	Adv  string
}

func (e *Entry) UnmarshalJSON(b []byte) error {
	if len(b) > 0 && b[0] == '"' {
		return json.Unmarshal(b, &e.Any)
	}
	var m struct {
		Noun string `json:"noun"`
		Verb string `json:"verb"`
		Adj  string `json:"adjective"`
		Adv  string `json:"adverb"`
	}
	if err := json.Unmarshal(b, &m); err != nil {
		return err
	}
	e.Noun, e.Verb, e.Adj, e.Adv = m.Noun, m.Verb, m.Adj, m.Adv
	return nil
}

type Lexicon struct {
	Tier1         map[string]Entry  `json:"tier1"`
	Tier2         map[string]Entry  `json:"tier2"`
	Tier3         map[string]Entry  `json:"tier3"`
	Phrases       map[string]string `json:"phrases"`
	Interjections []string          `json:"interjections"`
	Exclamations  []string          `json:"exclamations"`
	Misfortunes   []string          `json:"misfortunes"`
	Triumphs      []string          `json:"triumphs"`
	Asides        []string          `json:"asides"`
}

func LoadLexicon(path string) (*Lexicon, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var lex Lexicon
	if err := json.Unmarshal(b, &lex); err != nil {
		return nil, err
	}
	return &lex, nil
}

// ruleset is the lexicon compiled for one intensity level.
type ruleset struct {
	single map[string]Entry  // single-word lemma entries
	multi  map[string]string // multi-word literal entries
}

func buildRuleset(lex *Lexicon, intensity int) *ruleset {
	rs := &ruleset{single: map[string]Entry{}, multi: map[string]string{}}
	add := func(m map[string]Entry) {
		for k, e := range m {
			if strings.Contains(k, " ") {
				rs.multi[k] = firstSense(e)
			} else {
				rs.single[strings.ToLower(k)] = e
			}
		}
	}
	add(lex.Tier1)
	if intensity >= 2 {
		add(lex.Tier2)
	}
	if intensity >= 3 {
		add(lex.Tier3)
	}
	return rs
}

func firstSense(e Entry) string {
	for _, s := range []string{e.Any, e.Noun, e.Verb, e.Adj, e.Adv} {
		if s != "" {
			return s
		}
	}
	return ""
}

// pickSense chooses the replacement for a Penn tag; inflects to match.
func pickSense(e Entry, tag string) string {
	switch {
	case strings.HasPrefix(tag, "NN"):
		if e.Noun != "" {
			return inflectNounPhrase(e.Noun, tag == "NNS" || tag == "NNPS")
		}
		if e.Any != "" {
			return inflectNounPhrase(e.Any, tag == "NNS" || tag == "NNPS")
		}
	case strings.HasPrefix(tag, "VB"):
		if e.Verb != "" {
			return inflectVerbPhrase(e.Verb, tag)
		}
		if e.Any != "" {
			return inflectVerbPhrase(e.Any, tag)
		}
	case strings.HasPrefix(tag, "JJ"):
		if e.Adj != "" {
			return e.Adj
		}
		return e.Any // never use a verb/noun sense in an adjective slot
	case strings.HasPrefix(tag, "RB"):
		if e.Adv != "" {
			return e.Adv
		}
		return e.Any
	default:
		return e.Any
	}
	// noun/verb slots: mis-tags between the two are common, so cross-fall
	if e.Any != "" {
		return e.Any
	}
	if strings.HasPrefix(tag, "NN") {
		return e.Verb // "and compile" mis-tagged NN
	}
	if strings.HasPrefix(tag, "VB") {
		return e.Noun
	}
	return ""
}

// ---------------------------------------------------------------- protection

const placeholder = "\x00WIZ%d\x00"

var (
	reFrontmatter = regexp.MustCompile(`(?s)\A---\n.*?\n---\n`)
	reInlineCode  = regexp.MustCompile("`[^`\n]+`")
	reLinkURL     = regexp.MustCompile(`\]\([^)\s]+(?:\s+"[^"]*")?\)`)
	reHTMLTag     = regexp.MustCompile(`<[^>\n]+>`)
	rePlaceholder = regexp.MustCompile("\x00WIZ[0-9]+\x00")
	reBareURL     = regexp.MustCompile(`https?://[^\s)]+`)
)

func protectFences(text string, stash *[]string) string {
	lines := strings.Split(text, "\n")
	var out []string
	var block []string
	fence := ""
	for _, line := range lines {
		trimmed := strings.TrimLeft(line, " ")
		if fence == "" {
			if strings.HasPrefix(trimmed, "```") || strings.HasPrefix(trimmed, "~~~") {
				fence = trimmed[:3]
				block = []string{line}
			} else {
				out = append(out, line)
			}
		} else {
			block = append(block, line)
			if strings.HasPrefix(trimmed, fence) {
				*stash = append(*stash, strings.Join(block, "\n"))
				out = append(out, fmt.Sprintf(placeholder, len(*stash)-1))
				fence = ""
				block = nil
			}
		}
	}
	if fence != "" {
		*stash = append(*stash, strings.Join(block, "\n"))
		out = append(out, fmt.Sprintf(placeholder, len(*stash)-1))
	}
	return strings.Join(out, "\n")
}

func protect(text string) (string, []string) {
	var stash []string
	stashRe := func(re *regexp.Regexp, s string) string {
		return re.ReplaceAllStringFunc(s, func(m string) string {
			stash = append(stash, m)
			return fmt.Sprintf(placeholder, len(stash)-1)
		})
	}
	text = stashRe(reFrontmatter, text)
	text = protectFences(text, &stash)
	text = stashRe(reInlineCode, text)
	text = stashRe(reLinkURL, text)
	text = stashRe(reHTMLTag, text)
	text = stashRe(reBareURL, text)
	return text, stash
}

func restoreStash(text string, stash []string) string {
	for i, orig := range stash {
		text = strings.Replace(text, fmt.Sprintf(placeholder, i), orig, 1)
	}
	return text
}

// ------------------------------------------------------------- text helpers

func matchCase(source, repl string) string {
	if repl == "" {
		return repl
	}
	if len(source) > 1 && source == strings.ToUpper(source) &&
		strings.ToLower(source) != source {
		return strings.ToUpper(repl)
	}
	first := source[:1]
	if first == strings.ToUpper(first) && first != strings.ToLower(first) {
		return strings.ToUpper(repl[:1]) + repl[1:]
	}
	return repl
}

// applyLiteral does whole-word, longest-first replacement (multi-word entries).
func applyLiteral(text string, mapping map[string]string) string {
	if len(mapping) == 0 {
		return text
	}
	keys := make([]string, 0, len(mapping))
	lower := make(map[string]string, len(mapping))
	for k, v := range mapping {
		keys = append(keys, k)
		lower[strings.ToLower(k)] = v
	}
	sort.Slice(keys, func(i, j int) bool { return len(keys[i]) > len(keys[j]) })
	for i, k := range keys {
		keys[i] = regexp.QuoteMeta(k)
	}
	re := regexp.MustCompile(`(?i)\b(` + strings.Join(keys, "|") + `)\b`)
	return re.ReplaceAllStringFunc(text, func(m string) string {
		return matchCase(m, lower[strings.ToLower(m)])
	})
}

func applyPhrases(text string, phrases map[string]string) string {
	keys := make([]string, 0, len(phrases))
	for k := range phrases {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool { return len(keys[i]) > len(keys[j]) })
	for _, k := range keys {
		parts := strings.SplitN(k, "*", 2)
		if len(parts) != 2 {
			continue
		}
		head := regexp.QuoteMeta(strings.TrimSpace(parts[0]))
		tail := regexp.QuoteMeta(strings.TrimSpace(parts[1]))
		re := regexp.MustCompile(`(?i)\b(` + head + `)\s+(\S+(?:\s+\S+){0,3}?)\s+(` + tail + `)\b`)
		repl := phrases[k]
		text = re.ReplaceAllStringFunc(text, func(m string) string {
			g := re.FindStringSubmatch(m)
			out := strings.Replace(repl, "*", g[2], 1)
			return matchCase(g[1], out[:1]) + out[1:]
		})
	}
	return text
}

var (
	reAnCons = regexp.MustCompile(`\b([Aa])n ([bcdfgjklmnpqrstvwxyz])`)
	reAVowel = regexp.MustCompile(`\b([Aa]) ([aeiou])`)
)

func fixArticles(text string) string {
	text = reAnCons.ReplaceAllString(text, "$1 $2")
	return reAVowel.ReplaceAllString(text, "${1}n $2")
}

// ---------------------------------------------------------------- flourishes

var reSentence = regexp.MustCompile(`(?m)([.!?][ \t]+|^)([A-Z][a-z])`)
var reLowerCap = regexp.MustCompile(`([,:] \x01)([A-Z])([a-z])`)
var reExclaim = regexp.MustCompile(`([a-z][.!?])(\s|$)`)

func addInterjections(text string, interjections []string, rng *rand.Rand, rate float64) string {
	text = reSentence.ReplaceAllStringFunc(text, func(m string) string {
		g := reSentence.FindStringSubmatch(m)
		if rng.Float64() < rate {
			return g[1] + interjections[rng.Intn(len(interjections))] + " \x01" + g[2]
		}
		return m
	})
	text = reLowerCap.ReplaceAllStringFunc(text, func(m string) string {
		g := reLowerCap.FindStringSubmatch(m)
		return g[1] + strings.ToLower(g[2]) + g[3]
	})
	return strings.ReplaceAll(text, "\x01", "")
}

func addExclamations(text string, exclamations []string, rng *rand.Rand, rate float64) string {
	return reExclaim.ReplaceAllStringFunc(text, func(m string) string {
		g := reExclaim.FindStringSubmatch(m)
		if rng.Float64() < rate {
			return g[1] + " " + exclamations[rng.Intn(len(exclamations))] + g[2]
		}
		return m
	})
}


var pluralVerbMap = map[string]string{"was": "were", "is": "are", "has": "have", "does": "do"}

func looksPlural(word string) bool {
	w := strings.ToLower(word)
	return len(w) > 3 && strings.HasSuffix(w, "s") &&
		!strings.HasSuffix(w, "ss") && !strings.HasSuffix(w, "us")
}

func lastWordOf(s string) string {
	f := strings.Fields(s)
	if len(f) == 0 {
		return s
	}
	return f[len(f)-1]
}

// --------------------------------------------------------------- token pass

// tokenPass tags a segment and applies lemma lookup + inflection +
// (at intensity 3) generic archaization.
func tokenPass(text string, rs *ruleset, intensity int) string {
	if strings.TrimSpace(text) == "" {
		return text
	}
	toks := tagSegment(text)
	if len(toks) == 0 {
		return text
	}
	repl := make([]string, len(toks))
	for i, t := range toks {
		if !isWordTag(t.Tag) {
			continue
		}
		lower := strings.ToLower(t.Text)
		lemma := lemmatize(lower, t.Tag)
		if e, ok := rs.single[lemma]; ok {
			if r := pickSense(e, t.Tag); r != "" {
				if intensity >= 3 && t.Tag == "VBZ" {
					r = ethifyPhrase(r)
				}
				repl[i] = matchCase(t.Text, r)
				continue
			}
		}
		if e, ok := rs.single[lower]; ok && lower != lemma {
			if r := firstSense(e); r != "" {
				repl[i] = matchCase(t.Text, r)
				continue
			}
		}
		if strings.HasSuffix(lower, "ing") && (strings.HasPrefix(t.Tag, "NN") || t.Tag == "JJ") {
			if e, ok := rs.single[unIng(lower)]; ok && e.Verb != "" {
				repl[i] = matchCase(t.Text, inflectVerbPhrase(e.Verb, "VBG"))
				continue
			}
		}
		if intensity >= 3 {
			if r := archaizeToken(toks, i); r != "" {
				repl[i] = matchCase(t.Text, r)
			}
		}
	}
	// "Ask not" + ask->"inquire of" must yield "Inquire not of", not "Inquire of not"
	for i := range toks {
		if repl[i] == "" || !strings.Contains(repl[i], " ") {
			continue
		}
		if n := nextWord(toks, i); n >= 0 && strings.ToLower(toks[n].Text) == "not" && repl[n] == "" {
			words := strings.SplitN(repl[i], " ", 2)
			repl[i] = words[0] + " not " + words[1]
			repl[n] = "" // swallow the original "not"
		}
	}

	// agreement: singular noun replaced by plural phrase -> fix following verb
	for i := range toks {
		if repl[i] == "" || toks[i].Tag != "NN" || !looksPlural(lastWordOf(repl[i])) {
			continue
		}
		if n := nextWord(toks, i); n >= 0 && repl[n] == "" {
			if v, ok := pluralVerbMap[strings.ToLower(toks[n].Text)]; ok {
				repl[n] = v
			}
		}
	}

	// reconstruct, preserving all inter-token text
	var b strings.Builder
	cur := 0
	for i, t := range toks {
		if repl[i] == "" {
			continue
		}
		b.WriteString(text[cur:t.Start])
		if repl[i] != "" {
			b.WriteString(repl[i])
		} else if t.Start > cur || true {
			// swallowed token: also eat one trailing space
			cur = t.End
			if cur < len(text) && text[cur] == ' ' {
				cur++
			}
			continue
		}
		cur = t.End
	}
	b.WriteString(text[cur:])
	return b.String()
}

// tokenPassAll runs tokenPass on the stretches between placeholders.
func tokenPassAll(text string, rs *ruleset, intensity int) string {
	locs := rePlaceholder.FindAllStringIndex(text, -1)
	if locs == nil {
		return tokenPass(text, rs, intensity)
	}
	var b strings.Builder
	cur := 0
	for _, loc := range locs {
		b.WriteString(tokenPass(text[cur:loc[0]], rs, intensity))
		b.WriteString(text[loc[0]:loc[1]])
		cur = loc[1]
	}
	b.WriteString(tokenPass(text[cur:], rs, intensity))
	return b.String()
}

// --------------------------------------------------------------------- core

func transformText(text string, rs *ruleset, lex *Lexicon, intensity int, rng *rand.Rand) string {
	if len(lex.Phrases) > 0 {
		text = applyPhrases(text, lex.Phrases)
	}
	text = applyLiteral(text, rs.multi)
	if intensity >= 3 {
		text = expandContractions(text)
		text = removeDoSupport(text)
		text = invertQuestions(text)
	}
	text = tokenPassAll(text, rs, intensity)
	text = fixArticles(text)
	if intensity >= 3 {
		text = applyTis(text, rng, 0.5)
		text = applyMineThine(text)
		text = applyInversions(text, rng)
	}
	if intensity >= 2 && len(lex.Interjections) > 0 {
		rate := 0.12
		if intensity == 3 {
			rate = 0.2
		}
		text = addInterjections(text, lex.Interjections, rng, rate)
	}
	if intensity >= 3 {
		text = addToneFlourishes(text, lex, rng)
	}
	return text
}

// WizardifyMarkdown transmutes markdown text at the given intensity (1-3).
func WizardifyMarkdown(text string, lex *Lexicon, intensity int, seed int64) string {
	text, stash := protect(text)
	rng := rand.New(rand.NewSource(seed))
	text = transformText(text, buildRuleset(lex, intensity), lex, intensity, rng)
	return restoreStash(text, stash)
}

// WizardifyFile picks the right pipeline from the file extension.
func WizardifyFile(path, content string, lex *Lexicon, intensity int, seed int64) (string, error) {
	ext := strings.ToLower(filepath.Ext(path))
	if ext == ".html" || ext == ".htm" {
		return WizardifyHTML(content, lex, intensity, seed)
	}
	return WizardifyMarkdown(content, lex, intensity, seed), nil
}
