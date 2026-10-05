package measure

import (
	"math"
	"sort"

	"github.com/tehgeii/wifi-speed-tester/internal/model"
)

// Jitter returns the mean absolute difference between consecutive RTT
// samples (the "inter-packet delay variation" used by most speed tests).
// Fewer than two samples yields 0.
func Jitter(samplesMs []float64) float64 {
	if len(samplesMs) < 2 {
		return 0
	}
	var sum float64
	for i := 1; i < len(samplesMs); i++ {
		sum += math.Abs(samplesMs[i] - samplesMs[i-1])
	}
	return sum / float64(len(samplesMs)-1)
}

// PacketLoss returns the lost percentage of sent packets (0..100).
func PacketLoss(sent, received int) float64 {
	if sent <= 0 {
		return 0
	}
	if received > sent {
		received = sent
	}
	return float64(sent-received) * 100 / float64(sent)
}

// Median returns the median of xs without modifying it; 0 for empty input.
func Median(xs []float64) float64 {
	if len(xs) == 0 {
		return 0
	}
	s := append([]float64(nil), xs...)
	sort.Float64s(s)
	n := len(s)
	if n%2 == 1 {
		return s[n/2]
	}
	return (s[n/2-1] + s[n/2]) / 2
}

// Summarize fills min/avg/max/jitter/loss from the samples and counters.
// samples holds the RTT of every reply in send order.
func Summarize(label, target string, sent int, samples []float64) model.PingStats {
	st := model.PingStats{
		Label:     label,
		Target:    target,
		Sent:      sent,
		Received:  len(samples),
		SamplesMs: samples,
	}
	st.PacketLossPct = PacketLoss(sent, len(samples))
	if len(samples) == 0 {
		return st
	}
	st.MinMs, st.MaxMs = samples[0], samples[0]
	var sum float64
	for _, v := range samples {
		sum += v
		st.MinMs = math.Min(st.MinMs, v)
		st.MaxMs = math.Max(st.MaxMs, v)
	}
	st.AvgMs = sum / float64(len(samples))
	st.JitterMs = Jitter(samples)
	return st
}

// Unit is a throughput display unit.
type Unit string

const (
	Mbps Unit = "Mbps" // megabits per second (10^6 bits)
	MBps Unit = "MB/s" // megabytes per second (10^6 bytes)
)

// BytesToMbps converts a byte count over a duration in seconds to megabits
// per second.
func BytesToMbps(bytes int64, seconds float64) float64 {
	if seconds <= 0 {
		return 0
	}
	return float64(bytes) * 8 / seconds / 1e6
}

// Convert expresses a megabit-per-second value in the given unit.
// 80 Mbps == 10 MB/s.
func Convert(mbps float64, u Unit) float64 {
	if u == MBps {
		return mbps / 8
	}
	return mbps
}
