package main

import (
	"bufio"
	"fmt"
	"strings"

	"d2rhl/internal/common/config"
	"d2rhl/internal/switcher"
)

func setupSwitcher(cfg *config.Config, scanner *bufio.Scanner) {
	fmt.Println()
	fmt.Println("  === 視窗切換設定 ===")

	if cfg.Switcher != nil && cfg.Switcher.Enabled {
		fmt.Printf("  目前設定：%s\n", switcher.FormatSwitcherDisplay(cfg.Switcher.Modifiers, cfg.Switcher.Key, cfg.Switcher.GamepadIndex))
	} else {
		fmt.Println("  目前狀態：未啟用")
	}

	fmt.Println()
	fmt.Println("  [1] 設定切換按鍵")
	fmt.Println("  [0] 關閉切換功能")
	printSubMenuNav()
	fmt.Print("  > 請選擇：")

	if !scanner.Scan() {
		return
	}
	choice := strings.TrimSpace(scanner.Text())
	if isMenuNav(choice) != "" {
		return
	}

	switch choice {
	case "1":
		wasRunning := switcher.IsRunning()
		switcher.Stop()

		fmt.Println()
		fmt.Println("  請按下想用來切換視窗的按鍵組合...")
		fmt.Println("  （支援：鍵盤任意鍵 + Ctrl/Alt/Shift、滑鼠側鍵、搖桿按鈕）")
		fmt.Println("  （搖桿組合鍵：先按住修飾按鈕，再按觸發按鈕，放開後完成偵測）")
		fmt.Println("  （按 Esc 取消）")
		fmt.Println()

		modifiers, key, gamepadIndex, err := switcher.DetectKeyPress()
		if err != nil {
			fmt.Printf("  ⚠ 偵測失敗：%v\n", err)
			restartSwitcherIfNeeded(cfg, wasRunning)
			return
		}
		if key == "" {
			fmt.Println("  已取消。")
			restartSwitcherIfNeeded(cfg, wasRunning)
			return
		}

		display := switcher.FormatSwitcherDisplay(modifiers, key, gamepadIndex)
		fmt.Printf("  偵測到：%s\n", display)
		fmt.Print("  確認使用此組合？(Y/n)：")

		if !scanner.Scan() {
			return
		}
		answer := strings.ToLower(strings.TrimSpace(scanner.Text()))
		if answer != "" && answer != "y" {
			fmt.Println("  已取消。")
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
			fmt.Printf("  ⚠ 設定儲存失敗：%v\n", err)
			return
		}

		if err := switcher.Start(cfg.Switcher); err != nil {
			fmt.Printf("  ⚠ 切換功能啟動失敗：%v\n", err)
			return
		}

		fmt.Printf("  ✔ 已儲存切換設定：%s\n", display)

	case "0":
		switcher.Stop()
		if cfg.Switcher != nil {
			cfg.Switcher.Enabled = false
		}
		if err := config.Save(cfg); err != nil {
			fmt.Printf("  ⚠ 設定儲存失敗：%v\n", err)
			return
		}
		fmt.Println("  ✔ 已關閉切換功能")
	default:
		showInvalidInputAndPause()
	}

	fmt.Println()
}

func restartSwitcherIfNeeded(cfg *config.Config, wasRunning bool) {
	if wasRunning && cfg.Switcher != nil && cfg.Switcher.Enabled {
		if err := switcher.Start(cfg.Switcher); err != nil {
			fmt.Printf("  ⚠ 重新啟動切換功能失敗：%v\n", err)
		}
	}
}
