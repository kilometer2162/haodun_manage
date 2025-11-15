package utils

import (
	"fmt"
	"math"
	"runtime"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/mem"
)

type MemoryStats struct {
	Total   uint64  `json:"total"`
	Used    uint64  `json:"used"`
	Free    uint64  `json:"free"`
	Percent float64 `json:"percent"`
}

type CPUStats struct {
	Cores       int     `json:"cores"`
	Usage       float64 `json:"usage"`
	User        float64 `json:"user"`
	System      float64 `json:"system"`
	Idle        float64 `json:"idle"`
	Description string  `json:"description"`
}

type DiskStats struct {
	Path    string  `json:"path"`
	Total   uint64  `json:"total"`
	Used    uint64  `json:"used"`
	Free    uint64  `json:"free"`
	Percent float64 `json:"percent"`
}

type SystemMetrics struct {
	Memory     MemoryStats `json:"memory"`
	CPU        CPUStats    `json:"cpu"`
	Disk       DiskStats   `json:"disk"`
	UpdateTime time.Time   `json:"update_time"`
}

func GetSystemMetrics() (*SystemMetrics, error) {
	memStats, err := collectMemoryStats()
	if err != nil {
		return nil, err
	}

	cpuStats, err := collectCPUStats()
	if err != nil {
		return nil, err
	}

	diskStats, err := collectDiskStats("/")
	if err != nil {
		return nil, err
	}

	return &SystemMetrics{
		Memory:     memStats,
		CPU:        cpuStats,
		Disk:       diskStats,
		UpdateTime: time.Now(),
	}, nil
}

func collectMemoryStats() (MemoryStats, error) {
	vm, err := mem.VirtualMemory()
	if err != nil {
		return MemoryStats{}, err
	}

	return MemoryStats{
		Total:   vm.Total,
		Used:    vm.Used,
		Free:    vm.Available,
		Percent: roundToTwoDecimals(vm.UsedPercent),
	}, nil
}

func collectCPUStats() (CPUStats, error) {
	snapshot1, err := cpu.Times(false)
	if err != nil || len(snapshot1) == 0 {
		return cpuFallback(err)
	}

	time.Sleep(200 * time.Millisecond)

	snapshot2, err := cpu.Times(false)
	if err != nil || len(snapshot2) == 0 {
		return cpuFallback(err)
	}

	t1 := snapshot1[0]
	t2 := snapshot2[0]

	total1 := totalCPUTime(t1)
	total2 := totalCPUTime(t2)
	totalDiff := total2 - total1
	if totalDiff <= 0 {
		return cpuFallback(fmt.Errorf("invalid cpu time delta"))
	}

	idleDiff := (t2.Idle + t2.Iowait) - (t1.Idle + t1.Iowait)
	userDiff := (t2.User + t2.Nice) - (t1.User + t1.Nice)
	systemDiff := (t2.System + t2.Irq + t2.Softirq) - (t1.System + t1.Irq + t1.Softirq)

	usage := (totalDiff - idleDiff) / totalDiff * 100

	return CPUStats{
		Cores:       runtime.NumCPU(),
		Usage:       roundToTwoDecimals(usage),
		User:        roundToTwoDecimals(userDiff / totalDiff * 100),
		System:      roundToTwoDecimals(systemDiff / totalDiff * 100),
		Idle:        roundToTwoDecimals(idleDiff / totalDiff * 100),
		Description: runtime.GOOS,
	}, nil
}

func cpuFallback(err error) (CPUStats, error) {
	desc := runtime.GOOS
	if err != nil {
		desc = fmt.Sprintf("%s fallback: %v", runtime.GOOS, err)
	} else {
		desc = fmt.Sprintf("%s fallback", runtime.GOOS)
	}

	return CPUStats{
		Cores:       runtime.NumCPU(),
		Usage:       0,
		User:        0,
		System:      0,
		Idle:        0,
		Description: desc,
	}, nil
}

func totalCPUTime(t cpu.TimesStat) float64 {
	return t.User + t.System + t.Idle + t.Nice + t.Iowait + t.Irq + t.Softirq + t.Steal + t.Guest + t.GuestNice
}

func collectDiskStats(path string) (DiskStats, error) {
	usage, err := disk.Usage(path)
	if err != nil {
		return DiskStats{}, err
	}

	return DiskStats{
		Path:    usage.Path,
		Total:   usage.Total,
		Used:    usage.Used,
		Free:    usage.Free,
		Percent: roundToTwoDecimals(usage.UsedPercent),
	}, nil
}

func roundToTwoDecimals(val float64) float64 {
	return math.Round(val*100) / 100
}
