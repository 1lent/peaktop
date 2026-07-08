package alert

import (
	"fmt"
	"time"

	"github.com/brodie/peaktop/internal/types"
)

const defaultCooldown = 30 * time.Second
const sustainedCooldown = 10 * time.Second

type AlertLevel = string

const (
	LevelInfo     AlertLevel = "info"
	LevelWarning  AlertLevel = "warning"
	LevelCritical AlertLevel = "critical"
)

type alertRule struct {
	source       string
	field        string
	threshold    float64
	above        bool
	level        AlertLevel
	cooldown     time.Duration
	sustained    time.Duration
	lastFired    time.Time
	sustainedSince time.Time
}

type AlertEngine struct {
	rules []alertRule
}

func NewAlertEngine() *AlertEngine {
	engine := &AlertEngine{}
	engine.rules = engine.defaultRules()
	return engine
}

func (ae *AlertEngine) defaultRules() []alertRule {
	return []alertRule{
		{
			source:    "thermal",
			field:     "thermal.cpu_temp",
			threshold: 90,
			above:     true,
			level:     LevelCritical,
			cooldown:  defaultCooldown,
		},
		{
			source:    "thermal",
			field:     "thermal.gpu_temp",
			threshold: 85,
			above:     true,
			level:     LevelCritical,
			cooldown:  defaultCooldown,
		},
		{
			source:    "thermal",
			field:     "thermal.pressure",
			threshold: 3,
			above:     true,
			level:     LevelCritical,
			cooldown:  defaultCooldown,
		},
		{
			source:    "thermal",
			field:     "thermal.pressure",
			threshold: 2,
			above:     true,
			level:     LevelWarning,
			cooldown:  defaultCooldown,
		},
		{
			source:    "battery",
			field:     "battery.percent",
			threshold: 10,
			above:     false,
			level:     LevelWarning,
			cooldown:  defaultCooldown,
		},
		{
			source:    "memory",
			field:     "memory.pressure",
			threshold: 90,
			above:     true,
			level:     LevelWarning,
			cooldown:  defaultCooldown,
		},
		{
			source:    "gpu",
			field:     "gpu.usage",
			threshold: 95,
			above:     true,
			level:     LevelInfo,
			cooldown:  defaultCooldown,
			sustained: sustainedCooldown,
		},
	}
}

func (ae *AlertEngine) Check(cpuStats types.CPUStats, gpuStats types.GPUStats, memStats types.MemoryStats, thermalStats types.ThermalStats, batteryStats types.BatteryStats) []types.AlertEvent {
	now := time.Now()
	payload := ae.collectPayload(cpuStats, gpuStats, memStats, thermalStats, batteryStats)

	var events []types.AlertEvent
	for i := range ae.rules {
		rule := &ae.rules[i]
		if event, ok := ae.evaluateRule(rule, payload, now); ok {
			events = append(events, event)
		}
	}

	return events
}

func (ae *AlertEngine) collectPayload(cpu types.CPUStats, gpu types.GPUStats, mem types.MemoryStats, thermal types.ThermalStats, battery types.BatteryStats) map[string]float64 {
	payload := map[string]float64{
		"thermal.cpu_temp": thermal.CputempC,
		"thermal.gpu_temp": thermal.GPUTempC,
		"thermal.pressure": thermalPressureToFloat(thermal.Pressure),
		"cpu.usage":        cpu.UsagePercent,
		"gpu.usage":        gpu.UsagePercent,
		"memory.pressure":  float64(mem.PressurePercent),
		"battery.percent":  float64(battery.Percent),
	}
	return payload
}

func thermalPressureToFloat(pressure string) float64 {
	switch pressure {
	case "Nominal":
		return 0
	case "Fair":
		return 1
	case "Serious":
		return 2
	case "Critical":
		return 3
	default:
		return 0
	}
}

func (ae *AlertEngine) evaluateRule(rule *alertRule, payload map[string]float64, now time.Time) (types.AlertEvent, bool) {
	value, exists := payload[rule.field]
	if !exists {
		return types.AlertEvent{}, false
	}

	isTriggered := false
	if rule.above {
		isTriggered = value >= rule.threshold
	} else {
		isTriggered = value <= rule.threshold
	}

	if !isTriggered && rule.sustainedSince.IsZero() {
		return types.AlertEvent{}, false
	}

	if rule.sustained > 0 {
		if isTriggered {
			if rule.sustainedSince.IsZero() {
				rule.sustainedSince = now
				return types.AlertEvent{}, false
			}
			if now.Sub(rule.sustainedSince) < rule.sustained {
				return types.AlertEvent{}, false
			}
		} else {
			rule.sustainedSince = time.Time{}
			return types.AlertEvent{}, false
		}
	}

	if !rule.lastFired.IsZero() && now.Sub(rule.lastFired) < rule.cooldown {
		return types.AlertEvent{}, false
	}

	rule.lastFired = now
	rule.sustainedSince = time.Time{}

	message := ae.buildMessage(rule, value)

	return types.AlertEvent{
		Timestamp: now.Format(time.RFC3339),
		Level:     rule.level,
		Source:    rule.source,
		Message:   message,
	}, true
}

func (ae *AlertEngine) buildMessage(rule *alertRule, value float64) string {
	direction := "above"
	if !rule.above {
		direction = "below"
	}
	return fmt.Sprintf("%s %s is %.1f (%s threshold %.0f)",
		rule.source, rule.field, value, direction, rule.threshold)
}

func (ae *AlertEngine) RulesCount() int {
	return len(ae.rules)
}
