package mods

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInstalledDir(t *testing.T) {
	d2rPath := `D:\Games\Diablo II Resurrected\D2R.exe`
	assert.Equal(t, filepath.Join(`D:\Games\Diablo II Resurrected`, "mods"), InstalledDir(d2rPath))
}

func TestDiscoverInstalled(t *testing.T) {
	root := t.TempDir()
	d2rPath := filepath.Join(root, "D2R.exe")
	modsDir := InstalledDir(d2rPath)

	assert.NoError(t, os.MkdirAll(filepath.Join(modsDir, "z-last"), 0o755))
	assert.NoError(t, os.WriteFile(filepath.Join(modsDir, "z-last", "modinfo.json"), []byte("{}"), 0o644))
	assert.NoError(t, os.MkdirAll(filepath.Join(modsDir, "a-first"), 0o755))
	assert.NoError(t, os.WriteFile(filepath.Join(modsDir, "a-first", "modinfo.json"), []byte("{}"), 0o644))
	assert.NoError(t, os.MkdirAll(filepath.Join(modsDir, "ignored"), 0o755))
	assert.NoError(t, os.WriteFile(filepath.Join(modsDir, "plain-file.txt"), []byte("x"), 0o644))

	discovered, err := DiscoverInstalled(d2rPath)
	assert.NoError(t, err)
	assert.Equal(t, []string{"a-first", "z-last"}, discovered)
}

func TestDiscoverInstalledWithMatchingMPQ(t *testing.T) {
	root := t.TempDir()
	d2rPath := filepath.Join(root, "D2R.exe")
	modsDir := InstalledDir(d2rPath)

	assert.NoError(t, os.MkdirAll(filepath.Join(modsDir, "MCMod"), 0o755))
	assert.NoError(t, os.WriteFile(filepath.Join(modsDir, "MCMod", "MCMod.mpq"), []byte("x"), 0o644))

	discovered, err := DiscoverInstalled(d2rPath)
	assert.NoError(t, err)
	assert.Equal(t, []string{"MCMod"}, discovered)
}

func TestDiscoverInstalledMissingDir(t *testing.T) {
	root := t.TempDir()
	d2rPath := filepath.Join(root, "D2R.exe")

	discovered, err := DiscoverInstalled(d2rPath)
	assert.NoError(t, err)
	assert.Empty(t, discovered)
}

func TestBuildLaunchArgs(t *testing.T) {
	assert.Nil(t, BuildLaunchArgs(""))
	assert.Nil(t, BuildLaunchArgs("   "))
	assert.Equal(t, []string{"-mod", "sample-mod", "-txt"}, BuildLaunchArgs("sample-mod"))
}
