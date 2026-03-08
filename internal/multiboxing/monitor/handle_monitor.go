package monitor

import (
	"sync"
	"time"

	"d2rhl/internal/common/d2r"
	"d2rhl/internal/common/process"
	"d2rhl/internal/multiboxing/launcher"
)

var startHandleMonitorOnce sync.Once

func StartHandleMonitor() {
	startHandleMonitorOnce.Do(func() {
		go runHandleMonitor()
	})
}

func runHandleMonitor() {
	processedPIDs := make(map[uint32]bool)

	for {
		time.Sleep(2 * time.Second)

		d2rProcesses, err := process.FindProcessesByName(d2r.ProcessName)
		if err != nil {
			continue
		}

		activePIDs := make(map[uint32]bool)
		for _, p := range d2rProcesses {
			activePIDs[p.PID] = true
		}
		for pid := range processedPIDs {
			if !activePIDs[pid] {
				delete(processedPIDs, pid)
			}
		}

		for _, p := range d2rProcesses {
			if processedPIDs[p.PID] {
				continue
			}
			processedPIDs[p.PID] = true
			_, _ = launcher.CloseHandlesByName(p.PID, d2r.SingleInstanceEventName)
		}
	}
}
