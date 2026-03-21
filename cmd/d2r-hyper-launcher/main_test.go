package main

import (
	"bufio"
	"io"
	"os"
	"path/filepath"
	"strconv"
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

func TestDisplayReleaseTime(t *testing.T) {
	assert.Equal(t, "2026-03-08 13:40:45 release", displayReleaseTime("2026-03-08 13:40:45"))
	assert.Equal(t, "尚未 release", displayReleaseTime(""))
	assert.Equal(t, "尚未 release", displayReleaseTime("   "))
}

func TestDisplayReleaseSummary(t *testing.T) {
	assert.Equal(t, "v1.0.0（2026-03-08 13:40:45 release）", displayReleaseSummary("1.0.0", "2026-03-08 13:40:45"))
	assert.Equal(t, "vdev（尚未 release）", displayReleaseSummary("dev", ""))
}

func TestMaybeShowStartupAnnouncementShowsAnnouncementWhenAccountsFileExists(t *testing.T) {
	originalVersion := version
	originalReleaseTime := releaseTime
	version = "1.0.0"
	releaseTime = "2026-03-08 13:40:45"
	t.Cleanup(func() {
		version = originalVersion
		releaseTime = originalReleaseTime
	})

	originalWaitForAnyKey := ui.waitForAnyKey
	originalCanSingleKeyContinue := ui.canSingleKeyContinue
	t.Cleanup(func() {
		ui.waitForAnyKey = originalWaitForAnyKey
		ui.canSingleKeyContinue = originalCanSingleKeyContinue
	})
	ui.canSingleKeyContinue = func() bool { return true }
	ui.waitForAnyKey = func() error { return nil }

	output := captureStdout(t, func() {
		maybeShowStartupAnnouncement(`C:\Users\User\AppData\Roaming\d2r-hyper-launcher`, false)
	})

	assert.NotContains(t, output, "d2r-hyper-launcher (")
	assert.Contains(t, output, "• 目前版本：v1.0.0（2026-03-08 13:40:45 release）\n")
	assert.Contains(t, output, "? 請按任意鍵繼續...")
}

func TestMaybeShowStartupAnnouncementSkipsAnnouncementWhenAccountsFileWasCreated(t *testing.T) {
	output := captureStdout(t, func() {
		maybeShowStartupAnnouncement(`C:\Users\User\AppData\Roaming\d2r-hyper-launcher`, true)
	})

	assert.Equal(t, "", output)
}

func TestPrintStartupHeaderShowsAppHeader(t *testing.T) {
	originalVersion := version
	version = "1.0.0"
	t.Cleanup(func() {
		version = originalVersion
	})

	output := captureStdout(t, func() {
		printStartupHeader()
	})

	assert.Contains(t, output, "d2r-hyper-launcher (v1.0.0)")
}

func TestAccountLaunchArgsIncludesPerAccountFlagsAfterMods(t *testing.T) {
	acc := account.Account{
		DisplayName: "Alpha",
		LaunchFlags: account.LaunchFlagNoSound,
	}

	args := accountLaunchArgs(acc, []string{"-mod", "sample", "-txt"})

	assert.Equal(t, []string{"-mod", "sample", "-txt", "-ns"}, args)
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
		"  <未啟動> Alpha (alpha@example.com)",
		"  <已啟動> Bravo (bravo@example.com)",
	}, lines)
}

func TestLaunchTargetAccountLinesShowsAccountsToLaunch(t *testing.T) {
	accounts := []*account.Account{
		{DisplayName: "Alpha", Email: "alpha@example.com"},
		{DisplayName: "Bravo", Email: "bravo@example.com"},
	}

	lines := launchTargetAccountLines(accounts)

	assert.Equal(t, []string{
		"  Alpha (alpha@example.com)",
		"  Bravo (bravo@example.com)",
	}, lines)
}

