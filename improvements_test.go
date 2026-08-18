package main

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"sync"
	"testing"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/ansi"
)

func TestHeadingRestoresNestedProtectedContent(t *testing.T) {
	lex := lx(t)
	input := "# Testing ``code ` span`` and [the link](https://example.com/path_(one))\n"
	out := WizardifyMarkdown(input, lex, 3, 42)
	if strings.Contains(out, "\x00WIZ") {
		t.Fatalf("placeholder leaked into output: %q", out)
	}
	if !strings.Contains(out, "``code ` span``") {
		t.Errorf("inline code changed: %q", out)
	}
	if !strings.Contains(out, "](https://example.com/path_(one))") {
		t.Errorf("nested link destination changed: %q", out)
	}
}

func TestCRLFFrontmatterPreserved(t *testing.T) {
	lex := lx(t)
	frontmatter := "---\r\ntitle: Test code\r\ntags: [test]\r\n---\r\n"
	out := WizardifyMarkdown(frontmatter+"The test passed.", lex, 2, 42)
	if !strings.HasPrefix(out, frontmatter) {
		t.Fatalf("CRLF frontmatter changed:\n%q", out)
	}
}

func TestHTMLKeepsGrammarContextAcrossInlineTags(t *testing.T) {
	lex := lx(t)
	out, err := WizardifyHTML(`<p><strong>You</strong> edit the code.</p>`, lex, 3, 42)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(out, "HTMLNODE") || strings.ContainsRune(out, '\x00') {
		t.Fatalf("HTML node marker leaked: %q", out)
	}
	low := strings.ToLower(out)
	if !strings.Contains(low, "<strong>thou</strong>") || !strings.Contains(low, "editest") {
		t.Fatalf("inline tags broke grammatical context: %q", out)
	}
}

func TestEmbeddedLexicon(t *testing.T) {
	lex, err := LoadLexicon("")
	if err != nil {
		t.Fatal(err)
	}
	if len(lex.Tier1) < 100 {
		t.Fatalf("embedded lexicon looks incomplete: %d tier-1 entries", len(lex.Tier1))
	}
}

func TestNewsSafeProfileProtectsFacts(t *testing.T) {
	lex := lx(t)
	wizard, err := NewWizardifier(lex)
	if err != nil {
		t.Fatal(err)
	}
	input := `# Anthropic Releases Fable 5

Anthropic said the government approved a new AI model on June 23, 2026. The decision is important.

"The new model is important," CEO Dario Amodei said.

Source: [Reuters — original article](https://example.com/news?id=123)`
	out := wizard.Markdown(input, DefaultNewsSafeOptions(42))
	for _, preserved := range []string{
		"# Anthropic Releases Fable 5",
		"Anthropic", "government", "AI model", "June 23, 2026",
		`"The new model is important,"`, "CEO Dario Amodei",
		"Source: [Reuters — original article](https://example.com/news?id=123)",
	} {
		if !strings.Contains(out, preserved) {
			t.Errorf("news-safe output did not preserve %q:\n%s", preserved, out)
		}
	}
	for _, unsafe := range []string{"the crown", "arcane mind", "effigy", "simulacrum"} {
		if strings.Contains(strings.ToLower(out), unsafe) {
			t.Errorf("news-safe output introduced factual substitution %q:\n%s", unsafe, out)
		}
	}
	if !strings.Contains(out, "newly-wrought") || !strings.Contains(out, "weighty") {
		t.Errorf("news-safe allowlisted styling was not applied:\n%s", out)
	}
}

func TestNormalizeNewsSafeProfile(t *testing.T) {
	for input, want := range map[string]string{
		"wizard": "wizard", "NEWS": "news-safe", " news-safe ": "news-safe",
	} {
		got, err := normalizeProfile(input)
		if err != nil || got != want {
			t.Errorf("normalizeProfile(%q) = %q, %v; want %q", input, got, err, want)
		}
	}
	if _, err := normalizeProfile("satire"); err == nil {
		t.Fatal("invalid profile was accepted")
	}
}

