package widgets

import (
	"testing"
)

func TestRenderSparklineEmpty(t *testing.T) {
	result := RenderSparkline(nil, 10, 3)
	if result != "" {
		t.Errorf("expected empty, got %q", result)
	}
}

func TestRenderSparklineSingleValue(t *testing.T) {
	result := RenderSparkline([]float64{50}, 5, 3)
	if len(result) == 0 {
		t.Error("expected non-empty sparkline")
	}
}

func TestRenderSparklineAllSame(t *testing.T) {
	data := []float64{5, 5, 5, 5, 5}
	result := RenderSparkline(data, 5, 3)
	if len(result) == 0 {
		t.Error("expected non-empty sparkline")
	}
}

func TestRenderSparklineRising(t *testing.T) {
	data := []float64{0, 25, 50, 75, 100}
	result := RenderSparkline(data, 5, 3)
	if len(result) == 0 {
		t.Error("expected non-empty sparkline")
	}
}

func TestRenderSparklineWidthLarger(t *testing.T) {
	data := []float64{10, 20, 30}
	result := RenderSparkline(data, 10, 3)
	if len([]rune(result)) > 10 {
		t.Errorf("expected at most 10 chars, got %d", len([]rune(result)))
	}
}

func TestRenderSparklineWidthSmaller(t *testing.T) {
	data := []float64{10, 20, 30, 40, 50, 60, 70, 80, 90, 100}
	result := RenderSparkline(data, 5, 3)
	runes := []rune(result)
	if len(runes) > 5 {
		t.Errorf("expected at most 5 chars, got %d", len(runes))
	}
}

func TestSparkMinMax(t *testing.T) {
	tests := []struct {
		data     []float64
		wantMin  float64
		wantMax  float64
	}{
		{nil, 0, 1},
		{[]float64{}, 0, 1},
		{[]float64{5}, 5, 5},
		{[]float64{1, 2, 3, 4, 5}, 1, 5},
		{[]float64{0, 100, 50}, 0, 100},
		{[]float64{-10, 0, 10}, -10, 10},
	}

	for _, tt := range tests {
		gotMin, gotMax := sparkMinMax(tt.data)
		if gotMin != tt.wantMin || gotMax != tt.wantMax {
			t.Errorf("sparkMinMax(%v) = (%v, %v), want (%v, %v)",
				tt.data, gotMin, gotMax, tt.wantMin, tt.wantMax)
		}
	}
}
