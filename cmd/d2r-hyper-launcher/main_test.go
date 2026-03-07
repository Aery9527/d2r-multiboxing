package main

import (
	"bufio"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"d2rhl/internal/account"
	"d2rhl/internal/config"
	"d2rhl/internal/d2r"

	"github.com/stretchr/testify/assert"
)

func TestPendingBatchAccountsSkipsRunningAccounts(t *testing.T) {
	accounts := []account.Account{
		{DisplayName: "Alpha"},
		{DisplayName: "Bravo"},
		{DisplayName: "Charlie"},
	}
	runningTitles := map[string]bool{
		d2r.WindowTitle("Bravo"): true,
	}

	pending := pendingBatchAccounts(accounts, runningTitles)

	assert.Len(t, pending, 2)
	assert.Equal(t, "Alpha", pending[0].DisplayName)
	assert.Equal(t, "Charlie", pending[1].DisplayName)
}

func TestDisplayVersion(t *testing.T) {
	assert.Equal(t, "v0.1.0", displayVersion("0.1.0"))
	assert.Equal(t, "v0.1.0", displayVersion("v0.1.0"))
	assert.Equal(t, "vdev", displayVersion("dev"))
}

func TestAccountLaunchArgsIncludesPerAccountFlagsAfterMods(t *testing.T) {
	acc := account.Account{
		DisplayName: "Alpha",
		LaunchFlags: account.LaunchFlagNoSound | account.LaunchFlagSkipLogoVideo,
	}

	args := accountLaunchArgs(acc, []string{"-mod", "sample", "-txt"})

	assert.Equal(t, []string{"-mod", "sample", "-txt", "-ns", "-skiplogovideo"}, args)
}

func TestPendingBatchAccountsReturnsEmptyWhenAllRunning(t *testing.T) {
	accounts := []account.Account{
		{DisplayName: "Alpha"},
		{DisplayName: "Bravo"},
	}
	runningTitles := map[string]bool{
		d2r.WindowTitle("Alpha"): true,
		d2r.WindowTitle("Bravo"): true,
	}

	pending := pendingBatchAccounts(accounts, runningTitles)

	assert.Empty(t, pending)
}

func TestRunningBatchAccountsReturnsOnlyRunningAccounts(t *testing.T) {
	accounts := []account.Account{
		{DisplayName: "Alpha"},
		{DisplayName: "Bravo"},
		{DisplayName: "Charlie"},
	}
	runningTitles := map[string]bool{
		d2r.WindowTitle("Alpha"):   true,
		d2r.WindowTitle("Charlie"): true,
	}

	running := runningBatchAccounts(accounts, runningTitles)

	assert.Len(t, running, 2)
	assert.Equal(t, "Alpha", running[0].DisplayName)
	assert.Equal(t, "Charlie", running[1].DisplayName)
}

func TestBatchAccountStatusLinesShowsRunningAndPendingAccounts(t *testing.T) {
	accounts := []account.Account{
		{DisplayName: "Alpha", Email: "alpha@example.com"},
		{DisplayName: "Bravo", Email: "bravo@example.com"},
	}
	runningTitles := map[string]bool{
		d2r.WindowTitle("Bravo"): true,
	}

	lines := batchAccountStatusLines(accounts, runningTitles)

	assert.Equal(t, []string{
		"  [未啟動] Alpha (alpha@example.com)",
		"  [已啟動] Bravo (bravo@example.com)",
	}, lines)
}

func TestFormatLaunchDelayMessage(t *testing.T) {
	assert.Equal(t, "  等待 30 秒後啟動下一個帳號：VoidLife", formatLaunchDelayMessage(30, "VoidLife"))
}

func TestIsAccountRunningReturnsFalseForMissingWindow(t *testing.T) {
	assert.False(t, isAccountRunning("DefinitelyNotRunningAccount"))
}

func TestEnsureLaunchReadyD2RPathWithSetupAcceptsExistingPath(t *testing.T) {
	tmpDir := t.TempDir()
	d2rPath := filepath.Join(tmpDir, "D2R.exe")
	assert.NoError(t, os.WriteFile(d2rPath, []byte("binary"), 0o600))

	cfg := &config.Config{D2RPath: d2rPath}
	called := false
	scanner := bufio.NewScanner(strings.NewReader(""))

	ok := ensureLaunchReadyD2RPathWithSetup(cfg, scanner, func(*config.Config) bool {
		called = true
		return true
	})

	assert.True(t, ok)
	assert.False(t, called)
}