func TestPrintMenuKeepsChoicePromptInsideOptionGroup(t *testing.T) {
	cfg := &config.Config{
		D2RPath:     `C:\Games\D2R\D2R.exe`,
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
	assert.Contains(t, output, "--------------------------------------------------------\n[數字] 啟動指定帳號\n")
	assert.Contains(t, output, "[0]    離線遊玩")
	assert.Contains(t, output, "可選 mod，不需帳密\n")
	assert.Contains(t, output, "[d]    設定啟動間隔")
	assert.Contains(t, output, "30-60 秒（隨機）\n")
	assert.Contains(t, output, "[p]    選擇 D2R.exe 路徑")
	assert.Contains(t, output, "C:\\Games\\D2R\\D2R.exe\n")
	assert.Contains(t, output, "[s]    視窗切換設定")
	assert.Contains(t, output, "已啟用設定：Tab（Tab 鍵）\n")
	assert.Contains(t, output, "[q]    退出\n")
	assert.NotContains(t, output, "是否已啟動的判斷基準")
	assert.NotContains(t, output, "? 請選擇：")
}

func TestSwitcherMenuOptionStatusKeepsSavedBindingWhenDisabled(t *testing.T) {
	cfg := &config.Config{
		Switcher: &config.SwitcherConfig{
			Enabled: false,
			Key:     "Tab",
		},
	}

	assert.Equal(t, "未啟用設定：Tab（Tab 鍵）", switcherMenuOptionStatus(cfg))
}

func TestSwitcherMenuOptionStatusShowsUnsetWhenNoBindingSaved(t *testing.T) {
	assert.Equal(t, "未設定", switcherMenuOptionStatus(&config.Config{}))
}

func TestSwitcherToggleOptionLabelShowsEnableWhenDisabled(t *testing.T) {
	cfg := &config.Config{
		Switcher: &config.SwitcherConfig{
			Enabled: false,
			Key:     "Tab",
		},
	}

	assert.Equal(t, "切換為開啟", switcherToggleOptionLabel(cfg))
}

func TestSwitcherToggleOptionLabelShowsDisableWhenEnabled(t *testing.T) {
	cfg := &config.Config{
		Switcher: &config.SwitcherConfig{
			Enabled: true,
			Key:     "Tab",
		},
	}

	assert.Equal(t, "切換為關閉", switcherToggleOptionLabel(cfg))
}

func TestPrintStartupAnnouncementShowsDisplayNameStatusNote(t *testing.T) {
	originalVersion := version
	originalReleaseTime := releaseTime
	version = "1.0.0"
	releaseTime = "2026-03-08 13:40:45"
	t.Cleanup(func() {
		version = originalVersion
		releaseTime = originalReleaseTime
	})

	output := captureStdout(t, func() {
		printStartupAnnouncement(`C:\Users\User\AppData\Roaming\d2r-hyper-launcher`)
	})

	assert.NotContains(t, output, "d2r-hyper-launcher (")
	assert.Contains(t, output, "• 目前版本：v1.0.0（2026-03-08 13:40:45 release）\n")
	assert.Contains(t, output, "• 資料目錄：C:\\Users\\User\\AppData\\Roaming\\d2r-hyper-launcher\n")
	assert.NotContains(t, output, "D2R 路徑：")
	assert.Contains(t, output, "⚠ 帳號啟動狀態的偵測是用 account.csv 裡的 DisplayName 去對應視窗名稱，\n  所以已經透過該工具開啟 D2R 然後又去修改 DisplayName的話，\n  就會導致啟動狀態顯示不正確。\n")
	assert.Contains(t, output, "⚠ 注意事項：\n")
	assert.Contains(t, output, "  - 建議先把 D2R 設成「視窗化」或「無邊框視窗」\n")
	assert.Contains(t, output, "  - a 批次啟動預設 launch_delay 是 10 秒；舊版預設留下的 5 秒會自動按 10 秒處理，如要調整請回主選單輸入 d。\n")
	assert.Contains(t, output, "  - 本工具為社群自用工具，與 Blizzard Entertainment 無關；使用風險自負。\n")
	assert.NotContains(t, output, "啟動間隔：")
	assert.NotContains(t, output, "視窗切換已啟用：")
}

func TestPauseAfterStartupAnnouncementWaitsForAnyKey(t *testing.T) {
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
		pauseAfterStartupAnnouncement()
	})

	assert.Contains(t, output, "? 請按任意鍵繼續...")
	assert.Equal(t, 1, waitCalled)
}

func TestPauseAfterStartupAnnouncementWarnsWhenWaitFails(t *testing.T) {
	originalWaitForAnyKey := ui.waitForAnyKey
	originalCanSingleKeyContinue := ui.canSingleKeyContinue
	t.Cleanup(func() {
		ui.waitForAnyKey = originalWaitForAnyKey
		ui.canSingleKeyContinue = originalCanSingleKeyContinue
	})

	ui.canSingleKeyContinue = func() bool { return true }
	ui.waitForAnyKey = func() error { return assert.AnError }

	output := captureStdout(t, func() {
		pauseAfterStartupAnnouncement()
	})

	assert.Contains(t, output, "? 請按任意鍵繼續...")
	assert.Contains(t, output, "⚠ 等待按鍵失敗：assert.AnError general error for testing")
}

