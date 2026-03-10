package metrics

// Snapshot holds all system metrics for one collection tick.
type Snapshot struct {
	CPU       CPUInfo
	Memory    MemInfo
	Processes []ProcessInfo
	Network   NetworkInfo
	Disks     []DiskInfo
	Sensors   SensorInfo
	Host      HostInfo
}

// HostInfo contains system identity data.
type HostInfo struct {
	Hostname     string
	Kernel       string
	Uptime       string
	OS           string
	Architecture string
}

// CPUInfo contains CPU usage metrics.
type CPUInfo struct {
	TotalPercent float64
	PerCore      []float64
	ModelName    string
	Cores        int
	Threads      int
}

// MemInfo contains memory and swap usage.
type MemInfo struct {
	TotalRAM    uint64
	UsedRAM     uint64
	PercentRAM  float64
	TotalSwap   uint64
	UsedSwap    uint64
	PercentSwap float64
	Cached      uint64
	Buffers     uint64
}

// ProcessInfo represents one process.
type ProcessInfo struct {
	PID     int32
	Name    string
	CPU     float64
	Memory  float32
	MemRSS  uint64
	Status  string
	User    string
	Command string
}

// NetworkInfo contains network interface data.
type NetworkInfo struct {
	BytesSent     uint64
	BytesRecv     uint64
	BytesSentRate uint64
	BytesRecvRate uint64
}

// DiskInfo represents one mounted filesystem.
type DiskInfo struct {
	Mountpoint  string
	Device      string
	Filesystem  string
	Total       uint64
	Used        uint64
	UsedPercent float64
}

// SensorInfo contains temperature and battery data.
type SensorInfo struct {
	Temperatures []TemperatureReading
	Battery      *BatteryInfo
}

// TemperatureReading is a single sensor reading.
type TemperatureReading struct {
	Label       string
	Temperature float64 // Celsius
}

// BatteryInfo holds battery state if available.
type BatteryInfo struct {
	Percent    float64
	IsCharging bool
	TimeLeft   string
}
