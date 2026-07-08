package apple

import (
	"bytes"
	"context"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

func GetFanRPMs() []float64 {
	rpms := tryPowermetricsFans()
	if len(rpms) > 0 {
		return rpms
	}

	rpms = trySMCFans()
	if len(rpms) > 0 {
		return rpms
	}

	rpms = tryPmsetFans()
	if len(rpms) > 0 {
		return rpms
	}

	rpms = tryIoregFans()
	return rpms
}

func tryPowermetricsFans() []float64 {
	args := []string{"--samplers", "thermal", "-i", "200", "-n", "1"}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "/usr/bin/powermetrics", args...)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return nil
	}

	return parsePowermetricsFans(stdout.String())
}

func parsePowermetricsFans(output string) []float64 {
	var rpms []float64
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if !strings.Contains(line, "Fan") || !strings.Contains(line, "speed") {
			continue
		}
		fields := strings.Fields(line)
		for _, field := range fields {
			val, err := strconv.ParseFloat(field, 64)
			if err == nil && val > 0 && val < 20000 {
				rpms = append(rpms, val)
			}
		}
	}
	return rpms
}

func trySMCFans() []float64 {
	if err := ensureSMCConnection(); err != nil {
		return nil
	}

	var rpms []float64
	for i := 0; i < 2; i++ {
		rpm, err := GetFanRPMViaSMC(i)
		if err != nil || rpm <= 0 {
			break
		}
		rpms = append(rpms, rpm)
	}
	return rpms
}

func tryPmsetFans() []float64 {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "pmset", "-g", "thermlog")

	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = nil

	if err := cmd.Run(); err != nil {
		return nil
	}

	return parsePmsetFans(stdout.String())
}

func parsePmsetFans(output string) []float64 {
	var rpms []float64
	lines := strings.Split(output, "\n")

	for _, line := range lines {
		if !strings.Contains(line, "Fan") && !strings.Contains(line, "RPM") && !strings.Contains(line, "rpm") {
			continue
		}

		fields := strings.Fields(line)
		for _, field := range fields {
			val, err := strconv.ParseFloat(strings.TrimRight(field, "RrPpMm,;"), 64)
			if err == nil && val > 100 && val < 20000 {
				rpms = append(rpms, val)
			}
		}
	}

	return rpms
}

func tryIoregFans() []float64 {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "ioreg", "-c", "AppleSMC", "-r", "-d", "1")

	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = nil

	if err := cmd.Run(); err != nil {
		return nil
	}

	return parseIoregFans(stdout.String())
}

func parseIoregFans(output string) []float64 {
	var rpms []float64
	lines := strings.Split(output, "\n")

	for _, line := range lines {
		if !strings.Contains(line, "F") || !strings.Contains(line, "Ac") {
			continue
		}
		if strings.Contains(line, "Tg") {
			continue
		}

		fields := strings.Fields(line)
		for _, field := range fields {
			if field == "=" {
				continue
			}
			val, err := strconv.ParseFloat(strings.Trim(field, "\"= "), 64)
			if err == nil && val > 0 && val < 20000 {
				rpms = append(rpms, val)
			}
		}
	}

	return rpms
}
