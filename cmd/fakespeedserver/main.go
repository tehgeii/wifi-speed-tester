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
	log.Printf("fake speed server on http://%s (download: /down?bytes={bytes}, upload: /up)", ln.Addr())
	log.Fatal(http.Serve(ln, s.Config.Handler))
}
