package settings

import (
	"strings"

	"github.com/riipandi/elph/internal/appconst"
)

// SessionSettings stores per-user UI and runtime preferences.
type SessionSettings struct {
	ProviderID    string `json:"providerId,omitempty"`
	ModelID       string `json:"modelId,omitempty"`
	AgentMode     string `json:"agentMode,omitempty"`
	ThinkingLevel string `json:"thinkingLevel,omitempty"`
}

// AgentMode returns the persisted agent mode, defaulting to build.
func (s Settings) AgentMode() appconst.AgentMode {
	return normalizeAgentMode(s.Session.AgentMode)
}

// ThinkingLevel returns the persisted thinking level, defaulting to high.
func (s Settings) ThinkingLevel() appconst.ThinkingLevel {
	return normalizeThinkingLevel(s.Session.ThinkingLevel)
}

// ActiveProviderID returns the last selected provider id from settings.
func (s Settings) ActiveProviderID() string {
	return strings.TrimSpace(s.Session.ProviderID)
}

// ActiveModelID returns the last selected model id from settings.
func (s Settings) ActiveModelID() string {
	return strings.TrimSpace(s.Session.ModelID)
}

// Update loads home settings, applies mutator, and saves to ~/.elph.
func Update(mutator func(*Settings)) error {
	cfg, err := loadHomeSettings()
	if err != nil {
		return err
	}
	mutator(&cfg)
	return Save(cfg)
}

// SetActiveModel records the active provider/model selection.
func SetActiveModel(providerID, modelID string) error {
	return Update(func(cfg *Settings) {
		cfg.Session.ProviderID = strings.TrimSpace(providerID)
		cfg.Session.ModelID = strings.TrimSpace(modelID)
	})
}

// SetAgentMode records the active agent mode.
func SetAgentMode(mode appconst.AgentMode) error {
	return Update(func(cfg *Settings) {
		cfg.Session.AgentMode = string(normalizeAgentMode(string(mode)))
	})
}

// SetThinkingLevel records the active thinking level.
func SetThinkingLevel(level appconst.ThinkingLevel) error {
	return Update(func(cfg *Settings) {
		cfg.Session.ThinkingLevel = string(normalizeThinkingLevel(string(level)))
	})
}

func normalizeAgentMode(raw string) appconst.AgentMode {
	mode := appconst.AgentMode(strings.TrimSpace(raw))
	switch mode {
	case appconst.ModeBuild, appconst.ModePlan, appconst.ModeAsk, appconst.ModeBrave:
		return mode
	default:
		return appconst.ModeBuild
	}
}

func normalizeThinkingLevel(raw string) appconst.ThinkingLevel {
	level := appconst.ThinkingLevel(strings.TrimSpace(raw))
	switch level {
	case appconst.ThinkingOff,
		appconst.ThinkingMinimal,
		appconst.ThinkingLow,
		appconst.ThinkingMedium,
		appconst.ThinkingHigh,
		appconst.ThinkingXHigh:
		return level
	default:
		return appconst.ThinkingHigh
	}
}
