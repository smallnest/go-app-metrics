package system

import (
	"testing"
	"time"
)

func TestCollectorOnce(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping test because testing.Short is enabled")
	}

	c := New(nil)
	time.Sleep(time.Second)
	stats := c.Once()

	expKeys := []string{
		"cpu.user",
		"load.load1",
		"mem.total",
		"swap.total",
	}

	for _, expKey := range expKeys {
		if _, ok := stats.Values()[expKey]; !ok {
			t.Errorf("expected key (%s) not found", expKey)
		}
	}
}
func TestCollector(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping test because testing.Short is enabled")
	}

	latestStats := []SystemStats{}
	callback := func(stats SystemStats) {
		latestStats = append(latestStats, stats)
	}

	done := make(chan struct{})
	collectorShutdown := make(chan struct{})
	c := New(callback)
	c.CollectInterval = 100 * time.Millisecond
	c.Done = done

	go func() {
		defer close(collectorShutdown)
		c.Run()
	}()
	time.Sleep(time.Second)
	close(done)
	<-collectorShutdown

	expKeys := []string{
		"cpu.user",
		"load.load1",
		"mem.total",
		"swap.total",
	}

	for _, stats := range latestStats {
		for _, expKey := range expKeys {
			if _, ok := stats.Values()[expKey]; !ok {
				t.Errorf("expected key (%s) not found", expKey)
			}
		}
	}

	expected := 10
	if stats := len(latestStats); stats < expected {
		t.Errorf("num of points is lower than expected:\ngot: %d\nexp: %d", stats, expected)
	}

}
