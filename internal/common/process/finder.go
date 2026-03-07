// Package process provides utilities for Windows process management,
// including process discovery, launching, and window manipulation.
package process

import (
	"fmt"
	"strings"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

// ProcessInfo represents information about a process.
type ProcessInfo struct {
	PID  uint32
	Name string
}

// FindProcessesByName finds all processes with the given name (case-insensitive).
func FindProcessesByName(name string) ([]ProcessInfo, error) {
	snapshot, err := windows.CreateToolhelp32Snapshot(windows.TH32CS_SNAPPROCESS, 0)
	if err != nil {
		return nil, fmt.Errorf("CreateToolhelp32Snapshot failed: %w", err)
	}
	defer func() {
		_ = windows.CloseHandle(snapshot)
	}()

	var procEntry windows.ProcessEntry32
	procEntry.Size = uint32(unsafe.Sizeof(procEntry))

	err = windows.Process32First(snapshot, &procEntry)
	if err != nil {
		return nil, fmt.Errorf("Process32First failed: %w", err)
	}

	var processes []ProcessInfo
	for {
		processName := syscall.UTF16ToString(procEntry.ExeFile[:])
		if strings.EqualFold(processName, name) {
			processes = append(processes, ProcessInfo{
				PID:  procEntry.ProcessID,
				Name: processName,
			})
		}

		err = windows.Process32Next(snapshot, &procEntry)
		if err != nil {
			break
		}
	}

	return processes, nil
}

// IsProcessRunning checks if a process with the given name is running.
func IsProcessRunning(name string) (bool, error) {
	processes, err := FindProcessesByName(name)
	if err != nil {
		return false, err
	}
	return len(processes) > 0, nil
}
