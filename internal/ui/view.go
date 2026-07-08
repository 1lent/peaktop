package ui

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/1lent/peaktop/internal/throttle"
	"github.com/1lent/peaktop/internal/types"
	"github.com/1lent/peaktop/internal/ui/widgets"
)

var (
	dividerStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("238"))
)

func (m *Model) View() string {
	if m.width == 0 {
		return "loading..."
	}

	now := time.Now()
	if !m.lastViewTime.IsZero() {
		elapsed := now.Sub(m.lastViewTime).Seconds()
		if elapsed > 0 {
			m.fps = 1.0 / elapsed
		}
	}
	m.lastViewTime = now

	header := m.renderHeader()
	tabBar := m.renderTabBar()
	content := m.renderActiveTab()
	footer := m.renderFooter()

	divider := dividerStyle.Render(strings.Repeat("─", m.width))

	main := lipgloss.JoinVertical(
		lipgloss.Top,
		header,
		divider,
		tabBar,
		content,
		footer,
	)

	if m.confirmQuit {
		return m.renderQuitOverlay(main)
	}

	return main
}

func (m *Model) renderHeader() string {
	now := time.Now().Format("15:04:05")
	alertBadge := widgets.RenderAlertBadge(m.alerts)
	chip := m.detectChipName()

	left := m.styles.header.Render(fmt.Sprintf("peaktop v0.1 · %s · %s · %s", chip, m.theme.Label, now))
	right := fmt.Sprintf("%s", alertBadge)

	leftWidth := lipgloss.Width(left)
	rightWidth := lipgloss.Width(right)
	gap := m.width - leftWidth - rightWidth - 2
	if gap < 1 {
		gap = 1
	}
	padding := strings.Repeat(" ", gap)

	return left + padding + "  " + right
}

func (m *Model) detectChipName() string {
	if m.chipName != "" {
		return m.chipName
	}
	return "Apple Silicon"
}

func (m *Model) renderTabBar() string {
	tabs := []struct {
		index int
		label string
	}{
		{tabOverview, "[1]Overview"},
		{tabProcesses, "[2]Processes"},
		{tabThermal, "[3]Thermal"},
		{tabNetwork, "[4]Network"},
		{tabBattery, "[5]Battery"},
	}

	var parts []string
	for _, t := range tabs {
		if m.activeTab == t.index {
			parts = append(parts, m.styles.tabActive.Render(t.label))
		} else {
			parts = append(parts, m.styles.tabInactive.Render(t.label))
		}
	}

	return strings.Join(parts, " ")
}

func (m *Model) renderActiveTab() string {
	switch m.activeTab {
	case tabOverview:
		return m.renderOverview()
	case tabProcesses:
		return m.renderProcesses()
	case tabThermal:
		return m.renderThermal()
	case tabNetwork:
		return m.renderNetwork()
	case tabBattery:
		return m.renderBattery()
	}
	return ""
}

func (m *Model) renderFooter() string {
	help := "q:quit  h:help  j/k:scroll  1-5:tabs  t:theme  +/-:speed"
	if m.showHelp {
		help = fmt.Sprintf("theme: %s | tick: %dms | TAB: Overview | Process table sorted by CPU%% | Energy: green=low red=critical | CSV saves to ~/.peaktop/logs/",
			m.theme.Label, m.tickInterval.Milliseconds())
	}

	powerWarn := ""
	if warn := m.power.RootWarning(); warn != "" {
		powerWarn = "  " + lipgloss.NewStyle().Foreground(lipgloss.Color("220")).Render(warn)
	}

	left := m.styles.footer.Render(help)
	right := powerWarn
	rightWidth := lipgloss.Width(right)
	gap := m.width - lipgloss.Width(left) - rightWidth - 2
	if gap < 1 {
		gap = 1
	}

	return left + strings.Repeat(" ", gap) + right
}

