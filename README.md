# go-app-metrics

provider a out-of-the-box go metrics lib to collect stats of the machine and go runtime.


## metrics

### package rmetric

Package `rmetric` provides method to collect metrics of go **runtime** so it is called as rmetric (runtime metrics).

You can use it to get:
```go
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
```

You can check `expvar` to see how to use them to collect metrics which add metrics to `expvar`, and you can use the below url to see metrics:
```sh
http://xxx.xxx.xxx.xxx/debug/vars
```

Of course you can add these metrics in your web frameworks just like `expvar`.

### package system

Package `system` provides method to collect metrics of **machines**. 

You can use it to get:

```go
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
```


## Credits

- [shirou/gopsutil](https://github.com/shirou/gopsutil)
- [tevjef/go-runtime-metrics](https://github.com/tevjef/go-runtime-metrics)
- [bmhatfield/go-runtime-metrics](https://github.com/bmhatfield/go-runtime-metrics)
