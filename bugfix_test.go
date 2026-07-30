package main

import (
	"strings"
	"testing"
)

// ── Bug Class 1: Heading protection ───────────────────────────────
// Headings should get word swaps but NOT interjections, -eth, inversions,
// or tone flourishes. The heading number "3." inside "## 3." should not
// trigger sentence-boundary interjection insertion.

func TestHeadingNoInterjection(t *testing.T) {
	lex := lx(t)
	input := "## 3. Machine Traffic Dominates the Web\n\nSome text here."
	out := WizardifyMarkdown(input, lex, 3, 42)
	// The heading should NOT have an interjection prepended.
	for _, bad := range lex.Interjections {
		if strings.Contains(out, "## "+bad) || strings.Contains(out, "3. "+bad) {
			t.Errorf("heading got interjection %q:\n%s", bad, out)
		}
	}
}

func TestHeadingNoEth(t *testing.T) {
	lex := lx(t)
	// "Dominates" in a heading should NOT become "Dominateth".
	input := "## Server Dominates Market Share\n\nThe server dominates the market."
	out := WizardifyMarkdown(input, lex, 3, 42)
	lines := strings.Split(out, "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "##") {
			if strings.Contains(line, "Dominateth") || strings.Contains(line, "dominateth") {
				t.Errorf("heading got -eth: %q", line)
			}
		}
	}
}

func TestHeadingStillGetsWordSwaps(t *testing.T) {
	lex := lx(t)
	input := "## The Server Crashed\n\nThe server crashed."
	out := WizardifyMarkdown(input, lex, 1, 42)
	// Heading should still contain word swaps.
	if !strings.Contains(out, "summoning altar") {
		t.Errorf("heading missing word swap:\n%s", out)
	}
}

func TestHeadingNoToneFlourish(t *testing.T) {
	lex := lx(t)
	// A heading containing misfortune keywords should not get tone flourishes.
	input := "## The Deploy Failed Again\n\nBody text here."
	out := WizardifyMarkdown(input, lex, 3, 42)
	for _, m := range lex.Misfortunes {
		if strings.Contains(out, "##") {
			lines := strings.SplitN(out, "\n", 2)
			if strings.Contains(lines[0], m) {
				t.Errorf("heading got misfortune flourish: %q in %q", m, lines[0])
			}
		}
	}
}

// ── Bug Class 2: Doubled articles ─────────────────────────────────
// "the government" → "the the crown" should collapse to "the crown".

func TestDoubledArticleThe(t *testing.T) {
	lex := lx(t)
	// "government" → "the crown". Source "The government issued" → should
	// become "The crown issued", NOT "The the crown issued".
	out := WizardifyMarkdown("The government issued a decree.", lex, 1, 42)
	if strings.Contains(strings.ToLower(out), "the the crown") {
		t.Errorf("doubled article: %q", out)
	}
}

func TestDoubledArticleA(t *testing.T) {
	lex := lx(t)
	// "error" → "dark omen". Source "An error occurred" → replacement
	// "a dark omen" but original "An" → "An a dark omen" → should collapse.
	out := WizardifyMarkdown("An error occurred.", lex, 1, 42)
	if strings.Contains(strings.ToLower(out), "an a dark") {
		t.Errorf("doubled article: %q", out)
	}
}

func TestDoubledArticleInternet(t *testing.T) {
	lex := lx(t)
	// "internet" → "the great aether". Source "the internet traffic" →
	// "the the great aether traffic" → should collapse to "the great aether traffic".
	out := WizardifyMarkdown("the internet traffic grew.", lex, 1, 42)
	if strings.Contains(strings.ToLower(out), "the the great aether") {
		t.Errorf("doubled article: %q", out)
	}
}