func TestValidateOutputMode(t *testing.T) {
	valid := [][5]interface{}{
		{false, false, false, false, ""},
		{false, false, true, false, ""},
		{false, false, false, true, ""},
		{false, false, false, false, "out"},
	}
	for _, mode := range valid {
		if err := validateOutputMode(mode[0].(bool), mode[1].(bool), mode[2].(bool), mode[3].(bool), mode[4].(string)); err != nil {
			t.Errorf("valid mode rejected: %v: %v", mode, err)
		}
	}
	invalid := [][5]interface{}{
		{true, false, false, false, "out"},
		{false, true, false, false, "out"},
		{false, false, true, true, ""},
		{false, false, true, false, "out"},
		{false, false, false, true, "out"},
	}
	for _, mode := range invalid {
		if err := validateOutputMode(mode[0].(bool), mode[1].(bool), mode[2].(bool), mode[3].(bool), mode[4].(string)); err == nil {
			t.Errorf("invalid mode accepted: %v", mode)
		}
	}
}

func TestReorderArgsKeepsBooleanFlagsFromConsumingInputs(t *testing.T) {
	got := reorderArgs([]string{"first.md", "-dry-run", "second.md", "-check"})
	want := []string{"-dry-run", "-check", "first.md", "second.md"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("reorderArgs = %#v, want %#v", got, want)
	}
}

func TestNewsSafeForcesFlourishesOff(t *testing.T) {
	lex := lx(t)
	wizard, err := NewWizardifier(lex)
	if err != nil {
		t.Fatal(err)
	}
	opts := DefaultNewsSafeOptions(42)
	opts.Grammar, opts.Interjections, opts.Flourishes = true, true, true
	out := wizard.Markdown("You edit the new story. The government failed.", opts)
	if strings.Contains(strings.ToLower(out), "thou") || strings.Contains(strings.ToLower(out), "the crown") {
		t.Fatalf("unsafe layers escaped news-safe enforcement: %q", out)
	}
	for _, pool := range [][]string{lex.Interjections, lex.Exclamations, lex.Misfortunes, lex.Triumphs, lex.Asides} {
		for _, phrase := range pool {
			if strings.Contains(out, phrase) {
				t.Fatalf("news-safe output contains flourish %q: %q", phrase, out)
			}
		}
	}
}

func TestNewsSafeHTMLProtectsAttributionElements(t *testing.T) {
	lex := lx(t)
	wizard, err := NewWizardifier(lex)
	if err != nil {
		t.Fatal(err)
	}
	input := `<h1>Anthropic Releases Fable 5</h1><p>Anthropic announced a new AI model on June 23, 2026.</p><blockquote>The new model is important.</blockquote><footer><a href="https://example.com">Original story</a></footer>`
	out, err := wizard.HTML(input, DefaultNewsSafeOptions(42))
	if err != nil {
		t.Fatal(err)
	}
	for _, preserved := range []string{
		"<h1>Anthropic Releases Fable 5</h1>",
		"Anthropic", "AI model", "June 23, 2026",
		"<blockquote>The new model is important.</blockquote>",
		`<a href="https://example.com">Original story</a>`,
	} {
		if !strings.Contains(out, preserved) {
			t.Errorf("news-safe HTML did not preserve %q:\n%s", preserved, out)
		}
	}
	if !strings.Contains(out, "newly-wrought") {
		t.Errorf("news-safe HTML did not style eligible prose: %s", out)
	}
}

func TestLexiconValidation(t *testing.T) {
	if _, err := NewWizardifier(&Lexicon{}); err == nil {
		t.Fatal("empty lexicon was accepted")
	}
	bad := &Lexicon{Tier1: map[string]Entry{"test": {Any: "trial|"}}}
	if _, err := NewWizardifier(bad); err == nil {
		t.Fatal("empty replacement variant was accepted")
	}
}

