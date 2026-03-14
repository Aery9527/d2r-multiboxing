package main

import (
	"errors"
	"strconv"
	"strings"

	"d2rhl/internal/common/config"
	"d2rhl/internal/multiboxing/account"
	"d2rhl/internal/switcher"
)

func setupSwitcher(cfg *config.Config, accounts []account.Account, accountsFile string) {
	showUI := func() {
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
			options.option("2", lang.Switcher.OptSetAccounts, "")
			options.option("0", switcherToggleOptionLabel(cfg), "")
		})
		ui.menuBlock(func() {
			options.render()
		})
	}

	_ = runMenu(showUI, func(choice string) error {
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
				return errNavDone
			}
			if key == "" {
				ui.infof("%s", lang.Switcher.DetectCancelled)
				restartSwitcherIfNeeded(cfg, wasRunning)
				return errNavDone
			}

			display := switcher.FormatSwitcherDisplay(modifiers, key, gamepadIndex)
			ui.infof(lang.Switcher.DetectedKey, display)
			answer, ok := ui.readInputf("%s", lang.Switcher.DetectConfirmPrompt)
			if !ok {
				return errNavDone
			}
			answer = strings.ToLower(answer)
			if answer != "" && answer != "y" {
				ui.infof("%s", lang.Switcher.DetectCancelled)
				restartSwitcherIfNeeded(cfg, wasRunning)
				return errNavDone
			}

			cfg.Switcher = &config.SwitcherConfig{
				Enabled:      true,
				Modifiers:    modifiers,
				Key:          key,
				GamepadIndex: gamepadIndex,
			}
			if err := config.Save(cfg); err != nil {
				ui.warningf(lang.Common.SaveFailed, err)
				return errNavDone
			}

			if err := switcher.Start(cfg.Switcher); err != nil {
				ui.warningf(lang.Switcher.StartFailed, err)
				return errNavDone
			}

			ui.successf(lang.Switcher.KeySet, display)
			ui.blankLine()
			return errNavDone

		case "2":
			if err := setupSwitcherAccounts(cfg, accountsFile); errors.Is(err, ErrNavHome) {
				return ErrNavHome
			}
			if reloaded, err := account.LoadAccounts(accountsFile); err == nil {
				accounts = reloaded
			}

		case "0":
			if !toggleSwitcherEnabled(cfg) {
				return nil
			}
			ui.blankLine()
			return errNavDone

		default:
			showInvalidInputAndPause()
		}
		return nil
	})
}

func setupSwitcherAccounts(cfg *config.Config, accountsFile string) error {
	accounts, err := account.LoadAccounts(accountsFile)
	if err != nil || len(accounts) == 0 {
		ui.warningf("%s", lang.Switcher.AccountFilterNoAccounts)
		ui.blankLine()
		return nil
	}

	showUI := func() {
		ui.headf("%s", lang.Switcher.AccountFilterTitle)
		ui.infoLines(
			lang.Switcher.AccountFilterDescIncluded,
			lang.Switcher.AccountFilterDescExcluded,
		)
		printAccountList(accounts, func(acc account.Account) string {
			if account.SkipSwitcher(acc.ToolFlags) {
				return lang.Switcher.AccountExcluded
			}
			return lang.Switcher.AccountIncluded
		})
		ui.blankLine()
		toggleKey := "1~" + strconv.Itoa(len(accounts))
		options := ui.subMenuOptions(func(options *cliMenuOptions) {
			options.option(toggleKey, lang.Switcher.AccountFilterOptToggle, "")
			options.option("a", lang.Switcher.AccountFilterOptAll, "")
			options.option("n", lang.Switcher.AccountFilterOptNone, "")
		})
		ui.menuBlock(func() {
			options.render()
		})
	}

	return runMenu(showUI, func(choice string) error {
		switch choice {
		case "a":
			for i := range accounts {
				accounts[i].ToolFlags &^= account.ToolFlagSkipSwitcher
			}
		case "n":
			for i := range accounts {
				accounts[i].ToolFlags |= account.ToolFlagSkipSwitcher
			}
		default:
			idx, parseErr := strconv.Atoi(choice)
			if parseErr != nil || idx < 1 || idx > len(accounts) {
				showInvalidInputAndPause()
				return nil
			}
			i := idx - 1
			if account.SkipSwitcher(accounts[i].ToolFlags) {
				accounts[i].ToolFlags &^= account.ToolFlagSkipSwitcher
			} else {
				accounts[i].ToolFlags |= account.ToolFlagSkipSwitcher
			}
		}

		if saveErr := account.SaveAccounts(accountsFile, accounts); saveErr != nil {
			ui.warningf(lang.Common.SaveFailed, saveErr)
			if reloaded, loadErr := account.LoadAccounts(accountsFile); loadErr == nil {
				accounts = reloaded
			}
			return nil
		}

		switcher.UpdateExcludedAccounts(account.ExcludedFromSwitcher(accounts))
		ui.successf("%s", lang.Switcher.AccountFilterSaved)

		included := 0
		for _, acc := range accounts {
			if !account.SkipSwitcher(acc.ToolFlags) {
				included++
			}
		}
		switch {
		case included == 0:
			showWarningAndPause(lang.Switcher.AccountFilterWarnNoneIncluded)
		case included == 1:
			showWarningAndPause(lang.Switcher.AccountFilterWarnOneIncluded)
		default:
			ui.blankLine()
		}
		return nil
	})
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
