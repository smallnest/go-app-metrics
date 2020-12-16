// Package rmetric provides method to collect metrics of go runtime so it is called as rmetric (runtime metrics).
package rmetric

import (
	"runtime"
	"runtime/pprof"
	"time"
)

// threadProfile for getting number of threads
var threadProfile = pprof.Lookup("threadcreate")

// RuntimeStatsHandler represents a handler to handle stats after successfully gathering statistics
type RuntimeStatsHandler func(RuntimeStats)

// Collector implements the periodic grabbing of informational data of go runtime to a RuntimeStatsHandler.
type Collector struct {
	// CollectInterval represents the interval in-between each set of stats output.
	// Defaults to 10 seconds.
	CollectInterval time.Duration

	// EnableCPU determines whether CPU statistics will be output. Defaults to true.
	EnableCPU bool

	// EnableMem determines whether memory statistics will be output. Defaults to true.
	EnableMem bool

	// EnableGC determines whether garbage collection statistics will be output. EnableMem
	// must also be set to true for this to take affect. Defaults to true.
	EnableGC bool

	// Done, when closed, is used to signal Collector that is should stop collecting
	// statistics and the Run function should return.
	Done <-chan struct{}

	statsHandler RuntimeStatsHandler
}

// New creates a new Collector that will periodically output statistics to statsHandler. It
// will also set the values of the exported stats to the described defaults. The values
// of the exported defaults can be changed at any point before Run is called.
func New(statsHandler RuntimeStatsHandler) *Collector {
	if statsHandler == nil {
		statsHandler = func(RuntimeStats) {}
	}

	return &Collector{
		CollectInterval: 10 * time.Second,
		EnableCPU:       true,
		EnableMem:       true,
		EnableGC:        true,
		statsHandler:    statsHandler,
	}
}

// Run gathers statistics then outputs them to the configured RuntimeStatsHandler every
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
func (c *Collector) Once() RuntimeStats {
	return c.collectStats()
}

// collectStats collects all configured stats once.
func (c *Collector) collectStats() RuntimeStats {
	stats := RuntimeStats{}

	if c.EnableCPU {
		cStats := cpuStats{
			NumGoroutine: int64(runtime.NumGoroutine()),
			NumThread:    int64(threadProfile.Count()),
			NumCgoCall:   int64(runtime.NumCgoCall()),
			NumCPU:       int64(runtime.NumCPU()),
		}
		c.collectCPUStats(&stats, &cStats)
	}
	if c.EnableMem {
		m := &runtime.MemStats{}
		runtime.ReadMemStats(m)
		c.collectMemStats(&stats, m)
		if c.EnableGC {
			c.collectGCStats(&stats, m)
		}
	}

	stats.Goos = runtime.GOOS
	stats.Goarch = runtime.GOARCH
	stats.Version = runtime.Version()

	return stats
}

func (*Collector) collectCPUStats(stats *RuntimeStats, s *cpuStats) {
	stats.NumCPU = s.NumCPU
	stats.NumGoroutine = s.NumGoroutine
	stats.NumThread = s.NumThread
	stats.NumCgoCall = s.NumCgoCall
}

func (*Collector) collectMemStats(stats *RuntimeStats, m *runtime.MemStats) {
	// General
	stats.Alloc = int64(m.Alloc)
	stats.TotalAlloc = int64(m.TotalAlloc)
	stats.Sys = int64(m.Sys)
	stats.Lookups = int64(m.Lookups)
	stats.Mallocs = int64(m.Mallocs)
	stats.Frees = int64(m.Frees)

	// Heap
	stats.HeapAlloc = int64(m.HeapAlloc)
	stats.HeapSys = int64(m.HeapSys)
	stats.HeapIdle = int64(m.HeapIdle)
	stats.HeapInuse = int64(m.HeapInuse)
	stats.HeapReleased = int64(m.HeapReleased)
	stats.HeapObjects = int64(m.HeapObjects)

	// Stack
	stats.StackInuse = int64(m.StackInuse)
	stats.StackSys = int64(m.StackSys)
	stats.MSpanInuse = int64(m.MSpanInuse)
	stats.MSpanSys = int64(m.MSpanSys)
	stats.MCacheInuse = int64(m.MCacheInuse)
	stats.MCacheSys = int64(m.MCacheSys)

	stats.OtherSys = int64(m.OtherSys)
}

