// Package metrics reads CPU, memory, and system resource usage via /proc and syscall.
package metrics

import (
	"bytes"
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"
	"syscall"
	"time"
)

// Result holds a point-in-time resource snapshot for one process.
type Result struct {
	CPUPercent float64   `json:"cpu_percent"`
	MemRSSMB   float64   `json:"mem_rss_mb"`
	VmPeakMB   float64   `json:"vm_peak_mb"`
	PID        int       `json:"pid"`
	SampledAt  time.Time `json:"sampled_at"`
}

// SystemMetrics holds VM-level resource metrics.
type SystemMetrics struct {
	MemTotalMB    float64   `json:"mem_total_mb"`
	MemUsedMB     float64   `json:"mem_used_mb"`
	MemFreeMB     float64   `json:"mem_free_mb"`
	LoadAvg1      float64   `json:"load_avg_1"`
	LoadAvg5      float64   `json:"load_avg_5"`
	LoadAvg15     float64   `json:"load_avg_15"`
	UptimeSeconds float64   `json:"uptime_seconds"`
	DiskTotalGB   float64   `json:"disk_total_gb"`
	DiskUsedGB    float64   `json:"disk_used_gb"`
	DiskFreeGB    float64   `json:"disk_free_gb"`
	SampledAt     time.Time `json:"sampled_at"`
}

// Read samples CPU usage over 1 second and reads memory from /proc.
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

	cpuPct := math.Round(float64(t2-t1)*10) / 10

	rssKB, peakKB, err := procStatus(pid)
	if err != nil {
		return nil, err
	}

	return &Result{
		CPUPercent: cpuPct,
		MemRSSMB:   math.Round(float64(rssKB)/1024.0*10) / 10,
		VmPeakMB:   math.Round(float64(peakKB)/1024.0*10) / 10,
		PID:        pid,
		SampledAt:  time.Now(),
	}, nil
}

// ReadSystem reads VM-level metrics from /proc and syscall.
func ReadSystem() (*SystemMetrics, error) {
	mem, err := sysMemory()
	if err != nil {
		return nil, err
	}
	load, err := loadAverage()
	if err != nil {
		return nil, err
	}
	uptime, err := sysUptime()
	if err != nil {
		return nil, err
	}
	disk, err := diskUsage("/")
	if err != nil {
		return nil, err
	}

	return &SystemMetrics{
		MemTotalMB:    mem[0],
		MemUsedMB:     mem[1],
		MemFreeMB:     mem[2],
		LoadAvg1:      load[0],
		LoadAvg5:      load[1],
		LoadAvg15:     load[2],
		UptimeSeconds: uptime,
		DiskTotalGB:   disk[0],
		DiskUsedGB:    disk[1],
		DiskFreeGB:    disk[2],
		SampledAt:     time.Now(),
	}, nil
}

// --- per-process helpers ---

func cpuTicks(pid int) (uint64, error) {
	data, err := os.ReadFile(fmt.Sprintf("/proc/%d/stat", pid))
	if err != nil {
		return 0, fmt.Errorf("reading /proc/%d/stat: %w", pid, err)
	}
	closeParen := bytes.LastIndexByte(data, ')')
	if closeParen < 0 {
		return 0, fmt.Errorf("unexpected /proc/%d/stat format", pid)
	}
	// Fields after ')': state(0) ppid(1) ... utime(11) stime(12)
	fields := strings.Fields(string(data[closeParen+2:]))
	if len(fields) < 13 {
		return 0, fmt.Errorf("short /proc/%d/stat", pid)
	}
	utime, _ := strconv.ParseUint(fields[11], 10, 64)
	stime, _ := strconv.ParseUint(fields[12], 10, 64)
	return utime + stime, nil
}

// procStatus returns VmRSS and VmPeak in kB from /proc/{pid}/status.
func procStatus(pid int) (rssKB, peakKB int64, err error) {
	data, err := os.ReadFile(fmt.Sprintf("/proc/%d/status", pid))
	if err != nil {
		return 0, 0, fmt.Errorf("reading /proc/%d/status: %w", pid, err)
	}
	for _, line := range strings.Split(string(data), "\n") {
		if strings.HasPrefix(line, "VmRSS:") {
			if f := strings.Fields(line); len(f) >= 2 {
				rssKB, _ = strconv.ParseInt(f[1], 10, 64)
			}
		} else if strings.HasPrefix(line, "VmPeak:") {
			if f := strings.Fields(line); len(f) >= 2 {
				peakKB, _ = strconv.ParseInt(f[1], 10, 64)
			}
		}
	}
	if rssKB == 0 {
		return 0, 0, fmt.Errorf("VmRSS not found for pid %d", pid)
	}
	return rssKB, peakKB, nil
}

// --- system helpers ---

// sysMemory returns [totalMB, usedMB, freeMB] from /proc/meminfo.
func sysMemory() ([3]float64, error) {
	data, err := os.ReadFile("/proc/meminfo")
	if err != nil {
		return [3]float64{}, fmt.Errorf("reading /proc/meminfo: %w", err)
	}
	var totalKB, availKB int64
	for _, line := range strings.Split(string(data), "\n") {
		if strings.HasPrefix(line, "MemTotal:") {
			if f := strings.Fields(line); len(f) >= 2 {
				totalKB, _ = strconv.ParseInt(f[1], 10, 64)
			}
		} else if strings.HasPrefix(line, "MemAvailable:") {
			if f := strings.Fields(line); len(f) >= 2 {
				availKB, _ = strconv.ParseInt(f[1], 10, 64)
			}
		}
	}
	totalMB := math.Round(float64(totalKB)/1024.0*10) / 10
	freeMB := math.Round(float64(availKB)/1024.0*10) / 10
	usedMB := math.Round((totalMB-freeMB)*10) / 10
	return [3]float64{totalMB, usedMB, freeMB}, nil
}

// loadAverage returns [1m, 5m, 15m] from /proc/loadavg.
func loadAverage() ([3]float64, error) {
	data, err := os.ReadFile("/proc/loadavg")
	if err != nil {
		return [3]float64{}, fmt.Errorf("reading /proc/loadavg: %w", err)
	}
	fields := strings.Fields(string(data))
	if len(fields) < 3 {
		return [3]float64{}, fmt.Errorf("unexpected /proc/loadavg format")
	}
	var vals [3]float64
	for i := range vals {
		vals[i], _ = strconv.ParseFloat(fields[i], 64)
	}
	return vals, nil
}

// sysUptime returns uptime in seconds from /proc/uptime.
func sysUptime() (float64, error) {
	data, err := os.ReadFile("/proc/uptime")
	if err != nil {
		return 0, fmt.Errorf("reading /proc/uptime: %w", err)
	}
	fields := strings.Fields(string(data))
	if len(fields) < 1 {
		return 0, fmt.Errorf("unexpected /proc/uptime format")
	}
	return strconv.ParseFloat(fields[0], 64)
}

// diskUsage returns [totalGB, usedGB, freeGB] for the given mount point.
func diskUsage(path string) ([3]float64, error) {
	var stat syscall.Statfs_t
	if err := syscall.Statfs(path, &stat); err != nil {
		return [3]float64{}, fmt.Errorf("statfs %s: %w", path, err)
	}
	bsize := float64(stat.Bsize)
	totalGB := math.Round(float64(stat.Blocks)*bsize/1e9*10) / 10
	freeGB := math.Round(float64(stat.Bavail)*bsize/1e9*10) / 10
	usedGB := math.Round((totalGB-freeGB)*10) / 10
	return [3]float64{totalGB, usedGB, freeGB}, nil
}
