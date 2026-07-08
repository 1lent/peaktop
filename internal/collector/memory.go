package collector

import (
	"github.com/brodie/peaktop/internal/apple"
	"github.com/brodie/peaktop/internal/types"
)

const memoryCollectorName = "memory"

type MemoryCollector struct {
	stats types.MemoryStats
}

func NewMemoryCollector() *MemoryCollector {
	return &MemoryCollector{}
}

func (c *MemoryCollector) Name() string {
	return memoryCollectorName
}

func (c *MemoryCollector) Collect() error {
	vm, err := apple.GetVMStats()
	if err != nil {
		return err
	}

	swap, err := apple.GetSwapUsage()
	if err != nil {
		return err
	}

	pressure, _ := apple.GetVMPressure()

	usedBytes := vm.WiredBytes + vm.ActiveBytes + vm.CompressedBytes
	pressurePercent := computePressurePercent(vm)

	c.stats = types.MemoryStats{
		TotalBytes:      vm.TotalBytes,
		UsedBytes:       usedBytes,
		FreeBytes:       vm.FreeBytes + vm.InactiveBytes,
		WiredBytes:      vm.WiredBytes,
		CompressedBytes: vm.CompressedBytes,
		SwapTotalBytes:  swap.TotalBytes,
		SwapUsedBytes:   swap.UsedBytes,
		PressurePercent: pressurePercent,
	}

	_ = pressure

	return nil
}

func (c *MemoryCollector) Stats() types.MemoryStats {
	return c.stats
}

func computePressurePercent(vm apple.VMStats) int {
	if vm.TotalBytes == 0 {
		return 0
	}
	active := vm.WiredBytes + vm.ActiveBytes + vm.CompressedBytes
	pct := float64(active) / float64(vm.TotalBytes) * 100.0
	return int(pct)
}
