package provider

import (
	"encoding/json"
	"strconv"
	"strings"

	"github.com/riipandi/elph/pkg/ai/protocol"
)

// ThinkingMapState describes a per-level entry in thinkingLevelMap.
type ThinkingMapState int

const (
	ThinkingMapDefault ThinkingMapState = iota
	ThinkingMapUnsupported
	ThinkingMapExplicit
)

// ThinkingMapValue is a parsed thinkingLevelMap entry.
type ThinkingMapValue struct {
	State ThinkingMapState
	Value string
}

var defaultThinkingBudgets = map[protocol.ThinkingLevel]int{
	protocol.ThinkingMinimal: 1024,
	protocol.ThinkingLow:     4096,
	protocol.ThinkingMedium:  10240,
	protocol.ThinkingHigh:    32768,
	protocol.ThinkingXHigh:   65536,
}

var thinkingLevelOrder = []protocol.ThinkingLevel{
	protocol.ThinkingOff,
	protocol.ThinkingMinimal,
	protocol.ThinkingLow,
	protocol.ThinkingMedium,
	protocol.ThinkingHigh,
	protocol.ThinkingXHigh,
}

// ParseThinkingLevelMap decodes Pi-style tristate thinkingLevelMap values.
func ParseThinkingLevelMap(raw map[string]json.RawMessage) map[protocol.ThinkingLevel]ThinkingMapValue {
	if len(raw) == 0 {
		return nil
	}
	out := make(map[protocol.ThinkingLevel]ThinkingMapValue, len(raw))
	for key, value := range raw {
		level, ok := parseThinkingLevelKey(key)
		if !ok {
			continue
		}
		out[level] = parseThinkingLevelMapEntry(value)
	}
	return out
}

func parseThinkingLevelKey(key string) (protocol.ThinkingLevel, bool) {
	switch strings.TrimSpace(key) {
	case string(protocol.ThinkingOff):
		return protocol.ThinkingOff, true
	case string(protocol.ThinkingMinimal):
		return protocol.ThinkingMinimal, true
	case string(protocol.ThinkingLow):
		return protocol.ThinkingLow, true
	case string(protocol.ThinkingMedium):
		return protocol.ThinkingMedium, true
	case string(protocol.ThinkingHigh):
		return protocol.ThinkingHigh, true
	case string(protocol.ThinkingXHigh):
		return protocol.ThinkingXHigh, true
	default:
		return "", false
	}
}

func parseThinkingLevelMapEntry(raw json.RawMessage) ThinkingMapValue {
	if len(raw) == 0 {
		return ThinkingMapValue{State: ThinkingMapDefault}
	}
	if string(raw) == "null" {
		return ThinkingMapValue{State: ThinkingMapUnsupported}
	}
	var value string
	if err := json.Unmarshal(raw, &value); err != nil {
		return ThinkingMapValue{State: ThinkingMapDefault}
	}
	return ThinkingMapValue{State: ThinkingMapExplicit, Value: strings.TrimSpace(value)}
}

// IsThinkingLevelSupported reports whether level is available for model.
func IsThinkingLevelSupported(level protocol.ThinkingLevel, model ResolvedModel) bool {
	if !model.Reasoning && level != protocol.ThinkingOff {
		return false
	}
	if entry, ok := model.ThinkingLevelMap[level]; ok && entry.State == ThinkingMapUnsupported {
		return false
	}
	return true
}

// ClampThinkingLevel returns the nearest supported thinking level for the model.
func ClampThinkingLevel(level protocol.ThinkingLevel, model ResolvedModel) protocol.ThinkingLevel {
	if IsThinkingLevelSupported(level, model) {
		return level
	}
	idx := thinkingLevelIndex(level)
	for i := idx + 1; i < len(thinkingLevelOrder); i++ {
		candidate := thinkingLevelOrder[i]
		if IsThinkingLevelSupported(candidate, model) {
			return candidate
		}
	}
	for i := idx - 1; i >= 0; i-- {
		candidate := thinkingLevelOrder[i]
		if IsThinkingLevelSupported(candidate, model) {
			return candidate
		}
	}
	return protocol.ThinkingOff
}

// NextSupportedThinkingLevel cycles forward, skipping unsupported levels.
func NextSupportedThinkingLevel(current protocol.ThinkingLevel, model ResolvedModel) protocol.ThinkingLevel {
	next := current
	for range len(thinkingLevelOrder) {
		next = protocol.NextThinkingLevel(next)
		if IsThinkingLevelSupported(next, model) {
			return next
		}
	}
	return current
}

// PrevSupportedThinkingLevel cycles backward, skipping unsupported levels.
func PrevSupportedThinkingLevel(current protocol.ThinkingLevel, model ResolvedModel) protocol.ThinkingLevel {
	next := current
	for range len(thinkingLevelOrder) {
		next = protocol.PrevThinkingLevel(next)
		if IsThinkingLevelSupported(next, model) {
			return next
		}
	}
	return current
}

