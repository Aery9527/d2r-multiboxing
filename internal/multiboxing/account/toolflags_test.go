package account

import (
	"os"
	"path/filepath"
	"testing"

	"d2rhl/internal/multiboxing/mods"

	"github.com/stretchr/testify/assert"
)

func TestSkipSwitcher(t *testing.T) {
	assert.False(t, SkipSwitcher(0))
	assert.True(t, SkipSwitcher(ToolFlagSkipSwitcher))
	assert.False(t, SkipSwitcher(ToolFlagSkipSwitcher<<1)) // different bit
}

func TestSanitizeToolFlags(t *testing.T) {
	assert.Equal(t, uint32(0), SanitizeToolFlags(0))
	assert.Equal(t, uint32(ToolFlagSkipSwitcher), SanitizeToolFlags(ToolFlagSkipSwitcher))
	// unsupported bits should be stripped
	assert.Equal(t, uint32(ToolFlagSkipSwitcher), SanitizeToolFlags(ToolFlagSkipSwitcher|(1<<1)|(1<<5)))
}

func TestExcludedFromSwitcher(t *testing.T) {
	accounts := []Account{
		{DisplayName: "帳號A", ToolFlags: 0},
		{DisplayName: "帳號B", ToolFlags: ToolFlagSkipSwitcher},
		{DisplayName: "帳號C", ToolFlags: 0},
		{DisplayName: "帳號D", ToolFlags: ToolFlagSkipSwitcher},
	}
	excluded := ExcludedFromSwitcher(accounts)
	assert.Equal(t, []string{"帳號B", "帳號D"}, excluded)
}

func TestExcludedFromSwitcher_NoneExcluded(t *testing.T) {
	accounts := []Account{
		{DisplayName: "帳號A", ToolFlags: 0},
		{DisplayName: "帳號B", ToolFlags: 0},
	}
	excluded := ExcludedFromSwitcher(accounts)
	assert.Nil(t, excluded)
}

func TestToolFlagOptions(t *testing.T) {
	options := ToolFlagOptions()
	assert.NotEmpty(t, options)
	// ToolFlagSkipSwitcher must be present
	var found bool
	for _, o := range options {
		if o.Bit == ToolFlagSkipSwitcher {
			found = true
		}
	}
	assert.True(t, found, "ToolFlagSkipSwitcher should be in ToolFlagOptions")
}

func TestLoadAccounts_ToolFlagsColumn(t *testing.T) {
	dir := t.TempDir()
	csvPath := filepath.Join(dir, "accounts.csv")

	accounts := []Account{
		{Email: "a@b.com", Password: "pass", DisplayName: "AccA", LaunchFlags: 0, ToolFlags: ToolFlagSkipSwitcher, GraphicsProfile: "boss-run", DefaultRegion: "NA", DefaultMod: mods.DefaultModVanilla},
		{Email: "c@d.com", Password: "pass", DisplayName: "AccB", LaunchFlags: 0, ToolFlags: 0},
	}

	err := SaveAccounts(csvPath, accounts)
	assert.NoError(t, err)

	loaded, err := LoadAccounts(csvPath)
	assert.NoError(t, err)
	assert.Len(t, loaded, 2)
	assert.Equal(t, uint32(ToolFlagSkipSwitcher), loaded[0].ToolFlags)
	assert.Equal(t, uint32(0), loaded[1].ToolFlags)
	assert.Equal(t, "boss-run", loaded[0].GraphicsProfile)
	assert.Equal(t, "NA", loaded[0].DefaultRegion)
	assert.Equal(t, mods.DefaultModVanilla, loaded[0].DefaultMod)
	assert.Equal(t, "", loaded[1].GraphicsProfile)
	assert.Equal(t, "", loaded[1].DefaultRegion)
	assert.Equal(t, "", loaded[1].DefaultMod)
}

func TestLoadAccounts_BackwardCompatWithoutToolFlagsColumn(t *testing.T) {
	dir := t.TempDir()
	csvPath := filepath.Join(dir, "accounts.csv")

	// old 4-column format
	content := append(utf8BOM, []byte("Email,Password,DisplayName,LaunchFlags\nacc@b.com,pass,AccA,1\n")...)
	err := os.WriteFile(csvPath, content, 0o644)
	assert.NoError(t, err)

	loaded, err := LoadAccounts(csvPath)
	assert.NoError(t, err)
	assert.Len(t, loaded, 1)
	assert.Equal(t, uint32(0), loaded[0].ToolFlags, "ToolFlags should default to 0 for old 4-column CSVs")
}

func TestLoadAccounts_InvalidToolFlagsFallsBackToZero(t *testing.T) {
	dir := t.TempDir()
	csvPath := filepath.Join(dir, "accounts.csv")

	content := append(utf8BOM, []byte("Email,Password,DisplayName,LaunchFlags,ToolFlags\nacc@b.com,pass,AccA,0,bad\n")...)
	err := os.WriteFile(csvPath, content, 0o644)
	assert.NoError(t, err)

	loaded, err := LoadAccounts(csvPath)
	assert.NoError(t, err)
	assert.Equal(t, uint32(0), loaded[0].ToolFlags)
}
