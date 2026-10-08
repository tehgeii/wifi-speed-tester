// Package monitor runs the stability monitor: one probe per second to an
// internet target (and the router, when known) for a user-chosen time while
// the window is open. It never runs in the background.
package monitor

import (
	"context"
	"math"
	"net"
	"sort"
	"sync"
	"time"

	"github.com/tehgeii/wifi-speed-tester/internal/measure"
	"github.com/tehgeii/wifi-speed-tester/internal/model"
)

// MaxDuration caps a run.
const MaxDuration = 30 * time.Minute

// Options controls a run.
type Options struct {
	Duration time.Duration
	Interval time.Duration // default 1 s
	Timeout  time.Duration // per probe, default 1 s
	// OnSample is called after each probe, serialized, possibly out of order.
	OnSample func(model.MonitorSample)
}

// Run probes target (and gateway, when gwPinger is non-nil) until the
// duration ends or ctx is cancelled. It returns the samples in time order
// and whether the run was cut short.
func Run(ctx context.Context, p measure.Pinger, ip net.IP, gwPinger measure.Pinger, gwIP net.IP, o Options) ([]model.MonitorSample, bool) {
	if o.Interval <= 0 {
		o.Interval = time.Second
	}
	if o.Timeout <= 0 || o.Timeout > o.Interval {
		o.Timeout = o.Interval
	}
	if o.Duration <= 0 || o.Duration > MaxDuration {
		o.Duration = MaxDuration
	}
	runCtx, cancel := context.WithTimeout(ctx, o.Duration)
	defer cancel()

	var mu sync.Mutex
	var samples []model.MonitorSample
	var wg sync.WaitGroup
	start := time.Now()
	ticker := time.NewTicker(o.Interval)
	defer ticker.Stop()
	for i := 0; ; i++ {
		t := float64(i) * o.Interval.Seconds()
		wg.Add(1)
		go func() {
			defer wg.Done()
			s := model.MonitorSample{T: t}
			var inner sync.WaitGroup
			if gwPinger != nil && gwIP != nil {
				inner.Add(1)
				go func() {
					defer inner.Done()
					_, err := gwPinger.Ping(runCtx, gwIP, o.Timeout)
					ok := err == nil
					s.GwOK = &ok
				}()
			}
			rtt, err := p.Ping(runCtx, ip, o.Timeout)
			inner.Wait()
			if runCtx.Err() != nil && err != nil {
				return // interrupted at the end, not lost
			}
			s.OK = err == nil
			if s.OK {
				s.RttMs = measure.RTTMs(rtt)
			}
			mu.Lock()
			samples = append(samples, s)
			if o.OnSample != nil {
				o.OnSample(s)
			}
			mu.Unlock()
		}()
		select {
		case <-runCtx.Done():
		case <-ticker.C:
			if time.Since(start)+o.Interval/2 < o.Duration {
				continue
			}
			<-runCtx.Done()
		}
		break
	}
	wg.Wait()
	sort.Slice(samples, func(a, b int) bool { return samples[a].T < samples[b].T })
	return samples, ctx.Err() != nil
}

// Summarize computes the statistics of a run. interval is the probe spacing.
func Summarize(samples []model.MonitorSample, interval time.Duration) model.MonitorSummary {
	var s model.MonitorSummary
	s.Sent = len(samples)
	var rtts []float64
	run, longest := 0, 0
	closeRun := func() {
		if run >= 3 {
			s.Outages++
		}
		longest = max(longest, run)
		run = 0
	}
	var prev float64
	havePrev := false
	var jsum float64
	jn := 0
	for _, x := range samples {
		if x.GwOK != nil {
			s.GatewaySent++
			if !*x.GwOK {
				s.GatewayLost++
			}
		}
		if !x.OK {
			run++
			continue
		}
		closeRun()
		rtts = append(rtts, x.RttMs)
		if havePrev {
			jsum += math.Abs(x.RttMs - prev)
			jn++
		}
		prev, havePrev = x.RttMs, true
	}
	closeRun()
	if longest >= 3 {
		s.LongestOutage = float64(longest) * interval.Seconds()
	}
	s.Received = len(rtts)
	s.LossPct = measure.PacketLoss(s.Sent, s.Received)
	if jn > 0 {
		s.JitterMs = jsum / float64(jn)
	}
	if len(rtts) == 0 {
		return s
	}
	sorted := append([]float64(nil), rtts...)
	sort.Float64s(sorted)
	s.MinMs, s.MaxMs = sorted[0], sorted[len(sorted)-1]
	var sum float64
	for _, v := range rtts {
		sum += v
	}
	s.AvgMs = sum / float64(len(rtts))
	s.P95Ms = sorted[int(math.Ceil(0.95*float64(len(sorted))))-1]
	med := measure.Median(rtts)
	limit := math.Max(2*med, med+50)
	for _, v := range rtts {
		if v > limit {
			s.Spikes++
		}
	}
	return s
}
