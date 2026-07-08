package collector

import (
	"testing"

	"github.com/brodie/peaktop/internal/apple"
)

func TestTickDelta(t *testing.T) {
	tests := []struct {
		name     string
		current  apple.CPUTick
		previous apple.CPUTick
		want     apple.CPUTick
	}{
		{
			name:     "normal increment",
			current:  apple.CPUTick{User: 100, Sys: 50, Idle: 200, Nice: 0},
			previous: apple.CPUTick{User: 80, Sys: 40, Idle: 180, Nice: 0},
			want:     apple.CPUTick{User: 20, Sys: 10, Idle: 20, Nice: 0},
		},
		{
			name:     "counter wrap-around",
			current:  apple.CPUTick{User: 10, Sys: 0, Idle: 5, Nice: 0},
			previous: apple.CPUTick{User: 4294967290, Sys: 0, Idle: 0, Nice: 0},
			want:     apple.CPUTick{User: 0, Sys: 0, Idle: 5, Nice: 0},
		},
		{
			name:     "no change",
			current:  apple.CPUTick{User: 100, Sys: 50, Idle: 200, Nice: 0},
			previous: apple.CPUTick{User: 100, Sys: 50, Idle: 200, Nice: 0},
			want:     apple.CPUTick{User: 0, Sys: 0, Idle: 0, Nice: 0},
		},
		{
			name:     "zero previous",
			current:  apple.CPUTick{User: 100, Sys: 50, Idle: 200, Nice: 0},
			previous: apple.CPUTick{User: 0, Sys: 0, Idle: 0, Nice: 0},
			want:     apple.CPUTick{User: 100, Sys: 50, Idle: 200, Nice: 0},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tickDelta(tt.current, tt.previous)
			if got != tt.want {
				t.Errorf("tickDelta() = %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestPercentFromDelta(t *testing.T) {
	tests := []struct {
		name  string
		delta apple.CPUTick
		want  float64
	}{
		{
			name:  "half busy",
			delta: apple.CPUTick{User: 50, Sys: 0, Idle: 50, Nice: 0},
			want:  50.0,
		},
		{
			name:  "fully idle",
			delta: apple.CPUTick{User: 0, Sys: 0, Idle: 100, Nice: 0},
			want:  0.0,
		},
		{
			name:  "fully busy",
			delta: apple.CPUTick{User: 60, Sys: 40, Idle: 0, Nice: 0},
			want:  100.0,
		},
		{
			name:  "zero delta",
			delta: apple.CPUTick{User: 0, Sys: 0, Idle: 0, Nice: 0},
			want:  0.0,
		},
		{
			name:  "25 percent",
			delta: apple.CPUTick{User: 10, Sys: 15, Idle: 75, Nice: 0},
			want:  25.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := percentFromDelta(tt.delta)
			if got != tt.want {
				t.Errorf("percentFromDelta() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSumTicks(t *testing.T) {
	ticks := []apple.CPUTick{
		{User: 10, Sys: 5, Idle: 20, Nice: 0},
		{User: 20, Sys: 10, Idle: 30, Nice: 1},
		{User: 30, Sys: 15, Idle: 40, Nice: 2},
	}

	want := apple.CPUTick{User: 60, Sys: 30, Idle: 90, Nice: 3}
	got := sumTicks(ticks)

	if got != want {
		t.Errorf("sumTicks() = %+v, want %+v", got, want)
	}
}

func TestSafeDivide(t *testing.T) {
	tests := []struct {
		num, den float64
		want     float64
	}{
		{10, 2, 5},
		{0, 0, 0},
		{10, 0, 0},
		{0, 5, 0},
		{75, 3, 25},
	}

	for _, tt := range tests {
		got := safeDivide(tt.num, tt.den)
		if got != tt.want {
			t.Errorf("safeDivide(%v, %v) = %v, want %v", tt.num, tt.den, got, tt.want)
		}
	}
}

func TestDiffOrZero(t *testing.T) {
	tests := []struct {
		current, previous uint32
		want              uint32
	}{
		{100, 80, 20},
		{10, 20, 0},
		{0, 0, 0},
		{4294967295, 4294967294, 1},
	}

	for _, tt := range tests {
		got := diffOrZero(tt.current, tt.previous)
		if got != tt.want {
			t.Errorf("diffOrZero(%d, %d) = %d, want %d", tt.current, tt.previous, got, tt.want)
		}
	}
}

func TestCPUCollectorName(t *testing.T) {
	c := NewCPUCollector()
	if c.Name() != cpuCollectorName {
		t.Errorf("Name() = %s, want %s", c.Name(), cpuCollectorName)
	}
}

func TestCPUCollectorFirstCollectInitializes(t *testing.T) {
	c := NewCPUCollector()

	if err := c.Collect(); err != nil {
		t.Skip("skipping: requires macOS host_processor_info: " + err.Error())
	}

	if !c.initialized {
		t.Error("expected collector to be initialized after first Collect()")
	}

	stats := c.Stats()
	if stats.UsagePercent != 0 {
		t.Error("first Collect() should return zero usage (no delta yet)")
	}
}

func TestCPUCollectorSecondCollectHasUsage(t *testing.T) {
	c := NewCPUCollector()

	if err := c.Collect(); err != nil {
		t.Skip("skipping: requires macOS: " + err.Error())
	}
	if err := c.Collect(); err != nil {
		t.Fatal(err)
	}

	stats := c.Stats()
	if stats.UsagePercent < 0 || stats.UsagePercent > 100 {
		t.Errorf("UsagePercent out of range: %v", stats.UsagePercent)
	}

	hasE := false
	hasP := false
	for name := range stats.PerCore {
		if len(name) > 0 && name[0] == 'E' {
			hasE = true
		}
		if len(name) > 0 && name[0] == 'P' {
			hasP = true
		}
	}
	if !hasE && !hasP {
		t.Errorf("no E or P cores found in labels: %v", stats.PerCore)
	}
}
