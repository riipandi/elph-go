package provider

// TotalEnabledModels counts enabled models across enabled providers without allocating.
func (c Catalog) TotalEnabledModels() int {
	total := 0
	for _, reg := range c.Providers {
		if !ProviderConfigEnabled(reg.Config) {
			continue
		}
		total += EnabledModelCount(reg)
	}
	return total
}

// SlimModel keeps only fields needed for model picker labels and filtering.
func SlimModel(model ResolvedModel) ResolvedModel {
	return ResolvedModel{
		ID:            model.ID,
		Enabled:       model.Enabled,
		Name:          model.Name,
		ProviderID:    model.ProviderID,
		ProviderName:  model.ProviderName,
		API:           model.API,
		Reasoning:     model.Reasoning,
		ContextWindow: model.ContextWindow,
		Input:         model.Input,
	}
}

// TrimCatalogForRuntime drops disabled models and strips heavy metadata from
// non-active models. The active model retains full compat/thinking metadata.
func TrimCatalogForRuntime(catalog Catalog, activeProviderID, activeModelID string) Catalog {
	out := Catalog{Dir: catalog.Dir}
	for _, reg := range catalog.Providers {
		if !ProviderConfigEnabled(reg.Config) {
			continue
		}
		trimmed := RegisteredProvider{
			ID: reg.ID,
			Config: FileConfig{
				Enabled: reg.Config.Enabled,
				Name:    reg.Config.Name,
			},
		}
		for _, model := range reg.Models {
			if !model.Enabled {
				continue
			}
			if reg.ID == activeProviderID && model.ID == activeModelID {
				trimmed.Models = append(trimmed.Models, model)
			} else {
				trimmed.Models = append(trimmed.Models, SlimModel(model))
			}
		}
		if len(trimmed.Models) > 0 {
			out.Providers = append(out.Providers, trimmed)
		}
	}
	return out
}
