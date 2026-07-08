package collector

import (
	"testing"
)

func TestGPUCollectorName(t *testing.T) {
	c := NewGPUCollector()
	if c.Name() != gpuCollectorName {
		t.Errorf("Name() = %s, want %s", c.Name(), gpuCollectorName)
	}
}

func TestParseIOAcceleratorOutput(t *testing.T) {
	mockOutput := `+-o IOAccelerator  <class IOAccelerator, id 0x1000002d1>
  | "PerformanceStatistics" = {"Device Utilization %"=45.5,"Current Frequency"=1398,"vramUsedBytes"=536870912,"vramTotalBytes"=8589934592}`

	usage, freq, vramUsed, vramTotal, err := parseIOAcceleratorOutput(mockOutput)
	if err != nil {
		t.Fatal(err)
	}

	if usage != 45.5 {
		t.Errorf("usage = %v, want 45.5", usage)
	}
	if freq != 1398 {
		t.Errorf("freq = %v, want 1398", freq)
	}
	if vramUsed != 512 {
		t.Errorf("vramUsed = %v MB, want 512 MB", vramUsed)
	}
	if vramTotal != 8192 {
		t.Errorf("vramTotal = %v MB, want 8192 MB", vramTotal)
	}
}

func TestParseIOAcceleratorOutputPartial(t *testing.T) {
	mockOutput := `+-o IOAccelerator  <class IOAccelerator, id 0x1000002d1>
  | "PerformanceStatistics" = {"Device Utilization %"=12.3,"vramUsedBytes"=0}`

	usage, freq, vramUsed, vramTotal, err := parseIOAcceleratorOutput(mockOutput)
	if err != nil {
		t.Fatal(err)
	}

	if usage != 12.3 {
		t.Errorf("usage = %v, want 12.3", usage)
	}
	if freq != 0 {
		t.Errorf("freq = %v, want 0", freq)
	}
	if vramUsed != 0 {
		t.Errorf("vramUsed = %v, want 0", vramUsed)
	}
	if vramTotal != 0 {
		t.Errorf("vramTotal = %v, want 0", vramTotal)
	}
}

func TestGPUCollectorCollect(t *testing.T) {
	c := NewGPUCollector()

	if err := c.Collect(); err != nil {
		t.Skip("skipping: GPU not available: " + err.Error())
	}

	stats := c.Stats()
	if stats.UsagePercent < 0 || stats.UsagePercent > 100 {
		t.Errorf("UsagePercent out of range: %v", stats.UsagePercent)
	}
}
