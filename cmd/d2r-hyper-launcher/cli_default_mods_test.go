package main

import (
	"path/filepath"
	"testing"

	"d2rhl/internal/multiboxing/account"
	"d2rhl/internal/multiboxing/mods"

	"github.com/stretchr/testify/assert"
)

func TestDefaultModStatusLabel(t *testing.T) {
	assert.Equal(t, lang.DefaultMods.StatusUnassigned, defaultModStatusLabel(account.Account{}, nil))
	assert.Equal(t, lang.DefaultMods.StatusVanilla, defaultModStatusLabel(account.Account{DefaultMod: mods.DefaultModVanilla}, nil))
	assert.Equal(t, "SampleMod", defaultModStatusLabel(account.Account{DefaultMod: "samplemod"}, []string{"SampleMod"}))
	assert.Contains(t, defaultModStatusLabel(account.Account{DefaultMod: "missing-mod"}, []string{"SampleMod"}), "missing-mod")
}

func TestAssignDefaultModsByAccountPersistsVanillaSelection(t *testing.T) {
	accounts := []account.Account{
		{Email: "alpha@example.com", Password: "pass", DisplayName: "Alpha"},
	}
	accountsFile := filepath.Join(t.TempDir(), "accounts.csv")
	err := account.SaveAccounts(accountsFile, accounts)
	assert.NoError(t, err)

	withTestInput(t, "1\n0\n\n", func() {
		err = assignDefaultModsByAccount(accounts, accountsFile, nil)
	})

	assert.ErrorIs(t, err, errNavDone)
	assert.Equal(t, mods.DefaultModVanilla, accounts[0].DefaultMod)

	reloaded, loadErr := account.LoadAccounts(accountsFile)
	assert.NoError(t, loadErr)
	assert.Equal(t, mods.DefaultModVanilla, reloaded[0].DefaultMod)
}

func TestClearDefaultModsPersistsSelection(t *testing.T) {
	accounts := []account.Account{
		{Email: "alpha@example.com", Password: "pass", DisplayName: "Alpha", DefaultMod: mods.DefaultModVanilla},
	}
	accountsFile := filepath.Join(t.TempDir(), "accounts.csv")
	err := account.SaveAccounts(accountsFile, accounts)
	assert.NoError(t, err)

	withTestInput(t, "1\n\n", func() {
		err = clearDefaultMods(accounts, accountsFile, nil)
	})

	assert.NoError(t, err)
	assert.Equal(t, "", accounts[0].DefaultMod)

	reloaded, loadErr := account.LoadAccounts(accountsFile)
	assert.NoError(t, loadErr)
	assert.Equal(t, "", reloaded[0].DefaultMod)
}
