package modfile

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoad(t *testing.T) {
	modDir := filepath.Join("..", "..", "mods", "d2r-hyper-show")

	mod, err := Load(modDir)
	assert.NoError(t, err)
	assert.Equal(t, "d2r-hyper-show", mod.Name)
	assert.Len(t, mod.Files, 3)

	names := mod.FileNames()
	assert.Contains(t, names, "item-names")
	assert.Contains(t, names, "skills")
	assert.Contains(t, names, "ui")
}

func TestLoad_NotFound(t *testing.T) {
	_, err := Load("/nonexistent/mod")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "strings directory not found")
}

func TestFindFile(t *testing.T) {
	modDir := filepath.Join("..", "..", "mods", "d2r-hyper-show")
	mod, err := Load(modDir)
	assert.NoError(t, err)

	sf := mod.FindFile("item-names")
	assert.NotNil(t, sf)
	assert.Greater(t, len(sf.Entries), 0)

	// First entry should be El Rune
	assert.Equal(t, "r01", sf.Entries[0].Key)

	sf = mod.FindFile("nonexistent")
	assert.Nil(t, sf)
}

func TestAllEntries(t *testing.T) {
	modDir := filepath.Join("..", "..", "mods", "d2r-hyper-show")
	mod, err := Load(modDir)
	assert.NoError(t, err)

	refs := mod.AllEntries()
	assert.Greater(t, len(refs), 0)

	// Each ref should have valid pointers
	for _, ref := range refs {
		assert.NotNil(t, ref.File)
		assert.NotNil(t, ref.Entry)
		assert.GreaterOrEqual(t, ref.Index, 0)
	}
}

func TestStringSave(t *testing.T) {
	// Create a temp file with test data
	tmpDir := t.TempDir()
	testPath := filepath.Join(tmpDir, "test.json")

	sf := &StringFile{
		Name: "test",
		Path: testPath,
		Entries: []StringEntry{
			{ID: 1, Key: "test1", EnUS: "Hello", ZhTW: "你好"},
			{ID: 2, Key: "test2", EnUS: "World", ZhTW: "世界"},
		},
	}

	err := sf.Save()
	assert.NoError(t, err)

	// Reload and verify
	data, err := os.ReadFile(testPath)
	assert.NoError(t, err)
	assert.Contains(t, string(data), `"Key": "test1"`)
	assert.Contains(t, string(data), `"enUS": "Hello"`)

	// Verify round-trip
	sf2, err := loadStringFile(testPath)
	assert.NoError(t, err)
	assert.Len(t, sf2.Entries, 2)
	assert.Equal(t, "test1", sf2.Entries[0].Key)
	assert.Equal(t, "你好", sf2.Entries[0].ZhTW)
}

func TestDiscoverMods(t *testing.T) {
	modsDir := filepath.Join("..", "..", "mods")
	mods, err := DiscoverMods(modsDir)
	assert.NoError(t, err)
	assert.Contains(t, mods, "d2r-hyper-show")
}
