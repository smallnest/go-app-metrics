package stat

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStats(t *testing.T) {
	r, err := http.NewRequest("GET", "http://localhost:8000/debug/stats?seconds=1", nil)
	assert.Nil(t, err)

	w := httptest.NewRecorder()
	Stats(w, r)

	resp := w.Result()
	body, _ := io.ReadAll(resp.Body)
	stats := string(body)

	expKeys := []string{
		"cpu.goroutines",
		"mem.lookups",
		"mem.gc.count",
		"cpu.user",
		"load.load1",
		"mem.total",
		"swap.total",
	}

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	for _, k := range expKeys {
		assert.Contains(t, stats, k)
	}
}
