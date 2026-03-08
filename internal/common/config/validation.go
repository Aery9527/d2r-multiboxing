package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ValidateD2RPath verifies that the configured path points to an existing D2R.exe file.
func ValidateD2RPath(path string) error {
	if strings.TrimSpace(path) == "" {
		return fmt.Errorf("selected D2R path is empty")
	}
	if !strings.EqualFold(filepath.Base(path), "D2R.exe") {
		return fmt.Errorf("selected file must be D2R.exe")
	}

	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("selected D2R.exe does not exist: %w", err)
	}
	if info.IsDir() {
		return fmt.Errorf("selected D2R path points to a directory")
	}
	return nil
}
