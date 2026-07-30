package main

// Bubble Tea review UI for wizardify.

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/filepicker"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
)

var intensityNames = [...]string{"", "light", "enchanted", "full wizard"}

var (
	purple      = lipgloss.Color("99")
	gold        = lipgloss.Color("220")
	dim         = lipgloss.Color("241")
	titleSty    = lipgloss.NewStyle().Foreground(gold).Bold(true)
	headerSty   = lipgloss.NewStyle().Foreground(purple).Bold(true)
	keySty      = lipgloss.NewStyle().Foreground(gold)
	dimSty      = lipgloss.NewStyle().Foreground(dim)
	statusSty   = lipgloss.NewStyle().Foreground(lipgloss.Color("42"))
	errorSty    = lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
	removedSty  = lipgloss.NewStyle().Foreground(lipgloss.Color("203")).Bold(true)
	addedSty    = lipgloss.NewStyle().Foreground(lipgloss.Color("84")).Bold(true)
	borderSty   = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(purple).Padding(0, 1)
	activePanel = lipgloss.NewStyle().Foreground(gold).Bold(true)
	reDiffToken = regexp.MustCompile(`[[:alnum:]_]+|[^[:space:][:alnum:]_]`)
)

type state int

const (
	statePick state = iota
	statePreview
)

type transformKey struct {
	intensity                  int
	seed, flourishSeed         int64
	grammar, interject, comedy bool
	newsSafe                   bool
	newsFlair                  int
}

type transformedMsg struct {
	id       uint64
	key      transformKey
	text     string
	err      error
	duration time.Duration
}

type model struct {
	state      state
	picker     filepicker.Model
	originalVP viewport.Model
	renderedVP viewport.Model
	wizard     *Wizardifier
	file       string
	original   string
	rendered   string
	intensity  int
	seed       int64
	flourish   int64
	grammar    bool
	interject  bool
	comedy     bool
	newsSafe   bool
	newsFlair  int
	status     string
	statusErr  bool
	width      int
	height     int
	ready      bool
	working    bool
	requestID  uint64
	changes    int
	activePane int
	confirm    bool
	cache      map[transformKey]string
}

func newModel(wizard *Wizardifier, startDir string, newsSafe bool, initialFlair ...int) model {
	fp := filepicker.New()
	fp.AllowedTypes = []string{".md", ".markdown", ".txt", ".html", ".htm"}
	fp.CurrentDirectory = startDir
	m := model{state: statePick, picker: fp, wizard: wizard, intensity: 2,
		seed: time.Now().UnixNano(), flourish: time.Now().UnixNano() + 1,
		cache: map[transformKey]string{}}
	m.applyPreset(2)
	if newsSafe {
		m.applyNewsSafe()
		if len(initialFlair) > 0 {
			m.newsFlair = initialFlair[0]
		}
	}
	return m
}

func (m model) Init() tea.Cmd { return m.picker.Init() }

func (m *model) applyPreset(intensity int) {
	m.newsSafe = false
	m.newsFlair = 0
	m.intensity = intensity
	m.grammar = intensity >= 3
	m.interject = intensity >= 2
	m.comedy = intensity >= 3
}

func (m *model) applyNewsSafe() {
	m.newsSafe = true
	m.intensity = 2
	m.grammar, m.interject, m.comedy = false, false, false
}

func (m model) currentKey() transformKey {
	return transformKey{intensity: m.intensity, seed: m.seed, flourishSeed: m.flourish,
		grammar: m.grammar, interject: m.interject, comedy: m.comedy, newsSafe: m.newsSafe, newsFlair: m.newsFlair}
}

func (m model) options() TransformOptions {
	if m.newsSafe {
		opts := DefaultNewsSafeOptions(m.seed)
		opts.NewsFlair = m.newsFlair
		return opts
	}
	return TransformOptions{Intensity: m.intensity, Seed: m.seed, FlourishSeed: m.flourish,
		Grammar: m.grammar, Interjections: m.interject, Flourishes: m.comedy}
}

func (m *model) resizeViewports() {
	h := maxInt(3, m.height-9)
	w := maxInt(20, m.width-4)
	if m.width >= 100 {
		w = maxInt(20, (m.width-11)/2)
	}
	m.originalVP.Width, m.originalVP.Height = w, h
	m.renderedVP.Width, m.renderedVP.Height = w, h
}

