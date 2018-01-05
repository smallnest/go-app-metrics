package monitor

import (
	"time"

	metrics "github.com/rcrowley/go-metrics"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/disk"
	"github.com/shirou/gopsutil/load"
	"github.com/shirou/gopsutil/mem"
	"github.com/shirou/gopsutil/net"
)

var (
	cpuStat       *cpu.TimesStat
	partitions    []string
	netStats      = make(map[string]*net.IOCountersStat)
	systemMetrics struct {
		CPUStat struct {
			User   metrics.Gauge
			System metrics.Gauge
			Idle   metrics.Gauge
			Iowait metrics.Gauge
		}
		LoadStat struct {
			Load1  metrics.Gauge
			Load5  metrics.Gauge
			Load15 metrics.Gauge
		}
		MemStat struct {
			Total     metrics.Gauge
			Available metrics.Gauge
			Used      metrics.Gauge
		}
		SwapMemStat struct {
			Total metrics.Gauge
			Free  metrics.Gauge
			Used  metrics.Gauge
		}
		DiskStat map[string]*struct {
			Total metrics.Gauge
			Free  metrics.Gauge
		}
		BandwidthStat map[string]*struct {
			BytesSent   metrics.Gauge
			BytesRecv   metrics.Gauge
			PacketsSent metrics.Gauge
			PacketsRecv metrics.Gauge
		}
		captureSystemTimer metrics.Timer
	}
)

// CaptureSystemStats captures new values for the system statistics.
// This is designed to be called as a goroutine.
func CaptureSystemStats(r metrics.Registry, d time.Duration) {
	for _ = range time.Tick(d) {
		CaptureSystemStatsOnce(r)
	}
}

// CaptureSystemStatsOnce captures new values for the system statistics.
// This is designed to be called in a background goroutine.
func CaptureSystemStatsOnce(r metrics.Registry) {
	t := time.Now()

	//cpu * 100
	cpustats, err := cpu.Times(false)
	if err == nil && len(cpustats) > 0 {
		cpustat2 := cpustats[0]
		if cpuStat == nil {
			cpuStat = &cpustat2
		}
		total1, _ := getAllCPUTime(*cpuStat)
		total2, _ := getAllCPUTime(cpustat2)
		total := total2 - total1
		if total != 0 {
			systemMetrics.CPUStat.User.Update(int64((cpustat2.User - cpuStat.User) * 100 / total))
			systemMetrics.CPUStat.System.Update(int64((cpustat2.System - cpuStat.System) * 100 / total))
			systemMetrics.CPUStat.Iowait.Update(int64((cpustat2.Iowait - cpuStat.Iowait) * 100 / total))
			systemMetrics.CPUStat.Idle.Update(int64((cpustat2.Idle - cpuStat.Idle) * 100 / total))
		}
		cpuStat = &cpustat2
	}

	//load * 100
	avg, err := load.Avg()
	if err == nil {
		systemMetrics.LoadStat.Load1.Update(int64(avg.Load1 * 100))
		systemMetrics.LoadStat.Load5.Update(int64(avg.Load5 * 100))
		systemMetrics.LoadStat.Load15.Update(int64(avg.Load15 * 100))
	}

	//mem
	vmem, err := mem.VirtualMemory()
	if err == nil {
		systemMetrics.MemStat.Total.Update(int64(vmem.Total))
		systemMetrics.MemStat.Available.Update(int64(vmem.Available))
		systemMetrics.MemStat.Used.Update(int64(vmem.Used))
	}
	swapmem, err := mem.SwapMemory()
	if err == nil {
		systemMetrics.SwapMemStat.Total.Update(int64(swapmem.Total))
		systemMetrics.SwapMemStat.Free.Update(int64(swapmem.Free))
		systemMetrics.SwapMemStat.Used.Update(int64(swapmem.Used))
	}

	//disk
	for _, p := range partitions {
		s, err := disk.Usage(p)
		if err != nil {
			continue
		}

		systemMetrics.DiskStat[p].Total.Update(int64(s.Total))
		systemMetrics.DiskStat[p].Free.Update(int64(s.Free))
	}

	//bandwidth
	netstats, err := net.IOCounters(true)
	if err == nil {
		for _, s := range netstats {
			if netStats[s.Name] == nil {
				netStats[s.Name] = &s
			}
			s2 := netStats[s.Name]
			if systemMetrics.BandwidthStat[s.Name] == nil {
				registerBandwidthMetrics(r, s.Name)
			}

			systemMetrics.BandwidthStat[s.Name].BytesSent.Update(int64(s.BytesSent - s2.BytesSent))
			systemMetrics.BandwidthStat[s.Name].BytesRecv.Update(int64(s.BytesRecv - s2.BytesRecv))
			systemMetrics.BandwidthStat[s.Name].PacketsSent.Update(int64(s.PacketsSent - s2.PacketsSent))
			systemMetrics.BandwidthStat[s.Name].PacketsRecv.Update(int64(s.PacketsRecv - s2.PacketsRecv))
			news := s
			netStats[s.Name] = &news
		}
	}

	systemMetrics.captureSystemTimer.UpdateSince(t)
}

