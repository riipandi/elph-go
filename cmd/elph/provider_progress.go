package main

import (
	"fmt"
	"os"
	"reflect"
	"strings"
	"time"

	"charm.land/bubbles/v2/spinner"
	"charm.land/bubbles/v2/stopwatch"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/term"
	"github.com/riipandi/elph/internal/uiconst"
	"github.com/riipandi/elph/internal/settings"
	"github.com/riipandi/elph/pkg/ai/provider"
)

type providerProgressMode int

const (
	providerProgressConnect providerProgressMode = iota
	providerProgressUpdate
)

type providerProgressModel struct {
	mode       providerProgressMode
	force      bool
	spinner    spinner.Model
	stopwatch  stopwatch.Model
	progressCh chan provider.ProviderProgressEvent
	doneCh     chan providerStepDoneMsg

	phaseLabel      string
	currentProvider string
	currentLabel    string
	currentAction   provider.ProviderProgressAction
	providerIndex   int
	providerTotal   int
	completed       []string
	viewLines       int
	workStarted     bool

	err       error
	bootstrap provider.BootstrapResult
	sync      provider.UpdateModelsResult
}

type providerStepDoneMsg struct {
	phase     provider.ProviderProgressPhase
	bootstrap *provider.BootstrapResult
	sync      *provider.UpdateModelsResult
	err       error
}

func isInteractiveStderr() bool {
	return term.IsTerminal(os.Stderr.Fd())
}

func runConnectWithProgress(force bool) (provider.BootstrapResult, error) {
	if !isInteractiveStderr() {
		return provider.BootstrapProviders("", force)
	}

	m, err := runInteractiveProviderProgress(providerProgressConnect, force)
	if err != nil {
		return provider.BootstrapResult{}, err
	}
	return m.bootstrap, nil
}

func runUpdateWithProgress(force bool) (provider.BootstrapResult, provider.UpdateModelsResult, error) {
	if !isInteractiveStderr() {
		bootstrap, err := provider.BootstrapProviders("", force)
		if err != nil {
			return bootstrap, provider.UpdateModelsResult{}, err
		}
		sync, err := settings.RunModelsSync()
		return bootstrap, sync, err
	}

	m, err := runInteractiveProviderProgress(providerProgressUpdate, force)
	if err != nil {
		return m.bootstrap, m.sync, err
	}
	return m.bootstrap, m.sync, nil
}

func runInteractiveProviderProgress(mode providerProgressMode, force bool) (*providerProgressModel, error) {
	m := newProviderProgressModel(mode, force)
	m.stopwatch = applyStopwatchCmd(m.stopwatch, m.stopwatch.Start())

	spinInterval := m.spinner.Spinner.FPS
	if spinInterval <= 0 {
		spinInterval = time.Second / 12
	}
	spinTicker := time.NewTicker(spinInterval)
	clockTicker := time.NewTicker(100 * time.Millisecond)
	defer spinTicker.Stop()
	defer clockTicker.Stop()

	drawn := false
	redraw := func() {
		if drawn && m.viewLines > 0 {
			fmt.Fprintf(os.Stderr, "\033[%dF\r\033[J", m.viewLines)
		}
		fmt.Fprint(os.Stderr, m.progressContent())
		drawn = true
	}

	redraw()
	m.beginWork()

	for {
		select {
		case <-spinTicker.C:
			m.spinner, _ = m.spinner.Update(m.spinner.Tick())
			redraw()
		case <-clockTicker.C:
			m.stopwatch, _ = m.stopwatch.Update(stopwatch.TickMsg{ID: m.stopwatch.ID()})
			redraw()
		case evt := <-m.progressCh:
			m.applyProgressEvent(evt)
			redraw()
		case done := <-m.doneCh:
			if done.err != nil {
				m.err = done.err
				if done.bootstrap != nil {
					m.bootstrap = *done.bootstrap
				}
				if done.sync != nil {
					m.sync = *done.sync
				}
				break
			}
			if done.bootstrap != nil {
				m.bootstrap = *done.bootstrap
			}
			if done.sync != nil {
				m.sync = *done.sync
			}
			if done.phase == provider.ProviderProgressConnect && m.mode == providerProgressUpdate {
				m.resetForSyncPhase()
				redraw()
				go m.runSyncPhase()
				continue
			}
			goto finish
		}
		if m.err != nil {
			break
		}
	}

finish:
	m.stopwatch = applyStopwatchCmd(m.stopwatch, m.stopwatch.Stop())
	if drawn {
		clearProviderProgress(m.viewLines)
	}
	return m, m.err
}

func newProviderProgressModel(mode providerProgressMode, force bool) *providerProgressModel {
	spin := spinner.New(
		spinner.WithSpinner(spinner.MiniDot),
		spinner.WithStyle(lipgloss.NewStyle().Foreground(uiconst.Yellow)),
	)
	clock := stopwatch.New(stopwatch.WithInterval(100 * time.Millisecond))

	return &providerProgressModel{
		mode:         mode,
		force:        force,
		spinner:      spin,
		stopwatch:    clock,
		progressCh:   make(chan provider.ProviderProgressEvent, 64),
		doneCh:       make(chan providerStepDoneMsg, 2),
		phaseLabel:   "Connecting providers",
		currentLabel: "starting",
		viewLines:    3,
	}
}

func (m *providerProgressModel) beginWork() {
	if m.workStarted {
		return
	}
	m.workStarted = true
	go m.runConnectPhase()
}

