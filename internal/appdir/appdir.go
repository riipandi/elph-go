package appdir

import (
	"os"
	"path/filepath"
)

const (
	// configDirName is the XDG config directory name.
	configDirName = ".elph"

	// dataDirName is the XDG data directory name.
	dataDirName = "elph"
)

// ConfigDir returns the configuration directory path (~/.elph or XDG_CONFIG_HOME/elph).
func ConfigDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, configDirName), nil
}

// DataDir returns the data directory path (~/.local/share/elph or XDG_DATA_HOME/elph).
func DataDir() (string, error) {
	if xdg := os.Getenv("XDG_DATA_HOME"); xdg != "" {
		return filepath.Join(xdg, dataDirName), nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".local", "share", dataDirName), nil
}

// EnsureDataDir creates the data directory if it doesn't exist.
func EnsureDataDir() (string, error) {
	dir, err := DataDir()
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	return dir, nil
}

// EnsureConfigDir creates the config directory if it doesn't exist.
func EnsureConfigDir() (string, error) {
	dir, err := ConfigDir()
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	return dir, nil
}

// AttachmentsDir returns the global attachments directory path (~/.local/share/elph/attachments).
func AttachmentsDir() (string, error) {
	dataDir, err := DataDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dataDir, "attachments"), nil
}

// VersionPath returns the version.json file path in the data directory.
func VersionPath() (string, error) {
	dataDir, err := DataDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dataDir, "version.json"), nil
}

// DatabasePath returns the metadata.db file path in the data directory.
func DatabasePath() (string, error) {
	dataDir, err := DataDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dataDir, "metadata.db"), nil
}

// LogsDir returns the logs directory path in the data directory.
func LogsDir() (string, error) {
	dataDir, err := DataDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dataDir, "logs"), nil
}

// ProvidersDir returns the providers directory path in the config directory.
func ProvidersDir() (string, error) {
	configDir, err := ConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(configDir, "providers"), nil
}
