package process

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
)

// LaunchD2R starts D2R.exe with the given account parameters and returns the PID.
func LaunchD2R(d2rPath string, username string, password string, address string, extraArgs ...string) (uint32, error) {
	args := append([]string{}, extraArgs...)
	args = append(args,
		"-username", username,
		"-password", password,
		"-address", address,
	)

	cmd := exec.Command(d2rPath, args...)
	cmd.Dir = filepath.Dir(d2rPath)
	fmt.Printf("  > %s %s\n", d2rPath, strings.Join(redactArgs(args), " "))

	err := cmd.Start()
	if err != nil {
		return 0, fmt.Errorf("failed to start D2R: %w", err)
	}

	return uint32(cmd.Process.Pid), nil
}

// LaunchD2ROffline starts D2R.exe without account parameters (offline/single-player mode).
func LaunchD2ROffline(d2rPath string, extraArgs ...string) (uint32, error) {
	cmd := exec.Command(d2rPath, extraArgs...)
	cmd.Dir = filepath.Dir(d2rPath)
	fmt.Printf("  > %s %s\n", d2rPath, strings.Join(extraArgs, " "))

	if err := cmd.Start(); err != nil {
		return 0, fmt.Errorf("failed to start D2R: %w", err)
	}

	return uint32(cmd.Process.Pid), nil
}

// redactArgs masks the value after -password to avoid leaking credentials in console output.
func redactArgs(args []string) []string {
	out := make([]string, len(args))
	copy(out, args)
	for i, a := range out {
		if a == "-password" && i+1 < len(out) {
			out[i+1] = "****"
		}
	}
	return out
}
