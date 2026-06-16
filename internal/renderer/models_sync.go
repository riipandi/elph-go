package renderer

import (
	"fmt"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/huh/v2"
	"charm.land/lipgloss/v2"
	"github.com/riipandi/elph/internal/settings"
	"github.com/riipandi/elph/internal/uiconst"
	"github.com/riipandi/elph/pkg/ai/provider"
)

const (
	modelsSyncUpdatingLabel = "Updating model metadata from models.dev"
	modelsSyncDialogLabel   = "Model update"

	modelsSyncChoiceUpdate = "update"
	modelsSyncChoiceSkip   = "skip"
	modelsSyncChoiceCancel = "cancel"
)

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

const modelsSyncUpdateAvailableText = "Model metadata updates are available."

func formatModelsSyncDescription(width int) string {
	desc := clampMultilineText(modelsSyncUpdateAvailableText, width, maxApprovalDescriptionLines)
	if desc == "" {
		return ""
	}
	return desc + "\n"
}

func newModelsSyncForm(providers []string, width int) *huh.Form {
	choice := modelsSyncChoiceUpdate
	return huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Key("choice").
				Title("Apply model metadata updates?").
				Description(formatModelsSyncDescription(width)).
				Options(
					huh.NewOption("Update", modelsSyncChoiceUpdate),
					huh.NewOption("Skip", modelsSyncChoiceSkip),
					huh.NewOption("Cancel", modelsSyncChoiceCancel),
				).
				Value(&choice),
		),
	).
		WithWidth(width).
		WithShowHelp(false).
		WithTheme(toolInteractFormTheme())
}

func normalizeModelsSyncChoice(raw string) string {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case modelsSyncChoiceUpdate:
		return modelsSyncChoiceUpdate
	case modelsSyncChoiceSkip:
		return modelsSyncChoiceSkip
	case modelsSyncChoiceCancel:
		return modelsSyncChoiceCancel
	default:
		if raw == "" {
			return modelsSyncChoiceUpdate
		}
		return raw
	}
}

func (m Model) modelsSyncingActive() bool {
	return m.modelsSyncing
}

func (m Model) modelsSyncDialogActive() bool {
	return m.modelsSyncForm != nil
}

func (m Model) modelsSyncFormWidth() int {
	return m.toolInteractFormWidth()
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

	case tea.KeyPressMsg:
		switch strings.ToLower(msg.String()) {
		case "y", "1":
			return m.resolveModelsSyncConfirm(true)
		case "n", "2", "3", "c":
			return m.resolveModelsSyncConfirm(false)
		}
	}

	form, cmd := m.modelsSyncForm.Update(msg)
	if f, ok := form.(*huh.Form); ok {
		m.modelsSyncForm = f
	}
	if cmd != nil {
		cmds = append(cmds, cmd)
	}
	m = m.syncLayout(m.content.AtBottom())

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
		update := normalizeModelsSyncChoice(form.GetString("choice")) == modelsSyncChoiceUpdate
		return m.resolveModelsSyncConfirm(update)
	case huh.StateAborted:
		return m.resolveModelsSyncConfirm(false)
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

func (m Model) modelsSyncDialogBody() string {
	formView := trimTrailingLineSpaces(strings.TrimSuffix(m.modelsSyncForm.View(), "\n\n"))
	labelLine := lipgloss.NewStyle().Foreground(uiconst.Yellow).Bold(true).Render(modelsSyncDialogLabel)
	hintLine := lipgloss.NewStyle().Foreground(uiconst.DimText).Render("y update · n skip · c cancel · 1-3 · ↑/↓ · Enter · Esc")
	return lipgloss.JoinVertical(lipgloss.Left, labelLine, "", formView, "", hintLine)
}

func (m Model) modelsSyncChromeView() string {
	boxW := borderedChromeWidth(m.chromeOuterWidth())
	inner := m.modelsSyncDialogBody()
	return lipgloss.NewStyle().MarginTop(1).Render(
		cachedInputBorder(m.mode).Width(boxW).Render(inner),
	)
}

func (m Model) modelsSyncDialogHeight() int {
	if !m.modelsSyncDialogActive() {
		return 0
	}
	return lipgloss.Height(m.modelsSyncChromeView())
}

func (m Model) startModelsSync() (Model, tea.Cmd) {
	m.modelsSyncing = true
	m.agent.SpinnerFrame = 0
	m, _ = m.setModelsSyncMessage(m.modelsSyncStatusText())
	return m, tea.Batch(runModelsSyncCmd(), m.spinnerTickCmd())
}

func (m Model) setModelsSyncMessage(text string) (Model, tea.Cmd) {
	msg := message{text: text, kind: uiconst.MessageSystem}
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
