package network

import (
	"context"
	"errors"
	"net"
	"net/http"
	"time"
)

// CheckDNS resolves host with the system resolver.
func CheckDNS(ctx context.Context, host string) (time.Duration, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	start := time.Now()
	addrs, err := net.DefaultResolver.LookupHost(ctx, host)
	if err != nil {
		return 0, err
	}
	if len(addrs) == 0 {
		return 0, errors.New("no addresses returned")
	}
	return time.Since(start), nil
}

// CheckInternet makes a tiny HTTPS request to url and reports whether any
// HTTP response came back.
func CheckInternet(ctx context.Context, client *http.Client, url string) error {
	ctx, cancel := context.WithTimeout(ctx, 8*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	resp.Body.Close()
	return nil
}