func (m *Model) renderOverview() string {
	cpu := m.cpu.Stats()
	gpu := m.gpu.Stats()
	ane := m.ane.Stats()
	mem := m.memory.Stats()
	thermal := m.thermal.Stats()
	battery := m.battery.Stats()
	net := m.network.Stats()
	disk := m.disk.Stats()
	power := m.power.Stats()

	colWidth := m.width / 2
	if colWidth < 30 {
		colWidth = 30
	}

	leftCol := lipgloss.JoinVertical(lipgloss.Top,
		m.styles.section.Render(m.renderCPUTitle(cpu)),
		widgets.RenderGauge("Total", cpu.UsagePercent, colWidth),
		widgets.RenderGauge("P-Cluster", cpu.PCoreAvg, colWidth),
		widgets.RenderGauge("E-Cluster", cpu.ECoreAvg, colWidth),
		m.renderCoreBars(cpu),
		m.renderCPUTemps(thermal, cpu),
		m.renderSparklineBlock("CPU History", m.cpuHistory, colWidth),
		"",
		m.styles.section.Render("GPU"),
		widgets.RenderGauge("Usage", gpu.UsagePercent, colWidth),
		m.renderGPUDetail(gpu, thermal),
		m.renderSparklineBlock("GPU History", m.gpuHistory, colWidth),
		"",
		m.styles.section.Render("ANE"),
		m.renderANEBlock(ane.UsagePercent, m.ane.FirstError(), colWidth),
	)

	rightCol := lipgloss.JoinVertical(lipgloss.Top,
		m.styles.section.Render("Memory"),
		m.renderMemoryBar(mem, colWidth),
		"",
		m.styles.section.Render("Storage"),
		m.renderDiskStorage(disk),
		"",
		m.styles.section.Render("Network"),
		m.renderNetworkDetail(net, colWidth),
		"",
		m.styles.section.Render("Thermal"),
		widgets.RenderThrottleBar(thermal),
		"",
		m.renderBatteryBlock(battery),
		"",
		m.renderPowerBlock(power),
		m.renderFPSLabel(),
	)

	cols := lipgloss.JoinHorizontal(lipgloss.Top, leftCol, strings.Repeat(" ", 4), rightCol)
	return cols
}

func (m *Model) renderProcesses() string {
	processes := m.procList.List()
	return widgets.RenderProcessTable(processes, m.width, m.selectedProc, "cpu")
}

func (m *Model) renderThermal() string {
	thermal := m.thermal.Stats()

	parts := []string{
		m.styles.section.Render("Thermal State"),
		widgets.RenderThrottleBar(thermal),
	}

	if thermal.CputempC > 0 || thermal.GPUTempC > 0 {
		parts = append(parts, "",
			m.styles.section.Render("Temperature History"),
			m.renderTempBlock(thermal),
		)
	}

	if len(thermal.FanRPMs) > 0 {
		parts = append(parts, "",
			m.styles.section.Render("Fan Speeds"),
			m.renderFanDetail(thermal),
			m.renderSparklineBlock("Fan History", m.fanHistory, m.width),
		)
	} else {
		msg := "  no fans (SMC not available on Apple Silicon)"
		if strings.Contains(m.detectChipName(), "A1") || strings.Contains(m.detectChipName(), "A18") {
			msg = "  no fans (passive cooling on Neo devices)"
		}
		msg += "\n  run with sudo for fan + temperature data"
		parts = append(parts, "",
			dimText(msg),
		)
	}

	return lipgloss.JoinVertical(lipgloss.Top, parts...)
}

func (m *Model) renderNetwork() string {
	net := m.network.Stats()

	parts := []string{
		m.styles.section.Render(fmt.Sprintf("Throughput  RX: %.1f MB/s  TX: %.1f MB/s",
			net.RxBytesPerSec/1e6, net.TxBytesPerSec/1e6)),
		"",
	}

	if len(net.Interfaces) > 0 {
		for _, iface := range net.Interfaces {
			rx := m.styles.netRx.Render(fmt.Sprintf("RX %7.1f KB/s", iface.RxBps/1e3))
			tx := m.styles.netTx.Render(fmt.Sprintf("TX %7.1f KB/s", iface.TxBps/1e3))
			parts = append(parts, fmt.Sprintf("  %-8s %s  %s", iface.Name, rx, tx))
		}
	} else {
		parts = append(parts, dimText("no active network interfaces"))
	}

	return lipgloss.JoinVertical(lipgloss.Top, parts...)
}

