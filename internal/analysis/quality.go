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
	"github.com/tehgeii/wifi-speed-tester/internal/i18n"
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

// Evaluate grades r with profile p (named profileName) and writes notes,
// tips and the summary in lang. It never mutates r.
func Evaluate(r *model.TestResult, profileName string, p config.Profile, lang i18n.Lang) *model.Quality {
	T := func(key string, args ...any) string { return i18n.T(lang, key, args...) }
	q := &model.Quality{Profile: profileName, Lang: string(lang)}
	uses := func(m string) bool { return slices.Contains(p.Use, m) }
	havePing := r.PingMs > 0

	graded := 0
	add := func(key, value string, g model.QualityLevel, b config.Band, unit string, higher bool) {
		counts := uses(key)
		q.Grades = append(q.Grades, model.MetricGrade{
			Key: key, Metric: T("metric." + key), Value: value, Grade: g, Counted: counts,
			Limits: [3]float64{b.Excellent, b.Good, b.Fair}, Unit: unit, HigherBetter: higher,
		})
		if counts {
			graded++
		}
		if counts && (q.Level == "" || rank(g) > rank(q.Level)) {
			q.Level = g
		}
	}

	if havePing {
		add("ping", FormatMs(r.PingMs), lowerBetter(r.PingMs, p.PingMs), p.PingMs, "ms", false)
		add("jitter", FormatMs(r.JitterMs), lowerBetter(r.JitterMs, p.JitterMs), p.JitterMs, "ms", false)
		add("packetLoss", FormatPct(r.PacketLoss), lowerBetter(r.PacketLoss, p.PacketLossPct), p.PacketLossPct, "%", false)
	} else {
		q.Notes = append(q.Notes, T("note.pingMissing"))
	}
	if speedOK(r.Download) {
		add("download", FormatSpeed(r.Download.Mbps, "Mbps"), higherBetter(r.Download.Mbps, p.DownloadMbps), p.DownloadMbps, "Mbps", true)
	} else if uses("download") {
		q.Notes = append(q.Notes, T("note.downloadMissing"))
	}
	if speedOK(r.Upload) {
		add("upload", FormatSpeed(r.Upload.Mbps, "Mbps"), higherBetter(r.Upload.Mbps, p.UploadMbps), p.UploadMbps, "Mbps", true)
	} else if uses("upload") {
		q.Notes = append(q.Notes, T("note.uploadMissing"))
	}
	inc, haveLoaded := LoadedIncrease(r)
	bloat := false
	if haveLoaded {
		add("loadedLatency", "+"+FormatMs(inc), lowerBetter(inc, p.LoadedLatencyIncMs), p.LoadedLatencyIncMs, "ms", false)
		if inc > p.LoadedLatencyIncMs.Fair || (inc > 30 && inc > 2*r.PingMs) {
			bloat = true
			q.Notes = append(q.Notes, T("note.bufferbloat"))
		}
	}

	if havePing && (r.PacketLoss >= p.UnstableLossPct || r.JitterMs >= p.UnstableJitterMs) {
		q.Level = model.QualityUnstable
		q.Notes = append(q.Notes, T("note.unstable"))
	}
	// Rating from a minority of the profile's metrics would mislead (e.g.
	// EXCELLENT from upload alone when ping is blocked and download failed).
	if q.Level != model.QualityUnstable && graded*2 < len(p.Use) {
		if graded > 0 {
			q.Notes = append(q.Notes, T("note.tooFew"))
		}
		q.Level = ""
	}
	if q.Level == "" {
		q.Level = model.QualityUnknown
	}

	for _, ps := range r.Pings {
		if ps.PrimaryTarget && ps.Method == "tcp" {
			q.Notes = append(q.Notes, T("note.tcpPing"))
		}
	}
	loc := locate(r)
	q.Notes = append(q.Notes, locationNotes(r, loc, T)...)
	q.Tips = tips(r, q, p, loc, bloat || (haveLoaded && inc > p.LoadedLatencyIncMs.Good), T)
	q.Summary = summary(r, q, T)
	return q
}

// where tells on which side of the router latency or loss appears.
type where struct {
	gw, inet      *model.PingStats
	gatewayIssue  bool // delay/loss already between device and router
	beyondRouter  bool // router fine, problem added after it
	routerNoReply bool
}

func locate(r *model.TestResult) where {
	var w where
	for i := range r.Pings {
		p := &r.Pings[i]
		if p.IsGateway {
			w.gw = p
		} else if p.PrimaryTarget {
			w.inet = p
		}
	}
	if gw := w.gw; gw != nil && gw.Sent > 0 {
		switch {
		case gw.Received == 0:
			w.routerNoReply = true
		case gw.PacketLossPct > 0 || gw.AvgMs > 20 || gw.JitterMs > 10:
			w.gatewayIssue = true
		case w.inet != nil && w.inet.Received > 0 && (w.inet.AvgMs-gw.AvgMs > 80 || w.inet.PacketLossPct > 0):
			w.beyondRouter = true
		}
	}
	return w
}

