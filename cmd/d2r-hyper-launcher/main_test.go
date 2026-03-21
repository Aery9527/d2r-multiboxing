package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"d2rhl/internal/common/config"
	"d2rhl/internal/common/d2r"
	"d2rhl/internal/common/locale"
	"d2rhl/internal/multiboxing/account"
	"d2rhl/internal/multiboxing/mods"

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
	released := displayReleaseTime("2026-03-08 13:40:45")
	assert.Contains(t, released, "2026-03-08 13:40:45")
	assert.NotEqual(t, "2026-03-08 13:40:45", released)

	unreleased := displayReleaseTime("")
	assert.NotEmpty(t, unreleased)
	assert.Equal(t, unreleased, displayReleaseTime("   "))
}

func TestDisplayReleaseSummary(t *testing.T) {
	released := displayReleaseSummary("1.0.0", "2026-03-08 13:40:45")
	assert.Contains(t, released, displayVersion("1.0.0"))
	assert.Contains(t, released, displayReleaseTime("2026-03-08 13:40:45"))

	unreleased := displayReleaseSummary("dev", "")
	assert.Contains(t, unreleased, displayVersion("dev"))
	assert.Contains(t, unreleased, displayReleaseTime(""))
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
	waitCalled := 0
	ui.waitForAnyKey = func() error {
		waitCalled++
		return nil
	}

	output := captureStdout(t, func() {
		maybeShowStartupAnnouncement(`C:\Users\User\AppData\Roaming\d2r-hyper-launcher`, false)
	})

	assert.Equal(t, -1, firstLineIndex(nonEmptyOutputLines(output), func(line string) bool {
		return line == ui.style.headerDivider
	}))
	assert.Len(t, linesWithPrefix(output, ui.prefix(uiMessageInfo)+" "), 2)
	assert.Len(t, linesWithPrefix(output, ui.prefix(uiMessageWarning)+" "), 2)
	assert.Len(t, linesWithPrefix(output, ui.prefix(uiMessagePrompt)+" "), 1)
	assert.Equal(t, 1, waitCalled)
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

	lines := nonEmptyOutputLines(output)
	assert.Len(t, lines, 3)
	assert.Equal(t, ui.style.headerDivider, lines[0])
	assert.Equal(t, ui.style.headerDivider, lines[2])
	assert.Contains(t, lines[1], displayVersion("1.0.0"))
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

	assert.Len(t, lines, 2)
	assert.Contains(t, lines[0], "Alpha")
	assert.Contains(t, lines[0], "alpha@example.com")
	assert.Contains(t, lines[1], "Bravo")
	assert.Contains(t, lines[1], "bravo@example.com")
	assert.NotEmpty(t, between(lines[0], "<", ">"))
	assert.NotEmpty(t, between(lines[1], "<", ">"))
	assert.NotEqual(t, between(lines[0], "<", ">"), between(lines[1], "<", ">"))
}

func TestLaunchTargetAccountLinesShowsAccountsToLaunch(t *testing.T) {
	accounts := []*account.Account{
		{DisplayName: "Alpha", Email: "alpha@example.com", GraphicsProfile: "1080p-low", DefaultRegion: "EU", DefaultMod: mods.DefaultModVanilla},
		{DisplayName: "Bravo", Email: "bravo@example.com", DefaultMod: "sample-mod"},
	}

	lines := launchTargetAccountLines(accounts, []string{"sample-mod"}, []string{"1080p-low"})

	assert.Len(t, lines, 2)
	assert.Contains(t, lines[0], "Alpha")
	assert.Contains(t, lines[0], "alpha@example.com")
	assert.Contains(t, lines[0], "EU")
	assert.Contains(t, lines[0], lang.DefaultMods.StatusVanilla)
	assert.Contains(t, lines[0], "1080p-low")
	assert.Contains(t, lines[1], "Bravo")
	assert.Contains(t, lines[1], "bravo@example.com")
	assert.Contains(t, lines[1], lang.RegionDefaults.StatusUnassigned)
	assert.Contains(t, lines[1], "sample-mod")
	assert.Contains(t, lines[1], lang.GraphicsProfiles.StatusUnassigned)
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

	assert.Equal(t, 1, countMenuBlocksWithKeys(output, []string{lang.MainMenu.OptByNumberKey, "0", "a", "d", "f", "g", "m", "v", "p", "s", "r", "l", "q"}))
	assert.Empty(t, linesWithPrefix(output, ui.prefix(uiMessagePrompt)+" "))

	delayLine, ok := findMenuOptionLine(output, "d")
	assert.True(t, ok)
	assert.Contains(t, delayLine, displayDelay(cfg.LaunchDelay))

	pathLine, ok := findMenuOptionLine(output, "p")
	assert.True(t, ok)
	assert.Contains(t, pathLine, cfg.D2RPath)

	switcherLine, ok := findMenuOptionLine(output, "s")
	assert.True(t, ok)
	assert.Contains(t, switcherLine, switcherMenuOptionStatus(cfg))
}

