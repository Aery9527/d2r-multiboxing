package main

import (
	"fmt"
	"strconv"
	"strings"

	"d2rhl/internal/common/config"
	"d2rhl/internal/multiboxing/account"
	"d2rhl/internal/multiboxing/launcher"
	"d2rhl/internal/multiboxing/monitor"
	"d2rhl/internal/switcher"

	"golang.org/x/sys/windows"
)

// version and releaseTime are set at build time via
// -ldflags "-X main.version=x.y.z -X main.releaseTime=yyyy-mm-dd hh:mm:ss".
var version = "dev"
var releaseTime = ""

func displayVersion(version string) string {
	if strings.HasPrefix(version, "v") {
		return version
	}
	return "v" + version
}

func displayReleaseTime(releaseTime string) string {
	releaseTime = strings.TrimSpace(releaseTime)
	if releaseTime == "" {
		return "尚未 release"
	}
	return fmt.Sprintf("%s release", releaseTime)
}

func displayReleaseSummary(version string, releaseTime string) string {
	return fmt.Sprintf("%s（%s）", displayVersion(version), displayReleaseTime(releaseTime))
}

func maybeShowStartupAnnouncement(cfgDir string, createdAccountsFile bool) {
	if createdAccountsFile {
		return
	}
	printStartupAnnouncement(cfgDir)
	pauseAfterStartupAnnouncement()
}

func main() {
	_ = windows.SetConsoleCP(65001)
	_ = windows.SetConsoleOutputCP(65001)

	launcher.SetCommandLogger(func(message string) {
		ui.commandf("%s", message)
	})

	cfg, err := config.Load()
	if err != nil {
		showInputErrorAndPause(fmt.Sprintf("設定檔載入失敗：%v", err))
		return
	}
	cfgDir, _ := config.Dir()
	printStartupHeader()

	accountsFile, err := config.AccountsPath()
	if err != nil {
		showInputErrorAndPause(fmt.Sprintf("無法取得帳號檔案路徑：%v", err))
		return
	}

	createdAccountsFile, err := account.EnsureAccountsFile(accountsFile)
	if err != nil {
		showInputErrorAndPause(fmt.Sprintf("建立帳號檔案失敗：%v", err))
		return
	}
	if createdAccountsFile {
		handleCreatedAccountsFile(cfgDir, accountsFile)
		return
	}

	maybeShowStartupAnnouncement(cfgDir, createdAccountsFile)

	if cfg.Switcher != nil && cfg.Switcher.Enabled {
		if err := switcher.Start(cfg.Switcher); err != nil {
			ui.warningf("視窗切換啟動失敗：%v", err)
		}
		ui.blankLine()
	}

	accounts, err := account.LoadAccounts(accountsFile)
	if err != nil {
		showInputErrorAndPause(fmt.Sprintf("讀取帳號失敗：%v", err))
		return
	}

	changed, err := account.EncryptPlaintextPasswords(accountsFile, accounts)
	if err != nil {
		showInputErrorAndPause(fmt.Sprintf("密碼加密失敗：%v", err))
		return
	}
	if changed {
		ui.successf("已加密明文密碼並回寫至 CSV")
	}

	monitor.StartHandleMonitor()

	for {
		printMenu(accounts, cfg)
		input, ok := ui.readInput()
		if !ok {
			break
		}

		switch strings.ToLower(input) {
		case menuQuit:
			switcher.Stop()
			ui.infof("再見！")
			return
		case "r":
			accounts, err = account.LoadAccounts(accountsFile)
			if err != nil {
				ui.errorf("讀取帳號失敗：%v", err)
			}
		case "0":
			launchOffline(cfg)
		case "a":
			launchAll(accounts, cfg)
		case "d":
			setupLaunchDelay(cfg)
		case "p":
			setupD2RPath(cfg)
		case "s":
			setupSwitcher(cfg)
		case "f":
			setupAccountLaunchFlags(accounts, accountsFile)
		default:
			id, err := strconv.Atoi(input)
			if err != nil || id < 1 || id > len(accounts) {
				showInvalidInputAndPause()
				continue
			}
			launchAccount(&accounts[id-1], cfg)
		}
	}
}
