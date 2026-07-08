package widgets

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/1lent/peaktop/internal/types"
)

var (
	throttleNominal  = lipgloss.NewStyle().Foreground(lipgloss.Color("42"))
	throttleFair     = lipgloss.NewStyle().Foreground(lipgloss.Color("220"))
	throttleSerious  = lipgloss.NewStyle().Foreground(lipgloss.Color("208"))
	throttleCritical = lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Bold(true)
	throttleLabel    = lipgloss.NewStyle().Foreground(lipgloss.Color("248"))
	throttleFanLabel = lipgloss.NewStyle().Foreground(lipgloss.Color("246"))
	throttleFanBar   = lipgloss.NewStyle().Foreground(lipgloss.Color("75"))
	throttleFanBg    = lipgloss.NewStyle().Foreground(lipgloss.Color("238"))
)

const (
	throttleFanBarWidth = 20
	throttleMaxRPM      = 7000.0
)

func RenderThrottleBar(thermal types.ThermalStats) string {
	var parts []string

	parts = append(parts, renderThermalPressure(thermal.Pressure))
	parts = append(parts, renderTemperatures(thermal.CputempC, thermal.GPUTempC))
	parts = append(parts, renderFans(thermal.FanRPMs))

	return strings.Join(parts, " │ ")
}

func renderThermalPressure(pressure string) string {
	var style lipgloss.Style
	var pct int
	switch pressure {
	case "N/A":
		style = throttleLabel
		pct = -1
	case "Nominal":
		style = throttleNominal
		pct = 0
	case "Fair":
		style = throttleFair
		pct = 33
	case "Serious":
		style = throttleSerious
		pct = 66
	case "Critical":
		style = throttleCritical
		pct = 100
	default:
		style = throttleLabel
	}

	label := throttleLabel.Render("Throttle:")
	var value string
	if pct < 0 {
		value = style.Render(pressure)
	} else {
		value = style.Render(fmt.Sprintf("%s (%d%%)", pressure, pct))
	}
	return fmt.Sprintf("%s %s", label, value)
}

func renderTemperatures(cpuTemp, gpuTemp float64) string {
	parts := []string{}
	if cpuTemp > 0 {
		style := tempStyle(cpuTemp)
		parts = append(parts, throttleLabel.Render(fmt.Sprintf("CPU: %s", style.Render(fmt.Sprintf("%.1f°C", cpuTemp)))))
	}
	if gpuTemp > 0 {
		style := tempStyle(gpuTemp)
		parts = append(parts, throttleLabel.Render(fmt.Sprintf("GPU: %s", style.Render(fmt.Sprintf("%.1f°C", gpuTemp)))))
	}
	if len(parts) == 0 {
		return ""
	}
	return strings.Join(parts, " ")
}

func tempStyle(temp float64) lipgloss.Style {
	switch {
	case temp >= 90:
		return throttleCritical
	case temp >= 75:
		return throttleSerious
	case temp >= 60:
		return throttleFair
	default:
		return throttleNominal
	}
}

func renderFans(rpms []float64) string {
	if len(rpms) == 0 {
		return throttleLabel.Render("Fans: --")
	}

	var parts []string
	for i, rpm := range rpms {
		label := throttleFanLabel.Render(fmt.Sprintf("F%d", i))
		rpmText := throttleFanLabel.Render(fmt.Sprintf("%4.0f RPM", rpm))
		bar := renderFanBar(rpm)
		parts = append(parts, fmt.Sprintf("%s %s %s", label, bar, rpmText))
	}

	return strings.Join(parts, " │ ")
}

func renderFanBar(rpm float64) string {
	ratio := rpm / throttleMaxRPM
	if ratio > 1 {
		ratio = 1
	}
	filled := int(ratio * throttleFanBarWidth)
	if filled > throttleFanBarWidth {
		filled = throttleFanBarWidth
	}

	bar := throttleFanBar.Render(strings.Repeat("█", filled))
	empty := throttleFanBg.Render(strings.Repeat("░", throttleFanBarWidth-filled))
	return bar + empty
}