func registerBandwidthMetrics(r metrics.Registry, name string) {
	bsGauge := metrics.NewGauge()
	bcGauge := metrics.NewGauge()
	psGauge := metrics.NewGauge()
	pcGauge := metrics.NewGauge()

	systemMetrics.BandwidthStat[name] = &struct {
		BytesSent   metrics.Gauge
		BytesRecv   metrics.Gauge
		PacketsSent metrics.Gauge
		PacketsRecv metrics.Gauge
	}{
		BytesSent:   bsGauge,
		BytesRecv:   bcGauge,
		PacketsSent: psGauge,
		PacketsRecv: pcGauge,
	}

	r.Register("bandwidth."+name+".BytesSent", bsGauge)
	r.Register("bandwidth."+name+".BytesRecv", bsGauge)
	r.Register("bandwidth."+name+".PacketsSent", bsGauge)
	r.Register("bandwidth."+name+".PacketsRecv", bsGauge)
}

// RegisterSystemStats registers systemMetrics for the system statistics.
//  The systemMetrics are named by their categories and names, i.e. cpu.Usage.
func RegisterSystemStats(r metrics.Registry) {
	stats, _ := disk.Partitions(true)
	for _, s := range stats {
		partitions = append(partitions, s.Mountpoint)
	}

	systemMetrics.CPUStat.User = metrics.NewGauge()
	systemMetrics.CPUStat.System = metrics.NewGauge()
	systemMetrics.CPUStat.Idle = metrics.NewGauge()
	systemMetrics.CPUStat.Iowait = metrics.NewGauge()
	systemMetrics.LoadStat.Load1 = metrics.NewGauge()
	systemMetrics.LoadStat.Load5 = metrics.NewGauge()
	systemMetrics.LoadStat.Load15 = metrics.NewGauge()
	systemMetrics.MemStat.Total = metrics.NewGauge()
	systemMetrics.MemStat.Available = metrics.NewGauge()
	systemMetrics.MemStat.Used = metrics.NewGauge()
	systemMetrics.SwapMemStat.Total = metrics.NewGauge()
	systemMetrics.SwapMemStat.Free = metrics.NewGauge()
	systemMetrics.SwapMemStat.Used = metrics.NewGauge()
	systemMetrics.DiskStat = make(map[string]*struct {
		Total metrics.Gauge
		Free  metrics.Gauge
	}, len(partitions))
	for _, p := range partitions {
		systemMetrics.DiskStat[p] = &struct {
			Total metrics.Gauge
			Free  metrics.Gauge
		}{
			Total: metrics.NewGauge(),
			Free:  metrics.NewGauge(),
		}
	}

	systemMetrics.BandwidthStat = make(map[string]*struct {
		BytesSent   metrics.Gauge
		BytesRecv   metrics.Gauge
		PacketsSent metrics.Gauge
		PacketsRecv metrics.Gauge
	})
	systemMetrics.captureSystemTimer = metrics.NewTimer()

	r.Register("cpu.User", systemMetrics.CPUStat.User)
	r.Register("cpu.System", systemMetrics.CPUStat.System)
	r.Register("cpu.Idle", systemMetrics.CPUStat.Idle)
	r.Register("cpu.Iowait", systemMetrics.CPUStat.Iowait)
	r.Register("load.Load1", systemMetrics.LoadStat.Load1)
	r.Register("load.Load5", systemMetrics.LoadStat.Load5)
	r.Register("load.Load15", systemMetrics.LoadStat.Load15)
	r.Register("mem.Total", systemMetrics.MemStat.Total)
	r.Register("mem.Available", systemMetrics.MemStat.Available)
	r.Register("mem.Used", systemMetrics.MemStat.Used)
	r.Register("swapmem.Total", systemMetrics.SwapMemStat.Total)
	r.Register("swapmem.Free", systemMetrics.SwapMemStat.Free)
	r.Register("swapmem.Used", systemMetrics.SwapMemStat.Used)
	for _, p := range partitions {
		systemMetrics.DiskStat[p].Total = metrics.NewGauge()
		systemMetrics.DiskStat[p].Free = metrics.NewGauge()

		r.Register("disk."+p+".Total", systemMetrics.DiskStat[p].Total)
		r.Register("disk."+p+".Free", systemMetrics.DiskStat[p].Free)
	}
	r.Register("capture_system", systemMetrics.captureSystemTimer)
}

func getAllCPUTime(t cpu.TimesStat) (float64, float64) {
	return 0, t.User + t.System + t.Nice + t.Iowait + t.Irq +
		t.Softirq + t.Steal + t.Guest + t.GuestNice + t.Stolen + +t.Idle
}
