package main

import (
	"bufio"
	"fmt"
	"math/rand"
	"time"

	"d2rhl/internal/common/config"
	"d2rhl/internal/common/d2r"
	"d2rhl/internal/common/process"
	"d2rhl/internal/multiboxing/account"
	"d2rhl/internal/multiboxing/launcher"
)

var launchDelayRandIntN = rand.Intn
var launchDelaySleep = time.Sleep

func launchAccount(acc *account.Account, cfg *config.Config, scanner *bufio.Scanner) {
	if !ensureLaunchReadyD2RPath(cfg, scanner) {
		return
	}
	if isAccountRunning(acc.DisplayName) {
		fmt.Printf("  ⏭ %s 已在執行中，請先切回既有視窗或改用其他帳號。\n", acc.DisplayName)
		fmt.Println()
		return
	}

	fmt.Println()
	fmt.Println("  選擇區域 (1=NA, 2=EU, 3=Asia)")
	printSubMenuNav()
	fmt.Print("  > 請選擇：")
	if !scanner.Scan() {
		return
	}
	input := scanner.Text()
	if nav := isMenuNav(input); nav != "" {
		return
	}
	region := parseRegionInput(input)
	if region == nil {
		showInputErrorAndPause("無效的區域選擇。")
		return
	}

	modArgs, ok := selectLaunchMod(cfg.D2RPath, scanner)
	if !ok {
		return
	}

	password, err := account.GetDecryptedPassword(acc)
	if err != nil {
		fmt.Printf("  密碼解密失敗：%v\n", err)
		return
	}

	fmt.Printf("  正在啟動 %s (%s)...\n", acc.DisplayName, region.Name)
	pid, err := launcher.LaunchD2R(cfg.D2RPath, acc.Email, password, region.Address, accountLaunchArgs(*acc, modArgs)...)
	if err != nil {
		fmt.Printf("  啟動失敗：%v\n", err)
		return
	}
	fmt.Printf("  ✔ D2R 已啟動 (PID: %d)\n", pid)

	time.Sleep(2 * time.Second)
	closed, err := launcher.CloseHandlesByName(pid, d2r.SingleInstanceEventName)
	if err != nil {
		fmt.Printf("  ⚠ 關閉 Handle 失敗：%v\n", err)
	} else if closed > 0 {
		fmt.Printf("  ✔ 已關閉 %d 個 Event Handle\n", closed)
	}

	renameLaunchedWindow(pid, acc.DisplayName)
	fmt.Println()
}

func launchAll(accounts []account.Account, cfg *config.Config, scanner *bufio.Scanner) {
	if !ensureLaunchReadyD2RPath(cfg, scanner) {
		return
	}

	runningTitles := runningAccountWindowTitles()
	pendingAccounts := pendingBatchAccounts(accounts, runningTitles)
	fmt.Println("  已預先掃描目前執行中的 D2R 視窗：")
	for _, line := range batchAccountStatusLines(accounts, runningTitles) {
		fmt.Println(line)
	}
	if len(pendingAccounts) == 0 {
		fmt.Println("  所有帳號都已在執行中。")
		fmt.Println()
		return
	}
	fmt.Printf("  本次只會啟動上面標示為 [未啟動] 的帳號，共 %d 個。\n", len(pendingAccounts))

	fmt.Println()
	fmt.Println("  選擇區域 (1=NA, 2=EU, 3=Asia)")
	printSubMenuNav()
	fmt.Print("  > 請選擇：")
	if !scanner.Scan() {
		return
	}
	input := scanner.Text()
	if nav := isMenuNav(input); nav != "" {
		return
	}
	region := parseRegionInput(input)
	if region == nil {
		showInputErrorAndPause("無效的區域選擇。")
		return
	}

	modArgs, ok := selectLaunchMod(cfg.D2RPath, scanner)
	if !ok {
		return
	}

	for i, acc := range pendingAccounts {
		password, err := account.GetDecryptedPassword(acc)
		if err != nil {
			fmt.Printf("  ⚠ 帳號 %s 密碼解密失敗：%v\n", acc.DisplayName, err)
			continue
		}

		fmt.Printf("  正在啟動 %s (%s)...\n", acc.DisplayName, region.Name)
		pid, err := launcher.LaunchD2R(cfg.D2RPath, acc.Email, password, region.Address, accountLaunchArgs(*acc, modArgs)...)
		if err != nil {
			fmt.Printf("  ⚠ 帳號 %s 啟動失敗：%v\n", acc.DisplayName, err)
			continue
		}
		fmt.Printf("  ✔ %s 已啟動 (PID: %d)\n", acc.DisplayName, pid)

		time.Sleep(3 * time.Second)
		closed, err := launcher.CloseHandlesByName(pid, d2r.SingleInstanceEventName)
		if err != nil {
			fmt.Printf("  ⚠ %s Handle 關閉失敗：%v\n", acc.DisplayName, err)
		} else if closed > 0 {
			fmt.Printf("  ✔ %s 已關閉 %d 個 Handle\n", acc.DisplayName, closed)
		}

		renameLaunchedWindow(pid, acc.DisplayName)

		if i+1 < len(pendingAccounts) {
			delaySeconds := cfg.LaunchDelay.NextSeconds(launchDelayRandIntN)
			waitForNextBatchLaunch(delaySeconds, pendingAccounts[i+1].DisplayName)
		}
	}
	fmt.Println()
}

func launchOffline(cfg *config.Config, scanner *bufio.Scanner) {
	if !ensureLaunchReadyD2RPath(cfg, scanner) {
		return
	}

	fmt.Println()
	fmt.Println("  === 離線遊玩模式 ===")

	modArgs, ok := selectLaunchMod(cfg.D2RPath, scanner)
	if !ok {
		return
	}

	fmt.Println("  正在啟動 D2R（離線模式）...")
	pid, err := launcher.LaunchD2ROffline(cfg.D2RPath, modArgs...)
	if err != nil {
		fmt.Printf("  啟動失敗：%v\n", err)
		return
	}
	fmt.Printf("  ✔ D2R 已啟動 (PID: %d)\n", pid)
	fmt.Println()
}

func accountLaunchArgs(acc account.Account, modArgs []string) []string {
	args := make([]string, 0, len(modArgs)+4)
	args = append(args, modArgs...)
	args = append(args, account.LaunchArgs(acc.LaunchFlags)...)
	return args
}

func renameLaunchedWindow(pid uint32, displayName string) {
	fmt.Printf("  正在準備重命名視窗：%s\n", displayName)
	err := process.RenameWindow(pid, d2r.WindowTitle(displayName), 15, 2*time.Second)
	if err != nil {
		fmt.Printf("  ⚠ 視窗重命名失敗 (%s)：%v\n", displayName, err)
		return
	}

	fmt.Printf("  ✔ 視窗已重命名為 \"%s\"\n", d2r.WindowTitle(displayName))
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
		lines = append(lines, fmt.Sprintf("  [%s] %s (%s)", status, accounts[i].DisplayName, accounts[i].Email))
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

	fmt.Println(formatLaunchDelayMessage(delaySeconds, nextDisplayName))
	remainingSeconds := delaySeconds
	for remainingSeconds > 0 {
		stepSeconds := 5
		if remainingSeconds < stepSeconds {
			stepSeconds = remainingSeconds
		}

		launchDelaySleep(time.Duration(stepSeconds) * time.Second)
		remainingSeconds -= stepSeconds
		if remainingSeconds > 0 {
			fmt.Println(formatLaunchDelayRemainingMessage(remainingSeconds, nextDisplayName))
		}
	}
}