func (m *Model) renderBattery() string {
	battery := m.battery.Stats()

	if !battery.IsPresent {
		return dimText("no battery (desktop system)")
	}

	chargeWidth := 30
	chargeFilled := int(float64(chargeWidth) * float64(battery.Percent) / 100.0)
	if chargeFilled > chargeWidth {
		chargeFilled = chargeWidth
	}

	var chargeStyle lipgloss.Style
	switch {
	case battery.Percent <= 10:
		chargeStyle = m.styles.battLow
	case battery.Percent <= 30:
		chargeStyle = m.styles.battMid
	default:
		chargeStyle = m.styles.battFull
	}

	chargeBar := chargeStyle.Render(strings.Repeat("█", chargeFilled)) +
		m.styles.memEmpty.Render(strings.Repeat("░", chargeWidth-chargeFilled))

	health := battery.MaxCapacity
	if battery.MaxCapacity > 100 && battery.DesignCapacity > 0 {
		health = battery.MaxCapacity * 100 / battery.DesignCapacity
	}

	status := "discharging"
	if battery.IsCharging {
		status = "charging"
	}

	parts := []string{
		m.styles.section.Render("Battery"),
		"",
		fmt.Sprintf("  Charge:   %s %d%%", chargeBar, battery.Percent),
		fmt.Sprintf("  Status:   %s", status),
	}

	if battery.TimeRemaining > 0 {
		hrs := battery.TimeRemaining / 60
		mins := battery.TimeRemaining % 60
		parts = append(parts, fmt.Sprintf("  Remaining: %dh %dm", hrs, mins))
	}

	if battery.MaxCapacity > 100 {
		parts = append(parts,
			fmt.Sprintf("  Health:   %d%% (%d/%d mAh)", health, battery.MaxCapacity, battery.DesignCapacity),
			fmt.Sprintf("  Cycles:   %d", battery.CycleCount),
		)
	} else if battery.MaxCapacity < 100 && battery.DesignCapacity > 0 {
		effective := battery.MaxCapacity * battery.DesignCapacity / 100
		parts = append(parts,
			fmt.Sprintf("  Health:   %d%% (%d/%d mAh)", health, effective, battery.DesignCapacity),
			fmt.Sprintf("  Cycles:   %d", battery.CycleCount),
		)
	} else {
		parts = append(parts,
			fmt.Sprintf("  Health:   %d%% (%d mAh design)", health, battery.DesignCapacity),
			fmt.Sprintf("  Cycles:   %d", battery.CycleCount),
		)
	}

	return lipgloss.JoinVertical(lipgloss.Top, parts...)
}

func (m *Model) renderMemoryBar(mem types.MemoryStats, width int) string {
	if mem.TotalBytes == 0 {
		return dimText("memory stats unavailable")
	}

	barWidth := width - 4
	if barWidth < 10 {
		barWidth = 10
	}

	total := float64(mem.TotalBytes)
	wiredPct := int(float64(mem.WiredBytes) / total * float64(barWidth))
	activePct := int(float64(mem.UsedBytes-mem.WiredBytes) / total * float64(barWidth))
	compPct := int(float64(mem.CompressedBytes) / total * float64(barWidth))

	remaining := barWidth - wiredPct - activePct - compPct
	if remaining < 0 {
		remaining = 0
	}

	bar := m.styles.memFull.Render(strings.Repeat("█", wiredPct)) +
		m.styles.memCached.Render(strings.Repeat("█", activePct)) +
		m.styles.memSwap.Render(strings.Repeat("█", compPct)) +
		m.styles.memEmpty.Render(strings.Repeat("░", remaining))

	usedGB := float64(mem.UsedBytes) / 1e9
	totalGB := float64(mem.TotalBytes) / 1e9

	info := fmt.Sprintf("\n  %s/%s GB  wired: %.1f GB  compressed: %.1f GB  swap: %.1f GB  pressure: %d%%",
		fmt.Sprintf("%.1f", usedGB), fmt.Sprintf("%.1f", totalGB),
		float64(mem.WiredBytes)/1e9, float64(mem.CompressedBytes)/1e9,
		float64(mem.SwapUsedBytes)/1e9, mem.PressurePercent)

	return bar + m.styles.footer.Render(info)
}

func (m *Model) renderBatteryBlock(battery types.BatteryStats) string {
	if !battery.IsPresent {
		return ""
	}

	status := "discharging"
	style := m.styles.battMid
	if battery.IsCharging {
		status = "charging"
		style = m.styles.battFull
	}

	return lipgloss.JoinVertical(lipgloss.Top,
		m.styles.section.Render("Battery"),
		fmt.Sprintf("  %s %d%% %s", style.Render("█"), battery.Percent, status),
	)
}

func (m *Model) renderNetworkBlock(net types.NetworkStats, width int) string {
	return m.renderNetworkDetail(net, width)
}

