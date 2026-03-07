// Package account provides account management including CSV read/write
// and password encryption using Windows DPAPI.
package account

import (
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

var utf8BOM = []byte{0xEF, 0xBB, 0xBF}

const encryptedPrefix = "ENC:"

// Account represents a Battle.net account for D2R.
type Account struct {
	Email       string
	Password    string // 加密後以 "ENC:" 前綴標記
	DisplayName string
	LaunchFlags uint32
}

// IsPasswordEncrypted checks if the password is already encrypted.
func IsPasswordEncrypted(password string) bool {
	return strings.HasPrefix(password, encryptedPrefix)
}

// LoadAccounts reads accounts from a CSV file.
// CSV format: Email,Password,DisplayName[,LaunchFlags] (first row is header).
func LoadAccounts(path string) ([]Account, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open accounts file: %w", err)
	}
	defer f.Close()

	reader := csv.NewReader(f)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("failed to read CSV: %w", err)
	}

	if len(records) < 2 {
		return nil, fmt.Errorf("accounts file is empty (only header or no data)")
	}

	var (
		accounts         []Account
		sanitizedInvalid bool
	)
	for i, record := range records[1:] { // 跳過 header
		if len(record) < 3 {
			return nil, fmt.Errorf("invalid record at line %d: expected 3 fields, got %d", i+2, len(record))
		}

		var launchFlags uint32
		if len(record) >= 4 {
			value := strings.TrimSpace(record[3])
			if value != "" {
				parsed, err := strconv.ParseUint(value, 10, 32)
				if err != nil {
					sanitizedInvalid = true
				} else {
					launchFlags = uint32(parsed)
				}
			}
		}

		accounts = append(accounts, Account{
			Email:       strings.TrimSpace(record[0]),
			Password:    strings.TrimSpace(record[1]),
			DisplayName: strings.TrimSpace(record[2]),
			LaunchFlags: launchFlags,
		})
	}

	if sanitizedInvalid {
		if err := SaveAccounts(path, accounts); err != nil {
			return nil, fmt.Errorf("failed to reset invalid LaunchFlags to 0: %w", err)
		}
	}

	return accounts, nil
}

// SaveAccounts writes accounts to a CSV file.
func SaveAccounts(path string, accounts []Account) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return fmt.Errorf("failed to create accounts directory: %w", err)
	}

	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0o600)
	if err != nil {
		return fmt.Errorf("failed to create accounts file: %w", err)
	}
	defer f.Close()

	if _, err := f.Write(utf8BOM); err != nil {
		return fmt.Errorf("failed to write UTF-8 BOM: %w", err)
	}

	writer := csv.NewWriter(f)
	defer writer.Flush()

	// header
	if err := writer.Write([]string{"Email", "Password", "DisplayName", "LaunchFlags"}); err != nil {
		return fmt.Errorf("failed to write header: %w", err)
	}

	for i, acc := range accounts {
		record := []string{
			acc.Email,
			acc.Password,
			acc.DisplayName,
			strconv.FormatUint(uint64(acc.LaunchFlags), 10),
		}
		if err := writer.Write(record); err != nil {
			return fmt.Errorf("failed to write account #%d: %w", i+1, err)
		}
	}

	return nil
}

// EnsureAccountsFile creates accounts.csv with default template rows when it does not exist yet.
func EnsureAccountsFile(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return false, nil
	}
	if !os.IsNotExist(err) {
		return false, fmt.Errorf("failed to stat accounts file: %w", err)
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return false, fmt.Errorf("failed to create accounts directory: %w", err)
	}
	if err := os.WriteFile(path, accountsCSVTemplate, 0o600); err != nil {
		return false, fmt.Errorf("failed to create default accounts file: %w", err)
	}
	return true, nil
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
