// Command fakespeedserver is a development aid: a local speed-test server
// with a bandwidth cap, for exercising the UI without internet access.
//
//	go run ./cmd/fakespeedserver -addr 127.0.0.1:8099 -rate 3000000
//
// Then point WiFiSpeedTester.config.json at
// http://127.0.0.1:8099/down?bytes={bytes} and http://127.0.0.1:8099/up.
package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"

	"github.com/tehgeii/wifi-speed-tester/internal/testutil"
)

func main() {
	addr := flag.String("addr", "127.0.0.1:8099", "listen address")
	rate := flag.Int64("rate", 3_000_000, "bytes per second per connection (0 = unlimited)")
	status := flag.Int("status", 0, "force this HTTP status on every request (simulate failures)")
	flag.Parse()
	s := testutil.NewSpeedServer(*rate)
	s.Status = *status
	s.Close() // only the handler is reused
	ln, err := net.Listen("tcp", *addr)
	if err != nil {
		log.Fatal(err)
	}
	base := "http://" + ln.Addr().String()
	mux := http.NewServeMux()
	mux.Handle("/", s.Config.Handler)
	// A LibreSpeed-style server list naming this server's /backend/ plus one
	// unreachable entry, for trying the server picker
	// (set "serverListUrl": "<base>/servers.php" in the config).
	mux.HandleFunc("/servers.php", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `[{"id":1,"name":"Local LibreSpeed (fake)","server":"%s/backend/","dlURL":"garbage.php","ulURL":"empty.php","pingURL":"empty.php","sponsorName":"localhost"},`+
			`{"id":2,"name":"Unreachable example","server":"http://127.0.0.1:1/","dlURL":"garbage.php","ulURL":"empty.php","pingURL":"empty.php"}]`, base)
	})
	log.Printf("fake speed server on %s (download: /down?bytes={bytes}, upload: /up, LibreSpeed: /backend/, list: /servers.php)", base)
	log.Fatal(http.Serve(ln, mux))
}