func (m *Model) renderNetworkDetail(net types.NetworkStats, width int) string {
	parts := []string{
		fmt.Sprintf("  RX: %.2f MB/s  TX: %.2f MB/s",
			net.RxBytesPerSec/1e6, net.TxBytesPerSec/1e6),
	}

	if len(net.Interfaces) > 0 {
		for _, iface := range net.Interfaces {
			rx := m.styles.netRx.Render(fmt.Sprintf("RX %7.1f KB/s", iface.RxBps/1e3))
			tx := m.styles.netTx.Render(fmt.Sprintf("TX %7.1f KB/s", iface.TxBps/1e3))
			parts = append(parts, fmt.Sprintf("  %-8s %s  %s", iface.Name, rx, tx))
		}
	}

	return lipgloss.JoinVertical(lipgloss.Top, parts...)
}

func (m *Model) renderCoreBars(cpu types.CPUStats) string {
	if len(cpu.PerCore) == 0 {
		return ""
	}

	const barW = 10
	full := lipgloss.NewStyle().Foreground(lipgloss.Color("75"))
	mid := lipgloss.NewStyle().Foreground(lipgloss.Color("220"))
	low := lipgloss.NewStyle().Foreground(lipgloss.Color("42"))
	empty := lipgloss.NewStyle().Foreground(lipgloss.Color("238"))
	nameStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("248"))

	var lines []string
	for _, name := range sortedCoreNames(cpu.PerCore) {
		pct := cpu.PerCore[name]
		filled := int(pct / 100.0 * float64(barW))
		if filled > barW {
			filled = barW
		}

		var bar string
		for i := 0; i < barW; i++ {
			if i < filled {
				switch {
				case pct >= 80:
					bar += full.Render("█")
				case pct >= 50:
					bar += mid.Render("▓")
				default:
					bar += low.Render("▒")
				}
			} else {
				bar += empty.Render("░")
			}
		}
		lines = append(lines, fmt.Sprintf("  %s%s  %s %5.1f%%", nameStyle.Render(name), bar, bar, pct))
	}
	return strings.Join(lines, "\n")
}

func sortedCoreNames(perCore map[string]float64) []string {
	var pNames, eNames, other []string
	for name := range perCore {
		if len(name) > 0 && name[0] == 'P' {
			pNames = append(pNames, name)
		} else if len(name) > 0 && name[0] == 'E' {
			eNames = append(eNames, name)
		} else {
			other = append(other, name)
		}
	}
	sort.Strings(pNames)
	sort.Strings(eNames)
	sort.Strings(other)
	return append(append(pNames, eNames...), other...)
}

func (m *Model) renderGPUTemps(thermal types.ThermalStats) string {
	if thermal.GPUTempC <= 0 {
		return ""
	}
	return fmt.Sprintf("  GPU Temp: %.1f°C", thermal.GPUTempC)
}

func (m *Model) renderCPUTitle(cpu types.CPUStats) string {
	if cpu.CoreCount == 0 {
		return "CPU"
	}
	return fmt.Sprintf("CPU (%dP + %dE = %d cores)", cpu.PCoreCount, cpu.ECoreCount, cpu.CoreCount)
}

func (m *Model) renderCPUTemps(thermal types.ThermalStats, cpu types.CPUStats) string {
	var parts []string
	if thermal.CputempC > 0 {
		parts = append(parts, fmt.Sprintf("CPU Temp: %.1f°C", thermal.CputempC))
	}
	if cpu.FrequencyMHz > 0 {
		parts = append(parts, fmt.Sprintf("Freq: %.0f MHz", cpu.FrequencyMHz))
	}
	if len(parts) == 0 {
		return ""
	}
	return "  " + strings.Join(parts, "  ")
}

func (m *Model) renderGPUDetail(gpu types.GPUStats, thermal types.ThermalStats) string {
	var parts []string
	if gpu.ActiveMHz > 0 {
		parts = append(parts, fmt.Sprintf("Freq: %.0f MHz", gpu.ActiveMHz))
	}
	if thermal.GPUTempC > 0 {
		parts = append(parts, fmt.Sprintf("Temp: %.1f°C", thermal.GPUTempC))
	}
	if gpu.VRAMTotalMB > 0 {
		parts = append(parts, fmt.Sprintf("VRAM: %d / %d MB", gpu.VRAMUsedMB, gpu.VRAMTotalMB))
	}
	if len(parts) == 0 {
		return ""
	}
	return "  " + strings.Join(parts, "  ")
}

func (m *Model) renderDiskStorage(disk types.DiskStats) string {
	if disk.TotalBytes == 0 {
		return dimText("  disk info unavailable")
	}
	used := disk.TotalBytes - disk.FreeBytes
	usedStr := formatBytes(used)
	totalStr := formatBytes(disk.TotalBytes)
	return fmt.Sprintf("  %s / %s used", usedStr, totalStr)
}

