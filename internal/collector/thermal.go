package collector

import (
	"sync"

	"github.com/brodie/peaktop/internal/apple"
	"github.com/brodie/peaktop/internal/types"
)

const thermalCollectorName = "thermal"

type ThermalCollector struct {
	mu         sync.Mutex
	stats      types.ThermalStats
	collecting bool
}

func NewThermalCollector() *ThermalCollector {
	return &ThermalCollector{}
}

func (c *ThermalCollector) Name() string {
	return thermalCollectorName
}

func (c *ThermalCollector) Collect() error {
	pressure, err := apple.GetThermalPressure()
	if err != nil {
		c.mu.Lock()
		c.stats.Pressure = "N/A"
		c.stats.FanRPMs = apple.GetFanRPMs()
		c.mu.Unlock()
		return nil
	}

	c.mu.Lock()
	c.stats.Pressure = apple.ThermalPressureString(pressure)
	c.stats.FanRPMs = apple.GetFanRPMs()
	c.mu.Unlock()

	if c.collecting {
		return nil
	}
	c.collecting = true

	go func() {
		defer func() { c.collecting = false }()

		cpuTemp, gpuTemp, _ := apple.GetTemperatures()

		c.mu.Lock()
		c.stats.CputempC = cpuTemp
		c.stats.GPUTempC = gpuTemp
		c.mu.Unlock()
	}()

	return nil
}

func (c *ThermalCollector) Stats() types.ThermalStats {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.stats
}

func (c *ThermalCollector) IsThrottling() bool {
	pressure, err := apple.GetThermalPressure()
	if err != nil {
		return false
	}
	return pressure >= 2
}
