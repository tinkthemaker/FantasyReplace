# Contributing to FantasyReplace

Thank you for your interest in contributing! This guide will help you get started.

## Code of Conduct

Be respectful, inclusive, and constructive. We welcome all contributions regardless of experience level.

## Getting Started

### Fork & Clone
```bash
git clone https://github.com/yourusername/wizardify.git
cd wizardify
```

### Build & Test
```bash
go build -o wizardify .
go test ./...
```

### Run the TUI
```bash
./wizardify
```

## Development

### File Organization
- **main.go** — CLI entry point and flag parsing
- **wizard.go** — Core transformation engine and lexicon loading
- **pos.go** — Part-of-speech tagging and corrections
- **inflect.go** — Lemmatization and verb/noun inflection
- **archaic.go** — Archaic English transformations (thou/thee, -eth, etc.)
- **invert.go** — Sentence inversions
- **html.go** — HTML parsing and text-node transformation
- **tui.go** — Bubble Tea interactive interface
- **tone.go** — Sentiment-based flourishes
- **lexicon.json** — Replacement dictionary

### Code Style
- Follow standard Go conventions
- Use `gofmt` to format code
- Run `go vet` before committing
- Keep functions focused and well-commented
- Write tests for new features

### Adding Features

1. **New transformation type?** Add to `wizard.go` and call from `transformText()`
2. **New intensity-3 feature?** Implement in a new file (e.g., `fancy.go`)
3. **New lexicon entries?** Edit `lexicon.json` — see [Lexicon Format](#lexicon-format)
4. **Bug fix?** Add a regression test first

### Testing

```bash
go test -v ./...
go test -cover ./...      # check coverage
go test -race ./...        # race condition detection
```

**Write tests for:**
- New transformations
- Edge cases (empty strings, special characters)
- POS tagging correctness
- Lexicon loading and validation
- HTML/Markdown protection and restoration

Example:
```go
func TestInflectNounPhrase(t *testing.T) {
	tests := []struct {
		phrase   string
		plural   bool
		expected string
	}{
		{"trial of truth", false, "trial of truth"},
		{"trial of truth", true, "trials of truth"},
	}
	for _, tt := range tests {
		got := inflectNounPhrase(tt.phrase, tt.plural)
		if got != tt.expected {
			t.Errorf("inflectNounPhrase(%q, %v) = %q, want %q", 
				tt.phrase, tt.plural, got, tt.expected)
		}
	}
}
```

## Lexicon Format

The `lexicon.json` file defines transformations. Keys are **lemmas** (dictionary forms).

### Single-word entries
```json
"test": "trial of truth"
```
This matches: test, tests, tested, testing, etc.

### POS-specific entries
```json
"crash": {
  "noun": "catastrophic unraveling",
  "verb": "unravel"
}
```

### Multi-word entries
```json
"command line": "incantation specification"
```
Match as a literal phrase (case-insensitive, whole-word).

### Phrases with object slots
```json
"committed * to": "inscribed * into"
```
The `*` is replaced with the captured middle words.

### Guidelines
- Keep replacements **concise** (1-4 words usually)
- Avoid **circular references** — don't use a key as part of another replacement value
- Use **common phrases** — test tier1 entries with real documents
- Balance **whimsy with readability** — tier1 should be clever but clear
- For **tier3**, embrace drama and complexity

### Adding Entries
1. Identify the lemma (base form)
2. Determine part(s) of speech
3. Choose witty but understandable replacements
4. Test: `wizardify test.md -i 1 -stdout`

Example contribution:
```json
"deploy": {
  "verb": "unleash upon the realm",
  "noun": "unleashing upon the realm"
},
"bug": {
  "noun": "hexed sprite"
}
```

## Submitting Changes

### 1. Create a branch
```bash
git checkout -b feature/your-feature-name
# or
git checkout -b fix/your-bug-fix
```

### 2. Make changes and commit
```bash
git add .
git commit -m "Add feature X"
```

Use clear, concise commit messages. Reference issues if applicable.

### 3. Run tests and lint
```bash
go test ./...
golangci-lint run
```

### 4. Push and open a pull request
```bash
git push origin feature/your-feature-name
```

On GitHub, click "New Pull Request" and describe your changes.

### PR Guidelines
- **Title**: Brief description of changes
- **Description**: 
  - What problem does this solve?
  - How do you solve it?
  - Any breaking changes?
- **Tests**: Include tests for your changes
- **Documentation**: Update README/CONTRIBUTING if needed

## Reporting Issues

Found a bug or have an idea?

1. **Check existing issues** — might already be reported
2. **Describe clearly**:
   - What did you do?
   - What happened?
   - What did you expect?
   - OS, Go version, wizardify version
3. **Include examples** — sample text, error messages
4. **Label appropriately** — `bug`, `enhancement`, `documentation`, etc.

## Recognition

Contributors are recognized in:
- This file's contributors section (come back later!)
- Release notes for significant contributions

## Questions?

- Open an [Issue](https://github.com/yourusername/wizardify/issues) for questions
- Check [README.md](README.md) for general usage
- Discuss in PR comments for implementation details

Thank you for making wizardify better! 🧙
