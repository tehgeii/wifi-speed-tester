package measure

import (
	"context"
	"crypto/rand"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/tehgeii/wifi-speed-tester/internal/model"
)

// HTTPStatusError reports a non-2xx response from a test server.
type HTTPStatusError struct {
	StatusCode int
	Status     string
}

func (e *HTTPStatusError) Error() string { return "server responded " + e.Status }

// ErrNoData means the test finished without transferring anything measurable.
var ErrNoData = errors.New("no data was transferred")

// SpeedOptions controls a download or upload run.
type SpeedOptions struct {
	Duration time.Duration // measured duration, including warm-up
	Warmup   time.Duration // initial period excluded from the result
	Streams  int           // parallel TCP connections
	// OnProgress receives the elapsed fraction (0..1) and the current
	// throughput over roughly the last second, about four times per second.
	OnProgress func(fraction, currentMbps float64)
}

// NewHTTPClient returns a client tuned for throughput testing: HTTP/1.1 only
// (so every stream is its own TCP connection) and no response compression.
func NewHTTPClient(streams int) *http.Client {
	dialer := &net.Dialer{Timeout: 10 * time.Second, KeepAlive: 30 * time.Second}
	tr := &http.Transport{
		Proxy:                 http.ProxyFromEnvironment,
		DialContext:           dialer.DialContext,
		TLSHandshakeTimeout:   10 * time.Second,
		ResponseHeaderTimeout: 15 * time.Second,
		MaxIdleConns:          streams * 2,
		MaxIdleConnsPerHost:   streams * 2,
		MaxConnsPerHost:       streams * 2,
		IdleConnTimeout:       30 * time.Second,
		DisableCompression:    true,
		ForceAttemptHTTP2:     false,
		TLSNextProto:          map[string]func(string, *tls.Conn) http.RoundTripper{},
		ReadBufferSize:        256 << 10,
		WriteBufferSize:       256 << 10,
	}
	return &http.Client{Transport: tr}
}

const (
	downloadMinChunk = 10_000_000  // first request size per stream
	downloadMaxChunk = 100_000_000 // largest request size
	maxStreamFails   = 5
	sampleEvery      = 250 * time.Millisecond
)

// sampler records a running byte counter at a fixed interval so the result
// can exclude the warm-up and produce a throughput curve.
type sampler struct {
	start   time.Time
	times   []time.Duration
	totals  []int64
	samples []float64
}

// add records the counter. Unless final, it also appends a chart sample:
// throughput over the trailing second, which smooths the bursts of writes
// into socket buffers. The final (partial) interval is not charted.
func (s *sampler) add(t time.Duration, total int64, final bool) {
	s.times = append(s.times, t)
	s.totals = append(s.totals, total)
	if !final && len(s.times) > 1 {
		s.samples = append(s.samples, s.recent(time.Second))
	}
}

// recent returns throughput over the trailing window.
func (s *sampler) recent(window time.Duration) float64 {
	n := len(s.times)
	if n < 2 {
		return 0
	}
	j := n - 1
	for j > 0 && s.times[n-1]-s.times[j-1] <= window {
		j--
	}
	return BytesToMbps(s.totals[n-1]-s.totals[j], (s.times[n-1] - s.times[j]).Seconds())
}

// at returns the counter value at the first sample at or after t.
func (s *sampler) at(t time.Duration) (time.Duration, int64) {
	for i, ti := range s.times {
		if ti >= t {
			return ti, s.totals[i]
		}
	}
	n := len(s.times) - 1
	return s.times[n], s.totals[n]
}

// streamErrors collects the first error and counts how many happened.
type streamErrors struct {
	mu    sync.Mutex
	first error
	count int
}

func (e *streamErrors) add(err error) int {
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.first == nil {
		e.first = err
	}
	e.count++
	return e.count
}

// retryDelay backs off longer when the server says it is rate limiting.
func retryDelay(err error) time.Duration {
	var se *HTTPStatusError
	if errors.As(err, &se) && se.StatusCode == http.StatusTooManyRequests {
		return time.Second
	}
	return 300 * time.Millisecond
}

func fatal(err error) bool {
	var se *HTTPStatusError
	if errors.As(err, &se) {
		return se.StatusCode >= 400 && se.StatusCode != http.StatusTooManyRequests
	}
	var dnsErr *net.DNSError
	return errors.As(err, &dnsErr)
}

