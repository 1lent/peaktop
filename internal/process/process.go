package process

import (
	"sort"
	"time"

	"github.com/brodie/peaktop/internal/apple"
	"github.com/brodie/peaktop/internal/types"
)

const (
	processCollectorName = "process"
	maxProcesses         = 50
)

type ProcessList struct {
	processes    []types.ProcessInfo
	prevTicks    map[int32]uint64
	prevTime     time.Time
	initialized  bool
	totalRAM     uint64
}

func NewProcessList() *ProcessList {
	return &ProcessList{
		prevTicks: make(map[int32]uint64),
	}
}

func (p *ProcessList) Name() string {
	return processCollectorName
}

func (p *ProcessList) Collect() error {
	pids, err := apple.ListPIDs()
	if err != nil {
		return err
	}

	now := time.Now()
	elapsedSec := 1.0
	if p.initialized {
		elapsedSec = now.Sub(p.prevTime).Seconds()
		if elapsedSec < 0.001 {
			elapsedSec = 0.001
		}
	}

	currentTicks := make(map[int32]uint64, len(pids))
	var allProcesses []procEntry
	var totalDeltaTicks uint64

	for _, pid := range pids {
		if pid == 0 {
			continue
		}

		info, err := apple.GetProcInfo(pid)
		if err != nil {
			continue
		}

		if info.Name == "kernel_task" {
			continue
		}

		currentTicks[info.PID] = info.TotalTicks

		var cpuPercent float64
		if p.initialized {
			prevTicks, exists := p.prevTicks[info.PID]
			if exists {
				delta := safeDelta(info.TotalTicks, prevTicks)
				totalDeltaTicks += delta
				cpuPercent = float64(delta) / elapsedSec * 100.0 / 1e9
			}
		}

		memPercent := p.computeMemPercent(info.MemoryBytes)

		allProcesses = append(allProcesses, procEntry{
			pid:        info.PID,
			name:       info.Name,
			cpuPercent: cpuPercent,
			memPercent: memPercent,
			memBytes:   info.MemoryBytes,
		})
	}

	_ = totalDeltaTicks

	sort.Slice(allProcesses, func(i, j int) bool {
		return allProcesses[i].cpuPercent > allProcesses[j].cpuPercent
	})

	limit := len(allProcesses)
	if limit > maxProcesses {
		limit = maxProcesses
	}

	processes := make([]types.ProcessInfo, limit)
	for i := 0; i < limit; i++ {
		entry := allProcesses[i]
		processes[i] = types.ProcessInfo{
			PID:        entry.pid,
			Name:       entry.name,
			CPUPercent: entry.cpuPercent,
			MemPercent: entry.memPercent,
		}
	}

	p.processes = processes
	p.prevTicks = currentTicks
	p.prevTime = now
	p.initialized = true

	return nil
}

func (p *ProcessList) List() []types.ProcessInfo {
	return p.processes
}

type procEntry struct {
	pid        int32
	name       string
	cpuPercent float64
	memPercent float64
	memBytes   uint64
}

func safeDelta(current, previous uint64) uint64 {
	if current >= previous {
		return current - previous
	}
	return 0
}

func (p *ProcessList) computeMemPercent(memBytes uint64) float64 {
	if p.totalRAM == 0 {
		vm, err := apple.GetVMStats()
		if err == nil && vm.TotalBytes > 0 {
			p.totalRAM = vm.TotalBytes
		}
	}
	if p.totalRAM == 0 {
		return 0
	}
	return float64(memBytes) / float64(p.totalRAM) * 100.0
}
