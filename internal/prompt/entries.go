package prompt

import (
	inttools "github.com/riipandi/elph/internal/tools"
	"github.com/riipandi/elph/pkg/tools"
)

// Entry describes any tool available to the agent (built-in, MCP, plugin, etc.).
type Entry struct {
	Name                 string
	Section              string
	DefaultApproval      string
	Description          string
	RequiresConfirmation bool
}

type toolSection struct {
	Name  string
	Tools []Entry
}

// TemplateData is passed to system prompt templates.
type TemplateData struct {
	AvailableTools string
}

var pkgSectionByCategory = map[tools.Category]string{
	tools.CategoryFile:            "File Tools",
	tools.CategoryShell:           "Shell Tools",
	tools.CategoryWeb:             "Web Tools",
	tools.CategoryPlanMode:        "Plan Mode",
	tools.CategoryStateManagement: "State Management",
	tools.CategoryCollaboration:   "Collaboration Tools",
}

var pkgCategoryOrder = []tools.Category{
	tools.CategoryFile,
	tools.CategoryShell,
	tools.CategoryWeb,
	tools.CategoryPlanMode,
	tools.CategoryStateManagement,
	tools.CategoryCollaboration,
}

// EntryFromBuiltin converts a built-in tool definition into a catalog entry.
func EntryFromBuiltin(def tools.Definition) Entry {
	return Entry{
		Name:                 def.Name,
		Section:              pkgSectionByCategory[def.Category],
		DefaultApproval:      string(def.DefaultApproval),
		Description:          def.Description,
		RequiresConfirmation: def.RequiresConfirmation,
	}
}

// ExternalEntry creates an entry for MCP, plugin, or other externally connected tools.
func ExternalEntry(name, section, approval, description string) Entry {
	return Entry{
		Name:            name,
		Section:         section,
		DefaultApproval: approval,
		Description:     description,
	}
}

func catalogEntries(explicit []Entry) []Entry {
	if explicit == nil {
		explicit = entriesFromExposedBuiltins()
	}

	entries := make([]Entry, 0, len(explicit)+len(inttools.Diagnostic()))
	entries = append(entries, explicit...)

	for _, def := range inttools.Diagnostic() {
		if hasEntryName(entries, def.Name) {
			continue
		}
		entries = append(entries, entryFromDiagnostic(def))
	}

	return entries
}

func entriesFromExposedBuiltins() []Entry {
	defs := make([]tools.Definition, 0, 4)
	for _, def := range tools.All() {
		if tools.IsProviderExposed(def.Name) {
			defs = append(defs, def)
		}
	}
	return entriesFromBuiltins(defs)
}

func entriesFromBuiltins(defs []tools.Definition) []Entry {
	entries := make([]Entry, 0, len(defs))

	for _, category := range pkgCategoryOrder {
		section := pkgSectionByCategory[category]
		for _, def := range defs {
			if def.Category != category {
				continue
			}
			entry := EntryFromBuiltin(def)
			entry.Section = section
			entries = append(entries, entry)
		}
	}

	return entries
}

func entryFromDiagnostic(def inttools.Definition) Entry {
	return Entry{
		Name:            def.Name,
		Section:         "Diagnostic Tools",
		DefaultApproval: string(def.DefaultApproval),
		Description:     def.Description,
	}
}

func hasEntryName(entries []Entry, name string) bool {
	for _, entry := range entries {
		if entry.Name == name {
			return true
		}
	}
	return false
}

func groupBySection(entries []Entry) []toolSection {
	sectionOrder := make([]string, 0)
	seen := make(map[string]bool)
	bySection := make(map[string][]Entry)

	for _, entry := range entries {
		if !seen[entry.Section] {
			seen[entry.Section] = true
			sectionOrder = append(sectionOrder, entry.Section)
		}
		bySection[entry.Section] = append(bySection[entry.Section], entry)
	}

	sections := make([]toolSection, 0, len(sectionOrder))
	for _, name := range sectionOrder {
		sections = append(sections, toolSection{
			Name:  name,
			Tools: bySection[name],
		})
	}

	return sections
}
