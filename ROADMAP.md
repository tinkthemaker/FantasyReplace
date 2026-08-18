# Wizardify Product Roadmap

This checklist is ordered by user value and implementation dependency. A milestone
is complete only when its code, tests, documentation, and packaged executable are
updated together.

## 1. Release foundation

- [x] Add `--version` output with version, commit, and build date.
- [x] Make release builds reproducible and embed build metadata.
- [x] Smoke-test packaged executables in CI.
- [x] Generate checksums and retain consistent platform names.
- [x] Ensure every distributed binary is rebuilt from the same revision.

## 2. Contextual news flair

- [x] Add `-flair 0..3` for the news-safe profile.
- [x] Distinguish money, audiences, downloads, distance, and generic quantities.
- [x] Preserve the exact quantity and append, rather than replace, imagery.
- [x] Suppress flair for casualties, crime, medicine, war, and disasters.
- [x] Keep quotations, headings, citations, links, and source lines untouched.
- [x] Add a TUI control and visibly report the active flair level.
- [x] Keep variations deterministic for a given seed.

## 3. Review and control TUI

- [ ] Represent transformations as structured changes with source ranges and reasons.
- [ ] Navigate to the next and previous change.
- [ ] Accept or reject individual changes and change categories.
- [ ] Add undo and redo.
- [ ] Support direct editing of transformed text.
- [ ] Copy output to the clipboard.
- [ ] Add Save As and Open Output Folder actions.
- [ ] Add search, recent files, first-run help, and a keyboard-help overlay.
- [ ] Add accessible themes and improve narrow-terminal behavior.
- [ ] Replace line-indexed comparison with hunk-aware diff alignment.

## 4. Profiles and factual safety

- [ ] Add named profile files with vocabulary, grammar, comedy, and flair controls.
- [ ] Add personal replacements and a `never_replace` list.
- [ ] Compare numbers, dates, currencies, percentages, URLs, quotations, acronyms,
      headings, and citations before and after news transformations.
- [ ] Fail closed when a protected fact changes.
- [ ] Produce a human-readable and JSON fact-preservation report.
- [ ] Explain that news-safe protects text but does not verify source truth.

## 5. Engine hardening

- [ ] Add transformation provenance and confidence levels.
- [ ] Add repetition and style-density limits.
- [ ] Improve Unicode, smart-punctuation, emoji, and name handling.
- [ ] Move Markdown protection toward an AST-based pipeline.
- [ ] Avoid unnecessary HTML reserialization and formatting churn.
- [ ] Add graceful context cancellation for long transformations.

## 6. CLI and batch workflows

- [ ] Support stdin and conventional stdout pipelines.
- [x] Add `--dry-run` and `--check`; JSON reporting remains pending.
- [ ] Add include/exclude globs and configurable recursion.
- [ ] Continue batch jobs after per-file failures and print a final summary.
- [ ] Add bounded parallel folder processing and progress reporting.
- [ ] Add optional in-place backups and shell completion.
- [ ] Extract the engine into an importable Go package.

## 7. Quality and distribution

- [ ] Add a curated golden-output corpus.
- [ ] Add fuzz tests for malformed markup and placeholder leakage.
- [ ] Add compiled-CLI and TUI snapshot tests.
- [ ] Track performance benchmarks in CI.
- [ ] Add vulnerability scanning, an SBOM, and dependency-license checks.
- [ ] Publish installers or package-manager entries where practical.
- [ ] Add signed releases and optional private update checks.
- [ ] Unify the FantasyReplace/Wizardify repository, module, and product naming.
