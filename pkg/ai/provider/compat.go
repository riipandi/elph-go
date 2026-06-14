package provider

// Compat holds provider-specific API compatibility overrides (Pi-style).
type Compat struct {
	SupportsDeveloperRole    *bool  `json:"supportsDeveloperRole,omitempty"`
	SupportsReasoningEffort  *bool  `json:"supportsReasoningEffort,omitempty"`
	SupportsUsageInStreaming *bool  `json:"supportsUsageInStreaming,omitempty"`
	ForceAdaptiveThinking    bool   `json:"forceAdaptiveThinking,omitempty"`
	AllowEmptySignature      bool   `json:"allowEmptySignature,omitempty"`
	ThinkingFormat           string `json:"thinkingFormat,omitempty"`
	MaxTokensField           string `json:"maxTokensField,omitempty"`
}

func mergeCompat(providerCompat, modelCompat Compat) Compat {
	out := providerCompat
	if modelCompat.SupportsDeveloperRole != nil {
		out.SupportsDeveloperRole = modelCompat.SupportsDeveloperRole
	}
	if modelCompat.SupportsReasoningEffort != nil {
		out.SupportsReasoningEffort = modelCompat.SupportsReasoningEffort
	}
	if modelCompat.SupportsUsageInStreaming != nil {
		out.SupportsUsageInStreaming = modelCompat.SupportsUsageInStreaming
	}
	if modelCompat.ForceAdaptiveThinking {
		out.ForceAdaptiveThinking = true
	}
	if modelCompat.AllowEmptySignature {
		out.AllowEmptySignature = true
	}
	if modelCompat.ThinkingFormat != "" {
		out.ThinkingFormat = modelCompat.ThinkingFormat
	}
	if modelCompat.MaxTokensField != "" {
		out.MaxTokensField = modelCompat.MaxTokensField
	}
	return out
}

func (c Compat) supportsDeveloperRole() bool {
	if c.SupportsDeveloperRole != nil {
		return *c.SupportsDeveloperRole
	}
	return true
}

func (c Compat) supportsReasoningEffort() bool {
	if c.SupportsReasoningEffort != nil {
		return *c.SupportsReasoningEffort
	}
	return true
}

func (c Compat) supportsUsageInStreaming() bool {
	if c.SupportsUsageInStreaming != nil {
		return *c.SupportsUsageInStreaming
	}
	return true
}
