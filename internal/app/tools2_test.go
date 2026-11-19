package app

import (
	"encoding/json"
	"net"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/tehgeii/wifi-speed-tester/internal/config"
	"github.com/tehgeii/wifi-speed-tester/internal/model"
)

// all returns every payload of the events called name, as JSON maps.
func (e *events) all(name string) []map[string]any {
	e.mu.Lock()
	defer e.mu.Unlock()
	var out []map[string]any
	for _, ev := range e.list {
		if ev.name != name {
			continue
		}
		b, _ := json.Marshal(ev.payload)
		var m map[string]any
		_ = json.Unmarshal(b, &m)
		out = append(out, m)
	}
	return out
}

// until polls cond until it holds or the timeout passes.
func until(t *testing.T, what string, timeout time.Duration, cond func() bool) {
	t.Helper()
	for end := time.Now().Add(timeout); time.Now().Before(end); time.Sleep(10 * time.Millisecond) {
		if cond() {
			return
		}
	}
	t.Fatalf("timed out waiting for %s", what)
}

func signalEvents(ev *events, token float64, key string) int {
	n := 0
	for _, m := range ev.all("signal") {
		if _, ok := m[key]; ok && m["token"] == token {
			n++
		}
	}
	return n
}

func TestSignalFinder(t *testing.T) {
	a, ev, _ := newTestApp(t)
	a.signalEvery = 20 * time.Millisecond
	f := filepath.Join(t.TempDir(), "signal.json")
	if err := os.WriteFile(f, []byte(`[{"ssid":"Home","signalPercent":40,"rssi":-70},{"ssid":"Home","signalPercent":80,"rssi":-50}]`), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("WST_FAKE_WIFI_SIGNAL", f)

	if _, err := call(t, a, "signalStart", 7); err != nil {
		t.Fatal(err)
	}
	// Windows reads the real adapter, which CI runners lack: each tick then
	// reports an error instead of a reading, still once per interval.
	until(t, "readings", 5*time.Second, func() bool { return signalEvents(ev, 7, "t") >= 3 })
	if runtime.GOOS != "windows" {
		for _, m := range ev.all("signal") {
			if w, ok := m["wifi"].(map[string]any); ok && w["ssid"] != "Home" {
				t.Fatalf("reading = %v", m)
			}
		}
	}
	// Starting again replaces the session; the old one reports its end.
	if _, err := call(t, a, "signalStart", 8); err != nil {
		t.Fatal(err)
	}
	until(t, "old session stopped", 5*time.Second, func() bool { return signalEvents(ev, 7, "stopped") == 1 })
	until(t, "new readings", 5*time.Second, func() bool { return signalEvents(ev, 8, "t") >= 2 })
	if _, err := call(t, a, "signalStop"); err != nil {
		t.Fatal(err)
	}
	until(t, "new session stopped", 5*time.Second, func() bool { return signalEvents(ev, 8, "stopped") == 1 })
	n := signalEvents(ev, 8, "t")
	time.Sleep(100 * time.Millisecond)
	if signalEvents(ev, 8, "t") != n {
		t.Fatal("readings continued after stop")
	}
}

func TestSpots(t *testing.T) {
	a, _, _ := newTestApp(t)
	if _, err := call(t, a, "saveSpot", Spot{Name: "  ", SignalPercent: 50}); err == nil {
		t.Fatal("a spot without a name was saved")
	}
	if _, err := call(t, a, "saveSpot", Spot{Name: "Kamar", SignalPercent: 101}); err == nil {
		t.Fatal("a signal above 100% was saved")
	}
	_, _ = call(t, a, "saveSpot", Spot{Name: "Kamar", SignalPercent: 60, RSSI: -65})
	_, _ = call(t, a, "saveSpot", Spot{Name: "Dapur", SignalPercent: 40})
	got, err := call(t, a, "saveSpot", Spot{Name: "kamar ", SignalPercent: 70})
	if err != nil {
		t.Fatal(err)
	}
	list := got.([]Spot)
	if len(list) != 2 || list[0].Name != "kamar" || list[0].SignalPercent != 70 || list[1].Name != "Dapur" {
		t.Fatalf("spots = %+v", list)
	}
	long := strings.Repeat("é", 40)
	for i := 0; i < MaxSpots+5; i++ {
		_, _ = call(t, a, "saveSpot", Spot{Name: long[:2*(i%20+1)] + string(rune('a'+i)), SignalPercent: i})
	}
	stored, _ := call(t, a, "getSpots")
	if s := stored.([]Spot); len(s) != MaxSpots {
		t.Fatalf("kept %d spots, want %d", len(s), MaxSpots)
	}
	got, _ = call(t, a, "saveSpot", Spot{Name: long, SignalPercent: 1})
	if n := len([]rune(got.([]Spot)[0].Name)); n != 30 {
		t.Fatalf("name kept %d runes, want 30", n)
	}
	if _, err := call(t, a, "clearSpots"); err != nil {
		t.Fatal(err)
	}
	if s, _ := call(t, a, "getSpots"); len(s.([]Spot)) != 0 {
		t.Fatalf("after clear: %+v", s)
	}
}

func TestRegionTest(t *testing.T) {
	a, ev, _ := newTestApp(t)
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()
	go func() {
		for {
			c, err := ln.Accept()
			if err != nil {
				return
			}
			c.Close()
		}
	}()
	dead, _ := net.Listen("tcp", "127.0.0.1:0")
	deadAddr := dead.Addr().String()
	dead.Close()
	a.cfg.Regions = []config.PingTarget{{Label: "Closed", Host: deadAddr}, {Label: "Singapore", Host: ln.Addr().String()}}

	if _, err := call(t, a, "regionTest"); err != nil {
		t.Fatal(err)
	}
	m := ev.wait(t, "region", "results", 20*time.Second)
	res := m["results"].([]any)
	if len(res) != 2 || res[0].(map[string]any)["label"] != "Singapore" || res[1].(map[string]any)["label"] != "Closed" {
		t.Fatalf("results = %v", res)
	}
	if adv, _ := json.Marshal(m["advice"]); !strings.Contains(string(adv), "Singapore") {
		t.Fatalf("advice = %s", adv)
	}
	// Each region is reported as soon as it is measured.
	n := 0
	for _, e := range ev.all("region") {
		if _, ok := e["result"]; ok {
			n++
		}
	}
	if n != 2 || ev.all("region")[0]["total"] != float64(2) {
		t.Fatalf("events = %v", ev.all("region"))
	}
}

func TestDiagnosisRunIsNotSaved(t *testing.T) {
	a, ev, _ := newTestApp(t)
	if _, err := call(t, a, "startTest", "general", TestOptions{Diagnosis: true}); err != nil {
		t.Fatal(err)
	}
	res := ev.wait(t, "test", "result", 20*time.Second)["result"]
	if list, _ := call(t, a, "history"); len(list.([]HistoryEntry)) != 0 {
		t.Fatalf("a diagnosis run was saved: %+v", list)
	}
	// The page sends the results back as they came; Go concludes.
	got, err := call(t, a, "diagnose", []map[string]any{{"kind": "now", "result": res}})
	if err != nil {
		t.Fatal(err)
	}
	d := got.(model.Diagnosis)
	if d.Kind == "" || d.Title == "" || len(d.Facts) == 0 || !strings.HasPrefix(d.Facts[0], "Where you are: download") {
		t.Fatalf("diagnosis = %+v", d)
	}
}