func (m *providerProgressModel) applyProgressEvent(evt provider.ProviderProgressEvent) {
	switch evt.Phase {
	case provider.ProviderProgressConnect:
		m.phaseLabel = "Connecting providers"
	case provider.ProviderProgressSync:
		m.phaseLabel = "Syncing model catalogs"
	}

	m.currentAction = evt.Action
	switch evt.Action {
	case provider.ProviderProgressFetchMeta:
		m.currentProvider = ""
		m.currentLabel = "models.dev"
		m.providerIndex = 0
		m.providerTotal = 0
	case provider.ProviderProgressWorking:
		m.currentProvider = evt.ProviderID
		m.currentLabel = currentProviderName(evt.ProviderID, evt.Label)
		m.providerIndex = evt.Index
		m.providerTotal = evt.Total
	default:
		m.completed = appendCompletedLine(m.completed, formatCompletedProvider(evt))
		m.providerIndex = evt.Index
		m.providerTotal = evt.Total
		m.currentProvider = ""
		m.currentLabel = ""
	}
}

func (m *providerProgressModel) resetForSyncPhase() {
	m.phaseLabel = "Syncing model catalogs"
	m.currentProvider = ""
	m.currentLabel = "models.dev"
	m.currentAction = provider.ProviderProgressFetchMeta
	m.providerIndex = 0
	m.providerTotal = 0
	m.completed = nil
}

func (m *providerProgressModel) View() string {
	return m.progressContent()
}

func (m *providerProgressModel) progressContent() string {
	if m.err != nil {
		return fmt.Sprintf("\n✗ %v\n", m.err)
	}

	var b strings.Builder
	b.WriteString("\n")

	b.WriteString(m.spinner.View())
	b.WriteString(" ")
	b.WriteString(progressMutedStyle.Render(shortPhaseLabel(m.phaseLabel)))
	if m.providerTotal > 0 {
		b.WriteString(progressDimStyle.Render(" · "))
		b.WriteString(progressActiveStyle.Render(fmt.Sprintf("%d/%d", m.displayProviderIndex(), m.providerTotal)))
	}
	current := currentProviderName(m.currentProvider, m.currentLabel)
	if current != "" {
		b.WriteString(progressDimStyle.Render(" · "))
		b.WriteString(progressActiveStyle.Render(current))
	}
	b.WriteString(progressDimStyle.Render(" · "))
	b.WriteString(progressMutedStyle.Render(formatCompactElapsed(m.stopwatch.Elapsed())))
	b.WriteString("\n")

	lines := 3

	b.WriteString("  ")
	b.WriteString(renderMinimalBar(m.progressPercent()))
	b.WriteString("\n")

	for _, line := range tailCompletedLines(m.completed, maxCompletedProviderLines) {
		b.WriteString(line)
		b.WriteString("\n")
		lines++
	}

	m.viewLines = lines
	return b.String()
}

func (m *providerProgressModel) displayProviderIndex() int {
	if m.providerTotal == 0 {
		return 0
	}
	switch m.currentAction {
	case provider.ProviderProgressWorking:
		return m.providerIndex
	default:
		return min(m.providerIndex, m.providerTotal)
	}
}

func (m *providerProgressModel) progressPercent() float64 {
	if m.providerTotal <= 0 {
		return 0
	}
	completed := m.providerIndex
	if m.currentAction == provider.ProviderProgressWorking {
		completed = m.providerIndex - 1
	}
	if completed < 0 {
		completed = 0
	}
	if completed > m.providerTotal {
		completed = m.providerTotal
	}
	return float64(completed) / float64(m.providerTotal)
}

func clearProviderProgress(lines int) {
	if !isInteractiveStderr() || lines <= 0 {
		return
	}
	fmt.Fprintf(os.Stderr, "\033[%dF\r\033[J", lines)
}

func (m *providerProgressModel) runConnectPhase() {
	result, err := provider.BootstrapProvidersWithOptions(provider.BootstrapOptions{
		Force: m.force,
		Reporter: func(evt provider.ProviderProgressEvent) {
			m.progressCh <- evt
		},
	})
	m.doneCh <- providerStepDoneMsg{
		phase:     provider.ProviderProgressConnect,
		bootstrap: &result,
		err:       err,
	}
}

func (m *providerProgressModel) runSyncPhase() {
	result, err := settings.RunModelsSyncWithReporter(func(evt provider.ProviderProgressEvent) {
		m.progressCh <- evt
	})
	m.doneCh <- providerStepDoneMsg{
		phase: provider.ProviderProgressSync,
		sync:  &result,
		err:   err,
	}
}

func applyStopwatchCmd(sw stopwatch.Model, cmd tea.Cmd) stopwatch.Model {
	for _, msg := range drainTeaCmd(cmd) {
		sw, _ = sw.Update(msg)
	}
	return sw
}

func drainTeaCmd(cmd tea.Cmd) []tea.Msg {
	if cmd == nil {
		return nil
	}
	return drainTeaMsg(cmd())
}

func drainTeaMsg(msg tea.Msg) []tea.Msg {
	if msg == nil {
		return nil
	}
	switch batch := msg.(type) {
	case tea.BatchMsg:
		var msgs []tea.Msg
		for _, c := range batch {
			msgs = append(msgs, drainTeaCmd(c)...)
		}
		return msgs
	}
	val := reflect.ValueOf(msg)
	if val.Kind() == reflect.Slice && val.Type().Elem().Kind() == reflect.Func {
		var msgs []tea.Msg
		for i := 0; i < val.Len(); i++ {
			if c, ok := val.Index(i).Interface().(tea.Cmd); ok {
				msgs = append(msgs, drainTeaCmd(c)...)
			}
		}
		if len(msgs) > 0 {
			return msgs
		}
	}
	return []tea.Msg{msg}
}
