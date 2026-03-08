package main

import (
	"bufio"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"d2rhl/internal/common/config"
	"d2rhl/internal/common/d2r"
	"d2rhl/internal/multiboxing/account"

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

func TestPrintMenuKeepsChoicePromptInsideOptionGroup(t *testing.T) {
	cfg := &config.Config{
		LaunchDelay: config.LaunchDelayRange{MinSeconds: 30, MaxSeconds: 60},
		Switcher: &config.SwitcherConfig{
			Enabled: true,
			Key:     "Tab",
		},
	}
	output := captureStdout(t, func() {
		printMenu(nil, cfg)
	})

	assert.Contains(t, output, "========================================================\n"+strings.Repeat(" ", 25)+"主選單"+strings.Repeat(" ", 25)+"\n========================================================\n\n")
	assert.Contains(t, output, "--------------------------------------------------------\n[數字] 啟動指定帳號\n[0]    離線遊玩（可選 mod，不需帳密）")
	assert.Contains(t, output, "[d]    設定啟動間隔（目前：30-60 秒（隨機））\n")
	assert.Contains(t, output, "[s]    視窗切換設定（目前：Tab（Tab 鍵））\n")
	assert.Contains(t, output, "[q]    退出\n")
	assert.NotContains(t, output, "是否已啟動的判斷基準")
	assert.NotContains(t, output, "? 請選擇：")
}

func TestPrintStartupAnnouncementShowsDisplayNameStatusNote(t *testing.T) {
	cfg := &config.Config{
		D2RPath: `C:\Games\D2R\D2R.exe`,
		LaunchDelay: config.LaunchDelayRange{
			MinSeconds: 30,
			MaxSeconds: 60,
		},
	}

	output := captureStdout(t, func() {
		printStartupAnnouncement(`C:\Users\User\AppData\Roaming\d2r-hyper-launcher`, cfg)
	})

	assert.Contains(t, output, "d2r-hyper-launcher (")
	assert.Contains(t, output, "• 資料目錄：C:\\Users\\User\\AppData\\Roaming\\d2r-hyper-launcher\n")
	assert.Contains(t, output, "• D2R 路徑：C:\\Games\\D2R\\D2R.exe\n")
	assert.Contains(t, output, "• 說明：帳號啟動狀態是用 account.csv 裡的 DisplayName 對應視窗名稱；若 D2R 還開著，請先關掉工具再修改 DisplayName，否則狀態偵測可能不正確。\n")
	assert.NotContains(t, output, "啟動間隔：")
	assert.NotContains(t, output, "視窗切換已啟用：")
}

func TestFormatLaunchDelayMessage(t *testing.T) {
	assert.Equal(t, "  等待 30 秒後啟動下一個帳號：VoidLife", formatLaunchDelayMessage(30, "VoidLife"))
}

func TestFormatLaunchDelayRemainingMessage(t *testing.T) {
	assert.Equal(t, "  還剩 25 秒後啟動下一個帳號：VoidLife", formatLaunchDelayRemainingMessage(25, "VoidLife"))
}

func TestParseLaunchDelayInput(t *testing.T) {
	delay, err := parseLaunchDelayInput("45")
	assert.NoError(t, err)
	assert.Equal(t, config.LaunchDelayRange{MinSeconds: 45, MaxSeconds: 45}, delay)

	delay, err = parseLaunchDelayInput("30-60")
	assert.NoError(t, err)
	assert.Equal(t, config.LaunchDelayRange{MinSeconds: 30, MaxSeconds: 60}, delay)
}

func TestParseLaunchDelayInputRejectsNegative(t *testing.T) {
	_, err := parseLaunchDelayInput("9")
	assert.EqualError(t, err, "啟動間隔下限不可小於 10 秒")
}

func TestParseLaunchDelayInputRejectsNonInteger(t *testing.T) {
	_, err := parseLaunchDelayInput("abc")
	assert.EqualError(t, err, "啟動間隔必須是整數，或使用像 30-60 的範圍格式")
}

func TestParseLaunchDelayInputRejectsInvalidRangeOrder(t *testing.T) {
	_, err := parseLaunchDelayInput("60-30")
	assert.EqualError(t, err, "啟動間隔下限不可大於上限")
}

func TestLaunchDelayRangeUsesRandomValue(t *testing.T) {
	delay := config.LaunchDelayRange{MinSeconds: 30, MaxSeconds: 60}
	assert.Equal(t, 42, delay.NextSeconds(func(n int) int {
		assert.Equal(t, 31, n)
		return 12
	}))
}

func TestWaitForNextBatchLaunchReportsRemainingEveryFiveSeconds(t *testing.T) {
	originalSleep := launchDelaySleep
	t.Cleanup(func() {
		launchDelaySleep = originalSleep
	})

	var sleeps []time.Duration
	launchDelaySleep = func(d time.Duration) {
		sleeps = append(sleeps, d)
	}

	output := captureStdout(t, func() {
		waitForNextBatchLaunch(12, "VoidLife")
	})

	assert.Contains(t, output, "• 等待 12 秒後啟動下一個帳號：VoidLife")
	assert.Contains(t, output, "• 還剩 7 秒後啟動下一個帳號：VoidLife")
	assert.Contains(t, output, "• 還剩 2 秒後啟動下一個帳號：VoidLife")
	assert.Equal(t, []time.Duration{5 * time.Second, 5 * time.Second, 2 * time.Second}, sleeps)
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
	withTestInput(t, "", func() {
		ok := ensureLaunchReadyD2RPathWithSetup(cfg, func(*config.Config) bool {
			called = true
			return true
		})

		assert.True(t, ok)
	})
	assert.False(t, called)
}

func TestEnsureLaunchReadyD2RPathWithSetupRunsPathSetup(t *testing.T) {
	tmpDir := t.TempDir()
	validPath := filepath.Join(tmpDir, "D2R.exe")
	assert.NoError(t, os.WriteFile(validPath, []byte("binary"), 0o600))

	cfg := &config.Config{D2RPath: filepath.Join(tmpDir, "missing", "D2R.exe")}
	withTestInput(t, "p\n", func() {
		ok := ensureLaunchReadyD2RPathWithSetup(cfg, func(cfg *config.Config) bool {
			cfg.D2RPath = validPath
			return true
		})

		assert.True(t, ok)
	})
	assert.Equal(t, validPath, cfg.D2RPath)
}

func TestEnsureLaunchReadyD2RPathWithSetupAllowsBackNavigation(t *testing.T) {
	cfg := &config.Config{D2RPath: `C:\missing\D2R.exe`}
	withTestInput(t, "b\n", func() {
		ok := ensureLaunchReadyD2RPathWithSetup(cfg, func(*config.Config) bool {
			t.Fatal("setup should not be called")
			return false
		})

		assert.False(t, ok)
	})
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

func TestShowInputErrorAndPause(t *testing.T) {
	originalWaitForAnyKey := ui.waitForAnyKey
	originalCanSingleKeyContinue := ui.canSingleKeyContinue
	t.Cleanup(func() {
		ui.waitForAnyKey = originalWaitForAnyKey
		ui.canSingleKeyContinue = originalCanSingleKeyContinue
	})

	waitCalled := 0
	ui.canSingleKeyContinue = func() bool { return true }
	ui.waitForAnyKey = func() error {
		waitCalled++
		return nil
	}

	output := captureStdout(t, func() {
		showInputErrorAndPause(`解析失敗：區間 "1-4" 超出可選範圍 1-2`)
	})

	assert.Contains(t, output, `✘ 解析失敗：區間 "1-4" 超出可選範圍 1-2`)
	assert.Contains(t, output, "? 請按任意鍵繼續...")
	assert.Equal(t, 1, waitCalled)
}

func TestShowInvalidInputAndPause(t *testing.T) {
	originalWaitForAnyKey := ui.waitForAnyKey
	originalCanSingleKeyContinue := ui.canSingleKeyContinue
	t.Cleanup(func() {
		ui.waitForAnyKey = originalWaitForAnyKey
		ui.canSingleKeyContinue = originalCanSingleKeyContinue
	})

	ui.canSingleKeyContinue = func() bool { return true }
	ui.waitForAnyKey = func() error { return nil }

	output := captureStdout(t, func() {
		showInvalidInputAndPause()
	})

	assert.Contains(t, output, "✘ 無效輸入，請重試。")
	assert.Contains(t, output, "? 請按任意鍵繼續...")
}

func TestShowInputErrorAndPauseFallsBackToEnterWhenSingleKeyUnavailable(t *testing.T) {
	originalCanSingleKeyContinue := ui.canSingleKeyContinue
	t.Cleanup(func() {
		ui.canSingleKeyContinue = originalCanSingleKeyContinue
	})

	ui.canSingleKeyContinue = func() bool { return false }
	output := captureStdout(t, func() {
		withTestInput(t, "\n", func() {
			showInputErrorAndPause("無效輸入，請重試。")
		})
	})

	assert.Contains(t, output, "✘ 無效輸入，請重試。")
	assert.Contains(t, output, "? 請按 Enter 繼續...")
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

func withTestInput(t *testing.T, input string, fn func()) {
	t.Helper()

	originalReadLine := ui.readLine
	scanner := bufio.NewScanner(strings.NewReader(input))
	ui.readLine = func() (string, bool) {
		if !scanner.Scan() {
			return "", false
		}
		return strings.TrimSpace(scanner.Text()), true
	}
	defer func() {
		ui.readLine = originalReadLine
	}()

	fn()
}
