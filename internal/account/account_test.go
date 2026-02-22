package account

import (
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
		{Email: "test1@email.com", Password: "pass1", DisplayName: "Account1"},
		{Email: "test2@email.com", Password: "pass2", DisplayName: "Account2"},
	}

	err := SaveAccounts(csvPath, accounts)
	assert.NoError(t, err)

	// 讀取
	loaded, err := LoadAccounts(csvPath)
	assert.NoError(t, err)
	assert.Len(t, loaded, 2)

	assert.Equal(t, "test1@email.com", loaded[0].Email)
	assert.Equal(t, "pass1", loaded[0].Password)
	assert.Equal(t, "Account1", loaded[0].DisplayName)

	assert.Equal(t, "test2@email.com", loaded[1].Email)
	assert.Equal(t, "Account2", loaded[1].DisplayName)
}

func TestLoadAccounts_FileNotFound(t *testing.T) {
	_, err := LoadAccounts("nonexistent.csv")
	assert.Error(t, err)
}

func TestLoadAccounts_EmptyFile(t *testing.T) {
	dir := t.TempDir()
	csvPath := filepath.Join(dir, "accounts.csv")

	err := os.WriteFile(csvPath, []byte("Email,Password,DisplayName\n"), 0644)
	assert.NoError(t, err)

	_, err = LoadAccounts(csvPath)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "empty")
}

func TestLoadAccounts_UTF8BOM(t *testing.T) {
	dir := t.TempDir()
	csvPath := filepath.Join(dir, "accounts.csv")

	// 模擬 Windows 記事本存檔（UTF-8 with BOM）
	bom := []byte{0xEF, 0xBB, 0xBF}
	content := "Email,Password,DisplayName\n主帳號@gmail.com,pass123,主帳號-法師\n"
	err := os.WriteFile(csvPath, append(bom, []byte(content)...), 0644)
	assert.NoError(t, err)

	loaded, err := LoadAccounts(csvPath)
	assert.NoError(t, err)
	assert.Len(t, loaded, 1)
	// BOM 不應汙染 Email 欄位
	assert.Equal(t, "主帳號@gmail.com", loaded[0].Email)
	assert.Equal(t, "主帳號-法師", loaded[0].DisplayName)
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