func (m *Model) renderPowerBlock(power types.PowerStats) string {
	if power.PackageWatts <= 0 && power.CPUWatts <= 0 {
		return ""
	}
	var parts []string
	if power.PackageWatts > 0 {
		parts = append(parts, fmt.Sprintf("Pkg: %.2f W", power.PackageWatts))
	}
	if power.CPUWatts > 0 {
		parts = append(parts, fmt.Sprintf("CPU: %.2f W", power.CPUWatts))
	}
	if power.GPUWatts > 0 {
		parts = append(parts, fmt.Sprintf("GPU: %.2f W", power.GPUWatts))
	}
	if len(parts) == 0 {
		return ""
	}
	return m.styles.section.Render("Power") + "\n  " + strings.Join(parts, "  ")
}

func formatBytes(b uint64) string {
	switch {
	case b >= 1<<40:
		return fmt.Sprintf("%.1f TB", float64(b)/(1<<40))
	case b >= 1<<30:
		return fmt.Sprintf("%.1f GB", float64(b)/(1<<30))
	case b >= 1<<20:
		return fmt.Sprintf("%.1f MB", float64(b)/(1<<20))
	default:
		return fmt.Sprintf("%d KB", b/(1<<10))
	}
}

func (m *Model) renderFPSLabel() string {
	return dimText(fmt.Sprintf("  FPS: %.0f", m.fps))
}

func (m *Model) renderSparklineBlock(label string, rb *throttle.RingBuffer, width int) string {
	if rb.Len() == 0 {
		return dimText("  (no data)")
	}

	sparkWidth := width - 4
	if sparkWidth < 10 {
		sparkWidth = 10
	}
	if sparkWidth > rb.Len() {
		sparkWidth = rb.Len()
	}

	spark := widgets.RenderSparkline(rb.Values(), sparkWidth, 1)
	if label != "" {
		return fmt.Sprintf("  %s %s", m.styles.header.Render(label), spark)
	}
	return spark
}

func (m *Model) renderTempBlock(thermal types.ThermalStats) string {
	var parts []string
	if thermal.CputempC > 0 {
		parts = append(parts, fmt.Sprintf("  CPU: %.1f°C", thermal.CputempC))
	}
	if thermal.GPUTempC > 0 {
		parts = append(parts, fmt.Sprintf("  GPU: %.1f°C", thermal.GPUTempC))
	}
	return strings.Join(parts, "  ")
}

func (m *Model) renderFanDetail(thermal types.ThermalStats) string {
	const maxRPM = 7000.0
	const barWidth = 20
	full := lipgloss.NewStyle().Foreground(lipgloss.Color("75"))
	empty := lipgloss.NewStyle().Foreground(lipgloss.Color("238"))

	var parts []string
	for i, rpm := range thermal.FanRPMs {
		ratio := rpm / maxRPM
		if ratio > 1 {
			ratio = 1
		}
		filled := int(ratio * barWidth)
		bar := full.Render(strings.Repeat("█", filled))
		bg := empty.Render(strings.Repeat("░", barWidth-filled))
		parts = append(parts, fmt.Sprintf("  Fan %d: %s %6.0f RPM", i, bar+bg, rpm))
	}
	return strings.Join(parts, "\n")
}

func (m *Model) renderANEBlock(usage float64, firstErr error, colWidth int) string {
	if firstErr != nil {
		return dimText("  ANE data not available on this device")
	}
	return widgets.RenderGauge("Usage", usage, colWidth)
}

func dimText(s string) string {
	return lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render(s)
}

func (m *Model) renderQuitOverlay(main string) string {
	dialogStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(m.theme.CPU)).
		Padding(1, 2).
		Align(lipgloss.Center)

	dialog := dialogStyle.Render(
		lipgloss.JoinVertical(lipgloss.Center,
			m.styles.header.Render("Quit peaktop?"),
			"",
			fmt.Sprintf("  Logged %d samples to ~/.peaktop/logs/", m.tickCount),
			"",
			"  [y] save log & quit",
			"  [n] quit without saving",
			"  [esc] cancel",
		),
	)

	return lipgloss.Place(
		m.width, m.height,
		lipgloss.Center, lipgloss.Center,
		dialog,
		lipgloss.WithWhitespaceChars(" "),
	)
}
