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
		ui.errorf("讀取 mods 失敗：%v", err)
		return nil, false
	}

	if len(installedMods) == 0 {
		ui.infof("找不到已安裝 mod，將以原版啟動。")
		return nil, true
	}

	ui.blankLine()
	ui.headf("選擇 mod")
	for {
		options := ui.newMenuOptions()
		options.option("0", "不使用 mod")
		for i, modName := range installedMods {
			options.option(strconv.Itoa(i+1), modName)
		}
		options.subMenuNav()
		ui.menuBlock(func() {
			options.render(ui)
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
			ui.infof("本次啟動不使用 mod。")
			return nil, true
		}

		modName := installedMods[selected-1]
		ui.successf("本次使用 mod：%s", modName)
		return mods.BuildLaunchArgs(modName), true
	}
}
