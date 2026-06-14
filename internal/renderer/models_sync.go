package renderer

import (
	"fmt"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/huh/v2"
	"charm.land/lipgloss/v2"
	"github.com/riipandi/elph/internal/constants"
	"github.com/riipandi/elph/internal/settings"
	"github.com/riipandi/elph/pkg/ai/provider"
)

const modelsSyncUpdatingLabel = "Updating model metadata from models.dev"

type modelsSyncOfferMsg struct {
	providers []string
}

type modelsSyncCheckDoneMsg struct {
	err error
}

type modelsSyncDoneMsg struct {
	err    error
	result provider.UpdateModelsResult
}

func checkModelsSyncAtStartupCmd() tea.Cmd {
	return func() tea.Msg {
		cfg, err := settings.Load()
		if err != nil {
			return modelsSyncCheckDoneMsg{err: err}
		}
		if !cfg.SyncDue(time.Now()) {
			return nil
		}

		result, err := provider.PreviewModelsDevUpdates(provider.UpdateModelsOptions{})
		if err != nil {
			return modelsSyncCheckDoneMsg{err: err}
		}
		if len(result.Updated) == 0 {
			_ = settings.MarkModelsSynced(time.Now())
			return nil
		}
		return modelsSyncOfferMsg{providers: result.Updated}
	}
}

func runModelsSyncCmd() tea.Cmd {
	return func() tea.Msg {
		result, err := settings.RunModelsSync()
		return modelsSyncDoneMsg{err: err, result: result}
	}
}

func modelsSyncFormDescription(providers []string) string {
	return fmt.Sprintf("models.dev has updates for:\n%s", strings.Join(providers, ", "))
}

func newModelsSyncForm(providers []string, width int) *huh.Form {
	var update bool
	return huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Key("update").
				Title("Model metadata update available").
				Description(modelsSyncFormDescription(providers)).
				Affirmative("Update").
				Negative("Skip").
				Value(&update),
		),
	).
		WithWidth(width).
		WithShowHelp(true).
		WithTheme(huh.ThemeFunc(huh.ThemeCharm))
}

func (m Model) modelsSyncingActive() bool {
	return m.modelsSyncing
}

func (m Model) modelsSyncDialogActive() bool {
	return m.modelsSyncForm != nil
}

func (m Model) modelsSyncFormWidth() int {
	w := m.width - 6
	if w > 72 {
		return 72
	}
	if w < 32 {
		return 32
	}
	return w
}

func (m Model) syncModelsSyncFormWidth() Model {
	if m.modelsSyncForm == nil {
		return m
	}
	m.modelsSyncForm = m.modelsSyncForm.WithWidth(m.modelsSyncFormWidth())
	return m
}

func (m Model) modelsSyncStatusText() string {
	frame := spinnerFrames[m.agent.SpinnerFrame%len(spinnerFrames)]
	return frame + " " + modelsSyncUpdatingLabel + "..."
}

func (m Model) offerModelsSync(providers []string) (Model, tea.Cmd) {
	m.input.Blur()
	m.modelsSyncForm = newModelsSyncForm(providers, m.modelsSyncFormWidth())
	return m, m.modelsSyncForm.Init()
}

func (m Model) updateModelsSyncForm(msg tea.Msg) (Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.ready = true
		m = m.syncModelsSyncFormWidth()
		m.layout.ContentDirty = true
		m = m.syncLayout(false)
	}

	form, cmd := m.modelsSyncForm.Update(msg)
	if f, ok := form.(*huh.Form); ok {
		m.modelsSyncForm = f
	}
	if cmd != nil {
		cmds = append(cmds, cmd)
	}

	switch m.modelsSyncForm.State {
	case huh.StateCompleted, huh.StateAborted:
		var completeCmd tea.Cmd
		m, completeCmd = m.completeModelsSyncForm()
		if completeCmd != nil {
			cmds = append(cmds, completeCmd)
		}
	}

	return m, tea.Batch(cmds...)
}

func (m Model) completeModelsSyncForm() (Model, tea.Cmd) {
	form := m.modelsSyncForm
	m.modelsSyncForm = nil
	m.input.Focus()

	switch form.State {
	case huh.StateCompleted:
		if form.GetBool("update") {
			return m.startModelsSync()
		}
		return m.declineModelsSync(), nil
	case huh.StateAborted:
		return m.declineModelsSync(), nil
	default:
		return m, nil
	}
}

func (m Model) resolveModelsSyncConfirm(update bool) (Model, tea.Cmd) {
	m.modelsSyncForm = nil
	m.input.Focus()
	if update {
		return m.startModelsSync()
	}
	return m.declineModelsSync(), nil
}

func (m Model) declineModelsSync() Model {
	text := "Model metadata update skipped."
	m, _ = m.setModelsSyncMessage(text)
	m.session.AppendLog("system", text)
	m.modelsSyncMsgID = -1
	_ = settings.MarkModelsSynced(time.Now())
	return m
}

func (m Model) modelsSyncDialogView() string {
	formView := strings.TrimSuffix(m.modelsSyncForm.View(), "\n\n")
	boxW := borderedChromeWidth(m.chromeOuterWidth())
	border := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(constants.Blue).
		Padding(1, 2)
	return lipgloss.NewStyle().MarginTop(1).Render(
		border.Width(boxW).Render(formView),
	)
}

func (m Model) modelsSyncDialogHeight() int {
	if !m.modelsSyncDialogActive() {
		return 0
	}
	return lipgloss.Height(m.modelsSyncDialogView())
}

func (m Model) startModelsSync() (Model, tea.Cmd) {
	m.modelsSyncing = true
	m.agent.SpinnerFrame = 0
	m, _ = m.setModelsSyncMessage(m.modelsSyncStatusText())
	return m, tea.Batch(runModelsSyncCmd(), m.spinnerTickCmd())
}

func (m Model) setModelsSyncMessage(text string) (Model, tea.Cmd) {
	msg := message{text: text, kind: constants.MessageSystem}
	if m.modelsSyncMsgID >= 0 && m.modelsSyncMsgID < len(m.messages) {
		m.messages[m.modelsSyncMsgID] = msg
	} else {
		m.messages = append(m.messages, msg)
		m.modelsSyncMsgID = len(m.messages) - 1
	}
	m.layout.ContentDirty = true
	m = m.syncLayout(true)
	return m, nil
}

func (m Model) refreshModelsSyncStatus() Model {
	if !m.modelsSyncing {
		return m
	}
	m, _ = m.setModelsSyncMessage(m.modelsSyncStatusText())
	return m
}

func (m Model) finishModelsSync(msg modelsSyncDoneMsg) Model {
	m.modelsSyncing = false
	text := formatModelsSyncResult(msg)
	m, _ = m.setModelsSyncMessage(text)
	m.session.AppendLog("system", text)
	m.modelsSyncMsgID = -1

	if msg.err != nil {
		return m
	}

	catalog, err := provider.LoadCatalog("")
	if err != nil {
		return m
	}
	m.session.EnabledModelCount = catalog.TotalEnabledModels()
	m.session.Catalog = provider.TrimCatalogForRuntime(catalog, m.session.ProviderID, m.session.ModelID)
	return m
}

func formatModelsSyncResult(msg modelsSyncDoneMsg) string {
	if msg.err != nil {
		return fmt.Sprintf("Model metadata update failed: %v", msg.err)
	}
	if len(msg.result.Updated) > 0 {
		return "Model metadata updated: " + strings.Join(msg.result.Updated, ", ")
	}
	return "Model metadata is up to date."
}
