package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"d2rhl/internal/common/config"
	"d2rhl/internal/multiboxing/account"
	"d2rhl/internal/multiboxing/graphicsprofile"

	"github.com/stretchr/testify/assert"
)

func newTestGraphicsProfileStore(t *testing.T) *graphicsprofile.Store {
	t.Helper()

	dir := t.TempDir()
	settingsPath := filepath.Join(dir, "Saved Games", "Diablo II Resurrected", "Settings.json")
	err := os.MkdirAll(filepath.Dir(settingsPath), 0o700)
	assert.NoError(t, err)

	return graphicsprofile.NewStore(filepath.Join(dir, "launcher-home"), settingsPath)
}

func TestGraphicsProfileStatusLabel(t *testing.T) {
	unassigned := graphicsProfileStatusLabel(account.Account{})
	assert.NotEmpty(t, unassigned)
	assert.NotEqual(t, "boss-low", unassigned)
	assert.Equal(t, "boss-low", graphicsProfileStatusLabel(account.Account{GraphicsProfile: "boss-low"}))
}

func TestApplyGraphicsProfileAssignmentsSavesAccounts(t *testing.T) {
	accounts := []account.Account{
		{Email: "alpha@example.com", Password: "pass", DisplayName: "Alpha"},
		{Email: "bravo@example.com", Password: "pass", DisplayName: "Bravo"},
	}
	accountsFile := filepath.Join(t.TempDir(), "accounts.csv")

	err := applyGraphicsProfileAssignments(accounts, accountsFile, []int{0, 1}, "boss-low")
	assert.NoError(t, err)
	assert.Equal(t, "boss-low", accounts[0].GraphicsProfile)
	assert.Equal(t, "boss-low", accounts[1].GraphicsProfile)

	reloaded, err := account.LoadAccounts(accountsFile)
	assert.NoError(t, err)
	assert.Equal(t, "boss-low", reloaded[0].GraphicsProfile)
	assert.Equal(t, "boss-low", reloaded[1].GraphicsProfile)
}

func TestClearGraphicsProfileAssignmentsClearsSelectedAccounts(t *testing.T) {
	accounts := []account.Account{
		{Email: "alpha@example.com", Password: "pass", DisplayName: "Alpha", GraphicsProfile: "boss-low"},
		{Email: "bravo@example.com", Password: "pass", DisplayName: "Bravo", GraphicsProfile: "boss-high"},
	}
	accountsFile := filepath.Join(t.TempDir(), "accounts.csv")
	err := account.SaveAccounts(accountsFile, accounts)
	assert.NoError(t, err)

	err = clearGraphicsProfileAssignments(accounts, accountsFile, []int{1})
	assert.NoError(t, err)
	assert.Equal(t, "boss-low", accounts[0].GraphicsProfile)
	assert.Equal(t, "", accounts[1].GraphicsProfile)

	reloaded, err := account.LoadAccounts(accountsFile)
	assert.NoError(t, err)
	assert.Equal(t, "boss-low", reloaded[0].GraphicsProfile)
	assert.Equal(t, "", reloaded[1].GraphicsProfile)
}

func TestApplyGraphicsProfileForLaunchSkipsUnassignedAccount(t *testing.T) {
	store, err := applyGraphicsProfileForLaunch(account.Account{DisplayName: "Alpha"}, nil)
	assert.NoError(t, err)
	assert.Nil(t, store)
}

func TestApplyGraphicsProfileForLaunchAppliesAssignedProfile(t *testing.T) {
	originalFactory := newGraphicsProfileStore
	t.Cleanup(func() {
		newGraphicsProfileStore = originalFactory
	})

	dir := t.TempDir()
	store := graphicsprofile.NewStore(
		filepath.Join(dir, "launcher-home"),
		filepath.Join(dir, "Saved Games", "Diablo II Resurrected", "Settings.json"),
	)
	err := os.MkdirAll(store.ProfilesDir(), 0o700)
	assert.NoError(t, err)
	err = os.WriteFile(filepath.Join(store.ProfilesDir(), "boss-low.json"), []byte(`{"quality":"low"}`), 0o600)
	assert.NoError(t, err)

	newGraphicsProfileStore = func() (*graphicsprofile.Store, error) {
		return store, nil
	}

	returnedStore, err := applyGraphicsProfileForLaunch(account.Account{DisplayName: "Alpha", GraphicsProfile: "boss-low"}, nil)
	assert.NoError(t, err)
	assert.Same(t, store, returnedStore)

	data, err := os.ReadFile(store.SettingsPath())
	assert.NoError(t, err)
	assert.JSONEq(t, `{"quality":"low"}`, string(data))
}

