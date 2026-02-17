package account

import (
	"encoding/base64"
	"fmt"
	"strings"
	"syscall"
	"unsafe"
)

var (
	crypt32              = syscall.NewLazyDLL("crypt32.dll")
	kernel32             = syscall.NewLazyDLL("kernel32.dll")
	procCryptProtectData = crypt32.NewProc("CryptProtectData")
	procCryptUnprotData  = crypt32.NewProc("CryptUnprotectData")
	procLocalFree        = kernel32.NewProc("LocalFree")
)

// dataBlob represents the Windows DATA_BLOB structure used by DPAPI.
type dataBlob struct {
	Size uint32
	Data *byte
}

// EncryptPassword encrypts a plaintext password using Windows DPAPI
// and returns it as "ENC:<base64>" string.
func EncryptPassword(plaintext string) (string, error) {
	input := []byte(plaintext)
	inputBlob := dataBlob{
		Size: uint32(len(input)),
		Data: &input[0],
	}

	var outputBlob dataBlob
	ret, _, err := procCryptProtectData.Call(
		uintptr(unsafe.Pointer(&inputBlob)),
		0, // description
		0, // optional entropy
		0, // reserved
		0, // prompt struct
		0, // flags
		uintptr(unsafe.Pointer(&outputBlob)),
	)
	if ret == 0 {
		return "", fmt.Errorf("CryptProtectData failed: %w", err)
	}
	defer procLocalFree.Call(uintptr(unsafe.Pointer(outputBlob.Data)))

	encryptedBytes := unsafe.Slice(outputBlob.Data, outputBlob.Size)
	encoded := base64.StdEncoding.EncodeToString(encryptedBytes)

	return encryptedPrefix + encoded, nil
}

// DecryptPassword decrypts an "ENC:<base64>" password using Windows DPAPI.
func DecryptPassword(encrypted string) (string, error) {
	if !strings.HasPrefix(encrypted, encryptedPrefix) {
		return encrypted, nil
	}

	encoded := strings.TrimPrefix(encrypted, encryptedPrefix)
	cipherBytes, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return "", fmt.Errorf("failed to decode base64: %w", err)
	}

	inputBlob := dataBlob{
		Size: uint32(len(cipherBytes)),
		Data: &cipherBytes[0],
	}

	var outputBlob dataBlob
	ret, _, callErr := procCryptUnprotData.Call(
		uintptr(unsafe.Pointer(&inputBlob)),
		0, // description
		0, // optional entropy
		0, // reserved
		0, // prompt struct
		0, // flags
		uintptr(unsafe.Pointer(&outputBlob)),
	)
	if ret == 0 {
		return "", fmt.Errorf("CryptUnprotectData failed: %w", callErr)
	}
	defer procLocalFree.Call(uintptr(unsafe.Pointer(outputBlob.Data)))

	decryptedBytes := unsafe.Slice(outputBlob.Data, outputBlob.Size)
	return string(decryptedBytes), nil
}
