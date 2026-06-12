package tui

import (
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/riipandi/elph/internal/config"
	"github.com/riipandi/elph/internal/constants"
)

// ─── Braille Logo ────────────────────────────────────────────────────────────

const (
	logoLine1 = "\u28FF\u28FF\u285F\u28FF\u285F\u28FF\u28FF"
	logoLine2 = "\u28FF\u28FF\u28FF\u28FF\u28FF\u28FF\u28FF"
)

// ─── Tips ────────────────────────────────────────────────────────────────────

var tips = []string{
	"Use --no-session for ephemeral mode — no session file is saved, useful for one-off queries.",
	"Send /changelog to show version history.",
	"Use /help to see all available commands.",
	"Press Ctrl+C once to cancel, twice to exit.",
	"Type :q and press Enter to quit (vim-style exit).",
	"Press Ctrl+D to exit the application.",
	"Use Tab and Shift+Tab to switch between agent modes.",
}

func randomTip() string {
	return tips[rand.Intn(len(tips))]
}

// ─── Styles ──────────────────────────────────────────────────────────────────

var (
	blueCol   = lipgloss.Color("#3B82F6")
	highlight = lipgloss.AdaptiveColor{Light: "#874BFD", Dark: "#7C56DC"}
	special   = lipgloss.AdaptiveColor{Light: "#43BF6D", Dark: "#73F59F"}
	dimText   = lipgloss.AdaptiveColor{Light: "#9B9B9B", Dark: "#5C5C5C"}
)

func modeBorderColor(m constants.AgentMode) lipgloss.Color {
	switch m {
	case constants.ModeBrave:
		return lipgloss.Color("#EF4444")
	case constants.ModePlan:
		return lipgloss.Color("#06B6D4")
	case constants.ModeAsk:
		return lipgloss.Color("#22C55E")
	default:
		return lipgloss.Color("#A855F7")
	}
}

func bannerStyle(w int) lipgloss.Style {
	return lipgloss.NewStyle().
		Width(w-2).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(blueCol).
		Padding(1, 2)
}

func inputStyle(w int, m constants.AgentMode) lipgloss.Style {
	return lipgloss.NewStyle().
		Width(w-2).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(modeBorderColor(m)).
		Padding(0, 1)
}

func footerStyle(w int) lipgloss.Style {
	return lipgloss.NewStyle().
		Width(w).
		Padding(0, 1)
}

// ─── Mode Ordering ───────────────────────────────────────────────────────────

var modes = []constants.AgentMode{
	constants.ModeBuild,
	constants.ModePlan,
	constants.ModeAsk,
	constants.ModeBrave,
}

func nextMode(m constants.AgentMode) constants.AgentMode {
	for i, mode := range modes {
		if mode == m {
			return modes[(i+1)%len(modes)]
		}
	}
	return modes[0]
}

func prevMode(m constants.AgentMode) constants.AgentMode {
	for i, mode := range modes {
		if mode == m {
			prev := i - 1
			if prev < 0 {
				prev = len(modes) - 1
			}
			return modes[prev]
		}
	}
	return modes[len(modes)-1]
}

// ─── Exit Messages ──────────────────────────────────────────────────────────

const doubleTapTimeout = 3 * time.Second

type ctrlCResetMsg struct{}

// ─── Model ───────────────────────────────────────────────────────────────────

type Model struct {
	ready         bool
	width         int
	height        int
	input         textinput.Model
	vp            viewport.Model
	messages      []string
	modelName     string
	mode          constants.AgentMode
	sessionID     string
	workDir       string
	branch        string
	tip           string
	quitting      bool
	ctrlCPress    int // 0=none, 1=first, 2=second (input cleared)
	ctrlCNoticeID int // index in messages of the notice (-1 = none)
}

func New() Model {
	wd, _ := os.Getwd()

	ti := textinput.New()
	ti.Placeholder = "Type a message or /command..."
	ti.Focus()
	ti.CharLimit = 4096
	ti.Width = 60
	ti.PromptStyle = lipgloss.NewStyle().Foreground(highlight)
	ti.TextStyle = lipgloss.NewStyle()

	return Model{
		input:         ti,
		modelName:     config.AppName,
		mode:          constants.ModeBuild,
		sessionID:     "sess_abc123",
		workDir:       wd,
		branch:        "main",
		messages:      []string{},
		tip:           randomTip(),
		ctrlCNoticeID: -1,
	}
}