func TestSwitcherMenuOptionStatusKeepsSavedBindingWhenDisabled(t *testing.T) {
	cfg := &config.Config{
		Switcher: &config.SwitcherConfig{
			Enabled: false,
			Key:     "Tab",
		},
	}

	status := switcherMenuOptionStatus(cfg)
	assert.NotEmpty(t, status)
	assert.Contains(t, status, "Tab")
}

func TestSwitcherMenuOptionStatusShowsUnsetWhenNoBindingSaved(t *testing.T) {
	assert.NotEmpty(t, switcherMenuOptionStatus(&config.Config{}))
}

func TestSwitcherToggleOptionLabelShowsEnableWhenDisabled(t *testing.T) {
	cfg := &config.Config{
		Switcher: &config.SwitcherConfig{
			Enabled: false,
			Key:     "Tab",
		},
	}

	assert.NotEmpty(t, switcherToggleOptionLabel(cfg))
}

func TestSwitcherToggleOptionLabelShowsDisableWhenEnabled(t *testing.T) {
	cfg := &config.Config{
		Switcher: &config.SwitcherConfig{
			Enabled: true,
			Key:     "Tab",
		},
	}

	assert.NotEmpty(t, switcherToggleOptionLabel(cfg))
	assert.NotEqual(t, switcherToggleOptionLabel(cfg), switcherToggleOptionLabel(&config.Config{
		Switcher: &config.SwitcherConfig{Enabled: false, Key: "Tab"},
	}))
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

	assert.Equal(t, -1, firstLineIndex(nonEmptyOutputLines(output), func(line string) bool {
		return line == ui.style.headerDivider
	}))
	assert.Len(t, linesWithPrefix(output, ui.prefix(uiMessageInfo)+" "), 2)
	assert.Len(t, linesWithPrefix(output, ui.prefix(uiMessageWarning)+" "), 2)
	assert.Contains(t, output, `C:\Users\User\AppData\Roaming\d2r-hyper-launcher`)
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

	assert.Len(t, linesWithPrefix(output, ui.prefix(uiMessagePrompt)+" "), 1)
	assert.Empty(t, linesWithPrefix(output, ui.prefix(uiMessageWarning)+" "))
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

	assert.Len(t, linesWithPrefix(output, ui.prefix(uiMessagePrompt)+" "), 1)
	assert.Len(t, linesWithPrefix(output, ui.prefix(uiMessageWarning)+" "), 1)
	assert.Contains(t, output, assert.AnError.Error())
}

func TestFormatLaunchDelayMessage(t *testing.T) {
	message := formatLaunchDelayMessage(30, "VoidLife")
	assert.True(t, strings.HasPrefix(message, "  "))
	assert.Contains(t, message, "30")
	assert.Contains(t, message, "VoidLife")
}

func TestFormatLaunchDelayRemainingMessage(t *testing.T) {
	message := formatLaunchDelayRemainingMessage(25, "VoidLife")
	assert.True(t, strings.HasPrefix(message, "  "))
	assert.Contains(t, message, "25")
	assert.Contains(t, message, "VoidLife")
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
	display := displayDelay(delay)
	assert.Contains(t, display, "30")
	assert.NotContains(t, display, "60")
}

func TestDisplayDelayRandom(t *testing.T) {
	delay := config.LaunchDelayRange{MinSeconds: 30, MaxSeconds: 60}
	display := displayDelay(delay)
	assert.Contains(t, display, "30")
	assert.Contains(t, display, "60")
}

func TestLangDefaultsToZhTW(t *testing.T) {
	defaultLang := locale.Get(locale.LocaleZhTW)
	assert.Equal(t, defaultLang.Common.SelectPrompt, lang.Common.SelectPrompt)
	assert.Equal(t, defaultLang.Common.Goodbye, lang.Common.Goodbye)
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
	assert.Error(t, err)
}

func TestParseLaunchDelayInputRejectsNonInteger(t *testing.T) {
	_, err := parseLaunchDelayInput("abc")
	assert.Error(t, err)
}

func TestParseLaunchDelayInputRejectsInvalidRangeOrder(t *testing.T) {
	_, err := parseLaunchDelayInput("60-30")
	assert.Error(t, err)
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

	lines := linesWithPrefix(output, ui.prefix(uiMessageInfo)+" ")
	assert.Len(t, lines, 3)
	assert.Equal(t, []int{12}, extractIntsFromString(lines[0]))
	assert.Equal(t, []int{7}, extractIntsFromString(lines[1]))
	assert.Equal(t, []int{2}, extractIntsFromString(lines[2]))
	for _, line := range lines {
		assert.Contains(t, line, "VoidLife")
	}
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
	assert.Error(t, err)
}

