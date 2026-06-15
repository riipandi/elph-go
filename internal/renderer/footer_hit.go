package renderer

import (
	"fmt"
	"path/filepath"

	"charm.land/lipgloss/v2"
	"github.com/riipandi/elph/internal/uiconst"
)

type footerZone int

const (
	footerZoneModel footerZone = iota
	footerZoneThinking
	footerZoneMode
	footerZoneWorkdir
	footerZoneSession
	footerZoneBranch
	footerZoneGit
)

type footerHitRect struct {
	zone   footerZone
	startX int
	endX   int
	row    int
}

func (m Model) footerHitRects() []footerHitRect {
	cw := footerContentWidth(m.width)
	var rects []footerHitRect
	add := func(row int, zone footerZone, start, width int) {
		if width <= 0 {
			return
		}
		rects = append(rects, footerHitRect{
			zone:   zone,
			startX: start,
			endX:   start + width,
			row:    row,
		})
	}

	modelW := lipgloss.Width(m.modelName)
	add(0, footerZoneModel, 0, modelW)

	thinkStart := modelW + lipgloss.Width(metaSty.Render(fmt.Sprintf(" | %s", m.provider)))
	thinkW := lipgloss.Width(metaSty.Render(fmt.Sprintf(" | T: %s", m.thinkingLevel)))
	add(0, footerZoneThinking, thinkStart, thinkW)

	wd := lipgloss.Width(primaryBoldSty.Render(filepath.Base(m.workDir)))
	add(1, footerZoneWorkdir, 0, wd)

	sid := fmt.Sprintf(" [%s] ", m.sessionID.Suffix())
	sidStart := wd
	sidW := lipgloss.Width(sidSty.Render(sid))
	add(1, footerZoneSession, sidStart, sidW)

	modeStart := sidStart + sidW
	modeW := lipgloss.Width(lipgloss.NewStyle().Foreground(uiconst.ModeBorderColor(m.mode)).Bold(true).Render(string(m.mode)))
	add(1, footerZoneMode, modeStart, modeW)

	gitStr := "[-]"
	if m.gitAdded > 0 || m.gitDeleted > 0 {
		gitStr = fmt.Sprintf("[+%d -%d]", m.gitAdded, m.gitDeleted)
	}
	branchPart := fmt.Sprintf("turn: %d | %s ", m.turnCount, m.branch)
	branchW := lipgloss.Width(primarySty.Render(branchPart))
	gitW := lipgloss.Width(gitStr)
	row2RightStart := max(cw-branchW-gitW, 0)
	add(1, footerZoneBranch, row2RightStart, branchW)
	add(1, footerZoneGit, row2RightStart+branchW, gitW)

	return rects
}

func (m Model) footerZoneAt(x, rowY int) (footerZone, bool) {
	contentX := x - 1 // footer left padding
	if contentX < 0 {
		return 0, false
	}
	row := 0
	if rowY > 0 {
		row = 1
	}
	for _, rect := range m.footerHitRects() {
		if rect.row != row {
			continue
		}
		if contentX >= rect.startX && contentX < rect.endX {
			return rect.zone, true
		}
	}
	return 0, false
}
