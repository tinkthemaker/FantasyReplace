# FantasyReplace 🧙

Transmute any text into wizardly proclamations before posting it to thy public chronicle (blog).

Pure Go. One self-contained binary, two modes: a CLI and a charm.land (Bubble Tea) TUI. The default lexicon is embedded; `-lexicon` can load a custom one. Supports **Markdown** and **HTML** sources.

The engine tags every word with its part of speech (via `jdkato/prose`), so it works on arbitrary text — not just words in the dictionary.

## Quick Example

**Before:**
> I tested the code and found a critical bug in the deploy process. After three hours of debugging, the issue was finally resolved.

**After (intensity 3):**
> I essayed the incantation and discovered a most grievous curse in the unleashing ceremony. After three hours of untangling the matter, the hex hath been thusly mended.

## What the engine does

- **Lemma matching with auto-inflection** — one entry `"crash"` matches crash/crashes/crashed/crashing, and the replacement is inflected to fit (unravels, unraveled, unraveling). Plurals too: `tests` → `trials of truth`.
- **POS disambiguation** — "the test" (noun) → "the trial of truth", "I will test it" (verb) → "I will essay it".
- **Generic archaization (level 3)** — any third-person verb gets -eth ("the dog loveth", "he runneth"), "you" becomes thou/thee by subject/object position, auxiliaries conjugate ("Dost thou want tea?", "thou wouldst enjoy it").
- **Sentence inversions (level 3)** — occasional "thought I", "Never did I see", "A perfect trip it was."
- **Early Modern English grammar (level 3)** — negation without do-support ("I don't know" → "I know not", "Do not trust" → "Trust not"), question inversion ("What do you mean?" → "What meanest thou?"), 'tis/'twas, mine/thine before vowels, said → quoth.
- **Dark-comedy flourishes (level 3)** — exclamations are sentiment-keyed for Pratchett-style reversal: gallows humor only after misfortune ("The gods watched, and did nothing."), outsized pomp only after triumphs ("Bards shall sing of this, badly."), plus rate-limited deadpan asides ("(do not ask what it cost)"). Pools live in `misfortunes`, `triumphs`, and `asides` in the lexicon.
- **Grammar repair** — a/an agreement, subject-verb agreement after plural replacements, "not" placement inside replaced verb phrases.

What stays untouched:

- Markdown: YAML frontmatter, fenced code blocks, inline code, link URLs, bare URLs, HTML tags
- HTML: all tags and attributes, plus the contents of `<code>`, `<pre>`, `<script>`, `<style>`, `<kbd>`, `<samp>`, `<var>`, `<textarea>`

## Intensity levels

1. **Light seasoning** — tier1 term swaps only (computer → thinking engine)
2. **Enchanted** — adds tier2 phrases and occasional interjections ("By my beard,")
3. **Full wizard** — tier3 swaps, generic -eth/-est archaization, thou/thee, inversions, exclamations

## Usage

```
wizardify                          launch the TUI (file picker + live preview)
wizardify post.md -i 3             full wizard, writes post.wizard.md
wizardify page.html -i 2           HTML in, HTML out (text nodes only)
wizardify notes/ -o out/           folder tree (.md, .html, .txt)
wizardify post.md -stdout          print, write nothing
wizardify post.md -in-place        overwrite original
wizardify post.md -seed 42         reproducible flourishes
wizardify news.md -profile news-safe  factual, restricted news styling
wizardify news.md -profile news-safe -flair 2  factual styling with contextual imagery
```

The TUI shows an original/transformed diff. Keys: `1/2/3` preset, `w` toggle news-safe mode, `v` cycle news flair, `g/j/f` toggle grammar/interjections/comedy, `n` reroll flourishes, `tab` switch panes on narrow terminals, `s` save a `.wizard` copy, `esc` back, `q` quit.

`lexicon.json` is found in the current directory or next to the executable. If neither exists, the embedded default is used; override with `-lexicon path`.

## News-safe profile

Use `-profile news-safe` for factual summaries and attributed news copy. With no input files, the same flag opens the TUI directly in news-safe mode.

News-safe mode uses a deliberately small vocabulary allowlist and always disables archaic grammar, interjections, and comedy. It preserves Markdown headings, quotations, blockquotes, figures, dates, currencies, percentages, acronyms, proper names, links, and source-attribution lines. In HTML it also preserves headings, links, quotations, citations, timestamps, and footers.

This profile is a guardrail for already-reviewed summaries; it is not a fact checker. Keep the neutral source-linked summary alongside the styled output and review it before publishing.

`-flair 1..3` optionally appends category-aware imagery to large quantities while preserving the exact figure. Money, audiences, downloads, and distances receive different language; level 3 also permits restrained generic quantity flair. Flair is suppressed in quotations, headings, citations, source lines, links, and reporting about casualties, crime, medicine, war, or disasters.

## Installation

CI builds binaries for Windows, macOS, and Linux. See [INSTALL.md](INSTALL.md) for source and CI-artifact installation instructions.

## Building

```bash
go build -o wizardify .        # or wizardify.exe on Windows
go test -v .                   # run the test suite with verbose output
go test -cover .               # check code coverage
```

## Lexicon format (v2)

Keys are **lemmas** (dictionary forms) — the engine matches every inflection automatically. A value is either a plain string (used for any part of speech) or per-POS senses:

```json
"test":  { "noun": "trial of truth", "verb": "essay" },
"crash": { "noun": "catastrophic unraveling", "verb": "unravel" },
"file":  "parchment"
```

Replacement phrases inflect their head word: first word for verb phrases ("conjure down" → "conjured down"), the word before "of" (else the last word) for noun phrases ("trial of truth" → "trials of truth").

Multi-word keys ("command line", "I think", "figure out") match literally. `phrases` entries use `*` as an object slot: `"committed * to": "inscribed * into the chronicle of"`. `interjections` are prepended to sentences, `exclamations` appended.

The shipped lexicon has ~490 entries covering tech, work, home, travel, food, people, nature, and everyday verbs/adjectives/adverbs. When adding entries, keep replacement values free of words that are themselves keys — multi-word and `phrases` values are re-scanned by the engine.

## Files

- `main.go` — CLI entry point and flags
- `tui.go` — Bubble Tea TUI
- `wizard.go` — engine core: lexicon, protection, token pass
- `pos.go` — POS tagging + tech-jargon tag corrections
- `inflect.go` — lemmatizer, pluralization, verb conjugation
- `archaic.go` — -eth/-est, thou/thee, auxiliary conjugation
- `invert.go` — archaic sentence inversions
- `html.go` — HTML pipeline (transforms text nodes only)
- `lexicon.json` — the replacement dictionary
- `sample-post.md`, `sample-post.html` — test fixtures
- `examples/` — before/after transformation examples at each intensity level

## Contributing & Documentation

- **[INSTALL.md](INSTALL.md)** — Installation guide for all platforms
- **[CONTRIBUTING.md](CONTRIBUTING.md)** — Contribution guidelines and lexicon format
- **[examples/](examples/)** — Example transformations showcasing all intensity levels
- **[LICENSE](LICENSE)** — MIT License