func TestParseSelectionInputRejectsOutOfRange(t *testing.T) {
	_, err := parseSelectionInput("1,9", 8)
	assert.Error(t, err)
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

	lines := nonEmptyOutputLines(output)
	accountLines := make([]string, 0, len(lines))
	for _, line := range lines {
		if strings.HasPrefix(line, "[") {
			accountLines = append(accountLines, line)
		}
	}

	assert.Len(t, accountLines, 2)
	assert.Contains(t, accountLines[0], "Alpha")
	assert.Contains(t, accountLines[0], "alpha@example.com")
	assert.True(t, strings.HasSuffix(accountLines[0], "(alpha@example.com)"))
	assert.Contains(t, accountLines[1], "Bravo")
	assert.Contains(t, accountLines[1], "bravo@example.com")
	assert.True(t, strings.HasSuffix(accountLines[1], "(bravo@example.com)"))
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

	lines := nonEmptyOutputLines(output)
	accountListIndex := firstLineIndex(lines, func(line string) bool {
		return strings.HasPrefix(line, "[1] <") && strings.Contains(line, "alpha@example.com")
	})
	flagTableIndex := firstLineIndex(lines, func(line string) bool {
		return strings.HasPrefix(line, "|")
	})
	menuOptionIndex := firstLineIndex(lines, func(line string) bool {
		return line == ui.style.menuDivider
	})
	assert.NotEqual(t, -1, accountListIndex)
	assert.NotEqual(t, -1, flagTableIndex)
	assert.NotEqual(t, -1, menuOptionIndex)
	assert.Less(t, accountListIndex, flagTableIndex)
	assert.Less(t, flagTableIndex, menuOptionIndex)
	assert.Equal(t, 1, countMenuBlocksWithKeys(output, []string{"1", "2", "b", "h", "q"}))
	for _, line := range buildAccountLaunchFlagTableLines(accounts) {
		assert.Contains(t, output, line)
	}
}

func TestSetupAccountLaunchFlagsShowsFriendlyModeLabels(t *testing.T) {
	accounts := []account.Account{{DisplayName: "Alpha", Email: "alpha@example.com"}}

	output := captureStdout(t, func() {
		withTestInput(t, "1\nb\nb\n", func() {
			setupAccountLaunchFlags(accounts, "")
		})
	})

	assert.Equal(t, 1, countMenuBlocksWithKeys(output, []string{"1", "2", "3", "b", "h", "q"}))
}

func TestSetupAccountLaunchFlagsShowsCancelModeLabels(t *testing.T) {
	accounts := []account.Account{{DisplayName: "Alpha", Email: "alpha@example.com"}}

	output := captureStdout(t, func() {
		withTestInput(t, "2\nb\nb\n", func() {
			setupAccountLaunchFlags(accounts, "")
		})
	})

	assert.Equal(t, 1, countMenuBlocksWithKeys(output, []string{"1", "2", "3", "b", "h", "q"}))
}

func TestSetupAccountLaunchFlagsShowsAllAccountsAllFlagsOption(t *testing.T) {
	accounts := []account.Account{{DisplayName: "Alpha", Email: "alpha@example.com"}}

	output := captureStdout(t, func() {
		withTestInput(t, "1\nb\nb\n", func() {
			setupAccountLaunchFlags(accounts, "")
		})
	})

	assert.Equal(t, 1, countMenuBlocksWithKeys(output, []string{"1", "2", "3", "b", "h", "q"}))
}

func TestSetupAccountLaunchFlagsDoesNotShowDeprecatedLowQualityFlag(t *testing.T) {
	accounts := []account.Account{{DisplayName: "Alpha", Email: "alpha@example.com"}}

	output := captureStdout(t, func() {
		withTestInput(t, "1\n1\nb\nb\nb\n", func() {
			setupAccountLaunchFlags(accounts, "")
		})
	})

	assert.Len(t, account.LaunchFlagOptions(), 1)
	assert.Equal(t, 1, countMenuBlocksWithKeys(output, []string{"1", "b", "h", "q"}))
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
	assert.Equal(t, 1, strings.Count(result, ui.prefix(uiMessageSuccess)+" "))

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

	assert.Equal(t, 2, countMenuBlocksWithKeys(output, []string{"1", "2", "b", "h", "q"}))
	assert.Equal(t, 1, strings.Count(output, ui.prefix(uiMessageSuccess)+" "))
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

	assert.Equal(t, 2, countMenuBlocksWithKeys(output, []string{"1", "2", "b", "h", "q"}))
	assert.Equal(t, uint32(0), accounts[0].LaunchFlags)
}

func TestSetupAccountLaunchFlagsBackFromByFlagReturnsToModeMenu(t *testing.T) {
	accounts := []account.Account{{DisplayName: "Alpha", Email: "alpha@example.com"}}

	output := captureStdout(t, func() {
		withTestInput(t, "2\n1\nb\nb\nb\n", func() {
			setupAccountLaunchFlags(accounts, "")
		})
	})

	assert.Equal(t, 2, countMenuBlocksWithKeys(output, []string{"1", "2", "3", "b", "h", "q"}))
	assert.Equal(t, 2, countMenuBlocksWithKeys(output, []string{"1", "2", "b", "h", "q"}))
}

