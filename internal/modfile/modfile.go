// Package modfile provides loading and saving of D2R mod string files.
// It handles the JSON string files located under <mod>.mpq/data/local/lng/strings/.
package modfile

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// StringEntry represents a single entry in a D2R string JSON file.
type StringEntry struct {
	ID   int    `json:"id"`
	Key  string `json:"Key"`
	EnUS string `json:"enUS"`
	ZhTW string `json:"zhTW"`
}

// StringFile represents a loaded JSON string file with its entries.
type StringFile struct {
	// Name is the base filename without extension (e.g., "item-names").
	Name string

	// Path is the full filesystem path to the JSON file.
	Path string

	// Entries is the list of string entries in the file.
	Entries []StringEntry
}

// Mod represents a loaded D2R mod with all its string files.
type Mod struct {
	// Name is the mod directory name.
	Name string

	// Dir is the full filesystem path to the mod directory.
	Dir string

	// Files is the list of loaded string files, keyed by filename.
	Files []*StringFile
}

// stringsRelPath is the relative path from <mod>.mpq to the strings directory.
const stringsRelPath = "data/local/lng/strings"

// Load reads a mod directory and loads all JSON string files.
// modDir should point to the mod root (e.g., mods/d2r-hyper-show).
func Load(modDir string) (*Mod, error) {
	modName := filepath.Base(modDir)
	mpqDir := filepath.Join(modDir, modName+".mpq")
	stringsDir := filepath.Join(mpqDir, stringsRelPath)

	if _, err := os.Stat(stringsDir); err != nil {
		return nil, fmt.Errorf("strings directory not found: %s", stringsDir)
	}

	entries, err := os.ReadDir(stringsDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read strings directory: %w", err)
	}

	mod := &Mod{
		Name: modName,
		Dir:  modDir,
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}

		sf, err := loadStringFile(filepath.Join(stringsDir, entry.Name()))
		if err != nil {
			return nil, fmt.Errorf("failed to load %s: %w", entry.Name(), err)
		}
		mod.Files = append(mod.Files, sf)
	}

	sort.Slice(mod.Files, func(i, j int) bool {
		return mod.Files[i].Name < mod.Files[j].Name
	})

	return mod, nil
}

// utf8BOM is the UTF-8 Byte Order Mark required by D2R string JSON files.
var utf8BOM = []byte{0xEF, 0xBB, 0xBF}

// loadStringFile reads and parses a single JSON string file.
func loadStringFile(path string) (*StringFile, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	// Strip UTF-8 BOM if present
	if len(data) >= 3 && data[0] == 0xEF && data[1] == 0xBB && data[2] == 0xBF {
		data = data[3:]
	}

	var entries []StringEntry
	if err := json.Unmarshal(data, &entries); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	name := strings.TrimSuffix(filepath.Base(path), ".json")
	return &StringFile{
		Name:    name,
		Path:    path,
		Entries: entries,
	}, nil
}

// Save writes a string file back to disk with UTF-8 BOM (required by D2R).
func (sf *StringFile) Save() error {
	data, err := json.MarshalIndent(sf.Entries, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	// Prepend UTF-8 BOM + append newline
	out := make([]byte, 0, len(utf8BOM)+len(data)+1)
	out = append(out, utf8BOM...)
	out = append(out, data...)
	out = append(out, '\n')

	if err := os.WriteFile(sf.Path, out, 0o644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

// AllEntries returns a flat list of all entries across all files, with file reference.
func (m *Mod) AllEntries() []EntryRef {
	var refs []EntryRef
	for _, sf := range m.Files {
		for i := range sf.Entries {
			refs = append(refs, EntryRef{
				File:  sf,
				Index: i,
				Entry: &sf.Entries[i],
			})
		}
	}
	return refs
}

// EntryRef is a reference to a specific entry within a specific file.
type EntryRef struct {
	File  *StringFile
	Index int
	Entry *StringEntry
}

// FileNames returns the list of string file names in the mod.
func (m *Mod) FileNames() []string {
	names := make([]string, len(m.Files))
	for i, f := range m.Files {
		names[i] = f.Name
	}
	return names
}

// FindFile returns the StringFile with the given name, or nil.
func (m *Mod) FindFile(name string) *StringFile {
	for _, f := range m.Files {
		if f.Name == name {
			return f
		}
	}
	return nil
}

// DiscoverMods scans a directory for mod subdirectories (those containing modinfo.json).
func DiscoverMods(modsDir string) ([]string, error) {
	entries, err := os.ReadDir(modsDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read mods directory: %w", err)
	}

	var mods []string
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		modInfoPath := filepath.Join(modsDir, entry.Name(), "modinfo.json")
		if _, err := os.Stat(modInfoPath); err == nil {
			mods = append(mods, entry.Name())
		}
	}

	return mods, nil
}
