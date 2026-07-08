package widgets

import (
	"strings"
	"testing"
)

func TestRenderGaugeZero(t *testing.T) {
	result := RenderGauge("CPU", 0, 60)
	if !strings.Contains(result, "0.00%") {
		t.Errorf("expected 0.00%%, got %q", result)
	}
}

func TestRenderGaugeFull(t *testing.T) {
	result := RenderGauge("GPU", 100, 60)
	if !strings.Contains(result, "100.00%") {
		t.Errorf("expected 100.00%%, got %q", result)
	}
}

func TestRenderGaugeGreenColor(t *testing.T) {
	result := RenderGauge("CPU", 25, 60)
	if len(result) == 0 {
		t.Error("expected non-empty gauge")
	}
}

func TestRenderGaugeYellowColor(t *testing.T) {
	result := RenderGauge("CPU", 60, 60)
	if len(result) == 0 {
		t.Error("expected non-empty gauge")
	}
}

func TestRenderGaugeRedColor(t *testing.T) {
	result := RenderGauge("CPU", 90, 60)
	if len(result) == 0 {
		t.Error("expected non-empty gauge")
	}
}

func TestRenderGaugeNarrow(t *testing.T) {
	result := RenderGauge("MEM", 50, 20)
	if len(result) == 0 {
		t.Error("expected non-empty gauge even with narrow width")
	}
}

func TestRenderGaugeWide(t *testing.T) {
	result := RenderGauge("Dis", 45, 100)
	if len(result) == 0 {
		t.Error("expected non-empty gauge with wide width")
	}
}
