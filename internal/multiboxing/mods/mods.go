package mods

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
)

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