func TestSetupAccountLaunchFlagsBackFromByAccountFlagsReturnsToAccountMenu(t *testing.T) {
	accounts := []account.Account{{DisplayName: "Alpha", Email: "alpha@example.com"}}

	output := captureStdout(t, func() {
		withTestInput(t, "2\n2\n1\nb\nb\nb\nb\n", func() {
			setupAccountLaunchFlags(accounts, "")
		})
	})

	accountSelectionBlocks := 0
	for _, block := range parseMenuBlocks(output) {
		if sameStrings(menuBlockKeys(block), []string{"1", "b", "h", "q"}) && len(blockLinesWithPrefix(block, ui.prefix(uiMessagePrompt)+" ")) == 0 {
			accountSelectionBlocks++
		}
	}

	assert.Equal(t, 2, accountSelectionBlocks)
	assert.Equal(t, 2, countMenuBlocksWithKeys(output, []string{"1", "2", "3", "b", "h", "q"}))
	assert.Equal(t, 2, countMenuBlocksWithKeys(output, []string{"1", "2", "b", "h", "q"}))
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
		showInputErrorAndPause("sentinel-error")
	})

	assert.Equal(t, 1, strings.Count(output, ui.prefix(uiMessageError)+" "))
	assert.Len(t, linesWithPrefix(output, ui.prefix(uiMessagePrompt)+" "), 1)
	assert.Contains(t, output, "sentinel-error")
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

	assert.Equal(t, 1, strings.Count(output, ui.prefix(uiMessageError)+" "))
	assert.Len(t, linesWithPrefix(output, ui.prefix(uiMessagePrompt)+" "), 1)
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
		showInfoAndPause("sentinel-info")
	})

	assert.Len(t, linesWithPrefix(output, ui.prefix(uiMessageInfo)+" "), 1)
	assert.Len(t, linesWithPrefix(output, ui.prefix(uiMessagePrompt)+" "), 1)
	assert.Contains(t, output, "sentinel-info")
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
		showWarningAndPause("sentinel-warning")
	})

	assert.Len(t, linesWithPrefix(output, ui.prefix(uiMessageWarning)+" "), 1)
	assert.Len(t, linesWithPrefix(output, ui.prefix(uiMessagePrompt)+" "), 1)
	assert.Contains(t, output, "sentinel-warning")
	assert.Equal(t, 1, waitCalled)
}

func TestShowInputErrorAndPauseFallsBackToEnterWhenSingleKeyUnavailable(t *testing.T) {
	originalCanSingleKeyContinue := ui.canSingleKeyContinue
	originalWaitForAnyKey := ui.waitForAnyKey
	t.Cleanup(func() {
		ui.canSingleKeyContinue = originalCanSingleKeyContinue
		ui.waitForAnyKey = originalWaitForAnyKey
	})

	ui.canSingleKeyContinue = func() bool { return false }
	waitCalled := 0
	ui.waitForAnyKey = func() error {
		waitCalled++
		return nil
	}
	output := captureStdout(t, func() {
		withTestInput(t, "\n", func() {
			showInputErrorAndPause("sentinel-enter")
		})
	})

	assert.Equal(t, 1, strings.Count(output, ui.prefix(uiMessageError)+" "))
	assert.Len(t, linesWithPrefix(output, ui.prefix(uiMessagePrompt)+" "), 1)
	assert.Contains(t, output, "sentinel-enter")
	assert.Equal(t, 0, waitCalled)
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

	assert.Equal(t, 1, strings.Count(output, ui.prefix(uiMessageError)+" "))
	assert.Equal(t, 2, countMenuBlocksWithKeys(output, []string{"1", "2", "0", "b", "h", "q"}))
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

	assert.Equal(t, 1, strings.Count(output, ui.prefix(uiMessageError)+" "))
	assert.Equal(t, 2, countMenuBlocksWithKeys(output, []string{"b", "h", "q"}))
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

	assert.Equal(t, 1, strings.Count(output, ui.prefix(uiMessageError)+" "))
	assert.Equal(t, 2, countMenuBlocksWithKeys(output, []string{"1", "2", "b", "h", "q"}))
}

func TestPromptLaunchRegionKeepsCurrentMenuAfterInvalidInput(t *testing.T) {
	originalCanSingleKeyContinue := ui.canSingleKeyContinue
	t.Cleanup(func() {
		ui.canSingleKeyContinue = originalCanSingleKeyContinue
	})
	ui.canSingleKeyContinue = func() bool { return false }

	output := captureStdout(t, func() {
		withTestInput(t, "x\n\nb\n", func() {
			choice, ok := promptLaunchRegion("啟動指定帳號：選擇區域", []*account.Account{
				{DisplayName: "Alpha", Email: "alpha@example.com"},
			}, nil, nil)
			assert.False(t, ok)
			assert.Nil(t, choice.ManualRegion)
			assert.False(t, choice.UseDefaults)
		})
	})

	assert.Equal(t, 1, strings.Count(output, ui.prefix(uiMessageError)+" "))
	assert.Equal(t, 2, countMenuBlocksWithKeys(output, []string{"1", "2", "3", "b", "h", "q"}))
}

