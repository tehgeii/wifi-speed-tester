// Package dnsbench compares how fast DNS servers answer, using plain UDP
// queries so the result does not depend on the operating system's cache.
package dnsbench

import (
	"context"
	"crypto/rand"
	"encoding/binary"
	"errors"
	"net"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"golang.org/x/net/dns/dnsmessage"

	"github.com/tehgeii/wifi-speed-tester/internal/config"
	"github.com/tehgeii/wifi-speed-tester/internal/i18n"
	"github.com/tehgeii/wifi-speed-tester/internal/measure"
	"github.com/tehgeii/wifi-speed-tester/internal/model"
)

// Domains are looked up on every server: popular sites, so answers come
// from the resolver's cache the way they do in everyday browsing.
var Domains = []string{
	"google.com", "youtube.com", "whatsapp.net", "instagram.com", "tiktok.com",
	"tokopedia.com", "shopee.co.id", "detik.com", "wikipedia.org", "netflix.com",
}

// Server is a DNS server to test.
type Server struct {
	Label  string
	Addr   string // IP address
	System bool   // configured on this computer
}

// giveUpAfter is how many lookups a server may fail before its first answer
// before the rest of its lookups are skipped.
const giveUpAfter = 3

// Options controls a run.
type Options struct {
	Rounds  int           // lookups per domain (default 2)
	Timeout time.Duration // per query (default 2s)
	Port    int           // default 53 (tests use another)
	// OnProgress receives finished/total query counts.
	OnProgress func(done, total int)
}

// Run queries every server and returns results fastest first; servers that
// answered nothing come last.
func Run(ctx context.Context, servers []Server, domains []string, o Options) []model.DNSResult {
	if o.Rounds <= 0 {
		o.Rounds = 2
	}
	if o.Timeout <= 0 {
		o.Timeout = 2 * time.Second
	}
	if o.Port == 0 {
		o.Port = 53
	}
	total := len(servers) * len(domains) * o.Rounds
	var mu sync.Mutex
	done := 0
	tick := func() {
		mu.Lock()
		done++
		if o.OnProgress != nil {
			o.OnProgress(done, total)
		}
		mu.Unlock()
	}
	out := make([]model.DNSResult, len(servers))
	var wg sync.WaitGroup
	for i, s := range servers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			res := model.DNSResult{Label: s.Label, Server: s.Addr, System: s.System}
			var times []float64
			var lastErr error
			for round := 0; round < o.Rounds; round++ {
				for _, d := range domains {
					// A server that never answered after a few tries is
					// unreachable; waiting out every timeout would only
					// make the test slow.
					if ctx.Err() != nil || (res.OK == 0 && res.Failed >= giveUpAfter) {
						tick()
						continue
					}
					rtt, err := Query(ctx, net.JoinHostPort(s.Addr, strconv.Itoa(o.Port)), d, o.Timeout)
					if err != nil {
						res.Failed++
						lastErr = err
					} else {
						res.OK++
						times = append(times, float64(rtt)/float64(time.Millisecond))
					}
					tick()
				}
			}
			if len(times) > 0 {
				res.MedianMs = measure.Median(times)
				sort.Float64s(times)
				res.MinMs = times[0]
			} else if lastErr != nil {
				res.Error = lastErr.Error()
			}
			out[i] = res
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

// ErrServerFailure means the server answered with an error code.
var ErrServerFailure = errors.New("DNS server returned an error")

// Query sends one A-record query for name to addr (host:port) over UDP and
// returns the round-trip time. A NXDOMAIN answer counts as a valid response.
func Query(ctx context.Context, addr, name string, timeout time.Duration) (time.Duration, error) {
	var idb [2]byte
	_, _ = rand.Read(idb[:])
	id := binary.BigEndian.Uint16(idb[:])
	qname, err := dnsmessage.NewName(strings.TrimSuffix(name, ".") + ".")
	if err != nil {
		return 0, err
	}
	msg := dnsmessage.Message{
		Header:    dnsmessage.Header{ID: id, RecursionDesired: true},
		Questions: []dnsmessage.Question{{Name: qname, Type: dnsmessage.TypeA, Class: dnsmessage.ClassINET}},
	}
	pkt, err := msg.Pack()
	if err != nil {
		return 0, err
	}
	d := net.Dialer{Timeout: timeout}
	conn, err := d.DialContext(ctx, "udp", addr)
	if err != nil {
		return 0, err
	}
	defer conn.Close()
	deadline := time.Now().Add(timeout)
	_ = conn.SetDeadline(deadline)
	stop := context.AfterFunc(ctx, func() { _ = conn.SetDeadline(time.Now()) })
	defer stop()

	start := time.Now()
	if _, err := conn.Write(pkt); err != nil {
		return 0, err
	}
	buf := make([]byte, 1500)
	for {
		n, err := conn.Read(buf)
		if err != nil {
			if ctx.Err() != nil {
				return 0, ctx.Err()
			}
			return 0, err
		}
		var p dnsmessage.Parser
		h, err := p.Start(buf[:n])
		if err != nil || h.ID != id || !h.Response {
			continue // stray or malformed packet
		}
		rtt := time.Since(start)
		switch h.RCode {
		case dnsmessage.RCodeSuccess, dnsmessage.RCodeNameError:
			return rtt, nil
		}
		return 0, ErrServerFailure
	}
}

// Servers builds the list to test: the computer's own DNS servers first
// (link-local ones skipped), then the public ones. A public server that is
// also the computer's DNS appears once, marked as the system DNS.
func Servers(system []string, public []config.PingTarget, lang i18n.Lang) []Server {
	var list []Server
	seen := map[string]int{}
	for _, d := range system {
		if ip := net.ParseIP(d); ip == nil || ip.IsLinkLocalUnicast() {
			continue
		}
		if _, dup := seen[d]; !dup {
			seen[d] = len(list)
			list = append(list, Server{Label: i18n.T(lang, "dns.system", d), Addr: d, System: true})
		}
	}
	for _, s := range public {
		if i, dup := seen[s.Host]; dup {
			list[i].Label = s.Label + " · " + list[i].Label
			continue
		}
		seen[s.Host] = len(list)
		list = append(list, Server{Label: s.Label, Addr: s.Host})
	}
	return list
}
