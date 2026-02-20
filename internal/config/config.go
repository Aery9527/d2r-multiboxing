// Package config provides application configuration management.
// All files are stored under ~/.d2r-multiboxing/ by default.
// The directory can be overridden via the D2R_MULTIBOXING_HOME environment variable.
package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"d2r-multiboxing/internal/d2r"
)

const (
	// configDirName is the directory name under user home.
	configDirName = ".d2r-multiboxing"

	// configFileName is the config file name.
	configFileName = "config.json"

	// accountsFileName is the accounts CSV file name.
	accountsFileName = "accounts.csv"

	// HomeDirEnv is the environment variable to override the config directory.
	HomeDirEnv = "D2R_MULTIBOXING_HOME"
)

// SwitcherConfig holds the window switcher configuration.
type SwitcherConfig struct {
	// Enabled controls whether the window switcher is active.
	Enabled bool `json:"enabled"`

	// Modifiers is the list of modifier keys (e.g., "ctrl", "alt", "shift").
	Modifiers []string `json:"modifiers,omitempty"`

	// Key is the trigger key name (e.g., "Tab", "F1", "XButton1").
	Key string `json:"key"`
}

// Config represents the application configuration.
type Config struct {
	// D2RPath is the path to D2R.exe.
	D2RPath string `json:"d2r_path"`

	// LaunchDelay is the delay in seconds between launching each account.
	LaunchDelay int `json:"launch_delay"`

	// Switcher holds the window switcher settings.
	Switcher *SwitcherConfig `json:"switcher,omitempty"`
}

// DefaultConfig returns a Config with default values.
func DefaultConfig() Config {
	return Config{
		D2RPath:     d2r.DefaultGamePath,
		LaunchDelay: 5,
	}
}

// Dir returns the config directory path.
// Priority: D2R_MULTIBOXING_HOME env > ~/.d2r-multiboxing
func Dir() (string, error) {
	if dir := os.Getenv(HomeDirEnv); dir != "" {
		return dir, nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user home directory: %w", err)
	}
	return filepath.Join(home, configDirName), nil
}

// Path returns the full path to the config file.
func Path() (string, error) {
	dir, err := Dir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, configFileName), nil
}

// AccountsPath returns the full path to the accounts CSV file.
func AccountsPath() (string, error) {
	dir, err := Dir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, accountsFileName), nil
}

// Load reads the config file. If the file does not exist, it creates one with default values.
func Load() (*Config, error) {
	cfgPath, err := Path()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(cfgPath)
	if err != nil {
		if os.IsNotExist(err) {
			cfg := DefaultConfig()
			if writeErr := Save(&cfg); writeErr != nil {
				return nil, fmt.Errorf("failed to create default config: %w", writeErr)
			}
			return &cfg, nil
		}
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	return &cfg, nil
}

// Save writes the config to file, creating the directory if needed.
func Save(cfg *Config) error {
	cfgPath, err := Path()
	if err != nil {
		return err
	}

	dir := filepath.Dir(cfgPath)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(cfgPath, data, 0o600); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	return nil
}
