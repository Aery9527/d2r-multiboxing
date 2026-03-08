package main

import (
	"strings"

	"d2rhl/internal/common/config"
)

func ensureLaunchReadyD2RPath(cfg *config.Config) bool {
	return ensureLaunchReadyD2RPathWithSetup(cfg, setupD2RPath)
}

func ensureLaunchReadyD2RPathWithSetup(cfg *config.Config, setup func(*config.Config) bool) bool {
	for {
		err := config.ValidateD2RPath(cfg.D2RPath)
		if err == nil {
			return true
		}

		ui.blankLine()
		ui.headf("啟動前檢查 D2R 路徑")
		ui.warningf("找不到可啟動的 D2R.exe：%s", cfg.D2RPath)
		ui.warningf("原因：%v", err)
		ui.promptf("請先設定正確的 D2R.exe 路徑，完成後再繼續啟動。")
		ui.option("p", "立即設定 D2R.exe 路徑")
		printSubMenuNav()
		input, ok := ui.readInput()
		if !ok {
			return false
		}

		if nav := isMenuNav(input); nav != "" {
			return false
		}
		if strings.EqualFold(input, "p") {
			if !setup(cfg) {
				return false
			}
			continue
		}

		showInputErrorAndPause("無效輸入，請輸入 p / b / h / q。")
	}
}

func setupD2RPath(cfg *config.Config) bool {
	ui.blankLine()
	ui.headf("設定 D2R 路徑")
	ui.promptf("即將開啟 Windows 檔案選擇視窗，請選擇 D2R.exe。")

	selectedPath, err := PickD2RPath(cfg.D2RPath)
	if err != nil {
		ui.warningf("D2R 路徑設定失敗：%v", err)
		ui.blankLine()
		return false
	}
	if selectedPath == "" {
		ui.infof("已取消。")
		ui.blankLine()
		return false
	}

	cfg.D2RPath = selectedPath
	if err := config.Save(cfg); err != nil {
		ui.warningf("設定儲存失敗：%v", err)
		ui.blankLine()
		return false
	}

	ui.successf("已更新 D2R 路徑：%s", cfg.D2RPath)
	ui.blankLine()
	return true
}
