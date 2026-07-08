package collector

import (
	"strconv"
	"strings"

	"github.com/1lent/peaktop/internal/apple"
	"github.com/1lent/peaktop/internal/types"
)

const batteryCollectorName = "battery"

type BatteryCollector struct {
	stats types.BatteryStats
}

func NewBatteryCollector() *BatteryCollector {
	return &BatteryCollector{}
}

func (c *BatteryCollector) Name() string {
	return batteryCollectorName
}

func (c *BatteryCollector) Collect() error {
	percent, charging, cycleCount, maxCapacity, designCapacity, timeRemaining, hasBattery, err := apple.GetBatteryInfo()
	if err != nil {
		return err
	}

	if !hasBattery {
		c.stats = types.BatteryStats{IsPresent: false}
		return nil
	}

	timeMin := parseTimeRemaining(timeRemaining)

	c.stats = types.BatteryStats{
		Percent:        int(percent),
		IsCharging:     charging,
		CycleCount:     cycleCount,
		MaxCapacity:    maxCapacity,
		DesignCapacity: designCapacity,
		TimeRemaining:  timeMin,
		Watts:          0,
		IsPresent:      true,
	}

	return nil
}

func (c *BatteryCollector) Stats() types.BatteryStats {
	return c.stats
}

func (c *BatteryCollector) HasBattery() bool {
	return c.stats.IsPresent
}

func parseTimeRemaining(raw string) int {
	if raw == "" {
		return 0
	}

	raw = strings.ToLower(raw)
	raw = strings.ReplaceAll(raw, "h", " ")
	raw = strings.ReplaceAll(raw, "m", "")
	fields := strings.Fields(raw)

	if len(fields) == 0 {
		return 0
	}

	hours, errH := strconv.Atoi(fields[0])
	if errH != nil {
		return 0
	}

	mins := 0
	if len(fields) > 1 {
		mins, _ = strconv.Atoi(fields[1])
	}

	return hours*60 + mins
}
