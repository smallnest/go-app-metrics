// Package system provides method to collect metrics of machines.
package system

import (
	"time"

	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/disk"
	"github.com/shirou/gopsutil/load"
	"github.com/shirou/gopsutil/mem"
	"github.com/shirou/gopsutil/net"
)

// SystemStatsHandler represents a handler to handle stats after successfully gathering statistics
type SystemStatsHandler func(SystemStats)

// Collector implements the periodic grabbing of informational data of go runtime to a SystemStatsHandler.
type Collector struct {
	// CollectInterval represents the interval in-between each set of stats output.
	// Defaults to 10 seconds.
	CollectInterval time.Duration

	cpuStat    *cpu.TimesStat
	partitions []string
	netStats   map[string]*net.IOCountersStat

	// Done, when closed, is used to signal Collector that is should stop collecting
	// statistics and the Run function should return.
	Done <-chan struct{}

	statsHandler SystemStatsHandler
}

// New creates a new Collector that will periodically output statistics to statsHandler. It
// will also set the values of the exported stats to the described defaults. The values
// of the exported defaults can be changed at any point before Run is called.
func New(statsHandler SystemStatsHandler) *Collector {
	if statsHandler == nil {
		statsHandler = func(SystemStats) {}
	}

	var partitions []string
	stats, _ := disk.Partitions(true)
	for _, s := range stats {
		partitions = append(partitions, s.Mountpoint)
	}

	return &Collector{
		CollectInterval: 10 * time.Second,
		partitions:      partitions,
		netStats:        make(map[string]*net.IOCountersStat),
		statsHandler:    statsHandler,
	}
}

// Run gathers statistics then outputs them to the configured SystemStatsHandler every
// CollectInterval. Unlike Once, this function will return until Done has been closed
// (or never if Done is nil), therefore it should be called in its own goroutine.
func (c *Collector) Run() {
	c.statsHandler(c.collectStats())

	tick := time.NewTicker(c.CollectInterval)
	defer tick.Stop()
	for {
		select {
		case <-c.Done:
			return
		case <-tick.C:
			c.statsHandler(c.collectStats())
		}
	}
}

// Once returns a map containing all statistics. It is safe for use from multiple go routines。
func (c *Collector) Once() SystemStats {
	return c.collectStats()
}

// collectStats collects all configured stats once.
func (c *Collector) collectStats() SystemStats {
	stats := SystemStats{
		DiskStat:      make(map[string]DiskStat),
		BandwidthStat: make(map[string]BandwidthStat),
	}

	cpuStat := c.cpuStat

	//cpu * 100
	cpustats, err := cpu.Times(false)
	if err == nil && len(cpustats) > 0 {
		cpustat2 := cpustats[0]
		if cpuStat == nil {
			cpuStat = &cpustat2
		}
		total1 := cpuStat.Total()
		total2 := cpustat2.Total()
		total := total2 - total1
		if total > 0 {
			stats.CPUStat.User = (cpustat2.User - cpuStat.User) * 100 / total
			stats.CPUStat.System = (cpustat2.System - cpuStat.System) * 100 / total
			stats.CPUStat.Iowait = (cpustat2.Iowait - cpuStat.Iowait) * 100 / total
			stats.CPUStat.Idle = (cpustat2.Idle - cpuStat.Idle) * 100 / total
		}
		c.cpuStat = &cpustat2
	}

	//load * 100
	avg, err := load.Avg()
	if err == nil {
		stats.LoadStat.Load1 = avg.Load1
		stats.LoadStat.Load5 = avg.Load5
		stats.LoadStat.Load15 = avg.Load15
	}

	//mem
	vmem, err := mem.VirtualMemory()
	if err == nil {
		stats.MemStat.Total = vmem.Total
		stats.MemStat.Available = vmem.Available
		stats.MemStat.Used = vmem.Used
	}
	swapmem, err := mem.SwapMemory()
	if err == nil {
		stats.SwapMemStat.Total = swapmem.Total
		stats.SwapMemStat.Free = swapmem.Free
		stats.SwapMemStat.Used = swapmem.Used
	}

	//disk
	for _, p := range c.partitions {
		s, err := disk.Usage(p)
		if err != nil {
			continue
		}

		var diskStat DiskStat
		diskStat.Total = s.Total
		diskStat.Free = s.Free
		stats.DiskStat[p] = diskStat
	}

	//bandwidth
	netstats, err := net.IOCounters(true)
	netStats := c.netStats
	if err == nil {
		for _, s := range netstats {
			s := s
			if netStats[s.Name] == nil {
				netStats[s.Name] = &s
			}
			s2 := netStats[s.Name]

			var bandwidthStat BandwidthStat
			bandwidthStat.BytesSent = s.BytesSent - s2.BytesSent
			bandwidthStat.BytesRecv = s.BytesRecv - s2.BytesRecv
			bandwidthStat.PacketsSent = s.PacketsSent - s2.PacketsSent
			bandwidthStat.PacketsRecv = s.PacketsRecv - s2.PacketsRecv
			stats.BandwidthStat[s.Name] = bandwidthStat
			netStats[s.Name] = &s
		}
	}

	return stats
}

type SystemStats struct {
	CPUStat struct {
		User   float64
		System float64
		Idle   float64
		Iowait float64
	}
	LoadStat struct {
		Load1  float64
		Load5  float64
		Load15 float64
	}
	MemStat struct {
		Total     uint64
		Available uint64
		Used      uint64
	}
	SwapMemStat struct {
		Total uint64
		Free  uint64
		Used  uint64
	}
	DiskStat      map[string]DiskStat
	BandwidthStat map[string]BandwidthStat
}

type DiskStat struct {
	Total uint64
	Free  uint64
}

type BandwidthStat struct {
	BytesSent   uint64
	BytesRecv   uint64
	PacketsSent uint64
	PacketsRecv uint64
}

// Values returns metrics which you can write into TSDB.
func (ss *SystemStats) Values() map[string]interface{} {
	values := map[string]interface{}{
		"cpu.user":   ss.CPUStat.User,
		"cpu.system": ss.CPUStat.System,
		"cpu.idle":   ss.CPUStat.Idle,
		"cpu.iowait": ss.CPUStat.Iowait,

		"load.load1":  ss.LoadStat.Load1,
		"load.load5":  ss.LoadStat.Load5,
		"load.load15": ss.LoadStat.Load15,

		"mem.total":     ss.MemStat.Total,
		"mem.available": ss.MemStat.Available,
		"mem.used":      ss.MemStat.Used,
		"swap.total":    ss.SwapMemStat.Total,
		"swap.free":     ss.SwapMemStat.Free,
		"swap.used":     ss.SwapMemStat.Used,
	}

	for partition, stat := range ss.DiskStat {
		values["disk."+partition+".total"] = stat.Total
		values["disk."+partition+".free"] = stat.Free
	}

	for n, stat := range ss.BandwidthStat {
		values["net."+n+".bytes_sent"] = stat.BytesSent
		values["net."+n+".bytes_recv"] = stat.BytesRecv
		values["net."+n+".packets_sent"] = stat.PacketsSent
		values["net."+n+".packets_recv"] = stat.PacketsRecv
	}

	return values
}