func thinkingLevelIndex(level protocol.ThinkingLevel) int {
	for i, candidate := range thinkingLevelOrder {
		if candidate == level {
			return i
		}
	}
	return -1
}

// ResolveThinking maps the UI thinking level to provider request parameters.
func ResolveThinking(model ResolvedModel, level protocol.ThinkingLevel, budgets map[string]int) ThinkingConfig {
	level = ClampThinkingLevel(level, model)
	if level == protocol.ThinkingOff || !model.Reasoning {
		return ThinkingConfig{}
	}

	entry := mapEntry(model.ThinkingLevelMap, level)
	if entry.State == ThinkingMapUnsupported {
		return ThinkingConfig{}
	}

	format := thinkingFormatFor(model)
	adaptive := model.API == APIAnthropicMessages && model.Compat.ForceAdaptiveThinking

	cfg := ThinkingConfig{
		Enabled:        true,
		ThinkingFormat: format,
	}
	if adaptive {
		cfg.Adaptive = true
		cfg.AdaptiveEffort = resolveAdaptiveEffort(level, entry)
		return cfg
	}

	switch model.API {
	case APIAnthropicMessages:
		cfg.BudgetTokens = resolveBudgetTokens(level, entry, budgets)
	case APIOpenAICompletions:
		cfg = resolveOpenAIThinking(cfg, model, level, entry)
	default:
		cfg.Enabled = false
	}
	return cfg
}

func mapEntry(m map[protocol.ThinkingLevel]ThinkingMapValue, level protocol.ThinkingLevel) ThinkingMapValue {
	if m == nil {
		return ThinkingMapValue{State: ThinkingMapDefault}
	}
	entry, ok := m[level]
	if !ok {
		return ThinkingMapValue{State: ThinkingMapDefault}
	}
	return entry
}

func thinkingFormatFor(model ResolvedModel) ThinkingFormat {
	switch ThinkingFormat(strings.TrimSpace(model.Compat.ThinkingFormat)) {
	case ThinkingFormatOpenRouter:
		return ThinkingFormatOpenRouter
	case ThinkingFormatQwen:
		return ThinkingFormatQwen
	case ThinkingFormatDeepSeek:
		return ThinkingFormatDeepSeek
	default:
		return ThinkingFormatReasoningEffort
	}
}

func resolveAdaptiveEffort(level protocol.ThinkingLevel, entry ThinkingMapValue) string {
	if entry.State == ThinkingMapExplicit && entry.Value != "" {
		return entry.Value
	}
	switch level {
	case protocol.ThinkingMinimal, protocol.ThinkingLow:
		return "low"
	case protocol.ThinkingMedium:
		return "medium"
	case protocol.ThinkingHigh:
		return "high"
	case protocol.ThinkingXHigh:
		return "max"
	default:
		return "medium"
	}
}

func resolveBudgetTokens(level protocol.ThinkingLevel, entry ThinkingMapValue, budgets map[string]int) int {
	if entry.State == ThinkingMapExplicit && entry.Value != "" {
		if tokens, err := strconv.Atoi(entry.Value); err == nil && tokens > 0 {
			return tokens
		}
	}
	if token, ok := budgets[string(level)]; ok && token > 0 {
		return token
	}
	if token, ok := defaultThinkingBudgets[level]; ok && token > 0 {
		return token
	}
	return defaultThinkingBudgets[protocol.ThinkingHigh]
}

func resolveOpenAIThinking(cfg ThinkingConfig, model ResolvedModel, level protocol.ThinkingLevel, entry ThinkingMapValue) ThinkingConfig {
	effort := resolveReasoningEffort(level, entry)
	switch cfg.ThinkingFormat {
	case ThinkingFormatQwen:
		cfg.EnableThinking = true
		return cfg
	case ThinkingFormatOpenRouter:
		if effort != "" {
			cfg.ReasoningEffort = effort
		}
		return cfg
	case ThinkingFormatDeepSeek:
		// DeepSeek reasoner models infer thinking from model id; no extra param required.
		return cfg
	default:
		if model.Compat.ReasoningEffortSupported() && effort != "" {
			cfg.ReasoningEffort = effort
			return cfg
		}
		cfg.EnableThinking = true
		return cfg
	}
}

func resolveReasoningEffort(level protocol.ThinkingLevel, entry ThinkingMapValue) string {
	if entry.State == ThinkingMapExplicit && entry.Value != "" {
		return entry.Value
	}
	switch level {
	case protocol.ThinkingMinimal:
		return "minimal"
	case protocol.ThinkingLow:
		return "low"
	case protocol.ThinkingMedium:
		return "medium"
	case protocol.ThinkingHigh:
		return "high"
	case protocol.ThinkingXHigh:
		return "high"
	default:
		return ""
	}
}
