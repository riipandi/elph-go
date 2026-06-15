package settings

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/riipandi/elph/internal/projectdir"
	"github.com/riipandi/elph/pkg/jsoncfg"
)

func homeSettingsDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, defaultElphHomeDir), nil
}

// activeSettingsPath picks settings.json over settings.jsonc when both exist.
func activeSettingsPath(dir string) (path string, ok bool) {
	jsonPath := filepath.Join(dir, settingsFileName)
	if settingsFileExists(jsonPath) {
		return jsonPath, true
	}
	jsoncPath := filepath.Join(dir, settingsJSONCFileName)
	if settingsFileExists(jsoncPath) {
		return jsoncPath, true
	}
	return "", false
}

func readSettingsFile(path string) (Settings, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return Settings{}, fmt.Errorf("read settings %q: %w", path, err)
	}
	var cfg Settings
	if err := jsoncfg.Unmarshal(raw, &cfg); err != nil {
		return Settings{}, fmt.Errorf("decode settings %q: %w", path, err)
	}
	return cfg, nil
}

func loadSettingsDir(dir string) (Settings, bool, error) {
	path, ok := activeSettingsPath(dir)
	if !ok {
		return Settings{}, false, nil
	}
	cfg, err := readSettingsFile(path)
	if err != nil {
		return Settings{}, false, err
	}
	return cfg, true, nil
}

// loadHomeSettings reads ~/.elph settings merged onto defaults (no project overlay).
func loadHomeSettings() (Settings, error) {
	cfg := defaultSettings()

	homeDir, err := homeSettingsDir()
	if err != nil {
		return Settings{}, err
	}
	if homeCfg, ok, err := loadSettingsDir(homeDir); err != nil {
		return Settings{}, err
	} else if ok {
		cfg = mergeSettings(cfg, homeCfg)
	}
	return cfg.withDefaults(), nil
}

// LoadFor reads settings merged from defaults, ~/.elph, then workDir/.agents/elph.
// Project settings override home settings field-by-field.
func LoadFor(workDir string) (Settings, error) {
	cfg := defaultSettings()

	homeDir, err := homeSettingsDir()
	if err != nil {
		return Settings{}, err
	}
	if homeCfg, ok, err := loadSettingsDir(homeDir); err != nil {
		return Settings{}, err
	} else if ok {
		cfg = mergeSettings(cfg, homeCfg)
	}

	workDir = strings.TrimSpace(workDir)
	if workDir != "" {
		if abs, err := filepath.Abs(workDir); err == nil {
			workDir = abs
		}
		if projectCfg, ok, err := loadSettingsDir(projectdir.Root(workDir)); err != nil {
			return Settings{}, err
		} else if ok {
			cfg = mergeSettings(cfg, projectCfg)
		}
	}

	return cfg.withDefaults(), nil
}