func TestPromptLaunchRegionShowsSingleTargetAccount(t *testing.T) {
	output := captureStdout(t, func() {
		withTestInput(t, "b\n", func() {
			choice, ok := promptLaunchRegion("啟動指定帳號：選擇區域", []*account.Account{
				{DisplayName: "Alpha", Email: "alpha@example.com", GraphicsProfile: "1080p-low", DefaultRegion: "EU", DefaultMod: mods.DefaultModVanilla},
			}, []string{"sample-mod"}, []string{"1080p-low"})
			assert.False(t, ok)
			assert.Nil(t, choice.ManualRegion)
			assert.False(t, choice.UseDefaults)
		})
	})

	assert.Contains(t, output, "Alpha")
	assert.Contains(t, output, "alpha@example.com")
	assert.Contains(t, output, "EU")
	assert.Contains(t, output, lang.DefaultMods.StatusVanilla)
	assert.Contains(t, output, "1080p-low")
	assert.Equal(t, 1, countMenuBlocksWithKeys(output, []string{"1", "2", "3", "b", "h", "q"}))
}

func TestPromptLaunchRegionShowsBatchTargetAccounts(t *testing.T) {
	output := captureStdout(t, func() {
		withTestInput(t, "b\n", func() {
			choice, ok := promptLaunchRegion("啟動所有帳號：選擇區域", []*account.Account{
				{DisplayName: "Alpha", Email: "alpha@example.com", GraphicsProfile: "1080p-low", DefaultRegion: "NA", DefaultMod: mods.DefaultModVanilla},
				{DisplayName: "Bravo", Email: "bravo@example.com", GraphicsProfile: "1440p-high", DefaultRegion: "EU", DefaultMod: "sample-mod"},
			}, []string{"sample-mod"}, []string{"1080p-low", "1440p-high"})
			assert.False(t, ok)
			assert.Nil(t, choice.ManualRegion)
			assert.False(t, choice.UseDefaults)
		})
	})

	assert.Contains(t, output, "Alpha")
	assert.Contains(t, output, "alpha@example.com")
	assert.Contains(t, output, "Bravo")
	assert.Contains(t, output, "bravo@example.com")
	assert.Contains(t, output, "NA")
	assert.Contains(t, output, "EU")
	assert.Contains(t, output, lang.DefaultMods.StatusVanilla)
	assert.Contains(t, output, "sample-mod")
	assert.Contains(t, output, "1080p-low")
	assert.Contains(t, output, "1440p-high")
	assert.Equal(t, 1, countMenuBlocksWithKeys(output, []string{"1", "2", "3", "b", "h", "q"}))
}

func TestPromptLaunchRegionEnterUsesStoredDefaultMode(t *testing.T) {
	withTestInput(t, "\n", func() {
		choice, ok := promptLaunchRegion("啟動指定帳號：選擇區域", []*account.Account{
			{DisplayName: "Alpha", Email: "alpha@example.com", DefaultRegion: "EU"},
		}, nil, nil)
		assert.True(t, ok)
		assert.True(t, choice.UseDefaults)
		assert.Nil(t, choice.ManualRegion)
	})
}

func TestPromptLaunchRegionEnterRequiresDefaultsForAllTargets(t *testing.T) {
	originalCanSingleKeyContinue := ui.canSingleKeyContinue
	t.Cleanup(func() {
		ui.canSingleKeyContinue = originalCanSingleKeyContinue
	})
	ui.canSingleKeyContinue = func() bool { return false }

	output := captureStdout(t, func() {
		withTestInput(t, "\n\nb\n", func() {
			choice, ok := promptLaunchRegion("啟動所有帳號：選擇區域", []*account.Account{
				{DisplayName: "Alpha", Email: "alpha@example.com", DefaultRegion: "NA"},
				{DisplayName: "Bravo", Email: "bravo@example.com"},
				{DisplayName: "Charlie", Email: "charlie@example.com"},
			}, nil, nil)
			assert.False(t, ok)
			assert.False(t, choice.UseDefaults)
			assert.Nil(t, choice.ManualRegion)
		})
	})

	assert.Equal(t, 1, strings.Count(output, ui.prefix(uiMessageError)+" "))
	assert.Contains(t, output, "- Bravo")
	assert.Contains(t, output, "- Charlie")
	assert.NotContains(t, output, "Bravo, Charlie")
	assert.Equal(t, 2, countMenuBlocksWithKeys(output, []string{"1", "2", "3", "b", "h", "q"}))
}

func TestPromptLaunchRegionAllowsManualOverrideWithoutStoredDefaults(t *testing.T) {
	withTestInput(t, "2\n", func() {
		choice, ok := promptLaunchRegion("啟動指定帳號：選擇區域", []*account.Account{
			{DisplayName: "Alpha", Email: "alpha@example.com"},
		}, nil, nil)
		assert.True(t, ok)
		if assert.NotNil(t, choice.ManualRegion) {
			assert.Equal(t, "EU", choice.ManualRegion.Name)
		}
		assert.False(t, choice.UseDefaults)
	})
}

