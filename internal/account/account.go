// Package account provides account management including CSV read/write
// and password encryption using Windows DPAPI.
package account

import (
	"encoding/csv"
	"fmt"
	"os"
	"strconv"
	"strings"
)

const encryptedPrefix = "ENC:"

// Account represents a Battle.net account for D2R.
type Account struct {
	ID          int
	Email       string
	Password    string // 加密後以 "ENC:" 前綴標記
	DisplayName string
	Region      string
}

// IsPasswordEncrypted checks if the password is already encrypted.
func IsPasswordEncrypted(password string) bool {
	return strings.HasPrefix(password, encryptedPrefix)
}

// LoadAccounts reads accounts from a CSV file.
// CSV format: ID,Email,Password,DisplayName,Region (first row is header).
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

	var accounts []Account
	for i, record := range records[1:] { // 跳過 header
		if len(record) < 5 {
			return nil, fmt.Errorf("invalid record at line %d: expected 5 fields, got %d", i+2, len(record))
		}

		id, err := strconv.Atoi(strings.TrimSpace(record[0]))
		if err != nil {
			return nil, fmt.Errorf("invalid ID at line %d: %w", i+2, err)
		}

		accounts = append(accounts, Account{
			ID:          id,
			Email:       strings.TrimSpace(record[1]),
			Password:    strings.TrimSpace(record[2]),
			DisplayName: strings.TrimSpace(record[3]),
			Region:      strings.TrimSpace(record[4]),
		})
	}

	return accounts, nil
}

// SaveAccounts writes accounts to a CSV file.
func SaveAccounts(path string, accounts []Account) error {
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create accounts file: %w", err)
	}
	defer f.Close()

	writer := csv.NewWriter(f)
	defer writer.Flush()

	// header
	if err := writer.Write([]string{"ID", "Email", "Password", "DisplayName", "Region"}); err != nil {
		return fmt.Errorf("failed to write header: %w", err)
	}

	for _, acc := range accounts {
		record := []string{
			strconv.Itoa(acc.ID),
			acc.Email,
			acc.Password,
			acc.DisplayName,
			acc.Region,
		}
		if err := writer.Write(record); err != nil {
			return fmt.Errorf("failed to write account %d: %w", acc.ID, err)
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
			return false, fmt.Errorf("failed to encrypt password for account %d: %w", accounts[i].ID, err)
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
