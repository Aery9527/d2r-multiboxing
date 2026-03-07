package main

import (
	"fmt"
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
	fmt.Println("  帳號列表：")
	for i, acc := range accounts {
		status := "未啟動"
		if process.FindWindowByTitle(d2r.WindowTitle(acc.DisplayName)) {
			status = "已啟動"
		}
		fmt.Printf("  [%d] %-15s (%s)  [%s]\n", i+1, acc.DisplayName, acc.Email, status)
	}

	fmt.Println()
	fmt.Println("  *是否已啟動的判斷基準是用 account.csv 裡的 DisplayName 來對應視窗名稱。")
	fmt.Println("   如果 D2R 還開著就先關掉工具再去改 DisplayName，之後這裡的啟動狀態偵測可能會不正確。")
	fmt.Println()
	fmt.Println("--------------------------------------------")
	fmt.Println("  <數字>  啟動指定帳號")
	fmt.Println("  0       離線遊玩（可選 mod，不需帳密）")
	fmt.Println("  a       啟動所有帳號（可選 mod，只啟動未啟動的）")
	fmt.Println("  f       設定帳號啟動 flag")
	fmt.Println("  p       選擇 D2R.exe 路徑")
	fmt.Println("  s       視窗切換設定")
	fmt.Println("  r       重新整理狀態")
	fmt.Println("  q       退出")
	fmt.Println("--------------------------------------------")
}

func printSubMenuNav() {
	fmt.Printf("  %s       回上一層\n", menuBack)
	fmt.Printf("  %s       回主選單\n", menuHome)
	fmt.Printf("  %s       離開程式\n", menuQuit)
}

func isMenuNav(input string) string {
	switch strings.ToLower(strings.TrimSpace(input)) {
	case menuBack:
		return "back"
	case menuHome:
		return "home"
	case menuQuit:
		fmt.Println("  再見！")
		os.Exit(0)
		return ""
	default:
		return ""
	}
}
