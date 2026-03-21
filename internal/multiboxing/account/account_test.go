package account

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoadAndSaveAccounts(t *testing.T) {
	dir := t.TempDir()
	csvPath := filepath.Join(dir, "accounts.csv")

	// 建立測試 CSV
	accounts := []Account{
		{Email: "test1@email.com", Password: "pass1", DisplayName: "Account1", LaunchFlags: LaunchFlagNoSound, GraphicsProfile: "main-high"},
		{Email: "test2@email.com", Password: "pass2", DisplayName: "Account2"},
	}

	err := SaveAccounts(csvPath, accounts)
	assert.NoError(t, err)

	data, err := os.ReadFile(csvPath)
	assert.NoError(t, err)
	assert.True(t, bytes.HasPrefix(data, utf8BOM))

	// 讀取
	loaded, err := LoadAccounts(csvPath)
	assert.NoError(t, err)
	assert.Len(t, loaded, 2)

	assert.Equal(t, "test1@email.com", loaded[0].Email)
	assert.Equal(t, "pass1", loaded[0].Password)
	assert.Equal(t, "Account1", loaded[0].DisplayName)
	assert.Equal(t, uint32(LaunchFlagNoSound), loaded[0].LaunchFlags)
	assert.Equal(t, "main-high", loaded[0].GraphicsProfile)

	assert.Equal(t, "test2@email.com", loaded[1].Email)
	assert.Equal(t, "Account2", loaded[1].DisplayName)
	assert.Equal(t, uint32(0), loaded[1].LaunchFlags)
	assert.Equal(t, "", loaded[1].GraphicsProfile)
}

func TestEnsureAccountsFileCreatesTemplate(t *testing.T) {
	dir := t.TempDir()
	csvPath := filepath.Join(dir, "nested", "accounts.csv")

	created, err := EnsureAccountsFile(csvPath)
	assert.NoError(t, err)
	assert.True(t, created)

	data, err := os.ReadFile(csvPath)
	assert.NoError(t, err)
	assert.True(t, bytes.HasPrefix(data, utf8BOM))
	assert.Equal(t, string(accountsCSVTemplate), string(data))
}

func TestEnsureAccountsFileDoesNotOverwriteExistingFile(t *testing.T) {
	dir := t.TempDir()
	csvPath := filepath.Join(dir, "accounts.csv")
	original := []Account{
		{Email: "keep@example.com", Password: "keep", DisplayName: "Keep"},
	}

	err := SaveAccounts(csvPath, original)
	assert.NoError(t, err)

	created, err := EnsureAccountsFile(csvPath)
	assert.NoError(t, err)
	assert.False(t, created)

	loaded, err := LoadAccounts(csvPath)
	assert.NoError(t, err)
	assert.Equal(t, original, loaded)
}

func TestLoadAccounts_FileNotFound(t *testing.T) {
	_, err := LoadAccounts("nonexistent.csv")
	assert.Error(t, err)
}

func TestLoadAccounts_EmptyFile(t *testing.T) {
	dir := t.TempDir()
	csvPath := filepath.Join(dir, "accounts.csv")

	err := os.WriteFile(csvPath, []byte("Email,Password,DisplayName,LaunchFlags\n"), 0644)
	assert.NoError(t, err)

	_, err = LoadAccounts(csvPath)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "empty")
}

func TestLoadAccounts_BackwardCompatibleWithoutLaunchFlagsColumn(t *testing.T) {
	dir := t.TempDir()
	csvPath := filepath.Join(dir, "accounts.csv")

	err := os.WriteFile(csvPath, append(utf8BOM, []byte("Email,Password,DisplayName\nlegacy@example.com,plain,Legacy\n")...), 0o644)
	assert.NoError(t, err)

	loaded, err := LoadAccounts(csvPath)
	assert.NoError(t, err)
	assert.Len(t, loaded, 1)
	assert.Equal(t, uint32(0), loaded[0].LaunchFlags)
}

func TestLoadAccounts_InvalidLaunchFlagsFallsBackToZeroAndRewritesFile(t *testing.T) {
	dir := t.TempDir()
	csvPath := filepath.Join(dir, "accounts.csv")

	content := append(utf8BOM, []byte("Email,Password,DisplayName,LaunchFlags\nlegacy@example.com,plain,Legacy,abc\n")...)
	err := os.WriteFile(csvPath, content, 0o644)
	assert.NoError(t, err)

	loaded, err := LoadAccounts(csvPath)
	assert.NoError(t, err)
	assert.Len(t, loaded, 1)
	assert.Equal(t, uint32(0), loaded[0].LaunchFlags)

	data, err := os.ReadFile(csvPath)
	assert.NoError(t, err)
	assert.Contains(t, string(data), "legacy@example.com,plain,Legacy,0")
}

