package regions

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/tehgeii/wifi-speed-tester/internal/config"
	"github.com/tehgeii/wifi-speed-tester/internal/model"
)

func listener(t *testing.T) (string, func()) {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	go func() {
		for {
			c, err := ln.Accept()
			if err != nil {
				return
			}
			c.Close()
		}
	}()
	return ln.Addr().String(), func() { ln.Close() }
}

func TestRunOrdersAndReportsEach(t *testing.T) {
	ok, stop := listener(t)
	defer stop()
	// A closed port: connections are refused.
	dead, stopDead := listener(t)
	stopDead()
	var got []string
	res := Run(context.Background(), []config.PingTarget{
		{Label: "Dead", Host: dead},
		{Label: "Local", Host: ok},
	}, Options{Probes: 3, Interval: 10 * time.Millisecond, Timeout: time.Second, OnResult: func(r model.RegionResult) { got = append(got, r.Label) }})
	if len(got) != 2 {
		t.Fatalf("OnResult calls = %v", got)
	}
	if res[0].Label != "Local" || res[0].OK != 3 || res[0].Sent != 3 || res[0].MedianMs <= 0 {
		t.Fatalf("first = %+v", res[0])
	}
	if res[1].Label != "Dead" || res[1].OK != 0 || res[1].Error == "" {
		t.Fatalf("second = %+v", res[1])
	}
}

func TestCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	start := time.Now()
	res := Run(ctx, []config.PingTarget{{Label: "X", Host: "127.0.0.1"}}, Options{Probes: 50})
	if time.Since(start) > time.Second || res[0].Sent > 1 {
		t.Fatalf("cancel not honoured: %+v after %v", res[0], time.Since(start))
	}
}
