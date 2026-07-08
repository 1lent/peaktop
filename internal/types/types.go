package types

type Collector interface {
	Collect() error
	Name() string
}

type CPUStats struct {
	UsagePercent float64
	PerCore      map[string]float64
	ECoreAvg     float64
	PCoreAvg     float64
	SCoreAvg     float64
	FrequencyMHz float64
}

type GPUStats struct {
	UsagePercent float64
	ActiveMHz    float64
	VRAMUsedMB   uint64
	VRAMTotalMB  uint64
}

type MemoryStats struct {
	TotalBytes      uint64
	UsedBytes       uint64
	FreeBytes       uint64
	WiredBytes      uint64
	CompressedBytes uint64
	SwapTotalBytes  uint64
	SwapUsedBytes   uint64
	PressurePercent int
}

type NetworkStats struct {
	RxBytesPerSec float64
	TxBytesPerSec float64
	Interfaces    []NetInterface
}

type NetInterface struct {
	Name  string
	RxBps float64
	TxBps float64
}

type DiskStats struct {
	ReadBytesPerSec  float64
	WriteBytesPerSec float64
	IOPS             float64
}

type PowerStats struct {
	PackageWatts float64
	CPUWatts     float64
	GPUWatts     float64
	ANEWatts     float64
	DRAMWatts    float64
}

type ThermalStats struct {
	Pressure    string
	CputempC    float64
	GPUTempC    float64
	FanRPMs     []float64
	FanModes    []string
}

type BatteryStats struct {
	Percent        int
	IsCharging     bool
	Watts          float64
	TimeRemaining  int
	CycleCount     int
	MaxCapacity    int
	DesignCapacity int
	IsPresent      bool
}

type ProcessInfo struct {
	PID           int32
	Name          string
	CPUPercent    float64
	MemPercent    float64
	GPUPercent    float64
	ANEPercent    float64
	EnergyImpact  float64
}

type AlertEvent struct {
	Timestamp string
	Level     string
	Source    string
	Message   string
}
