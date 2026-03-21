package graphicsprofile

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDefaultSettingsPath(t *testing.T) {
	assert.Equal(
		t,
		filepath.Join(`C:\Users\User`, "Saved Games", "Diablo II Resurrected", "Settings.json"),
		DefaultSettingsPath(`C:\Users\User`),
	)
}

func TestValidateProfileName(t *testing.T) {
	assert.NoError(t, ValidateProfileName("main-high"))
	assert.NoError(t, ValidateProfileName(" boss run "))
	assert.Error(t, ValidateProfileName(""))
	assert.Error(t, ValidateProfileName("bad/name"))
	assert.Error(t, ValidateProfileName("bad*name"))
	assert.Error(t, ValidateProfileName("con"))
	assert.Error(t, ValidateProfileName("."))
}

func TestListReturnsEmptyWhenProfilesDirMissing(t *testing.T) {
	store := NewStore(filepath.Join(t.TempDir(), "launcher-home"), filepath.Join(t.TempDir(), "Settings.json"))

	profiles, err := store.List()
	assert.NoError(t, err)
	assert.Empty(t, profiles)
}

func TestSaveCurrentAsCreatesNamedProfile(t *testing.T) {
	dir := t.TempDir()
	settingsPath := filepath.Join(dir, "saved-games", "Settings.json")
	err := os.MkdirAll(filepath.Dir(settingsPath), 0o700)
	assert.NoError(t, err)
	err = os.WriteFile(settingsPath, []byte(`{"quality":"high"}`), 0o600)
	assert.NoError(t, err)

	store := NewStore(filepath.Join(dir, "launcher-home"), settingsPath)

	err = store.SaveCurrentAs("main-high", false)
	assert.NoError(t, err)

	data, err := os.ReadFile(filepath.Join(store.ProfilesDir(), "main-high.json"))
	assert.NoError(t, err)
	assert.JSONEq(t, `{"quality":"high"}`, string(data))
}

func TestSaveCurrentAsRejectsExistingProfileWithoutOverwrite(t *testing.T) {
	dir := t.TempDir()
	settingsPath := filepath.Join(dir, "saved-games", "Settings.json")
	err := os.MkdirAll(filepath.Dir(settingsPath), 0o700)
	assert.NoError(t, err)
	err = os.WriteFile(settingsPath, []byte(`{"quality":"high"}`), 0o600)
	assert.NoError(t, err)

	store := NewStore(filepath.Join(dir, "launcher-home"), settingsPath)
	err = os.MkdirAll(store.ProfilesDir(), 0o700)
	assert.NoError(t, err)
	err = os.WriteFile(filepath.Join(store.ProfilesDir(), "main-high.json"), []byte(`{"quality":"low"}`), 0o600)
	assert.NoError(t, err)

	err = store.SaveCurrentAs("main-high", false)
	assert.Error(t, err)
	assert.True(t, errors.Is(err, ErrProfileExists))
}

func TestSaveCurrentAsAllowsOverwrite(t *testing.T) {
	dir := t.TempDir()
	settingsPath := filepath.Join(dir, "saved-games", "Settings.json")
	err := os.MkdirAll(filepath.Dir(settingsPath), 0o700)
	assert.NoError(t, err)
	err = os.WriteFile(settingsPath, []byte(`{"quality":"high"}`), 0o600)
	assert.NoError(t, err)

	store := NewStore(filepath.Join(dir, "launcher-home"), settingsPath)
	err = os.MkdirAll(store.ProfilesDir(), 0o700)
	assert.NoError(t, err)
	err = os.WriteFile(filepath.Join(store.ProfilesDir(), "main-high.json"), []byte(`{"quality":"low"}`), 0o600)
	assert.NoError(t, err)

	err = store.SaveCurrentAs("main-high", true)
	assert.NoError(t, err)

	data, err := os.ReadFile(filepath.Join(store.ProfilesDir(), "main-high.json"))
	assert.NoError(t, err)
	assert.JSONEq(t, `{"quality":"high"}`, string(data))
}

