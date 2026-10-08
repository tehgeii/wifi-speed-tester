package export

import (
	"bytes"
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	"github.com/tehgeii/wifi-speed-tester/internal/analysis"
	"github.com/tehgeii/wifi-speed-tester/internal/i18n"
	"github.com/tehgeii/wifi-speed-tester/internal/model"
)

// MonitorTXT renders a stability monitor run.
func MonitorTXT(s *model.MonitorSummary, o Options) []byte {
	T := func(key string, args ...any) string { return i18n.T(o.Lang, key, args...) }
	var b bytes.Buffer
	line := func(k, v string) { fmt.Fprintf(&b, "%-22s %s\n", k+":", v) }
	b.WriteString(T("rep.monTitle") + "\n" + strings.Repeat("=", 46) + "\n\n")
	line(T("rep.date"), dateText(s.StartedAt, o.Lang))
	line(T("rep.duration"), analysis.FormatDuration(s.DurationSec, o.Lang))
	target := s.Target
	if !o.IncludeSensitive {
		target = maskIfPrivate(target)
	}
	line(T("rep.target"), strings.TrimSpace(s.TargetLabel+" ("+target+")"))
	if n := s.Network; n != nil {
		line(T("rep.connection"), string(n.Connection))
	}
	b.WriteString("\n")
	line(T("rep.probes"), T("rep.probesVal", s.Sent, s.Received, analysis.FormatPct(math.Round(s.LossPct*10)/10)))
	if s.Received > 0 {
		line(T("rep.latency"), T("rep.latencyVal", analysis.FormatMs(s.AvgMs), analysis.FormatMs(s.MinMs), analysis.FormatMs(s.MaxMs), analysis.FormatMs(s.P95Ms), analysis.FormatMs(s.JitterMs)))
	}
	line(T("rep.spikes"), fmt.Sprint(s.Spikes))
	line(T("rep.outages"), T("rep.outagesVal", s.Outages, analysis.FormatDuration(s.LongestOutage, o.Lang)))
	if s.GatewaySent > 0 {
		line(T("rep.router"), fmt.Sprintf("%d / %d", s.GatewayLost, s.GatewaySent))
	}
	b.WriteString("\n" + T("rep.verdict") + ": " + s.Verdict + "\n")
	for _, n := range s.Notes {
		b.WriteString("  - " + n + "\n")
	}
	b.WriteString("\n" + T("rep.footer") + "\n")
	return b.Bytes()
}

// maskIfPrivate hides private (LAN) addresses; public hosts stay visible.
func maskIfPrivate(h string) string {
	if strings.Count(h, ".") == 3 && (strings.HasPrefix(h, "10.") || strings.HasPrefix(h, "192.168.") || strings.HasPrefix(h, "172.")) {
		return MaskIP(h)
	}
	return h
}

// ReportFilter selects the tests that go into an ISP report.
type ReportFilter struct {
	Since      time.Time // zero = all
	Connection string    // "" = all
}

