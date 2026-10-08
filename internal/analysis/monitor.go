package analysis

import (
	"math"

	"github.com/tehgeii/wifi-speed-tester/internal/i18n"
	"github.com/tehgeii/wifi-speed-tester/internal/model"
)

// FormatDuration renders seconds as "5 min" / "1 min 30 s" ("5 menit" ...).
func FormatDuration(sec float64, lang i18n.Lang) string {
	total := int(math.Round(sec))
	m, s := total/60, total%60
	switch {
	case m == 0:
		return i18n.T(lang, "dur.sec", s)
	case s == 0:
		return i18n.T(lang, "dur.min", m)
	}
	return i18n.T(lang, "dur.min", m) + " " + i18n.T(lang, "dur.sec", s)
}

// MonitorVerdict fills in the localized conclusion and notes of a summary.
func MonitorVerdict(s *model.MonitorSummary, lang i18n.Lang) {
	T := func(key string, args ...any) string { return i18n.T(lang, key, args...) }
	s.Notes = nil
	switch {
	case s.Sent > 0 && s.Received == 0:
		s.Verdict = T("mon.drops", 1, FormatDuration(s.DurationSec, lang))
	case s.Outages > 0:
		s.Verdict = T("mon.drops", s.Outages, FormatDuration(s.LongestOutage, lang))
	case s.LossPct >= 1:
		s.Verdict = T("mon.lossy", FormatPct(math.Round(s.LossPct*10)/10))
	case s.Spikes >= max(3, s.Sent/100):
		s.Verdict = T("mon.spiky", s.Spikes)
	default:
		s.Verdict = T("mon.stable", FormatDuration(s.DurationSec, lang))
	}
	problem := s.Outages > 0 || s.LossPct >= 1
	if problem && s.GatewaySent > 0 && s.GatewayLost < s.GatewaySent { // a router that never answers just blocks ping
		gwLoss := float64(s.GatewayLost) / float64(s.GatewaySent) * 100
		switch {
		case s.GatewayLost == 0:
			s.Notes = append(s.Notes, T("mon.isp"))
		case gwLoss >= s.LossPct/2:
			s.Notes = append(s.Notes, T("mon.local", FormatPct(math.Round(gwLoss*10)/10)))
		}
	}
	if s.Method == "tcp" {
		s.Notes = append(s.Notes, T("mon.tcp"))
	}
	if s.Cancelled {
		s.Notes = append(s.Notes, T("mon.cancelled", FormatDuration(s.DurationSec, lang)))
	}
}
