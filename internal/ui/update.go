package ui

import (
	"fmt"
	"os"
	"time"

	"github.com/charmbracelet/bubbletea"
)

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.confirmQuit {
		return m.handleQuitConfirm(msg)
	}

	switch msg := msg.(type) {
	case tickMsg:
		return m.handleTick()

	case tea.KeyMsg:
		return m.handleKey(msg)

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil
	}

	return m, nil
}

func (m *Model) handleTick() (tea.Model, tea.Cmd) {
	if m.quitting {
		return m, tea.Quit
	}

	defer func() {
		if r := recover(); r != nil {
			fmt.Fprintf(os.Stderr, "peaktop: tick panic recovered: %v\n", r)
		}
	}()

	m.collectCPU()
	m.collectGPU()
	m.collectANE()
	m.collectMemory()
	m.collectNetwork()
	m.collectDisk()
	m.collectBattery()

	if m.tickCount%4 == 0 && m.tickCount > 0 {
		m.collectThermal()
		m.collectProcesses()
	}

	if m.tickCount%8 == 0 && m.tickCount > 0 {
		m.collectPower()
	}

	m.collectAlerts()
	m.exportCSV()
	m.tickCount++

	return m, m.tickCmd()
}

func (m *Model) collectCPU() {
	if err := m.cpu.Collect(); err != nil {
		return
	}
	cpuStats := m.cpu.Stats()
	m.cpuHistory.Push(cpuStats.UsagePercent)
}

func (m *Model) collectGPU() {
	if err := m.gpu.Collect(); err != nil {
		return
	}
	gpuStats := m.gpu.Stats()
	m.gpuHistory.Push(gpuStats.UsagePercent)
}

func (m *Model) collectANE() {
	if err := m.ane.Collect(); err != nil {
		return
	}
	aneStats := m.ane.Stats()
	m.aneHistory.Push(aneStats.UsagePercent)
}

func (m *Model) collectMemory() {
	if err := m.memory.Collect(); err != nil {
		return
	}
}

func (m *Model) collectNetwork() {
	if err := m.network.Collect(); err != nil {
		return
	}
}

func (m *Model) collectDisk() {
	if err := m.disk.Collect(); err != nil {
		return
	}
}

func (m *Model) collectPower() {
	if err := m.power.Collect(); err != nil {
		return
	}
}

func (m *Model) collectThermal() {
	if err := m.thermal.Collect(); err != nil {
		return
	}
}

func (m *Model) collectBattery() {
	if err := m.battery.Collect(); err != nil {
		return
	}
}

func (m *Model) collectProcesses() {
	if err := m.procList.Collect(); err != nil {
		return
	}
}

func (m *Model) collectAlerts() {
	m.alerts = m.alertEngine.Check(
		m.cpu.Stats(),
		m.gpu.Stats(),
		m.memory.Stats(),
		m.thermal.Stats(),
		m.battery.Stats(),
	)
}

func (m *Model) exportCSV() {
	gpuStats := m.gpu.Stats()
	aneStats := m.ane.Stats()
	powerStats := m.power.Stats()
	memStats := m.memory.Stats()
	thermalStats := m.thermal.Stats()
	batteryStats := m.battery.Stats()
	netStats := m.network.Stats()

	m.csvExporter.WriteRow(
		m.cpu.Stats().UsagePercent,
		gpuStats.UsagePercent,
		aneStats.UsagePercent,
		powerStats.PackageWatts,
		powerStats.CPUWatts,
		powerStats.GPUWatts,
		powerStats.ANEWatts,
		memStats.UsedBytes,
		memStats.PressurePercent,
		memStats.SwapUsedBytes,
		thermalStats.Pressure,
		thermalStats.CputempC,
		thermalStats.GPUTempC,
		thermalStats.FanRPMs,
		batteryStats.Percent,
		batteryStats.IsCharging,
		netStats.RxBytesPerSec,
		netStats.TxBytesPerSec,
	)
}

func (m *Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "ctrl+c":
		m.confirmQuit = true
		return m, nil

	case "h":
		m.showHelp = !m.showHelp
		return m, nil

	case "j", "down":
		m.selectedProc++
		if m.selectedProc >= len(m.procList.List()) {
			m.selectedProc = len(m.procList.List()) - 1
		}
		return m, nil

	case "k", "up":
		m.selectedProc--
		if m.selectedProc < 0 {
			m.selectedProc = 0
		}
		return m, nil

	case "1":
		m.activeTab = tabOverview
		return m, nil
	case "2":
		m.activeTab = tabProcesses
		return m, nil
	case "3":
		m.activeTab = tabThermal
		return m, nil
	case "4":
		m.activeTab = tabNetwork
		return m, nil
	case "5":
		m.activeTab = tabBattery
		return m, nil

	case "t":
		return m.handleThemeCycle()

	case "+", "=":
		return m.handleTickFaster()
	case "-":
		return m.handleTickSlower()
	}

	return m, nil
}

func (m *Model) handleThemeCycle() (tea.Model, tea.Cmd) {
	themes := AllThemes()
	for i, t := range themes {
		if t.Name == m.theme.Name {
			m.theme = themes[(i+1)%len(themes)]
			m.rebuildStyles()
			return m, nil
		}
	}
	m.theme = themes[0]
	m.rebuildStyles()
	return m, nil
}

func (m *Model) handleTickFaster() (tea.Model, tea.Cmd) {
	if m.tickInterval > 150*time.Millisecond {
		m.tickInterval -= 100 * time.Millisecond
	}
	return m, nil
}

func (m *Model) handleTickSlower() (tea.Model, tea.Cmd) {
	if m.tickInterval < 3000*time.Millisecond {
		m.tickInterval += 100 * time.Millisecond
	}
	return m, nil
}

func (m *Model) handleQuit() (tea.Model, tea.Cmd) {
	m.quitting = true
	path := m.csvExporter.Close()
	if path != "" {
		fmt.Fprintf(os.Stderr, "Session saved to %s (%d samples)\n", path, m.tickCount)
	}
	return m, tea.Quit
}

func (m *Model) handleQuitConfirm(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "y":
			return m.handleQuit()
		case "n":
			m.quitting = true
			_ = m.csvExporter.Close()
			return m, tea.Quit
		case "q", "ctrl+c", "esc":
			m.confirmQuit = false
			return m, nil
		default:
			return m, nil
		}
	}
	return m, nil
}
