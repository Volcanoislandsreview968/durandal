package metrics

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"sort"
	"strings"
	"time"

	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/disk"
	"github.com/shirou/gopsutil/v4/host"
	"github.com/shirou/gopsutil/v4/mem"
	"github.com/shirou/gopsutil/v4/net"
	"github.com/shirou/gopsutil/v4/process"
	"github.com/shirou/gopsutil/v4/sensors"
)

var (
	prevBytesSent uint64
	prevBytesRecv uint64
	prevTime      time.Time
)

// CollectSnapshot gathers a full system metrics snapshot.
func CollectSnapshot() (Snapshot, error) {
	var snap Snapshot

	// Host info
	snap.Host = collectHost()

	// CPU
	snap.CPU = collectCPU()

	// Memory
	snap.Memory = collectMemory()

	// Processes (top 50 by CPU)
	snap.Processes = collectProcesses(50)

	// Network
	snap.Network = collectNetwork()

	// Disks
	snap.Disks = collectDisks()

	// Sensors (temperatures + battery)
	snap.Sensors = collectSensors()

	// GPUs
	snap.GPUs = collectGPUs()

	return snap, nil
}

func collectHost() HostInfo {
	info := HostInfo{}

	h, err := host.Info()
	if err == nil {
		info.Hostname = h.Hostname
		info.OS = h.Platform + " " + h.PlatformVersion
		info.Kernel = h.KernelVersion
		info.Architecture = h.KernelArch

		upSecs := h.Uptime
		days := upSecs / 86400
		hours := (upSecs % 86400) / 3600
		mins := (upSecs % 3600) / 60
		if days > 0 {
			info.Uptime = fmt.Sprintf("%dd %dh %dm", days, hours, mins)
		} else if hours > 0 {
			info.Uptime = fmt.Sprintf("%dh %dm", hours, mins)
		} else {
			info.Uptime = fmt.Sprintf("%dm", mins)
		}
	} else {
		info.Hostname, _ = os.Hostname()
	}

	return info
}

func collectCPU() CPUInfo {
	info := CPUInfo{}

	// Overall percent
	percents, err := cpu.Percent(0, false)
	if err == nil && len(percents) > 0 {
		info.TotalPercent = percents[0]
	}

	// Per-core percent
	perCore, err := cpu.Percent(0, true)
	if err == nil {
		info.PerCore = perCore
	}

	// CPU info
	cpuInfo, err := cpu.Info()
	if err == nil && len(cpuInfo) > 0 {
		info.ModelName = cpuInfo[0].ModelName
		info.Cores = int(cpuInfo[0].Cores)
	}
	info.Threads = len(info.PerCore)

	return info
}

func collectMemory() MemInfo {
	info := MemInfo{}

	v, err := mem.VirtualMemory()
	if err == nil {
		info.TotalRAM = v.Total
		info.UsedRAM = v.Used
		info.PercentRAM = v.UsedPercent
		info.Cached = v.Cached
		info.Buffers = v.Buffers
	}

	s, err := mem.SwapMemory()
	if err == nil {
		info.TotalSwap = s.Total
		info.UsedSwap = s.Used
		info.PercentSwap = s.UsedPercent
	}

	return info
}

func collectProcesses(limit int) []ProcessInfo {
	procs, err := process.Processes()
	if err != nil {
		return nil
	}

	var result []ProcessInfo

	for _, p := range procs {
		name, _ := p.Name()
		cpuPct, _ := p.CPUPercent()
		memPct, _ := p.MemoryPercent()
		memInfo, _ := p.MemoryInfo()
		status, _ := p.Status()
		user, _ := p.Username()
		cmdline, _ := p.Cmdline()

		if cmdline == "" {
			cmdline = name
		}

		// Truncate command
		if len(cmdline) > 40 {
			cmdline = cmdline[:37] + "..."
		}

		var rss uint64
		if memInfo != nil {
			rss = memInfo.RSS
		}

		statusStr := ""
		if len(status) > 0 {
			statusStr = strings.Join(status, ",")
		}

		result = append(result, ProcessInfo{
			PID:     p.Pid,
			Name:    name,
			CPU:     cpuPct,
			Memory:  memPct,
			MemRSS:  rss,
			Status:  statusStr,
			User:    user,
			Command: cmdline,
		})
	}

	// Sort by CPU usage descending
	sort.Slice(result, func(i, j int) bool {
		return result[i].CPU > result[j].CPU
	})

	if len(result) > limit {
		result = result[:limit]
	}

	return result
}

