package exp

import (
	"context"
	"expvar"
	"strings"
	"time"

	"github.com/smallnest/go-app-metrics/rmetric"
	"github.com/smallnest/go-app-metrics/system"
)

var (
	rmetricMap = expvar.NewMap("rmetricStats")
	systemMap  = expvar.NewMap("systemStats")
)

// Run starts a collector to collect system stats and go runtime stats,
// and writes them in expvar variables named as `rmetricStats` and `systemStats`.
func Run(ctx context.Context, interval time.Duration) {
	c := rmetric.New(runtimeStatsCallback)
	c.CollectInterval = interval
	c.Done = ctx.Done()
	go c.Run()

	sc := system.New(systemStatsCallback)
	sc.CollectInterval = interval
	sc.Done = ctx.Done()
	go sc.Run()
}

func runtimeStatsCallback(stats rmetric.RuntimeStats) {
	values := stats.Values()
	for k, v := range values {
		va := rmetricMap.Get(k)

		if k == "mem.gc.cpu_fraction" {
			if va == nil {
				va = new(expvar.Float)
				rmetricMap.Set(k, va)
			}
			va.(*expvar.Float).Set(v.(float64))
			continue
		}
		if va == nil {
			va = new(expvar.Int)
			rmetricMap.Set(k, va)
		}
		va.(*expvar.Int).Set(v.(int64))
	}
}

func systemStatsCallback(stats system.SystemStats) {
	values := stats.Values()
	for k, v := range values {
		va := systemMap.Get(k)

		if strings.HasPrefix(k, "cpu.") || strings.HasPrefix(k, "load.") {
			if va == nil {
				va = new(expvar.Float)
				systemMap.Set(k, va)
			}
			systemMap.Get(k).(*expvar.Float).Set(v.(float64))
			continue
		}
		if va == nil {
			va = new(expvar.Int)
			systemMap.Set(k, va)
		}
		va.(*expvar.Int).Set(int64(v.(uint64)))
	}
}
