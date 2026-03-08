package main

import (
	"strconv"
	"strings"

	"d2rhl/internal/common/config"
	"d2rhl/internal/multiboxing/account"
	"d2rhl/internal/multiboxing/launcher"
	"d2rhl/internal/multiboxing/monitor"
	"d2rhl/internal/switcher"

	"golang.org/x/sys/windows"
)

// version is set at build time via -ldflags "-X main.version=x.y.z".
var version = "dev"

func displayVersion(version string) string {
	if strings.HasPrefix(version, "v") {
		return version
	}
	return "v" + version
}

func main() {
	_ = windows.SetConsoleCP(65001)
	_ = windows.SetConsoleOutputCP(65001)

	launcher.SetCommandLogger(func(message string) {
		ui.commandf("%s", message)
	})

	cfg, err := config.Load()
	if err != nil {
		ui.errorf("設定檔載入失敗：%v", err)
		return
	}
	cfgDir, _ := config.Dir()
	printStartupAnnouncement(cfgDir)
	pauseAfterStartupAnnouncement()

	if cfg.Switcher != nil && cfg.Switcher.Enabled {
		if err := switcher.Start(cfg.Switcher); err != nil {
			ui.warningf("視窗切換啟動失敗：%v", err)
		}
	}
	ui.blankLine()

	accountsFile, err := config.AccountsPath()
	if err != nil {
		ui.errorf("無法取得帳號檔案路徑：%v", err)
		return
	}

	createdAccountsFile, err := account.EnsureAccountsFile(accountsFile)
	if err != nil {
		ui.errorf("建立帳號檔案失敗：%v", err)
		return
	}
	if createdAccountsFile {
		handleCreatedAccountsFile(cfgDir, accountsFile)
		return
	}

	accounts, err := account.LoadAccounts(accountsFile)
	if err != nil {
		ui.errorf("讀取帳號失敗：%v", err)
		return
	}

	changed, err := account.EncryptPlaintextPasswords(accountsFile, accounts)
	if err != nil {
		ui.errorf("密碼加密失敗：%v", err)
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
