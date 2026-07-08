package collector

import (
	"fmt"

	"github.com/1lent/peaktop/internal/apple"
	"github.com/1lent/peaktop/internal/types"
)

const cpuCollectorName = "cpu"

type CPUCollector struct {
	prevTicks     []apple.CPUTick
	prevTotalTick apple.CPUTick
	coreLabels    []string
	initialized   bool
	stats         types.CPUStats
}

func NewCPUCollector() *CPUCollector {
	return &CPUCollector{}
}

func (c *CPUCollector) Name() string {
	return cpuCollectorName
}

func (c *CPUCollector) Collect() error {
	info, err := apple.GetHostCPUInfo()
	if err != nil {
		return err
	}

	c.coreLabels = info.CoreLabels

	freqMHz := float64(info.FrequencyHz) / 1_000_000.0

	if !c.initialized {
		c.prevTicks = make([]apple.CPUTick, len(info.PerCoreTicks))
		copy(c.prevTicks, info.PerCoreTicks)
		c.prevTotalTick = sumTicks(info.PerCoreTicks)
		c.initialized = true
		c.stats = types.CPUStats{
			PerCore:      make(map[string]float64),
			FrequencyMHz: freqMHz,
		}
		return nil
	}

	totalDelta := tickDelta(sumTicks(info.PerCoreTicks), c.prevTotalTick)
	totalUsage := percentFromDelta(totalDelta)

	perCore := make(map[string]float64, len(info.PerCoreTicks))
	var eCoreTotal float64
	var eCoreCount int
	var pCoreTotal float64
	var pCoreCount int
	var sCoreTotal float64
	var sCoreCount int

	for i, current := range info.PerCoreTicks {
		if i >= len(c.prevTicks) {
			break
		}
		delta := tickDelta(current, c.prevTicks[i])
		usage := percentFromDelta(delta)

		coreName := fmt.Sprintf("C%d", i)
		if i < len(info.CoreLabels) {
			coreName = info.CoreLabels[i]
		}
		perCore[coreName] = usage

		if len(coreName) > 0 {
			switch coreName[0] {
			case 'E':
				eCoreTotal += usage
				eCoreCount++
			case 'P':
				pCoreTotal += usage
				pCoreCount++
			case 'S':
				sCoreTotal += usage
				sCoreCount++
			}
		}
	}

	copy(c.prevTicks, info.PerCoreTicks)
	c.prevTotalTick = sumTicks(info.PerCoreTicks)

	c.stats = types.CPUStats{
		UsagePercent: totalUsage,
		PerCore:      perCore,
		ECoreAvg:     safeDivide(eCoreTotal, float64(eCoreCount)),
		PCoreAvg:     safeDivide(pCoreTotal, float64(pCoreCount)),
		SCoreAvg:     safeDivide(sCoreTotal, float64(sCoreCount)),
		FrequencyMHz: freqMHz,
		CoreCount:    info.CoreCount,
		ECoreCount:   eCoreCount,
		PCoreCount:   pCoreCount,
	}

	return nil
}

func (c *CPUCollector) Stats() types.CPUStats {
	return c.stats
}

func sumTicks(ticks []apple.CPUTick) apple.CPUTick {
	var total apple.CPUTick
	for _, t := range ticks {
		total.User += t.User
		total.Sys += t.Sys
		total.Idle += t.Idle
		total.Nice += t.Nice
	}
	return total
}

func tickDelta(current, previous apple.CPUTick) apple.CPUTick {
	return apple.CPUTick{
		User: diffOrZero(current.User, previous.User),
		Sys:  diffOrZero(current.Sys, previous.Sys),
		Idle: diffOrZero(current.Idle, previous.Idle),
		Nice: diffOrZero(current.Nice, previous.Nice),
	}
}

func diffOrZero(current, previous uint32) uint32 {
	if current >= previous {
		return current - previous
	}
	return 0
}

func percentFromDelta(delta apple.CPUTick) float64 {
	total := delta.User + delta.Sys + delta.Idle + delta.Nice
	if total == 0 {
		return 0
	}
	return float64(delta.User+delta.Sys) / float64(total) * 100.0
}

func safeDivide(numerator, denominator float64) float64 {
	if denominator == 0 {
		return 0
	}
	return numerator / denominator
}
