// Package config holds user-tunable settings: test servers, ping targets,
// test durations and the connection-quality thresholds.
//
// The portable build reads an optional "WiFiSpeedTester.config.json" placed
// next to the executable. Any field left out keeps its default.
package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// FileName is the optional config file looked up next to the executable.
const FileName = "WiFiSpeedTester.config.json"

// Server is an HTTP speed-test endpoint.
//
// DownloadURL must contain "{bytes}", which is replaced with the requested
// payload size. UploadURL must accept a POST body of arbitrary size and
// discard it.
type Server struct {
	Name        string `json:"name"`
	DownloadURL string `json:"downloadUrl"`
	UploadURL   string `json:"uploadUrl"`
}

// PingTarget is a host to measure latency against. Host "gateway" means the
// default gateway of the active adapter.
type PingTarget struct {
	Label string `json:"label"`
	Host  string `json:"host"`
}

// Band is the threshold triple for one metric. For "lower is better" metrics a
// value <= Excellent is excellent, <= Good is good, <= Fair is fair, otherwise
// poor. For "higher is better" metrics the comparisons are >=.
type Band struct {
	Excellent float64 `json:"excellent"`
	Good      float64 `json:"good"`
	Fair      float64 `json:"fair"`
}

// Profile is a set of thresholds. Metrics listed in Use are graded; others are
// shown but do not affect the overall level.
type Profile struct {
	Use                []string `json:"use"`
	PingMs             Band     `json:"pingMs"`
	JitterMs           Band     `json:"jitterMs"`
	PacketLossPct      Band     `json:"packetLossPct"`
	DownloadMbps       Band     `json:"downloadMbps"`
	UploadMbps         Band     `json:"uploadMbps"`
	LoadedLatencyIncMs Band     `json:"loadedLatencyIncreaseMs"`
	UnstableLossPct    float64  `json:"unstableLossPct"`
	UnstableJitterMs   float64  `json:"unstableJitterMs"`
}

// Config is the full application configuration.
type Config struct {
	Servers           []Server           `json:"servers"`
	PingTargets       []PingTarget       `json:"pingTargets"`
	PingCount         int                `json:"pingCount"`
	PingIntervalMs    int                `json:"pingIntervalMs"`
	PingTimeoutMs     int                `json:"pingTimeoutMs"`
	DownloadSeconds   float64            `json:"downloadSeconds"`
	UploadSeconds     float64            `json:"uploadSeconds"`
	WarmupSeconds     float64            `json:"warmupSeconds"`
	Streams           int                `json:"streams"`
	MeasureLoadedPing bool               `json:"measureLoadedPing"`
	ConnectivityHost  string             `json:"connectivityHost"`
	HistoryLimit      int                `json:"historyLimit"`
	Profiles          map[string]Profile `json:"profiles"`
}

// Default returns the built-in configuration.
func Default() Config {
	return Config{
		Servers: []Server{
			{
				Name:        "Cloudflare",
				DownloadURL: "https://speed.cloudflare.com/__down?bytes={bytes}",
				UploadURL:   "https://speed.cloudflare.com/__up",
			},
		},
		PingTargets: []PingTarget{
			{Label: "Local Gateway", Host: "gateway"},
			{Label: "Cloudflare DNS", Host: "1.1.1.1"},
			{Label: "Google DNS", Host: "8.8.8.8"},
		},
		PingCount:         20,
		PingIntervalMs:    200,
		PingTimeoutMs:     1000,
		DownloadSeconds:   10,
		UploadSeconds:     10,
		WarmupSeconds:     1.5,
		Streams:           4,
		MeasureLoadedPing: true,
		ConnectivityHost:  "speed.cloudflare.com",
		HistoryLimit:      200,
		Profiles: map[string]Profile{
			"general": {
				Use:                []string{"ping", "jitter", "packetLoss", "download", "upload"},
				PingMs:             Band{Excellent: 20, Good: 50, Fair: 100},
				JitterMs:           Band{Excellent: 5, Good: 15, Fair: 30},
				PacketLossPct:      Band{Excellent: 0, Good: 0.5, Fair: 2},
				DownloadMbps:       Band{Excellent: 100, Good: 25, Fair: 10},
				UploadMbps:         Band{Excellent: 20, Good: 5, Fair: 2},
				LoadedLatencyIncMs: Band{Excellent: 10, Good: 30, Fair: 100},
				UnstableLossPct:    5,
				UnstableJitterMs:   50,
			},
			"gaming": {
				Use:                []string{"ping", "jitter", "packetLoss", "loadedLatency"},
				PingMs:             Band{Excellent: 20, Good: 40, Fair: 70},
				JitterMs:           Band{Excellent: 3, Good: 8, Fair: 20},
				PacketLossPct:      Band{Excellent: 0, Good: 0, Fair: 1},
				DownloadMbps:       Band{Excellent: 50, Good: 15, Fair: 5},
				UploadMbps:         Band{Excellent: 10, Good: 3, Fair: 1},
				LoadedLatencyIncMs: Band{Excellent: 10, Good: 30, Fair: 80},
				UnstableLossPct:    2,
				UnstableJitterMs:   30,
			},
		},
	}
}

