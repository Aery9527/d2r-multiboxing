package main

import (
	"strings"

	"d2rhl/internal/common/config"
	"d2rhl/internal/switcher"
)

func setupSwitcher(cfg *config.Config) {
	for {
		ui.headf("視窗切換設定")

		if display, ok := switcherSavedDisplay(cfg); ok {
			if cfg.Switcher.Enabled {
				ui.infof("目前狀態：已啟用")
				ui.infof("目前設定：%s", display)
			} else {
				ui.infof("目前狀態：未啟用")
				ui.infof("已保存設定：%s", display)
			}
		} else {
			ui.infof("目前狀態：未設定")
		}

		ui.blankLine()
		options := ui.subMenuOptions(func(options *cliMenuOptions) {
			options.option("1", "設定切換按鍵", "")
			options.option("0", switcherToggleOptionLabel(cfg), "")
		})
		ui.menuBlock(func() {
			options.render()
		})
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
			ui.infof("請按下想用來切換視窗的按鍵組合...")
			ui.infof("支援：鍵盤任意鍵 + Ctrl/Alt/Shift、滑鼠側鍵、搖桿按鈕")
			ui.infof("搖桿組合鍵：先按住修飾按鈕，再按觸發按鈕，放開後完成偵測")
			ui.infof("按 Esc 取消")

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
			ui.blankLine()
			return

		case "0":
			if !toggleSwitcherEnabled(cfg) {
				continue
			}
			ui.blankLine()
			return
		default:
			showInvalidInputAndPause()
		}
	}
}

func restartSwitcherIfNeeded(cfg *config.Config, wasRunning bool) {
	if wasRunning && cfg.Switcher != nil && cfg.Switcher.Enabled {
		if err := switcher.Start(cfg.Switcher); err != nil {
			ui.warningf("重新啟動切換功能失敗：%v", err)
		}
	}
}

func switcherToggleOptionLabel(cfg *config.Config) string {
	if cfg != nil && cfg.Switcher != nil && cfg.Switcher.Enabled {
		return "切換為關閉"
	}
	return "切換為開啟"
}

func toggleSwitcherEnabled(cfg *config.Config) bool {
	if cfg == nil || cfg.Switcher == nil || cfg.Switcher.Key == "" {
		showInputErrorAndPause("尚未設定切換按鍵，請先使用 [1] 設定。")
		return false
	}

	if cfg.Switcher.Enabled {
		disableSwitcher(cfg)
		return true
	}

	enableSwitcher(cfg)
	return true
}

func enableSwitcher(cfg *config.Config) {
	candidate := *cfg.Switcher
	candidate.Enabled = true
	if err := switcher.Start(&candidate); err != nil {
		ui.warningf("切換功能啟動失敗：%v", err)
		return
	}

	cfg.Switcher.Enabled = true
	if err := config.Save(cfg); err != nil {
		cfg.Switcher.Enabled = false
		switcher.Stop()
		ui.warningf("設定儲存失敗：%v", err)
		return
	}

	ui.successf("已開啟切換功能：%s", switcher.FormatSwitcherDisplay(candidate.Modifiers, candidate.Key, candidate.GamepadIndex))
}

func disableSwitcher(cfg *config.Config) {
	switcher.Stop()
	cfg.Switcher.Enabled = false
	if err := config.Save(cfg); err != nil {
		cfg.Switcher.Enabled = true
		if restartErr := switcher.Start(cfg.Switcher); restartErr != nil {
			ui.warningf("設定儲存失敗：%v；且無法恢復切換功能：%v", err, restartErr)
			return
		}
		ui.warningf("設定儲存失敗：%v", err)
		return
	}

	ui.successf("已關閉切換功能，原設定會保留。")
}
