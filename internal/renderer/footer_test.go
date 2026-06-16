package renderer

import (
	"testing"

	"github.com/riipandi/elph/internal/appconst"
	"github.com/riipandi/elph/pkg/ai/provider"
	"github.com/stretchr/testify/require"
)

func TestFooterShowsTurnCountAndCost(t *testing.T) {
	m := New()
	m.width = 100
	m.turnCount = 3
	m.sessionCost = 1.25
	m.tokensUsed = 131_072
	m.contextWindow = 262_144

	footer := stripANSI(m.footerView())
	require.Contains(t, footer, "turn: 3")
	require.Contains(t, footer, "$1.25")
	require.Contains(t, footer, "50.0%")
}

func TestFooterImageLabelReflectsModelCapability(t *testing.T) {
	m := New()
	m.width = 100
	m.modelSupportsImage = true
	require.Contains(t, stripANSI(m.footerView()), "IMG")

	m.modelSupportsImage = false
	require.Contains(t, stripANSI(m.footerView()), "—")
}

func TestApplyTurnUsageAccumulates(t *testing.T) {
	m := New()
	m.modelCost = provider.Cost{Input: 3, Output: 15}
	m.contextWindow = 1000

	m = m.applyTurnUsage(provider.TurnUsage{InputTokens: 100, OutputTokens: 50})
	require.Equal(t, 150, m.tokensUsed)
	require.InDelta(t, 0.00105, m.sessionCost, 0.00001)
}

func TestFooterClickCyclesMode(t *testing.T) {
	m := testInputModel(t)
	m.mode = appconst.ModeBuild
	m.ready = true
	m = m.syncLayout(false)

	top := m.footerViewportTop()
	rects := m.footerHitRects()
	var modeRect footerHitRect
	for _, r := range rects {
		if r.zone == footerZoneMode {
			modeRect = r
			break
		}
	}
	require.NotZero(t, modeRect.endX)

	updated, cmd := m.handleFooterClick(modeRect.startX+1, top+modeRect.row)
	m = updated
	require.Nil(t, cmd)
	require.Equal(t, appconst.ModePlan, m.mode)
}

func TestFooterClickCopiesSessionID(t *testing.T) {
	m := testInputModel(t)
	m.ready = true
	m = m.syncLayout(false)

	top := m.footerViewportTop()
	var sessRect footerHitRect
	for _, r := range m.footerHitRects() {
		if r.zone == footerZoneSession {
			sessRect = r
			break
		}
	}
	require.NotZero(t, sessRect.endX)

	updated, cmd := m.handleFooterClick(sessRect.startX+1, top+sessRect.row)
	m = updated
	require.Nil(t, cmd)
	require.Contains(t, stripANSI(m.messages[len(m.messages)-1].text), "Copied session id")
}
