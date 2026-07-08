package alert

import (
	"testing"
	"time"

	"github.com/brodie/peaktop/internal/types"
)

func TestNewAlertEngine(t *testing.T) {
	ae := NewAlertEngine()
	count := ae.RulesCount()
	if count == 0 {
		t.Error("expected non-zero default rules")
	}
}

func TestCPUTempCritical(t *testing.T) {
	ae := NewAlertEngine()
	thermal := types.ThermalStats{CputempC: 92}

	events := ae.Check(
		types.CPUStats{},
		types.GPUStats{},
		types.MemoryStats{},
		thermal,
		types.BatteryStats{},
	)

	if len(events) == 0 {
		t.Error("expected critical alert for CPU temp > 90")
	}
	found := false
	for _, e := range events {
		if e.Level == LevelCritical && e.Source == "thermal" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected thermal critical alert, got %v", events)
	}
}

func TestCPUTempBelowThreshold(t *testing.T) {
	ae := NewAlertEngine()
	thermal := types.ThermalStats{CputempC: 50}

	events := ae.Check(
		types.CPUStats{},
		types.GPUStats{},
		types.MemoryStats{},
		thermal,
		types.BatteryStats{},
	)

	for _, e := range events {
		if e.Source == "thermal" && e.Level == LevelCritical {
			t.Errorf("unexpected thermal alert for temp 50: %v", e)
		}
	}
}

func TestBatteryWarning(t *testing.T) {
	ae := NewAlertEngine()
	battery := types.BatteryStats{Percent: 8, IsPresent: true}

	events := ae.Check(
		types.CPUStats{},
		types.GPUStats{},
		types.MemoryStats{},
		types.ThermalStats{},
		battery,
	)

	found := false
	for _, e := range events {
		if e.Source == "battery" && e.Level == LevelWarning {
			found = true
		}
	}
	if !found {
		t.Error("expected battery warning for 8%")
	}
}

func TestBatteryAboveThreshold(t *testing.T) {
	ae := NewAlertEngine()
	battery := types.BatteryStats{Percent: 85, IsPresent: true}

	events := ae.Check(
		types.CPUStats{},
		types.GPUStats{},
		types.MemoryStats{},
		types.ThermalStats{},
		battery,
	)

	for _, e := range events {
		if e.Source == "battery" {
			t.Errorf("unexpected battery alert for 85%%: %v", e)
		}
	}
}

func TestMemoryPressureWarning(t *testing.T) {
	ae := NewAlertEngine()
	mem := types.MemoryStats{PressurePercent: 95}

	events := ae.Check(
		types.CPUStats{},
		types.GPUStats{},
		mem,
		types.ThermalStats{},
		types.BatteryStats{},
	)

	found := false
	for _, e := range events {
		if e.Source == "memory" && e.Level == LevelWarning {
			found = true
		}
	}
	if !found {
		t.Error("expected memory pressure warning for 95%")
	}
}

func TestMemoryPressureNormal(t *testing.T) {
	ae := NewAlertEngine()
	mem := types.MemoryStats{PressurePercent: 30}

	events := ae.Check(
		types.CPUStats{},
		types.GPUStats{},
		mem,
		types.ThermalStats{},
		types.BatteryStats{},
	)

	for _, e := range events {
		if e.Source == "memory" {
			t.Errorf("unexpected memory alert for 30%%: %v", e)
		}
	}
}

func TestThermalPressureWarning(t *testing.T) {
	ae := NewAlertEngine()
	thermal := types.ThermalStats{Pressure: "Serious"}

	events := ae.Check(
		types.CPUStats{},
		types.GPUStats{},
		types.MemoryStats{},
		thermal,
		types.BatteryStats{},
	)

	hasWarning := false
	hasCritical := false
	for _, e := range events {
		if e.Source == "thermal" && e.Level == LevelWarning {
			hasWarning = true
		}
		if e.Source == "thermal" && e.Level == LevelCritical {
			hasCritical = true
		}
	}
	if !hasWarning {
		t.Error("expected thermal warning for Serious pressure")
	}
	if hasCritical {
		t.Error("did not expect critical for Serious (level 2)")
	}
}