// Download measures receive throughput. urlFor returns the URL that serves
// about n bytes (see config.Server.DownloadFor).
func Download(ctx context.Context, client *http.Client, urlFor func(n int) string, opts SpeedOptions) (model.SpeedResult, error) {
	res := model.SpeedResult{Streams: opts.Streams}
	runCtx, cancel := context.WithTimeout(ctx, opts.Duration)
	defer cancel()

	var total atomic.Int64
	var errs streamErrors
	var wg sync.WaitGroup
	for i := 0; i < opts.Streams; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			buf := make([]byte, 64<<10)
			chunk, maxChunk, lastGood := downloadMinChunk, downloadMaxChunk, 0
			for runCtx.Err() == nil {
				url := urlFor(chunk)
				t0 := time.Now()
				err := downloadOnce(runCtx, client, url, buf, &total)
				if err == nil {
					// Fewer, larger requests on fast links: servers rate-limit
					// by request count.
					lastGood = chunk
					if time.Since(t0) < time.Second && chunk < maxChunk {
						chunk = min(chunk*2, maxChunk)
					}
					continue
				}
				var se *HTTPStatusError
				if errors.As(err, &se) && se.StatusCode/100 == 4 && se.StatusCode != http.StatusTooManyRequests && lastGood > 0 && chunk > lastGood {
					// The server refuses requests this large; stay at the
					// largest size that worked.
					chunk, maxChunk = lastGood, lastGood
					continue
				}
				if runCtx.Err() != nil {
					continue
				}
				if errs.add(err) >= maxStreamFails*opts.Streams || fatal(err) {
					cancel()
					return
				}
				select {
				case <-runCtx.Done():
				case <-time.After(retryDelay(err)):
				}
			}
		}()
	}

	smp := runSampler(runCtx, &total, opts)
	wg.Wait()
	return finishSpeed(ctx, res, smp, opts, errs.first, total.Load())
}

func downloadOnce(ctx context.Context, client *http.Client, url string, buf []byte, total *atomic.Int64) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Cache-Control", "no-store")
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode/100 != 2 {
		return &HTTPStatusError{StatusCode: resp.StatusCode, Status: resp.Status}
	}
	for {
		n, err := resp.Body.Read(buf)
		total.Add(int64(n))
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
	}
}

// runSampler samples total until ctx ends and returns the samples. It also
// drives the progress callback.
func runSampler(ctx context.Context, total *atomic.Int64, opts SpeedOptions) *sampler {
	s := &sampler{start: time.Now()}
	s.add(0, 0, false)
	t := time.NewTicker(sampleEvery)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			s.add(time.Since(s.start), total.Load(), true)
			return s
		case <-t.C:
			el := time.Since(s.start)
			s.add(el, total.Load(), false)
			if opts.OnProgress != nil {
				f := el.Seconds() / opts.Duration.Seconds()
				if f > 1 {
					f = 1
				}
				opts.OnProgress(f, s.recent(time.Second))
			}
		}
	}
}

func finishSpeed(ctx context.Context, res model.SpeedResult, smp *sampler, opts SpeedOptions, firstErr error, total int64) (model.SpeedResult, error) {
	res.Bytes = total
	if ctx.Err() != nil {
		return res, ctx.Err()
	}
	end := smp.times[len(smp.times)-1]
	if firstErr != nil {
		// After failures the counter may have stalled; don't let the idle
		// tail dilute the rate.
		for i := len(smp.totals) - 1; i > 0 && smp.totals[i] == smp.totals[i-1]; i-- {
			end = smp.times[i-1]
		}
	}
	wStart, wBytes := smp.at(opts.Warmup)
	if end-wStart < time.Second || wBytes >= total {
		// The run was cut short (e.g. by errors) before or soon after the
		// warm-up ended; use everything we have.
		wStart, wBytes = 0, 0
	}
	res.DurationSec = (end - wStart).Seconds()
	res.Mbps = BytesToMbps(total-wBytes, res.DurationSec)
	res.SamplesMbps = smp.samples
	if total == 0 || res.Mbps == 0 {
		if firstErr != nil {
			return res, firstErr
		}
		return res, ErrNoData
	}
	if firstErr != nil && end < opts.Duration*3/4 {
		// Stopped early because of repeated failures.
		return res, fmt.Errorf("test interrupted: %w", firstErr)
	}
	return res, nil
}

// --- upload ---

// payload is incompressible data reused for every upload body.
var payload = func() []byte {
	b := make([]byte, 1<<20)
	_, _ = rand.Read(b)
	return b
}()

// countingBody streams size bytes of payload and counts what was read.
type countingBody struct {
	remaining int64
	off       int
	counter   *atomic.Int64
}

func (c *countingBody) Read(p []byte) (int, error) {
	if c.remaining <= 0 {
		return 0, io.EOF
	}
	if int64(len(p)) > c.remaining {
		p = p[:c.remaining]
	}
	n := copy(p, payload[c.off:])
	c.off = (c.off + n) % len(payload)
	c.remaining -= int64(n)
	c.counter.Add(int64(n))
	return n, nil
}