func TestPromptLaunchModKeepsCurrentMenuAfterInvalidInput(t *testing.T) {
	originalCanSingleKeyContinue := ui.canSingleKeyContinue
	t.Cleanup(func() {
		ui.canSingleKeyContinue = originalCanSingleKeyContinue
	})
	ui.canSingleKeyContinue = func() bool { return false }

	accounts := []account.Account{{DisplayName: "Alpha", Email: "alpha@example.com"}}
	output := captureStdout(t, func() {
		withTestInput(t, "x\n\nb\n", func() {
			choice, ok := promptLaunchMod("啟動指定帳號：選擇 mod", accounts, filepath.Join(t.TempDir(), "accounts.csv"), []*account.Account{&accounts[0]}, []string{"sample-mod"}, nil)
			assert.False(t, ok)
			assert.False(t, choice.UseDefaults)
			assert.False(t, choice.HasManual)
		})
	})

	assert.Equal(t, 1, strings.Count(output, ui.prefix(uiMessageError)+" "))
	assert.Equal(t, 2, countMenuBlocksWithKeys(output, []string{"0", "1", "b", "h", "q"}))
}

func TestPromptLaunchModShowsBatchTargetAccounts(t *testing.T) {
	accounts := []account.Account{
		{DisplayName: "Alpha", Email: "alpha@example.com"},
		{DisplayName: "Bravo", Email: "bravo@example.com"},
	}

	output := captureStdout(t, func() {
		withTestInput(t, "b\n", func() {
			choice, ok := promptLaunchMod("啟動所有帳號：選擇 mod", accounts, filepath.Join(t.TempDir(), "accounts.csv"), []*account.Account{&accounts[0], &accounts[1]}, []string{"sample-mod"}, nil)
			assert.False(t, ok)
			assert.False(t, choice.UseDefaults)
			assert.False(t, choice.HasManual)
		})
	})

	assert.Contains(t, output, "Alpha")
	assert.Contains(t, output, "alpha@example.com")
	assert.Contains(t, output, "Bravo")
	assert.Contains(t, output, "bravo@example.com")
	assert.Contains(t, output, lang.RegionDefaults.StatusUnassigned)
	assert.Contains(t, output, lang.DefaultMods.StatusUnassigned)
	assert.Contains(t, output, lang.GraphicsProfiles.StatusUnassigned)
	assert.Equal(t, 1, countMenuBlocksWithKeys(output, []string{"0", "1", "b", "h", "q"}))
}

func TestPromptLaunchModShowsPreparedDefaultsForTargets(t *testing.T) {
	accounts := []account.Account{
		{DisplayName: "Alpha", Email: "alpha@example.com", GraphicsProfile: "1080p-low", DefaultRegion: "NA", DefaultMod: mods.DefaultModVanilla},
		{DisplayName: "Bravo", Email: "bravo@example.com", GraphicsProfile: "1440p-high", DefaultRegion: "EU", DefaultMod: "sample-mod"},
	}

	output := captureStdout(t, func() {
		withTestInput(t, "b\n", func() {
			choice, ok := promptLaunchMod("啟動所有帳號：選擇 mod", accounts, filepath.Join(t.TempDir(), "accounts.csv"), []*account.Account{&accounts[0], &accounts[1]}, []string{"sample-mod"}, []string{"1080p-low", "1440p-high"})
			assert.False(t, ok)
			assert.False(t, choice.UseDefaults)
			assert.False(t, choice.HasManual)
		})
	})

	assert.Contains(t, output, "NA")
	assert.Contains(t, output, "EU")
	assert.Contains(t, output, lang.DefaultMods.StatusVanilla)
	assert.Contains(t, output, "sample-mod")
	assert.Contains(t, output, "1080p-low")
	assert.Contains(t, output, "1440p-high")
}

func TestPromptLaunchModEnterUsesStoredDefaultMode(t *testing.T) {
	accounts := []account.Account{
		{DisplayName: "Alpha", Email: "alpha@example.com", DefaultMod: mods.DefaultModVanilla},
	}
	accountsFile := filepath.Join(t.TempDir(), "accounts.csv")
	assert.NoError(t, account.SaveAccounts(accountsFile, accounts))

	withTestInput(t, "\n", func() {
		choice, ok := promptLaunchMod("啟動指定帳號：選擇 mod", accounts, accountsFile, []*account.Account{&accounts[0]}, nil, nil)
		assert.True(t, ok)
		assert.True(t, choice.UseDefaults)
		assert.False(t, choice.HasManual)
	})
}