func (*Collector) collectGCStats(stats *RuntimeStats, m *runtime.MemStats) {
	stats.GCSys = int64(m.GCSys)
	stats.NextGC = int64(m.NextGC)
	stats.LastGC = int64(m.LastGC)
	stats.PauseTotalNs = int64(m.PauseTotalNs)
	stats.PauseNs = int64(m.PauseNs[(m.NumGC+255)%256])
	stats.NumGC = int64(m.NumGC)
	stats.GCCPUFraction = float64(m.GCCPUFraction)
}

type cpuStats struct {
	NumCPU       int64
	NumGoroutine int64
	NumThread    int64
	NumCgoCall   int64
}

// RuntimeStats represents metrics of go runtime.
type RuntimeStats struct {
	// CPU
	NumCPU       int64 `json:"cpu.count"`
	NumThread    int64 `json:"cpu.threads"`
	NumGoroutine int64 `json:"cpu.goroutines"`
	NumCgoCall   int64 `json:"cpu.cgo_calls"`

	// General
	Alloc      int64 `json:"mem.alloc"`
	TotalAlloc int64 `json:"mem.total"`
	Sys        int64 `json:"mem.sys"`
	Lookups    int64 `json:"mem.lookups"`
	Mallocs    int64 `json:"mem.malloc"`
	Frees      int64 `json:"mem.frees"`

	// Heap
	HeapAlloc    int64 `json:"mem.heap.alloc"`
	HeapSys      int64 `json:"mem.heap.sys"`
	HeapIdle     int64 `json:"mem.heap.idle"`
	HeapInuse    int64 `json:"mem.heap.inuse"`
	HeapReleased int64 `json:"mem.heap.released"`
	HeapObjects  int64 `json:"mem.heap.objects"`

	// Stack
	StackInuse  int64 `json:"mem.stack.inuse"`
	StackSys    int64 `json:"mem.stack.sys"`
	MSpanInuse  int64 `json:"mem.stack.mspan_inuse"`
	MSpanSys    int64 `json:"mem.stack.mspan_sys"`
	MCacheInuse int64 `json:"mem.stack.mcache_inuse"`
	MCacheSys   int64 `json:"mem.stack.mcache_sys"`

	OtherSys int64 `json:"mem.othersys"`

	// GC
	GCSys         int64   `json:"mem.gc.sys"`
	NextGC        int64   `json:"mem.gc.next"`
	LastGC        int64   `json:"mem.gc.last"`
	PauseTotalNs  int64   `json:"mem.gc.pause_total"`
	PauseNs       int64   `json:"mem.gc.pause"`
	NumGC         int64   `json:"mem.gc.count"`
	GCCPUFraction float64 `json:"mem.gc.cpu_fraction"`

	Goarch  string `json:"-"`
	Goos    string `json:"-"`
	Version string `json:"-"`
}

// Tags return go arch.
func (f *RuntimeStats) Tags() map[string]string {
	return map[string]string{
		"go.os":      f.Goos,
		"go.arch":    f.Goarch,
		"go.version": f.Version,
	}
}

// Values returns metrics which you can write into TSDB.
func (f *RuntimeStats) Values() map[string]interface{} {
	return map[string]interface{}{
		"cpu.count":      f.NumCPU,
		"cpu.threads":    f.NumThread,
		"cpu.goroutines": f.NumGoroutine,
		"cpu.cgo_calls":  f.NumCgoCall,

		"mem.alloc":   f.Alloc,
		"mem.total":   f.TotalAlloc,
		"mem.sys":     f.Sys,
		"mem.lookups": f.Lookups,
		"mem.malloc":  f.Mallocs,
		"mem.frees":   f.Frees,

		"mem.heap.alloc":    f.HeapAlloc,
		"mem.heap.sys":      f.HeapSys,
		"mem.heap.idle":     f.HeapIdle,
		"mem.heap.inuse":    f.HeapInuse,
		"mem.heap.released": f.HeapReleased,
		"mem.heap.objects":  f.HeapObjects,

		"mem.stack.inuse":        f.StackInuse,
		"mem.stack.sys":          f.StackSys,
		"mem.stack.mspan_inuse":  f.MSpanInuse,
		"mem.stack.mspan_sys":    f.MSpanSys,
		"mem.stack.mcache_inuse": f.MCacheInuse,
		"mem.stack.mcache_sys":   f.MCacheSys,
		"mem.othersys":           f.OtherSys,

		"mem.gc.sys":          f.GCSys,
		"mem.gc.next":         f.NextGC,
		"mem.gc.last":         f.LastGC,
		"mem.gc.pause_total":  f.PauseTotalNs,
		"mem.gc.pause":        f.PauseNs,
		"mem.gc.count":        f.NumGC,
		"mem.gc.cpu_fraction": float64(f.GCCPUFraction),
	}
}
