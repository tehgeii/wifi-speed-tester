package measure_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/tehgeii/wifi-speed-tester/internal/measure"
	"github.com/tehgeii/wifi-speed-tester/internal/testutil"
)

func opts() measure.SpeedOptions {
	return measure.SpeedOptions{Duration: 2500 * time.Millisecond, Warmup: 500 * time.Millisecond, Streams: 2}
}

// The server is throttled per stream to 1 MB/s, so 2 streams ~ 16 Mbps.
func TestDownloadMeasuresThrottledRate(t *testing.T) {
	srv := testutil.NewSpeedServer(1_000_000)
	defer srv.Close()
	var calls int
	o := opts()
	o.OnProgress = func(f, mbps float64) { calls++ }
	res, err := measure.Download(context.Background(), measure.NewHTTPClient(2), srv.DownloadURL(), o)
	if err != nil {
		t.Fatal(err)
	}
	if res.Mbps < 10 || res.Mbps > 22 {
		t.Fatalf("download = %.1f Mbps, want about 16", res.Mbps)
	}
	if calls == 0 || len(res.SamplesMbps) == 0 {
		t.Fatal("expected progress callbacks and samples")
	}
}

func TestUploadMeasuresThrottledRate(t *testing.T) {
	srv := testutil.NewSpeedServer(1_000_000)
	defer srv.Close()
	res, err := measure.Upload(context.Background(), measure.NewHTTPClient(2), srv.UploadURL(), opts())
	if err != nil {
		t.Fatal(err)
	}
	if res.Mbps < 10 || res.Mbps > 22 {
		t.Fatalf("upload = %.1f Mbps, want about 16", res.Mbps)
	}
}

func TestSpeedHTTPErrorIsReported(t *testing.T) {
	srv := testutil.NewSpeedServer(0)
	srv.Status = 503
	defer srv.Close()
	start := time.Now()
	_, err := measure.Download(context.Background(), measure.NewHTTPClient(2), srv.DownloadURL(), opts())
	var se *measure.HTTPStatusError
	if !errors.As(err, &se) || se.StatusCode != 503 {
		t.Fatalf("err = %v, want HTTP 503", err)
	}
	if time.Since(start) > 2*time.Second {
		t.Fatal("a fatal HTTP error should stop the test early")
	}
	_, err = measure.Upload(context.Background(), measure.NewHTTPClient(2), srv.UploadURL(), opts())
	if !errors.As(err, &se) {
		t.Fatalf("upload err = %v, want HTTP error", err)
	}
}

func TestSpeedCancellation(t *testing.T) {
	srv := testutil.NewSpeedServer(500_000)
	defer srv.Close()
	for name, run := range map[string]func(context.Context) error{
		"download": func(ctx context.Context) error {
			_, err := measure.Download(ctx, measure.NewHTTPClient(2), srv.DownloadURL(), measure.SpeedOptions{Duration: 20 * time.Second, Warmup: time.Second, Streams: 2})
			return err
		},
		"upload": func(ctx context.Context) error {
			_, err := measure.Upload(ctx, measure.NewHTTPClient(2), srv.UploadURL(), measure.SpeedOptions{Duration: 20 * time.Second, Warmup: time.Second, Streams: 2})
			return err
		},
	} {
		ctx, cancel := context.WithCancel(context.Background())
		time.AfterFunc(300*time.Millisecond, cancel)
		start := time.Now()
		err := run(ctx)
		if !errors.Is(err, context.Canceled) {
			t.Errorf("%s: err = %v, want context.Canceled", name, err)
		}
		if d := time.Since(start); d > 2*time.Second {
			t.Errorf("%s: cancellation took %v", name, d)
		}
	}
}

func TestUnreachableServer(t *testing.T) {
	o := opts()
	o.Duration = 2 * time.Second
	_, err := measure.Download(context.Background(), measure.NewHTTPClient(1), "http://127.0.0.1:1/down?bytes={bytes}", o)
	if err == nil {
		t.Fatal("expected an error for an unreachable server")
	}
}