func TestThermalPressureCritical(t *testing.T) {
	ae := NewAlertEngine()
	thermal := types.ThermalStats{Pressure: "Critical"}

	events := ae.Check(
		types.CPUStats{},
		types.GPUStats{},
		types.MemoryStats{},
		thermal,
		types.BatteryStats{},
	)

	hasCritical := false
	for _, e := range events {
		if e.Source == "thermal" && e.Level == LevelCritical {
			hasCritical = true
		}
	}
	if !hasCritical {
		t.Error("expected thermal critical for Critical pressure")
	}
}

func TestCooldownPreventsReFire(t *testing.T) {
	ae := NewAlertEngine()

	thermal := types.ThermalStats{CputempC: 95}

	events1 := ae.Check(
		types.CPUStats{},
		types.GPUStats{},
		types.MemoryStats{},
		thermal,
		types.BatteryStats{},
	)

	countFirst := 0
	for _, e := range events1 {
		if e.Source == "thermal" && e.Level == LevelCritical {
			countFirst++
		}
	}

	events2 := ae.Check(
		types.CPUStats{},
		types.GPUStats{},
		types.MemoryStats{},
		thermal,
		types.BatteryStats{},
	)

	countSecond := 0
	for _, e := range events2 {
		if e.Source == "thermal" && e.Level == LevelCritical {
			countSecond++
		}
	}

	if countFirst == 0 {
		t.Error("expected alert on first check")
	}
	if countSecond != 0 {
		t.Errorf("expected cooldown to suppress second alert, got %d", countSecond)
	}
}

func TestCooldownExpiry(t *testing.T) {
	ae := NewAlertEngine()

	thermal := types.ThermalStats{CputempC: 95}

	ae.Check(
		types.CPUStats{},
		types.GPUStats{},
		types.MemoryStats{},
		thermal,
		types.BatteryStats{},
	)

	for i := range ae.rules {
		if ae.rules[i].source == "thermal" && ae.rules[i].field == "thermal.cpu_temp" {
			ae.rules[i].lastFired = time.Now().Add(-defaultCooldown - time.Second)
			ae.rules[i].sustainedSince = time.Time{}
		}
	}

	events := ae.Check(
		types.CPUStats{},
		types.GPUStats{},
		types.MemoryStats{},
		thermal,
		types.BatteryStats{},
	)

	found := false
	for _, e := range events {
		if e.Source == "thermal" && e.Level == LevelCritical {
			found = true
		}
	}
	if !found {
		t.Error("expected alert after cooldown expired")
	}
}

func TestGPUSustainedAlert(t *testing.T) {
	ae := NewAlertEngine()
	gpu := types.GPUStats{UsagePercent: 98}

	events := ae.Check(
		types.CPUStats{},
		gpu,
		types.MemoryStats{},
		types.ThermalStats{},
		types.BatteryStats{},
	)

	for _, e := range events {
		if e.Source == "gpu" && e.Level == LevelInfo {
			t.Errorf("GPU sustained alert should not fire on first check: %v", e)
		}
	}
}

func TestAlertEventHasFields(t *testing.T) {
	ae := NewAlertEngine()
	thermal := types.ThermalStats{CputempC: 92}

	events := ae.Check(
		types.CPUStats{},
		types.GPUStats{},
		types.MemoryStats{},
		thermal,
		types.BatteryStats{},
	)

	if len(events) == 0 {
		t.Fatal("expected events")
	}

	for _, e := range events {
		if e.Timestamp == "" {
			t.Error("timestamp empty")
		}
		if e.Level == "" {
			t.Error("level empty")
		}
		if e.Source == "" {
			t.Error("source empty")
		}
		if e.Message == "" {
			t.Error("message empty")
		}
	}
}
