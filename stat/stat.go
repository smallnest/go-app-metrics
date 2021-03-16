package stat

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/smallnest/go-app-metrics/rmetric"
	"github.com/smallnest/go-app-metrics/system"
)

func init() {
	http.HandleFunc("/debug/stats/", Stats)
}

// Stats responds with system stats and go runtime stats.
// Each metric is a line and has key=value format.
func Stats(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")

	sec, err := strconv.ParseInt(r.FormValue("seconds"), 10, 64)
	if sec <= 0 || err != nil {
		sec = 30
	}

	c := rmetric.New(nil)
	sc := system.New(nil)

	time.Sleep(time.Duration(sec) * time.Second)

	rstats := c.Once()
	sstats := sc.Once()

	var buf strings.Builder
	for k, v := range rstats.Values() {
		buf.WriteString(fmt.Sprintf("%s=%v\n", k, v))
	}
	for k, v := range sstats.Values() {
		buf.WriteString(fmt.Sprintf("%s=%v\n", k, v))
	}
	w.Write([]byte(buf.String()))
}