func TestCollectSourceFilesRecursive(t *testing.T) {
	root := t.TempDir()
	nested := filepath.Join(root, "posts", "drafts")
	if err := os.MkdirAll(nested, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(nested, "post.md"), []byte("test"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(nested, "ignore.png"), []byte("x"), 0644); err != nil {
		t.Fatal(err)
	}
	files, err := collectSourceFiles([]string{root})
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 1 || files[0].Rel != filepath.Join("posts", "drafts", "post.md") {
		t.Fatalf("unexpected recursive files: %#v", files)
	}
}

func TestSafeWriteReplacesAndPreservesMode(t *testing.T) {
	path := filepath.Join(t.TempDir(), "post.md")
	if err := os.WriteFile(path, []byte("before"), 0600); err != nil {
		t.Fatal(err)
	}
	beforeInfo, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := safeWriteFile(path, []byte("after"), 0644); err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(b) != "after" {
		t.Fatalf("got %q", b)
	}
	if info, err := os.Stat(path); err != nil || info.Mode().Perm() != beforeInfo.Mode().Perm() {
		t.Fatalf("mode was not preserved: %v, %v", info, err)
	}
}

func TestHighlightedDiff(t *testing.T) {
	left, right, changes := highlightedDiff("The test passed.", "The trial passed.")
	if changes != 1 {
		t.Fatalf("changes = %d, want 1", changes)
	}
	if !strings.Contains(left, "test") || !strings.Contains(right, "trial") {
		t.Fatal("diff content was lost")
	}
}

func TestWrapAlignedDiff(t *testing.T) {
	left, right, _ := highlightedDiff(
		"The government issued a surprise export control affecting several models.",
		"The crown issued a most unexpected proclamation affecting several enchanted effigies.")
	left, right = wrapAlignedDiff(left, right, 24)
	leftLines, rightLines := strings.Split(left, "\n"), strings.Split(right, "\n")
	if len(leftLines) != len(rightLines) {
		t.Fatalf("wrapped panes lost alignment: %d lines vs %d", len(leftLines), len(rightLines))
	}
	if len(leftLines) < 2 {
		t.Fatal("long content did not wrap")
	}
	for i, line := range append(leftLines, rightLines...) {
		if width := ansi.StringWidth(line); width > 24 {
			t.Fatalf("line %d is %d cells wide: %q", i, width, ansi.Strip(line))
		}
	}
}

func TestTUIRewrapsOnResize(t *testing.T) {
	lex := lx(t)
	wizard, err := NewWizardifier(lex)
	if err != nil {
		t.Fatal(err)
	}
	m := newModel(wizard, ".", false)
	m.state, m.ready = statePreview, true
	m.width, m.height = 160, 30
	m.original = strings.Repeat("original text ", 20)
	m.rendered = strings.Repeat("wizardly proclamation ", 20)
	m.originalVP, m.renderedVP = viewport.New(20, 3), viewport.New(20, 3)
	m.resizeViewports()
	m.applyRendered(m.rendered)
	wideLines := m.originalVP.TotalLineCount()
	updated, _ := m.Update(tea.WindowSizeMsg{Width: 100, Height: 30})
	narrow := updated.(model)
	if narrow.originalVP.TotalLineCount() <= wideLines {
		t.Fatalf("narrower viewport did not add wrapped lines: wide=%d narrow=%d", wideLines, narrow.originalVP.TotalLineCount())
	}
}

func TestTUITransformsAsynchronously(t *testing.T) {
	lex := lx(t)
	wizard, err := NewWizardifier(lex)
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "post.md")
	if err := os.WriteFile(path, []byte("The test passed."), 0644); err != nil {
		t.Fatal(err)
	}
	m := newModel(wizard, filepath.Dir(path), false)
	m.width, m.height, m.ready = 120, 30, true
	cmd, err := m.openFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !m.working || cmd == nil {
		t.Fatal("transform did not start asynchronously")
	}
	updated, _ := m.Update(cmd())
	got := updated.(model)
	if got.working || got.rendered == "" || got.changes == 0 {
		t.Fatalf("transform result not applied: working=%v rendered=%q changes=%d", got.working, got.rendered, got.changes)
	}
	if _, ok := updated.(tea.Model); !ok {
		t.Fatal("updated model does not implement tea.Model")
	}
}

func TestTUINewsSafeProfileIsEnforced(t *testing.T) {
	lex := lx(t)
	wizard, err := NewWizardifier(lex)
	if err != nil {
		t.Fatal(err)
	}
	m := newModel(wizard, ".", true)
	m.state = statePreview
	if !m.newsSafe || m.grammar || m.interject || m.comedy || !m.options().NewsSafe {
		t.Fatalf("unexpected news-safe defaults: %#v", m.options())
	}
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'g'}})
	locked := updated.(model)
	if locked.grammar || !strings.Contains(locked.status, "locked off") {
		t.Fatalf("grammar escaped news-safe lock: grammar=%v status=%q", locked.grammar, locked.status)
	}
	updated, _ = locked.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'w'}})
	normal := updated.(model)
	if normal.newsSafe || normal.options().NewsSafe {
		t.Fatal("news-safe toggle did not return to the normal preset")
	}
}