// ISPReport summarizes many speed tests into one document a user can send
// to their ISP: averages, % of plan, a time-of-day breakdown, the slowest
// tests and the latest stability monitor runs in the period. Quick Ping
// results are left out of the speed figures.
func ISPReport(all []*model.TestResult, monitors []*model.MonitorSummary, plan *model.Plan, f ReportFilter, o Options) []byte {
	T := func(key string, args ...any) string { return i18n.T(o.Lang, key, args...) }
	var tests []*model.TestResult
	for _, r := range all {
		if r.Mode == "quick" || r.Cancelled || !r.StartedAt.After(f.Since) {
			continue
		}
		if f.Connection != "" && (r.Network == nil || string(r.Network.Connection) != f.Connection) {
			continue
		}
		if (r.Download == nil || r.Download.Mbps <= 0) && r.PingMs <= 0 {
			continue
		}
		tests = append(tests, r)
	}
	sort.Slice(tests, func(a, b int) bool { return tests[a].StartedAt.Before(tests[b].StartedAt) })

	var b bytes.Buffer
	line := func(k, v string) { fmt.Fprintf(&b, "%-26s %s\n", k+":", v) }
	b.WriteString(T("isp.title") + "\n" + strings.Repeat("=", 52) + "\n" + T("isp.intro") + "\n\n")
	// Stability monitor runs in the period are evidence too, even without
	// speed tests.
	writeMonitors := func() {
		var mons []*model.MonitorSummary
		for _, m := range monitors {
			if m.StartedAt.After(f.Since) && (f.Connection == "" || (m.Network != nil && string(m.Network.Connection) == f.Connection)) {
				mons = append(mons, m)
			}
		}
		if len(mons) == 0 {
			return
		}
		b.WriteString("\n" + T("rep.monTitle") + "\n")
		for _, m := range mons[:min(10, len(mons))] {
			fmt.Fprintf(&b, "  %s  %s: %s\n", m.StartedAt.Format("2006-01-02 15:04"), analysis.FormatDuration(m.DurationSec, o.Lang), m.Verdict)
		}
	}
	if len(tests) == 0 {
		b.WriteString(T("isp.none") + "\n")
		writeMonitors()
		b.WriteString("\n" + T("rep.footer") + "\n")
		return b.Bytes()
	}
	line(T("isp.period"), dateText(tests[0].StartedAt, o.Lang)+" – "+dateText(tests[len(tests)-1].StartedAt, o.Lang))
	line(T("isp.tests"), fmt.Sprint(len(tests)))
	conn := f.Connection
	if conn == "" {
		conn = connectionsOf(tests)
	}
	line(T("isp.connection"), conn)
	if plan != nil && (plan.DownMbps > 0 || plan.UpMbps > 0) {
		line(T("isp.plan"), T("isp.planVal", speedOrDash(plan.DownMbps, o.Unit), speedOrDash(plan.UpMbps, o.Unit)))
	} else {
		line(T("isp.plan"), T("isp.planNone"))
	}

	speedOf := func(s *model.SpeedResult) float64 {
		if v := mbps(s); v > 0 {
			return v
		}
		return -1
	}
	down := collect(tests, func(r *model.TestResult) float64 { return speedOf(r.Download) })
	up := collect(tests, func(r *model.TestResult) float64 { return speedOf(r.Upload) })
	ping := collect(tests, func(r *model.TestResult) float64 {
		if r.PingMs > 0 {
			return r.PingMs
		}
		return -1
	})
	jit := collect(tests, func(r *model.TestResult) float64 {
		if r.PingMs > 0 {
			return r.JitterMs
		}
		return -1
	})
	loss := collect(tests, func(r *model.TestResult) float64 {
		if r.PingMs > 0 {
			return r.PacketLoss
		}
		return -1
	})
	b.WriteString("\n" + T("isp.summary") + "\n")
	spd := func(v float64) string { return analysis.FormatSpeed(v, o.Unit) }
	if len(down) > 0 {
		line("Download", T("isp.avgMinMax", spd(avg(down)), spd(minOf(down)), spd(maxOf(down))))
	}
	if len(up) > 0 {
		line("Upload", T("isp.avgMinMax", spd(avg(up)), spd(minOf(up)), spd(maxOf(up))))
	}
	if len(ping) > 0 {
		line("Ping", T("isp.avgMinMax", analysis.FormatMs(avg(ping)), analysis.FormatMs(minOf(ping)), analysis.FormatMs(maxOf(ping))))
	}
	if len(jit) > 0 {
		line("Jitter", T("isp.avgMinMax", analysis.FormatMs(avg(jit)), analysis.FormatMs(minOf(jit)), analysis.FormatMs(maxOf(jit))))
	}
	if len(loss) > 0 {
		line("Packet Loss", T("isp.avgMinMax", pctText(avg(loss)), pctText(minOf(loss)), pctText(maxOf(loss))))
	}
	if plan != nil && plan.DownMbps > 0 && len(down) > 0 {
		upPct := "—"
		if plan.UpMbps > 0 && len(up) > 0 {
			upPct = pctText(model.PlanPercent(avg(up), plan.UpMbps))
		}
		line(T("isp.ofPlan"), T("isp.ofPlanVal", pctText(model.PlanPercent(avg(down), plan.DownMbps)), upPct))
		low := plan.LowPct
		if low <= 0 {
			low = 50
		}
		below := 0
		for _, v := range down {
			if model.PlanPercent(v, plan.DownMbps) < low {
				below++
			}
		}
		line(T("isp.belowPlan", low), T("isp.belowVal", below, len(down)))
	}

	// Time of day: four 6-hour slots.
	type slot struct{ down, ping []float64 }
	var slots [4]slot
	for _, r := range tests {
		i := r.StartedAt.Hour() / 6
		if v := mbps(r.Download); v > 0 {
			slots[i].down = append(slots[i].down, v)
		}
		if r.PingMs > 0 {
			slots[i].ping = append(slots[i].ping, r.PingMs)
		}
	}
	b.WriteString("\n" + T("isp.byTime") + "\n")
	fast, slow := -1, -1
	for i, s := range slots {
		n := max(len(s.down), len(s.ping))
		if n == 0 {
			continue
		}
		d, p := "—", "—"
		if len(s.down) > 0 {
			d = spd(avg(s.down))
			if fast < 0 || avg(s.down) > avg(slots[fast].down) {
				fast = i
			}
			if slow < 0 || avg(s.down) < avg(slots[slow].down) {
				slow = i
			}
		}
		if len(s.ping) > 0 {
			p = analysis.FormatMs(avg(s.ping))
		}
		b.WriteString("  " + T("isp.timeRow", T(fmt.Sprintf("isp.slot%d", i)), n, d, p) + "\n")
	}
	if fast >= 0 && slow >= 0 && fast != slow {
		if drop := (1 - avg(slots[slow].down)/avg(slots[fast].down)) * 100; drop >= 25 {
			b.WriteString("  " + T("isp.peak", T(fmt.Sprintf("isp.slot%d", slow)), drop, T(fmt.Sprintf("isp.slot%d", fast))) + "\n")
		}
	}

	row := func(r *model.TestResult) string {
		c := ""
		if r.Network != nil {
			c = string(r.Network.Connection)
		}
		p, l := "—", "—"
		if r.PingMs > 0 {
			p, l = analysis.FormatMs(r.PingMs), pctText(r.PacketLoss)
		}
		return "  " + T("isp.row", r.StartedAt.Format("2006-01-02 15:04"), c, speedOrDash(mbps(r.Download), o.Unit), speedOrDash(mbps(r.Upload), o.Unit), p, l)
	}
	withSpeed := make([]*model.TestResult, 0, len(tests))
	for _, r := range tests {
		if mbps(r.Download) > 0 {
			withSpeed = append(withSpeed, r)
		}
	}
	sort.SliceStable(withSpeed, func(a, c int) bool { return mbps(withSpeed[a].Download) < mbps(withSpeed[c].Download) })
	if len(withSpeed) > 0 {
		b.WriteString("\n" + T("isp.worst") + "\n")
		for _, r := range withSpeed[:min(5, len(withSpeed))] {
			b.WriteString(row(r) + "\n")
		}
	}

	writeMonitors()

	b.WriteString("\n" + T("isp.all") + "\n")
	for _, r := range tests {
		b.WriteString(row(r) + "\n")
	}
	if wifi := countConn(tests, model.ConnWiFi); wifi > 0 {
		b.WriteString("\n" + T("isp.wifiNote", wifi) + "\n")
	}
	b.WriteString("\n" + T("rep.footer") + "\n")
	return b.Bytes()
}

