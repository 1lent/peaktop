package collector

import (
	"testing"
)

func TestANECollectorName(t *testing.T) {
	c := NewANECollector()
	if c.Name() != aneCollectorName {
		t.Errorf("Name() = %s, want %s", c.Name(), aneCollectorName)
	}
}

func TestANECollectorCollect(t *testing.T) {
	c := NewANECollector()

	err := c.Collect()
	if err != nil {
		t.Skip("skipping: ANE not available: " + err.Error())
	}

	stats := c.Stats()
	if stats.UsagePercent < 0 || stats.UsagePercent > 100 {
		t.Errorf("UsagePercent out of range: %v", stats.UsagePercent)
	}
}
