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

// TransformOptions exposes the stylistic layers independently while keeping
// the three familiar intensity presets as defaults.
type TransformOptions struct {
	Intensity     int
	Seed          int64
	FlourishSeed  int64
	Grammar       bool
	Interjections bool
	Flourishes    bool
	NewsSafe      bool
	NewsFlair     int
}

func DefaultTransformOptions(intensity int, seed int64) TransformOptions {
	return TransformOptions{
		Intensity: intensity, Seed: seed, FlourishSeed: seed,
		Grammar: intensity >= 3, Interjections: intensity >= 2, Flourishes: intensity >= 3,
	}
}

func DefaultNewsSafeOptions(seed int64) TransformOptions {
	return TransformOptions{
		Intensity: 2, Seed: seed, FlourishSeed: seed, NewsSafe: true,
		Grammar: false, Interjections: false, Flourishes: false,
	}
}

// Wizardifier owns immutable, precompiled rulesets and is safe to reuse for
// many transformations (including concurrent TUI commands).
type Wizardifier struct {
	lex       *Lexicon
	rules     [4]*ruleset
	newsRules *ruleset
}

func NewWizardifier(lex *Lexicon) (*Wizardifier, error) {
	if lex == nil {
		return nil, fmt.Errorf("nil lexicon")
	}
	if err := validateLexicon(lex); err != nil {
		return nil, err
	}
	w := &Wizardifier{lex: lex}
	for intensity := 1; intensity <= 3; intensity++ {
		w.rules[intensity] = buildRuleset(lex, intensity)
	}
	w.newsRules = buildNewsSafeRuleset(lex)
	return w, nil
}

func (w *Wizardifier) ruleset(intensity int) (*ruleset, error) {
	if intensity < 1 || intensity > 3 {
		return nil, fmt.Errorf("intensity must be 1, 2, or 3")
	}
	return w.rules[intensity], nil
}

func (w *Wizardifier) rulesetFor(opts TransformOptions) (*ruleset, error) {
	if opts.NewsFlair < 0 || opts.NewsFlair > 3 {
		return nil, fmt.Errorf("news flair must be between 0 and 3")
	}
	if opts.NewsFlair > 0 && !opts.NewsSafe {
		return nil, fmt.Errorf("news flair requires the news-safe profile")
	}
	if opts.NewsSafe {
		if opts.Intensity < 1 || opts.Intensity > 3 {
			return nil, fmt.Errorf("intensity must be 1, 2, or 3")
		}
		return w.newsRules, nil
	}
	return w.ruleset(opts.Intensity)
}

func LoadLexicon(path string) (*Lexicon, error) {
	b := embeddedLexicon
	if path != "" {
		var err error
		b, err = os.ReadFile(path)
		if err != nil {
			return nil, err
		}
	}
	var lex Lexicon
	if err := json.Unmarshal(b, &lex); err != nil {
		return nil, err
	}
	if err := validateLexicon(&lex); err != nil {
		return nil, err
	}
	return &lex, nil
}

func validateLexicon(lex *Lexicon) error {
	if len(lex.Tier1)+len(lex.Tier2)+len(lex.Tier3) == 0 {
		return fmt.Errorf("lexicon has no replacement entries")
	}
	for tier, entries := range map[string]map[string]Entry{
		"tier1": lex.Tier1, "tier2": lex.Tier2, "tier3": lex.Tier3,
	} {
		for key, entry := range entries {
			if strings.TrimSpace(key) == "" {
				return fmt.Errorf("%s contains an empty key", tier)
			}
			if firstSense(entry) == "" {
				return fmt.Errorf("%s entry %q has no replacement", tier, key)
			}
			for _, value := range []string{entry.Any, entry.Noun, entry.Verb, entry.Adj, entry.Adv} {
				if value != "" && hasEmptyVariant(value) {
					return fmt.Errorf("%s entry %q contains an empty variant", tier, key)
				}
			}
		}
	}
	for key, value := range lex.Phrases {
		if strings.Count(key, "*") != 1 || strings.Count(value, "*") != 1 {
			return fmt.Errorf("phrase %q must contain exactly one * in its key and replacement", key)
		}
	}
	for name, values := range map[string][]string{
		"interjections": lex.Interjections, "exclamations": lex.Exclamations,
		"misfortunes": lex.Misfortunes, "triumphs": lex.Triumphs, "asides": lex.Asides,
	} {
		for i, value := range values {
			if strings.TrimSpace(value) == "" {
				return fmt.Errorf("%s entry %d is empty", name, i)
			}
		}
	}
	return nil
}