func mbps(s *model.SpeedResult) float64 {
	if s == nil || s.Mbps <= 0 {
		return 0
	}
	return s.Mbps
}

// collect gathers f over rs. f returns a negative value for "not
// measured"; a speed of 0 also means not measured.
func collect(rs []*model.TestResult, f func(*model.TestResult) float64) []float64 {
	var out []float64
	for _, r := range rs {
		if v := f(r); v >= 0 {
			out = append(out, v)
		}
	}
	return out
}

func avg(xs []float64) float64 {
	var s float64
	for _, x := range xs {
		s += x
	}
	return s / float64(len(xs))
}

func minOf(xs []float64) float64 {
	m := xs[0]
	for _, x := range xs {
		m = math.Min(m, x)
	}
	return m
}

func maxOf(xs []float64) float64 {
	m := xs[0]
	for _, x := range xs {
		m = math.Max(m, x)
	}
	return m
}

func pctText(v float64) string { return analysis.FormatPct(math.Round(v*10) / 10) }

func speedOrDash(v float64, unit string) string {
	if v <= 0 {
		return "—"
	}
	return analysis.FormatSpeed(v, unit)
}

func connectionsOf(rs []*model.TestResult) string {
	seen := map[string]bool{}
	var out []string
	for _, r := range rs {
		if r.Network != nil && !seen[string(r.Network.Connection)] {
			seen[string(r.Network.Connection)] = true
			out = append(out, string(r.Network.Connection))
		}
	}
	return strings.Join(out, ", ")
}

func countConn(rs []*model.TestResult, c model.ConnectionType) int {
	n := 0
	for _, r := range rs {
		if r.Network != nil && r.Network.Connection == c {
			n++
		}
	}
	return n
}