func TestApplyNamedGraphicsProfileForLaunchPreservesNonMissingProfileErrors(t *testing.T) {
	store := newTestGraphicsProfileStore(t)
	err := os.WriteFile(store.SettingsPath(), []byte(`{"quality":"existing"}`), 0o600)
	assert.NoError(t, err)
	err = os.MkdirAll(store.ProfilesDir(), 0o700)
	assert.NoError(t, err)
	err = os.WriteFile(filepath.Join(store.ProfilesDir(), "broken.json"), []byte(`{"quality":`), 0o600)
	assert.NoError(t, err)

	returnedStore, err := applyNamedGraphicsProfileForLaunch("broken", store)
	assert.Error(t, err)
	assert.Same(t, store, returnedStore)
}

func TestPrintMenuShowsGraphicsProfilesOption(t *testing.T) {
	cfg := &config.Config{
		D2RPath:     `C:\Games\D2R\D2R.exe`,
		LaunchDelay: config.LaunchDelayRange{MinSeconds: 10, MaxSeconds: 10},
	}

	output := captureStdout(t, func() {
		printMenu(nil, cfg)
	})

	assert.Equal(t, 1, countMenuBlocksWithKeys(output, []string{"數字", "0", "a", "d", "f", "g", "m", "v", "p", "s", "r", "l", "q"}))
	graphicsOptionLine, ok := findMenuOptionLine(output, "g")
	assert.True(t, ok)
	assert.NotEmpty(t, graphicsOptionLine)
}

func TestSetupAccountGraphicsProfilesShowsUsageGuidanceAboveAccountList(t *testing.T) {
	originalFactory := newGraphicsProfileStore
	t.Cleanup(func() {
		newGraphicsProfileStore = originalFactory
	})

	store := newTestGraphicsProfileStore(t)
	err := os.MkdirAll(store.ProfilesDir(), 0o700)
	assert.NoError(t, err)

	newGraphicsProfileStore = func() (*graphicsprofile.Store, error) {
		return store, nil
	}

	accounts := []account.Account{
		{Email: "alpha@example.com", Password: "pass", DisplayName: "Alpha", GraphicsProfile: "boss-low"},
	}

	output := captureStdout(t, func() {
		withTestInput(t, "b\n", func() {
			setupAccountGraphicsProfiles(accounts, filepath.Join(t.TempDir(), "accounts.csv"))
		})
	})

	lines := nonEmptyOutputLines(output)
	accountListIndex := firstLineIndex(lines, func(line string) bool {
		return strings.HasPrefix(line, "[1] <") && strings.Contains(line, "alpha@example.com")
	})
	assert.NotEqual(t, -1, accountListIndex)

	infoLinesBeforeAccount := 0
	for _, line := range lines[:accountListIndex] {
		if strings.HasPrefix(line, ui.prefix(uiMessageInfo)+" ") {
			infoLinesBeforeAccount++
		}
	}
	assert.Equal(t, 5, infoLinesBeforeAccount)
	assert.Equal(t, 1, countMenuBlocksWithKeys(output, []string{"1", "2", "3", "4", "b", "h", "q"}))
}

func TestSaveCurrentGraphicsProfileShowsExistingProfilesAsOverwriteOptions(t *testing.T) {
	originalFactory := newGraphicsProfileStore
	t.Cleanup(func() {
		newGraphicsProfileStore = originalFactory
	})

	store := newTestGraphicsProfileStore(t)
	err := os.MkdirAll(store.ProfilesDir(), 0o700)
	assert.NoError(t, err)
	err = os.WriteFile(filepath.Join(store.ProfilesDir(), "boss-low.json"), []byte(`{"quality":"low"}`), 0o600)
	assert.NoError(t, err)

	newGraphicsProfileStore = func() (*graphicsprofile.Store, error) {
		return store, nil
	}

	output := captureStdout(t, func() {
		withTestInput(t, "b\n", func() {
			err := saveCurrentGraphicsProfile()
			assert.NoError(t, err)
		})
	})

	assert.Equal(t, 1, countMenuBlocksWithKeys(output, []string{"1", "b", "h", "q"}))
	overwriteOptionLine, ok := findMenuOptionLine(output, "1")
	assert.True(t, ok)
	assert.Contains(t, overwriteOptionLine, "boss-low")
}