func (m *model) startTransform() tea.Cmd {
	key := m.currentKey()
	if cached, ok := m.cache[key]; ok {
		m.applyRendered(cached)
		m.status = "cached preview"
		m.statusErr = false
		m.working = false
		return nil
	}
	m.requestID++
	id := m.requestID
	m.working = true
	m.confirm = false
	m.status = "transmuting…"
	m.statusErr = false
	wizard, path, original, opts := m.wizard, m.file, m.original, m.options()
	return func() tea.Msg {
		started := time.Now()
		text, err := wizard.File(path, original, opts)
		return transformedMsg{id: id, key: key, text: text, err: err, duration: time.Since(started)}
	}
}

func (m *model) applyRendered(text string) {
	m.rendered = text
	original, rendered, changes := highlightedDiff(m.original, text)
	m.changes = changes
	original, rendered = wrapAlignedDiff(original, rendered, m.originalVP.Width)
	m.originalVP.SetContent(original)
	m.renderedVP.SetContent(rendered)
}

func (m *model) refreshWrappedContent() {
	if m.rendered == "" {
		return
	}
	leftOffset, rightOffset := m.originalVP.YOffset, m.renderedVP.YOffset
	m.applyRendered(m.rendered)
	m.originalVP.YOffset = minInt(leftOffset, maxInt(0, m.originalVP.TotalLineCount()-m.originalVP.Height))
	m.renderedVP.YOffset = minInt(rightOffset, maxInt(0, m.renderedVP.TotalLineCount()-m.renderedVP.Height))
}

func (m *model) openFile(path string) (tea.Cmd, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	m.file, m.original = path, string(b)
	m.rendered, m.status = "", ""
	m.state, m.activePane = statePreview, 0
	m.cache = map[transformKey]string{}
	m.originalVP, m.renderedVP = viewport.New(20, 3), viewport.New(20, 3)
	m.resizeViewports()
	m.originalVP.SetContent(m.original)
	return m.startTransform(), nil
}

func (m *model) save() {
	if m.working || m.rendered == "" {
		m.status, m.statusErr = "wait for the preview before saving", true
		return
	}
	ext := filepath.Ext(m.file)
	out := strings.TrimSuffix(m.file, ext) + ".wizard" + ext
	if _, err := os.Stat(out); err == nil && !m.confirm {
		m.confirm = true
		m.status, m.statusErr = "output exists; press s again to replace it", true
		return
	}
	if err := safeWriteFile(out, []byte(m.rendered), 0644); err != nil {
		m.status, m.statusErr = "save failed: "+err.Error(), true
		return
	}
	m.confirm = false
	m.status, m.statusErr = "saved "+out, false
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height, m.ready = msg.Width, msg.Height, true
		m.picker.Height = maxInt(3, msg.Height-6)
		if m.state == statePreview {
			m.resizeViewports()
			m.refreshWrappedContent()
		}

	case transformedMsg:
		if msg.id != m.requestID {
			return m, nil
		}
		m.working = false
		if msg.err != nil {
			m.status, m.statusErr = "transmute failed: "+msg.err.Error(), true
			return m, nil
		}
		m.cache[msg.key] = msg.text
		m.applyRendered(msg.text)
		m.status = fmt.Sprintf("%d changes · %s", m.changes, friendlyDuration(msg.duration))
		m.statusErr = false
		return m, nil

	case tea.KeyMsg:
		key := msg.String()
		if key == "ctrl+c" || key == "q" {
			return m, tea.Quit
		}
		if m.state == statePreview {
			switch key {
			case "1", "2", "3":
				m.applyPreset(int(key[0] - '0'))
				return m, m.startTransform()
			case "n":
				if m.newsSafe {
					m.status, m.statusErr = "flourishes are disabled in news-safe mode", false
					return m, nil
				}
				m.flourish = time.Now().UnixNano()
				return m, m.startTransform()
			case "w":
				if m.newsSafe {
					m.applyPreset(2)
				} else {
					m.applyNewsSafe()
				}
				return m, m.startTransform()
			case "v":
				if !m.newsSafe {
					m.status, m.statusErr = "news flair requires news-safe mode", false
					return m, nil
				}
				m.newsFlair = (m.newsFlair + 1) % 4
				return m, m.startTransform()
			case "g":
				if m.newsSafe {
					m.status, m.statusErr = "grammar is locked off in news-safe mode", false
					return m, nil
				}
				m.grammar = !m.grammar
				return m, m.startTransform()
			case "j":
				if m.newsSafe {
					m.status, m.statusErr = "interjections are locked off in news-safe mode", false
					return m, nil
				}
				m.interject = !m.interject
				return m, m.startTransform()
			case "f":
				if m.newsSafe {
					m.status, m.statusErr = "comedy is locked off in news-safe mode", false
					return m, nil
				}
				m.comedy = !m.comedy
				return m, m.startTransform()
			case "tab":
				m.activePane = 1 - m.activePane
				return m, nil
			case "s":
				m.save()
				return m, nil
			case "esc":
				m.requestID++ // discard any result still in flight
				m.state, m.status, m.confirm = statePick, "", false
				return m, m.picker.Init()
			}
		}
	}

	var cmd tea.Cmd
	if m.state == statePick {
		m.picker, cmd = m.picker.Update(msg)
		if ok, path := m.picker.DidSelectFile(msg); ok {
			openCmd, err := m.openFile(path)
			if err != nil {
				m.status, m.statusErr = err.Error(), true
				return m, nil
			}
			return m, openCmd
		}
		return m, cmd
	}

	if m.activePane == 0 {
		m.originalVP, cmd = m.originalVP.Update(msg)
		m.renderedVP.YOffset = m.originalVP.YOffset
	} else {
		m.renderedVP, cmd = m.renderedVP.Update(msg)
		m.originalVP.YOffset = m.renderedVP.YOffset
	}
	return m, cmd
}

