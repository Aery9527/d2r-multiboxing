package main

import (
	"bufio"
	"fmt"
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

func selectLaunchMod(d2rPath string, scanner *bufio.Scanner) ([]string, bool) {
	installedMods, err := mods.DiscoverInstalled(d2rPath)
	if err != nil {
		fmt.Printf("  讀取 mods 失敗：%v\n", err)
		return nil, false
	}

	if len(installedMods) == 0 {
		fmt.Println("  找不到已安裝 mod，將以原版啟動。")
		return nil, true
	}

	fmt.Println()
	fmt.Println("  選擇 mod")
	for {
		fmt.Println("  [0] 不使用 mod")
		for i, modName := range installedMods {
			fmt.Printf("  [%d] %s\n", i+1, modName)
		}
		printSubMenuNav()
		fmt.Print("  > 請選擇：")

		if !scanner.Scan() {
			return nil, false
		}
		input := strings.TrimSpace(scanner.Text())
		if nav := isMenuNav(input); nav != "" {
			return nil, false
		}

		selected, err := strconv.Atoi(input)
		if err != nil || selected < 0 || selected > len(installedMods) {
			showInvalidInputAndPause()
			continue
		}

		if selected == 0 {
			fmt.Println("  本次啟動不使用 mod。")
			return nil, true
		}

		modName := installedMods[selected-1]
		fmt.Printf("  本次使用 mod：%s\n", modName)
		return mods.BuildLaunchArgs(modName), true
	}
}
