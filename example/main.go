package main

import (
	"net"
	"time"

	graphite "github.com/cyberdelia/go-metrics-graphite"
	metrics "github.com/rcrowley/go-metrics"
	"github.com/smallnest/go-metrics-system"
)

func main() {

	addr, _ := net.ResolveTCPAddr("tcp", "127.0.0.1:2003")
	go graphite.Graphite(metrics.DefaultRegistry, 1e9, "system.127_0_0_1", addr)

	monitor.RegisterSystemStats(metrics.DefaultRegistry)
	go monitor.CaptureSystemStats(metrics.DefaultRegistry, time.Second)

	i := 0
	for {
		i = i + 1
	}
}
