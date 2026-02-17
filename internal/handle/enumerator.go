//nolint:govet // unsafe pointer operations required for Windows API interop with variable-length structures
package handle

import (
	"fmt"
	"strings"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

// HandleInfo represents information about a handle.
type HandleInfo struct {
	ProcessID uint32
	Handle    windows.Handle
	Name      string
	TypeName  string
}

// findHandlesByName finds handles by name in a specific process.
func findHandlesByName(processID uint32, targetName string) ([]HandleInfo, error) {
	var matchedHandles []HandleInfo

	bufferSize := uint32(1024 * 1024) // 1MB 初始大小
	var buffer []byte
	var returnLength uint32

	for {
		buffer = make([]byte, bufferSize)
		err := ntQuerySystemInformation(
			SystemExtendedHandleInformation,
			uintptr(unsafe.Pointer(&buffer[0])),
			bufferSize,
			&returnLength,
		)
		if err == nil {
			break
		}
		if errno, ok := err.(syscall.Errno); ok && errno == StatusInfoLengthMismatch {
			bufferSize = returnLength + 1024*1024
			continue
		}
		return nil, fmt.Errorf("NtQuerySystemInformation failed: %w", err)
	}

	handleInfo := (*SystemExtendedHandleInformationEx)(unsafe.Pointer(&buffer[0]))
	numberOfHandles := int(handleInfo.NumberOfHandles)

	firstEntryPtr := (*SystemHandleTableEntryInfoEx)(unsafe.Pointer(&handleInfo.Handles[0]))
	entries := unsafe.Slice(firstEntryPtr, numberOfHandles)

	processHandle, err := windows.OpenProcess(
		windows.PROCESS_DUP_HANDLE,
		false,
		processID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to open process %d: %w", processID, err)
	}
	defer func() {
		_ = windows.CloseHandle(processHandle)
	}()

	currentProcess := windows.CurrentProcess()

	for i := 0; i < numberOfHandles; i++ {
		entry := &entries[i]

		if uint32(entry.UniqueProcessID) != processID {
			continue
		}

		var duplicatedHandle windows.Handle
		err = ntDuplicateObject(
			processHandle,
			windows.Handle(entry.HandleValue),
			currentProcess,
			&duplicatedHandle,
			0,
			0,
			DuplicateSameAccess,
		)
		if err != nil {
			continue
		}

		typeName := queryObjectType(duplicatedHandle)

		// 僅對 Event 類型查詢名稱，避免對 pipe 等 handle 查詢時 hang 住
		var name string
		if typeName == "Event" {
			name = queryObjectName(duplicatedHandle)
		}

		_ = windows.CloseHandle(duplicatedHandle)

		if name != "" && strings.Contains(name, targetName) {
			matchedHandles = append(matchedHandles, HandleInfo{
				ProcessID: processID,
				Handle:    windows.Handle(entry.HandleValue),
				Name:      name,
				TypeName:  typeName,
			})
		}
	}

	return matchedHandles, nil
}

// queryObjectType queries the type name of a handle.
func queryObjectType(handle windows.Handle) string {
	alignedBuf := make([]uint64, 128)
	buffer := unsafe.Slice((*byte)(unsafe.Pointer(&alignedBuf[0])), len(alignedBuf)*8)
	var returnLength uint32

	err := ntQueryObject(
		handle,
		ObjectTypeInformation,
		uintptr(unsafe.Pointer(&buffer[0])),
		uint32(len(buffer)),
		&returnLength,
	)
	if err != nil {
		return ""
	}

	typeInfo := (*ObjectTypeInfo)(unsafe.Pointer(&buffer[0]))
	return getUnicodeString(&typeInfo.TypeName)
}

// queryObjectName queries the name of a handle.
func queryObjectName(handle windows.Handle) string {
	alignedBuf := make([]uint64, 512)
	buffer := unsafe.Slice((*byte)(unsafe.Pointer(&alignedBuf[0])), len(alignedBuf)*8)
	var returnLength uint32

	err := ntQueryObject(
		handle,
		ObjectNameInformation,
		uintptr(unsafe.Pointer(&buffer[0])),
		uint32(len(buffer)),
		&returnLength,
	)
	if err != nil {
		if errno, ok := err.(syscall.Errno); ok && errno == StatusInfoLengthMismatch {
			requiredUint64s := (returnLength + 7) / 8
			alignedBuf = make([]uint64, requiredUint64s)
			buffer = unsafe.Slice((*byte)(unsafe.Pointer(&alignedBuf[0])), len(alignedBuf)*8)

			err = ntQueryObject(
				handle,
				ObjectNameInformation,
				uintptr(unsafe.Pointer(&buffer[0])),
				uint32(len(buffer)),
				&returnLength,
			)
			if err != nil {
				return ""
			}
		} else {
			return ""
		}
	}

	nameInfo := (*ObjectNameInfo)(unsafe.Pointer(&buffer[0]))
	return getUnicodeString(&nameInfo.Name)
}
