package main

import _ "embed"

// embeddedLexicon is the built-in fallback used by release binaries. A local
// lexicon.json (or -lexicon) may still override it during development.
//
//go:embed lexicon.json
var embeddedLexicon []byte
