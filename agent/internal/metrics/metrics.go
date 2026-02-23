// Package metrics reads CPU and memory usage for a process via /proc.
package metrics

import (
	"bytes"
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"
	"time"
)

// Result holds a point-in-time resource snapshot for one process.
type Result struct {
	CPUPercent float64   `json:"cpu_percent"`
	MemRSSMB   float64   `json:"mem_rss_mb"`
	PID        int       `json:"pid"`
	SampledAt  time.Time `json:"sampled_at"`
}

// Read samples CPU usage over 1 second and reads memory from /proc.
// CPU % is wall-clock percentage (can exceed 100% on multi-core systems).
func Read(pid int) (*Result, error) {
	t1, err := cpuTicks(pid)
	if err != nil {
		return nil, err
	}
	time.Sleep(time.Second)
	t2, err := cpuTicks(pid)
	if err != nil {
		return nil, err
	}

	// USER_HZ is 100 on all modern Linux — delta ticks over 1s = cpu %
	cpuPct := math.Round(float64(t2-t1)*10) / 10

	rssKB, err := memRSSKB(pid)
	if err != nil {
		return nil, err
	}

	return &Result{
		CPUPercent: cpuPct,
		MemRSSMB:   math.Round(float64(rssKB)/1024.0*10) / 10,
		PID:        pid,
		SampledAt:  time.Now(),
	}, nil
}

// cpuTicks returns utime+stime from /proc/{pid}/stat.
func cpuTicks(pid int) (uint64, error) {
	data, err := os.ReadFile(fmt.Sprintf("/proc/%d/stat", pid))
	if err != nil {
		return 0, fmt.Errorf("reading /proc/%d/stat: %w", pid, err)
	}
	// Format: pid (comm) state ppid ... utime stime ...
	// Find last ')' to skip the comm field which may contain spaces/parens.
	closeParen := bytes.LastIndexByte(data, ')')
	if closeParen < 0 {
		return 0, fmt.Errorf("unexpected /proc/%d/stat format", pid)
	}
	// Fields after ')': state(0) ppid(1) pgrp(2) session(3) tty(4) tpgid(5)
	//   flags(6) minflt(7) cminflt(8) majflt(9) cmajflt(10) utime(11) stime(12)
	fields := strings.Fields(string(data[closeParen+2:]))
	if len(fields) < 13 {
		return 0, fmt.Errorf("short /proc/%d/stat", pid)
	}
	utime, _ := strconv.ParseUint(fields[11], 10, 64)
	stime, _ := strconv.ParseUint(fields[12], 10, 64)
	return utime + stime, nil
}

// memRSSKB returns VmRSS in kB from /proc/{pid}/status.
func memRSSKB(pid int) (int64, error) {
	data, err := os.ReadFile(fmt.Sprintf("/proc/%d/status", pid))
	if err != nil {
		return 0, fmt.Errorf("reading /proc/%d/status: %w", pid, err)
	}
	for _, line := range strings.Split(string(data), "\n") {
		if strings.HasPrefix(line, "VmRSS:") {
			fields := strings.Fields(line)
			if len(fields) >= 2 {
				val, err := strconv.ParseInt(fields[1], 10, 64)
				if err == nil {
					return val, nil
				}
			}
		}
	}
	return 0, fmt.Errorf("VmRSS not found in /proc/%d/status", pid)
}
