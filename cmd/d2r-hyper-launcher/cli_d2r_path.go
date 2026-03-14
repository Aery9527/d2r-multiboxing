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

		ui.headf("%s", lang.D2RPath.PreCheckTitle)
		ui.warningf(lang.D2RPath.PathNotFound, cfg.D2RPath)
		ui.warningf(lang.D2RPath.PathError, err)
		ui.promptf("%s", lang.D2RPath.PromptFix)
		options := ui.subMenuOptions(func(options *cliMenuOptions) {
			options.option("p", lang.D2RPath.OptSetPath, "")
		})
		ui.menuBlock(func() {
			options.render()
		})
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

		showInputErrorAndPause(lang.D2RPath.PreCheckInvalidInput)
	}
}

func setupD2RPath(cfg *config.Config) bool {
	ui.headf("%s", lang.D2RPath.SetTitle)
	ui.promptf("%s", lang.D2RPath.SetPrompt)

	selectedPath, err := PickD2RPath(cfg.D2RPath, lang.D2RPath.PickerDialogTitle)
	if err != nil {
		ui.warningf(lang.D2RPath.SetFailed, err)
		ui.blankLine()
		return false
	}
	if selectedPath == "" {
		ui.infof("%s", lang.Common.Cancelled)
		ui.blankLine()
		return false
	}

	cfg.D2RPath = selectedPath
	if err := config.Save(cfg); err != nil {
		ui.warningf(lang.Common.SaveFailed, err)
		ui.blankLine()
		return false
	}

	ui.successf(lang.D2RPath.SetOK, cfg.D2RPath)
	ui.blankLine()
	return true
}
