package provider

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
)

// BackfillThinkingResult reports provider files that gained thinking metadata.
type BackfillThinkingResult struct {
	Dir        string
	Backfilled []string
}

func thinkingTemplateByProvider(providerID string) (FileConfig, bool) {
	templates := PrimaryProviderTemplates()
	cfg, ok := templates[providerID]
	return cfg, ok
}

func thinkingTemplateModel(providerID, modelID string) (ModelConfig, bool) {
	tmpl, ok := thinkingTemplateByProvider(providerID)
	if !ok {
		return ModelConfig{}, false
	}
	for _, model := range tmpl.Models {
		if model.ID == modelID {
			return model, true
		}
	}
	return ModelConfig{}, false
}

// BackfillProviderThinking fills missing reasoning, thinkingLevelMap, and compat
// fields from built-in templates without overwriting existing user configuration.
func BackfillProviderThinking(providerID string, cfg FileConfig) (FileConfig, bool) {
	tmpl, ok := thinkingTemplateByProvider(providerID)
	if !ok {
		return cfg, false
	}

	changed := false
	nextCompat := backfillCompat(cfg.Compat, tmpl.Compat)
	if !compatEqual(cfg.Compat, nextCompat) {
		cfg.Compat = nextCompat
		changed = true
	}

	tmplModels := indexModelConfigsByID(tmpl.Models)
	for i, model := range cfg.Models {
		tmplModel, ok := tmplModels[model.ID]
		if !ok {
			continue
		}
		updated := backfillModelThinking(model, tmplModel)
		if !modelConfigsEqual(model, updated) {
			cfg.Models[i] = updated
			changed = true
		}
	}
	return cfg, changed
}

// BackfillAllProviderThinking backfills thinking metadata for every provider file in dir.
func BackfillAllProviderThinking(dir string) (BackfillThinkingResult, error) {
	if dir == "" {
		var err error
		dir, err = ProvidersDir()
		if err != nil {
			return BackfillThinkingResult{}, err
		}
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return BackfillThinkingResult{Dir: dir}, nil
		}
		return BackfillThinkingResult{}, err
	}

	result := BackfillThinkingResult{Dir: dir}
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}
		providerID := strings.TrimSuffix(entry.Name(), ".json")
		if providerID == "" {
			continue
		}

		path := filepath.Join(dir, entry.Name())
		raw, err := os.ReadFile(path)
		if err != nil {
			return result, err
		}

		var cfg FileConfig
		if err := json.Unmarshal(raw, &cfg); err != nil {
			return result, err
		}

		updated, changed := BackfillProviderThinking(providerID, cfg)
		if !changed {
			continue
		}

		payload, err := json.MarshalIndent(updated, "", "  ")
		if err != nil {
			return result, err
		}
		payload = append(payload, '\n')
		if err := os.WriteFile(path, payload, 0o644); err != nil {
			return result, err
		}
		result.Backfilled = append(result.Backfilled, entry.Name())
	}
	return result, nil
}

func indexModelConfigsByID(models []ModelConfig) map[string]ModelConfig {
	out := make(map[string]ModelConfig, len(models))
	for _, model := range models {
		if model.ID == "" {
			continue
		}
		out[model.ID] = model
	}
	return out
}

func backfillModelThinking(existing, template ModelConfig) ModelConfig {
	out := existing
	if !out.Reasoning && template.Reasoning {
		out.Reasoning = true
	}
	if len(out.ThinkingLevelMap) == 0 && len(template.ThinkingLevelMap) > 0 {
		out.ThinkingLevelMap = cloneThinkingLevelMap(template.ThinkingLevelMap)
	}
	out.Compat = backfillCompat(out.Compat, template.Compat)
	return out
}

func cloneThinkingLevelMap(src map[string]json.RawMessage) map[string]json.RawMessage {
	if len(src) == 0 {
		return nil
	}
	out := make(map[string]json.RawMessage, len(src))
	for key, value := range src {
		out[key] = append(json.RawMessage(nil), value...)
	}
	return out
}

func backfillCompat(existing, template Compat) Compat {
	out := existing
	if template.ForceAdaptiveThinking {
		out.ForceAdaptiveThinking = out.ForceAdaptiveThinking || template.ForceAdaptiveThinking
	}
	if template.AllowEmptySignature {
		out.AllowEmptySignature = out.AllowEmptySignature || template.AllowEmptySignature
	}
	if out.ThinkingFormat == "" && template.ThinkingFormat != "" {
		out.ThinkingFormat = template.ThinkingFormat
	}
	if out.MaxTokensField == "" && template.MaxTokensField != "" {
		out.MaxTokensField = template.MaxTokensField
	}
	if out.SupportsDeveloperRole == nil && template.SupportsDeveloperRole != nil {
		v := *template.SupportsDeveloperRole
		out.SupportsDeveloperRole = &v
	}
	if out.SupportsReasoningEffort == nil && template.SupportsReasoningEffort != nil {
		v := *template.SupportsReasoningEffort
		out.SupportsReasoningEffort = &v
	}
	if out.SupportsUsageInStreaming == nil && template.SupportsUsageInStreaming != nil {
		v := *template.SupportsUsageInStreaming
		out.SupportsUsageInStreaming = &v
	}
	return out
}

func compatEqual(a, b Compat) bool {
	aa, err := json.Marshal(a)
	if err != nil {
		return false
	}
	bb, err := json.Marshal(b)
	if err != nil {
		return false
	}
	return string(aa) == string(bb)
}
