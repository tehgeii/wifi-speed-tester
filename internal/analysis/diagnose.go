package analysis

import (
	"strings"

	"github.com/tehgeii/wifi-speed-tester/internal/i18n"
	"github.com/tehgeii/wifi-speed-tester/internal/model"
)

// Diagnose concludes the guided diagnosis. The steps are the test where the
// user is now, optionally near the router, and optionally on a LAN cable;
// comparing them separates weak Wi-Fi coverage, Wi-Fi interference, a
// struggling router and a problem beyond the router (ISP or route).
func Diagnose(steps []model.DiagStep, lang i18n.Lang) model.Diagnosis {
	T := func(key string, args ...any) string { return i18n.T(lang, key, args...) }
	get := func(kind string) *model.TestResult {
		for _, s := range steps {
			if s.Kind == kind && s.Result != nil {
				return s.Result
			}
		}
		return nil
	}
	now, near, cable := get("now"), get("near"), get("cable")
	var d model.Diagnosis
	for _, s := range steps {
		if s.Result != nil {
			d.Facts = append(d.Facts, diagFact(s.Kind, s.Result, T))
		}
	}
	if cable != nil && (cable.Network == nil || cable.Network.Connection != model.ConnEthernet) {
		d.Facts = append(d.Facts, T("diag.cableIgnored"))
		cable = nil
	}
	set := func(kind string, actions ...string) {
		d.Kind, d.Title, d.Detail, d.Actions = kind, T("diag."+kind), T("diag."+kind+".detail"), actions
	}
	wifi := now != nil && now.Network != nil && now.Network.Connection == model.ConnWiFi
	switch {
	case now == nil:
		set("unclear")
	case offline(now):
		if checkFailed(now, "gateway") || checkFailed(now, "adapter") {
			set("offlineLocal", "router")
		} else {
			set("offline", "router", "ispReport")
		}
	case cable != nil:
		switch {
		case healthy(cable) && !healthy(now):
			set("wifi", "channel")
		case !healthy(cable) && locate(cable).gatewayIssue:
			set("router", "router")
		case !healthy(cable):
			set("isp", "ispReport", "monitor")
		default:
			set("healthy", "monitor")
		}
	case near != nil:
		switch {
		case healthy(near) && !healthy(now):
			set("coverage", "channel")
		case !healthy(near) && (locate(near).gatewayIssue || locate(now).gatewayIssue):
			set("router", "channel", "router")
		case !healthy(near):
			set("isp", "ispReport", "monitor")
		default:
			set("healthy", "monitor")
		}
	case healthy(now):
		set("healthy", "monitor")
	case viaPhone(now):
		// Walking to a router or plugging a cable does not apply: the limit
		// is the phone's mobile signal or the link to the phone.
		set("hotspot", "monitor")
		if loc := locate(now); loc.gatewayIssue {
			d.Facts = append(d.Facts, T("diag.hotspotLink", FormatMs(loc.gw.AvgMs)))
		}
	case locate(now).gatewayIssue:
		set("router", "channel", "router")
	case wifi && signalOf(now) >= StrongSignal:
		// Moving closer cannot improve a strong signal, and the router
		// answers normally, so the slowness is most likely past it.
		set("isp", "ispReport", "monitor")
		d.Facts = append(d.Facts, T("diag.strongSignal", signalOf(now)))
	case wifi:
		set("unclear", "monitor")
	default:
		set("isp", "ispReport", "monitor")
	}
	if now != nil && now.DNSMs > 150 {
		d.Facts = append(d.Facts, T("diag.dnsSlow", FormatMs(now.DNSMs)))
		d.Actions = append(d.Actions, "dns")
	}
	return d
}

// StrongSignal is the Wi-Fi signal (%) from which a test near the router is
// skipped: it would not change the result.
const StrongSignal = 70

func signalOf(r *model.TestResult) int {
	if w := wifiOf(r); w != nil {
		return w.SignalPercent
	}
	return 0
}

func offline(r *model.TestResult) bool { return r.Failure != nil && r.Quality == nil }

func checkFailed(r *model.TestResult, key string) bool {
	for _, c := range r.Checks {
		if c.Key == key && c.Status == model.CheckFail {
			return true
		}
	}
	return false
}

// healthy means rated good or better, without delay or loss at the router
// and not far below the user's plan.
func healthy(r *model.TestResult) bool {
	if r == nil || r.Quality == nil {
		return false
	}
	if l := r.Quality.Level; l != model.QualityExcellent && l != model.QualityGood {
		return false
	}
	if locate(r).gatewayIssue {
		return false
	}
	if p := r.Plan; p != nil && p.DownMbps > 0 && p.LowPct > 0 && r.Download != nil && r.Download.Mbps > 0 &&
		model.PlanPercent(r.Download.Mbps, p.DownMbps) < p.LowPct {
		return false
	}
	return true
}

func diagFact(kind string, r *model.TestResult, T func(string, ...any) string) string {
	sp := func(s *model.SpeedResult) string {
		if s == nil || s.Mbps <= 0 {
			return "—"
		}
		return FormatSpeed(s.Mbps, "Mbps")
	}
	ping := "—"
	if r.PingMs > 0 {
		ping = FormatMs(r.PingMs)
	}
	var extra []string
	if w := wifiOf(r); w != nil && w.SignalPercent > 0 {
		extra = append(extra, T("diag.signal", w.SignalPercent))
	}
	if r.Network != nil && r.Network.Connection == model.ConnEthernet {
		extra = append(extra, T("diag.onCable"))
	}
	s := T("diag.fact", T("diag.step."+kind), sp(r.Download), sp(r.Upload), ping)
	if len(extra) > 0 {
		s += ", " + strings.Join(extra, ", ")
	}
	if r.Quality != nil {
		s += " · " + T("level."+string(r.Quality.Level))
	}
	return s
}

// RegionAdvice explains region latencies. Most game servers for Southeast
// Asia run in Singapore, so that region gets its own verdict.
func RegionAdvice(res []model.RegionResult, lang i18n.Lang) []string {
	T := func(key string, args ...any) string { return i18n.T(lang, key, args...) }
	var best *model.RegionResult
	var sg *model.RegionResult
	for i := range res {
		r := &res[i]
		if r.OK == 0 {
			continue
		}
		if best == nil || r.MedianMs < best.MedianMs {
			best = r
		}
		if r.Label == "Singapore" || strings.Contains(r.Host, "ap-southeast-1.") {
			sg = r
		}
	}
	if best == nil {
		return []string{T("region.allFailed")}
	}
	out := []string{T("region.best", best.Label, FormatMs(best.MedianMs))}
	if sg != nil {
		key := "region.sgHigh"
		switch {
		case sg.MedianMs <= 40:
			key = "region.sgGood"
		case sg.MedianMs <= 80:
			key = "region.sgOk"
		}
		out = append(out, T(key, FormatMs(sg.MedianMs)))
	}
	return out
}