func TestPromptLaunchModEnterRequiresDefaultsForAllTargets(t *testing.T) {
	originalCanSingleKeyContinue := ui.canSingleKeyContinue
	t.Cleanup(func() {
		ui.canSingleKeyContinue = originalCanSingleKeyContinue
	})
	ui.canSingleKeyContinue = func() bool { return false }

	accounts := []account.Account{
		{DisplayName: "Alpha", Email: "alpha@example.com", DefaultMod: mods.DefaultModVanilla},
		{DisplayName: "Bravo", Email: "bravo@example.com"},
		{DisplayName: "Charlie", Email: "charlie@example.com"},
	}
	accountsFile := filepath.Join(t.TempDir(), "accounts.csv")
	assert.NoError(t, account.SaveAccounts(accountsFile, accounts))

	output := captureStdout(t, func() {
		withTestInput(t, "\n\nb\n", func() {
			choice, ok := promptLaunchMod("啟動所有帳號：選擇 mod", accounts, accountsFile, []*account.Account{&accounts[0], &accounts[1], &accounts[2]}, []string{"sample-mod"}, nil)
			assert.False(t, ok)
			assert.False(t, choice.UseDefaults)
			assert.False(t, choice.HasManual)
		})
	})

	assert.Equal(t, 1, strings.Count(output, ui.prefix(uiMessageError)+" "))
	assert.Contains(t, output, "- Bravo")
	assert.Contains(t, output, "- Charlie")
	assert.NotContains(t, output, "Bravo, Charlie")
	assert.Equal(t, 2, countMenuBlocksWithKeys(output, []string{"0", "1", "b", "h", "q"}))
}

func TestPromptLaunchModEnterClearsMissingInstalledDefault(t *testing.T) {
	originalCanSingleKeyContinue := ui.canSingleKeyContinue
	t.Cleanup(func() {
		ui.canSingleKeyContinue = originalCanSingleKeyContinue
	})
	ui.canSingleKeyContinue = func() bool { return false }

	accounts := []account.Account{
		{DisplayName: "Alpha", Email: "alpha@example.com", DefaultMod: "ghost-mod"},
	}
	accountsFile := filepath.Join(t.TempDir(), "accounts.csv")
	assert.NoError(t, account.SaveAccounts(accountsFile, accounts))

	output := captureStdout(t, func() {
		withTestInput(t, "\n\nb\n", func() {
			choice, ok := promptLaunchMod("啟動指定帳號：選擇 mod", accounts, accountsFile, []*account.Account{&accounts[0]}, []string{"sample-mod"}, nil)
			assert.False(t, ok)
			assert.False(t, choice.UseDefaults)
			assert.False(t, choice.HasManual)
		})
	})

	assert.Equal(t, "", accounts[0].DefaultMod)
	reloaded, err := account.LoadAccounts(accountsFile)
	assert.NoError(t, err)
	assert.Equal(t, "", reloaded[0].DefaultMod)
	assert.Equal(t, 1, strings.Count(output, ui.prefix(uiMessageError)+" "))
}

func TestPromptLaunchModAllowsManualOverrideWithoutStoredDefaults(t *testing.T) {
	accounts := []account.Account{
		{DisplayName: "Alpha", Email: "alpha@example.com"},
	}
	withTestInput(t, "0\n", func() {
		choice, ok := promptLaunchMod("啟動指定帳號：選擇 mod", accounts, filepath.Join(t.TempDir(), "accounts.csv"), []*account.Account{&accounts[0]}, []string{"sample-mod"}, nil)
		assert.True(t, ok)
		assert.True(t, choice.HasManual)
		assert.Equal(t, mods.DefaultModVanilla, choice.ManualMod)
		assert.False(t, choice.UseDefaults)
	})
}

func TestPromptLaunchGraphicsKeepsCurrentMenuAfterInvalidInput(t *testing.T) {
	originalCanSingleKeyContinue := ui.canSingleKeyContinue
	t.Cleanup(func() {
		ui.canSingleKeyContinue = originalCanSingleKeyContinue
	})
	ui.canSingleKeyContinue = func() bool { return false }

	accounts := []account.Account{{DisplayName: "Alpha", Email: "alpha@example.com"}}
	output := captureStdout(t, func() {
		withTestInput(t, "x\n\nb\n", func() {
			choice, ok := promptLaunchGraphics("啟動指定帳號：選擇畫質", accounts, filepath.Join(t.TempDir(), "accounts.csv"), []*account.Account{&accounts[0]}, nil, []string{"boss-low"})
			assert.False(t, ok)
			assert.False(t, choice.UseDefaults)
			assert.False(t, choice.HasManual)
		})
	})

	assert.Equal(t, 1, strings.Count(output, ui.prefix(uiMessageError)+" "))
	assert.Equal(t, 2, countMenuBlocksWithKeys(output, []string{"0", "1", "b", "h", "q"}))
}

