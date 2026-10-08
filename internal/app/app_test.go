package app

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/tehgeii/wifi-speed-tester/internal/config"
	"github.com/tehgeii/wifi-speed-tester/internal/measure"
	"github.com/tehgeii/wifi-speed-tester/internal/model"
	"github.com/tehgeii/wifi-speed-tester/internal/testutil"
)

type events struct {
	mu   sync.Mutex
	list []struct {
		name    string
		payload any
	}
}

func (e *events) emit(name string, p any) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.list = append(e.list, struct {
		name    string
		payload any
	}{name, p})
}

// wait returns the first event called name whose payload has key.
func (e *events) wait(t *testing.T, name, key string, timeout time.Duration) map[string]any {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		e.mu.Lock()
		for _, ev := range e.list {
			if ev.name != name {
				continue
			}
			b, _ := json.Marshal(ev.payload)
			var m map[string]any
			_ = json.Unmarshal(b, &m)
			if _, ok := m[key]; ok {
				e.mu.Unlock()
				return m
			}
		}
		e.mu.Unlock()
		time.Sleep(20 * time.Millisecond)
	}
	t.Fatalf("no %s event with %q", name, key)
	return nil
}

func newTestApp(t *testing.T) (*App, *events, *testutil.SpeedServer) {
	t.Helper()
	srv := testutil.NewSpeedServer(2_000_000)
	t.Cleanup(srv.Close)
	dir := t.TempDir()
	cfg := `{"servers":[{"name":"Local","downloadUrl":"` + srv.DownloadURL() + `","uploadUrl":"` + srv.UploadURL() + `"}],
	  "connectivityHost":"localhost","pingCount":5,"pingIntervalMs":10,
	  "downloadSeconds":2,"uploadSeconds":2,"warmupSeconds":0.5,"streams":2}`
	if err := os.WriteFile(filepath.Join(dir, config.FileName), []byte(cfg), 0o644); err != nil {
		t.Fatal(err)
	}
	ev := &events{}
	a := New(dir, Host{Emit: ev.emit}, nil)
	if a.cfgErr != nil {
		t.Fatal(a.cfgErr)
	}
	p := &testutil.Pinger{RTT: time.Millisecond}
	a.eng.Detect = func(context.Context) (*model.NetworkInfo, error) {
		return &model.NetworkInfo{Connection: model.ConnEthernet, AdapterName: "eth", Gateway: "192.168.1.1", DNS: []string{"192.168.1.1"}}, nil
	}
	a.eng.NewPinger = func() (measure.Pinger, error) { return p, nil }
	a.monitorUnit = 300 * time.Millisecond
	return a, ev, srv
}

func call(t *testing.T, a *App, method string, args ...any) (any, error) {
	t.Helper()
	var params []json.RawMessage
	for _, x := range args {
		b, _ := json.Marshal(x)
		params = append(params, b)
	}
	return a.Call(method, params)
}

func TestPrefsAndPlan(t *testing.T) {
	a, _, _ := newTestApp(t)
	if a.plan() != nil {
		t.Fatal("no plan by default")
	}
	if _, err := call(t, a, "setPrefs", map[string]string{"lang": "id", "planDown": "50,5", "planUp": "10", "evil": "x"}); err != nil {
		t.Fatal(err)
	}
	p, _ := call(t, a, "getPrefs")
	if _, bad := p.(map[string]string)["evil"]; bad {
		t.Fatal("unknown pref key saved")
	}
	pl := a.plan()
	if pl == nil || pl.DownMbps != 50.5 || pl.UpMbps != 10 || pl.LowPct != 50 {
		t.Fatalf("plan = %+v", pl)
	}
	_, _ = call(t, a, "setPrefs", map[string]string{"planDown": "-3"})
	if a.plan() != nil {
		t.Fatal("negative plan must be ignored")
	}
}

func TestCustomTargets(t *testing.T) {
	a, _, _ := newTestApp(t)
	if _, err := call(t, a, "setTargets", []config.PingTarget{{Label: "  SG game ", Host: " sg.example.com "}, {Host: "203.0.113.5"}}); err != nil {
		t.Fatal(err)
	}
	got := a.customTargets()
	if len(got) != 2 || got[0].Label != "SG game" || got[0].Host != "sg.example.com" || got[1].Label != "203.0.113.5" {
		t.Fatalf("targets = %+v", got)
	}
	mt := a.monitorTargets()
	if mt[0].Host != "1.1.1.1" || mt[len(mt)-1].Host != "203.0.113.5" {
		t.Fatalf("monitor targets = %+v", mt)
	}
	for _, bad := range [][]config.PingTarget{
		{{Label: "x", Host: "bad host!"}},
		{{Label: "x", Host: "gateway"}},
		{{Host: "1.1.1.1"}, {Host: "1.1.1.2"}, {Host: "1.1.1.3"}, {Host: "1.1.1.4"}, {Host: "1.1.1.5"}, {Host: "1.1.1.6"}},
	} {
		if _, err := call(t, a, "setTargets", bad); err == nil {
			t.Errorf("accepted %+v", bad)
		}
	}
}