func hasEmptyVariant(value string) bool {
	if !strings.Contains(value, "|") {
		return false
	}
	for _, variant := range strings.Split(value, "|") {
		if strings.TrimSpace(variant) == "" {
			return true
		}
	}
	return false
}

// ruleset is the lexicon compiled for one intensity level.
type ruleset struct {
	single  map[string]Entry  // single-word lemma entries
	multi   map[string]string // lowercase multi-word literal entries
	literal *regexp.Regexp
	phrases []phraseRule
}

type phraseRule struct {
	re   *regexp.Regexp
	repl string
}

func buildRuleset(lex *Lexicon, intensity int) *ruleset {
	rs := &ruleset{single: map[string]Entry{}, multi: map[string]string{}}
	add := func(m map[string]Entry) {
		for k, e := range m {
			if strings.Contains(k, " ") {
				rs.multi[strings.ToLower(k)] = firstSense(e)
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
	rs.compile(lex.Phrases)
	return rs
}

// newsSafeKeys is intentionally small. These replacements add voice without
// renaming institutions, technologies, places, roles, or other factual nouns.
// Expanding this list requires a news-safety regression test.
var newsSafeKeys = map[string]bool{
	"news": true, "story": true, "message": true,
	"again": true, "always": true, "begin": true, "change": true,
	"continue": true, "create": true, "eventually": true, "finally": true,
	"finish": true, "help": true, "honestly": true, "immediately": true,
	"important": true, "improve": true, "learn": true, "make": true,
	"maybe": true, "nearly": true, "new": true, "often": true,
	"read": true, "remember": true, "show": true, "start": true,
	"stop": true, "together": true, "today": true, "tomorrow": true,
	"write": true, "yesterday": true,
}

func buildNewsSafeRuleset(lex *Lexicon) *ruleset {
	rs := &ruleset{single: map[string]Entry{}, multi: map[string]string{}}
	add := func(entries map[string]Entry) {
		for key, entry := range entries {
			lower := strings.ToLower(key)
			if !newsSafeKeys[lower] {
				continue
			}
			if strings.Contains(lower, " ") {
				rs.multi[lower] = firstSense(entry)
			} else {
				rs.single[lower] = entry
			}
		}
	}
	add(lex.Tier1)
	add(lex.Tier2)
	rs.compile(nil) // object-slot phrases are too broad for factual summaries
	return rs
}

func (rs *ruleset) compile(phrases map[string]string) {
	keys := make([]string, 0, len(rs.multi))
	for key := range rs.multi {
		keys = append(keys, key)
	}
	sort.Slice(keys, func(i, j int) bool { return len(keys[i]) > len(keys[j]) })
	quoted := make([]string, len(keys))
	for i, key := range keys {
		quoted[i] = regexp.QuoteMeta(key)
	}
	if len(quoted) > 0 {
		rs.literal = regexp.MustCompile(`(?i)\b(` + strings.Join(quoted, "|") + `)\b`)
	}

	phraseKeys := make([]string, 0, len(phrases))
	for key := range phrases {
		phraseKeys = append(phraseKeys, key)
	}
	sort.Slice(phraseKeys, func(i, j int) bool { return len(phraseKeys[i]) > len(phraseKeys[j]) })
	for _, key := range phraseKeys {
		parts := strings.SplitN(key, "*", 2)
		head := regexp.QuoteMeta(strings.TrimSpace(parts[0]))
		tail := regexp.QuoteMeta(strings.TrimSpace(parts[1]))
		rs.phrases = append(rs.phrases, phraseRule{
			re:   regexp.MustCompile(`(?i)\b(` + head + `)\s+(\S+(?:\s+\S+){0,3}?)\s+(` + tail + `)\b`),
			repl: phrases[key],
		})
	}
}

func firstSense(e Entry) string {
	for _, s := range []string{e.Any, e.Noun, e.Verb, e.Adj, e.Adv} {
		if s != "" {
			return s
		}
	}
	return ""
}

// pickVariant selects from a pipe-delimited alternative list.
// "common folk|mortal soul" → one of the two. Used for multi-word literal
// entries where per-occurrence rotation is not tracked.
func pickVariant(s string, rng *rand.Rand) string {
	if rng == nil || !strings.Contains(s, "|") {
		return s
	}
	parts := strings.Split(s, "|")
	return strings.TrimSpace(parts[rng.Intn(len(parts))])
}

// pickVariantRot selects from a pipe-delimited list, rotating so consecutive
// occurrences of the same key get DIFFERENT values (anti-repetition). The
// variant order is shuffled once per key (seeded), then rotated through, so
// the first n occurrences of a lemma are all distinct while staying
// deterministic for a given seed.
func pickVariantRot(s, key string, seen map[string]int, order map[string][]int, rng *rand.Rand) string {
	if !strings.Contains(s, "|") {
		return s
	}
	parts := strings.Split(s, "|")
	n := len(parts)
	seq, ok := order[key]
	if !ok {
		if rng != nil {
			seq = rng.Perm(n)
		} else {
			seq = make([]int, n)
			for i := range seq {
				seq[i] = i
			}
		}
		order[key] = seq
	}
	idx := seen[key] % n
	seen[key]++
	return strings.TrimSpace(parts[seq[idx]])
}

// leadingArticle returns the leading "the "/"a "/"an " of a replacement, or "".
func leadingArticle(s string) string {
	low := strings.ToLower(s)
	for _, art := range []string{"the ", "an ", "a "} {
		if strings.HasPrefix(low, art) {
			return art[:len(art)-1]
		}
	}
	return ""
}

// stripLeadingArticle removes a leading article from a replacement, preserving
// the original capitalization of the following word.
func stripLeadingArticle(s string) string {
	art := leadingArticle(s)
	if art == "" {
		return s
	}
	return s[len(art)+1:]
}

// isFunctionWord reports whether a Penn tag is a function word that can head
// a noun phrase requiring a determiner: determiners (DT), prepositions (IN),
// coordinating conjunctions (CC), the infinitive marker (TO), verbs (VB*),
// and modals (MD). Used by the article-strip pass: an expansion's leading
// article is kept after these, stripped after content words (modifiers).
func isFunctionWord(tag string) bool {
	return strings.HasPrefix(tag, "DT") || strings.HasPrefix(tag, "IN") ||
		strings.HasPrefix(tag, "CC") || strings.HasPrefix(tag, "VB") ||
		tag == "TO" || tag == "MD"
}

// rawSense returns the uninflected replacement string for a Penn tag (which
// may be a pipe-delimited variant list). It does NOT inflect — the caller
// resolves variants, then inflects with inflectChosen.
func rawSense(e Entry, tag string) string {
	switch {
	case strings.HasPrefix(tag, "NN"):
		if e.Noun != "" {
			return e.Noun
		}
		if e.Any != "" {
			return e.Any
		}
		return e.Verb // mis-tagged verb as noun
	case strings.HasPrefix(tag, "VB"):
		if e.Verb != "" {
			return e.Verb
		}
		// Noun-only entry in a verb slot (mis-tagged): return uninflected.
		// Inflecting "stratagem" as a verb produces non-words like "stratageming".
		if e.Any != "" {
			return e.Any
		}
		return e.Noun // mis-tagged noun as verb
	case strings.HasPrefix(tag, "JJ"):
		if e.Adj != "" {
			return e.Adj
		}
		return e.Any
	case strings.HasPrefix(tag, "RB"):
		if e.Adv != "" {
			return e.Adv
		}
		return e.Any
	default:
		return e.Any
	}
}

// inflectChosen inflects a single (already variant-resolved, pipe-free)
// replacement string to match the Penn tag.
func inflectChosen(s, tag string) string {
	switch {
	case strings.HasPrefix(tag, "NN"):
		return inflectNounPhrase(s, tag == "NNS" || tag == "NNPS")
	case strings.HasPrefix(tag, "VB"):
		return inflectVerbPhrase(s, tag)
	}
	return s
}

// ---------------------------------------------------------------- protection

const placeholder = "\x00WIZ%d\x00"

var (
	reFrontmatter = regexp.MustCompile(`(?s)\A---\r?\n.*?\r?\n---\r?\n`)
	reHTMLTag     = regexp.MustCompile(`<[^>\n]+>`)
	rePlaceholder = regexp.MustCompile("\x00WIZ[0-9]+\x00")
	reBareURL     = regexp.MustCompile(`https?://[^\s)]+`)
	reNewsQuote   = regexp.MustCompile(`(?m)"[^"\n]*"|“[^”\n]*”`)
	reNewsNumber  = regexp.MustCompile(`(?i)(?:[$€£¥]\s*)?\b\d[\d,.]*(?:%|[a-z]{1,4})?\b`)
	reNewsAcronym = regexp.MustCompile(`\b[A-Z][A-Z0-9]{1,}(?:-[A-Z0-9]+)*\b`)
	reNewsSource  = regexp.MustCompile(`(?mi)^(?:sources?|source links?|attribution|original article|reported by)\s*:.*$`)
	reNewsQuoteMD = regexp.MustCompile(`(?m)^>[^\n]*(?:\n>[^\n]*)*`)
	reNewsLink    = regexp.MustCompile(`\[[^\]\n]+\]\x00WIZ[0-9]+\x00`)
)

func stashRegexp(text string, re *regexp.Regexp, stash *[]string) string {
	return re.ReplaceAllStringFunc(text, func(match string) string {
		*stash = append(*stash, match)
		return fmt.Sprintf(placeholder, len(*stash)-1)
	})
}

// protectNewsFacts removes fact-sensitive spans from the stylistic pipeline.
// The restricted ruleset is the primary safety boundary; these placeholders
// additionally guarantee byte-for-byte preservation of quotations, figures,
// attribution, headlines, links, acronyms, and proper names.
func protectNewsFacts(text string, stash *[]string) string {
	text = stashRegexp(text, reHeadingLine, stash)
	text = stashRegexp(text, reNewsSource, stash)
	text = stashRegexp(text, reNewsQuoteMD, stash)
	text = stashRegexp(text, reNewsQuote, stash)
	text = stashRegexp(text, reNewsLink, stash)
	text = stashRegexp(text, reNewsNumber, stash)
	text = stashRegexp(text, reNewsAcronym, stash)
	return protectProperNames(text, stash)
}

func protectProperNames(text string, stash *[]string) string {
	tokens := tagSegment(text)
	type span struct{ start, end int }
	var spans []span
	for _, token := range tokens {
		if token.Tag == "NNP" || token.Tag == "NNPS" {
			spans = append(spans, span{token.Start, token.End})
		}
	}
	if len(spans) == 0 {
		return text
	}
	var b strings.Builder
	cur := 0
	for _, item := range spans {
		if item.start < cur {
			continue
		}
		b.WriteString(text[cur:item.start])
		*stash = append(*stash, text[item.start:item.end])
		b.WriteString(fmt.Sprintf(placeholder, len(*stash)-1))
		cur = item.end
	}
	b.WriteString(text[cur:])
	return b.String()
}

func protectInlineCode(text string, stash *[]string) string {
	var b strings.Builder
	for i := 0; i < len(text); {
		if text[i] != '`' {
			b.WriteByte(text[i])
			i++
			continue
		}
		run := 1
		for i+run < len(text) && text[i+run] == '`' {
			run++
		}
		delim := text[i : i+run]
		rel := strings.Index(text[i+run:], delim)
		if rel < 0 {
			b.WriteString(delim)
			i += run
			continue
		}
		end := i + run + rel + run
		*stash = append(*stash, text[i:end])
		b.WriteString(fmt.Sprintf(placeholder, len(*stash)-1))
		i = end
	}
	return b.String()
}

func protectLinkDestinations(text string, stash *[]string) string {
	var b strings.Builder
	for i := 0; i < len(text); {
		rel := strings.Index(text[i:], "](")
		if rel < 0 {
			b.WriteString(text[i:])
			break
		}
		start := i + rel
		b.WriteString(text[i:start])
		depth, end := 1, start+2
		for end < len(text) && depth > 0 {
			if text[end] == '\\' {
				end += 2
				continue
			}
			switch text[end] {
			case '(':
				depth++
			case ')':
				depth--
			}
			end++
		}
		if depth != 0 {
			b.WriteString(text[start : start+2])
			i = start + 2
			continue
		}
		*stash = append(*stash, text[start:end])
		b.WriteString(fmt.Sprintf(placeholder, len(*stash)-1))
		i = end
	}
	return b.String()
}

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
	text = protectInlineCode(text, &stash)
	text = protectLinkDestinations(text, &stash)
	text = stashRe(reHTMLTag, text)
	text = stashRe(reBareURL, text)
	return text, stash
}

func restoreStash(text string, stash []string) string {
	// Later entries may contain placeholders created earlier (for example an
	// inline-code span inside a heading). Restore in reverse nesting order.
	for i := len(stash) - 1; i >= 0; i-- {
		orig := stash[i]
		text = strings.Replace(text, fmt.Sprintf(placeholder, i), orig, 1)
	}
	return text
}

// ------------------------------------------------------------- text helpers

// massNouns are English nouns that are grammatically singular despite ending
// in -s or being inherently uncountable. The POS tagger often tags them NNS,
// which would cause the engine to pluralize their replacements ("knowledge
// essences", "arcane workingses"). Listing them here forces singular
// inflection regardless of the tagger's output.
var massNouns = map[string]bool{
	"data": true, "information": true, "software": true, "hardware": true,
	"knowledge": true, "research": true, "evidence": true, "traffic": true,
	"content": true, "infrastructure": true, "equipment": true, "machinery": true,
	"news": true, "advice": true, "feedback": true, "progress": true,
	"code":   true, // mass noun in "write code" (but tagged NN usually)
	"series": true, "species": true, "means": true, "lens": true,
}

func isMassNoun(lemma string) bool {
	return massNouns[strings.ToLower(lemma)]
}

func matchCase(source, repl string) string {
	if repl == "" {
		return repl
	}
	// ALL-CAPS source: for multi-word replacements, use Title Case (each word
	// capitalized) rather than ALL CAPS, because acronyms like "AI" → "ARCANE
	// MIND" read poorly in body text. Single-word replacements stay ALL CAPS.
	if len(source) > 1 && source == strings.ToUpper(source) &&
		strings.ToLower(source) != source {
		if strings.Contains(repl, " ") {
			return toTitleCase(repl)
		}
		return strings.ToUpper(repl)
	}
	first := source[:1]
	if first == strings.ToUpper(first) && first != strings.ToLower(first) {
		return strings.ToUpper(repl[:1]) + repl[1:]
	}
	return repl
}

// toTitleCase capitalizes the first letter of each word.
func toTitleCase(s string) string {
	var b strings.Builder
	capNext := true
	for _, r := range s {
		if capNext && r >= 'a' && r <= 'z' {
			b.WriteRune(r - 32)
		} else {
			b.WriteRune(r)
		}
		capNext = r == ' '
	}
	return b.String()
}

// applyLiteral does whole-word, longest-first replacement (multi-word entries).
func applyLiteral(text string, rs *ruleset, rng *rand.Rand) string {
	if rs.literal == nil {
		return text
	}
	return rs.literal.ReplaceAllStringFunc(text, func(m string) string {
		return matchCase(m, pickVariant(rs.multi[strings.ToLower(m)], rng))
	})
}

func applyPhrases(text string, phrases []phraseRule) string {
	for _, rule := range phrases {
		text = rule.re.ReplaceAllStringFunc(text, func(m string) string {
			g := rule.re.FindStringSubmatch(m)
			out := strings.Replace(rule.repl, "*", g[2], 1)
			return matchCase(g[1], out[:1]) + out[1:]
		})
	}
	return text
}

var (
	reAnCons = regexp.MustCompile(`\b([Aa])n ([bcdfgjklmnpqrstvwxyz])`)
	reAVowel = regexp.MustCompile(`\b([Aa]) ([aeiou])`)
	// "an" before words starting with a consonant-letter but vowel SOUND
	// (silent h: hour, honest, honor, heir) — the letter check misses these.
	reASilentH = regexp.MustCompile(`\b([Aa]) ([Hh](?:our|onest|onor|eir))\b`)
	// "a" before words starting with a vowel-letter but consonant SOUND
	// (u pronounced "you": unique, university, useful, user, European, etc.)
	reAnUconsonant = regexp.MustCompile(`\b([Aa])n ([Uu](?:nique|niversity|ser|seful|niversal|rl|rop|nanimous|nited)|[Ee]uropean|[Oo]ne[- ])\b`)
)

func fixArticles(text string) string {
	text = reAnCons.ReplaceAllString(text, "$1 $2")
	text = reAVowel.ReplaceAllString(text, "${1}n $2")
	// Silent-h: "a hour" → "an hour" (h is a consonant letter but the
	// word starts with a vowel sound).
	text = reASilentH.ReplaceAllString(text, "${1}n $2")
	// U-as-consonant: "an unique" → "a unique" (u is a vowel letter but
	// pronounced "yoo", a consonant sound).
	text = reAnUconsonant.ReplaceAllString(text, "$1 $2")
	return text
}

// reDoubledArticle matches two consecutive articles ("the the", "a an", etc.),
// even across sentence-final punctuation like "U.S." which doesn't create a
// real sentence boundary between an article and the next word.
// Never grammatical in English, so the rule has zero false positives.
var reDoubledArticle = regexp.MustCompile(`(?i)\b(the|a|an)[.\s]+(the|a|an)\b`)

func fixDoubledArticles(text string) string {
	return reDoubledArticle.ReplaceAllStringFunc(text, func(m string) string {
		g := reDoubledArticle.FindStringSubmatch(m)
		return g[1]
	})
}

// reAbbrevArticle matches a capitalized abbreviation (e.g. "U.S.", "U.K.")
// immediately followed by a replacement phrase that starts with an article.
// "The U.S. government" → replacement "the crown" → "The U.S. the crown"
// is awkward. We strip the replacement's leading article.
var reAbbrevArticle = regexp.MustCompile(`([A-Z]\.[A-Z]\.)\s+(the|a|an)\s+`)

func fixAbbrevArticles(text string) string {
	return reAbbrevArticle.ReplaceAllString(text, "$1 ")
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
func tokenPass(text string, rs *ruleset, intensity int, rng *rand.Rand) string {
	if strings.TrimSpace(text) == "" {
		return text
	}
	toks := tagSegment(text)
	if len(toks) == 0 {
		return text
	}
	seen := map[string]int{}    // per-lemma occurrence counter (rotation)
	order := map[string][]int{} // per-lemma shuffled variant order
	repl := make([]string, len(toks))
	for i, t := range toks {
		if !isWordTag(t.Tag) {
			continue
		}
		lower := strings.ToLower(t.Text)
		lemma := lemmatize(lower, t.Tag)
		// Mass nouns: force singular tag so replacements don't get pluralized
		// ("Cloudflare data" → "knowledge essence", not "knowledge essences").
		effectiveTag := t.Tag
		if isMassNoun(lemma) && (effectiveTag == "NNS" || effectiveTag == "NNPS") {
			effectiveTag = "NN"
		}
		if e, ok := rs.single[lemma]; ok {
			if raw := rawSense(e, effectiveTag); raw != "" {
				r := pickVariantRot(raw, lemma, seen, order, rng)
				r = inflectChosen(r, effectiveTag)
				if intensity >= 3 && effectiveTag == "VBZ" {
					r = ethifyPhrase(r)
				}
				repl[i] = matchCase(t.Text, r)
				continue
			}
		}
		if e, ok := rs.single[lower]; ok && lower != lemma {
			if r := firstSense(e); r != "" {
				r = pickVariantRot(r, lower, seen, order, rng)
				repl[i] = matchCase(t.Text, r)
				continue
			}
		}
		if strings.HasSuffix(lower, "ing") && (strings.HasPrefix(t.Tag, "NN") || t.Tag == "JJ") {
			if e, ok := rs.single[unIng(lower)]; ok && e.Verb != "" {
				v := pickVariantRot(e.Verb, unIng(lower), seen, order, rng)
				repl[i] = matchCase(t.Text, inflectVerbPhrase(v, "VBG"))
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
			repl[n] = "\u0002" // swallow the original "not"
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

	// valency: "help X <verb>" → "aid X to <verb>". "Aid" (from "help") does
	// not take a bare-infinitive complement the way "help" does, so insert
	// "to" before the bare verb that follows the object noun. Only applies to
	// verb-sense "help" (toks[i] tagged VB*), not noun-sense ("human help").
	for i := range toks {
		rl := strings.ToLower(repl[i])
		if (rl != "aid" && rl != "aids") || !strings.HasPrefix(toks[i].Tag, "VB") {
			continue
		}
		hasObj := false
		for j := i + 1; j < len(toks); j++ {
			tg := toks[j].Tag
			if strings.HasPrefix(tg, "VB") || tg == "MD" {
				if hasObj && repl[j] == "" && (tg == "VB" || tg == "VBP") {
					repl[j] = "to " + toks[j].Text
				}
				break
			}
			if strings.HasPrefix(tg, "NN") {
				hasObj = true
			}
		}
	}

	// strip leading article from a multi-word expansion unless the preceding
	// word is a function word (determiner, preposition, conjunction, verb) or
	// the expansion is clause-initial. A leading "the/a/an" on an expansion
	// acts as a determiner, but when a modifier (adjective, noun adjunct,
	// proper noun, possessive) already heads the noun phrase it is
	// grammatically redundant: "New Cloudflare" → "Newly-wrought the wardens"
	// must become "Newly-wrought wardens", and "Google's the wardens" →
	// "Google's wardens". We KEEP the article after function words: "by the
	// wardens", "and the wardens", "installed the wardens".
	for i := range toks {
		if repl[i] == "" || leadingArticle(repl[i]) == "" {
			continue
		}
		j := prevWord(toks, i)
		if j < 0 || isFunctionWord(toks[j].Tag) {
			continue
		}
		repl[i] = stripLeadingArticle(repl[i])
	}

	// reconstruct, preserving all inter-token text
	var b strings.Builder
	cur := 0
	for i, t := range toks {
		if repl[i] == "" {
			continue
		}
		b.WriteString(text[cur:t.Start])
		if repl[i] != "\u0002" {
			b.WriteString(repl[i])
		} else {
			// swallowed token: also eat one trailing space so we don't
			// leave a gap where the word used to be.
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
func tokenPassAll(text string, rs *ruleset, intensity int, rng *rand.Rand) string {
	locs := rePlaceholder.FindAllStringIndex(text, -1)
	if locs == nil {
		return tokenPass(text, rs, intensity, rng)
	}
	var b strings.Builder
	cur := 0
	for _, loc := range locs {
		b.WriteString(tokenPass(text[cur:loc[0]], rs, intensity, rng))
		b.WriteString(text[loc[0]:loc[1]])
		cur = loc[1]
	}
	b.WriteString(tokenPass(text[cur:], rs, intensity, rng))
	return b.String()
}

// --------------------------------------------------------------------- core

// reHeadingLine matches Markdown heading lines (## Title, ### Section, etc.).
var reHeadingLine = regexp.MustCompile(`(?m)^(#{1,6}\s+.+)$`)

// transformHeadingLine applies word-level replacement to a heading WITHOUT
// structural transforms (no -eth, no interjections, no inversions, no tone
// flourishes). Headings should read as titles, not as archaic prose.
func transformHeadingLine(line string, rs *ruleset, rng *rand.Rand) string {
	if len(rs.phrases) > 0 {
		line = applyPhrases(line, rs.phrases)
	}
	line = applyLiteral(line, rs, rng)
	line = tokenPassAll(line, rs, 1, rng) // intensity 1: word swaps only, no archaization
	line = fixDoubledArticles(line)
	line = fixAbbrevArticles(line)
	line = fixArticles(line)
	return line
}

// protectHeadings finds Markdown heading lines, transforms each with
// word-level replacement only, then stashes the result so the structural
// pipeline (interjections, -eth, inversions, tone) skips them entirely.
func protectHeadings(text string, rs *ruleset, stash *[]string, rng *rand.Rand) string {
	return reHeadingLine.ReplaceAllStringFunc(text, func(m string) string {
		transformed := transformHeadingLine(m, rs, rng)
		*stash = append(*stash, transformed)
		return fmt.Sprintf(placeholder, len(*stash)-1)
	})
}

func transformText(text string, rs *ruleset, lex *Lexicon, opts TransformOptions, rng, flourishRNG *rand.Rand) string {
	if opts.NewsSafe {
		opts.Grammar = false
		opts.Interjections = false
		opts.Flourishes = false
	}
	intensity := opts.Intensity
	if len(rs.phrases) > 0 {
		text = applyPhrases(text, rs.phrases)
	}
	text = applyLiteral(text, rs, rng)
	if opts.Grammar {
		text = expandContractions(text)
		text = removeDoSupport(text)
		text = invertQuestions(text)
	}
	tokenIntensity := intensity
	if !opts.Grammar && tokenIntensity >= 3 {
		tokenIntensity = 2
	}
	text = tokenPassAll(text, rs, tokenIntensity, rng)
	text = fixDoubledArticles(text)
	text = fixAbbrevArticles(text)
	text = fixArticles(text)
	if opts.Grammar {
		text = applyTis(text, rng, 0.5)
		text = applyMineThine(text)
		text = applyInversions(text, rng)
	}
	if opts.Interjections && len(lex.Interjections) > 0 {
		rate := 0.12
		if intensity == 3 {
			rate = 0.2
		}
		text = addInterjections(text, lex.Interjections, flourishRNG, rate)
	}
	if opts.Flourishes {
		text = addToneFlourishes(text, lex, flourishRNG)
	}
	return text
}

// WizardifyMarkdown transmutes markdown text at the given intensity (1-3).
func WizardifyMarkdown(text string, lex *Lexicon, intensity int, seed int64) string {
	w, err := NewWizardifier(lex)
	if err != nil {
		return text
	}
	return w.Markdown(text, DefaultTransformOptions(intensity, seed))
}

func (w *Wizardifier) Markdown(text string, opts TransformOptions) string {
	text, stash := protect(text)
	if opts.NewsSafe {
		text = protectNewsFacts(text, &stash)
	}
	rng := rand.New(rand.NewSource(opts.Seed))
	flourishSeed := opts.FlourishSeed
	if flourishSeed == 0 {
		flourishSeed = opts.Seed
	}
	flourishRNG := rand.New(rand.NewSource(flourishSeed))
	rs, err := w.rulesetFor(opts)
	if err != nil {
		return restoreStash(text, stash)
	}
	text = protectHeadings(text, rs, &stash, rng)
	text = transformText(text, rs, w.lex, opts, rng, flourishRNG)
	text = restoreStash(text, stash)
	if opts.NewsSafe && opts.NewsFlair > 0 {
		text = applyNewsFlair(text, opts.Seed, opts.NewsFlair)
	}
	return text
}

// WizardifyFile picks the right pipeline from the file extension.
func WizardifyFile(path, content string, lex *Lexicon, intensity int, seed int64) (string, error) {
	w, err := NewWizardifier(lex)
	if err != nil {
		return "", err
	}
	return w.File(path, content, DefaultTransformOptions(intensity, seed))
}

func (w *Wizardifier) File(path, content string, opts TransformOptions) (string, error) {
	if _, err := w.rulesetFor(opts); err != nil {
		return "", err
	}
	ext := strings.ToLower(filepath.Ext(path))
	if ext == ".html" || ext == ".htm" {
		return w.HTML(content, opts)
	}
	return w.Markdown(content, opts), nil
}
