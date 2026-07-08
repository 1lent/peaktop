package log

import (
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

const (
	logDir         = ".peaktop"
	logSubDir      = "logs"
	filePermission = 0600
	dirPermission  = 0700
)

type CSVExporter struct {
	file *os.File
	writer *csv.Writer
	mu   sync.Mutex
	path string
}

func NewCSVExporter() (*CSVExporter, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("home dir: %w", err)
	}

	logPath := filepath.Join(homeDir, logDir, logSubDir)
	if err := os.MkdirAll(logPath, dirPermission); err != nil {
		return nil, fmt.Errorf("mkdir %s: %w", logPath, err)
	}

	filename := fmt.Sprintf("%s.csv", time.Now().Format("2006-01-02"))
	fullPath := filepath.Join(logPath, filename)

	file, err := os.OpenFile(fullPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, filePermission)
	if err != nil {
		return nil, fmt.Errorf("open %s: %w", fullPath, err)
	}

	exporter := &CSVExporter{
		file:   file,
		writer: csv.NewWriter(file),
		path:   fullPath,
	}

	stat, err := file.Stat()
	if err != nil {
		file.Close()
		return nil, fmt.Errorf("stat: %w", err)
	}

	if stat.Size() == 0 {
		exporter.writeHeader()
	}

	return exporter, nil
}

func NewNoopExporter() *CSVExporter {
	return &CSVExporter{}
}

func (e *CSVExporter) IsEnabled() bool {
	return e.file != nil
}

func (e *CSVExporter) writeHeader() {
	header := []string{
		"timestamp",
		"cpu_percent",
		"gpu_percent",
		"ane_percent",
		"power_package_w",
		"power_cpu_w",
		"power_gpu_w",
		"power_ane_w",
		"mem_used_bytes",
		"mem_pressure_pct",
		"swap_used_bytes",
		"thermal_pressure",
		"cpu_temp_c",
		"gpu_temp_c",
		"fan_rpm_0",
		"fan_rpm_1",
		"battery_pct",
		"battery_charging",
		"network_rx_bps",
		"network_tx_bps",
	}
	e.writer.Write(header)
	e.writer.Flush()
}

func (e *CSVExporter) WriteRow(
	cpuPct float64,
	gpuPct float64,
	anePct float64,
	powerPkgW float64,
	powerCPUW float64,
	powerGPUW float64,
	powerANEW float64,
	memUsedBytes uint64,
	memPressurePct int,
	swapUsedBytes uint64,
	thermalPressure string,
	cpuTemp float64,
	gpuTemp float64,
	fanRPMs []float64,
	batteryPct int,
	batteryCharging bool,
	networkRxBps float64,
	networkTxBps float64,
) {
	if !e.IsEnabled() {
		return
	}

	e.mu.Lock()
	defer e.mu.Unlock()

	fan0 := 0.0
	if len(fanRPMs) > 0 {
		fan0 = fanRPMs[0]
	}
	fan1 := 0.0
	if len(fanRPMs) > 1 {
		fan1 = fanRPMs[1]
	}

	row := []string{
		time.Now().Format(time.RFC3339),
		fmt.Sprintf("%.2f", cpuPct),
		fmt.Sprintf("%.2f", gpuPct),
		fmt.Sprintf("%.2f", anePct),
		fmt.Sprintf("%.3f", powerPkgW),
		fmt.Sprintf("%.3f", powerCPUW),
		fmt.Sprintf("%.3f", powerGPUW),
		fmt.Sprintf("%.3f", powerANEW),
		fmt.Sprintf("%d", memUsedBytes),
		fmt.Sprintf("%d", memPressurePct),
		fmt.Sprintf("%d", swapUsedBytes),
		thermalPressure,
		fmt.Sprintf("%.1f", cpuTemp),
		fmt.Sprintf("%.1f", gpuTemp),
		fmt.Sprintf("%.0f", fan0),
		fmt.Sprintf("%.0f", fan1),
		fmt.Sprintf("%d", batteryPct),
		fmt.Sprintf("%t", batteryCharging),
		fmt.Sprintf("%.0f", networkRxBps),
		fmt.Sprintf("%.0f", networkTxBps),
	}

	e.writer.Write(row)
	e.writer.Flush()
}

func (e *CSVExporter) Close() string {
	if !e.IsEnabled() {
		return ""
	}

	e.mu.Lock()
	defer e.mu.Unlock()

	e.writer.Flush()
	e.file.Close()
	path := e.path
	e.file = nil
	e.writer = nil
	return path
}
