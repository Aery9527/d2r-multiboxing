package main

import (
	"fmt"
	"os"
	"strings"

	"d2rhl/internal/common/config"
	"d2rhl/internal/common/d2r"
	"d2rhl/internal/common/process"
	"d2rhl/internal/multiboxing/account"
	"d2rhl/internal/switcher"
)

const (
	menuBack = "b"
	menuHome = "h"
	menuQuit = "q"
)

func printStartupAnnouncement(cfgDir string, cfg *config.Config) {
	ui.headf("d2r-hyper-launcher (%s)", displayVersion(version))
	ui.infof("資料目錄：%s", cfgDir)
	ui.infof("D2R 路徑：%s", cfg.D2RPath)
	ui.warningLines(
		"帳號啟動狀態的偵測是用 account.csv 裡的 DisplayName 去對應視窗名稱，",
		"所以已經透過該工具開啟 D2R 然後又去修改 DisplayName的話，",
		"就會導致啟動狀態顯示不正確，請注意。",
	)
}

func printMenu(accounts []account.Account, cfg *config.Config) {
	ui.headf("主選單")
	ui.infof("帳號列表：")
	for i, acc := range accounts {
		status := "未啟動"
		if process.FindWindowByTitle(d2r.WindowTitle(acc.DisplayName)) {
			status = "已啟動"
		}
		ui.rawlnf("[%d] <%s> %-15s (%s)  ", i+1, status, acc.DisplayName, acc.Email)
	}

	ui.blankLine()
	options := ui.newMenuOptions()
	options.option("數字", "啟動指定帳號")
	options.option("0", "離線遊玩（可選 mod，不需帳密）")
	options.option("a", "啟動所有帳號（可選 mod，只啟動未啟動的）")
	options.option("d", fmt.Sprintf("設定啟動間隔（目前：%s）", cfg.LaunchDelay.DisplayString()))
	options.option("f", "設定帳號啟動 flag")
	options.option("p", "選擇 D2R.exe 路徑")
	options.option("s", switcherMenuOptionLabel(cfg))
	options.option("r", "重新整理狀態")
	options.blankLine()
	options.option(menuQuit, "退出")
	ui.menuBlock(func() {
		options.render(ui)
	})
}

func printSubMenuNav() {
	ui.subMenuNav()
}

func switcherMenuOptionLabel(cfg *config.Config) string {
	if cfg.Switcher == nil || !cfg.Switcher.Enabled {
		return "視窗切換設定（目前：未啟用）"
	}
	return fmt.Sprintf(
		"視窗切換設定（目前：%s）",
		switcher.FormatSwitcherDisplay(cfg.Switcher.Modifiers, cfg.Switcher.Key, cfg.Switcher.GamepadIndex),
	)
}

func isMenuNav(input string) string {
	switch strings.ToLower(strings.TrimSpace(input)) {
	case menuBack:
		return "back"
	case menuHome:
		return "home"
	case menuQuit:
		ui.infof("再見！")
		os.Exit(0)
		return ""
	default:
		return ""
	}
}