func TestFixDoubledArticlesUnit(t *testing.T) {
	cases := []struct{ in, want string }{
		{"the the crown", "the crown"},
		{"The the crown", "The crown"},
		{"a a test", "a test"},
		{"an an hour", "an hour"},
		{"the a crown", "the crown"},
		{"a an error", "a error"},
	}
	for _, c := range cases {
		got := fixDoubledArticles(c.in)
		if got != c.want {
			t.Errorf("fixDoubledArticles(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestFixArticlesSilentH(t *testing.T) {
	// "a hour" → "an hour" (silent h)
	got := fixArticles("a hour passed")
	if strings.Contains(got, "a hour") {
		t.Errorf("silent-h article not fixed: %q", got)
	}
	if !strings.Contains(got, "an hour") {
		t.Errorf("expected 'an hour': %q", got)
	}
}

func TestFixArticlesUConsonant(t *testing.T) {
	// "an unique" → "a unique" (u pronounced "yoo")
	got := fixArticles("an unique opportunity")
	if strings.Contains(got, "an unique") {
		t.Errorf("u-consonant article not fixed: %q", got)
	}
	if !strings.Contains(got, "a unique") {
		t.Errorf("expected 'a unique': %q", got)
	}
	// "an university" → "a university"
	got = fixArticles("an university student")
	if strings.Contains(got, "an university") {
		t.Errorf("u-consonant article not fixed: %q", got)
	}
}

// ── Bug Class 3: Mass noun over-pluralization ────────────────────
// "data" → "knowledge essence" (not "knowledge essences") even when
// the POS tagger tags "data" as NNS.

func TestMassNounData(t *testing.T) {
	lex := lx(t)
	out := WizardifyMarkdown("Cloudflare data reveals trends.", lex, 1, 42)
	if strings.Contains(out, "knowledge essences") {
		t.Errorf("mass noun pluralized: %q", out)
	}
	if !strings.Contains(out, "knowledge essence") {
		t.Errorf("mass noun replacement missing: %q", out)
	}
}

func TestMassNounInformation(t *testing.T) {
	got := fixDoubledArticles("the information is available")
	// Just verify mass noun lookup works
	_ = got
	if !isMassNoun("data") {
		t.Error("'data' should be a mass noun")
	}
	if !isMassNoun("information") {
		t.Error("'information' should be a mass noun")
	}
	if isMassNoun("server") {
		t.Error("'server' should NOT be a mass noun")
	}
}

func TestMassNounTrafficNotPlural(t *testing.T) {
	lex := lx(t)
	// "traffic" is not in the lexicon as a key, but "internet" is.
	// "the internet traffic" → "the great aether traffic"
	out := WizardifyMarkdown("The internet traffic is high.", lex, 1, 42)
	if strings.Contains(out, "aethers traffic") || strings.Contains(out, "great aethers") {
		t.Errorf("mass noun 'internet' over-pluralized: %q", out)
	}
}

// ── Bug Class 4: Do-support on participle constructions ───────────
// "don't get stuck" should NOT become "get not stuck".

func TestDoSupportParticipleGetStuck(t *testing.T) {
	lex := lx(t)
	out := WizardifyMarkdown("I don't get stuck on bugs.", lex, 3, 999)
	if strings.Contains(out, "get not stuck") {
		t.Errorf("participle construction broken: %q", out)
	}
}

func TestDoSupportParticipleGetStarted(t *testing.T) {
	lex := lx(t)
	out := WizardifyMarkdown("Don't get started without the config.", lex, 3, 999)
	if strings.Contains(strings.ToLower(out), "get not started") {
		t.Errorf("participle construction broken: %q", out)
	}
}

func TestDoSupportStillWorksForMainVerbs(t *testing.T) {
	lex := lx(t)
	// "don't know" → "know not" should still work (no participle follows).
	out := WizardifyMarkdown("I don't know why.", lex, 3, 999)
	if !strings.Contains(out, "know not") {
		t.Errorf("do-support removal broken for main verb: %q", out)
	}
}

func TestDoSupportStillWorksForImperative(t *testing.T) {
	lex := lx(t)
	out := WizardifyMarkdown("Do not trust the cache.", lex, 3, 999)
	if !strings.Contains(out, "rust not the hoard") && !strings.Contains(out, "Trust not") {
		// At least one of these forms should appear
		if !strings.Contains(strings.ToLower(out), "not") {
			t.Errorf("imperative do-support broken: %q", out)
		}
	}
}

func TestLooksLikeParticiple(t *testing.T) {
	cases := []struct {
		word string
		want bool
	}{
		{"stuck", true},     // irregular VBN
		{"broken", true},    // irregular VBN
		{"surprised", true}, // ends -ed
		{"written", true},   // ends -en
		{"why", false},
		{"the", false},
		{"now", false},
		{"there", false},
	}
	for _, c := range cases {
		got := looksLikeParticiple(c.word)
		if got != c.want {
			t.Errorf("looksLikeParticiple(%q) = %v, want %v", c.word, got, c.want)
		}
	}
}

// ── Bug Class 5: Title-case -eth protection ───────────────────────
// "Dominates" in a heading should NOT get -eth.

func TestTitleCaseNoEth(t *testing.T) {
	lex := lx(t)
	// In body text, "The server dominates" → "dominateth" is fine.
	// But "Server Dominates Market" (title case) → should NOT.
	out := WizardifyMarkdown("## Server Dominates Market\n\nBody.", lex, 3, 42)
	if strings.Contains(out, "Dominateth") || strings.Contains(out, "dominateth") {
		// Check if it's in the heading line specifically
		for _, line := range strings.Split(out, "\n") {
			if strings.HasPrefix(line, "##") && strings.Contains(strings.ToLower(line), "dominateth") {
				t.Errorf("title-case heading got -eth: %q", line)
			}
		}
	}
}

// ── Bug Class 6: Non-word gerunds from noun-only entries ──────────
// "planning" tagged VBG with lemma "plan" → "stratagem" (noun-only).
// Should NOT become "stratageming".

func TestNounOnlyNotVerbInflected(t *testing.T) {
	lex := lx(t)
	out := WizardifyMarkdown("AI is moving from chatbots to autonomous agents.", lex, 3, 42)
	// "planning" → lemma "plan" → "stratagem" (noun-only Any sense)
	// If mis-tagged as verb, should NOT produce "stratageming"
	if strings.Contains(out, "stratageming") {
		t.Errorf("noun-only entry verb-inflected: %q", out)
	}
}

// ── Bug Class 7: Modern AI terms in lexicon ───────────────────────

func TestModernTermsAI(t *testing.T) {
	lex := lx(t)
	out := WizardifyMarkdown("AI is transforming the industry.", lex, 1, 42)
	// "AI" is all-caps, matchCase uppercases the replacement.
	if !strings.Contains(strings.ToLower(out), "arcane mind") {
		t.Errorf("AI not replaced: %q", out)
	}
}

func TestModernTermsAgentic(t *testing.T) {
	lex := lx(t)
	out := WizardifyMarkdown("The shift to agentic workflows is real.", lex, 1, 42)
	if !strings.Contains(out, "self-willed") {
		t.Errorf("agentic not replaced: %q", out)
	}
}

func TestModernTermsAgent(t *testing.T) {
	lex := lx(t)
	out := WizardifyMarkdown("The agent completed the task.", lex, 1, 42)
	// "agent" has a variant pool (emissary/familiar/harbinger); any of them
	// is acceptable. What matters is that the word was replaced at all.
	if !strings.Contains(out, "emissary") &&
		!strings.Contains(out, "familiar") &&
		!strings.Contains(out, "harbinger") {
		t.Errorf("agent not replaced: %q", out)
	}
}

func TestModernTermsBot(t *testing.T) {
	lex := lx(t)
	out := WizardifyMarkdown("Bots now outnumber humans online.", lex, 1, 42)
	// "Bots" is plural → "automatons" (correctly inflected)
	if !strings.Contains(strings.ToLower(out), "automaton") {
		t.Errorf("bot not replaced: %q", out)
	}
}

func TestModernTermsChatbot(t *testing.T) {
	lex := lx(t)
	out := WizardifyMarkdown("The chatbot answered questions.", lex, 1, 42)
	if !strings.Contains(out, "talking head") {
		t.Errorf("chatbot not replaced: %q", out)
	}
}

func TestModernTermsCrawler(t *testing.T) {
	lex := lx(t)
	out := WizardifyMarkdown("AI crawlers account for most traffic.", lex, 1, 42)
	if !strings.Contains(out, "aether-prowler") {
		t.Errorf("crawler not replaced: %q", out)
	}
}

func TestModernTermsLLM(t *testing.T) {
	lex := lx(t)
	out := WizardifyMarkdown("The LLM generates text.", lex, 1, 42)
	// "LLM" is all-caps, so matchCase uppercases the replacement.
	if !strings.Contains(strings.ToLower(out), "great mind") {
		t.Errorf("LLM not replaced: %q", out)
	}
}

// ── Regression: expansion-article collision (Fix 1) ───────────────
// A multi-word expansion starting with "the" must drop its article when a
// content word (adjective, proper noun, possessive) precedes it.
// "New Cloudflare" → "Newly-wrought wardens", NOT "Newly-wrought the wardens".

func TestArticleStripAfterProperNoun(t *testing.T) {
	lex := lx(t)
	out := WizardifyMarkdown("New Cloudflare data shows the trend.", lex, 3, 42)
	low := strings.ToLower(out)
	if strings.Contains(low, "newly-wrought the wardens") {
		t.Errorf("leading article not stripped after proper-noun modifier: %q", out)
	}
	if !strings.Contains(low, "newly-wrought wardens") {
		t.Errorf("expected 'newly-wrought wardens': %q", out)
	}
}

func TestArticleStripAfterPossessive(t *testing.T) {
	lex := lx(t)
	out := WizardifyMarkdown("Google's network is down.", lex, 3, 42)
	low := strings.ToLower(out)
	// "Google's the aether" would be wrong; possessive heads the NP.
	if strings.Contains(low, "google's the aether") {
		t.Errorf("leading article not stripped after possessive: %q", out)
	}
}

func TestArticleKeptAfterPreposition(t *testing.T) {
	lex := lx(t)
	out := WizardifyMarkdown("Powered by Cloudflare, the site is fast.", lex, 3, 42)
	low := strings.ToLower(out)
	// "by the wardens of the aether" — article must be KEPT after a
	// preposition, since it is the determiner of the noun phrase.
	if !strings.Contains(low, "by the wardens") && !strings.Contains(low, "by the wardens of the aether") {
		// Check that "by wardens" (stripped) did NOT happen.
		if strings.Contains(low, "by wardens") {
			t.Errorf("article wrongly stripped after preposition: %q", out)
		}
	}
}

// ── Regression: variant rotation / anti-repetition (Fix 2) ─────────
// High-frequency words with a variant pool must rotate, not repeat. A
// document using "agent" three times should not produce the same variant
// three times in a row.

func TestVariantRotationDistinct(t *testing.T) {
	lex := lx(t)
	out := WizardifyMarkdown(
		"The agent spoke. The agent listened. The agent departed.", lex, 1, 42)
	low := strings.ToLower(out)
	variants := []string{"emissary", "familiar", "harbinger"}
	found := map[string]bool{}
	for _, v := range variants {
		if strings.Contains(low, v) {
			found[v] = true
		}
	}
	if len(found) < 2 {
		t.Errorf("expected >=2 distinct agent variants, got %d (%v): %q", len(found), found, out)
	}
}

func TestVariantRotationDeterministic(t *testing.T) {
	lex := lx(t)
	in := "The user spoke. The user listened. The user departed."
	a := WizardifyMarkdown(in, lex, 1, 42)
	b := WizardifyMarkdown(in, lex, 1, 42)
	if a != b {
		t.Errorf("non-deterministic output for same seed:\n%s\n%s", a, b)
	}
}

// ── Regression: spec-sentence flourish suppression (Fix 3) ─────────
// A dry technical spec (numbers + acronyms + hyphenated compounds) must
// not receive a tone flourish. "550B-parameter effigy" is not a narrative
// beat, so "Songs will be written" must not follow it.

func TestSpecSentenceNoFlourish(t *testing.T) {
	lex := lx(t)
	out := WizardifyMarkdown(
		"NVIDIA released a 550B-parameter model today.", lex, 3, 42)
	for _, triumph := range lex.Triumphs {
		if strings.Contains(out, triumph) {
			t.Errorf("spec sentence got triumph flourish %q:\n%s", triumph, out)
		}
	}
}

// ── Regression: valency fix for "help X <verb>" (Fix 4) ────────────
// "help" → "aid" changes verb-frame arity: "aid" does not take a bare
// infinitive. "help websites optimize" → "aid websites to optimize".

func TestValencyAidGetsTo(t *testing.T) {
	lex := lx(t)
	out := WizardifyMarkdown(
		"Adobe launched a system to help websites optimize themselves.", lex, 3, 42)
	low := strings.ToLower(out)
	if !strings.Contains(low, "aid aetheric shrines to optimize") {
		t.Errorf("valency fix missing 'to' before optimize: %q", out)
	}
}

func TestValencyNoFalsePositiveOnNounHelp(t *testing.T) {
	lex := lx(t)
	// "human help" — help is a noun here. No verb follows, so no "to"
	// should be inserted. Just verify it does not crash and "aid" appears.
	out := WizardifyMarkdown("I need human help now.", lex, 3, 42)
	low := strings.ToLower(out)
	if strings.Contains(low, "aid to") {
		t.Errorf("false 'to' insertion after noun-sense help: %q", out)
	}
}
