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
		ui.warningf("%s 已在執行中，請先切回既有視窗或改用其他帳號。", acc.DisplayName)
		ui.blankLine()
		return
	}

	region, ok := promptLaunchRegion("啟動指定帳號：選擇區域")
	if !ok {
		return
	}

	modArgs, ok := selectLaunchMod(cfg.D2RPath)
	if !ok {
		return
	}

	password, err := account.GetDecryptedPassword(acc)
	if err != nil {
		ui.errorf("密碼解密失敗：%v", err)
		return
	}

	ui.infof("正在啟動 %s (%s)...", acc.DisplayName, region.Name)
	pid, err := launcher.LaunchD2R(cfg.D2RPath, acc.Email, password, region.Address, accountLaunchArgs(*acc, modArgs)...)
	if err != nil {
		ui.errorf("啟動失敗：%v", err)
		return
	}
	ui.successf("D2R 已啟動 (PID: %d)", pid)

	pauseAfterSuccessfulLaunch()
	closed, err := launcher.CloseHandlesByName(pid, d2r.SingleInstanceEventName)
	if err != nil {
		ui.warningf("關閉 Handle 失敗：%v", err)
	} else if closed > 0 {
		ui.successf("已關閉 %d 個 Event Handle", closed)
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
	ui.infof("已預先掃描目前執行中的 D2R 視窗：")
	for _, line := range batchAccountStatusLines(accounts, runningTitles) {
		ui.rawln(line)
	}
	if len(pendingAccounts) == 0 {
		ui.infof("所有帳號都已在執行中。")
		ui.blankLine()
		return
	}
	ui.infof("本次只會啟動上面標示為 <未啟動> 的帳號，共 %d 個。", len(pendingAccounts))

	region, ok := promptLaunchRegion("啟動所有帳號：選擇區域")
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
			ui.warningf("帳號 %s 密碼解密失敗：%v", acc.DisplayName, err)
			continue
		}

		ui.infof("正在啟動 %s (%s)...", acc.DisplayName, region.Name)
		pid, err := launcher.LaunchD2R(cfg.D2RPath, acc.Email, password, region.Address, accountLaunchArgs(*acc, modArgs)...)
		if err != nil {
			ui.warningf("帳號 %s 啟動失敗：%v", acc.DisplayName, err)
			continue
		}
		ui.successf("%s 已啟動 (PID: %d)", acc.DisplayName, pid)

		pauseAfterSuccessfulLaunch()
		closed, err := launcher.CloseHandlesByName(pid, d2r.SingleInstanceEventName)
		if err != nil {
			ui.warningf("%s Handle 關閉失敗：%v", acc.DisplayName, err)
		} else if closed > 0 {
			ui.successf("%s 已關閉 %d 個 Handle", acc.DisplayName, closed)
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

	ui.blankLine()
	ui.headf("離線遊玩模式")

	modArgs, ok := selectLaunchMod(cfg.D2RPath)
	if !ok {
		return
	}

	ui.infof("正在啟動 D2R（離線模式）...")
	pid, err := launcher.LaunchD2ROffline(cfg.D2RPath, modArgs...)
	if err != nil {
		ui.errorf("啟動失敗：%v", err)
		return
	}
	ui.successf("D2R 已啟動 (PID: %d)", pid)
	pauseAfterSuccessfulLaunch()
	ui.blankLine()
}

func promptLaunchRegion(title string) (*d2r.Region, bool) {
	for {
		ui.blankLine()
		ui.headf("%s", title)
		options := ui.subMenuOptions(func(options *cliMenuOptions) {
			options.option("1", "NA", "")
			options.option("2", "EU", "")
			options.option("3", "Asia", "")
		})
		ui.menuBlock(func() {
			options.render()
		})
		input, ok := ui.readInput()
		if !ok {
			return nil, false
		}
		if nav := isMenuNav(input); nav != "" {
			return nil, false
		}
		region := parseRegionInput(input)
		if region == nil {
			showInputErrorAndPause("無效的區域選擇。")
			continue
		}
		return region, true
	}
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
	ui.infof("正在準備重命名視窗：%s", displayName)
	err := process.RenameWindow(pid, d2r.WindowTitle(displayName), 15, 2*time.Second)
	if err != nil {
		ui.warningf("視窗重命名失敗 (%s)：%v", displayName, err)
		return
	}

	ui.successf("視窗已重命名為 %q", d2r.WindowTitle(displayName))
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
		status := "未啟動"
		if runningTitles[d2r.WindowTitle(accounts[i].DisplayName)] {
			status = "已啟動"
		}
		lines = append(lines, fmt.Sprintf("  <%s> %s (%s)", status, accounts[i].DisplayName, accounts[i].Email))
	}
	return lines
}

func formatLaunchDelayMessage(delaySeconds int, nextDisplayName string) string {
	return fmt.Sprintf("  等待 %d 秒後啟動下一個帳號：%s", delaySeconds, nextDisplayName)
}

func formatLaunchDelayRemainingMessage(remainingSeconds int, nextDisplayName string) string {
	return fmt.Sprintf("  還剩 %d 秒後啟動下一個帳號：%s", remainingSeconds, nextDisplayName)
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
