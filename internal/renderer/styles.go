package renderer

import (
	"charm.land/lipgloss/v2"
	"github.com/riipandi/elph/internal/appconst"
	"github.com/riipandi/elph/internal/uiconst"
)

var modeList = []appconst.AgentMode{
	appconst.ModeBuild,
	appconst.ModePlan,
	appconst.ModeAsk,
	appconst.ModeBrave,
}

var (
	inputBorderByMode             = make(map[appconst.AgentMode]lipgloss.Style, len(modeList))
	inputBorderAttachedByMode     = make(map[appconst.AgentMode]lipgloss.Style, len(modeList))
	paletteBorderByMode           = make(map[appconst.AgentMode]lipgloss.Style, len(modeList))
	modelSelectorListBorderByMode = make(map[appconst.AgentMode]lipgloss.Style, len(modeList))
)

func init() {
	for _, mode := range modeList {
		color := uiconst.ModeBorderColor(mode)
		inputBorderByMode[mode] = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(color).
			Padding(0, 1)
		inputBorderAttachedByMode[mode] = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(color).
			BorderTop(false).
			Padding(0, 1)
		paletteBorderByMode[mode] = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(color).
			BorderBottom(false).
			Padding(0, 1)
		modelSelectorListBorderByMode[mode] = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(color).
			BorderBottom(false).
			Padding(0, 1).
			PaddingBottom(1)
	}
}

func cachedInputBorder(mode appconst.AgentMode) lipgloss.Style {
	if style, ok := inputBorderByMode[mode]; ok {
		return style
	}
	return inputBorderByMode[appconst.ModeBuild]
}

func cachedInputBorderAttached(mode appconst.AgentMode) lipgloss.Style {
	if style, ok := inputBorderAttachedByMode[mode]; ok {
		return style
	}
	return inputBorderAttachedByMode[appconst.ModeBuild]
}

func paletteBorder(mode appconst.AgentMode) lipgloss.Style {
	if style, ok := paletteBorderByMode[mode]; ok {
		return style
	}
	return paletteBorderByMode[appconst.ModeBuild]
}

func modelSelectorListBorder(mode appconst.AgentMode) lipgloss.Style {
	if style, ok := modelSelectorListBorderByMode[mode]; ok {
		return style
	}
	return modelSelectorListBorderByMode[appconst.ModeBuild]
}