func TestTestRunStoresPlanTargetsAndBlocksMonitor(t *testing.T) {
	a, ev, _ := newTestApp(t)
	_, _ = call(t, a, "setPrefs", map[string]string{"planDown": "100"})
	_, _ = call(t, a, "setTargets", []config.PingTarget{{Label: "Mine", Host: "203.0.113.9"}})
	if _, err := call(t, a, "startTest", "general"); err != nil {
		t.Fatal(err)
	}
	if _, err := call(t, a, "startMonitor", MonitorRequest{Minutes: 1}); err == nil || !strings.Contains(err.Error(), "test is already running") {
		t.Fatalf("monitor during test: %v", err)
	}
	m := ev.wait(t, "test", "result", 20*time.Second)
	res := m["result"].(map[string]any)
	if res["plan"] == nil {
		t.Fatal("result has no plan")
	}
	found := false
	for _, p := range res["pings"].([]any) {
		if p.(map[string]any)["label"] == "Mine" {
			found = true
		}
	}
	if !found {
		t.Fatal("custom target was not pinged")
	}
	notes, _ := json.Marshal(res["quality"])
	if !strings.Contains(string(notes), "of your 100 Mbps plan") {
		t.Fatalf("quality = %s", notes)
	}
	list, _ := call(t, a, "history")
	if h := list.([]HistoryEntry); len(h) != 1 || h[0].PlanPct <= 0 {
		t.Fatalf("history = %+v", h)
	}
	// Export with an id that is not in history falls back to the latest result.
	if _, err := a.Export(ExportRequest{Format: "txt", ID: "missing"}); err != nil {
		t.Fatalf("export fallback: %v", err)
	}
}

func TestMonitorRunExportAndReport(t *testing.T) {
	a, ev, _ := newTestApp(t)
	if _, err := call(t, a, "startMonitor", MonitorRequest{Minutes: 1, Host: "8.8.8.8"}); err != nil {
		t.Fatal(err)
	}
	if _, err := call(t, a, "startTest", "general"); err == nil {
		t.Fatal("test must not start while the monitor runs")
	}
	m := ev.wait(t, "monitor", "summary", 5*time.Second)
	sum := m["summary"].(map[string]any)
	if sum["target"] != "8.8.8.8" || sum["sent"].(float64) == 0 || sum["verdict"] == "" || sum["gatewaySent"].(float64) == 0 {
		t.Fatalf("summary = %+v", sum)
	}
	if len(a.monitorHistory()) != 1 {
		t.Fatal("monitor run not stored")
	}
	if _, err := call(t, a, "startTest", "quick"); err != nil {
		t.Fatalf("test after monitor: %v", err)
	}
	ev.wait(t, "test", "result", 10*time.Second)
	p, err := a.exportMonitor(false)
	if err != nil || !strings.HasSuffix(p, ".txt") {
		t.Fatalf("export monitor: %q %v", p, err)
	}
	p, err = a.ispReport(ReportRequest{Unit: "Mbps"})
	if err != nil {
		t.Fatal(err)
	}
	b, _ := os.ReadFile(p)
	if !strings.Contains(string(b), "STABILITY MONITOR") {
		t.Fatalf("report lacks monitor runs:\n%s", b)
	}
}

func TestMonitorStop(t *testing.T) {
	a, ev, _ := newTestApp(t)
	a.monitorUnit = time.Minute
	if _, err := call(t, a, "startMonitor", MonitorRequest{Minutes: 5}); err != nil {
		t.Fatal(err)
	}
	ev.wait(t, "monitor", "sample", 5*time.Second)
	_, _ = call(t, a, "stopMonitor")
	m := ev.wait(t, "monitor", "summary", 5*time.Second)
	if !m["summary"].(map[string]any)["cancelled"].(bool) {
		t.Fatal("stopped run must be marked cancelled")
	}
}
