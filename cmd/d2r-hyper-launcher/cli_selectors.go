package main

import (
	"strconv"
	"strings"

	"d2rhl/internal/common/d2r"
	"d2rhl/internal/multiboxing/mods"
)

func parseRegionInput(input string) *d2r.Region {
	input = strings.TrimSpace(strings.ToUpper(input))
	switch input {
	case "1", "NA":
		return d2r.FindRegion("NA")
	case "2", "EU":
		return d2r.FindRegion("EU")
	case "3", "ASIA":
		return d2r.FindRegion("Asia")
	default:
		return d2r.FindRegion(input)
	}
}

func selectLaunchMod(d2rPath string) ([]string, bool) {
	installedMods, err := mods.DiscoverInstalled(d2rPath)
	if err != nil {
		ui.errorf(lang.Launch.ModLoadFailed, err)
		return nil, false
	}

	if len(installedMods) == 0 {
		ui.infof("%s", lang.Launch.ModNoMods)
		return nil, true
	}

	var result []string
	chosen := false
	_ = runMenu(func() {
		options := ui.subMenuOptions(func(options *cliMenuOptions) {
			options.option("0", lang.Launch.ModOptNone, "")
			for i, modName := range installedMods {
				options.option(strconv.Itoa(i+1), modName, "")
			}
		})
		ui.menuBlock(func() {
			options.render()
		})
	}, func(input string) error {
		selected, err := strconv.Atoi(input)
		if err != nil || selected < 0 || selected > len(installedMods) {
			showInvalidInputAndPause()
			return nil
		}
		if selected == 0 {
			ui.infof("%s", lang.Launch.ModNoneChosen)
			result = nil
		} else {
			modName := installedMods[selected-1]
			ui.infof(lang.Launch.ModUsing, modName)
			result = mods.BuildLaunchArgs(modName)
		}
		chosen = true
		return errNavDone
	})
	return result, chosen
}