func (m model) View() string {
	if !m.ready {
		return "summoning…"
	}
	title := titleSty.Render("🧙 wizardify") + dimSty.Render("  ~ review before inscription ~")
	if m.state == statePick {
		help := dimSty.Render("↑/↓ move · enter select · q quit")
		status := ""
		if m.status != "" {
			status = "\n" + errorSty.Render(m.status)
		}
		return fmt.Sprintf("%s\n\n%s\n%s%s\n\n%s", title,
			headerSty.Render("Choose a parchment:"), m.picker.View(), status, help)
	}

	profile := intensityNames[m.intensity]
	if m.newsSafe {
		profile = fmt.Sprintf("NEWS-SAFE · flair %d", m.newsFlair)
	}
	info := fmt.Sprintf("%s · %s · grammar %s · interjections %s · comedy %s",
		filepath.Base(m.file), profile, onOff(m.grammar), onOff(m.interject), onOff(m.comedy))
	status := m.status
	if m.working {
		status = "transmuting…"
	}
	statusView := statusSty.Render(status)
	if m.statusErr {
		statusView = errorSty.Render(status)
	}

	body := ""
	if m.width >= 100 {
		paneWidth := maxInt(20, (m.width-11)/2)
		left := borderSty.Copy().Width(paneWidth).Render(m.originalVP.View())
		right := borderSty.Copy().Width(paneWidth).Render(m.renderedVP.View())
		labels := fmt.Sprintf("%-*s  %s", paneWidth+2, activePanel.Render("ORIGINAL"), activePanel.Render("WIZARDIFIED"))
		body = labels + "\n" + lipgloss.JoinHorizontal(lipgloss.Top, left, " ", right)
	} else {
		label, content := "ORIGINAL", m.originalVP.View()
		if m.activePane == 1 {
			label, content = "WIZARDIFIED", m.renderedVP.View()
		}
		body = activePanel.Render(label) + dimSty.Render(" · tab switches pane") + "\n" +
			borderSty.Copy().Width(maxInt(20, m.width-4)).Render(content)
	}

	help := keySty.Render("1/2/3") + dimSty.Render(" preset · ") +
		keySty.Render("w") + dimSty.Render(" news-safe · ") +
		keySty.Render("v") + dimSty.Render(" news flair · ") +
		keySty.Render("g/j/f") + dimSty.Render(" grammar/interjections/comedy · ") +
		keySty.Render("n") + dimSty.Render(" reroll flourishes · ") +
		keySty.Render("s") + dimSty.Render(" save · ") +
		keySty.Render("esc") + dimSty.Render(" back · ") + keySty.Render("q") + dimSty.Render(" quit")
	return fmt.Sprintf("%s  %s\n%s\n%s\n%s", title, headerSty.Render(info), statusView, body, help)
}