func TestNewsFlairIsContextualAndPreservesQuantities(t *testing.T) {
	money := applyNewsFlair("The company raised $12 million in funding.", 42, 2)
	if !strings.Contains(money, "$12 million—") || !strings.Contains(money, "hoard") && !strings.Contains(money, "coin") {
		t.Fatalf("money did not receive money-specific flair: %q", money)
	}
	audience := applyNewsFlair("The service reached 12 million active users.", 42, 2)
	if !strings.Contains(audience, "12 million active users—") || !strings.Contains(audience, "multitude") && !strings.Contains(audience, "kingdoms") {
		t.Fatalf("audience did not receive audience-specific flair: %q", audience)
	}
	if strings.Contains(applyNewsFlair("Millions supported the proposal.", 42, 2), "—") {
		t.Fatal("generic quantity received flair below level 3")
	}
	if got := applyNewsFlair("Millions supported the proposal.", 42, 3); !strings.Contains(got, "Millions—") {
		t.Fatalf("level 3 generic quantity received no flair: %q", got)
	}
}

func TestNewsFlairSuppressesSensitiveAndProtectedCopy(t *testing.T) {
	input := strings.Join([]string{
		"# $12 million settlement",
		"The disaster killed 12 million people.",
		`The CEO said, "$12 million was raised."`,
		"Source: [12 million users](https://example.com/story)",
	}, "\n")
	if got := applyNewsFlair(input, 42, 3); got != input {
		t.Fatalf("protected or sensitive news copy changed:\n%s", got)
	}
}

func TestNewsFlairIntegrationAndValidation(t *testing.T) {
	lex := lx(t)
	wizard, err := NewWizardifier(lex)
	if err != nil {
		t.Fatal(err)
	}
	opts := DefaultNewsSafeOptions(42)
	opts.NewsFlair = 2
	out, err := wizard.File("story.md", "The company raised $12 million in funding.", opts)
	if err != nil || !strings.Contains(out, "$12 million—") {
		t.Fatalf("integrated news flair failed: %q, %v", out, err)
	}
	bad := DefaultTransformOptions(2, 42)
	bad.NewsFlair = 1
	if _, err := wizard.File("story.md", "Millions arrived.", bad); err == nil {
		t.Fatal("news flair was accepted outside news-safe mode")
	}
}

func TestTUICyclesNewsFlair(t *testing.T) {
	lex := lx(t)
	wizard, err := NewWizardifier(lex)
	if err != nil {
		t.Fatal(err)
	}
	m := newModel(wizard, ".", true)
	m.state = statePreview
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'v'}})
	got := updated.(model)
	if got.newsFlair != 1 || got.options().NewsFlair != 1 {
		t.Fatalf("news flair did not cycle: model=%d options=%d", got.newsFlair, got.options().NewsFlair)
	}
}

func TestWizardifierConcurrent(t *testing.T) {
	lex := lx(t)
	wizard, err := NewWizardifier(lex)
	if err != nil {
		t.Fatal(err)
	}
	input := "You edit the code, test it, and deploy it. The test passed."
	opts := DefaultTransformOptions(3, 42)
	want, err := wizard.File("post.md", input, opts)
	if err != nil {
		t.Fatal(err)
	}
	var wg sync.WaitGroup
	errs := make(chan error, 16)
	for range 16 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			got, err := wizard.File("post.md", input, opts)
			if err != nil {
				errs <- err
				return
			}
			if got != want {
				errs <- &outputMismatch{got: got, want: want}
			}
		}()
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		t.Fatal(err)
	}
}

type outputMismatch struct{ got, want string }

func (e *outputMismatch) Error() string {
	return "concurrent output differed:\n" + e.got + "\nwant:\n" + e.want
}

func TestPOSModelReuseAllocationBudget(t *testing.T) {
	lex := lx(t)
	wizard, err := NewWizardifier(lex)
	if err != nil {
		t.Fatal(err)
	}
	input, err := os.ReadFile("sample-post.md")
	if err != nil {
		t.Fatal(err)
	}
	opts := DefaultTransformOptions(3, 42)
	_, _ = wizard.File("sample-post.md", string(input), opts) // warm shared model
	allocs := testing.AllocsPerRun(3, func() {
		_, _ = wizard.File("sample-post.md", string(input), opts)
	})
	if allocs > 10_000 {
		t.Fatalf("steady-state allocations regressed: %.0f", allocs)
	}
}

func BenchmarkWizardifierMarkdown(b *testing.B) {
	lex, err := LoadLexicon("lexicon.json")
	if err != nil {
		b.Fatal(err)
	}
	wizard, err := NewWizardifier(lex)
	if err != nil {
		b.Fatal(err)
	}
	input, err := os.ReadFile("sample-post.md")
	if err != nil {
		b.Fatal(err)
	}
	opts := DefaultTransformOptions(3, 42)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = wizard.File("sample-post.md", string(input), opts)
	}
}