// ─── tea.Model Implementation ────────────────────────────────────────────────

func (m Model) Init() tea.Cmd {
	return textinput.Blink
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.ready = true

		reserved := 9 + 3 + 3 + 2
		vpHeight := msg.Height - reserved
		if vpHeight < 3 {
			vpHeight = 3
		}

		m.vp = viewport.New(msg.Width, vpHeight)
		m.vp.YPosition = 0
		m.vp.Style = lipgloss.NewStyle().Padding(0, 1)

	case ctrlCResetMsg:
		m = m.cancelCtrlC()

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			hasInput := m.input.Value() != ""

			if m.ctrlCPress == 1 && hasInput {
				// Second press, input non-empty → clear input
				m.ctrlCPress = 2
				m.input.SetValue("")
				m = m.replaceNotice("Input cleared, press ctrl+c again to exit")
				return m, tea.Tick(doubleTapTimeout, func(t time.Time) tea.Msg {
					return ctrlCResetMsg{}
				})
			}

			if m.ctrlCPress == 2 || (m.ctrlCPress == 1 && !hasInput) {
				// Third press, or second when input was empty → quit
				m.quitting = true
				return m, tea.Quit
			}

			// First Ctrl+C
			m.ctrlCPress = 1
			m = m.withMessage("Press ctrl+c again to exit")
			m.ctrlCNoticeID = len(m.messages) - 1
			return m, tea.Tick(doubleTapTimeout, func(t time.Time) tea.Msg {
				return ctrlCResetMsg{}
			})

		case "ctrl+d":
			m.quitting = true
			return m, tea.Quit

		case "tab":
			m.mode = nextMode(m.mode)
			m = m.withMessage(fmt.Sprintf("Switched to %s mode", m.mode))

		case "shift+tab":
			m.mode = prevMode(m.mode)
			m = m.withMessage(fmt.Sprintf("Switched to %s mode", m.mode))

		case "enter":
			val := strings.TrimSpace(m.input.Value())
			if val == "" {
				break
			}
			if val == ":q" || val == ":q!" {
				m.quitting = true
				return m, tea.Quit
			}
			m.messages = append(m.messages, val)
			m.input.SetValue("")
		}

		// Any other key cancels the pending Ctrl+C state.
		m = m.cancelCtrlC()
	}

	// Update input component
	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	cmds = append(cmds, cmd)

	// Update viewport component
	m.vp, cmd = m.vp.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m Model) View() string {
	if m.quitting {
		return ""
	}
	if !m.ready {
		return "\n  Initializing..."
	}

	bannerView := m.bannerView()
	inputView := m.inputView()
	footerView := m.footerView()

	bannerH := lipgloss.Height(bannerView)
	inputH := lipgloss.Height(inputView)
	footerH := lipgloss.Height(footerView)
	gaps := 2

	vpHeight := m.height - bannerH - inputH - footerH - gaps
	if vpHeight < 1 {
		vpHeight = 1
	}

	m.vp.Width = m.width
	m.vp.Height = vpHeight
	m.vp.SetContent(m.streamView())

	parts := []string{
		bannerView,
		"",
		m.vp.View(),
		"",
		inputView,
		footerView,
	}

	return lipgloss.JoinVertical(lipgloss.Top, parts...)
}

// ─── Sub-views ───────────────────────────────────────────────────────────────

