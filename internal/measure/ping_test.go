package measure_test

import (
	"context"
	"errors"
	"net"
	"testing"
	"time"

	"github.com/tehgeii/wifi-speed-tester/internal/measure"
	"github.com/tehgeii/wifi-speed-tester/internal/testutil"
)

func TestRunPingCountsLoss(t *testing.T) {
	p := &testutil.Pinger{RTT: time.Millisecond, LoseEvery: 4}
	var replies int
	st, err := measure.RunPing(context.Background(), p, "t", "1.1.1.1", net.IPv4(1, 1, 1, 1), measure.PingOptions{
		Count: 20, Interval: time.Millisecond, Timeout: 100 * time.Millisecond,
		OnReply: func(int, float64, bool) { replies++ },
	})
	if err != nil {
		t.Fatal(err)
	}
	if st.Sent != 20 || st.Received != 15 || st.PacketLossPct != 25 {
		t.Fatalf("sent/recv/loss = %d/%d/%v", st.Sent, st.Received, st.PacketLossPct)
	}
	if replies != 20 {
		t.Fatalf("OnReply called %d times", replies)
	}
}

func TestRunPingAllLostKeepsError(t *testing.T) {
	p := &testutil.Pinger{RTT: time.Millisecond, Lose: map[string]bool{"8.8.8.8": true}}
	st, _ := measure.RunPing(context.Background(), p, "t", "8.8.8.8", net.IPv4(8, 8, 8, 8), measure.PingOptions{Count: 3, Interval: time.Millisecond, Timeout: time.Second})
	if st.Received != 0 || st.PacketLossPct != 100 || st.Error == "" {
		t.Fatalf("unexpected stats %+v", st)
	}
}

func TestRunPingCancel(t *testing.T) {
	p := &testutil.Pinger{RTT: time.Millisecond}
	ctx, cancel := context.WithCancel(context.Background())
	time.AfterFunc(50*time.Millisecond, cancel)
	st, err := measure.RunPing(ctx, p, "t", "h", net.IPv4(1, 1, 1, 1), measure.PingOptions{Count: 1000, Interval: 10 * time.Millisecond, Timeout: time.Second})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("err = %v", err)
	}
	if st.Sent == 0 || st.Sent >= 1000 || st.Sent != st.Received {
		t.Fatalf("partial stats wrong: %+v", st)
	}
}

// A target that never answers must not make the run take Count*Timeout.
func TestRunPingDoesNotSerializeTimeouts(t *testing.T) {
	p := &slowLoss{}
	start := time.Now()
	st, err := measure.RunPing(context.Background(), p, "t", "h", net.IPv4(8, 8, 8, 8), measure.PingOptions{Count: 10, Interval: 10 * time.Millisecond, Timeout: 300 * time.Millisecond})
	if err != nil {
		t.Fatal(err)
	}
	if d := time.Since(start); d > 1500*time.Millisecond {
		t.Fatalf("run took %v; probes appear to be serialized", d)
	}
	if st.Sent != 10 || st.Received != 0 {
		t.Fatalf("stats %+v", st)
	}
}

type slowLoss struct{}

func (slowLoss) Ping(ctx context.Context, _ net.IP, timeout time.Duration) (time.Duration, error) {
	select {
	case <-ctx.Done():
		return 0, ctx.Err()
	case <-time.After(timeout):
	}
	return 0, measure.ErrTimeout
}
func (slowLoss) Close() error { return nil }
