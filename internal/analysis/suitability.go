package analysis

import (
	"fmt"

	"github.com/tehgeii/wifi-speed-tester/internal/i18n"
	"github.com/tehgeii/wifi-speed-tester/internal/model"
)

// Activity status values.
const (
	ActOK      = "ok"      // works smoothly
	ActWarn    = "warn"    // works, with occasional stutters
	ActBad     = "bad"     // not enough
	ActPending = "pending" // its measurements are still running
	ActNA      = "na"      // a needed measurement failed
)

type metric int

const (
	mDown metric = iota
	mUp
	mPing
	mJitter
	mLoss
	mLoaded // latency increase under load; optional
)

// need is one requirement of an activity. Speeds must be at least ok (good)
// or warn (usable); latency, jitter and loss at most.
type need struct {
	m        metric
	ok, warn float64
}

type activity struct {
	key   string
	needs []need
}

// activities and their thresholds (documented in docs/QUALITY.md). A big
// download is judged by how long 50 GB takes: 56 Mbps is about two hours,
// 14 Mbps about eight.
var activities = []activity{
	{"browse", []need{{mDown, 5, 1.5}, {mPing, 100, 300}, {mLoss, 1, 5}}},
	{"hd", []need{{mDown, 8, 5}, {mLoss, 2, 5}}},
	{"uhd", []need{{mDown, 25, 15}, {mLoss, 2, 5}}},
	{"call", []need{{mDown, 4, 1.5}, {mUp, 3, 1}, {mPing, 100, 200}, {mJitter, 30, 50}, {mLoss, 1, 3}}},
	{"game", []need{{mPing, 60, 100}, {mJitter, 15, 30}, {mLoss, 0.5, 2}, {mLoaded, 30, 80}}},
	{"live", []need{{mUp, 6, 3}, {mLoss, 1, 3}}},
	{"bigdl", []need{{mDown, 56, 14}}},
}

// quickActivities are the ones Quick Ping can judge from latency alone.
var quickActivities = map[string]bool{"browse": true, "call": true, "game": true}

// bigDownloadBits is the 50 GB used for the download-time estimate.
const bigDownloadBits = 50e9 * 8

// Suitability says how well r suits everyday activities. While a test runs
// (final false), activities whose measurements are still to come are
// "pending", so the list can be shown as it fills in; an activity that is
// already clearly not enough is reported as soon as that is known. Quick Ping
// only judges latency, so activities that need a speed test are left out.
func Suitability(r *model.TestResult, final bool, lang i18n.Lang) []model.Activity {
	T := func(key string, args ...any) string { return i18n.T(lang, key, args...) }
	quick := r.Mode == "quick"
	var out []model.Activity
	for _, a := range activities {
		needs := a.needs
		if quick {
			if !quickActivities[a.key] {
				continue // mostly about bandwidth; nothing useful to say
			}
			needs = nil
			for _, n := range a.needs {
				if n.m != mDown && n.m != mUp && n.m != mLoaded {
					needs = append(needs, n)
				}
			}
		}
		act := model.Activity{Key: a.key, Name: T("act." + a.key)}
		worst, worstReason := ActOK, ""
		pending, missing := false, false
		for _, n := range needs {
			v, state := value(r, n.m, final)
			switch state {
			case stPending:
				pending = true
				continue
			case stMissing:
				if n.m != mLoaded {
					missing = true
				}
				continue
			}
			st := grade(n, v)
			if rankAct(st) > rankAct(worst) {
				worst, worstReason = st, needReason(n, v, T)
			}
		}
		switch {
		case worst == ActBad:
			act.Status, act.Reason = ActBad, worstReason
		case pending:
			act.Status = ActPending
		case missing:
			act.Status, act.Reason = ActNA, T("act.na")
		default:
			act.Status, act.Reason = worst, worstReason
		}
		if a.key == "bigdl" && act.Status != ActPending && act.Status != ActNA {
			if v, st := value(r, mDown, final); st == stHave && v > 0 {
				act.Reason = T("act.bigdlTime", FormatDuration(bigDownloadBits/(v*1e6), lang))
			}
		}
		out = append(out, act)
	}
	return out
}

type valState int

const (
	stHave valState = iota
	stPending
	stMissing
)

func value(r *model.TestResult, m metric, final bool) (float64, valState) {
	speed := func(s *model.SpeedResult) (float64, valState) {
		switch {
		case s != nil && s.Mbps > 0:
			return s.Mbps, stHave
		case s == nil && !final:
			return 0, stPending
		}
		return 0, stMissing
	}
	pingDone := r.PingMs > 0 || final || len(r.Pings) > 0
	switch m {
	case mDown:
		return speed(r.Download)
	case mUp:
		return speed(r.Upload)
	case mLoaded:
		if v, ok := LoadedIncrease(r); ok {
			return v, stHave
		}
		if !final && (r.Download == nil || r.Upload == nil) {
			return 0, stPending
		}
		return 0, stMissing
	}
	if r.PingMs <= 0 {
		if !pingDone {
			return 0, stPending
		}
		return 0, stMissing
	}
	switch m {
	case mPing:
		return r.PingMs, stHave
	case mJitter:
		return r.JitterMs, stHave
	}
	return r.PacketLoss, stHave
}

func grade(n need, v float64) string {
	higher := n.m == mDown || n.m == mUp
	switch {
	case higher && v >= n.ok, !higher && v <= n.ok:
		return ActOK
	case higher && v >= n.warn, !higher && v <= n.warn:
		return ActWarn
	}
	return ActBad
}

func rankAct(s string) int {
	switch s {
	case ActWarn:
		return 1
	case ActBad:
		return 2
	}
	return 0
}

// needReason explains the limiting value, e.g. "Ping 120 ms (ideal ≤ 60 ms)".
func needReason(n need, v float64, T func(string, ...any) string) string {
	switch n.m {
	case mDown, mUp:
		name := T("metric.download")
		if n.m == mUp {
			name = T("metric.upload")
		}
		return T("act.need", name, FormatSpeed(v, "Mbps"), "≥ "+FormatSpeed(n.ok, "Mbps"))
	case mPing:
		return T("act.need", T("metric.ping"), FormatMs(v)+" ms", fmt.Sprintf("≤ %s ms", FormatMs(n.ok)))
	case mJitter:
		return T("act.need", T("metric.jitter"), FormatMs(v)+" ms", fmt.Sprintf("≤ %s ms", FormatMs(n.ok)))
	case mLoaded:
		return T("act.need", T("metric.loadedLatency"), "+"+FormatMs(v)+" ms", fmt.Sprintf("≤ +%s ms", FormatMs(n.ok)))
	}
	return T("act.need", T("metric.packetLoss"), FormatPct(v)+"%", fmt.Sprintf("≤ %s%%", FormatPct(n.ok)))
}

// BloatGrade grades how much latency rises under load, on the scale common
// bufferbloat tests use: A+ under 5 ms, A under 30, B under 60, C under 200,
// D under 400, otherwise F.
func BloatGrade(increaseMs float64) string {
	switch {
	case increaseMs < 5:
		return "A+"
	case increaseMs < 30:
		return "A"
	case increaseMs < 60:
		return "B"
	case increaseMs < 200:
		return "C"
	case increaseMs < 400:
		return "D"
	}
	return "F"
}

// GradeOf returns r's bufferbloat grade, or "" when latency under load was
// not measured.
func GradeOf(r *model.TestResult) string {
	if v, ok := LoadedIncrease(r); ok {
		return BloatGrade(v)
	}
	return ""
}
