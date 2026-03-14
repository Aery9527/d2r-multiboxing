package main

import (
	"strings"

	"d2rhl/internal/common/config"
	"d2rhl/internal/common/locale"
)

// pickInitialLanguage detects the system locale, confirms with the player,
// and saves the chosen language to cfg.  It is called on first run when
// cfg.Language is empty.
func pickInitialLanguage(cfg *config.Config) {
	detected := locale.DetectSystemLocale()
	cat := locale.Get(detected)

	ui.headf("%s", cat.Language.PickerTitle)
	ui.infof(cat.Language.AutoDetected, localeName(cat, detected))

	answer, ok := ui.readInputf("%s", cat.Language.ConfirmUse)
	if !ok {
		saveLanguage(cfg, detected)
		return
	}
	answer = strings.ToLower(strings.TrimSpace(answer))

	if answer == "" || answer == "y" {
		saveLanguage(cfg, detected)
		return
	}

	// Player rejected the auto-detected locale; show the full picker.
	chosen := showLanguagePicker(cat)
	saveLanguage(cfg, chosen)
}

// setupLanguage shows the language picker from the main menu and applies the
// selection immediately by reloading the package-level lang variable.
func setupLanguage(cfg *config.Config) {
	chosen := showLanguagePicker(lang)
	if err := saveLanguageToConfig(cfg, chosen); err != nil {
		ui.warningf(lang.Common.SaveFailed, err)
		return
	}
	lang = locale.Get(chosen)
	ui.successf(lang.Language.Saved, localeName(lang, chosen))
	ui.blankLine()
}

// showLanguagePicker renders a selection menu for all known locales and
// returns the chosen Locale.  Falls back to the current lang's locale on EOF or nav.
func showLanguagePicker(cat *locale.Catalog) locale.Locale {
	knownLocales := locale.KnownLocales()
	result := cat.Locale
	_ = runMenuRead(
		func() {
			ui.headf("%s", cat.Language.PickerTitle)
			options := ui.subMenuOptions(func(o *cliMenuOptions) {
				for i, l := range knownLocales {
					c := locale.Get(l)
					key := string(rune('1' + i))
					o.option(key, localeName(c, l), "")
				}
			})
			ui.menuBlock(func() { options.render() })
		},
		func() (string, bool) {
			return ui.readInputf("%s", cat.Language.PromptSelect)
		},
		func(input string) error {
			idx := int(input[0] - '1')
			if len(input) == 1 && idx >= 0 && idx < len(knownLocales) {
				result = knownLocales[idx]
				return errNavDone
			}
			showInvalidInputAndPause()
			return nil
		},
	)
	return result
}

// saveLanguage writes the chosen locale to cfg and persists the config file.
// On error the player sees a warning and the language is still applied in-memory.
func saveLanguage(cfg *config.Config, l locale.Locale) {
	if err := saveLanguageToConfig(cfg, l); err != nil {
		// Use the detected catalog because lang may not be initialised yet.
		cat := locale.Get(l)
		ui.warningf(cat.Common.SaveFailed, err)
	}
}

func saveLanguageToConfig(cfg *config.Config, l locale.Locale) error {
	cfg.Language = string(l)
	return config.Save(cfg)
}

// localeName returns the human-readable display name for a locale using the
// strings defined in the given catalog.
func localeName(cat *locale.Catalog, l locale.Locale) string {
	switch l {
	case locale.LocaleZhTW:
		return cat.Language.NameZhTW
	case locale.LocaleEn:
		return cat.Language.NameEn
	default:
		return string(l)
	}
}
