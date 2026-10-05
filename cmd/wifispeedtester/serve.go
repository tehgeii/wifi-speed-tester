package main

import (
	"net"
	"net/http"
	"time"

	"github.com/tehgeii/wifi-speed-tester/internal/ui"
)

func serveHTTP(ln net.Listener, srv *ui.DevServer) error {
	s := &http.Server{Handler: srv.Handler(), ReadHeaderTimeout: 10 * time.Second}
	return s.Serve(ln)
}