func TestEnsureLaunchReadyD2RPathWithSetupRunsPathSetup(t *testing.T) {
	tmpDir := t.TempDir()
	validPath := filepath.Join(tmpDir, "D2R.exe")
	assert.NoError(t, os.WriteFile(validPath, []byte("binary"), 0o600))

	cfg := &config.Config{D2RPath: filepath.Join(tmpDir, "missing", "D2R.exe")}
	scanner := bufio.NewScanner(strings.NewReader("p\n"))

	ok := ensureLaunchReadyD2RPathWithSetup(cfg, scanner, func(cfg *config.Config) bool {
		cfg.D2RPath = validPath
		return true
	})

	assert.True(t, ok)
	assert.Equal(t, validPath, cfg.D2RPath)
}

func TestEnsureLaunchReadyD2RPathWithSetupAllowsBackNavigation(t *testing.T) {
	cfg := &config.Config{D2RPath: `C:\missing\D2R.exe`}
	scanner := bufio.NewScanner(strings.NewReader("b\n"))

	ok := ensureLaunchReadyD2RPathWithSetup(cfg, scanner, func(*config.Config) bool {
		t.Fatal("setup should not be called")
		return false
	})

	assert.False(t, ok)
}

func TestParseSelectionInput(t *testing.T) {
	indexes, err := parseSelectionInput("1-3,5,7-8", 8)
	assert.NoError(t, err)
	assert.Equal(t, []int{0, 1, 2, 4, 6, 7}, indexes)
}

func TestParseSelectionInputRejectsReverseRange(t *testing.T) {
	_, err := parseSelectionInput("5-3", 8)
	assert.EqualError(t, err, `區間 "5-3" 起點不可大於終點`)
}

func TestParseSelectionInputRejectsOutOfRange(t *testing.T) {
	_, err := parseSelectionInput("1,9", 8)
	assert.EqualError(t, err, "編號 9 超出可選範圍 1-8")
}

func TestSelectedLaunchFlagMask(t *testing.T) {
	options := account.LaunchFlagOptions()
	mask := selectedLaunchFlagMask([]int{0, 2, 4}, options)

	assert.Equal(t, uint32(account.LaunchFlagNoSound|account.LaunchFlagLowQuality|account.LaunchFlagNoRumble), mask)
}

func TestHasConflictingLaunchFlags(t *testing.T) {
	assert.True(t, hasConflictingLaunchFlags(account.LaunchFlagNoSound|account.LaunchFlagSoundInBackground))
	assert.False(t, hasConflictingLaunchFlags(account.LaunchFlagNoSound|account.LaunchFlagSkipLogoVideo))
}

func TestNormalizeLaunchFlags(t *testing.T) {
	flags := normalizeLaunchFlags(account.LaunchFlagNoSound|account.LaunchFlagSoundInBackground, account.LaunchFlagNoSound)
	assert.Equal(t, uint32(account.LaunchFlagNoSound), flags)

	flags = normalizeLaunchFlags(account.LaunchFlagNoSound|account.LaunchFlagSoundInBackground, account.LaunchFlagSoundInBackground)
	assert.Equal(t, uint32(account.LaunchFlagSoundInBackground), flags)
}

func TestPrintAccountLaunchFlagSummary(t *testing.T) {
	accounts := []account.Account{
		{DisplayName: "Alpha", Email: "alpha@example.com", LaunchFlags: account.LaunchFlagNoSound},
		{DisplayName: "Bravo", Email: "bravo@example.com"},
	}

	output := captureStdout(t, func() {
		printAccountLaunchFlagSummary(accounts)
	})

	assert.Contains(t, output, "[1] Alpha (alpha@example.com)  flag：關閉聲音")
	assert.Contains(t, output, "[2] Bravo (bravo@example.com)  flag：無")
}

func captureStdout(t *testing.T, fn func()) string {
	t.Helper()

	original := os.Stdout
	r, w, err := os.Pipe()
	assert.NoError(t, err)
	os.Stdout = w

	done := make(chan string, 1)
	go func() {
		data, _ := io.ReadAll(r)
		done <- string(data)
	}()

	fn()

	assert.NoError(t, w.Close())
	os.Stdout = original
	output := <-done
	assert.NoError(t, r.Close())
	return output
}
