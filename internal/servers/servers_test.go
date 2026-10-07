package servers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/tehgeii/wifi-speed-tester/internal/config"
)

// Shape of https://librespeed.org/backend-servers/servers.php
const sample = `[
 {"id":1,"name":"Jakarta, Indonesia","server":"//jkt.example.net/backend/","dlURL":"garbage.php","ulURL":"empty.php","pingURL":"empty.php","getIpURL":"getIP.php","sponsorName":"Example ISP"},
 {"id":2,"name":"Plain http","server":"http://plain.example.net/","dlURL":"garbage.php","ulURL":"empty.php","pingURL":"empty.php"},
 {"id":3,"name":"","server":"//noname.example.net/"},
 {"id":4,"name":"Bad scheme","server":"ftp://x/"}
]`

func TestParse(t *testing.T) {
	list, err := Parse([]byte(sample))
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 2 {
		t.Fatalf("got %d servers: %+v", len(list), list)
	}
	s := list[0]
	if s.Type != config.LibreSpeed || s.URL != "https://jkt.example.net/backend/" || s.Sponsor != "Example ISP" {
		t.Fatalf("server = %+v", s)
	}
	if s.DownloadFor(1<<20) != "https://jkt.example.net/backend/garbage.php?ckSize=1" {
		t.Fatal(s.DownloadFor(1 << 20))
	}
	if _, err := Parse([]byte(`not json`)); err == nil {
		t.Fatal("expected error")
	}
	if _, err := Parse([]byte(`[]`)); err == nil {
		t.Fatal("expected error for empty list")
	}
}

func TestFetchAndProbe(t *testing.T) {
	mk := func(delay time.Duration) *httptest.Server {
		return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			time.Sleep(delay)
		}))
	}
	slow, fast := mk(40*time.Millisecond), mk(0)
	defer slow.Close()
	defer fast.Close()
	list := `[{"name":"Slow","server":"` + slow.URL + `/"},{"name":"Down","server":"http://127.0.0.1:1/"},{"name":"Fast","server":"` + fast.URL + `/"}]`
	src := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.Write([]byte(list)) }))
	defer src.Close()

	servers, err := Fetch(context.Background(), http.DefaultClient, src.URL)
	if err != nil {
		t.Fatal(err)
	}
	calls := 0
	ranked := Probe(context.Background(), servers, 4, func(done, total int) { calls++ })
	if calls != 3 {
		t.Fatalf("progress calls = %d", calls)
	}
	if ranked[0].Server.Name != "Fast" || ranked[1].Server.Name != "Slow" || ranked[2].Server.Name != "Down" {
		t.Fatalf("order = %s, %s, %s", ranked[0].Server.Name, ranked[1].Server.Name, ranked[2].Server.Name)
	}
	if ranked[2].Error == "" || ranked[0].LatencyMs <= 0 || ranked[1].LatencyMs < 40 {
		t.Fatalf("ranked = %+v", ranked)
	}
}
