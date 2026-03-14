package main

import (
	"strings"

	"d2rhl/internal/common/config"
	"d2rhl/internal/switcher"
)

func setupSwitcher(cfg *config.Config) {
	for {
		ui.headf("%s", lang.Switcher.Title)

		if display, ok := switcherSavedDisplay(cfg); ok {
			if cfg.Switcher.Enabled {
				ui.infof(lang.Switcher.StatusLabel, lang.Switcher.StatusEnabled)
				ui.infof(lang.Switcher.SettingLabel, display)
			} else {
				ui.infof(lang.Switcher.StatusLabel, lang.Switcher.StatusDisabled)
				ui.infof(lang.Switcher.SavedLabel, display)
			}
		} else {
			ui.infof(lang.Switcher.StatusLabel, lang.Switcher.StatusNotSet)
		}

		ui.blankLine()
		options := ui.subMenuOptions(func(options *cliMenuOptions) {
			options.option("1", lang.Switcher.OptSetKey, "")
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
			ui.infof("%s", lang.Switcher.DetectInstruction)
			ui.infof("%s", lang.Switcher.DetectSupport)
			ui.infof("%s", lang.Switcher.DetectGamepad)
			ui.infof("%s", lang.Switcher.DetectEscCancel)

			modifiers, key, gamepadIndex, err := switcher.DetectKeyPress()
			if err != nil {
				ui.warningf(lang.Switcher.DetectFailed, err)
				restartSwitcherIfNeeded(cfg, wasRunning)
				return
			}
			if key == "" {
				ui.infof("%s", lang.Switcher.DetectCancelled)
				restartSwitcherIfNeeded(cfg, wasRunning)
				return
			}

			display := switcher.FormatSwitcherDisplay(modifiers, key, gamepadIndex)
			ui.infof(lang.Switcher.DetectedKey, display)
			answer, ok := ui.readInputf("%s", lang.Switcher.DetectConfirmPrompt)
			if !ok {
				return
			}
			answer = strings.ToLower(answer)
			if answer != "" && answer != "y" {
				ui.infof("%s", lang.Switcher.DetectCancelled)
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
				ui.warningf(lang.Common.SaveFailed, err)
				return
			}

			if err := switcher.Start(cfg.Switcher); err != nil {
				ui.warningf(lang.Switcher.StartFailed, err)
				return
			}

			ui.successf(lang.Switcher.KeySet, display)
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
			ui.warningf(lang.Switcher.RestartFailed, err)
		}
	}
}

func switcherToggleOptionLabel(cfg *config.Config) string {
	if cfg != nil && cfg.Switcher != nil && cfg.Switcher.Enabled {
		return lang.Switcher.OptDisable
	}
	return lang.Switcher.OptEnable
}

func toggleSwitcherEnabled(cfg *config.Config) bool {
	if cfg == nil || cfg.Switcher == nil || cfg.Switcher.Key == "" {
		showInputErrorAndPause(lang.Switcher.ToggleNotSet)
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
		ui.warningf(lang.Switcher.StartFailed, err)
		return
	}

	cfg.Switcher.Enabled = true
	if err := config.Save(cfg); err != nil {
		cfg.Switcher.Enabled = false
		switcher.Stop()
		ui.warningf(lang.Common.SaveFailed, err)
		return
	}

	ui.successf(lang.Switcher.Enabled, switcher.FormatSwitcherDisplay(candidate.Modifiers, candidate.Key, candidate.GamepadIndex))
}

func disableSwitcher(cfg *config.Config) {
	switcher.Stop()
	cfg.Switcher.Enabled = false
	if err := config.Save(cfg); err != nil {
		cfg.Switcher.Enabled = true
		if restartErr := switcher.Start(cfg.Switcher); restartErr != nil {
			ui.warningf(lang.Switcher.DisableSaveAndRestoreFailed, err, restartErr)
			return
		}
		ui.warningf(lang.Common.SaveFailed, err)
		return
	}

	ui.successf("%s", lang.Switcher.Disabled)
}