func TestFormatLaunchDelayMessage(t *testing.T) {
	assert.Equal(t, "  等待 30 秒後啟動下一個帳號：VoidLife", formatLaunchDelayMessage(30, "VoidLife"))
}

func TestFormatLaunchDelayRemainingMessage(t *testing.T) {
	assert.Equal(t, "  還剩 25 秒後啟動下一個帳號：VoidLife", formatLaunchDelayRemainingMessage(25, "VoidLife"))
}

func TestPauseAfterSuccessfulLaunchWaitsThreeSeconds(t *testing.T) {
	originalSleep := launchSuccessPauseSleep
	t.Cleanup(func() {
		launchSuccessPauseSleep = originalSleep
	})

	var slept time.Duration
	launchSuccessPauseSleep = func(d time.Duration) {
		slept = d
	}

	pauseAfterSuccessfulLaunch()

	assert.Equal(t, 3*time.Second, slept)
}

func TestDisplayDelayFixed(t *testing.T) {
	delay := config.LaunchDelayRange{MinSeconds: 30, MaxSeconds: 30}
	assert.Equal(t, "30 秒", displayDelay(delay))
}

func TestDisplayDelayRandom(t *testing.T) {
	delay := config.LaunchDelayRange{MinSeconds: 30, MaxSeconds: 60}
	assert.Equal(t, "30-60 秒（隨機）", displayDelay(delay))
}

func TestLangDefaultsToZhTW(t *testing.T) {
	assert.Equal(t, "請選擇：", lang.Common.SelectPrompt)
	assert.Equal(t, "再見！", lang.Common.Goodbye)
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
	mask := selectedLaunchFlagMask([]int{0}, options)

	assert.Equal(t, uint32(account.LaunchFlagNoSound), mask)
}

func TestAllLaunchFlagMask(t *testing.T) {
	options := account.LaunchFlagOptions()
	mask := allLaunchFlagMask(options)

	assert.Equal(t, uint32(account.LaunchFlagNoSound), mask)
}

func TestPrintAccountList(t *testing.T) {
	accounts := []account.Account{
		{DisplayName: "Alpha", Email: "alpha@example.com", LaunchFlags: account.LaunchFlagNoSound},
		{DisplayName: "Bravo", Email: "bravo@example.com"},
	}

	output := captureStdout(t, func() {
		printAccountList(accounts, runningStatusLabel)
	})

	assert.Contains(t, output, "[1] <")
	assert.Contains(t, output, "Alpha")
	assert.Contains(t, output, "(alpha@example.com)")
	assert.Contains(t, output, "[2] <")
	assert.Contains(t, output, "Bravo")
	assert.Contains(t, output, "(bravo@example.com)")
	assert.NotContains(t, output, "flag：")
}

func TestBuildAccountLaunchFlagTableLines(t *testing.T) {
	options := account.LaunchFlagOptions()
	accounts := []account.Account{
		{LaunchFlags: account.LaunchFlagNoSound},
		{LaunchFlags: 0},
	}

	lines := buildAccountLaunchFlagTableLines(accounts)

	assert.Len(t, lines, len(accounts)+5)
	assert.Equal(t, lines[0], lines[3])
	assert.Equal(t, lines[0], lines[len(lines)-1])

	headerTopCells := parseLaunchFlagTableCells(lines[1])
	headerBottomCells := parseLaunchFlagTableCells(lines[2])
	expectedHeaderTop := []string{"帳號編號"}
	expectedHeaderBottom := []string{""}
	for _, option := range options {
		title, flag := launchFlagTableHeaderLines(option)
		expectedHeaderTop = append(expectedHeaderTop, title)
		expectedHeaderBottom = append(expectedHeaderBottom, flag)
	}
	assert.Equal(t, expectedHeaderTop, headerTopCells)
	assert.Equal(t, expectedHeaderBottom, headerBottomCells)

	assert.Equal(t, expectedLaunchFlagTableCells(1, accounts[0].LaunchFlags, options), parseLaunchFlagTableCells(lines[4]))
	assert.Equal(t, expectedLaunchFlagTableCells(2, accounts[1].LaunchFlags, options), parseLaunchFlagTableCells(lines[5]))
}

func TestCenterLaunchFlagTableCell(t *testing.T) {
	assert.Equal(t, "  v  ", centerLaunchFlagTableCell("v", 5))
	assert.Equal(t, " 帳號 ", centerLaunchFlagTableCell("帳號", 6))
	assert.Equal(t, "alpha", centerLaunchFlagTableCell("alpha", 3))
}

