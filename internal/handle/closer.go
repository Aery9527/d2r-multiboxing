// Package handle provides utilities for enumerating and manipulating
// Windows handles in remote processes.
package handle

import (
	"fmt"

	"golang.org/x/sys/windows"
)

// closeRemoteHandle closes a handle in a remote process.
func closeRemoteHandle(processID uint32, handle windows.Handle) error {
	processHandle, err := windows.OpenProcess(
		windows.PROCESS_DUP_HANDLE,
		false,
		processID,
	)
	if err != nil {
		return fmt.Errorf("failed to open process %d: %w", processID, err)
	}
	defer func() {
		_ = windows.CloseHandle(processHandle)
	}()

	// DuplicateCloseSource 會關閉來源 handle，等效於在目標進程中 CloseHandle()
	var duplicatedHandle windows.Handle
	err = ntDuplicateObject(
		processHandle,
		handle,
		0,
		&duplicatedHandle,
		0,
		0,
		DuplicateCloseSource,
	)
	if err != nil {
		return fmt.Errorf("failed to close handle 0x%X in process %d: %w", handle, processID, err)
	}

	return nil
}

// CloseHandlesByName finds and closes all handles matching the given name in a process.
func CloseHandlesByName(processID uint32, handleName string) (int, error) {
	handles, err := findHandlesByName(processID, handleName)
	if err != nil {
		return 0, fmt.Errorf("failed to find handles: %w", err)
	}

	if len(handles) == 0 {
		return 0, nil
	}

	closedCount := 0
	var lastError error
	for _, h := range handles {
		err := closeRemoteHandle(h.ProcessID, h.Handle)
		if err != nil {
			lastError = err
			continue
		}
		closedCount++
	}

	if closedCount == 0 && lastError != nil {
		return 0, fmt.Errorf("failed to close any handles: %w", lastError)
	}

	return closedCount, nil
}
