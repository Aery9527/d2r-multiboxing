package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"d2rhl/internal/common/config"
)

// PickD2RPath opens a Windows file picker and returns the selected D2R.exe path.
// dialogTitle is shown in the file picker window title bar.
func PickD2RPath(currentPath string, dialogTitle string) (string, error) {
	script := buildD2RPathPickerScript(dialogInitialDir(currentPath), dialogTitle)

	cmd := exec.Command("powershell.exe", "-NoProfile", "-STA", "-ExecutionPolicy", "Bypass", "-Command", script)
	output, err := cmd.CombinedOutput()
	if err != nil {
		message := strings.TrimSpace(string(output))
		if message != "" {
			return "", fmt.Errorf("failed to open D2R file picker: %s", message)
		}
		return "", fmt.Errorf("failed to open D2R file picker: %w", err)
	}

	selectedPath := strings.TrimSpace(string(output))
	if selectedPath == "" {
		return "", nil
	}

	if err := config.ValidateD2RPath(selectedPath); err != nil {
		return "", err
	}
	return selectedPath, nil
}

func dialogInitialDir(currentPath string) string {
	candidates := []string{filepath.Dir(currentPath)}
	for _, candidate := range candidates {
		if candidate == "" || candidate == "." {
			continue
		}

		info, err := os.Stat(candidate)
		if err == nil && info.IsDir() {
			return candidate
		}
	}
	return ""
}

func buildD2RPathPickerScript(initialDir string, title string) string {
	escapedInitialDir := powerShellSingleQuote(initialDir)
	escapedTitle := powerShellSingleQuote(title)

	return fmt.Sprintf(`
Add-Type -AssemblyName System.Windows.Forms | Out-Null
$dialog = New-Object System.Windows.Forms.OpenFileDialog
$dialog.Title = '%s'
$dialog.Filter = 'D2R.exe|D2R.exe|Executable files (*.exe)|*.exe|All files (*.*)|*.*'
$dialog.CheckFileExists = $true
$dialog.Multiselect = $false
$dialog.FileName = 'D2R.exe'
if ('%s' -ne '') {
    $dialog.InitialDirectory = '%s'
}
if ($dialog.ShowDialog() -eq [System.Windows.Forms.DialogResult]::OK) {
    [Console]::Out.Write($dialog.FileName)
}
`, escapedTitle, escapedInitialDir, escapedInitialDir)
}

func powerShellSingleQuote(value string) string {
	return strings.ReplaceAll(value, "'", "''")
}
