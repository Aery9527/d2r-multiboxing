// Package config provides application configuration management.
// All files are stored under ~/.d2r-hyper-launcher/ by default.
// The directory can be overridden via the D2R_HYPER_LAUNCHER_HOME environment variable.
package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"d2rhl/internal/common/d2r"
)

const (
	// configDirName is the directory name under user home.
	configDirName = ".d2r-hyper-launcher"

	// configFileName is the config file name.
	configFileName = "config.json"

	// accountsFileName is the accounts CSV file name.
	accountsFileName = "accounts.csv"

	// HomeDirEnv is the environment variable to override the config directory.
	HomeDirEnv = "D2R_HYPER_LAUNCHER_HOME"

	// MinLaunchDelaySeconds is the hard lower bound for batch launch delays.
	MinLaunchDelaySeconds = 10
)

// LaunchDelayRange describes the delay used between batch launches.
type LaunchDelayRange struct {
	MinSeconds int
	MaxSeconds int
}

// DefaultLaunchDelayRange returns the default fixed delay.
func DefaultLaunchDelayRange() LaunchDelayRange {
	return LaunchDelayRange{MinSeconds: 30, MaxSeconds: 30}
}

// ParseLaunchDelayRange parses either "30" or "30-60" style input.
func ParseLaunchDelayRange(input string) (LaunchDelayRange, error) {
	input = strings.TrimSpace(input)
	if input == "" {
		return LaunchDelayRange{}, fmt.Errorf("啟動間隔不可為空，請輸入像 30 或 30-60")
	}

	if !strings.Contains(input, "-") {
		value, err := strconv.Atoi(input)
		if err != nil {
			return LaunchDelayRange{}, fmt.Errorf("啟動間隔必須是整數，或使用像 30-60 的範圍格式")
		}
		delay := LaunchDelayRange{MinSeconds: value, MaxSeconds: value}
		if err := delay.Validate(); err != nil {
			return LaunchDelayRange{}, err
		}
		return delay, nil
	}

	parts := strings.Split(input, "-")
	if len(parts) != 2 {
		return LaunchDelayRange{}, fmt.Errorf("啟動間隔範圍格式錯誤，請輸入像 30-60")
	}

	minSeconds, err := strconv.Atoi(strings.TrimSpace(parts[0]))
	if err != nil {
		return LaunchDelayRange{}, fmt.Errorf("啟動間隔下限必須是整數")
	}
	maxSeconds, err := strconv.Atoi(strings.TrimSpace(parts[1]))
	if err != nil {
		return LaunchDelayRange{}, fmt.Errorf("啟動間隔上限必須是整數")
	}

	delay := LaunchDelayRange{MinSeconds: minSeconds, MaxSeconds: maxSeconds}
	if err := delay.Validate(); err != nil {
		return LaunchDelayRange{}, err
	}
	return delay, nil
}

// Validate checks whether the delay range is usable.
func (r LaunchDelayRange) Validate() error {
	if r.MinSeconds < MinLaunchDelaySeconds {
		return fmt.Errorf("啟動間隔下限不可小於 %d 秒", MinLaunchDelaySeconds)
	}
	if r.MaxSeconds < MinLaunchDelaySeconds {
		return fmt.Errorf("啟動間隔上限不可小於 %d 秒", MinLaunchDelaySeconds)
	}
	if r.MinSeconds > r.MaxSeconds {
		return fmt.Errorf("啟動間隔下限不可大於上限")
	}
	return nil
}

// String returns the persisted representation.
func (r LaunchDelayRange) String() string {
	if r.MinSeconds == r.MaxSeconds {
		return strconv.Itoa(r.MinSeconds)
	}
	return fmt.Sprintf("%d-%d", r.MinSeconds, r.MaxSeconds)
}

// DisplayString returns a player-facing summary.
func (r LaunchDelayRange) DisplayString() string {
	if r.MinSeconds == r.MaxSeconds {
		return fmt.Sprintf("%d 秒", r.MinSeconds)
	}
	return fmt.Sprintf("%d-%d 秒（隨機）", r.MinSeconds, r.MaxSeconds)
}

// NextSeconds resolves the actual delay used for the next launch.
func (r LaunchDelayRange) NextSeconds(randIntN func(int) int) int {
	if r.MinSeconds == r.MaxSeconds {
		return r.MinSeconds
	}
	return r.MinSeconds + randIntN(r.MaxSeconds-r.MinSeconds+1)
}

// MarshalJSON writes a fixed delay as an integer, and a range as a string.
func (r LaunchDelayRange) MarshalJSON() ([]byte, error) {
	if err := r.Validate(); err != nil {
		return nil, err
	}
	if r.MinSeconds == r.MaxSeconds {
		return json.Marshal(r.MinSeconds)
	}
	return json.Marshal(r.String())
}

// UnmarshalJSON accepts either an integer or a "min-max" string.
func (r *LaunchDelayRange) UnmarshalJSON(data []byte) error {
	var intValue int
	if err := json.Unmarshal(data, &intValue); err == nil {
		parsed := LaunchDelayRange{MinSeconds: intValue, MaxSeconds: intValue}
		if err := parsed.Validate(); err != nil {
			return err
		}
		*r = parsed
		return nil
	}

	var stringValue string
	if err := json.Unmarshal(data, &stringValue); err == nil {
		parsed, err := ParseLaunchDelayRange(stringValue)
		if err != nil {
			return err
		}
		*r = parsed
		return nil
	}

	return fmt.Errorf("launch_delay 必須是整數或像 30-60 的字串")
}

// SwitcherConfig holds the window switcher configuration.
type SwitcherConfig struct {
	// Enabled controls whether the window switcher is active.
	Enabled bool `json:"enabled"`

	// Modifiers is the list of modifier keys (e.g., "ctrl", "alt", "shift").
	Modifiers []string `json:"modifiers,omitempty"`

	// Key is the trigger key name (e.g., "Tab", "F1", "XButton1", "Gamepad_A").
	Key string `json:"key"`

	// GamepadIndex is the XInput controller index (0-3), used when Key is a gamepad button.
	GamepadIndex int `json:"gamepad_index,omitempty"`
}

// Config represents the application configuration.
type Config struct {
	// D2RPath is the path to D2R.exe.
	D2RPath string `json:"d2r_path"`

	// LaunchDelay is the delay range between launching each account.
	LaunchDelay LaunchDelayRange `json:"launch_delay"`

	// Switcher holds the window switcher settings.
	Switcher *SwitcherConfig `json:"switcher,omitempty"`
}

// DefaultConfig returns a Config with default values.
func DefaultConfig() Config {
	return Config{
		D2RPath:     d2r.DefaultGamePath,
		LaunchDelay: DefaultLaunchDelayRange(),
	}
}

// Dir returns the config directory path.
// Priority: D2R_HYPER_LAUNCHER_HOME env > ~/.d2r-hyper-launcher
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

	cfg := DefaultConfig()
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}
	if err := cfg.LaunchDelay.Validate(); err != nil {
		return nil, fmt.Errorf("invalid launch_delay: %w", err)
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
