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
	assert.Equal(t, lang.GraphicsProfiles.StatusUnassigned, graphicsProfileStatusLabel(account.Account{}))
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

func TestPrepareGraphicsProfileForLaunchClearsMissingAssignmentAndLeavesSettingsUntouched(t *testing.T) {
	originalFactory := newGraphicsProfileStore
	t.Cleanup(func() {
		newGraphicsProfileStore = originalFactory
	})

	store := newTestGraphicsProfileStore(t)
	err := os.WriteFile(store.SettingsPath(), []byte(`{"quality":"existing"}`), 0o600)
	assert.NoError(t, err)

	accounts := []account.Account{
		{Email: "alpha@example.com", Password: "pass", DisplayName: "Alpha", GraphicsProfile: "missing"},
	}
	accountsFile := filepath.Join(t.TempDir(), "accounts.csv")
	err = account.SaveAccounts(accountsFile, accounts)
	assert.NoError(t, err)

	newGraphicsProfileStore = func() (*graphicsprofile.Store, error) {
		return store, nil
	}

	output := captureStdout(t, func() {
		returnedStore, err := prepareGraphicsProfileForLaunch(accounts, accountsFile, &accounts[0], nil)
		assert.NoError(t, err)
		assert.Same(t, store, returnedStore)
	})

	data, err := os.ReadFile(store.SettingsPath())
	assert.NoError(t, err)
	assert.JSONEq(t, `{"quality":"existing"}`, string(data))
	assert.Equal(t, "", accounts[0].GraphicsProfile)

	reloaded, err := account.LoadAccounts(accountsFile)
	assert.NoError(t, err)
	assert.Equal(t, "", reloaded[0].GraphicsProfile)
	assert.Contains(t, output, `帳號 Alpha (alpha@example.com) 指派的畫質設定檔「missing」不存在；已自動清空該帳號的畫質設定，這次啟動不會改動 Settings.json。`)
}

func TestPrepareGraphicsProfileForLaunchPreservesNonMissingProfileErrors(t *testing.T) {
	originalFactory := newGraphicsProfileStore
	t.Cleanup(func() {
		newGraphicsProfileStore = originalFactory
	})

	store := newTestGraphicsProfileStore(t)
	err := os.WriteFile(store.SettingsPath(), []byte(`{"quality":"existing"}`), 0o600)
	assert.NoError(t, err)
	err = os.MkdirAll(store.ProfilesDir(), 0o700)
	assert.NoError(t, err)
	err = os.WriteFile(filepath.Join(store.ProfilesDir(), "broken.json"), []byte(`{"quality":`), 0o600)
	assert.NoError(t, err)

	accounts := []account.Account{
		{Email: "alpha@example.com", Password: "pass", DisplayName: "Alpha", GraphicsProfile: "broken"},
	}
	accountsFile := filepath.Join(t.TempDir(), "accounts.csv")
	err = account.SaveAccounts(accountsFile, accounts)
	assert.NoError(t, err)

	newGraphicsProfileStore = func() (*graphicsprofile.Store, error) {
		return store, nil
	}

	returnedStore, err := prepareGraphicsProfileForLaunch(accounts, accountsFile, &accounts[0], nil)
	assert.Error(t, err)
	assert.Same(t, store, returnedStore)
	assert.Contains(t, err.Error(), "invalid")
	assert.Equal(t, "broken", accounts[0].GraphicsProfile)

	reloaded, loadErr := account.LoadAccounts(accountsFile)
	assert.NoError(t, loadErr)
	assert.Equal(t, "broken", reloaded[0].GraphicsProfile)
}

func TestPrintMenuShowsGraphicsProfilesOption(t *testing.T) {
	cfg := &config.Config{
		D2RPath:     `C:\Games\D2R\D2R.exe`,
		LaunchDelay: config.LaunchDelayRange{MinSeconds: 10, MaxSeconds: 10},
	}

	output := captureStdout(t, func() {
		printMenu(nil, cfg)
	})

	assert.Contains(t, output, "[g]    帳號畫質設定檔")
	assert.Contains(t, output, "儲存目前 Settings.json 並指派給帳號")
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

	assert.Contains(t, output, "• "+lang.GraphicsProfiles.Intro1)
	assert.Contains(t, output, "• "+lang.GraphicsProfiles.Intro2)
	assert.Contains(t, output, "• "+lang.GraphicsProfiles.Intro3)
	assert.Contains(t, output, "• "+lang.GraphicsProfiles.Intro4)
	assert.Contains(t, output, lang.GraphicsProfiles.OptDeleteSaved)

	introIndex := strings.Index(output, lang.GraphicsProfiles.Intro1)
	accountListIndex := strings.Index(output, lang.MainMenu.AccountListHeader)
	assert.NotEqual(t, -1, introIndex)
	assert.NotEqual(t, -1, accountListIndex)
	assert.Less(t, introIndex, accountListIndex)
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

	assert.Contains(t, output, "已保存的畫質設定檔：")
	assert.Contains(t, output, "[1] boss-low")
	assert.Contains(t, output, "覆蓋既有設定")
	assert.Contains(t, output, "請輸入設定檔編號以覆蓋既有設定，或輸入新名稱另存：")
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

	output := captureStdout(t, func() {
		withTestInput(t, "1\n", func() {
			err := saveCurrentGraphicsProfile()
			assert.NoError(t, err)
		})
	})

	data, err := os.ReadFile(filepath.Join(store.ProfilesDir(), "boss-low.json"))
	assert.NoError(t, err)
	assert.JSONEq(t, `{"quality":"high"}`, string(data))
	assert.Contains(t, output, "已儲存畫質設定檔：boss-low")
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

	output := captureStdout(t, func() {
		withTestInput(t, "fresh-high\n", func() {
			err := saveCurrentGraphicsProfile()
			assert.NoError(t, err)
		})
	})

	data, err := os.ReadFile(filepath.Join(store.ProfilesDir(), "fresh-high.json"))
	assert.NoError(t, err)
	assert.JSONEq(t, `{"quality":"high"}`, string(data))
	assert.Contains(t, output, "已儲存畫質設定檔：fresh-high")
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

	output := captureStdout(t, func() {
		withTestInput(t, "1,2\ny\n", func() {
			err := deleteSavedGraphicsProfiles(nil)
			assert.NoError(t, err)
		})
	})

	_, err = os.Stat(filepath.Join(store.ProfilesDir(), "boss-low.json"))
	assert.ErrorIs(t, err, os.ErrNotExist)
	_, err = os.Stat(filepath.Join(store.ProfilesDir(), "boss-high.json"))
	assert.ErrorIs(t, err, os.ErrNotExist)
	assert.Contains(t, output, lang.GraphicsProfiles.DeleteDone)
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

	output := captureStdout(t, func() {
		withTestInput(t, "1\n\nb\n", func() {
			err := deleteSavedGraphicsProfiles(accounts)
			assert.NoError(t, err)
		})
	})

	_, err = os.Stat(filepath.Join(store.ProfilesDir(), "boss-low.json"))
	assert.NoError(t, err)
	assert.Contains(t, output, `畫質設定檔「boss-low」仍被以下帳號使用：Alpha (alpha@example.com)。請先清除或改派後再刪除。`)
}