func TestLoadAccounts_BackwardCompatWithoutGraphicsProfileColumn(t *testing.T) {
	dir := t.TempDir()
	csvPath := filepath.Join(dir, "accounts.csv")

	content := append(utf8BOM, []byte("Email,Password,DisplayName,LaunchFlags,ToolFlags\nlegacy@example.com,plain,Legacy,1,0\n")...)
	err := os.WriteFile(csvPath, content, 0o644)
	assert.NoError(t, err)

	loaded, err := LoadAccounts(csvPath)
	assert.NoError(t, err)
	assert.Len(t, loaded, 1)
	assert.Equal(t, uint32(LaunchFlagNoSound), loaded[0].LaunchFlags)
	assert.Equal(t, uint32(0), loaded[0].ToolFlags)
	assert.Equal(t, "", loaded[0].GraphicsProfile)
}

func TestLoadAccounts_GraphicsProfileColumn(t *testing.T) {
	dir := t.TempDir()
	csvPath := filepath.Join(dir, "accounts.csv")

	content := append(utf8BOM, []byte("Email,Password,DisplayName,LaunchFlags,ToolFlags,GraphicsProfile\nlegacy@example.com,plain,Legacy,1,0,alt-low\n")...)
	err := os.WriteFile(csvPath, content, 0o644)
	assert.NoError(t, err)

	loaded, err := LoadAccounts(csvPath)
	assert.NoError(t, err)
	assert.Len(t, loaded, 1)
	assert.Equal(t, "alt-low", loaded[0].GraphicsProfile)
}

func TestLoadAccounts_RemovesLegacyLowQualityFlagAndRewritesFile(t *testing.T) {
	dir := t.TempDir()
	csvPath := filepath.Join(dir, "accounts.csv")

	content := append(utf8BOM, []byte("Email,Password,DisplayName,LaunchFlags\nlegacy@example.com,plain,Legacy,4\n")...)
	err := os.WriteFile(csvPath, content, 0o644)
	assert.NoError(t, err)

	loaded, err := LoadAccounts(csvPath)
	assert.NoError(t, err)
	assert.Len(t, loaded, 1)
	assert.Equal(t, uint32(0), loaded[0].LaunchFlags)

	data, err := os.ReadFile(csvPath)
	assert.NoError(t, err)
	assert.Contains(t, string(data), "legacy@example.com,plain,Legacy,0,0,")
}

func TestLoadAccounts_RemovesUnsupportedLaunchFlagBitsAndRewritesFile(t *testing.T) {
	dir := t.TempDir()
	csvPath := filepath.Join(dir, "accounts.csv")

	content := append(utf8BOM, []byte("Email,Password,DisplayName,LaunchFlags\nlegacy@example.com,plain,Legacy,31\n")...)
	err := os.WriteFile(csvPath, content, 0o644)
	assert.NoError(t, err)

	loaded, err := LoadAccounts(csvPath)
	assert.NoError(t, err)
	assert.Len(t, loaded, 1)
	assert.Equal(t, uint32(LaunchFlagNoSound), loaded[0].LaunchFlags)

	data, err := os.ReadFile(csvPath)
	assert.NoError(t, err)
	assert.Contains(t, string(data), "legacy@example.com,plain,Legacy,1,0,")
}

func TestIsPasswordEncrypted(t *testing.T) {
	assert.False(t, IsPasswordEncrypted("plaintext"))
	assert.False(t, IsPasswordEncrypted(""))
	assert.True(t, IsPasswordEncrypted("ENC:base64data"))
}

func TestEncryptAndDecryptPassword(t *testing.T) {
	original := "mySecretPassword123!"

	encrypted, err := EncryptPassword(original)
	assert.NoError(t, err)
	assert.True(t, IsPasswordEncrypted(encrypted))

	decrypted, err := DecryptPassword(encrypted)
	assert.NoError(t, err)
	assert.Equal(t, original, decrypted)
}

func TestDecryptPassword_Plaintext(t *testing.T) {
	// 未加密的密碼應原樣返回
	result, err := DecryptPassword("plaintext")
	assert.NoError(t, err)
	assert.Equal(t, "plaintext", result)
}

func TestEncryptPlaintextPasswords(t *testing.T) {
	dir := t.TempDir()
	csvPath := filepath.Join(dir, "accounts.csv")

	accounts := []Account{
		{Email: "a@b.com", Password: "plain1", DisplayName: "A"},
		{Email: "c@d.com", Password: "ENC:already", DisplayName: "B"},
	}

	err := SaveAccounts(csvPath, accounts)
	assert.NoError(t, err)

	changed, err := EncryptPlaintextPasswords(csvPath, accounts)
	assert.NoError(t, err)
	assert.True(t, changed)

	// 第一個帳號密碼應被加密
	assert.True(t, IsPasswordEncrypted(accounts[0].Password))
	// 第二個不應被改動
	assert.Equal(t, "ENC:already", accounts[1].Password)

	// 重新讀取驗證持久化
	reloaded, err := LoadAccounts(csvPath)
	assert.NoError(t, err)
	assert.True(t, IsPasswordEncrypted(reloaded[0].Password))
}
