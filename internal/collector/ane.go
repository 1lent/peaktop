package collector

import (
	"fmt"

	"github.com/1lent/peaktop/internal/apple"
	"github.com/1lent/peaktop/internal/types"
)

const aneCollectorName = "ane"

type ANECollector struct {
	stats       types.GPUStats
	firstError  error
	warnedNoANE bool
}

func NewANECollector() *ANECollector {
	return &ANECollector{}
}

func (c *ANECollector) Name() string {
	return aneCollectorName
}

func (c *ANECollector) Collect() error {
	usage, err := collectANEUtilization()
	if err != nil {
		if !c.warnedNoANE {
			c.warnedNoANE = true
			c.firstError = fmt.Errorf("ANE unavailable: %w", err)
		}
		usage = 0
	}

	c.stats = types.GPUStats{
		UsagePercent: usage,
	}
	return nil
}

func (c *ANECollector) Stats() types.GPUStats {
	return c.stats
}

func (c *ANECollector) FirstError() error {
	return c.firstError
}

func collectANEUtilization() (float64, error) {
	if apple.IsIOReportAvailable() {
		usage, err := apple.GetIOReportANE()
		if err == nil {
			return usage, nil
		}
	}

	stats, err := apple.GetANEStats()
	if err != nil {
		return 0, err
	}
	return stats.UsagePercent, nil
}
