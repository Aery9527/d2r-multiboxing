package main

import (
	"fmt"
	"strings"

	"d2rhl/internal/common/config"
)

func setupLaunchDelay(cfg *config.Config) {
	for {
		ui.blankLine()
		ui.headf("啟動間隔設定")
		ui.infof("目前設定：%s", cfg.LaunchDelay.DisplayString())
		ui.infof("說明：這會影響主選單 [a]「啟動所有帳號」時，每個帳號之間的等待秒數。")
		options := ui.subMenuOptions(nil)
		ui.menuBlock(func() {
			ui.infof("固定下限：%d 秒", config.MinLaunchDelaySeconds)
			ui.infof("可輸入固定秒數，例如：30")
			ui.infof("也可輸入隨機範圍，例如：30-60")
			options.render()
		})
		input, ok := ui.readInputf("請輸入新的秒數或範圍：")
		if !ok {
			return
		}
		if isMenuNav(input) != "" {
			return
		}

		delay, err := parseLaunchDelayInput(input)
		if err != nil {
			showInputErrorAndPause(err.Error())
			continue
		}

		cfg.LaunchDelay = delay
		if err := config.Save(cfg); err != nil {
			showInputErrorAndPause(fmt.Sprintf("設定儲存失敗：%v", err))
			continue
		}

		ui.successf("已更新啟動間隔：%s", delay.DisplayString())
		ui.blankLine()
		return
	}
}

func parseLaunchDelayInput(input string) (config.LaunchDelayRange, error) {
	return config.ParseLaunchDelayRange(strings.TrimSpace(input))
}
