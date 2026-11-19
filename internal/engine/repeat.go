package engine

import (
	"context"

	"github.com/tehgeii/wifi-speed-tester/internal/analysis"
	"github.com/tehgeii/wifi-speed-tester/internal/measure"
	"github.com/tehgeii/wifi-speed-tester/internal/model"
)

// RunRepeated runs n full tests back to back and reports the median of each
// figure, which is steadier than one run (useful as evidence for an ISP).
// Every run's events are passed on, each run starting with a "run" event,
// and progress counts the data of all runs so far. Only the combined result
// is sent as a "result" event. A run that is cancelled or fails ends the
// series and is reported as it is. Quick Ping is never repeated.
func (e *Engine) RunRepeated(ctx context.Context, opts Options, n int, emit func(Event)) *model.TestResult {
	if n <= 1 || opts.Mode == "quick" {
		return e.Run(ctx, opts, emit)
	}
	if emit == nil {
		emit = func(Event) {}
	}
	var runs []*model.TestResult
	var used int64
	for i := 1; i <= n; i++ {
		emit(Event{Type: "run", Run: i, Runs: n})
		base := used
		res := e.Run(ctx, opts, func(ev Event) {
			if ev.Type == "result" {
				return
			}
			if ev.DataBytes > 0 {
				ev.DataBytes += base
			}
			emit(ev)
		})
		used += res.DataBytes
		runs = append(runs, res)
		if res.Cancelled || res.Failure != nil || res.Quality == nil {
			res.DataBytes = used
			emit(Event{Type: "result", Stage: StageDone, Progress: 1, Result: res})
			return res
		}
	}
	out := e.combine(runs, opts)
	out.DataBytes = used
	emit(Event{Type: "result", Stage: StageDone, Progress: 1, Result: out})
	return out
}

// combine builds one result from complete runs: the last run supplies the
// details (network, checks, ping table, throughput chart), and the headline
// figures are medians over all runs.
func (e *Engine) combine(runs []*model.TestResult, opts Options) *model.TestResult {
	last := *runs[len(runs)-1]
	out := &last
	out.StartedAt = runs[0].StartedAt
	var ping, jitter, loss []float64
	down, downLoaded := speeds(runs, func(r *model.TestResult) *model.SpeedResult { return r.Download })
	up, upLoaded := speeds(runs, func(r *model.TestResult) *model.SpeedResult { return r.Upload })
	out.Runs = nil
	for _, r := range runs {
		s := model.RunSummary{PingMs: r.PingMs, JitterMs: r.JitterMs, LossPct: r.PacketLoss}
		if r.Download != nil {
			s.DownMbps = r.Download.Mbps
		}
		if r.Upload != nil {
			s.UpMbps = r.Upload.Mbps
		}
		out.Runs = append(out.Runs, s)
		if r.PingMs > 0 {
			ping = append(ping, r.PingMs)
			jitter = append(jitter, r.JitterMs)
			loss = append(loss, r.PacketLoss)
		}
	}
	if len(ping) > 0 {
		out.PingMs, out.JitterMs, out.PacketLoss = measure.Median(ping), measure.Median(jitter), measure.Median(loss)
	}
	out.Download = medianSpeed(last.Download, down, downLoaded)
	out.Upload = medianSpeed(last.Upload, up, upLoaded)
	mode := last.Mode
	out.Quality = analysis.Evaluate(out, mode, e.Config.Profile(mode), opts.Lang)
	out.Activities = analysis.Suitability(out, true, opts.Lang)
	out.BloatGrade = analysis.GradeOf(out)
	return out
}

func speeds(runs []*model.TestResult, get func(*model.TestResult) *model.SpeedResult) (mbps, loaded []float64) {
	for _, r := range runs {
		if s := get(r); s != nil && s.Mbps > 0 {
			mbps = append(mbps, s.Mbps)
			if s.LoadedSamples > 0 {
				loaded = append(loaded, s.LoadedPingMs)
			}
		}
	}
	return mbps, loaded
}

func medianSpeed(last *model.SpeedResult, mbps, loaded []float64) *model.SpeedResult {
	if last == nil {
		return nil
	}
	s := *last
	if len(mbps) > 0 {
		s.Mbps = measure.Median(mbps)
		s.Error = "" // at least one run measured it
	}
	if len(loaded) > 0 {
		s.LoadedPingMs = measure.Median(loaded)
		s.LoadedSamples = max(s.LoadedSamples, 1)
	}
	return &s
}
