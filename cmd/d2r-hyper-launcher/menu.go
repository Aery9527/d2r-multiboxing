package main

import (
	"fmt"
	"os"
	"strings"

	"d2rhl/internal/common/config"
	"d2rhl/internal/multiboxing/account"
	"d2rhl/internal/switcher"
)

const (
	menuBack = "b"
	menuHome = "h"
	menuQuit = "q"
)

func printStartupAnnouncement(cfgDir string) {
	ui.headf("d2r-hyper-launcher (%s)", displayVersion(version))

	ui.infof("目前版本：%s", displayReleaseSummary(version, releaseTime))
	ui.infof("資料目錄：%s", cfgDir)
	ui.warningLines(
		"注意：帳號啟動狀態的偵測是用 account.csv 裡的 DisplayName 去對應視窗名稱，",
		"所以已經透過該工具開啟 D2R 然後又去修改 DisplayName的話，",
		"就會導致啟動狀態顯示不正確。",
	)
	ui.warningLines(
		"注意事項：",
		"- 建議先把 D2R 設成「視窗化」或「無邊框視窗」",
		"- 設定搖桿切換按鍵時，建議以管理員權限執行，否則可能抓不到搖桿訊號。",
		"- switcher 只有在 d2r-hyper-launcher 持續開著時才會生效；若把工具關掉，切窗功能也會停止作用。",
		"- a 批次啟動預設 launch_delay 是 10 秒；舊版預設留下的 5 秒會自動按 10 秒處理，如要調整請回主選單輸入 d。",
		"- 盡量不要手動修改 config.json，避免不小心破壞 JSON 格式；大部分設定請優先透過工具內建選單調整。",
		"- 僅支援 Battle.net 版 D2R；操作進程 Handle 也可能被部分防毒軟體誤報。",
		"- 本工具不會修改遊戲檔案、注入遊戲程式或自動化遊戲操作。",
		"- 本工具為社群自用工具，與 Blizzard Entertainment 無關；使用風險自負。",
	)
}

func pauseAfterStartupAnnouncement() {
	if err := ui.anyKeyContinue(); err != nil {
		ui.warningf("等待按鍵失敗：%v", err)
	}
}

func printMenu(accounts []account.Account, cfg *config.Config) {
	ui.headf("主選單")
	printAccountList(accounts)

	ui.blankLine()
	options := ui.mainMenuOptions(func(options *cliMenuOptions) {
		options.option("數字", "啟動指定帳號", "")
		options.option("0", "離線遊玩", "可選 mod，不需帳密")
		options.option("a", "啟動所有帳號", "可選 mod，只啟動未啟動的")
		options.option("d", "設定啟動間隔", cfg.LaunchDelay.DisplayString())
		options.option("f", "設定帳號啟動 flag", "進入可查看所有帳號的 flag 設定")
		options.option("p", "選擇 D2R.exe 路徑", cfg.D2RPath)
		options.option("s", "視窗切換設定", switcherMenuOptionStatus(cfg))
		options.option("r", "重新整理狀態", "")
	})
	ui.menuBlock(func() {
		options.render()
	})
}

func switcherMenuOptionStatus(cfg *config.Config) string {
	display, ok := switcherSavedDisplay(cfg)
	if !ok {
		return "未設定"
	}
	if !cfg.Switcher.Enabled {
		return fmt.Sprintf("未啟用設定：%s", display)
	}
	return fmt.Sprintf("已啟用設定：%s", display)
}

func switcherSavedDisplay(cfg *config.Config) (string, bool) {
	if cfg == nil || cfg.Switcher == nil || cfg.Switcher.Key == "" {
		return "", false
	}
	return switcher.FormatSwitcherDisplay(cfg.Switcher.Modifiers, cfg.Switcher.Key, cfg.Switcher.GamepadIndex), true
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
