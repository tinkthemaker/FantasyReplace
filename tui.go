package main

// wizardify-tui: a charm.land Bubble Tea front-end for the wizardify lexicon.

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/filepicker"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const intensityNames = "1 light seasoning  2 enchanted  3 full wizard"

var (
	purple    = lipgloss.Color("99")
	gold      = lipgloss.Color("220")
	dim       = lipgloss.Color("241")
	titleSty  = lipgloss.NewStyle().Foreground(gold).Bold(true)
	headerSty = lipgloss.NewStyle().Foreground(purple).Bold(true)
	keySty    = lipgloss.NewStyle().Foreground(gold)
	dimSty    = lipgloss.NewStyle().Foreground(dim)
	statusSty = lipgloss.NewStyle().Foreground(lipgloss.Color("42"))
	borderSty = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(purple).
			Padding(0, 1)
)

type state int

const (
	statePick state = iota
	statePreview
)

type model struct {
	state     state
	picker    filepicker.Model
	viewport  viewport.Model
	lex       *Lexicon
	file      string
	original  string
	rendered  string
	intensity int
	seed      int64
	showOrig  bool
	status    string
	width     int
	height    int
	ready     bool
}

func newModel(lex *Lexicon, startDir string) model {
	fp := filepicker.New()
	fp.AllowedTypes = []string{".md", ".markdown", ".txt", ".html", ".htm"}
	fp.CurrentDirectory = startDir
	return model{state: statePick, picker: fp, lex: lex,
		intensity: 2, seed: time.Now().UnixNano() % 100000}
}

func (m model) Init() tea.Cmd { return m.picker.Init() }

func (m *model) transform() {
	out, err := WizardifyFile(m.file, m.original, m.lex, m.intensity, m.seed)
	if err != nil {
		m.status = "transmute failed: " + err.Error()
		return
	}
	m.rendered = out
	if m.showOrig {
		m.viewport.SetContent(m.original)
	} else {
		m.viewport.SetContent(m.rendered)
	}
}

func (m *model) openFile(path string) error {
	b, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	m.file = path
	m.original = string(b)
	m.state = statePreview
	m.showOrig = false
	m.viewport = viewport.New(m.width-4, m.height-7)
	m.transform()
	m.status = ""
	return nil
}

func (m *model) save() {
	ext := filepath.Ext(m.file)
	out := strings.TrimSuffix(m.file, ext) + ".wizard" + ext
	if err := os.WriteFile(out, []byte(m.rendered), 0644); err != nil {
		m.status = "save failed: " + err.Error()
		return
	}
	m.status = "inscribed to " + filepath.Base(out)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		m.picker.Height = msg.Height - 6
		if m.state == statePreview {
			m.viewport.Width = msg.Width - 4
			m.viewport.Height = msg.Height - 7
		}
		m.ready = true

	case tea.KeyMsg:
		key := msg.String()
		if key == "ctrl+c" || (key == "q" && m.state == statePreview) {
			return m, tea.Quit
		}
		if m.state == statePick && key == "q" {
			return m, tea.Quit
		}
		if m.state == statePreview {
			switch key {
			case "1", "2", "3":
				m.intensity = int(key[0] - '0')
				m.transform()
			case "n":
				m.seed = time.Now().UnixNano() % 100000
				m.transform()
				m.status = fmt.Sprintf("new seed %d", m.seed)
			case "o":
				m.showOrig = !m.showOrig
				m.transform()
			case "s":
				m.save()
			case "esc":
				m.state = statePick
				m.status = ""
				return m, m.picker.Init()
			}
		}
	}

	var cmd tea.Cmd
	if m.state == statePick {
		m.picker, cmd = m.picker.Update(msg)
		if ok, path := m.picker.DidSelectFile(msg); ok {
			if err := m.openFile(path); err != nil {
				m.status = err.Error()
			}
		}
		return m, cmd
	}
	m.viewport, cmd = m.viewport.Update(msg)
	return m, cmd
}

func (m model) View() string {
	if !m.ready {
		return "summoning..."
	}
	title := titleSty.Render("🧙 wizardify") + dimSty.Render("  ~ transmute thy build notes ~")

	if m.state == statePick {
		help := dimSty.Render("↑/↓ move · enter select · q quit")
		return fmt.Sprintf("%s\n\n%s\n%s\n\n%s", title,
			headerSty.Render("Choose a parchment:"), m.picker.View(), help)
	}

	mode := "transmuted"
	if m.showOrig {
		mode = "original"
	}
	info := fmt.Sprintf("%s · intensity %d · seed %d · %s",
		filepath.Base(m.file), m.intensity, m.seed, mode)
	help := keySty.Render("1/2/3") + dimSty.Render(" intensity · ") +
		keySty.Render("n") + dimSty.Render(" reroll · ") +
		keySty.Render("o") + dimSty.Render(" original · ") +
		keySty.Render("s") + dimSty.Render(" save · ") +
		keySty.Render("esc") + dimSty.Render(" back · ") +
		keySty.Render("q") + dimSty.Render(" quit")
	status := ""
	if m.status != "" {
		status = "  " + statusSty.Render(m.status)
	}
	return fmt.Sprintf("%s  %s%s\n%s\n%s",
		title, headerSty.Render(info), status,
		borderSty.Width(m.width-2).Render(m.viewport.View()), help)
}

func runTUI(lex *Lexicon, dir string) {
	p := tea.NewProgram(newModel(lex, dir), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
