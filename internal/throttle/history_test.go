package throttle

import (
	"strings"
	"testing"
)

func TestNewRingBuffer(t *testing.T) {
	rb := NewRingBuffer(0)
	if rb.size != defaultBufferSize {
		t.Errorf("size = %d, want %d", rb.size, defaultBufferSize)
	}

	rb2 := NewRingBuffer(10)
	if rb2.size != 10 {
		t.Errorf("size = %d, want 10", rb2.size)
	}
}

func TestPushAndValues(t *testing.T) {
	rb := NewRingBuffer(5)

	rb.Push(1.0)
	rb.Push(2.0)
	rb.Push(3.0)

	values := rb.Values()
	if len(values) != 3 {
		t.Fatalf("len = %d, want 3", len(values))
	}
	if values[0] != 1.0 || values[1] != 2.0 || values[2] != 3.0 {
		t.Errorf("values = %v, want [1 2 3]", values)
	}
}

func TestRingBufferWrap(t *testing.T) {
	rb := NewRingBuffer(4)

	for i := 0; i < 6; i++ {
		rb.Push(float64(i))
	}

	values := rb.Values()
	if len(values) != 4 {
		t.Fatalf("len = %d, want 4", len(values))
	}

	want := []float64{2, 3, 4, 5}
	for i, v := range values {
		if v != want[i] {
			t.Errorf("values[%d] = %v, want %v", i, v, want[i])
		}
	}
}

func TestEmptyBufferSparkline(t *testing.T) {
	rb := NewRingBuffer(10)
	result := rb.Sparkline(5)
	if result != "" {
		t.Errorf("sparkline = %q, want empty", result)
	}
}

func TestSparklineSingleValue(t *testing.T) {
	rb := NewRingBuffer(10)
	rb.Push(50.0)

	result := rb.Sparkline(1)
	if len([]rune(result)) != 1 {
		t.Fatalf("len = %d, want 1", len([]rune(result)))
	}
}

func TestSparklineAllSameValues(t *testing.T) {
	rb := NewRingBuffer(10)
	for i := 0; i < 5; i++ {
		rb.Push(42.0)
	}

	result := rb.Sparkline(5)
	// All same values: all should be normalized to 0 → lowest char
	expected := strings.Repeat(string(sparkChars[0]), 5)
	if result != expected {
		t.Errorf("sparkline = %q, want %q", result, expected)
	}
}

func TestSparklineNormalization(t *testing.T) {
	rb := NewRingBuffer(10)
	rb.Push(0.0)
	rb.Push(50.0)
	rb.Push(100.0)

	result := rb.Sparkline(3)
	runes := []rune(result)
	if len(runes) != 3 {
		t.Fatalf("len = %d, want 3", len(runes))
	}

	if runes[0] != sparkChars[0] {
		t.Errorf("first rune = %q, want %q", runes[0], sparkChars[0])
	}
	if runes[2] != sparkChars[len(sparkChars)-1] {
		t.Errorf("last rune = %q, want %q", runes[2], sparkChars[len(sparkChars)-1])
	}
}

func TestLen(t *testing.T) {
	rb := NewRingBuffer(10)

	if rb.Len() != 0 {
		t.Errorf("Len() = %d, want 0", rb.Len())
	}

	rb.Push(1.0)
	if rb.Len() != 1 {
		t.Errorf("Len() = %d, want 1", rb.Len())
	}

	for i := 0; i < 20; i++ {
		rb.Push(float64(i))
	}
	if rb.Len() != 10 {
		t.Errorf("Len() = %d, want 10", rb.Len())
	}
}

func TestSparklineWidthLargerThanData(t *testing.T) {
	rb := NewRingBuffer(10)
	rb.Push(0.0)
	rb.Push(1.0)

	result := rb.Sparkline(5)
	if len([]rune(result)) != 5 {
		t.Errorf("len = %d, want 5", len([]rune(result)))
	}
}

func TestNormalizeValue(t *testing.T) {
	tests := []struct {
		value, min, max float64
		want            float64
	}{
		{50, 0, 100, 0.5},
		{0, 0, 100, 0},
		{100, 0, 100, 1.0},
		{5, 5, 5, 0},
		{0, 0, 0, 0},
		{7, 0, 10, 0.7},
	}

	for _, tt := range tests {
		got := normalizeValue(tt.value, tt.min, tt.max)
		if !floatNear(got, tt.want, 0.001) {
			t.Errorf("normalizeValue(%v, %v, %v) = %v, want %v",
				tt.value, tt.min, tt.max, got, tt.want)
		}
	}
}

func floatNear(a, b, epsilon float64) bool {
	diff := a - b
	if diff < 0 {
		diff = -diff
	}
	return diff < epsilon
}
