// Package analysis turns raw measurements into a deterministic connection
// quality rating and plain-language notes. Thresholds come from
// config.Profile; see docs/QUALITY.md.
package analysis

import (
	"fmt"
	"math"
	"slices"
	"strings"

	"github.com/tehgeii/wifi-speed-tester/internal/config"
	"github.com/tehgeii/wifi-speed-tester/internal/model"
)

var order = []model.QualityLevel{model.QualityExcellent, model.QualityGood, model.QualityFair, model.QualityPoor, model.QualityUnstable}

func rank(l model.QualityLevel) int { return slices.Index(order, l) }

func lowerBetter(v float64, b config.Band) model.QualityLevel {
	switch {
	case v <= b.Excellent:
		return model.QualityExcellent
	case v <= b.Good:
		return model.QualityGood
	case v <= b.Fair:
		return model.QualityFair
	}
	return model.QualityPoor
}

func higherBetter(v float64, b config.Band) model.QualityLevel {
	switch {
	case v >= b.Excellent:
		return model.QualityExcellent
	case v >= b.Good:
		return model.QualityGood
	case v >= b.Fair:
		return model.QualityFair
	}
	return model.QualityPoor
}

// FormatMs renders a latency value.
func FormatMs(v float64) string {
	if v < 10 {
		return fmt.Sprintf("%.1f ms", v)
	}
	return fmt.Sprintf("%.0f ms", v)
}

// FormatPct renders a percentage.
func FormatPct(v float64) string {
	if v == math.Trunc(v) {
		return fmt.Sprintf("%.0f%%", v)
	}
	return fmt.Sprintf("%.1f%%", v)
}

// FormatSpeed renders megabits per second in the requested unit.
func FormatSpeed(mbps float64, unit string) string {
	if unit == "MB/s" {
		return fmt.Sprintf("%.2f MB/s", mbps/8)
	}
	if mbps >= 100 {
		return fmt.Sprintf("%.0f Mbps", mbps)
	}
	return fmt.Sprintf("%.1f Mbps", mbps)
}

// LoadedIncrease returns how much the median latency under load exceeded the
// idle ping (the larger of download and upload), and whether it was measured.
func LoadedIncrease(r *model.TestResult) (float64, bool) {
	if r.PingMs <= 0 {
		return 0, false
	}
	worst, ok := 0.0, false
	for _, s := range []*model.SpeedResult{r.Download, r.Upload} {
		if s != nil && s.LoadedSamples > 0 {
			worst = math.Max(worst, s.LoadedPingMs-r.PingMs)
			ok = true
		}
	}
	return math.Max(worst, 0), ok
}

func speedOK(s *model.SpeedResult) bool { return s != nil && s.Error == "" && s.Mbps > 0 }

// Evaluate grades r with profile p (named profileName). It never mutates r.
func Evaluate(r *model.TestResult, profileName string, p config.Profile) *model.Quality {
	q := &model.Quality{Profile: profileName}
	uses := func(m string) bool { return slices.Contains(p.Use, m) }
	havePing := r.PingMs > 0

	graded := 0
	add := func(metric, value string, g model.QualityLevel, counts bool) {
		q.Grades = append(q.Grades, model.MetricGrade{Metric: metric, Value: value, Grade: g})
		if counts {
			graded++
		}
		if counts && (q.Level == "" || rank(g) > rank(q.Level)) {
			q.Level = g
		}
	}

	if havePing {
		add("Ping", FormatMs(r.PingMs), lowerBetter(r.PingMs, p.PingMs), uses("ping"))
		add("Jitter", FormatMs(r.JitterMs), lowerBetter(r.JitterMs, p.JitterMs), uses("jitter"))
		add("Packet Loss", FormatPct(r.PacketLoss), lowerBetter(r.PacketLoss, p.PacketLossPct), uses("packetLoss"))
	} else {
		q.Notes = append(q.Notes, "Ping could not be measured (ICMP may be blocked on this network).")
	}
	if speedOK(r.Download) {
		add("Download", FormatSpeed(r.Download.Mbps, "Mbps"), higherBetter(r.Download.Mbps, p.DownloadMbps), uses("download"))
	} else if uses("download") {
		q.Notes = append(q.Notes, "Download was not measured; the rating uses the remaining results.")
	}
	if speedOK(r.Upload) {
		add("Upload", FormatSpeed(r.Upload.Mbps, "Mbps"), higherBetter(r.Upload.Mbps, p.UploadMbps), uses("upload"))
	} else if uses("upload") {
		q.Notes = append(q.Notes, "Upload was not measured; the rating uses the remaining results.")
	}
	if inc, ok := LoadedIncrease(r); ok {
		add("Latency Under Load", "+"+FormatMs(inc), lowerBetter(inc, p.LoadedLatencyIncMs), uses("loadedLatency"))
		if inc > p.LoadedLatencyIncMs.Fair || (inc > 30 && inc > 2*r.PingMs) {
			q.Notes = append(q.Notes, "Latency increases significantly under load. This can be an early sign of bufferbloat; a dedicated test is needed to confirm.")
		}
	}

	if havePing && (r.PacketLoss >= p.UnstableLossPct || r.JitterMs >= p.UnstableJitterMs) {
		q.Level = model.QualityUnstable
		q.Notes = append(q.Notes, "Packet loss or large latency variation makes this connection unstable.")
	}
	// Rating from a minority of the profile's metrics would mislead (e.g.
	// EXCELLENT from upload alone when ping is blocked and download failed).
	if q.Level != model.QualityUnstable && graded*2 < len(p.Use) {
		if graded > 0 {
			q.Notes = append(q.Notes, "Too few results were measured to rate this connection reliably.")
		}
		q.Level = ""
	}
	if q.Level == "" {
		q.Level = model.QualityUnknown
	}

	for _, ps := range r.Pings {
		if ps.PrimaryTarget && ps.Method == "tcp" {
			q.Notes = append(q.Notes, "This network blocks ICMP ping, so latency was measured with TCP connections to the test server. These values can read slightly higher than a normal ping.")
		}
	}
	q.Notes = append(q.Notes, locationNotes(r)...)
	q.Summary = summary(r, q)
	return q
}

