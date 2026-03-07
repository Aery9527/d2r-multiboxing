package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"d2rhl/internal/common/config"
	"d2rhl/internal/multiboxing/account"
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

	fmt.Println("============================================")
	fmt.Printf("  d2r-hyper-launcher  %s\n", displayVersion(version))
	fmt.Println("============================================")
	fmt.Println()

	cfg, err := config.Load()
	if err != nil {
		fmt.Printf("  設定檔載入失敗：%v\n", err)
		return
	}
	cfgDir, _ := config.Dir()
	fmt.Printf("  資料目錄：%s\n", cfgDir)
	fmt.Printf("  D2R 路徑：%s\n", cfg.D2RPath)
	fmt.Printf("  啟動間隔：%d 秒\n", cfg.LaunchDelay)

	if cfg.Switcher != nil && cfg.Switcher.Enabled {
		if err := switcher.Start(cfg.Switcher); err != nil {
			fmt.Printf("  ⚠ 視窗切換啟動失敗：%v\n", err)
		} else {
			fmt.Printf("  ✔ 視窗切換已啟用：%s\n", switcher.FormatSwitcherDisplay(cfg.Switcher.Modifiers, cfg.Switcher.Key, cfg.Switcher.GamepadIndex))
		}
	}
	fmt.Println()

	accountsFile, err := config.AccountsPath()
	if err != nil {
		fmt.Printf("  無法取得帳號檔案路徑：%v\n", err)
		return
	}

	createdAccountsFile, err := account.EnsureAccountsFile(accountsFile)
	if err != nil {
		fmt.Printf("  建立帳號檔案失敗：%v\n", err)
		return
	}
	if createdAccountsFile {
		handleCreatedAccountsFile(cfgDir, accountsFile)
		return
	}

	accounts, err := account.LoadAccounts(accountsFile)
	if err != nil {
		fmt.Printf("  讀取帳號失敗：%v\n", err)
		return
	}

	changed, err := account.EncryptPlaintextPasswords(accountsFile, accounts)
	if err != nil {
		fmt.Printf("  密碼加密失敗：%v\n", err)
		return
	}
	if changed {
		fmt.Println("  ✔ 已加密明文密碼並回寫至 CSV")
	}

	monitor.StartHandleMonitor()

	scanner := bufio.NewScanner(os.Stdin)
	for {
		printMenu(accounts)
		fmt.Print("  > 請選擇：")
		if !scanner.Scan() {
			break
		}

		input := strings.TrimSpace(scanner.Text())
		switch strings.ToLower(input) {
		case menuQuit:
			switcher.Stop()
			fmt.Println("  再見！")
			return
		case "r":
			accounts, err = account.LoadAccounts(accountsFile)
			if err != nil {
				fmt.Printf("  讀取帳號失敗：%v\n", err)
			}
		case "0":
			launchOffline(cfg, scanner)
		case "a":
			launchAll(accounts, cfg, scanner)
		case "p":
			setupD2RPath(cfg)
		case "s":
			setupSwitcher(cfg, scanner)
		case "f":
			setupAccountLaunchFlags(accounts, accountsFile, scanner)
		default:
			id, err := strconv.Atoi(input)
			if err != nil || id < 1 || id > len(accounts) {
				showInvalidInputAndPause()
				continue
			}
			launchAccount(&accounts[id-1], cfg, scanner)
		}
	}
}
