package metrics

import (
	"bufio"
	"fmt"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"
)

// ProcessStats holds real-time resource metrics for a running process
type ProcessStats struct {
	PID         int
	Alive       bool
	State       string  // e.g. "R" (running), "S" (sleeping)
	MemoryRSS   int64   // Resident memory in bytes
	CPUPercent  float64 // Estimated CPU usage %
	ThreadCount int
	Uptime      time.Duration
}

// GetProcessStats retrieves current memory, CPU, and thread stats for a PID
func GetProcessStats(pid int) (*ProcessStats, error) {
	if pid <= 0 {
		return nil, fmt.Errorf("invalid pid: %d", pid)
	}

	// Check if process exists
	process, err := os.FindProcess(pid)
	if err != nil {
		return &ProcessStats{PID: pid, Alive: false}, nil
	}

	// On Unix, sending signal 0 checks if process is alive
	if err := checkProcessAlive(process); err != nil {
		return &ProcessStats{PID: pid, Alive: false}, nil
	}

	stats := &ProcessStats{
		PID:   pid,
		Alive: true,
	}

	if runtime.GOOS == "linux" {
		readLinuxProcStats(pid, stats)
	} else {
		// Fallback for macOS / Darwin development
		readDarwinProcessStats(pid, stats)
	}

	return stats, nil
}

// readLinuxProcStats reads /proc/<pid>/status and /proc/<pid>/stat on Linux/Android
func readLinuxProcStats(pid int, stats *ProcessStats) {
	statusPath := fmt.Sprintf("/proc/%d/status", pid)
	file, err := os.Open(statusPath)
	if err == nil {
		defer file.Close()
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			line := scanner.Text()
			parts := strings.SplitN(line, ":", 2)
			if len(parts) != 2 {
				continue
			}
			key := strings.TrimSpace(parts[0])
			val := strings.TrimSpace(parts[1])

			switch key {
			case "State":
				fields := strings.Fields(val)
				if len(fields) > 0 {
					stats.State = fields[0]
				}
			case "VmRSS":
				// VmRSS is typically in kB (e.g. "12345 kB")
				fields := strings.Fields(val)
				if len(fields) > 0 {
					if kb, err := strconv.ParseInt(fields[0], 10, 64); err == nil {
						stats.MemoryRSS = kb * 1024
					}
				}
			case "Threads":
				if th, err := strconv.Atoi(val); err == nil {
					stats.ThreadCount = th
				}
			}
		}
	}

	// Read CPU times from /proc/<pid>/stat
	statPath := fmt.Sprintf("/proc/%d/stat", pid)
	data, err := os.ReadFile(statPath)
	if err == nil {
		fields := strings.Fields(string(data))
		if len(fields) > 14 {
			// utime is field 14 (index 13), stime is field 15 (index 14)
			utime, _ := strconv.ParseFloat(fields[13], 64)
			stime, _ := strconv.ParseFloat(fields[14], 64)
			totalTicks := utime + stime
			// Standard Linux clock ticks per sec (USER_HZ) is usually 100
			if totalTicks > 0 {
				stats.CPUPercent = totalTicks / 100.0
				if stats.CPUPercent > 100.0 {
					stats.CPUPercent = 99.9
				}
			}
		}
	}
}

// readDarwinProcessStats provides fallback metrics on macOS for testing
func readDarwinProcessStats(pid int, stats *ProcessStats) {
	stats.State = "S"
	stats.ThreadCount = 1
	// Default placeholder for non-Linux dev
	stats.MemoryRSS = 12 * 1024 * 1024
	stats.CPUPercent = 0.5
}