// locationNotes describes where a problem sits, plus Wi-Fi observations.
func locationNotes(r *model.TestResult, loc where, T func(string, ...any) string) []string {
	var notes []string
	switch {
	case loc.routerNoReply:
		notes = append(notes, T("note.routerNoPing"))
	case loc.gatewayIssue:
		notes = append(notes, T("note.localProblem", FormatMs(loc.gw.AvgMs), FormatPct(loc.gw.PacketLossPct)))
	case loc.beyondRouter:
		notes = append(notes, T("note.beyondRouter"))
	}
	if w := wifiOf(r); w != nil {
		if w.SignalPercent > 0 && w.SignalPercent < 40 {
			notes = append(notes, T("note.weakSignal", w.SignalPercent))
		}
		if link := math.Max(w.RxRateMbps, w.TxRateMbps); link > 0 {
			notes = append(notes, T("note.linkSpeed", link))
		}
	}
	return notes
}

func gradeOf(q *model.Quality, key string) model.QualityLevel {
	for _, g := range q.Grades {
		if g.Key == key {
			return g.Grade
		}
	}
	return ""
}

func weak(g model.QualityLevel) bool { return g == model.QualityFair || g == model.QualityPoor }

// tips turns the findings into concrete actions, most useful first. Rules are
// deterministic and only fire on a measured problem.
func tips(r *model.TestResult, q *model.Quality, p config.Profile, loc where, loadedBad bool, T func(string, ...any) string) []string {
	var out []string
	add := func(key string) {
		if t := T(key); !slices.Contains(out, t) {
			out = append(out, t)
		}
	}
	isWiFi := r.Network != nil && r.Network.Connection == model.ConnWiFi
	w := wifiOf(r)
	jitterBad := weak(gradeOf(q, "jitter")) || r.JitterMs >= p.UnstableJitterMs
	lossBad := r.PingMs > 0 && r.PacketLoss > 0

	if loadedBad {
		add("tip.sqm")
	}
	if isWiFi && w != nil && w.SignalPercent > 0 && w.SignalPercent < 60 {
		add("tip.weakSignal")
	}
	if isWiFi && w != nil && w.Band == "2.4 GHz" {
		add("tip.use5GHz")
	}
	if isWiFi && (loc.gatewayIssue || jitterBad || lossBad) {
		add("tip.tryCable")
	}
	if loc.gatewayIssue {
		add("tip.restartRouter")
	}
	if r.Network != nil && r.Network.Connection == model.ConnVPN {
		add("tip.vpn")
	}
	if weak(gradeOf(q, "ping")) && !loc.gatewayIssue {
		add("tip.nearerServer")
	}
	if weak(gradeOf(q, "download")) || weak(gradeOf(q, "upload")) {
		add("tip.otherTraffic")
	}
	if loc.beyondRouter || (lossBad && !loc.gatewayIssue && !isWiFi) {
		add("tip.isp")
	}
	return out
}

func wifiOf(r *model.TestResult) *model.WiFiInfo {
	if r.Network == nil {
		return nil
	}
	return r.Network.WiFi
}

// ProfileLabel is the localized profile name.
func ProfileLabel(profile string, lang i18n.Lang) string {
	switch profile {
	case "gaming", "quick":
		return i18n.T(lang, "profile."+profile)
	}
	return i18n.T(lang, "profile.general")
}

func summary(r *model.TestResult, q *model.Quality, T func(string, ...any) string) string {
	var parts []string
	if speedOK(r.Download) {
		parts = append(parts, T("summary.download", FormatSpeed(r.Download.Mbps, "Mbps")))
	}
	if speedOK(r.Upload) {
		parts = append(parts, T("summary.upload", FormatSpeed(r.Upload.Mbps, "Mbps")))
	}
	if r.PingMs > 0 {
		parts = append(parts, T("summary.ping", FormatMs(r.PingMs)), T("summary.jitter", FormatMs(r.JitterMs)), T("summary.loss", FormatPct(r.PacketLoss)))
	}
	s := T("summary.head", T("level."+string(q.Level)), ProfileLabel(q.Profile, i18n.Lang(q.Lang)))
	if len(parts) > 0 {
		p := strings.Join(parts, ", ")
		s += " " + strings.ToUpper(p[:1]) + p[1:] + "."
	}
	return s + " " + T("summary.disclaimer")
}