func TestSetupAccountLaunchFlagsShowsFlagTableAfterAccountList(t *testing.T) {
	originalCanSingleKeyContinue := ui.canSingleKeyContinue
	t.Cleanup(func() {
		ui.canSingleKeyContinue = originalCanSingleKeyContinue
	})
	ui.canSingleKeyContinue = func() bool { return false }

	accounts := []account.Account{{DisplayName: "Alpha", Email: "alpha@example.com", LaunchFlags: account.LaunchFlagNoSound}}
	output := captureStdout(t, func() {
		withTestInput(t, "b\n", func() {
			setupAccountLaunchFlags(accounts, "")
		})
	})

	accountListIndex := strings.Index(output, "帳號列表：")
	flagTableIndex := strings.Index(output, "flag 對照表：")
	menuOptionIndex := strings.Index(output, "[1] 設定 flag")
	assert.NotEqual(t, -1, accountListIndex)
	assert.NotEqual(t, -1, flagTableIndex)
	assert.NotEqual(t, -1, menuOptionIndex)
	assert.Less(t, accountListIndex, flagTableIndex)
	assert.Less(t, flagTableIndex, menuOptionIndex)
	assert.Contains(t, output, "關閉聲音")
	assert.Contains(t, output, "-ns / -nosound")
	assert.Contains(t, output, "|    1     |")
	assert.Contains(t, output, "|       v        |")
}

func TestSetupAccountLaunchFlagsShowsFriendlyModeLabels(t *testing.T) {
	accounts := []account.Account{{DisplayName: "Alpha", Email: "alpha@example.com"}}

	output := captureStdout(t, func() {
		withTestInput(t, "1\nb\nb\n", func() {
			setupAccountLaunchFlags(accounts, "")
		})
	})

	assert.Contains(t, output, "[1] 選擇 flag 設定至多個帳號")
	assert.Contains(t, output, "[2] 選擇帳號設定多個 flag")
}

func TestSetupAccountLaunchFlagsShowsCancelModeLabels(t *testing.T) {
	accounts := []account.Account{{DisplayName: "Alpha", Email: "alpha@example.com"}}

	output := captureStdout(t, func() {
		withTestInput(t, "2\nb\nb\n", func() {
			setupAccountLaunchFlags(accounts, "")
		})
	})

	assert.Contains(t, output, "[1] 選擇 flag 取消至多個帳號")
	assert.Contains(t, output, "[2] 選擇帳號取消多個 flag")
}

func TestSetupAccountLaunchFlagsShowsAllAccountsAllFlagsOption(t *testing.T) {
	accounts := []account.Account{{DisplayName: "Alpha", Email: "alpha@example.com"}}

	output := captureStdout(t, func() {
		withTestInput(t, "1\nb\nb\n", func() {
			setupAccountLaunchFlags(accounts, "")
		})
	})

	assert.Contains(t, output, "[3] 設定所有帳號所有 flag")
}

func TestSetupAccountLaunchFlagsDoesNotShowDeprecatedLowQualityFlag(t *testing.T) {
	accounts := []account.Account{{DisplayName: "Alpha", Email: "alpha@example.com"}}

	output := captureStdout(t, func() {
		withTestInput(t, "1\n1\nb\nb\nb\n", func() {
			setupAccountLaunchFlags(accounts, "")
		})
	})

	assert.Contains(t, output, "關閉聲音")
	assert.NotContains(t, output, "-lq")
	assert.NotContains(t, output, "Large Font Mode")
	assert.NotContains(t, output, "術士版本似乎已失效")
}

func TestConfigureAllFlagsForAllAccountsSetsEveryCompatibleFlag(t *testing.T) {
	accounts := []account.Account{
		{DisplayName: "Alpha", Email: "alpha@example.com"},
		{DisplayName: "Bravo", Email: "bravo@example.com"},
	}
	accountsFile := filepath.Join(t.TempDir(), "accounts.csv")

	result := captureStdout(t, func() {
		withTestInput(t, "\n", func() {
			assert.ErrorIs(t, configureAllFlagsForAllAccounts(accounts, accountsFile, true), errNavDone)
		})
	})

	expectedFlags := uint32(account.LaunchFlagNoSound)
	assert.Equal(t, expectedFlags, accounts[0].LaunchFlags)
	assert.Equal(t, expectedFlags, accounts[1].LaunchFlags)
	assert.Contains(t, result, "設定所有帳號所有 flag")
	assert.Contains(t, result, "套用範圍：全部 2 個帳號")

	savedAccounts, err := account.LoadAccounts(accountsFile)
	assert.NoError(t, err)
	assert.Len(t, savedAccounts, 2)
	assert.Equal(t, expectedFlags, savedAccounts[0].LaunchFlags)
	assert.Equal(t, expectedFlags, savedAccounts[1].LaunchFlags)
}

