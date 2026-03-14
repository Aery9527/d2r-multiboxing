package main

import (
	"fmt"
	"math/rand"
	"strings"
	"time"

	"d2rhl/internal/common/config"
	"d2rhl/internal/common/d2r"
	"d2rhl/internal/common/process"
	"d2rhl/internal/multiboxing/account"
	"d2rhl/internal/multiboxing/launcher"
)

var launchDelayRandIntN = rand.Intn
var launchDelaySleep = time.Sleep
var launchSuccessPauseSleep = time.Sleep

const launchSuccessPauseDuration = 3 * time.Second

func launchAccount(acc *account.Account, cfg *config.Config) {
	if !ensureLaunchReadyD2RPath(cfg) {
		return
	}
	if isAccountRunning(acc.DisplayName) {
		showWarningAndPause(fmt.Sprintf(lang.Launch.AlreadyRunning, acc.DisplayName))
		return
	}

	region, ok := promptLaunchRegion(lang.Launch.RegionSingleTitle, []*account.Account{acc})
	if !ok {
		return
	}

	modArgs, ok := selectLaunchMod(cfg.D2RPath)
	if !ok {
		return
	}

	password, err := account.GetDecryptedPassword(acc)
	if err != nil {
		ui.errorf(lang.Launch.DecryptFailed, err)
		return
	}

	ui.infof(lang.Launch.Starting, acc.DisplayName, region.Name)
	pid, err := launcher.LaunchD2R(cfg.D2RPath, acc.Email, password, region.Address, accountLaunchArgs(*acc, modArgs)...)
	if err != nil {
		ui.errorf(lang.Launch.LaunchFailed, err)
		return
	}
	ui.successf(lang.Launch.LaunchOK, pid)

	pauseAfterSuccessfulLaunch()
	closed, err := launcher.CloseHandlesByName(pid, d2r.SingleInstanceEventName)
	if err != nil {
		ui.warningf(lang.Launch.CloseHandleFailed, err)
	} else if closed > 0 {
		ui.successf(lang.Launch.HandlesClosed, closed)
	}

	renameLaunchedWindow(pid, acc.DisplayName)
	ui.blankLine()
}

func launchAll(accounts []account.Account, cfg *config.Config) {
	if !ensureLaunchReadyD2RPath(cfg) {
		return
	}

	runningTitles := runningAccountWindowTitles()
	pendingAccounts := pendingBatchAccounts(accounts, runningTitles)
	ui.infof("%s", lang.Launch.BatchScanHeader)
	for _, line := range batchAccountStatusLines(accounts, runningTitles) {
		ui.rawln(line)
	}
	if len(pendingAccounts) == 0 {
		showInfoAndPause(lang.Launch.AllRunning)
		return
	}
	ui.infof(lang.Launch.BatchOnlyPending, len(pendingAccounts))

	region, ok := promptLaunchRegion(lang.Launch.RegionBatchTitle, pendingAccounts)
	if !ok {
		return
	}

	modArgs, ok := selectLaunchMod(cfg.D2RPath)
	if !ok {
		return
	}

	for i, acc := range pendingAccounts {
		password, err := account.GetDecryptedPassword(acc)
		if err != nil {
			ui.warningf(lang.Launch.BatchDecryptFailed, acc.DisplayName, err)
			continue
		}

		ui.infof(lang.Launch.Starting, acc.DisplayName, region.Name)
		pid, err := launcher.LaunchD2R(cfg.D2RPath, acc.Email, password, region.Address, accountLaunchArgs(*acc, modArgs)...)
		if err != nil {
			ui.warningf(lang.Launch.BatchLaunchFailed, acc.DisplayName, err)
			continue
		}
		ui.successf(lang.Launch.BatchLaunchOK, acc.DisplayName, pid)

		pauseAfterSuccessfulLaunch()
		closed, err := launcher.CloseHandlesByName(pid, d2r.SingleInstanceEventName)
		if err != nil {
			ui.warningf(lang.Launch.BatchHandleCloseFailed, acc.DisplayName, err)
		} else if closed > 0 {
			ui.successf(lang.Launch.BatchHandlesClosed, acc.DisplayName, closed)
		}

		renameLaunchedWindow(pid, acc.DisplayName)

		if i+1 < len(pendingAccounts) {
			delaySeconds := cfg.LaunchDelay.NextSeconds(launchDelayRandIntN)
			waitForNextBatchLaunch(delaySeconds, pendingAccounts[i+1].DisplayName)
		}
	}
	ui.blankLine()
}

func launchOffline(cfg *config.Config) {
	if !ensureLaunchReadyD2RPath(cfg) {
		return
	}

	ui.headf("%s", lang.Launch.OfflineTitle)

	modArgs, ok := selectLaunchMod(cfg.D2RPath)
	if !ok {
		return
	}

	ui.infof("%s", lang.Launch.OfflineLaunching)
	pid, err := launcher.LaunchD2ROffline(cfg.D2RPath, modArgs...)
	if err != nil {
		ui.errorf(lang.Launch.OfflineLaunchFailed, err)
		return
	}
	ui.successf(lang.Launch.OfflineLaunchOK, pid)
	pauseAfterSuccessfulLaunch()
	ui.blankLine()
}

