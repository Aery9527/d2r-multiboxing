package graphicsprofile

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"d2rhl/internal/common/config"
)

const (
	profilesDirName  = "graphics-profiles"
	settingsDirName  = "Diablo II Resurrected"
	settingsFileName = "Settings.json"
)

const invalidProfileNameChars = `<>:"/\|?*`

var (
	ErrProfileExists   = errors.New("graphics profile already exists")
	ErrProfileNotFound = errors.New("graphics profile not found")
)

var reservedWindowsNames = map[string]struct{}{
	"CON":  {},
	"PRN":  {},
	"AUX":  {},
	"NUL":  {},
	"COM1": {},
	"COM2": {},
	"COM3": {},
	"COM4": {},
	"COM5": {},
	"COM6": {},
	"COM7": {},
	"COM8": {},
	"COM9": {},
	"LPT1": {},
	"LPT2": {},
	"LPT3": {},
	"LPT4": {},
	"LPT5": {},
	"LPT6": {},
	"LPT7": {},
	"LPT8": {},
	"LPT9": {},
}

type Store struct {
	profilesDir  string
	settingsPath string
}

func NewDefaultStore() (*Store, error) {
	launcherHomeDir, err := config.Dir()
	if err != nil {
		return nil, err
	}

	userHomeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get Windows user home directory: %w", err)
	}

	return NewStore(launcherHomeDir, DefaultSettingsPath(userHomeDir)), nil
}

func NewStore(launcherHomeDir string, settingsPath string) *Store {
	return &Store{
		profilesDir:  filepath.Join(launcherHomeDir, profilesDirName),
		settingsPath: settingsPath,
	}
}

func DefaultSettingsPath(userHomeDir string) string {
	return filepath.Join(userHomeDir, "Saved Games", settingsDirName, settingsFileName)
}

func (s *Store) ProfilesDir() string {
	return s.profilesDir
}

func (s *Store) SettingsPath() string {
	return s.settingsPath
}

func (s *Store) List() ([]string, error) {
	entries, err := os.ReadDir(s.profilesDir)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to read graphics profile directory: %w", err)
	}

	profiles := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if !strings.EqualFold(filepath.Ext(entry.Name()), ".json") {
			continue
		}
		profiles = append(profiles, strings.TrimSuffix(entry.Name(), filepath.Ext(entry.Name())))
	}

	sort.Strings(profiles)
	return profiles, nil
}

func (s *Store) Exists(name string) (bool, error) {
	profilePath, err := s.ProfilePath(name)
	if err != nil {
		return false, err
	}

	_, err = os.Stat(profilePath)
	if err == nil {
		return true, nil
	}
	if errors.Is(err, fs.ErrNotExist) {
		return false, nil
	}
	return false, fmt.Errorf("failed to stat graphics profile: %w", err)
}

func (s *Store) SaveCurrentAs(name string, overwrite bool) error {
	profilePath, err := s.ProfilePath(name)
	if err != nil {
		return err
	}

	data, err := os.ReadFile(s.settingsPath)
	if err != nil {
		return fmt.Errorf("failed to read current Settings.json: %w", err)
	}
	if err := validateSettingsJSON(data); err != nil {
		return fmt.Errorf("current Settings.json is invalid: %w", err)
	}

	if err := os.MkdirAll(s.profilesDir, 0o700); err != nil {
		return fmt.Errorf("failed to create graphics profile directory: %w", err)
	}

	if !overwrite {
		exists, err := s.Exists(name)
		if err != nil {
			return err
		}
		if exists {
			return fmt.Errorf("%w: %s", ErrProfileExists, normalizeProfileName(name))
		}
	}

	if err := os.WriteFile(profilePath, data, 0o600); err != nil {
		return fmt.Errorf("failed to save graphics profile: %w", err)
	}
	return nil
}

func (s *Store) Apply(name string) error {
	profilePath, err := s.ProfilePath(name)
	if err != nil {
		return err
	}

	data, err := os.ReadFile(profilePath)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return fmt.Errorf("%w: %s", ErrProfileNotFound, normalizeProfileName(name))
		}
		return fmt.Errorf("failed to read graphics profile: %w", err)
	}
	if err := validateSettingsJSON(data); err != nil {
		return fmt.Errorf("graphics profile %q is invalid: %w", normalizeProfileName(name), err)
	}

	if err := os.MkdirAll(filepath.Dir(s.settingsPath), 0o700); err != nil {
		return fmt.Errorf("failed to create D2R settings directory: %w", err)
	}
	if err := os.WriteFile(s.settingsPath, data, 0o600); err != nil {
		return fmt.Errorf("failed to apply graphics profile: %w", err)
	}
	return nil
}

func (s *Store) Delete(name string) error {
	profilePath, err := s.ProfilePath(name)
	if err != nil {
		return err
	}

	if err := os.Remove(profilePath); err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return fmt.Errorf("%w: %s", ErrProfileNotFound, normalizeProfileName(name))
		}
		return fmt.Errorf("failed to delete graphics profile: %w", err)
	}
	return nil
}

func (s *Store) ProfilePath(name string) (string, error) {
	normalized := normalizeProfileName(name)
	if err := ValidateProfileName(normalized); err != nil {
		return "", err
	}
	return filepath.Join(s.profilesDir, normalized+".json"), nil
}

func ValidateProfileName(name string) error {
	normalized := normalizeProfileName(name)
	if normalized == "" {
		return errors.New("graphics profile name cannot be empty")
	}
	if normalized == "." || normalized == ".." {
		return errors.New("graphics profile name cannot be . or ..")
	}
	if strings.ContainsAny(normalized, invalidProfileNameChars) {
		return fmt.Errorf("graphics profile name cannot contain any of %q", invalidProfileNameChars)
	}
	if strings.HasSuffix(normalized, ".") {
		return errors.New("graphics profile name cannot end with a dot")
	}
	if strings.HasSuffix(normalized, " ") {
		return errors.New("graphics profile name cannot end with a space")
	}

	reservedCheckName := strings.ToUpper(strings.SplitN(normalized, ".", 2)[0])
	if _, ok := reservedWindowsNames[reservedCheckName]; ok {
		return fmt.Errorf("graphics profile name cannot use reserved Windows device name %q", reservedCheckName)
	}
	return nil
}

func normalizeProfileName(name string) string {
	return strings.TrimSpace(name)
}

func validateSettingsJSON(data []byte) error {
	if !json.Valid(data) {
		return errors.New("content is not valid JSON")
	}
	return nil
}