func TestSetupAccountLaunchFlagsReturnsToTopPageAfterSuccessfulChange(t *testing.T) {
	accounts := []account.Account{{DisplayName: "Alpha", Email: "alpha@example.com"}}
	accountsFile := filepath.Join(t.TempDir(), "accounts.csv")

	output := captureStdout(t, func() {
		withTestInput(t, "1\n1\n1\n1\ny\nb\n", func() {
			setupAccountLaunchFlags(accounts, accountsFile)
		})
	})

	assert.Equal(t, 2, strings.Count(output, "帳號啟動 flag 設定"))
	assert.Contains(t, output, "已完成設定。")
	assert.Equal(t, uint32(account.LaunchFlagNoSound), accounts[0].LaunchFlags)
}

func TestSetupAccountLaunchFlagsReturnsToTopPageAfterCanceledBulkChange(t *testing.T) {
	accounts := []account.Account{{DisplayName: "Alpha", Email: "alpha@example.com"}}
	accountsFile := filepath.Join(t.TempDir(), "accounts.csv")

	output := captureStdout(t, func() {
		withTestInput(t, "1\n3\nn\nb\n", func() {
			setupAccountLaunchFlags(accounts, accountsFile)
		})
	})

	assert.Equal(t, 2, strings.Count(output, "帳號啟動 flag 設定"))
	assert.Contains(t, output, "已取消。")
	assert.Equal(t, uint32(0), accounts[0].LaunchFlags)
}

func TestSetupAccountLaunchFlagsBackFromByFlagReturnsToModeMenu(t *testing.T) {
	accounts := []account.Account{{DisplayName: "Alpha", Email: "alpha@example.com"}}

	output := captureStdout(t, func() {
		withTestInput(t, "2\n1\nb\nb\nb\n", func() {
			setupAccountLaunchFlags(accounts, "")
		})
	})

	assert.Equal(t, 2, strings.Count(output, "取消 flag：選擇操作方式"))
	assert.Equal(t, 2, strings.Count(output, "帳號啟動 flag 設定"))
}

func TestSetupAccountLaunchFlagsBackFromByAccountFlagsReturnsToAccountMenu(t *testing.T) {
	accounts := []account.Account{{DisplayName: "Alpha", Email: "alpha@example.com"}}

	output := captureStdout(t, func() {
		withTestInput(t, "2\n2\n1\nb\nb\nb\nb\n", func() {
			setupAccountLaunchFlags(accounts, "")
		})
	})

	assert.Equal(t, 2, strings.Count(output, "取消 flag：先選帳號"))
	assert.Equal(t, 2, strings.Count(output, "取消 flag：選擇操作方式"))
	assert.Equal(t, 2, strings.Count(output, "帳號啟動 flag 設定"))
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

func TestShowInfoAndPause(t *testing.T) {
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
		showInfoAndPause("所有帳號都已在執行中。")
	})

	assert.Contains(t, output, "• 所有帳號都已在執行中。")
	assert.Contains(t, output, "? 請按任意鍵繼續...")
	assert.Equal(t, 1, waitCalled)
}

func TestShowWarningAndPause(t *testing.T) {
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
		showWarningAndPause("Alpha 已在執行中。")
	})

	assert.Contains(t, output, "⚠ Alpha 已在執行中。")
	assert.Contains(t, output, "? 請按任意鍵繼續...")
	assert.Equal(t, 1, waitCalled)
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

func TestSetupSwitcherKeepsCurrentMenuAfterInvalidInput(t *testing.T) {
	originalCanSingleKeyContinue := ui.canSingleKeyContinue
	t.Cleanup(func() {
		ui.canSingleKeyContinue = originalCanSingleKeyContinue
	})
	ui.canSingleKeyContinue = func() bool { return false }

	output := captureStdout(t, func() {
		withTestInput(t, "x\n\nb\n", func() {
			setupSwitcher(&config.Config{}, nil, "")
		})
	})

	assert.Contains(t, output, "✘ 無效輸入，請重試。")
	assert.Equal(t, 2, strings.Count(output, "視窗切換設定"))
}

