package apple

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

const thermalPressureSysctl = "kern.thermalpressure"

func GetThermalPressure() (int, error) {
	pressure := readSysctlInt(thermalPressureSysctl)
	if pressure < 0 {
		return 0, fmt.Errorf("sysctl %s returned %d", thermalPressureSysctl, pressure)
	}
	return pressure, nil
}

func GetTemperatures() (cpuTemp float64, gpuTemp float64, err error) {
	args := []string{
		"--samplers", "thermal",
		"-i", "200",
		"-n", "1",
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "/usr/bin/powermetrics", args...)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		stderrStr := stderr.String()
		if strings.Contains(stderrStr, "requires root") ||
			strings.Contains(stderrStr, "permission denied") ||
			strings.Contains(stderrStr, "Operation not permitted") ||
			strings.Contains(stderrStr, "must be invoked as the superuser") {
			return 0, 0, nil
		}
		return 0, 0, nil
	}

	cpu, gpu := parseThermalOutput(stdout.String())
	return cpu, gpu, nil
}

func parseThermalOutput(output string) (float64, float64) {
	var cpuTemp, gpuTemp float64
	lines := strings.Split(output, "\n")

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		switch {
		case strings.Contains(trimmed, "CPU die temperature"):
			cpuTemp = extractTempValue(trimmed)
		case strings.Contains(trimmed, "GPU die temperature"):
			gpuTemp = extractTempValue(trimmed)
		}
	}

	return cpuTemp, gpuTemp
}

func extractTempValue(line string) float64 {
	fields := strings.Fields(line)
	for i, field := range fields {
		if i+1 < len(fields) {
			if strings.HasPrefix(fields[i+1], "C") || strings.HasPrefix(fields[i+1], "F") {
				val, err := strconv.ParseFloat(field, 64)
				if err != nil {
					return 0
				}
				return val
			}
		}
	}
	return 0
}

func ThermalPressureString(pressure int) string {
	switch pressure {
	case 0:
		return "Nominal"
	case 1:
		return "Fair"
	case 2:
		return "Serious"
	case 3:
		return "Critical"
	default:
		return "Unknown"
	}
}
