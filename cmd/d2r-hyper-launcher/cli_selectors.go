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

func discoverInstalledMods(d2rPath string) ([]string, bool) {
	installedMods, err := mods.DiscoverInstalled(d2rPath)
	if err != nil {
		ui.errorf(lang.Launch.ModLoadFailed, err)
		return nil, false
	}
	return installedMods, true
}

func parseLaunchModInput(input string, installedMods []string) (string, bool) {
	selected, err := strconv.Atoi(strings.TrimSpace(input))
	if err != nil || selected < 0 || selected > len(installedMods) {
		return "", false
	}
	if selected == 0 {
		return mods.DefaultModVanilla, true
	}
	return installedMods[selected-1], true
}

func renderLaunchModOptions(installedMods []string) {
	options := ui.subMenuOptions(func(options *cliMenuOptions) {
		options.option("0", lang.Launch.ModOptNone, "")
		for i, modName := range installedMods {
			options.option(strconv.Itoa(i+1), modName, "")
		}
	})
	options.render()
}

func selectOfflineLaunchMod(installedMods []string) ([]string, bool) {
	if len(installedMods) == 0 {
		ui.infof("%s", lang.Launch.ModNoMods)
		return nil, true
	}

	var result []string
	chosen := false
	_ = runMenu(func() {
		ui.headf("%s", lang.Launch.ModOfflineTitle)
		ui.menuBlock(func() {
			renderLaunchModOptions(installedMods)
		})
	}, func(input string) error {
		selectedMod, ok := parseLaunchModInput(input, installedMods)
		if !ok {
			showInvalidInputAndPause()
			return nil
		}
		if selectedMod == mods.DefaultModVanilla {
			ui.infof("%s", lang.Launch.ModNoneChosen)
			result = nil
		} else {
			ui.infof(lang.Launch.ModUsing, selectedMod)
			result = mods.BuildLaunchArgs(selectedMod)
		}
		chosen = true
		return errNavDone
	})
	return result, chosen
}
