package apple

import (
	"bytes"
	"context"
	"encoding/xml"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

const (
	powermetricsPath   = "/usr/bin/powermetrics"
	powermetricsSample = 200
	powermetricsCount  = 1
)

type PowerStats struct {
	PackageWatts float64
	CPUWatts     float64
	GPUWatts     float64
	ANEWatts     float64
	DRAMWatts    float64
	IsRoot       bool
}

func GetPowerBreakdown() (PowerStats, error) {
	if !isPowermetricsAvailable() {
		return PowerStats{}, nil
	}

	args := []string{
		"--samplers", "cpu_power,gpu_power,ane_power",
		"-i", strconv.Itoa(powermetricsSample),
		"-n", strconv.Itoa(powermetricsCount),
		"--format", "plist",
	}

	output, err := runPowermetrics(args)
	if err != nil || output == "" {
		return PowerStats{}, nil
	}

	stats := parsePowerPlist(output)
	stats.IsRoot = true
	if stats.hasAnyData() {
		return stats, nil
	}

	stats = parsePowermetricsText(output)
	stats.IsRoot = true
	return stats, nil
}

func isPowermetricsAvailable() bool {
	_, err := os.Stat(powermetricsPath)
	return err == nil
}

func isPermissionError(stderr string) bool {
	return strings.Contains(stderr, "requires root") ||
		strings.Contains(stderr, "permission denied") ||
		strings.Contains(stderr, "Operation not permitted") ||
		strings.Contains(stderr, "must be invoked as the superuser")
}

func runPowermetrics(args []string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, powermetricsPath, args...)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		stderrStr := stderr.String()
		if isPermissionError(stderrStr) {
			return "", nil
		}
		if ctx.Err() != nil {
			return "", nil
		}
		return "", nil
	}

	return stdout.String(), nil
}

func (s PowerStats) hasAnyData() bool {
	return s.CPUWatts > 0 || s.GPUWatts > 0 || s.PackageWatts > 0 || s.ANEWatts > 0 || s.DRAMWatts > 0
}

func parsePowerPlist(data string) PowerStats {
	stats := PowerStats{}

	type plistValue struct {
		XMLName xml.Name
		Content string `xml:",chardata"`
	}

	type plistEntry struct {
		Key   string      `xml:"key"`
		Value *plistValue `xml:"real"`
	}

	decoder := xml.NewDecoder(strings.NewReader(data))

	var (
		inArray   bool
		inDict    bool
		currentKey string
	)

	for {
		tok, err := decoder.Token()
		if err != nil {
			break
		}

		switch elem := tok.(type) {
		case xml.StartElement:
			switch elem.Name.Local {
			case "array":
				inArray = true
			case "dict":
				inDict = true
			}
		case xml.EndElement:
			switch elem.Name.Local {
			case "array":
				inArray = false
			case "dict":
				inDict = false
			}
		case xml.CharData:
			if !inArray || !inDict {
				continue
			}
			text := strings.TrimSpace(string(elem))
			if text == "" {
				continue
			}

			if currentKey == "" {
				currentKey = text
			} else {
				val, err := strconv.ParseFloat(text, 64)
				if err == nil {
					switch currentKey {
					case "cpu_power", "CPU Power":
						stats.CPUWatts = val
					case "gpu_power", "GPU Power":
						stats.GPUWatts = val
					case "ane_power", "ANE Power":
						stats.ANEWatts = val
					case "package_power", "Package Power", "combined_power":
						stats.PackageWatts = val
					case "dram_power", "DRAM Power":
						stats.DRAMWatts = val
					}
				}
				currentKey = ""
			}
		}
		_ = plistEntry{}
	}

	return stats
}

func parsePowermetricsText(text string) PowerStats {
	stats := PowerStats{}
	lines := strings.Split(text, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		switch {
		case strings.Contains(line, "CPU Power:"):
			stats.CPUWatts = extractWatts(line)
		case strings.Contains(line, "GPU Power:"):
			stats.GPUWatts = extractWatts(line)
		case strings.Contains(line, "ANE Power:"):
			stats.ANEWatts = extractWatts(line)
		case strings.Contains(line, "Package Power:"):
			stats.PackageWatts = extractWatts(line)
		case strings.Contains(line, "DRAM Power:"):
			stats.DRAMWatts = extractWatts(line)
		}
	}

	return stats
}

func extractWatts(line string) float64 {
	fields := strings.Fields(line)
	for i, field := range fields {
		if i+1 >= len(fields) {
			continue
		}
		if strings.HasSuffix(field, "mW") {
			val, err := strconv.ParseFloat(fields[i-1], 64)
			if err != nil {
				continue
			}
			return val / 1000.0
		}
		if strings.HasSuffix(field, "W") && field != "Power:" && field != "W" {
			val, err := strconv.ParseFloat(fields[i-1], 64)
			if err != nil {
				continue
			}
			return val
		}
	}
	return 0
}
