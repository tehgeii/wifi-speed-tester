package dnsbench

import (
	"context"
	"strconv"
	"testing"
	"time"

	"github.com/tehgeii/wifi-speed-tester/internal/config"
	"github.com/tehgeii/wifi-speed-tester/internal/i18n"
	"github.com/tehgeii/wifi-speed-tester/internal/testutil"
)

// Each fake server gets its own port, so the loopback address is shared and
// the port selects the server; run them one at a time through Run.
func runOne(t *testing.T, srv *testutil.DNSServer, label string, system bool) []Server {
	t.Helper()
	return []Server{{Label: label, Addr: "127.0.0.1", System: system}}
}

func TestQueryAndRun(t *testing.T) {
	fast := testutil.NewDNSServer(2 * time.Millisecond)
	defer fast.Close()
	rtt, err := Query(context.Background(), "127.0.0.1:"+strconv.Itoa(fast.Port()), "example.com", time.Second)
	if err != nil || rtt <= 0 || rtt > 500*time.Millisecond {
		t.Fatalf("rtt=%v err=%v", rtt, err)
	}

	calls := 0
	res := Run(context.Background(), runOne(t, fast, "Fast", true), []string{"a.com", "b.com"}, Options{Port: fast.Port(), OnProgress: func(d, n int) { calls++ }})
	if len(res) != 1 || res[0].OK != 4 || res[0].Failed != 0 || res[0].MedianMs <= 0 {
		t.Fatalf("res = %+v", res)
	}
	if calls != 4 {
		t.Fatalf("progress calls = %d", calls)
	}
}

func TestFailuresAndTimeouts(t *testing.T) {
	srv := testutil.NewDNSServer(0)
	defer srv.Close()
	srv.Fail.Store(true)
	if _, err := Query(context.Background(), "127.0.0.1:"+strconv.Itoa(srv.Port()), "example.com", time.Second); err != ErrServerFailure {
		t.Fatalf("err = %v, want ErrServerFailure", err)
	}
	srv.Fail.Store(false)
	srv.Silent.Store(true)
	start := time.Now()
	res := Run(context.Background(), runOne(t, srv, "Silent", false), []string{"a.com"}, Options{Port: srv.Port(), Rounds: 1, Timeout: 200 * time.Millisecond})
	if res[0].OK != 0 || res[0].Failed != 1 || res[0].Error == "" {
		t.Fatalf("res = %+v", res)
	}
	if time.Since(start) > time.Second {
		t.Fatal("timeout not honoured")
	}
}

func TestUnreachableServerGivesUpEarly(t *testing.T) {
	srv := testutil.NewDNSServer(0)
	defer srv.Close()
	srv.Silent.Store(true)
	calls := 0
	start := time.Now()
	res := Run(context.Background(), runOne(t, srv, "Silent", false), Domains, Options{Port: srv.Port(), Timeout: 100 * time.Millisecond, OnProgress: func(d, n int) { calls++ }})
	if res[0].Failed != giveUpAfter || res[0].OK != 0 {
		t.Fatalf("res = %+v, want %d failures then skip", res, giveUpAfter)
	}
	if calls != 2*len(Domains) {
		t.Fatalf("progress calls = %d, want %d so the bar completes", calls, 2*len(Domains))
	}
	if time.Since(start) > time.Second {
		t.Fatalf("took %v", time.Since(start))
	}
}

func TestCancel(t *testing.T) {
	srv := testutil.NewDNSServer(0)
	defer srv.Close()
	srv.Silent.Store(true)
	ctx, cancel := context.WithCancel(context.Background())
	time.AfterFunc(100*time.Millisecond, cancel)
	start := time.Now()
	Run(ctx, runOne(t, srv, "x", false), Domains, Options{Port: srv.Port(), Timeout: 2 * time.Second})
	if time.Since(start) > time.Second {
		t.Fatal("cancel not prompt")
	}
}

func TestServers(t *testing.T) {
	list := Servers([]string{"192.168.1.1", "8.8.8.8", "fe80::1", "192.168.1.1"}, []config.PingTarget{{Label: "Google", Host: "8.8.8.8"}, {Label: "Quad9", Host: "9.9.9.9"}}, i18n.EN)
	if len(list) != 3 {
		t.Fatalf("list = %+v", list)
	}
	if list[0].Label != "Your DNS (192.168.1.1)" || !list[0].System || list[1].Label != "Google · Your DNS (8.8.8.8)" || !list[1].System || list[2].System {
		t.Fatalf("list = %+v", list)
	}
}
