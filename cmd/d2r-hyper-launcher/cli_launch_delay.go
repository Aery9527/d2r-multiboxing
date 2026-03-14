package main

import (
	"fmt"
	"strings"

	"d2rhl/internal/common/config"
)

// displayDelay formats a LaunchDelayRange for display using the active locale.
// Use this in the CLI layer instead of the domain's DisplayString method.
func displayDelay(r config.LaunchDelayRange) string {
	if r.MinSeconds == r.MaxSeconds {
		return fmt.Sprintf(lang.Delay.DisplayFixed, r.MinSeconds)
	}
	return fmt.Sprintf(lang.Delay.DisplayRandom, r.MinSeconds, r.MaxSeconds)
}

func setupLaunchDelay(cfg *config.Config) {
	showUI := func() {
		ui.headf("%s", lang.Delay.Title)
		ui.infof(lang.Delay.CurrentSetting, displayDelay(cfg.LaunchDelay))
		ui.infof("%s", lang.Delay.Description)
		options := ui.subMenuOptions(nil)
		ui.menuBlock(func() {
			ui.infof(lang.Delay.MinLabel, config.MinLaunchDelaySeconds)
			ui.infof("%s", lang.Delay.HintFixed)
			ui.infof("%s", lang.Delay.HintRange)
			options.render()
		})
	}
	_ = runMenuRead(showUI, func() (string, bool) {
		return ui.readInputf("%s", lang.Delay.InputPrompt)
	}, func(input string) error {
		delay, err := parseLaunchDelayInput(input)
		if err != nil {
			showInputErrorAndPause(err.Error())
			return nil
		}
		cfg.LaunchDelay = delay
		if err := config.Save(cfg); err != nil {
			showInputErrorAndPause(fmt.Sprintf(lang.Common.SaveFailed, err))
			return nil
		}
		ui.successf(lang.Delay.Updated, displayDelay(delay))
		ui.blankLine()
		return errNavDone
	})
}

func parseLaunchDelayInput(input string) (config.LaunchDelayRange, error) {
	return config.ParseLaunchDelayRange(strings.TrimSpace(input))
}