// PingInterval etc. are convenience accessors.
func (c Config) PingInterval() time.Duration {
	return time.Duration(c.PingIntervalMs) * time.Millisecond
}

func (c Config) PingTimeout() time.Duration {
	return time.Duration(c.PingTimeoutMs) * time.Millisecond
}

func secs(f float64) time.Duration { return time.Duration(f * float64(time.Second)) }

func (c Config) DownloadDuration() time.Duration { return secs(c.DownloadSeconds) }
func (c Config) UploadDuration() time.Duration   { return secs(c.UploadSeconds) }
func (c Config) Warmup() time.Duration           { return secs(c.WarmupSeconds) }

// Profile returns the named profile, falling back to "general".
func (c Config) Profile(name string) Profile {
	if p, ok := c.Profiles[name]; ok {
		return p
	}
	return c.Profiles["general"]
}

// Validate rejects settings that would make tests meaningless.
func (c Config) Validate() error {
	if len(c.Servers) == 0 {
		return fmt.Errorf("at least one test server is required")
	}
	for _, s := range c.Servers {
		if s.DownloadURL == "" || s.UploadURL == "" {
			return fmt.Errorf("server %q needs both downloadUrl and uploadUrl", s.Name)
		}
	}
	if c.PingCount < 1 || c.PingCount > 1000 {
		return fmt.Errorf("pingCount must be between 1 and 1000")
	}
	if c.PingIntervalMs < 10 || c.PingTimeoutMs < 100 {
		return fmt.Errorf("pingIntervalMs must be >= 10 and pingTimeoutMs >= 100")
	}
	if c.DownloadSeconds < 2 || c.UploadSeconds < 2 || c.DownloadSeconds > 120 || c.UploadSeconds > 120 {
		return fmt.Errorf("downloadSeconds/uploadSeconds must be between 2 and 120")
	}
	if c.WarmupSeconds < 0 || c.WarmupSeconds >= c.DownloadSeconds || c.WarmupSeconds >= c.UploadSeconds {
		return fmt.Errorf("warmupSeconds must be shorter than the test duration")
	}
	if c.Streams < 1 || c.Streams > 16 {
		return fmt.Errorf("streams must be between 1 and 16")
	}
	if _, ok := c.Profiles["general"]; !ok {
		return fmt.Errorf(`a "general" profile is required`)
	}
	return nil
}

// Load reads the config file in dir, merging it over the defaults. A missing
// file is not an error. The second return value is the path that was used,
// or "" when only defaults apply.
func Load(dir string) (Config, string, error) {
	cfg := Default()
	path := filepath.Join(dir, FileName)
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return cfg, "", nil
	}
	if err != nil {
		return cfg, path, err
	}
	if err := Parse(data, &cfg); err != nil {
		return Default(), path, fmt.Errorf("%s: %w", FileName, err)
	}
	return cfg, path, nil
}

// Parse merges JSON over cfg and validates the result. Profiles given in the
// JSON replace the default profile of the same name entirely.
func Parse(data []byte, cfg *Config) error {
	defaults := cfg.Profiles
	cfg.Profiles = nil
	if err := json.Unmarshal(data, cfg); err != nil {
		cfg.Profiles = defaults
		return err
	}
	merged := map[string]Profile{}
	for k, v := range defaults {
		merged[k] = v
	}
	for k, v := range cfg.Profiles {
		merged[k] = v
	}
	cfg.Profiles = merged
	return cfg.Validate()
}

// ExampleJSON returns the default config rendered as indented JSON, used for
// the example file shipped with the portable zip.
func ExampleJSON() []byte {
	b, _ := json.MarshalIndent(Default(), "", "  ")
	return append(b, '\n')
}
