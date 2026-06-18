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
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

var sourceExts = map[string]bool{
	".md": true, ".markdown": true, ".txt": true, ".html": true, ".htm": true,
}

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
	return "lexicon.json"
}

func collectFiles(inputs []string) ([]string, error) {
	var files []string
	for _, inp := range inputs {
		info, err := os.Stat(inp)
		if err != nil {
			return nil, fmt.Errorf("%s not found", inp)
		}
		if info.IsDir() {
			entries, err := os.ReadDir(inp)
			if err != nil {
				return nil, err
			}
			for _, e := range entries {
				if !e.IsDir() && sourceExts[strings.ToLower(filepath.Ext(e.Name()))] {
					files = append(files, filepath.Join(inp, e.Name()))
				}
			}
		} else {
			files = append(files, inp)
		}
	}
	sort.Strings(files)
	return files, nil
}

func outPath(in string) string {
	ext := filepath.Ext(in)
	return strings.TrimSuffix(in, ext) + ".wizard" + ext
}

// reorderArgs lets flags appear after positional arguments
// (Go's flag package normally stops parsing at the first positional).
func reorderArgs(args []string) []string {
	boolFlags := map[string]bool{
		"-stdout": true, "--stdout": true,
		"-in-place": true, "--in-place": true,
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
	lexPath := flag.String("lexicon", defaultLexiconPath(), "path to lexicon JSON")
	seed := flag.Int64("seed", 0, "RNG seed for reproducible flourishes (0 = random)")
	toStdout := flag.Bool("stdout", false, "print result, write nothing")
	inPlace := flag.Bool("in-place", false, "overwrite input files")
	dir := flag.String("dir", ".", "starting directory for the TUI file picker")
	flag.CommandLine.Parse(reorderArgs(os.Args[1:]))

	if *intensity < 1 || *intensity > 3 {
		fmt.Fprintln(os.Stderr, "intensity must be 1, 2, or 3")
		os.Exit(1)
	}
	lex, err := LoadLexicon(*lexPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "could not load lexicon: %v\n", err)
		os.Exit(1)
	}

	if flag.NArg() == 0 {
		runTUI(lex, *dir)
		return
	}

	s := *seed
	if s == 0 {
		s = time.Now().UnixNano() % 100000
	}
	files, err := collectFiles(flag.Args())
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
	for _, f := range files {
		b, err := os.ReadFile(f)
		if err != nil {
			fmt.Fprintln(os.Stderr, "error:", err)
			os.Exit(1)
		}
		result, err := WizardifyFile(f, string(b), lex, *intensity, s)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error transmuting %s: %v\n", f, err)
			os.Exit(1)
		}
		switch {
		case *toStdout:
			fmt.Print(result)
		case *inPlace:
			if err := os.WriteFile(f, []byte(result), 0644); err != nil {
				fmt.Fprintln(os.Stderr, "error:", err)
				os.Exit(1)
			}
			fmt.Println("transmuted in place:", f)
		default:
			out := outPath(f)
			if *outdir != "" {
				os.MkdirAll(*outdir, 0755)
				out = filepath.Join(*outdir, filepath.Base(out))
			}
			if err := os.WriteFile(out, []byte(result), 0644); err != nil {
				fmt.Fprintln(os.Stderr, "error:", err)
				os.Exit(1)
			}
			fmt.Printf("transmuted: %s -> %s\n", f, out)
		}
	}
}
