package app

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/tehgeii/wifi-speed-tester/internal/analysis"
	"github.com/tehgeii/wifi-speed-tester/internal/i18n"
	"github.com/tehgeii/wifi-speed-tester/internal/model"
	"github.com/tehgeii/wifi-speed-tester/internal/network"
	"github.com/tehgeii/wifi-speed-tester/internal/regions"
)

// ---- Wi-Fi signal finder ------------------------------------------------------

// signalMax bounds a signal-finder session; the page restarts it if needed.
const signalMax = 30 * time.Minute

// startSignal reads the Wi-Fi signal every interval and sends "signal"
// events until stopSignal, signalMax, or the window closes. It sends no
// network traffic. Every event carries the page's token, so readings from
// a session that was just stopped are never mixed into a new one; starting
// again replaces a running session.
func (a *App) startSignal(token int) {
	a.mu.Lock()
	if a.sigCancel != nil {
		a.sigCancel()
	}
	ctx, cancel := context.WithTimeout(context.Background(), signalMax)
	a.sigCancel = cancel
	a.mu.Unlock()
	go func() {
		defer func() {
			cancel()
			a.emit("signal", map[string]any{"token": token, "stopped": true})
		}()
		start := time.Now()
		t := time.NewTicker(a.signalEvery)
		defer t.Stop()
		for {
			w, err := network.WiFiSignal(ctx)
			if ctx.Err() != nil {
				return
			}
			ev := map[string]any{"token": token, "t": time.Since(start).Seconds()}
			if err != nil {
				ev["error"] = err.Error()
			} else {
				ev["wifi"] = w
			}
			a.emit("signal", ev)
			select {
			case <-ctx.Done():
				return
			case <-t.C:
			}
		}
	}()
}

func (a *App) stopSignal() {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.sigCancel != nil {
		a.sigCancel()
		a.sigCancel = nil
	}
}

// Spot is a place where the signal was saved ("Living room").
type Spot struct {
	Name          string    `json:"name"`
	SignalPercent int       `json:"signalPercent"`
	RSSI          int       `json:"rssi"`
	At            time.Time `json:"at"`
}

// MaxSpots limits the saved signal spots.
const MaxSpots = 20

func (a *App) spots() []Spot {
	var s []Spot
	a.readJSON("spots.json", &s)
	return s
}

// saveSpot stores a spot, replacing one with the same name.
func (a *App) saveSpot(s Spot) ([]Spot, error) {
	s.Name = strings.TrimSpace(s.Name)
	if r := []rune(s.Name); len(r) > 30 {
		s.Name = string(r[:30])
	}
	if s.Name == "" || s.SignalPercent < 0 || s.SignalPercent > 100 {
		return nil, errors.New(i18n.T(a.lang(), "app.badSpot"))
	}
	s.At = time.Now()
	list := []Spot{s}
	for _, x := range a.spots() {
		if !strings.EqualFold(x.Name, s.Name) && len(list) < MaxSpots {
			list = append(list, x)
		}
	}
	return list, a.writeJSON("spots.json", list)
}

// ---- ping to regions ------------------------------------------------------------

// regionTest measures every configured region and sends a "region" event
// as each one finishes, then the sorted list with advice.
func (a *App) regionTest() {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	a.emit("region", map[string]any{"total": len(a.cfg.Regions)})
	res := regions.Run(ctx, a.cfg.Regions, regions.Options{
		OnResult: func(r model.RegionResult) { a.emit("region", map[string]any{"result": r}) },
	})
	a.emit("region", map[string]any{"results": res, "advice": analysis.RegionAdvice(res, a.lang())})
}

// ---- guided diagnosis -------------------------------------------------------------

func (a *App) diagnose(steps []model.DiagStep) model.Diagnosis {
	return analysis.Diagnose(steps, a.lang())
}