// locationNotes compares gateway and internet pings to hint where a problem
// sits, plus Wi-Fi specific hints.
func locationNotes(r *model.TestResult) []string {
	var notes []string
	var gw, inet *model.PingStats
	for i := range r.Pings {
		p := &r.Pings[i]
		if p.IsGateway {
			gw = p
		} else if p.PrimaryTarget {
			inet = p
		}
	}
	if gw != nil && gw.Sent > 0 {
		switch {
		case gw.Received == 0:
			notes = append(notes, "The router did not answer ping. Many routers block ping, so this alone is not a fault.")
		case gw.PacketLossPct > 0 || gw.AvgMs > 20 || gw.JitterMs > 10:
			notes = append(notes, fmt.Sprintf("Delay or loss already appears between this device and the router (%s avg, %s loss). The local network or Wi-Fi is a likely cause.",
				FormatMs(gw.AvgMs), FormatPct(gw.PacketLossPct)))
		case inet != nil && inet.Received > 0 && (inet.AvgMs-gw.AvgMs > 80 || inet.PacketLossPct > 0):
			notes = append(notes, "The local network looks healthy; the extra latency or loss is added beyond the router (ISP or internet route).")
		}
	}
	if w := wifiOf(r); w != nil {
		if w.SignalPercent > 0 && w.SignalPercent < 40 {
			notes = append(notes, fmt.Sprintf("Wi-Fi signal is weak (%d%%). Moving closer to the router may help.", w.SignalPercent))
		}
		if link := math.Max(w.RxRateMbps, w.TxRateMbps); link > 0 {
			notes = append(notes, fmt.Sprintf("Wi-Fi link speed (%.0f Mbps) is the radio rate to the router, not your internet speed.", link))
		}
	}
	return notes
}

func wifiOf(r *model.TestResult) *model.WiFiInfo {
	if r.Network == nil {
		return nil
	}
	return r.Network.WiFi
}

func summary(r *model.TestResult, q *model.Quality) string {
	var parts []string
	if speedOK(r.Download) {
		parts = append(parts, "measured download "+FormatSpeed(r.Download.Mbps, "Mbps"))
	}
	if speedOK(r.Upload) {
		parts = append(parts, "upload "+FormatSpeed(r.Upload.Mbps, "Mbps"))
	}
	if r.PingMs > 0 {
		parts = append(parts, "ping "+FormatMs(r.PingMs), "jitter "+FormatMs(r.JitterMs), "packet loss "+FormatPct(r.PacketLoss))
	}
	label := "General"
	if q.Profile == "gaming" {
		label = "Gaming"
	}
	s := fmt.Sprintf("Connection quality: %s (%s profile).", q.Level, label)
	if len(parts) > 0 {
		p := strings.Join(parts, ", ")
		s += " " + strings.ToUpper(p[:1]) + p[1:] + "."
	}
	return s + " Results reflect this moment and this test server, not a guaranteed ISP speed."
}