func (m Model) bannerView() string {
	w := m.width

	versionLine := fmt.Sprintf("Welcome to Elph v%s", config.AppVersion)
	if config.BuildHash != "unknown" {
		versionLine = fmt.Sprintf("Welcome to Elph v%s (%s)", config.AppVersion, config.BuildHash[:7])
	}

	header := lipgloss.NewStyle().Bold(true).Render(versionLine)

	dirLine := fmt.Sprintf("Directory:  %s", m.workDir)
	modelLine := fmt.Sprintf("Model:      %s", m.modelName)
	statsLine := fmt.Sprintf("Stats:      00 ext, 00 commands, 00 skills, 00 tools | Mode: %s", m.mode)

	logo := lipgloss.JoinVertical(lipgloss.Left,
		lipgloss.NewStyle().Foreground(special).Render(logoLine1),
		lipgloss.NewStyle().Foreground(special).Render(logoLine2),
	)

	content := lipgloss.JoinHorizontal(lipgloss.Top,
		lipgloss.NewStyle().MarginRight(2).Render(logo),
		lipgloss.JoinVertical(lipgloss.Left,
			header,
			"",
			lipgloss.NewStyle().Foreground(dimText).Render(dirLine),
			lipgloss.NewStyle().Foreground(dimText).Render(modelLine),
			lipgloss.NewStyle().Foreground(dimText).Render(statsLine),
			"",
			lipgloss.NewStyle().Foreground(dimText).Italic(true).Render("Tip: "+m.tip),
		),
	)

	return bannerStyle(w).Render(content)
}

func (m Model) streamView() string {
	if len(m.messages) == 0 {
		return lipgloss.NewStyle().Foreground(dimText).Render("MCP: 0 servers connected (000 tools)")
	}

	var b strings.Builder
	for _, msg := range m.messages {
		b.WriteString(msg)
		b.WriteString("\n")
	}
	return strings.TrimRight(b.String(), "\n")
}

func (m Model) inputView() string {
	w := m.width
	m.input.Width = w - 6
	borderColor := modeBorderColor(m.mode)
	m.input.PromptStyle = lipgloss.NewStyle().Foreground(borderColor)
	return inputStyle(w, m.mode).Render(m.input.View())
}

func (m Model) footerView() string {
	w := m.width
	wd := filepath.Base(m.workDir)

	line1 := fmt.Sprintf("%s | opencode | T: high | IMG", m.modelName)
	line1right := "$0.00 | 0.0% (262k)"
	pad1 := w - lipgloss.Width(line1) - lipgloss.Width(line1right)
	if pad1 < 1 {
		pad1 = 1
	}

	line2 := fmt.Sprintf("%s [%s]", wd, m.sessionID)
	line2right := fmt.Sprintf("turn: 0 | %s [+00 -00]", m.branch)
	pad2 := w - lipgloss.Width(line2) - lipgloss.Width(line2right)
	if pad2 < 1 {
		pad2 = 1
	}

	s := lipgloss.NewStyle().Foreground(dimText)

	return footerStyle(w).Render(
		lipgloss.JoinVertical(lipgloss.Left,
			s.Render(line1+strings.Repeat(" ", pad1)+line1right),
			s.Render(line2+strings.Repeat(" ", pad2)+line2right),
		),
	)
}

// ─── Helpers ────────────────────────────────────────────────────────────────

func (m Model) withMessage(msg string) Model {
	styled := lipgloss.NewStyle().Foreground(highlight).Render("> ") + msg
	m.messages = append(m.messages, styled)
	return m
}

// replaceNotice replaces the existing Ctrl+C notice with a new message.
func (m Model) replaceNotice(msg string) Model {
	styled := lipgloss.NewStyle().Foreground(highlight).Render("> ") + msg
	if m.ctrlCNoticeID >= 0 && m.ctrlCNoticeID < len(m.messages) {
		m.messages[m.ctrlCNoticeID] = styled
	} else {
		m.messages = append(m.messages, styled)
		m.ctrlCNoticeID = len(m.messages) - 1
	}
	return m
}

// cancelCtrlC removes the Ctrl+C notice and resets the press state.
func (m Model) cancelCtrlC() Model {
	m.ctrlCPress = 0
	if m.ctrlCNoticeID >= 0 && m.ctrlCNoticeID < len(m.messages) {
		m.messages = append(m.messages[:m.ctrlCNoticeID], m.messages[m.ctrlCNoticeID+1:]...)
	}
	m.ctrlCNoticeID = -1
	return m
}

// ─── Public API ──────────────────────────────────────────────────────────────

func Run() error {
	m := New()
	p := tea.NewProgram(m, tea.WithAltScreen(), tea.WithMouseCellMotion())
	_, err := p.Run()
	return err
}