func TestSetupLaunchDelayKeepsCurrentMenuAfterInvalidInput(t *testing.T) {
	originalCanSingleKeyContinue := ui.canSingleKeyContinue
	t.Cleanup(func() {
		ui.canSingleKeyContinue = originalCanSingleKeyContinue
	})
	ui.canSingleKeyContinue = func() bool { return false }

	output := captureStdout(t, func() {
		withTestInput(t, "abc\n\nb\n", func() {
			setupLaunchDelay(&config.Config{LaunchDelay: config.DefaultLaunchDelayRange()})
		})
	})

	assert.Contains(t, output, "✘ 啟動間隔必須是整數，或使用像 30-60 的範圍格式")
	assert.Equal(t, 2, strings.Count(output, "啟動間隔設定"))
}

func TestSetupAccountLaunchFlagsKeepsCurrentMenuAfterInvalidInput(t *testing.T) {
	originalCanSingleKeyContinue := ui.canSingleKeyContinue
	t.Cleanup(func() {
		ui.canSingleKeyContinue = originalCanSingleKeyContinue
	})
	ui.canSingleKeyContinue = func() bool { return false }

	accounts := []account.Account{{DisplayName: "Alpha", Email: "alpha@example.com"}}
	output := captureStdout(t, func() {
		withTestInput(t, "x\n\nb\n", func() {
			setupAccountLaunchFlags(accounts, "")
		})
	})

	assert.Contains(t, output, "✘ 無效輸入，請重試。")
	assert.Equal(t, 2, strings.Count(output, "帳號啟動 flag 設定"))
}

func TestPromptLaunchRegionKeepsCurrentMenuAfterInvalidInput(t *testing.T) {
	originalCanSingleKeyContinue := ui.canSingleKeyContinue
	t.Cleanup(func() {
		ui.canSingleKeyContinue = originalCanSingleKeyContinue
	})
	ui.canSingleKeyContinue = func() bool { return false }

	output := captureStdout(t, func() {
		withTestInput(t, "x\n\nb\n", func() {
			region, ok := promptLaunchRegion("啟動指定帳號：選擇區域", []*account.Account{
				{DisplayName: "Alpha", Email: "alpha@example.com"},
			})
			assert.False(t, ok)
			assert.Nil(t, region)
		})
	})

	assert.Contains(t, output, "✘ 無效的區域選擇。")
	assert.Equal(t, 2, strings.Count(output, "啟動指定帳號：選擇區域"))
}

func TestPromptLaunchRegionShowsSingleTargetAccount(t *testing.T) {
	output := captureStdout(t, func() {
		withTestInput(t, "b\n", func() {
			region, ok := promptLaunchRegion("啟動指定帳號：選擇區域", []*account.Account{
				{DisplayName: "Alpha", Email: "alpha@example.com"},
			})
			assert.False(t, ok)
			assert.Nil(t, region)
		})
	})

	assert.Contains(t, output, "• 準備啟動的帳號：")
	assert.Contains(t, output, "  Alpha (alpha@example.com)")
}

func TestPromptLaunchRegionShowsBatchTargetAccounts(t *testing.T) {
	output := captureStdout(t, func() {
		withTestInput(t, "b\n", func() {
			region, ok := promptLaunchRegion("啟動所有帳號：選擇區域", []*account.Account{
				{DisplayName: "Alpha", Email: "alpha@example.com"},
				{DisplayName: "Bravo", Email: "bravo@example.com"},
			})
			assert.False(t, ok)
			assert.Nil(t, region)
		})
	})

	assert.Contains(t, output, "• 準備啟動的帳號：")
	assert.Contains(t, output, "  Alpha (alpha@example.com)")
	assert.Contains(t, output, "  Bravo (bravo@example.com)")
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

func parseLaunchFlagTableCells(line string) []string {
	parts := strings.Split(line, "|")
	cells := make([]string, 0, len(parts))
	for _, part := range parts[1 : len(parts)-1] {
		cells = append(cells, strings.TrimSpace(part))
	}
	return cells
}

func expectedLaunchFlagTableCells(accountNumber int, flags uint32, options []account.LaunchFlagOption) []string {
	cells := make([]string, 0, len(options)+1)
	cells = append(cells, strconv.Itoa(accountNumber))
	for _, option := range options {
		cell := ""
		if flags&option.Bit != 0 {
			cell = "v"
		}
		cells = append(cells, cell)
	}
	return cells
}