func TestSaveCurrentAsRejectsInvalidCurrentSettingsJSON(t *testing.T) {
	dir := t.TempDir()
	settingsPath := filepath.Join(dir, "saved-games", "Settings.json")
	err := os.MkdirAll(filepath.Dir(settingsPath), 0o700)
	assert.NoError(t, err)
	err = os.WriteFile(settingsPath, []byte(`{"quality":`), 0o600)
	assert.NoError(t, err)

	store := NewStore(filepath.Join(dir, "launcher-home"), settingsPath)

	err = store.SaveCurrentAs("main-high", false)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid")
}

func TestListReturnsSortedProfileNames(t *testing.T) {
	dir := t.TempDir()
	store := NewStore(filepath.Join(dir, "launcher-home"), filepath.Join(dir, "saved-games", "Settings.json"))
	err := os.MkdirAll(store.ProfilesDir(), 0o700)
	assert.NoError(t, err)
	err = os.WriteFile(filepath.Join(store.ProfilesDir(), "z-low.json"), []byte(`{}`), 0o600)
	assert.NoError(t, err)
	err = os.WriteFile(filepath.Join(store.ProfilesDir(), "a-high.json"), []byte(`{}`), 0o600)
	assert.NoError(t, err)
	err = os.WriteFile(filepath.Join(store.ProfilesDir(), "README.txt"), []byte("skip"), 0o600)
	assert.NoError(t, err)

	profiles, err := store.List()
	assert.NoError(t, err)
	assert.Equal(t, []string{"a-high", "z-low"}, profiles)
}

func TestApplyCopiesProfileIntoSettingsPath(t *testing.T) {
	dir := t.TempDir()
	store := NewStore(filepath.Join(dir, "launcher-home"), filepath.Join(dir, "Saved Games", "Diablo II Resurrected", "Settings.json"))
	err := os.MkdirAll(store.ProfilesDir(), 0o700)
	assert.NoError(t, err)
	err = os.WriteFile(filepath.Join(store.ProfilesDir(), "alt-low.json"), []byte(`{"quality":"low"}`), 0o600)
	assert.NoError(t, err)

	err = store.Apply("alt-low")
	assert.NoError(t, err)

	data, err := os.ReadFile(store.SettingsPath())
	assert.NoError(t, err)
	assert.JSONEq(t, `{"quality":"low"}`, string(data))
}

func TestApplyReturnsErrProfileNotFound(t *testing.T) {
	store := NewStore(filepath.Join(t.TempDir(), "launcher-home"), filepath.Join(t.TempDir(), "Settings.json"))

	err := store.Apply("missing")
	assert.Error(t, err)
	assert.True(t, errors.Is(err, ErrProfileNotFound))
}

func TestApplyRejectsInvalidProfileJSON(t *testing.T) {
	dir := t.TempDir()
	store := NewStore(filepath.Join(dir, "launcher-home"), filepath.Join(dir, "Saved Games", "Diablo II Resurrected", "Settings.json"))
	err := os.MkdirAll(store.ProfilesDir(), 0o700)
	assert.NoError(t, err)
	err = os.WriteFile(filepath.Join(store.ProfilesDir(), "alt-low.json"), []byte(`{"quality":`), 0o600)
	assert.NoError(t, err)

	err = store.Apply("alt-low")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid")
}

func TestDeleteRemovesSavedProfile(t *testing.T) {
	dir := t.TempDir()
	store := NewStore(filepath.Join(dir, "launcher-home"), filepath.Join(dir, "Saved Games", "Diablo II Resurrected", "Settings.json"))
	err := os.MkdirAll(store.ProfilesDir(), 0o700)
	assert.NoError(t, err)
	err = os.WriteFile(filepath.Join(store.ProfilesDir(), "alt-low.json"), []byte(`{"quality":"low"}`), 0o600)
	assert.NoError(t, err)

	err = store.Delete("alt-low")
	assert.NoError(t, err)

	_, err = os.Stat(filepath.Join(store.ProfilesDir(), "alt-low.json"))
	assert.ErrorIs(t, err, os.ErrNotExist)
}

func TestDeleteReturnsErrProfileNotFound(t *testing.T) {
	store := NewStore(filepath.Join(t.TempDir(), "launcher-home"), filepath.Join(t.TempDir(), "Settings.json"))

	err := store.Delete("missing")
	assert.Error(t, err)
	assert.True(t, errors.Is(err, ErrProfileNotFound))
}
