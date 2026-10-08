package measure

import (
	"context"
	"errors"
	"net"
	"sync"
	"time"

	"github.com/tehgeii/wifi-speed-tester/internal/model"
)

// ErrTimeout is returned by a Pinger when no echo reply arrived in time.
var ErrTimeout = errors.New("request timed out")

// MinRTTMs stands in for a reply faster than the clock can resolve: ICMP
// reports whole milliseconds and Windows' monotonic clock is coarse, so a
// LAN or loopback reply can read as 0. An answered probe must never look
// unmeasured, since a latency of 0 means "not measured" everywhere else.
const MinRTTMs = 0.01

// RTTMs converts the round-trip time of an answered probe to milliseconds.
func RTTMs(d time.Duration) float64 {
	return max(float64(d)/float64(time.Millisecond), MinRTTMs)
}

// Pinger sends a single ICMP echo request and returns the round-trip time.
// Implementations are platform specific (IcmpSendEcho on Windows, unprivileged
// ICMP sockets elsewhere).
type Pinger interface {
	Ping(ctx context.Context, ip net.IP, timeout time.Duration) (time.Duration, error)
	Close() error
}

// NewPinger returns the platform pinger.
func NewPinger() (Pinger, error) { return newPlatformPinger() }

// PingOptions controls a ping run.
type PingOptions struct {
	Count    int
	Interval time.Duration
	Timeout  time.Duration
	// OnReply is called when a probe completes with its index (0-based) and
	// the RTT, or ok=false when it was lost. Calls are serialized but may
	// arrive out of order. May be nil.
	OnReply func(i int, rttMs float64, ok bool)
}

// ResolveIP resolves host to a single IP, preferring IPv4.
func ResolveIP(ctx context.Context, host string) (net.IP, error) {
	if ip := net.ParseIP(host); ip != nil {
		return ip, nil
	}
	addrs, err := net.DefaultResolver.LookupIPAddr(ctx, host)
	if err != nil {
		return nil, err
	}
	for _, a := range addrs {
		if v4 := a.IP.To4(); v4 != nil {
			return v4, nil
		}
	}
	if len(addrs) > 0 {
		return addrs[0].IP, nil
	}
	return nil, errors.New("no address found")
}

// RunPing pings ip opts.Count times, one probe every opts.Interval, and
// summarises the run. Like the ping command, probes do not wait for earlier
// replies, so an unresponsive target takes Count*Interval + Timeout rather
// than Count*Timeout. A cancelled context stops early; probes that completed
// are still summarised, and the context error is returned.
func RunPing(ctx context.Context, p Pinger, label, host string, ip net.IP, opts PingOptions) (model.PingStats, error) {
	type reply struct {
		rtt  float64
		ok   bool
		done bool
	}
	replies := make([]reply, opts.Count)
	var mu sync.Mutex
	var lastErr error
	var wg sync.WaitGroup
	ticker := time.NewTicker(opts.Interval)
	defer ticker.Stop()

	launched := 0
loop:
	for i := 0; i < opts.Count; i++ {
		if i > 0 {
			select {
			case <-ctx.Done():
				break loop
			case <-ticker.C:
			}
		}
		if ctx.Err() != nil {
			break
		}
		launched++
		wg.Add(1)
		go func() {
			defer wg.Done()
			rtt, err := p.Ping(ctx, ip, opts.Timeout)
			if ctx.Err() != nil {
				return // interrupted, neither received nor lost
			}
			ms := 0.0
			if err == nil {
				ms = RTTMs(rtt)
			}
			mu.Lock()
			replies[i] = reply{rtt: ms, ok: err == nil, done: true}
			if err != nil {
				lastErr = err
			}
			if opts.OnReply != nil {
				opts.OnReply(i, ms, err == nil)
			}
			mu.Unlock()
		}()
	}
	wg.Wait()

	// Samples in send order, so jitter compares consecutive probes.
	var samples []float64
	sent := 0
	for _, r := range replies[:launched] {
		if !r.done {
			continue
		}
		sent++
		if r.ok {
			samples = append(samples, r.rtt)
		}
	}
	st := Summarize(label, host, sent, samples)
	if ctx.Err() != nil {
		return st, ctx.Err()
	}
	if st.Received == 0 && lastErr != nil {
		st.Error = lastErr.Error()
	}
	return st, nil
}
