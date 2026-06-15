package provider

import (
	"errors"
	"fmt"
	"strings"
)

// IsCredentialError reports whether err is a missing or unresolved API key.
func IsCredentialError(err error) bool {
	if errors.Is(err, ErrMissingAPIKey) {
		return true
	}
	if err == nil {
		return false
	}
	msg := err.Error()
	return strings.Contains(msg, "resolve apiKey") ||
		strings.Contains(msg, "environment variable")
}

// CredentialHint returns setup guidance for a provider missing credentials.
func CredentialHint(reg RegisteredProvider) string {
	path := fmt.Sprintf("~/.elph/providers/%s.json", reg.ID)
	raw := strings.TrimSpace(reg.Config.APIKey)
	switch {
	case raw == "":
		return fmt.Sprintf("add apiKey to %s", path)
	case strings.HasPrefix(raw, "!"):
		return fmt.Sprintf("configure apiKey command in %s", path)
	case strings.Contains(raw, "$"):
		return fmt.Sprintf("set the environment variable referenced by apiKey in %s", path)
	default:
		return fmt.Sprintf("configure apiKey in %s", path)
	}
}

// BuildModelConfig resolves a runtime provider when credentials are available.
// On credential errors it still returns display metadata with Provider == nil.
func BuildModelConfig(catalog Catalog, reg RegisteredProvider, model ResolvedModel) (Config, error) {
	cfg, err := SelectModel(catalog, reg, model)
	if err == nil {
		return cfg, nil
	}
	if !IsCredentialError(err) {
		return Config{Catalog: catalog}, err
	}
	return metadataConfig(catalog, reg, model), err
}

func metadataConfig(catalog Catalog, reg RegisteredProvider, model ResolvedModel) Config {
	providerName := reg.Config.Name
	if providerName == "" {
		providerName = reg.ID
	}
	if model.ProviderName != "" {
		providerName = model.ProviderName
	}
	modelName := model.Name
	if modelName == "" {
		modelName = model.ID
	}
	return Config{
		ModelID:       model.ID,
		ModelName:     modelName,
		ProviderID:    reg.ID,
		ProviderName:  providerName,
		ContextWindow: model.ContextWindow,
		MaxTokens:     model.MaxTokens,
		Input:         model.Input,
		Cost:          model.Cost,
		Catalog:       TrimCatalogForRuntime(catalog, reg.ID, model.ID),
	}
}