func TestSaveCurrentGraphicsProfileOverwritesByNumber(t *testing.T) {
	originalFactory := newGraphicsProfileStore
	t.Cleanup(func() {
		newGraphicsProfileStore = originalFactory
	})

	store := newTestGraphicsProfileStore(t)
	err := os.WriteFile(store.SettingsPath(), []byte(`{"quality":"high"}`), 0o600)
	assert.NoError(t, err)
	err = os.MkdirAll(store.ProfilesDir(), 0o700)
	assert.NoError(t, err)
	err = os.WriteFile(filepath.Join(store.ProfilesDir(), "boss-low.json"), []byte(`{"quality":"low"}`), 0o600)
	assert.NoError(t, err)

	newGraphicsProfileStore = func() (*graphicsprofile.Store, error) {
		return store, nil
	}

	_ = captureStdout(t, func() {
		withTestInput(t, "1\n", func() {
			err := saveCurrentGraphicsProfile()
			assert.NoError(t, err)
		})
	})

	data, err := os.ReadFile(filepath.Join(store.ProfilesDir(), "boss-low.json"))
	assert.NoError(t, err)
	assert.JSONEq(t, `{"quality":"high"}`, string(data))
}

func TestSaveCurrentGraphicsProfileCreatesNewProfileFromTypedName(t *testing.T) {
	originalFactory := newGraphicsProfileStore
	t.Cleanup(func() {
		newGraphicsProfileStore = originalFactory
	})

	store := newTestGraphicsProfileStore(t)
	err := os.WriteFile(store.SettingsPath(), []byte(`{"quality":"high"}`), 0o600)
	assert.NoError(t, err)
	err = os.MkdirAll(store.ProfilesDir(), 0o700)
	assert.NoError(t, err)
	err = os.WriteFile(filepath.Join(store.ProfilesDir(), "boss-low.json"), []byte(`{"quality":"low"}`), 0o600)
	assert.NoError(t, err)

	newGraphicsProfileStore = func() (*graphicsprofile.Store, error) {
		return store, nil
	}

	_ = captureStdout(t, func() {
		withTestInput(t, "fresh-high\n", func() {
			err := saveCurrentGraphicsProfile()
			assert.NoError(t, err)
		})
	})

	data, err := os.ReadFile(filepath.Join(store.ProfilesDir(), "fresh-high.json"))
	assert.NoError(t, err)
	assert.JSONEq(t, `{"quality":"high"}`, string(data))
}

func TestDeleteSavedGraphicsProfilesRemovesSelectedProfiles(t *testing.T) {
	originalFactory := newGraphicsProfileStore
	t.Cleanup(func() {
		newGraphicsProfileStore = originalFactory
	})

	store := newTestGraphicsProfileStore(t)
	err := os.MkdirAll(store.ProfilesDir(), 0o700)
	assert.NoError(t, err)
	err = os.WriteFile(filepath.Join(store.ProfilesDir(), "boss-low.json"), []byte(`{"quality":"low"}`), 0o600)
	assert.NoError(t, err)
	err = os.WriteFile(filepath.Join(store.ProfilesDir(), "boss-high.json"), []byte(`{"quality":"high"}`), 0o600)
	assert.NoError(t, err)

	newGraphicsProfileStore = func() (*graphicsprofile.Store, error) {
		return store, nil
	}

	_ = captureStdout(t, func() {
		withTestInput(t, "1,2\ny\n", func() {
			err := deleteSavedGraphicsProfiles(nil)
			assert.NoError(t, err)
		})
	})

	_, err = os.Stat(filepath.Join(store.ProfilesDir(), "boss-low.json"))
	assert.ErrorIs(t, err, os.ErrNotExist)
	_, err = os.Stat(filepath.Join(store.ProfilesDir(), "boss-high.json"))
	assert.ErrorIs(t, err, os.ErrNotExist)
}

func TestDeleteSavedGraphicsProfilesBlocksAssignedProfiles(t *testing.T) {
	originalFactory := newGraphicsProfileStore
	t.Cleanup(func() {
		newGraphicsProfileStore = originalFactory
	})

	store := newTestGraphicsProfileStore(t)
	err := os.MkdirAll(store.ProfilesDir(), 0o700)
	assert.NoError(t, err)
	err = os.WriteFile(filepath.Join(store.ProfilesDir(), "boss-low.json"), []byte(`{"quality":"low"}`), 0o600)
	assert.NoError(t, err)

	newGraphicsProfileStore = func() (*graphicsprofile.Store, error) {
		return store, nil
	}

	accounts := []account.Account{
		{Email: "alpha@example.com", Password: "pass", DisplayName: "Alpha", GraphicsProfile: "boss-low"},
	}

	_ = captureStdout(t, func() {
		withTestInput(t, "1\n\nb\n", func() {
			err := deleteSavedGraphicsProfiles(accounts)
			assert.NoError(t, err)
		})
	})

	_, err = os.Stat(filepath.Join(store.ProfilesDir(), "boss-low.json"))
	assert.NoError(t, err)
	assert.Equal(t, "boss-low", accounts[0].GraphicsProfile)
}