func TestPromptLaunchGraphicsShowsPreparedDefaultsForTargets(t *testing.T) {
	accounts := []account.Account{
		{DisplayName: "Alpha", Email: "alpha@example.com", GraphicsProfile: "1080p-low", DefaultRegion: "NA", DefaultMod: mods.DefaultModVanilla},
		{DisplayName: "Bravo", Email: "bravo@example.com", GraphicsProfile: "missing-profile", DefaultRegion: "EU", DefaultMod: "sample-mod"},
	}

	output := captureStdout(t, func() {
		withTestInput(t, "b\n", func() {
			choice, ok := promptLaunchGraphics("啟動所有帳號：選擇畫質", accounts, filepath.Join(t.TempDir(), "accounts.csv"), []*account.Account{&accounts[0], &accounts[1]}, []string{"sample-mod"}, []string{"1080p-low"})
			assert.False(t, ok)
			assert.False(t, choice.UseDefaults)
			assert.False(t, choice.HasManual)
		})
	})

	assert.Contains(t, output, "NA")
	assert.Contains(t, output, "EU")
	assert.Contains(t, output, lang.DefaultMods.StatusVanilla)
	assert.Contains(t, output, "sample-mod")
	assert.Contains(t, output, "1080p-low")
	assert.Contains(t, output, fmt.Sprintf(lang.GraphicsProfiles.StatusMissing, "missing-profile"))
	assert.Equal(t, 1, countMenuBlocksWithKeys(output, []string{"0", "1", "b", "h", "q"}))
}

func TestPromptLaunchGraphicsEnterUsesStoredDefaults(t *testing.T) {
	accounts := []account.Account{
		{DisplayName: "Alpha", Email: "alpha@example.com", GraphicsProfile: "boss-low"},
	}

	withTestInput(t, "\n", func() {
		choice, ok := promptLaunchGraphics("啟動指定帳號：選擇畫質", accounts, filepath.Join(t.TempDir(), "accounts.csv"), []*account.Account{&accounts[0]}, nil, []string{"boss-low"})
		assert.True(t, ok)
		assert.True(t, choice.UseDefaults)
		assert.False(t, choice.HasManual)
	})
}

func TestPromptLaunchGraphicsEnterRequiresDefaultsForAllTargets(t *testing.T) {
	originalCanSingleKeyContinue := ui.canSingleKeyContinue
	t.Cleanup(func() {
		ui.canSingleKeyContinue = originalCanSingleKeyContinue
	})
	ui.canSingleKeyContinue = func() bool { return false }

	accounts := []account.Account{
		{DisplayName: "Alpha", Email: "alpha@example.com", GraphicsProfile: "boss-low"},
		{DisplayName: "Bravo", Email: "bravo@example.com"},
		{DisplayName: "Charlie", Email: "charlie@example.com"},
	}

	output := captureStdout(t, func() {
		withTestInput(t, "\n\nb\n", func() {
			choice, ok := promptLaunchGraphics("啟動所有帳號：選擇畫質", accounts, filepath.Join(t.TempDir(), "accounts.csv"), []*account.Account{&accounts[0], &accounts[1], &accounts[2]}, nil, []string{"boss-low"})
			assert.False(t, ok)
			assert.False(t, choice.UseDefaults)
			assert.False(t, choice.HasManual)
		})
	})

	assert.Equal(t, 1, strings.Count(output, ui.prefix(uiMessageError)+" "))
	assert.Contains(t, output, "- Bravo")
	assert.Contains(t, output, "- Charlie")
	assert.NotContains(t, output, "Bravo, Charlie")
	assert.Equal(t, 2, countMenuBlocksWithKeys(output, []string{"0", "1", "b", "h", "q"}))
}

func TestPromptLaunchGraphicsEnterClearsMissingSavedDefault(t *testing.T) {
	accounts := []account.Account{
		{DisplayName: "Alpha", Email: "alpha@example.com", GraphicsProfile: "ghost-profile"},
	}
	accountsFile := filepath.Join(t.TempDir(), "accounts.csv")
	assert.NoError(t, account.SaveAccounts(accountsFile, accounts))

	output := captureStdout(t, func() {
		withTestInput(t, "\n\nb\n", func() {
			choice, ok := promptLaunchGraphics("啟動指定帳號：選擇畫質", accounts, accountsFile, []*account.Account{&accounts[0]}, nil, []string{"boss-low"})
			assert.False(t, ok)
			assert.False(t, choice.UseDefaults)
			assert.False(t, choice.HasManual)
		})
	})

	assert.Equal(t, "", accounts[0].GraphicsProfile)
	reloaded, err := account.LoadAccounts(accountsFile)
	assert.NoError(t, err)
	assert.Equal(t, "", reloaded[0].GraphicsProfile)
	assert.Equal(t, 1, strings.Count(output, ui.prefix(uiMessageWarning)+" "))
	assert.Equal(t, 1, strings.Count(output, ui.prefix(uiMessageError)+" "))
}

func TestPromptLaunchGraphicsAllowsManualOverrideWithoutStoredDefaults(t *testing.T) {
	accounts := []account.Account{
		{DisplayName: "Alpha", Email: "alpha@example.com"},
	}

	withTestInput(t, "0\n", func() {
		choice, ok := promptLaunchGraphics("啟動指定帳號：選擇畫質", accounts, filepath.Join(t.TempDir(), "accounts.csv"), []*account.Account{&accounts[0]}, nil, []string{"boss-low"})
		assert.True(t, ok)
		assert.True(t, choice.HasManual)
		assert.Equal(t, "", choice.ManualProfile)
		assert.False(t, choice.UseDefaults)
	})
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