func runTUI(wizard *Wizardifier, dir string, newsSafe bool, flair ...int) {
	p := tea.NewProgram(newModel(wizard, dir, newsSafe, flair...), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func onOff(value bool) string {
	if value {
		return "on"
	}
	return "off"
}

func friendlyDuration(d time.Duration) string {
	if d < time.Millisecond {
		return "<1ms"
	}
	if d < time.Second {
		return fmt.Sprintf("%dms", d.Milliseconds())
	}
	return fmt.Sprintf("%.1fs", d.Seconds())
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func highlightedDiff(original, rendered string) (string, string, int) {
	leftLines, rightLines := strings.Split(original, "\n"), strings.Split(rendered, "\n")
	n := maxInt(len(leftLines), len(rightLines))
	leftOut, rightOut := make([]string, n), make([]string, n)
	changes := 0
	for i := 0; i < n; i++ {
		var left, right string
		if i < len(leftLines) {
			left = leftLines[i]
		}
		if i < len(rightLines) {
			right = rightLines[i]
		}
		leftOut[i], rightOut[i], changes = diffLine(left, right, changes)
	}
	return strings.Join(leftOut, "\n"), strings.Join(rightOut, "\n"), changes
}

// wrapAlignedDiff performs ANSI-aware word wrapping and pads each logical
// line pair to the same rendered height. This keeps the two panes aligned even
// when a replacement makes one side wrap onto more rows than the other.
func wrapAlignedDiff(left, right string, width int) (string, string) {
	if width < 1 {
		return left, right
	}
	leftLines, rightLines := strings.Split(left, "\n"), strings.Split(right, "\n")
	n := maxInt(len(leftLines), len(rightLines))
	leftOut, rightOut := make([]string, 0, n), make([]string, 0, n)
	for i := 0; i < n; i++ {
		var leftLine, rightLine string
		if i < len(leftLines) {
			leftLine = leftLines[i]
		}
		if i < len(rightLines) {
			rightLine = rightLines[i]
		}
		wrappedLeft := strings.Split(ansi.Wrap(leftLine, width, "/"), "\n")
		wrappedRight := strings.Split(ansi.Wrap(rightLine, width, "/"), "\n")
		rows := maxInt(len(wrappedLeft), len(wrappedRight))
		for row := 0; row < rows; row++ {
			if row < len(wrappedLeft) {
				leftOut = append(leftOut, wrappedLeft[row])
			} else {
				leftOut = append(leftOut, "")
			}
			if row < len(wrappedRight) {
				rightOut = append(rightOut, wrappedRight[row])
			} else {
				rightOut = append(rightOut, "")
			}
		}
	}
	return strings.Join(leftOut, "\n"), strings.Join(rightOut, "\n")
}

func diffLine(left, right string, changes int) (string, string, int) {
	leftLocs, rightLocs := reDiffToken.FindAllStringIndex(left, -1), reDiffToken.FindAllStringIndex(right, -1)
	if left == right {
		return left, right, changes
	}
	if len(leftLocs)*len(rightLocs) > 40000 {
		return removedSty.Render(left), addedSty.Render(right), changes + maxInt(len(leftLocs), len(rightLocs))
	}
	leftTokens, rightTokens := tokenStrings(left, leftLocs), tokenStrings(right, rightLocs)
	dp := make([][]int, len(leftTokens)+1)
	for i := range dp {
		dp[i] = make([]int, len(rightTokens)+1)
	}
	for i := len(leftTokens) - 1; i >= 0; i-- {
		for j := len(rightTokens) - 1; j >= 0; j-- {
			if leftTokens[i] == rightTokens[j] {
				dp[i][j] = dp[i+1][j+1] + 1
			} else {
				dp[i][j] = maxInt(dp[i+1][j], dp[i][j+1])
			}
		}
	}
	leftSame, rightSame := make([]bool, len(leftTokens)), make([]bool, len(rightTokens))
	for i, j := 0, 0; i < len(leftTokens) && j < len(rightTokens); {
		if leftTokens[i] == rightTokens[j] {
			leftSame[i], rightSame[j] = true, true
			i, j = i+1, j+1
		} else if dp[i+1][j] >= dp[i][j+1] {
			i++
		} else {
			j++
		}
	}
	changed := 0
	for _, same := range rightSame {
		if !same {
			changed++
		}
	}
	return styleTokenRanges(left, leftLocs, leftSame, removedSty),
		styleTokenRanges(right, rightLocs, rightSame, addedSty), changes + changed
}

func tokenStrings(text string, locs [][]int) []string {
	out := make([]string, len(locs))
	for i, loc := range locs {
		out[i] = text[loc[0]:loc[1]]
	}
	return out
}

func styleTokenRanges(text string, locs [][]int, same []bool, style lipgloss.Style) string {
	var b strings.Builder
	cur := 0
	for i, loc := range locs {
		b.WriteString(text[cur:loc[0]])
		token := text[loc[0]:loc[1]]
		if same[i] {
			b.WriteString(token)
		} else {
			b.WriteString(style.Render(token))
		}
		cur = loc[1]
	}
	b.WriteString(text[cur:])
	return b.String()
}
