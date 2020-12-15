package system

import (
	"log"
	"os"
	"testing"
	"time"

	metrics "github.com/rcrowley/go-metrics"
)

func TestMonitor(t *testing.T) {
	go metrics.Log(metrics.DefaultRegistry, 4*time.Second, log.New(os.Stderr, "metrics: ", log.Lmicroseconds))

	RegisterSystemStats(metrics.DefaultRegistry)

	go CaptureSystemStats(metrics.DefaultRegistry, time.Second)
	time.Sleep(5 * time.Second)
}
