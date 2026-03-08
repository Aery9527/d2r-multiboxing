package launcher

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
)

var commandLogger func(string)

func SetCommandLogger(logger func(string)) {
	commandLogger = logger
}

// LaunchD2R starts D2R.exe with the given account parameters and returns the PID.
func LaunchD2R(d2rPath string, username string, password string, address string, extraArgs ...string) (uint32, error) {
	args := buildOnlineArgs(username, password, address, extraArgs...)

	cmd := exec.Command(d2rPath, args...)
	cmd.Dir = filepath.Dir(d2rPath)
	if commandLogger != nil {
		commandLogger(fmt.Sprintf("%s %s", d2rPath, strings.Join(redactArgs(args), " ")))
	}

	err := cmd.Start()
	if err != nil {
		return 0, fmt.Errorf("failed to start D2R: %w", err)
	}

	return uint32(cmd.Process.Pid), nil
}

// LaunchD2ROffline starts D2R.exe without account parameters (offline/single-player mode).
func LaunchD2ROffline(d2rPath string, extraArgs ...string) (uint32, error) {
	args := buildOfflineArgs(extraArgs...)
	cmd := exec.Command(d2rPath, args...)
	cmd.Dir = filepath.Dir(d2rPath)
	if commandLogger != nil {
		commandLogger(fmt.Sprintf("%s %s", d2rPath, strings.Join(args, " ")))
	}

	if err := cmd.Start(); err != nil {
		return 0, fmt.Errorf("failed to start D2R: %w", err)
	}

	return uint32(cmd.Process.Pid), nil
}

func buildOnlineArgs(username string, password string, address string, extraArgs ...string) []string {
	args := make([]string, 0, len(extraArgs)+6)
	args = append(args, extraArgs...)
	args = append(args,
		"-username", username,
		"-password", password,
		"-address", address,
	)
	return args
}

func buildOfflineArgs(extraArgs ...string) []string {
	args := make([]string, 0, len(extraArgs))
	args = append(args, extraArgs...)
	return args
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
