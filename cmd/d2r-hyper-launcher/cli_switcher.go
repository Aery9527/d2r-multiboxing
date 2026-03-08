package main

import (
	"strings"

	"d2rhl/internal/common/config"
	"d2rhl/internal/switcher"
)

func setupSwitcher(cfg *config.Config) {
	ui.blankLine()
	ui.headf("視窗切換設定")

	if cfg.Switcher != nil && cfg.Switcher.Enabled {
		ui.infof("目前設定：%s", switcher.FormatSwitcherDisplay(cfg.Switcher.Modifiers, cfg.Switcher.Key, cfg.Switcher.GamepadIndex))
	} else {
		ui.infof("目前狀態：未啟用")
	}

	ui.blankLine()
	ui.option("1", "設定切換按鍵")
	ui.option("0", "關閉切換功能")
	printSubMenuNav()
	choice, ok := ui.readInput()
	if !ok {
		return
	}
	if isMenuNav(choice) != "" {
		return
	}

	switch choice {
	case "1":
		wasRunning := switcher.IsRunning()
		switcher.Stop()

		ui.blankLine()
		ui.promptf("請按下想用來切換視窗的按鍵組合...")
		ui.infof("（支援：鍵盤任意鍵 + Ctrl/Alt/Shift、滑鼠側鍵、搖桿按鈕）")
		ui.infof("（搖桿組合鍵：先按住修飾按鈕，再按觸發按鈕，放開後完成偵測）")
		ui.infof("（按 Esc 取消）")
		ui.blankLine()

		modifiers, key, gamepadIndex, err := switcher.DetectKeyPress()
		if err != nil {
			ui.warningf("偵測失敗：%v", err)
			restartSwitcherIfNeeded(cfg, wasRunning)
			return
		}
		if key == "" {
			ui.infof("已取消。")
			restartSwitcherIfNeeded(cfg, wasRunning)
			return
		}

		display := switcher.FormatSwitcherDisplay(modifiers, key, gamepadIndex)
		ui.infof("偵測到：%s", display)
		answer, ok := ui.readInputf("確認使用此組合？([Y]/[n])：")
		if !ok {
			return
		}
		answer = strings.ToLower(answer)
		if answer != "" && answer != "y" {
			ui.infof("已取消。")
			restartSwitcherIfNeeded(cfg, wasRunning)
			return
		}

		cfg.Switcher = &config.SwitcherConfig{
			Enabled:      true,
			Modifiers:    modifiers,
			Key:          key,
			GamepadIndex: gamepadIndex,
		}
		if err := config.Save(cfg); err != nil {
			ui.warningf("設定儲存失敗：%v", err)
			return
		}

		if err := switcher.Start(cfg.Switcher); err != nil {
			ui.warningf("切換功能啟動失敗：%v", err)
			return
		}

		ui.successf("已儲存切換設定：%s", display)

	case "0":
		switcher.Stop()
		if cfg.Switcher != nil {
			cfg.Switcher.Enabled = false
		}
		if err := config.Save(cfg); err != nil {
			ui.warningf("設定儲存失敗：%v", err)
			return
		}
		ui.successf("已關閉切換功能")
	default:
		showInvalidInputAndPause()
	}

	ui.blankLine()
}

func restartSwitcherIfNeeded(cfg *config.Config, wasRunning bool) {
	if wasRunning && cfg.Switcher != nil && cfg.Switcher.Enabled {
		if err := switcher.Start(cfg.Switcher); err != nil {
			ui.warningf("重新啟動切換功能失敗：%v", err)
		}
	}
}
