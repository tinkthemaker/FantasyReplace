package main

import (
	"math/rand"
	"os"
	"strings"
	"testing"
)

func lx(t *testing.T) *Lexicon {
	lex, err := LoadLexicon("lexicon.json")
	if err != nil {
		t.Fatal(err)
	}
	return lex
}

func TestMarkdownProtection(t *testing.T) {
	src, err := os.ReadFile("sample-post.md")
	if err != nil {
		t.Fatal(err)
	}
	out := WizardifyMarkdown(string(src), lx(t), 3, 42)
	for _, want := range []string{
		"---\ntitle: Building a keyboard macro pad", // frontmatter untouched
		"qmk compile -kb mypad -km default",         // fence untouched
		"`keymap.c`",                                // inline code untouched
		"(https://qmk.fm)",                          // URL untouched
		"rune-board", "soul-script",                 // swaps applied
	} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q", want)
		}
	}
	for _, bad := range []string{"\x00", "\x01", "an dark", "a hour"} {
		if strings.Contains(out, bad) {
			t.Errorf("contains %q", bad)
		}
	}
}

func TestHTML(t *testing.T) {
	src, err := os.ReadFile("sample-post.html")
	if err != nil {
		t.Fatal(err)
	}
	out, err := WizardifyHTML(string(src), lx(t), 3, 42)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		"qmk compile -kb mypad -km default",
		"<code>keymap.c</code>",
		`<a href="https://qmk.fm">`,
		`console.log("my computer code stays untouched")`,
		"font-family: serif",
		"rune-board", "soul-script",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q", want)
		}
	}
}

func TestHTMLFragment(t *testing.T) {
	frag := `<p>My server crashed during the test.</p><p>Use <code>git push</code> to deploy.</p>`
	out, err := WizardifyHTML(frag, lx(t), 1, 7)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(out, "<html>") || strings.Contains(out, "<body>") {
		t.Errorf("fragment gained document wrapper: %s", out)
	}
	if !strings.Contains(out, "summoning altar unraveled") {
		t.Errorf("crashed not auto-inflected: %s", out)
	}
	if !strings.Contains(out, "<code>git push</code>") {
		t.Errorf("code content changed: %s", out)
	}
}

func TestAutoInflection(t *testing.T) {
	lex := lx(t)
	cases := []struct{ in, want string }{
		{"The server crashes every night.", "unravels"},        // VBZ
		{"Both servers crashed.", "summoning altars unraveled"}, // NNS + VBD
		{"The app keeps crashing.", "unraveling"},               // VBG
		{"I was testing the firmware.", "essaying"},             // verb sense, VBG
		{"We ran three tests.", "trials of truth"},              // head-of-phrase plural
		{"She merged the branches.", "wove together"},           // irregular replacement verb
		{"All my files are gone.", "parchments"},                // simple plural
	}
	for _, c := range cases {
		out := WizardifyMarkdown(c.in, lex, 1, 5)
		if !strings.Contains(out, c.want) {
			t.Errorf("in %q:\n  got  %q\n  want substring %q", c.in, out, c.want)
		}
	}
}

func TestPOSDisambiguation(t *testing.T) {
	lex := lx(t)
	nounOut := WizardifyMarkdown("The test failed again.", lex, 1, 5)
	if !strings.Contains(nounOut, "trial of truth") {
		t.Errorf("noun sense: %q", nounOut)
	}
	verbOut := WizardifyMarkdown("I will test it tomorrow.", lex, 1, 5)
	if !strings.Contains(verbOut, "essay") {
		t.Errorf("verb sense: %q", verbOut)
	}
}

func TestAgreement(t *testing.T) {
	out := WizardifyMarkdown("The code was easy to read.", lx(t), 2, 5)
	if !strings.Contains(out, "incantations were") {
		t.Errorf("agreement not fixed: %q", out)
	}
}

func TestArchaize(t *testing.T) {
	lex := lx(t)
	cases := []struct{ in, want string }{
		{"He walks to town and carries wood.", "strideth"},
		{"He walks to town and carries wood.", "carrieth"},
		{"The sun rises early.", "riseth"},
		{"It runs fast.", "runneth"},
		{"You are a wizard.", "hou art"},
		{"Do you want tea?", "ost thou desire"},
		{"I gave it to you.", "to thee"},
		{"You never know.", "hou never knowest"},
	}
	for _, c := range cases {
		out := WizardifyMarkdown(c.in, lex, 3, 999) // seed avoiding inversions is irrelevant here
		if !strings.Contains(out, c.want) {
			t.Errorf("in %q:\n  got  %q\n  want substring %q", c.in, out, c.want)
		}
	}
}

