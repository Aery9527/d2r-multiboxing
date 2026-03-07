package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateSelectedD2RPath(t *testing.T) {
	tmpDir := t.TempDir()
	exePath := filepath.Join(tmpDir, "D2R.exe")

	err := os.WriteFile(exePath, []byte("binary"), 0o600)
	assert.NoError(t, err)

	err = validateSelectedD2RPath(exePath)
	assert.NoError(t, err)
}

func TestValidateSelectedD2RPathRejectsNonD2RExecutable(t *testing.T) {
	tmpDir := t.TempDir()
	exePath := filepath.Join(tmpDir, "not-d2r.exe")

	err := os.WriteFile(exePath, []byte("binary"), 0o600)
	assert.NoError(t, err)

	err = validateSelectedD2RPath(exePath)
	assert.EqualError(t, err, "selected file must be D2R.exe")
}

func TestDialogInitialDirUsesExistingParent(t *testing.T) {
	tmpDir := t.TempDir()
	currentPath := filepath.Join(tmpDir, "D2R.exe")

	assert.Equal(t, tmpDir, dialogInitialDir(currentPath))
}

func TestPowerShellSingleQuote(t *testing.T) {
	assert.Equal(t, "C:\\Games\\D2R", powerShellSingleQuote("C:\\Games\\D2R"))
	assert.Equal(t, "C:\\Bob''s Games\\D2R", powerShellSingleQuote("C:\\Bob's Games\\D2R"))
}
