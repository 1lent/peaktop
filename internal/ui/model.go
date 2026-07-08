package ui

import (
	"strings"
	"time"

	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/brodie/peaktop/internal/alert"
	"github.com/brodie/peaktop/internal/apple"
	"github.com/brodie/peaktop/internal/collector"
	"github.com/brodie/peaktop/internal/log"
	"github.com/brodie/peaktop/internal/process"
	"github.com/brodie/peaktop/internal/throttle"
	"github.com/brodie/peaktop/internal/types"
)

const (
	tabOverview  = 0
	tabProcesses = 1
	tabThermal   = 2
	tabNetwork   = 3
	tabBattery   = 4
	tabCount     = 5

	ringBufferSize     = 60
	defaultTickInterval = 500 * time.Millisecond
)

type tickMsg time.Time

type Model struct {
	cpu      *collector.CPUCollector
	gpu      *collector.GPUCollector
	ane      *collector.ANECollector
	memory   *collector.MemoryCollector
	network  *collector.NetworkCollector
	disk     *collector.DiskCollector
	power    *collector.PowerCollector
	thermal  *collector.ThermalCollector
	battery  *collector.BatteryCollector
	procList *process.ProcessList

	cpuHistory *throttle.RingBuffer
	gpuHistory *throttle.RingBuffer
	aneHistory *throttle.RingBuffer

	alertEngine *alert.AlertEngine
	alerts      []types.AlertEvent

	csvExporter *log.CSVExporter

	theme Theme

	activeTab    int
	selectedProc int
	showHelp     bool
	quitting     bool
	width        int
	height       int
	tickCount    int
	tickInterval time.Duration
	chipName     string
	lastViewTime time.Time
	fps          float64
	styles       cachedStyles
}

type cachedStyles struct {
	header    lipgloss.Style
	tabActive lipgloss.Style
	tabInactive lipgloss.Style
	footer    lipgloss.Style
	section   lipgloss.Style
	memFull   lipgloss.Style
	memCached lipgloss.Style
	memSwap   lipgloss.Style
	memEmpty  lipgloss.Style
	battFull  lipgloss.Style
	battMid   lipgloss.Style
	battLow   lipgloss.Style
	netRx     lipgloss.Style
	netTx     lipgloss.Style
}

func (m *Model) rebuildStyles() {
	t := m.theme
	m.styles = cachedStyles{
		header:    lipgloss.NewStyle().Foreground(lipgloss.Color(t.CPU)).Bold(true),
		tabActive: lipgloss.NewStyle().Foreground(lipgloss.Color("0")).Background(lipgloss.Color(t.CPU)).Padding(0, 1),
		tabInactive: lipgloss.NewStyle().Foreground(lipgloss.Color("248")).Padding(0, 1),
		footer:    lipgloss.NewStyle().Foreground(lipgloss.Color(t.Dim)),
		section:   lipgloss.NewStyle().Foreground(lipgloss.Color(t.Header)).Bold(true),
		memFull:   lipgloss.NewStyle().Foreground(lipgloss.Color(t.Memory)),
		memCached: lipgloss.NewStyle().Foreground(lipgloss.Color(t.Network)),
		memSwap:   lipgloss.NewStyle().Foreground(lipgloss.Color(t.Thermal)),
		memEmpty:  lipgloss.NewStyle().Foreground(lipgloss.Color("238")),
		battFull:  lipgloss.NewStyle().Foreground(lipgloss.Color(t.Battery)),
		battMid:   lipgloss.NewStyle().Foreground(lipgloss.Color(t.Yellow)),
		battLow:   lipgloss.NewStyle().Foreground(lipgloss.Color(t.Red)),
		netRx:     lipgloss.NewStyle().Foreground(lipgloss.Color(t.Green)),
		netTx:     lipgloss.NewStyle().Foreground(lipgloss.Color(t.CPU)),
	}
}

func NewModel(interval time.Duration) Model {
	if interval <= 0 {
		interval = defaultTickInterval
	}

	m := Model{
		cpu:      collector.NewCPUCollector(),
		gpu:      collector.NewGPUCollector(),
		ane:      collector.NewANECollector(),
		memory:   collector.NewMemoryCollector(),
		network:  collector.NewNetworkCollector(),
		disk:     collector.NewDiskCollector(),
		power:    collector.NewPowerCollector(),
		thermal:  collector.NewThermalCollector(),
		battery:  collector.NewBatteryCollector(),
		procList: process.NewProcessList(),

		cpuHistory: throttle.NewRingBuffer(ringBufferSize),
		gpuHistory: throttle.NewRingBuffer(ringBufferSize),
		aneHistory: throttle.NewRingBuffer(ringBufferSize),

		alertEngine: alert.NewAlertEngine(),

		theme:        LoadTheme(),
		tickInterval: interval,
		chipName:     detectChipName(),
	}
	m.rebuildStyles()
	return m
}

func (m *Model) tickCmd() tea.Cmd {
	return tea.Tick(m.tickInterval, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func (m *Model) Init() tea.Cmd {
	exporter, err := log.NewCSVExporter()
	if err != nil {
		exporter = log.NewNoopExporter()
	}
	m.csvExporter = exporter

	return m.tickCmd()
}

func detectChipName() string {
	name, err := apple.ReadSysctlString("machdep.cpu.brand_string")
	if err != nil || name == "" {
		return "Apple Silicon"
	}
	name = strings.TrimSpace(name)
	name = strings.TrimPrefix(name, "Apple ")
	return name
}
