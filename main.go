package main

// wizardify: transmute build notes (markdown or HTML) into wizard-speak.
//
//   wizardify                        launch the charm.land TUI
//   wizardify post.md -i 3           CLI: transmute a file
//   wizardify notes/ -o out/         CLI: transmute a folder
//
// Both modes share lexicon.json (looked up in the current directory,
// then next to the executable).

import (
	"flag"
	"fmt"
	"hash/fnv"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

var sourceExts = map[string]bool{
	".md": true, ".markdown": true, ".txt": true, ".html": true, ".htm": true,
}

// Overridden by release builds through -ldflags. Keeping useful development
// defaults makes locally-built binaries self-identifying too.
var (
	version   = "dev"
	commit    = "unknown"
	buildDate = "unknown"
)

func defaultLexiconPath() string {
	if _, err := os.Stat("lexicon.json"); err == nil {
		return "lexicon.json"
	}
	if exe, err := os.Executable(); err == nil {
		p := filepath.Join(filepath.Dir(exe), "lexicon.json")
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}
	return ""
}

func collectFiles(inputs []string) ([]string, error) {
	sources, err := collectSourceFiles(inputs)
	if err != nil {
		return nil, err
	}
	files := make([]string, len(sources))
	for i, source := range sources {
		files[i] = source.Path
	}
	return files, nil
}

type sourceFile struct {
	Path string
	Rel  string
}

func collectSourceFiles(inputs []string) ([]sourceFile, error) {
	var files []sourceFile
	seen := map[string]bool{}
	for _, inp := range inputs {
		info, err := os.Stat(inp)
		if err != nil {
			return nil, fmt.Errorf("%s not found", inp)
		}
		if info.IsDir() {
			err := filepath.WalkDir(inp, func(path string, entry fs.DirEntry, walkErr error) error {
				if walkErr != nil {
					return walkErr
				}
				if entry.IsDir() || !sourceExts[strings.ToLower(filepath.Ext(entry.Name()))] {
					return nil
				}
				rel, err := filepath.Rel(inp, path)
				if err != nil {
					return err
				}
				abs, _ := filepath.Abs(path)
				if !seen[abs] {
					seen[abs] = true
					files = append(files, sourceFile{Path: path, Rel: rel})
				}
				return nil
			})
			if err != nil {
				return nil, err
			}
		} else {
			if !sourceExts[strings.ToLower(filepath.Ext(inp))] {
				return nil, fmt.Errorf("%s has an unsupported file type", inp)
			}
			abs, _ := filepath.Abs(inp)
			if !seen[abs] {
				seen[abs] = true
				files = append(files, sourceFile{Path: inp, Rel: filepath.Base(inp)})
			}
		}
	}
	sort.Slice(files, func(i, j int) bool { return files[i].Path < files[j].Path })
	return files, nil
}

func seedForFile(seed int64, rel string, multiple bool) int64 {
	if !multiple {
		return seed
	}
	h := fnv.New64a()
	_, _ = h.Write([]byte(filepath.ToSlash(rel)))
	return seed ^ int64(h.Sum64())
}

func outPath(in string) string {
	ext := filepath.Ext(in)
	return strings.TrimSuffix(in, ext) + ".wizard" + ext
}

func normalizeProfile(value string) (string, error) {
	profile := strings.ToLower(strings.TrimSpace(value))
	if profile == "news" {
		profile = "news-safe"
	}
	if profile != "wizard" && profile != "news-safe" {
		return "", fmt.Errorf("profile must be wizard or news-safe")
	}
	return profile, nil
}

func validateOutputMode(stdout, inPlace, dryRun, check bool, outdir string) error {
	if dryRun && check {
		return fmt.Errorf("-dry-run and -check are mutually exclusive")
	}
	if stdout && (inPlace || dryRun || check || outdir != "") {
		return fmt.Errorf("-stdout cannot be combined with -in-place, -dry-run, -check, or -o")
	}
	if inPlace && (dryRun || check || outdir != "") {
		return fmt.Errorf("-in-place cannot be combined with -dry-run, -check, or -o")
	}
	if dryRun && outdir != "" {
		return fmt.Errorf("-dry-run cannot be combined with -o")
	}
	if check && outdir != "" {
		return fmt.Errorf("-check cannot be combined with -o")
	}
	return nil
}

// reorderArgs lets flags appear after positional arguments
// (Go's flag package normally stops parsing at the first positional).
func reorderArgs(args []string) []string {
	boolFlags := map[string]bool{
		"-stdout": true, "--stdout": true,
		"-in-place": true, "--in-place": true,
		"-version": true, "--version": true,
		"-dry-run": true, "--dry-run": true,
		"-check": true, "--check": true,
		"-h": true, "--help": true,
	}
	var flags, pos []string
	for i := 0; i < len(args); i++ {
		a := args[i]
		if strings.HasPrefix(a, "-") {
			flags = append(flags, a)
			if !boolFlags[a] && !strings.Contains(a, "=") && i+1 < len(args) {
				i++
				flags = append(flags, args[i])
			}
		} else {
			pos = append(pos, a)
		}
	}
	return append(flags, pos...)
}

func main() {
	intensity := flag.Int("i", 1, "intensity 1-3")
	outdir := flag.String("o", "", "output directory (default: next to input)")
	lexPath := flag.String("lexicon", defaultLexiconPath(), "path to lexicon JSON (default: local file if present, otherwise embedded)")
	seed := flag.Int64("seed", 0, "RNG seed for reproducible flourishes (0 = random)")
	profile := flag.String("profile", "wizard", "style profile: wizard or news-safe")
	flair := flag.Int("flair", 0, "contextual news flair 0-3 (requires news-safe)")
	showVersion := flag.Bool("version", false, "print version and build information")
	toStdout := flag.Bool("stdout", false, "print result, write nothing")
	inPlace := flag.Bool("in-place", false, "overwrite input files")
	dryRun := flag.Bool("dry-run", false, "show what would be written, without changing files")
	check := flag.Bool("check", false, "exit 1 if any input would change, without writing files")
	dir := flag.String("dir", ".", "starting directory for the TUI file picker")
	flag.CommandLine.Parse(reorderArgs(os.Args[1:]))
	if *showVersion {
		fmt.Printf("wizardify %s (commit %s, built %s)\n", version, commit, buildDate)
		return
	}

	if *intensity < 1 || *intensity > 3 {
		fmt.Fprintln(os.Stderr, "intensity must be 1, 2, or 3")
		os.Exit(1)
	}
	normalizedProfile, err := normalizeProfile(*profile)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	*profile = normalizedProfile
	if *flair < 0 || *flair > 3 {
		fmt.Fprintln(os.Stderr, "flair must be between 0 and 3")
		os.Exit(1)
	}
	if *flair > 0 && *profile != "news-safe" {
		fmt.Fprintln(os.Stderr, "-flair requires -profile news-safe")
		os.Exit(1)
	}
	if err := validateOutputMode(*toStdout, *inPlace, *dryRun, *check, *outdir); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	lex, err := LoadLexicon(*lexPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "could not load lexicon: %v\n", err)
		os.Exit(1)
	}
	wizardifier, err := NewWizardifier(lex)
	if err != nil {
		fmt.Fprintf(os.Stderr, "could not compile lexicon: %v\n", err)
		os.Exit(1)
	}

	if flag.NArg() == 0 {
		if *dryRun || *check {
			fmt.Fprintln(os.Stderr, "-dry-run and -check require input files")
			os.Exit(1)
		}
		runTUI(wizardifier, *dir, *profile == "news-safe", *flair)
		return
	}

	s := *seed
	if s == 0 {
		s = time.Now().UnixNano() % 100000
	}
	files, err := collectSourceFiles(flag.Args())
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
	if len(files) == 0 {
		fmt.Fprintln(os.Stderr, "error: no supported files found")
		os.Exit(1)
	}
	if *toStdout && len(files) != 1 {
		fmt.Fprintln(os.Stderr, "-stdout requires exactly one input file")
		os.Exit(1)
	}

	outputs := map[string]string{}
	for _, source := range files {
		if *toStdout || *inPlace || *dryRun || *check || *outdir == "" {
			continue
		}
		out := filepath.Join(*outdir, outPath(source.Rel))
		key, _ := filepath.Abs(out)
		if previous, exists := outputs[key]; exists {
			fmt.Fprintf(os.Stderr, "output collision: %s and %s both map to %s\n", previous, source.Path, out)
			os.Exit(1)
		}
		outputs[key] = source.Path
	}

	checkChanged := false
	for _, source := range files {
		b, err := os.ReadFile(source.Path)
		if err != nil {
			fmt.Fprintln(os.Stderr, "error:", err)
			os.Exit(1)
		}
		fileSeed := seedForFile(s, source.Rel, len(files) > 1)
		opts := DefaultTransformOptions(*intensity, fileSeed)
		if *profile == "news-safe" {
			opts = DefaultNewsSafeOptions(fileSeed)
			opts.NewsFlair = *flair
		}
		result, err := wizardifier.File(source.Path, string(b), opts)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error transmuting %s: %v\n", source.Path, err)
			os.Exit(1)
		}
		switch {
		case *toStdout:
			fmt.Print(result)
		case *check:
			if result != string(b) {
				checkChanged = true
				fmt.Println("would change:", source.Path)
			}
		case *dryRun:
			fmt.Printf("would transmute: %s -> %s\n", source.Path, outPath(source.Path))
		case *inPlace:
			if err := safeWriteFile(source.Path, []byte(result), 0644); err != nil {
				fmt.Fprintln(os.Stderr, "error:", err)
				os.Exit(1)
			}
			fmt.Println("transmuted in place:", source.Path)
		default:
			out := outPath(source.Path)
			if *outdir != "" {
				out = filepath.Join(*outdir, outPath(source.Rel))
			}
			if err := safeWriteFile(out, []byte(result), 0644); err != nil {
				fmt.Fprintln(os.Stderr, "error:", err)
				os.Exit(1)
			}
			fmt.Printf("transmuted: %s -> %s\n", source.Path, out)
		}
	}
	if *check && checkChanged {
		os.Exit(1)
	}
}
