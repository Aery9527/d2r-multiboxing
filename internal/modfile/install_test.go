package modfile

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestD2RModsDir(t *testing.T) {
	got := D2RModsDir(`C:\Program Files (x86)\Diablo II Resurrected\D2R.exe`)
	assert.Equal(t, `C:\Program Files (x86)\Diablo II Resurrected\mods`, got)
}

func TestInstallMod(t *testing.T) {
	// Create a fake source mod
	srcDir := t.TempDir()
	modDir := filepath.Join(srcDir, "test-mod")
	mpqDir := filepath.Join(modDir, "test-mod.mpq", "data", "local", "lng", "strings")
	assert.NoError(t, os.MkdirAll(mpqDir, 0o755))
	assert.NoError(t, os.WriteFile(filepath.Join(modDir, "modinfo.json"), []byte(`{"name":"test-mod"}`), 0o644))
	assert.NoError(t, os.WriteFile(filepath.Join(mpqDir, "item-names.json"), []byte(`[]`), 0o644))

	// Create a fake D2R directory
	d2rDir := t.TempDir()
	d2rExe := filepath.Join(d2rDir, "D2R.exe")
	assert.NoError(t, os.WriteFile(d2rExe, []byte("fake"), 0o644))

	// Install
	err := InstallMod(modDir, d2rExe)
	assert.NoError(t, err)

	// Verify files were copied
	installedModInfo := filepath.Join(d2rDir, "mods", "test-mod", "modinfo.json")
	assert.FileExists(t, installedModInfo)

	installedJSON := filepath.Join(d2rDir, "mods", "test-mod", "test-mod.mpq", "data", "local", "lng", "strings", "item-names.json")
	assert.FileExists(t, installedJSON)
}

func TestDiscoverInstalledMods(t *testing.T) {
	// Create a fake D2R directory with an installed mod
	d2rDir := t.TempDir()
	d2rExe := filepath.Join(d2rDir, "D2R.exe")
	assert.NoError(t, os.WriteFile(d2rExe, []byte("fake"), 0o644))

	modDir := filepath.Join(d2rDir, "mods", "my-mod")
	assert.NoError(t, os.MkdirAll(modDir, 0o755))
	assert.NoError(t, os.WriteFile(filepath.Join(modDir, "modinfo.json"), []byte(`{}`), 0o644))

	mods, err := DiscoverInstalledMods(d2rExe)
	assert.NoError(t, err)
	assert.Contains(t, mods, "my-mod")
}

func TestDiscoverInstalledMods_NoModsDir(t *testing.T) {
	d2rDir := t.TempDir()
	d2rExe := filepath.Join(d2rDir, "D2R.exe")

	mods, err := DiscoverInstalledMods(d2rExe)
	assert.Error(t, err)
	assert.Nil(t, mods)
}