func promptLaunchRegion(title string, accounts []*account.Account) (*d2r.Region, bool) {
	var result *d2r.Region
	_ = runMenu(func() {
		ui.headf("%s", title)
		ui.infof("%s", lang.Launch.RegionTargetLabel)
		for _, line := range launchTargetAccountLines(accounts) {
			ui.rawln(line)
		}
		options := ui.subMenuOptions(func(options *cliMenuOptions) {
			options.option("1", "NA", "")
			options.option("2", "EU", "")
			options.option("3", "Asia", "")
		})
		ui.menuBlock(func() {
			options.render()
		})
	}, func(input string) error {
		region := parseRegionInput(input)
		if region == nil {
			showInputErrorAndPause(lang.Launch.RegionInvalid)
			return nil
		}
		result = region
		return errNavDone
	})
	return result, result != nil
}

func launchTargetAccountLines(accounts []*account.Account) []string {
	lines := make([]string, 0, len(accounts))
	for _, acc := range accounts {
		if acc == nil {
			continue
		}
		lines = append(lines, fmt.Sprintf("  %s (%s)", acc.DisplayName, acc.Email))
	}
	return lines
}

func pauseAfterSuccessfulLaunch() {
	launchSuccessPauseSleep(launchSuccessPauseDuration)
}

func accountLaunchArgs(acc account.Account, modArgs []string) []string {
	args := make([]string, 0, len(modArgs)+4)
	args = append(args, modArgs...)
	args = append(args, account.LaunchArgs(acc.LaunchFlags)...)
	return args
}

func renameLaunchedWindow(pid uint32, displayName string) {
	ui.infof(lang.Launch.WindowRenaming, displayName)
	err := process.RenameWindow(pid, d2r.WindowTitle(displayName), 15, 2*time.Second)
	if err != nil {
		ui.warningf(lang.Launch.WindowRenameFailed, displayName, err)
		return
	}

	ui.successf(lang.Launch.WindowRenamed, d2r.WindowTitle(displayName))
}

func runningAccountWindowTitles() map[string]bool {
	titles := process.FindWindowTitlesByPrefix(d2r.WindowTitlePrefix)
	running := make(map[string]bool, len(titles))
	for _, title := range titles {
		running[title] = true
	}
	return running
}

func isAccountRunning(displayName string) bool {
	return process.FindWindowByTitle(d2r.WindowTitle(displayName))
}

func pendingBatchAccounts(accounts []account.Account, runningTitles map[string]bool) []*account.Account {
	pending := make([]*account.Account, 0, len(accounts))
	for i := range accounts {
		if runningTitles[d2r.WindowTitle(accounts[i].DisplayName)] {
			continue
		}
		pending = append(pending, &accounts[i])
	}
	return pending
}

func runningBatchAccounts(accounts []account.Account, runningTitles map[string]bool) []*account.Account {
	running := make([]*account.Account, 0, len(accounts))
	for i := range accounts {
		if !runningTitles[d2r.WindowTitle(accounts[i].DisplayName)] {
			continue
		}
		running = append(running, &accounts[i])
	}
	return running
}

func batchAccountStatusLines(accounts []account.Account, runningTitles map[string]bool) []string {
	lines := make([]string, 0, len(accounts))
	for i := range accounts {
		status := lang.Launch.StatusStopped
		if runningTitles[d2r.WindowTitle(accounts[i].DisplayName)] {
			status = lang.Launch.StatusRunning
		}
		lines = append(lines, fmt.Sprintf("  <%s> %s (%s)", status, accounts[i].DisplayName, accounts[i].Email))
	}
	return lines
}

func formatLaunchDelayMessage(delaySeconds int, nextDisplayName string) string {
	return fmt.Sprintf("  "+lang.Launch.BatchDelayMsg, delaySeconds, nextDisplayName)
}

func formatLaunchDelayRemainingMessage(remainingSeconds int, nextDisplayName string) string {
	return fmt.Sprintf("  "+lang.Launch.BatchDelayRemaining, remainingSeconds, nextDisplayName)
}

func waitForNextBatchLaunch(delaySeconds int, nextDisplayName string) {
	if delaySeconds <= 0 {
		return
	}

	ui.infof("%s", strings.TrimSpace(formatLaunchDelayMessage(delaySeconds, nextDisplayName)))
	remainingSeconds := delaySeconds
	for remainingSeconds > 0 {
		stepSeconds := 5
		if remainingSeconds < stepSeconds {
			stepSeconds = remainingSeconds
		}

		launchDelaySleep(time.Duration(stepSeconds) * time.Second)
		remainingSeconds -= stepSeconds
		if remainingSeconds > 0 {
			ui.infof("%s", strings.TrimSpace(formatLaunchDelayRemainingMessage(remainingSeconds, nextDisplayName)))
		}
	}
}