func TestInversions(t *testing.T) {
	found := false
	for seed := int64(1); seed <= 30 && !found; seed++ {
		rng := rand.New(rand.NewSource(seed))
		out := applyInversions("I thought it would work.", rng)
		if strings.Contains(out, "Thought I") || strings.Contains(out, "thought I") {
			found = true
		}
	}
	if !found {
		t.Error("speech inversion never fired in 30 seeds")
	}
	found = false
	for seed := int64(1); seed <= 30 && !found; seed++ {
		rng := rand.New(rand.NewSource(seed))
		out := applyInversions("I never saw it again.", rng)
		if strings.Contains(out, "Never did I see") {
			found = true
		}
	}
	if !found {
		t.Error("negative fronting never fired in 30 seeds")
	}
}

func TestRealPostRegressions(t *testing.T) {
	lex := lx(t)
	cases := []struct {
		in, want, not string
		intensity     int
	}{
		{"I build a lot of things.", "I forge", "I forging", 1},
		{"This is what you'll find here.", "thou wilt find", "'llest", 3},
		{"You're a wizard, Harry.", "hou art", "'reest", 3},
		{"You've been busy.", "hou hast been", "'veest", 3},
		{"The interesting part is the middle — the bug that ate a Saturday.", "the curse", "—eth", 3},
		{"We test everything twice.", "essay", "trial", 1},
	}
	for _, c := range cases {
		out := WizardifyMarkdown(c.in, lex, c.intensity, 999)
		if !strings.Contains(out, c.want) {
			t.Errorf("in %q:\n  got  %q\n  want substring %q", c.in, out, c.want)
		}
		if c.not != "" && strings.Contains(out, c.not) {
			t.Errorf("in %q:\n  got  %q\n  must NOT contain %q", c.in, out, c.not)
		}
	}
}

func TestEMEGrammar(t *testing.T) {
	lex := lx(t)
	cases := []struct{ in, want string }{
		{"I don't know why.", "I know not why"},
		{"I didn't believe it.", "deemed not"},
		{"Do not trust the cache.", "rust not the hoard"},
		{"What do you mean?", "hat meanest thou?"},
		{"My eyes hurt.", "ine eyes"},
		{"It cannot be undone.", "cannot"}, // modals keep their negation
	}
	for _, c := range cases {
		out := WizardifyMarkdown(c.in, lex, 3, 999)
		if !strings.Contains(out, c.want) {
			t.Errorf("in %q:\n  got  %q\n  want substring %q", c.in, out, c.want)
		}
	}
	// 'tis is rate-limited; check it fires within a few seeds
	found := false
	for s := int64(1); s <= 20 && !found; s++ {
		if strings.Contains(WizardifyMarkdown("It is a small site.", lex, 3, s), "'Tis") {
			found = true
		}
	}
	if !found {
		t.Error("'tis never fired in 20 seeds")
	}
}

func TestToneFlourishes(t *testing.T) {
	lex := lx(t)
	dark := "The deploy failed and the server crashed after six hours."
	bright := "The fix finally worked and the release shipped."
	sawMis, sawTri, crossed := false, false, false
	for s := int64(1); s <= 60; s++ {
		dOut := WizardifyMarkdown(dark, lex, 3, s)
		bOut := WizardifyMarkdown(bright, lex, 3, s)
		for _, m := range lex.Misfortunes {
			if strings.Contains(dOut, m) {
				sawMis = true
			}
			if strings.Contains(bOut, m) {
				crossed = true // gallows humor after a triumph = bug
			}
		}
		for _, m := range lex.Triumphs {
			if strings.Contains(bOut, m) {
				sawTri = true
			}
		}
	}
	if !sawMis {
		t.Error("misfortune pool never used on dark sentence")
	}
	if !sawTri {
		t.Error("triumph pool never used on bright sentence")
	}
	if crossed {
		t.Error("misfortune flourish appeared after triumph sentence")
	}
}

// Edge case tests
func TestEmptyInput(t *testing.T) {
	lex := lx(t)
	out := WizardifyMarkdown("", lex, 1, 0)
	if out != "" {
		t.Errorf("empty input should produce empty output, got %q", out)
	}
}

func TestWhitespaceOnly(t *testing.T) {
	lex := lx(t)
	out := WizardifyMarkdown("   \n\n  ", lex, 1, 0)
	if strings.TrimSpace(out) != "" {
		t.Errorf("whitespace-only input should stay whitespace-only, got %q", out)
	}
}

func TestSpecialCharacters(t *testing.T) {
	lex := lx(t)
	cases := []string{
		"The café has good coffee.",
		"Naïve implementations fail.",
		"Test: 123 → success (100%).",
		"Price: $99.99 or €85.",
	}
	for _, c := range cases {
		out := WizardifyMarkdown(c, lex, 1, 42)
		// Should not crash and output should be non-empty
		if len(out) == 0 {
			t.Errorf("special chars input produced empty output: %q", c)
		}
	}
}

