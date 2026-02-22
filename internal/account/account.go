// Package account provides account management including CSV read/write
// and password encryption using Windows DPAPI.
package account

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"strings"
)

const encryptedPrefix = "ENC:"

// Account represents a Battle.net account for D2R.
type Account struct {
	Email       string
	Password    string // 加密後以 "ENC:" 前綴標記
	DisplayName string
}

// IsPasswordEncrypted checks if the password is already encrypted.
func IsPasswordEncrypted(password string) bool {
	return strings.HasPrefix(password, encryptedPrefix)
}

// utf8BOM is the UTF-8 byte order mark written by Windows Notepad and some editors.
var utf8BOM = []byte{0xEF, 0xBB, 0xBF}

// LoadAccounts reads accounts from a CSV file.
// CSV format: Email,Password,DisplayName (first row is header).
// Automatically strips UTF-8 BOM if present (written by Windows Notepad).
func LoadAccounts(path string) ([]Account, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open accounts file: %w", err)
	}
	defer f.Close()

	data, err := io.ReadAll(f)
	if err != nil {
		return nil, fmt.Errorf("failed to read accounts file: %w", err)
	}

	// 去除 UTF-8 BOM（Windows 記事本預設會加上 BOM）
	data = bytes.TrimPrefix(data, utf8BOM)

	reader := csv.NewReader(bytes.NewReader(data))
	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("failed to read CSV: %w", err)
	}

	if len(records) < 2 {
		return nil, fmt.Errorf("accounts file is empty (only header or no data)")
	}

	var accounts []Account
	for i, record := range records[1:] { // 跳過 header
		if len(record) < 3 {
			return nil, fmt.Errorf("invalid record at line %d: expected 3 fields, got %d", i+2, len(record))
		}

		accounts = append(accounts, Account{
			Email:       strings.TrimSpace(record[0]),
			Password:    strings.TrimSpace(record[1]),
			DisplayName: strings.TrimSpace(record[2]),
		})
	}

	return accounts, nil
}

// SaveAccounts writes accounts to a CSV file.
func SaveAccounts(path string, accounts []Account) error {
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0o600)
	if err != nil {
		return fmt.Errorf("failed to create accounts file: %w", err)
	}
	defer f.Close()

	writer := csv.NewWriter(f)
	defer writer.Flush()

	// header
	if err := writer.Write([]string{"Email", "Password", "DisplayName"}); err != nil {
		return fmt.Errorf("failed to write header: %w", err)
	}

	for i, acc := range accounts {
		record := []string{
			acc.Email,
			acc.Password,
			acc.DisplayName,
		}
		if err := writer.Write(record); err != nil {
			return fmt.Errorf("failed to write account #%d: %w", i+1, err)
		}
	}

	return nil
}

// EncryptPlaintextPasswords encrypts any plaintext passwords in-place
// and saves the updated accounts back to the file.
func EncryptPlaintextPasswords(path string, accounts []Account) (bool, error) {
	changed := false
	for i := range accounts {
		if accounts[i].Password == "" || IsPasswordEncrypted(accounts[i].Password) {
			continue
		}
		encrypted, err := EncryptPassword(accounts[i].Password)
		if err != nil {
			return false, fmt.Errorf("failed to encrypt password for account #%d: %w", i+1, err)
		}
		accounts[i].Password = encrypted
		changed = true
	}

	if changed {
		if err := SaveAccounts(path, accounts); err != nil {
			return false, fmt.Errorf("failed to save encrypted passwords: %w", err)
		}
	}

	return changed, nil
}

// GetDecryptedPassword returns the decrypted password for an account.
func GetDecryptedPassword(acc *Account) (string, error) {
	if !IsPasswordEncrypted(acc.Password) {
		return acc.Password, nil
	}
	return DecryptPassword(acc.Password)
}
