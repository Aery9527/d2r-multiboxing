package main

import (
	"os"
	"strings"

	"d2rhl/internal/common/d2r"
	"d2rhl/internal/common/process"
	"d2rhl/internal/multiboxing/account"
)

const (
	menuBack = "b"
	menuHome = "h"
	menuQuit = "q"
)

func printMenu(accounts []account.Account) {
	ui.headf("主選單")
	ui.infof("帳號列表：")
	for i, acc := range accounts {
		status := "未啟動"
		if process.FindWindowByTitle(d2r.WindowTitle(acc.DisplayName)) {
			status = "已啟動"
		}
		ui.rawlnf("  [%d] %-15s (%s)  [%s]", i+1, acc.DisplayName, acc.Email, status)
	}

	ui.blankLine()
	ui.infof("是否已啟動的判斷基準是用 account.csv 裡的 DisplayName 來對應視窗名稱。")
	ui.infof("如果 D2R 還開著就先關掉工具再去改 DisplayName，之後這裡的啟動狀態偵測可能會不正確。")
	ui.blankLine()
	ui.menuDividerLine()
	ui.option("數字", "啟動指定帳號")
	ui.option("0", "離線遊玩（可選 mod，不需帳密）")
	ui.option("a", "啟動所有帳號（可選 mod，只啟動未啟動的）")
	ui.option("d", "設定啟動間隔")
	ui.option("f", "設定帳號啟動 flag")
	ui.option("p", "選擇 D2R.exe 路徑")
	ui.option("s", "視窗切換設定")
	ui.option("r", "重新整理狀態")
	ui.blankLine()
	ui.option(menuQuit, "退出")
}

func printSubMenuNav() {
	ui.subMenuNav()
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
