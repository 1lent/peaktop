package collector

import (
	"testing"

	"github.com/1lent/peaktop/internal/apple"
)

func TestMemoryCollectorName(t *testing.T) {
	c := NewMemoryCollector()
	if c.Name() != memoryCollectorName {
		t.Errorf("Name() = %s, want %s", c.Name(), memoryCollectorName)
	}
}

func TestComputePressurePercent(t *testing.T) {
	tests := []struct {
		name string
		vm   apple.VMStats
		want int
	}{
		{
			name: "50% pressure",
			vm: apple.VMStats{
				TotalBytes:      16384 * 1024 * 1024,
				WiredBytes:      2048 * 1024 * 1024,
				ActiveBytes:     4096 * 1024 * 1024,
				CompressedBytes: 2048 * 1024 * 1024,
			},
			want: 50,
		},
		{
			name: "zero total",
			vm: apple.VMStats{
				TotalBytes:      0,
				WiredBytes:      1024 * 1024 * 1024,
				ActiveBytes:     1024 * 1024 * 1024,
				CompressedBytes: 1024 * 1024 * 1024,
			},
			want: 0,
		},
		{
			name: "idle system",
			vm: apple.VMStats{
				TotalBytes:      16384 * 1024 * 1024,
				WiredBytes:      1024 * 1024 * 1024,
				ActiveBytes:     512 * 1024 * 1024,
				CompressedBytes: 0,
			},
			want: 9,
		},
		{
			name: "near full",
			vm: apple.VMStats{
				TotalBytes:      16384 * 1024 * 1024,
				WiredBytes:      4096 * 1024 * 1024,
				ActiveBytes:     8192 * 1024 * 1024,
				CompressedBytes: 3072 * 1024 * 1024,
			},
			want: 93,
		},
		{
			name: "empty system",
			vm: apple.VMStats{
				TotalBytes:      8192 * 1024 * 1024,
				WiredBytes:      0,
				ActiveBytes:     0,
				CompressedBytes: 0,
			},
			want: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := computePressurePercent(tt.vm)
			if got != tt.want {
				t.Errorf("computePressurePercent() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestMemoryCollectorCollect(t *testing.T) {
	c := NewMemoryCollector()

	if err := c.Collect(); err != nil {
		t.Skip("skipping: memory collect failed: " + err.Error())
	}

	stats := c.Stats()

	if stats.TotalBytes == 0 {
		t.Error("TotalBytes should not be zero")
	}
	if stats.PressurePercent < 0 || stats.PressurePercent > 100 {
		t.Errorf("PressurePercent out of range: %d", stats.PressurePercent)
	}
}
