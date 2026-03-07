package main

import (
	"fmt"
	"strings"
	"time"
)

var (
	cliInputErrorPauseSleep = time.Sleep
	cliInputErrorPauseStep  = 500 * time.Millisecond
	cliInputErrorPauseCount = 6
)

func showInvalidInputAndPause() {
	showInputErrorAndPause("無效輸入，請重試。")
}

func showInputErrorAndPause(message string) {
	fmt.Printf("  %s\n", message)
	for i := 1; i <= cliInputErrorPauseCount; i++ {
		cliInputErrorPauseSleep(cliInputErrorPauseStep)
		fmt.Printf("\r  %s", strings.Repeat(".", i))
	}
	fmt.Print("\n\n")
}
