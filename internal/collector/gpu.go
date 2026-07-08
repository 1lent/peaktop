package collector

import (
	"bytes"
	"fmt"
	"os/exec"
	"strconv"
	"strings"

	"github.com/1lent/peaktop/internal/apple"
	"github.com/1lent/peaktop/internal/types"
)

const gpuCollectorName = "gpu"

type GPUCollector struct {
	stats types.GPUStats
}

func NewGPUCollector() *GPUCollector {
	return &GPUCollector{}
}

func (c *GPUCollector) Name() string {
	return gpuCollectorName
}

func (c *GPUCollector) Collect() error {
	usage, freq, vramUsed, vramTotal, err := collectGPUStats()
	if err != nil {
		return err
	}

	c.stats = types.GPUStats{
		UsagePercent: usage,
		ActiveMHz:    freq,
		VRAMUsedMB:   vramUsed,
		VRAMTotalMB:  vramTotal,
	}
	return nil
}

func (c *GPUCollector) Stats() types.GPUStats {
	return c.stats
}

func collectGPUStats() (usage float64, freq float64, vramUsedMB uint64, vramTotalMB uint64, err error) {
	if apple.IsIOReportAvailable() {
		return collectGPUFromIOReport()
	}

	stats, err := apple.GetGPUStats()
	if err == nil {
		return stats.UsagePercent, stats.ActiveMHz, stats.VRAMUsedMB, stats.VRAMTotalMB, nil
	}

	return collectGPUFromIoreg()
}

func collectGPUFromIOReport() (float64, float64, uint64, uint64, error) {
	channels, err := apple.GetIOReportPower()
	if err != nil {
		return 0, 0, 0, 0, fmt.Errorf("IOReport GPU: %w", err)
	}

	var gpuUsage float64
	for _, ch := range channels {
		if ch.Group == "GPU" {
			gpuUsage = ch.Value
			break
		}
	}

	stats, err := apple.GetGPUStats()
	if err != nil {
		return gpuUsage, 0, 0, 0, nil
	}

	return gpuUsage, stats.ActiveMHz, stats.VRAMUsedMB, stats.VRAMTotalMB, nil
}

func collectGPUFromIoreg() (float64, float64, uint64, uint64, error) {
	cmd := exec.Command("ioreg", "-c", "IOAccelerator", "-r", "-d", "1")
	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = nil

	if err := cmd.Run(); err != nil {
		return 0, 0, 0, 0, fmt.Errorf("ioreg GPU: %w", err)
	}

	return parseIOAcceleratorOutput(stdout.String())
}

func parseIOAcceleratorOutput(output string) (float64, float64, uint64, uint64, error) {
	var usage float64
	var freq float64
	var vramUsed, vramTotal uint64

	statStart := strings.Index(output, "PerformanceStatistics")
	if statStart < 0 {
		return 0, 0, 0, 0, nil
	}

	rest := output[statStart:]

	braceStart := strings.Index(rest, "{")
	if braceStart < 0 {
		return 0, 0, 0, 0, nil
	}

	braceEnd := strings.Index(rest[braceStart:], "}")
	if braceEnd < 0 {
		return 0, 0, 0, 0, nil
	}

	statsBlock := rest[braceStart+1 : braceStart+braceEnd]

	usage = extractFloatFromIoreg(statsBlock, "Device Utilization %")
	freq = extractFloatFromIoreg(statsBlock, "Current Frequency")
	vramUsed = extractUint64FromIoreg(statsBlock, "vramUsedBytes") / (1024 * 1024)
	vramTotal = extractUint64FromIoreg(statsBlock, "vramTotalBytes") / (1024 * 1024)

	return usage, freq, vramUsed, vramTotal, nil
}

func extractFloatFromIoreg(block, key string) float64 {
	idx := strings.Index(block, key)
	if idx < 0 {
		return 0
	}

	rest := block[idx+len(key):]
	eqIdx := strings.Index(rest, "=")
	if eqIdx < 0 {
		return 0
	}

	valStr := strings.TrimLeft(rest[eqIdx+1:], " ")
	end := strings.IndexAny(valStr, ",}")
	if end < 0 {
		end = len(valStr)
	}
	valStr = strings.TrimSpace(valStr[:end])
	valStr = strings.Trim(valStr, "\"")

	val, err := strconv.ParseFloat(valStr, 64)
	if err != nil {
		return 0
	}
	return val
}

func extractUint64FromIoreg(block, key string) uint64 {
	idx := strings.Index(block, key)
	if idx < 0 {
		return 0
	}

	rest := block[idx+len(key):]
	eqIdx := strings.Index(rest, "=")
	if eqIdx < 0 {
		return 0
	}

	valStr := strings.TrimLeft(rest[eqIdx+1:], " ")
	end := strings.IndexAny(valStr, ",}")
	if end < 0 {
		end = len(valStr)
	}
	valStr = strings.TrimSpace(valStr[:end])
	valStr = strings.Trim(valStr, "\"")

	val, err := strconv.ParseUint(valStr, 10, 64)
	if err != nil {
		return 0
	}
	return val
}
