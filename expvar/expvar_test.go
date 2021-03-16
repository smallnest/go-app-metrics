package exp

import (
	"context"
	"expvar"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCollector(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping test because testing.Short is enabled")
	}

	ctx, cancel := context.WithCancel(context.Background())
	go Run(ctx, time.Second)
	time.Sleep(time.Second)
	cancel()

	expKeys := []string{
		"cpu.goroutines",
		"mem.lookups",
		"mem.gc.count",
	}
	rmetricMap := expvar.Get("rmetricStats").(*expvar.Map)
	assert.NotNil(t, rmetricMap)
	for _, expKey := range expKeys {
		if v := rmetricMap.Get(expKey); v == nil {
			t.Errorf("expected key (%s) not found", expKey)
		}
	}

	expKeys = []string{
		"cpu.user",
		"load.load1",
		"mem.total",
		"swap.total",
	}
	systemMap := expvar.Get("systemStats").(*expvar.Map)
	assert.NotNil(t, systemMap)
	for _, expKey := range expKeys {
		if v := systemMap.Get(expKey); v == nil {
			t.Errorf("expected key (%s) not found", expKey)
		}
	}
}
