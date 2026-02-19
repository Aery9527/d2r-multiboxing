package process

import (
	"fmt"
	"os/exec"
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

	err := cmd.Start()
	if err != nil {
		return 0, fmt.Errorf("failed to start D2R: %w", err)
	}

	return uint32(cmd.Process.Pid), nil
}