type completed struct {
	start, end time.Duration
	bytes      int64
}

const (
	uploadMinChunk = 128 << 10
	uploadMaxChunk = 64 << 20
	uploadGrace    = 4 * time.Second
)

// Upload measures send throughput by POSTing generated data to uploadURL.
//
// Bytes handed to the socket are not yet delivered, so the result only
// counts requests the server acknowledged. Each request is credited for the
// part of it that overlaps the measurement window, and requests still in
// flight at the deadline get a short grace period to finish.
func Upload(ctx context.Context, client *http.Client, uploadURL string, opts SpeedOptions) (model.SpeedResult, error) {
	res := model.SpeedResult{Streams: opts.Streams}
	// sendCtx stops new requests at the deadline; reqCtx lets in-flight ones
	// finish within the grace period.
	sendCtx, stopSending := context.WithTimeout(ctx, opts.Duration)
	defer stopSending()
	reqCtx, cancelReqs := context.WithTimeout(ctx, opts.Duration+uploadGrace)
	defer cancelReqs()

	var sent atomic.Int64
	var errs streamErrors
	var mu sync.Mutex
	var done []completed
	var wg sync.WaitGroup
	start := time.Now()

	for i := 0; i < opts.Streams; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			chunk := int64(uploadMinChunk)
			for sendCtx.Err() == nil {
				t0 := time.Since(start)
				err := uploadOnce(reqCtx, client, uploadURL, chunk, &sent)
				t1 := time.Since(start)
				if err == nil {
					mu.Lock()
					done = append(done, completed{t0, t1, chunk})
					mu.Unlock()
					// Aim for roughly one second per request so the window
					// boundaries cut few requests.
					if d := t1 - t0; d < 500*time.Millisecond && chunk < uploadMaxChunk {
						chunk *= 2
					} else if d > 2*time.Second && chunk > uploadMinChunk {
						chunk /= 2
					}
					continue
				}
				if reqCtx.Err() != nil || sendCtx.Err() != nil {
					return
				}
				if errs.add(err) >= maxStreamFails*opts.Streams || fatal(err) {
					stopSending()
					return
				}
				select {
				case <-sendCtx.Done():
				case <-time.After(retryDelay(err)):
				}
			}
		}()
	}

	smp := runSampler(sendCtx, &sent, opts)
	wg.Wait()
	if ctx.Err() != nil {
		return res, ctx.Err()
	}

	end := smp.times[len(smp.times)-1]
	window := [2]time.Duration{opts.Warmup, end}
	if end-opts.Warmup < time.Second {
		window[0] = 0
	}
	var credited float64
	var acked int64
	for _, c := range done {
		acked += c.bytes
		lo, hi := max(c.start, window[0]), min(c.end, window[1])
		if hi > lo && c.end > c.start {
			credited += float64(c.bytes) * float64(hi-lo) / float64(c.end-c.start)
		}
	}
	res.Bytes = acked
	res.DurationSec = (window[1] - window[0]).Seconds()
	res.Mbps = BytesToMbps(int64(credited), res.DurationSec)
	res.SamplesMbps = smp.samples
	if acked == 0 || res.Mbps == 0 {
		if errs.first != nil {
			return res, errs.first
		}
		return res, ErrNoData
	}
	if errs.first != nil && end < opts.Duration*3/4 {
		return res, fmt.Errorf("test interrupted: %w", errs.first)
	}
	return res, nil
}

func uploadOnce(ctx context.Context, client *http.Client, url string, size int64, sent *atomic.Int64) error {
	body := &countingBody{remaining: size, counter: sent}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, body)
	if err != nil {
		return err
	}
	req.ContentLength = size
	req.Header.Set("Content-Type", "application/octet-stream")
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, 1<<20))
	if resp.StatusCode/100 != 2 {
		return &HTTPStatusError{StatusCode: resp.StatusCode, Status: resp.Status}
	}
	return nil
}

// SampleLatency pings ip every interval until ctx ends and returns the RTTs
// of the replies (lost probes are skipped). Used for latency under load.
func SampleLatency(ctx context.Context, p Pinger, ip net.IP, interval, timeout time.Duration) []float64 {
	var out []float64
	t := time.NewTicker(interval)
	defer t.Stop()
	for {
		rtt, err := p.Ping(ctx, ip, timeout)
		if ctx.Err() != nil {
			return out
		}
		if err == nil {
			out = append(out, float64(rtt)/float64(time.Millisecond))
		}
		select {
		case <-ctx.Done():
			return out
		case <-t.C:
		}
	}
}
