// Package regions measures latency to cloud regions (Jakarta, Singapore,
// Tokyo, ...) by timing TCP connections, the same way the speed test falls
// back to when ICMP is blocked. Game and streaming services run in these
// regions, so the numbers show what to expect from them, without claiming
// to be the ping of any particular game server.
package regions

import (
	"context"
	"net"
	"sort"
	"strconv"
	"sync"
	"time"

	"github.com/tehgeii/wifi-speed-tester/internal/config"
	"github.com/tehgeii/wifi-speed-tester/internal/measure"
	"github.com/tehgeii/wifi-speed-tester/internal/model"
)

// Options controls a run.
type Options struct {
	Probes   int           // connections per region (default 5)
	Interval time.Duration // between probes to one region (default 200 ms)
	Timeout  time.Duration // per probe (default 2 s)
	// OnResult is called as each region finishes, from its own goroutine.
	OnResult func(model.RegionResult)
}

// Run measures every region in parallel and returns the results fastest
// first, unreachable regions last.
func Run(ctx context.Context, targets []config.PingTarget, o Options) []model.RegionResult {
	if o.Probes <= 0 {
		o.Probes = 5
	}
	if o.Interval <= 0 {
		o.Interval = 200 * time.Millisecond
	}
	if o.Timeout <= 0 {
		o.Timeout = 2 * time.Second
	}
	out := make([]model.RegionResult, len(targets))
	var wg sync.WaitGroup
	var mu sync.Mutex
	for i, t := range targets {
		wg.Add(1)
		go func() {
			defer wg.Done()
			// Stagger the start a little so the regions do not all open
			// connections in the same instant.
			select {
			case <-ctx.Done():
			case <-time.After(time.Duration(i) * 30 * time.Millisecond):
			}
			r := measureOne(ctx, t, o)
			out[i] = r
			if o.OnResult != nil {
				mu.Lock()
				o.OnResult(r)
				mu.Unlock()
			}
		}()
	}
	wg.Wait()
	sort.SliceStable(out, func(a, b int) bool {
		if (out[a].OK == 0) != (out[b].OK == 0) {
			return out[b].OK == 0
		}
		return out[a].MedianMs < out[b].MedianMs
	})
	return out
}

func measureOne(ctx context.Context, t config.PingTarget, o Options) model.RegionResult {
	r := model.RegionResult{Label: t.Label, Host: t.Host}
	host, port := t.Host, 443
	if h, p, err := net.SplitHostPort(t.Host); err == nil {
		if n, err := strconv.Atoi(p); err == nil {
			host, port = h, n
		}
	}
	ip, err := measure.ResolveIP(ctx, host)
	if err != nil {
		r.Error = err.Error()
		return r
	}
	p := measure.TCPPinger{Port: port}
	var times []float64
	var lastErr error
	for k := 0; k < o.Probes && ctx.Err() == nil; k++ {
		if k > 0 {
			select {
			case <-ctx.Done():
			case <-time.After(o.Interval):
			}
		}
		r.Sent++
		rtt, err := p.Ping(ctx, ip, o.Timeout)
		if err != nil {
			lastErr = err
			continue
		}
		times = append(times, measure.RTTMs(rtt))
	}
	r.OK = len(times)
	if r.OK > 0 {
		r.MedianMs = measure.Median(times)
		sort.Float64s(times)
		r.MinMs = times[0]
	} else if lastErr != nil {
		r.Error = lastErr.Error()
	}
	return r
}
