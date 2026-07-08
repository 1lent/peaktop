package collector

import (
	"sync"

	"github.com/1lent/peaktop/internal/apple"
	"github.com/1lent/peaktop/internal/types"
)

const powerCollectorName = "power"

type PowerCollector struct {
	mu            sync.Mutex
	stats         types.PowerStats
	rootWarning   string
	warnedNoRoot  bool
	collecting    bool
}

func NewPowerCollector() *PowerCollector {
	return &PowerCollector{}
}

func (c *PowerCollector) Name() string {
	return powerCollectorName
}

func (c *PowerCollector) Collect() error {
	if !c.warnedNoRoot {
		result, err := apple.GetPowerBreakdown()
		if err != nil {
			return err
		}

		if !result.IsRoot {
			c.warnedNoRoot = true
			c.rootWarning = "Power data requires sudo (run with root to see power metrics)"
		}

		c.mu.Lock()
		c.stats = types.PowerStats{
			PackageWatts: result.PackageWatts,
			CPUWatts:     result.CPUWatts,
			GPUWatts:     result.GPUWatts,
			ANEWatts:     result.ANEWatts,
			DRAMWatts:    result.DRAMWatts,
		}
		c.mu.Unlock()
		return nil
	}

	if c.collecting {
		return nil
	}
	c.collecting = true

	go func() {
		defer func() { c.collecting = false }()

		result, err := apple.GetPowerBreakdown()
		if err != nil {
			return
		}

		if !result.IsRoot && !c.warnedNoRoot {
			c.warnedNoRoot = true
			c.rootWarning = "Power data requires sudo (run with root to see power metrics)"
		}

		c.mu.Lock()
		c.stats = types.PowerStats{
			PackageWatts: result.PackageWatts,
			CPUWatts:     result.CPUWatts,
			GPUWatts:     result.GPUWatts,
			ANEWatts:     result.ANEWatts,
			DRAMWatts:    result.DRAMWatts,
		}
		c.mu.Unlock()
	}()

	return nil
}

func (c *PowerCollector) Stats() types.PowerStats {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.stats
}

func (c *PowerCollector) RootWarning() string {
	return c.rootWarning
}