func TestIntensityLevels(t *testing.T) {
	lex := lx(t)
	input := "I test the code and see if it works well. It is good."
	i1 := WizardifyMarkdown(input, lex, 1, 42)
	i2 := WizardifyMarkdown(input, lex, 2, 42)
	i3 := WizardifyMarkdown(input, lex, 3, 42)

	// Intensity 1 should be mildest
	if len(i1) < len(input)-10 { // allow some shrink but not too much
		t.Errorf("intensity 1 too aggressive: %q", i1)
	}

	// Intensity 3 should typically be longer (added archaic constructions)
	if len(i3) <= len(i1) {
		t.Errorf("intensity 3 should generally be longer than intensity 1")
	}

	// All should be non-empty
	if len(i1) == 0 || len(i2) == 0 || len(i3) == 0 {
		t.Error("intensities produced empty output")
	}
}

func TestMultilineContent(t *testing.T) {
	lex := lx(t)
	input := `First paragraph has a test.

Second paragraph also tests things.

Third paragraph: we test thoroughly.`
	out := WizardifyMarkdown(input, lex, 1, 42)
	// Should preserve paragraph structure
	lines := strings.Split(out, "\n")
	if len(lines) < 5 {
		t.Errorf("multiline structure not preserved: %q", out)
	}
}

func TestFrontmatterPreservation(t *testing.T) {
	lex := lx(t)
	input := `---
title: Test Article
author: John Tester
---

This is the test content that should be modified.`
	out := WizardifyMarkdown(input, lex, 1, 42)
	if !strings.Contains(out, "---\ntitle: Test Article\nauthor: John Tester\n---") {
		t.Errorf("frontmatter not preserved: %q", out)
	}
	if !strings.Contains(out, "trial of truth") {
		t.Errorf("content not transformed: %q", out)
	}
}

func TestCodeBlockPreservation(t *testing.T) {
	lex := lx(t)
	input := "Here is some code:\n\n```go\nfunc test(t *testing.T) {\n\tif err != nil {\n\t\tt.Fatal(err)\n\t}\n}\n```\n\nThe code above should not change."
	out := WizardifyMarkdown(input, lex, 3, 42)
	if !strings.Contains(out, "func test(t *testing.T)") {
		t.Errorf("code block content changed: %q", out)
	}
	// At intensity 3, "code" can become "incantations" - just verify some transformation occurred
	if strings.Contains(out, "should not change") {
		t.Errorf("text outside code block not transformed: %q", out)
	}
}

func TestInlineCodePreservation(t *testing.T) {
	lex := lx(t)
	input := "Use the `test` function in your code to test things."
	out := WizardifyMarkdown(input, lex, 1, 42)
	if !strings.Contains(out, "`test`") {
		t.Errorf("inline code `test` was modified: %q", out)
	}
}

func TestURLPreservation(t *testing.T) {
	lex := lx(t)
	input := "Check out https://github.com/test for test frameworks."
	out := WizardifyMarkdown(input, lex, 1, 42)
	if !strings.Contains(out, "https://github.com/test") {
		t.Errorf("URL was modified: %q", out)
	}
}

func TestLinkPreservation(t *testing.T) {
	lex := lx(t)
	input := "See the [test guide](https://example.com/test-guide) for details."
	out := WizardifyMarkdown(input, lex, 1, 42)
	if !strings.Contains(out, "(https://example.com/test-guide)") {
		t.Errorf("link URL was modified: %q", out)
	}
}

func TestArticleAgreement(t *testing.T) {
	lex := lx(t)
	cases := []struct {
		in  string
		bad string // Should NOT contain this
	}{
		{"an test", "an test"},
		{"a apple", "a apple"},
	}
	for _, c := range cases {
		out := WizardifyMarkdown(c.in, lex, 2, 42)
		if strings.Contains(out, c.bad) {
			t.Errorf("article agreement not fixed in %q: %q", c.in, out)
		}
	}
}

func TestConsistency(t *testing.T) {
	lex := lx(t)
	input := "I test and test again. The test must pass. Testing is important."
	// Same input with same seed should produce same output
	out1 := WizardifyMarkdown(input, lex, 2, 42)
	out2 := WizardifyMarkdown(input, lex, 2, 42)
	if out1 != out2 {
		t.Errorf("same input with same seed produced different outputs:\n%q\nvs\n%q", out1, out2)
	}
}

func TestCaseSensitivity(t *testing.T) {
	lex := lx(t)
	// Test a simple, single-word replacement for consistent case preservation
	lower := WizardifyMarkdown("the crash happened.", lex, 1, 42)
	upper := WizardifyMarkdown("THE CRASH HAPPENED.", lex, 1, 42)

	// All should transform, preserving case on replacement words
	if !strings.Contains(lower, "catastrophic unraveling") {
		t.Errorf("lowercase transform failed: %q", lower)
	}
	if !strings.Contains(upper, "CATASTROPHIC UNRAVELING") {
		t.Errorf("uppercase not preserved: %q", upper)
	}
}
