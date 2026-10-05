// Package testutil provides fakes shared by tests.
package testutil

import (
	"context"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"strconv"
	"sync"
	"time"

	"github.com/tehgeii/wifi-speed-tester/internal/measure"
)

// Pinger answers with a fixed RTT; IPs listed in Lose never answer and
// every Nth probe (LoseEvery) is dropped.
type Pinger struct {
	RTT       time.Duration
	Lose      map[string]bool
	LoseEvery int
	mu        sync.Mutex
	n         int
}

func (p *Pinger) Ping(ctx context.Context, ip net.IP, timeout time.Duration) (time.Duration, error) {
	p.mu.Lock()
	p.n++
	n := p.n
	p.mu.Unlock()
	if p.Lose[ip.String()] || (p.LoseEvery > 0 && n%p.LoseEvery == 0) {
		select {
		case <-ctx.Done():
			return 0, ctx.Err()
		case <-time.After(5 * time.Millisecond):
		}
		return 0, measure.ErrTimeout
	}
	select {
	case <-ctx.Done():
		return 0, ctx.Err()
	case <-time.After(p.RTT):
	}
	return p.RTT, nil
}

func (p *Pinger) Close() error { return nil }

// SpeedServer is a local stand-in for a speed-test server, with an optional
// bandwidth limit so results are predictable.
type SpeedServer struct {
	*httptest.Server
	BytesPerSec int64 // 0 = unlimited
	Status      int   // non-zero forces this status on every request
}

// NewSpeedServer starts a server with /down?bytes=N and /up endpoints.
func NewSpeedServer(bytesPerSec int64) *SpeedServer {
	s := &SpeedServer{BytesPerSec: bytesPerSec}
	mux := http.NewServeMux()
	mux.HandleFunc("/down", func(w http.ResponseWriter, r *http.Request) {
		if s.Status != 0 {
			w.WriteHeader(s.Status)
			return
		}
		n, _ := strconv.Atoi(r.URL.Query().Get("bytes"))
		w.Header().Set("Content-Length", strconv.Itoa(n))
		buf := make([]byte, 32<<10)
		for n > 0 && r.Context().Err() == nil {
			c := min(n, len(buf))
			if _, err := w.Write(buf[:c]); err != nil {
				return
			}
			n -= c
			s.throttle(c)
		}
	})
	mux.HandleFunc("/up", func(w http.ResponseWriter, r *http.Request) {
		if s.Status != 0 {
			io.Copy(io.Discard, r.Body)
			w.WriteHeader(s.Status)
			return
		}
		buf := make([]byte, 32<<10)
		for {
			n, err := r.Body.Read(buf)
			s.throttle(n)
			if err != nil {
				break
			}
		}
		w.WriteHeader(http.StatusOK)
	})
	s.Server = httptest.NewServer(mux)
	return s
}

func (s *SpeedServer) throttle(n int) {
	if s.BytesPerSec > 0 && n > 0 {
		time.Sleep(time.Duration(int64(n) * int64(time.Second) / s.BytesPerSec))
	}
}

// DownloadURL and UploadURL are the config templates for this server.
func (s *SpeedServer) DownloadURL() string { return s.URL + "/down?bytes={bytes}" }
func (s *SpeedServer) UploadURL() string   { return s.URL + "/up" }
