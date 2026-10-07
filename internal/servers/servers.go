// Package servers finds speed-test servers: it fetches the public LibreSpeed
// server list and ranks entries by measured latency. Nothing here runs unless
// the user asks for nearby servers.
package servers

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/tehgeii/wifi-speed-tester/internal/config"
)

// listEntry is one item of the LibreSpeed servers.php JSON.
type listEntry struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Server      string `json:"server"`
	DLURL       string `json:"dlURL"`
	ULURL       string `json:"ulURL"`
	PingURL     string `json:"pingURL"`
	SponsorName string `json:"sponsorName"`
}

// Parse converts a LibreSpeed server list into config servers. Entries with
// a missing or non-http(s) address are skipped. Protocol-relative addresses
// ("//host/path") get https.
func Parse(data []byte) ([]config.Server, error) {
	var entries []listEntry
	if err := json.Unmarshal(data, &entries); err != nil {
		return nil, fmt.Errorf("server list is not valid JSON: %w", err)
	}
	var out []config.Server
	for _, e := range entries {
		base := strings.TrimSpace(e.Server)
		if strings.HasPrefix(base, "//") {
			base = "https:" + base
		}
		s := config.Server{
			Name: strings.TrimSpace(e.Name), Type: config.LibreSpeed, URL: base,
			DLPath: e.DLURL, ULPath: e.ULURL, PingPath: e.PingURL, Sponsor: e.SponsorName,
		}
		if s.Name == "" || s.Validate() != nil {
			continue
		}
		out = append(out, s)
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("server list contains no usable servers")
	}
	return out, nil
}

// Fetch downloads and parses the server list at listURL.
func Fetch(ctx context.Context, client *http.Client, listURL string) ([]config.Server, error) {
	ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, listURL, nil)
	if err != nil {
		return nil, err
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode/100 != 2 {
		return nil, fmt.Errorf("server list: %s", resp.Status)
	}
	data, err := io.ReadAll(io.LimitReader(resp.Body, 2<<20))
	if err != nil {
		return nil, err
	}
	return Parse(data)
}

// Candidate is a server with its measured latency.
type Candidate struct {
	Server    config.Server `json:"server"`
	LatencyMs float64       `json:"latencyMs"` // meaningful only when Error is ""
	Error     string        `json:"error,omitempty"`
}

// Probe measures each server's HTTP latency (best of a few small requests
// on a kept-alive connection, so TLS setup is not counted) and returns them
// fastest first, unreachable ones last. onDone, if set, is called after each
// server with the number finished so far.
func Probe(ctx context.Context, list []config.Server, concurrency int, onDone func(done, total int)) []Candidate {
	out := make([]Candidate, len(list))
	sem := make(chan struct{}, max(concurrency, 1))
	var wg sync.WaitGroup
	var mu sync.Mutex
	done := 0
	for i, s := range list {
		wg.Add(1)
		go func() {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()
			ms, err := latency(ctx, s)
			out[i] = Candidate{Server: s, LatencyMs: ms}
			if err != nil {
				out[i].Error = err.Error()
			}
			if onDone != nil {
				mu.Lock()
				done++
				onDone(done, len(list))
				mu.Unlock()
			}
		}()
	}
	wg.Wait()
	sort.SliceStable(out, func(a, b int) bool {
		ea, eb := out[a].Error != "", out[b].Error != ""
		if ea != eb {
			return eb
		}
		return out[a].LatencyMs < out[b].LatencyMs
	})
	return out
}

func latency(ctx context.Context, s config.Server) (float64, error) {
	client := &http.Client{
		Timeout:   3 * time.Second,
		Transport: &http.Transport{Proxy: http.ProxyFromEnvironment, MaxIdleConnsPerHost: 1},
	}
	defer client.CloseIdleConnections()
	best := -1.0
	var lastErr error
	for i := 0; i < 3; i++ {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, s.Ping(), nil)
		if err != nil {
			return 0, err
		}
		req.Header.Set("Cache-Control", "no-store")
		start := time.Now()
		resp, err := client.Do(req)
		if err != nil {
			lastErr = err
			continue
		}
		_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, 64<<10))
		resp.Body.Close()
		if resp.StatusCode/100 != 2 {
			lastErr = fmt.Errorf("HTTP %s", resp.Status)
			continue
		}
		ms := float64(time.Since(start)) / float64(time.Millisecond)
		if i == 0 {
			continue // includes connection and TLS setup
		}
		if best < 0 || ms < best {
			best = ms
		}
	}
	if best < 0 {
		if lastErr == nil {
			lastErr = fmt.Errorf("no response")
		}
		return 0, lastErr
	}
	return best, nil
}
