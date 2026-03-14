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

	for {
		options := ui.subMenuOptions(func(options *cliMenuOptions) {
			options.option("0", lang.Launch.ModOptNone, "")
			for i, modName := range installedMods {
				options.option(strconv.Itoa(i+1), modName, "")
			}
		})
		ui.menuBlock(func() {
			options.render()
		})
		input, ok := ui.readInput()
		if !ok {
			return nil, false
		}
		if nav := isMenuNav(input); nav != "" {
			return nil, false
		}

		selected, err := strconv.Atoi(input)
		if err != nil || selected < 0 || selected > len(installedMods) {
			showInvalidInputAndPause()
			continue
		}

		if selected == 0 {
			ui.infof("%s", lang.Launch.ModNoneChosen)
			return nil, true
		}

		modName := installedMods[selected-1]
		ui.infof(lang.Launch.ModUsing, modName)
		return mods.BuildLaunchArgs(modName), true
	}
}