func collectNetwork() NetworkInfo {
	info := NetworkInfo{}

	counters, err := net.IOCounters(false)
	if err == nil && len(counters) > 0 {
		info.BytesSent = counters[0].BytesSent
		info.BytesRecv = counters[0].BytesRecv

		now := time.Now()
		if !prevTime.IsZero() {
			elapsed := now.Sub(prevTime).Seconds()
			if elapsed > 0 {
				info.BytesSentRate = uint64(float64(info.BytesSent-prevBytesSent) / elapsed)
				info.BytesRecvRate = uint64(float64(info.BytesRecv-prevBytesRecv) / elapsed)
			}
		}

		prevBytesSent = info.BytesSent
		prevBytesRecv = info.BytesRecv
		prevTime = now
	}

	return info
}

func collectDisks() []DiskInfo {
	partitions, err := disk.Partitions(true)
	if err != nil {
		return nil
	}

	var result []DiskInfo
	seen := make(map[string]bool)

	for _, p := range partitions {
		// Filter out noisy virtual/loop filesystems but keep btrfs/lvm root
		if p.Fstype == "squashfs" || p.Fstype == "tmpfs" || p.Fstype == "devtmpfs" ||
			strings.HasPrefix(p.Mountpoint, "/snap") || strings.HasPrefix(p.Mountpoint, "/var/lib/docker") ||
			strings.HasPrefix(p.Mountpoint, "/run") || strings.HasPrefix(p.Mountpoint, "/sys") || strings.HasPrefix(p.Mountpoint, "/dev") {
			continue
		}

		if seen[p.Device] {
			continue
		}
		seen[p.Device] = true

		usage, err := disk.Usage(p.Mountpoint)
		if err != nil || usage.Total == 0 {
			continue
		}

		result = append(result, DiskInfo{
			Mountpoint:  p.Mountpoint,
			Device:      p.Device,
			Filesystem:  p.Fstype,
			Total:       usage.Total,
			Used:        usage.Used,
			UsedPercent: usage.UsedPercent,
		})
	}

	return result
}

func collectSensors() SensorInfo {
	info := SensorInfo{}

	// Temperature sensors
	temps, err := sensors.TemperaturesWithContext(context.Background())
	if err == nil {
		for _, t := range temps {
			if t.Temperature > 0 {
				info.Temperatures = append(info.Temperatures, TemperatureReading{
					Label:       t.SensorKey,
					Temperature: t.Temperature,
				})
			}
		}
	}

	// Battery — try reading from /sys
	battPath := "/sys/class/power_supply/BAT0/capacity"
	if data, err := os.ReadFile(battPath); err == nil {
		var pct float64
		fmt.Sscanf(strings.TrimSpace(string(data)), "%f", &pct)

		statusData, _ := os.ReadFile("/sys/class/power_supply/BAT0/status")
		status := strings.TrimSpace(string(statusData))

		info.Battery = &BatteryInfo{
			Percent:    pct,
			IsCharging: status == "Charging",
		}
	}

	return info
}

func collectGPUs() []GPUInfo {
	out, err := exec.Command("nvidia-smi", "--query-gpu=name,utilization.gpu,memory.used,memory.total,temperature.gpu", "--format=csv,noheader,nounits").Output()
	if err != nil {
		return nil
	}

	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	var gpus []GPUInfo
	for _, line := range lines {
		if line == "" {
			continue
		}
		parts := strings.Split(line, ",")
		if len(parts) != 5 {
			continue
		}

		var util float64
		var memU, memT uint64
		var temp float64

		fmt.Sscanf(strings.TrimSpace(parts[1]), "%f", &util)
		fmt.Sscanf(strings.TrimSpace(parts[2]), "%d", &memU)
		fmt.Sscanf(strings.TrimSpace(parts[3]), "%d", &memT)
		fmt.Sscanf(strings.TrimSpace(parts[4]), "%f", &temp)

		gpus = append(gpus, GPUInfo{
			Name:        strings.TrimSpace(parts[0]),
			Utilization: util,
			MemoryUsed:  memU,
			MemoryTotal: memT,
			Temperature: temp,
		})
	}
	return gpus
}
