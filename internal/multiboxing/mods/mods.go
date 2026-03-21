package mods

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
)

const DefaultModVanilla = "<vanilla>"

// InstalledDir returns the D2R mods directory next to D2R.exe.
func InstalledDir(d2rPath string) string {
	return filepath.Join(filepath.Dir(d2rPath), "mods")
}

// DiscoverInstalled returns installed mod names under the D2R mods directory.
// A mod is considered installed when its directory contains modinfo.json
// or a same-named <mod>.mpq file/directory.
func DiscoverInstalled(d2rPath string) ([]string, error) {
	modsDir := InstalledDir(d2rPath)

	entries, err := os.ReadDir(modsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var discovered []string
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		if isInstalledMod(filepath.Join(modsDir, entry.Name()), entry.Name()) {
			discovered = append(discovered, entry.Name())
		}
	}

	sort.Strings(discovered)
	return discovered, nil
}

func isInstalledMod(modDir string, modName string) bool {
	modInfoPath := filepath.Join(modDir, "modinfo.json")
	if _, err := os.Stat(modInfoPath); err == nil {
		return true
	}

	mpqPath := filepath.Join(modDir, modName+".mpq")
	if _, err := os.Stat(mpqPath); err == nil {
		return true
	}

	return false
}

// BuildLaunchArgs returns the command-line args required to launch a specific mod.
func BuildLaunchArgs(modName string) []string {
	modName = strings.TrimSpace(modName)
	if modName == "" {
		return nil
	}
	return []string{"-mod", modName, "-txt"}
}

// NormalizeSavedDefaultMod trims a saved DefaultMod value and canonicalises the
// special "launch vanilla" sentinel. Empty string means "no default assigned".
func NormalizeSavedDefaultMod(saved string) string {
	saved = strings.TrimSpace(saved)
	if saved == "" {
		return ""
	}
	if IsDefaultModVanilla(saved) {
		return DefaultModVanilla
	}
	return saved
}

// ResolveSavedDefaultMod returns the canonical saved default for the current
// installed-mod list. Empty string means the saved value is either unassigned
// or no longer available for this D2R install.
func ResolveSavedDefaultMod(saved string, installedMods []string) string {
	saved = NormalizeSavedDefaultMod(saved)
	if saved == "" {
		return ""
	}
	if saved == DefaultModVanilla {
		return DefaultModVanilla
	}
	for _, modName := range installedMods {
		if strings.EqualFold(strings.TrimSpace(modName), saved) {
			return modName
		}
	}
	return ""
}

// IsDefaultModVanilla checks whether a saved DefaultMod value means "launch
// the vanilla game without any mod".
func IsDefaultModVanilla(saved string) bool {
	switch strings.ToLower(strings.TrimSpace(saved)) {
	case strings.ToLower(DefaultModVanilla), "vanilla", "none", "no mod", "no-mod", "nomod", "0":
		return true
	default:
		return false
	}
}
