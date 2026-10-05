package measure

import (
	"math"
	"testing"
)

func near(a, b float64) bool { return math.Abs(a-b) < 1e-9 }

func TestJitter(t *testing.T) {
	// Example from the product brief: diffs 1,2,4,3,2 -> 2.4 ms.
	if j := Jitter([]float64{18, 19, 17, 21, 18, 20}); !near(j, 2.4) {
		t.Fatalf("jitter = %v, want 2.4", j)
	}
	if Jitter([]float64{5}) != 0 || Jitter(nil) != 0 {
		t.Fatal("jitter of <2 samples must be 0")
	}
}

func TestPacketLoss(t *testing.T) {
	cases := []struct {
		sent, recv int
		want       float64
	}{{100, 100, 0}, {100, 96, 4}, {20, 0, 100}, {0, 0, 0}, {10, 12, 0}}
	for _, c := range cases {
		if got := PacketLoss(c.sent, c.recv); !near(got, c.want) {
			t.Errorf("PacketLoss(%d,%d) = %v, want %v", c.sent, c.recv, got, c.want)
		}
	}
}

func TestUnits(t *testing.T) {
	if got := Convert(80, MBps); !near(got, 10) {
		t.Fatalf("80 Mbps = %v MB/s, want 10", got)
	}
	if got := Convert(80, Mbps); got != 80 {
		t.Fatalf("Mbps conversion changed the value: %v", got)
	}
	// 10 MB in 1 s = 80 Mbps.
	if got := BytesToMbps(10_000_000, 1); !near(got, 80) {
		t.Fatalf("BytesToMbps = %v, want 80", got)
	}
	if BytesToMbps(100, 0) != 0 {
		t.Fatal("zero duration must not divide by zero")
	}
}

func TestSummarize(t *testing.T) {
	st := Summarize("x", "1.1.1.1", 4, []float64{15, 18, 27})
	if st.Received != 3 || !near(st.PacketLossPct, 25) {
		t.Fatalf("received/loss = %d/%v", st.Received, st.PacketLossPct)
	}
	if st.MinMs != 15 || st.MaxMs != 27 || !near(st.AvgMs, 20) {
		t.Fatalf("min/avg/max = %v/%v/%v", st.MinMs, st.AvgMs, st.MaxMs)
	}
	if !near(st.JitterMs, 6) {
		t.Fatalf("jitter = %v, want 6", st.JitterMs)
	}
	empty := Summarize("x", "h", 3, nil)
	if empty.PacketLossPct != 100 || empty.AvgMs != 0 {
		t.Fatalf("all lost: %+v", empty)
	}
}

func TestMedian(t *testing.T) {
	if Median([]float64{3, 1, 2}) != 2 || Median([]float64{4, 1, 3, 2}) != 2.5 || Median(nil) != 0 {
		t.Fatal("median wrong")
	}
}
