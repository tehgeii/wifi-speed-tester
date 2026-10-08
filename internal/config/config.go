// Package config holds user-tunable settings: test servers, ping targets,
// test durations and the connection-quality thresholds.
//
// The portable build reads an optional "WiFiSpeedTester.config.json" placed
// next to the executable. Any field left out keeps its default.
package config

import (
	"encoding/json"
	"fmt"
	"net"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// FileName is the optional config file looked up next to the executable.
const FileName = "WiFiSpeedTester.config.json"

// Server is an HTTP speed-test endpoint, in one of two forms:
//
//   - URL templates (Type ""): DownloadURL contains "{bytes}", replaced with
//     the requested payload size; UploadURL accepts a POST of any size.
//   - A LibreSpeed backend (Type "librespeed"): URL is the backend base
//     address; downloads use garbage.php?ckSize=<MiB>, uploads and pings use
//     empty.php (paths can be overridden).
type Server struct {
	Name        string `json:"name"`
	Type        string `json:"type,omitempty"`
	DownloadURL string `json:"downloadUrl,omitempty"`
	UploadURL   string `json:"uploadUrl,omitempty"`

	URL      string `json:"url,omitempty"`
	DLPath   string `json:"dlPath,omitempty"`
	ULPath   string `json:"ulPath,omitempty"`
	PingPath string `json:"pingPath,omitempty"`
	Sponsor  string `json:"sponsor,omitempty"`
}

// LibreSpeed is the Type of a LibreSpeed backend.
const LibreSpeed = "librespeed"

func (s Server) base() string {
	if strings.HasSuffix(s.URL, "/") {
		return s.URL
	}
	return s.URL + "/"
}

func orDefault(v, d string) string {
	if v == "" {
		return d
	}
	return v
}

// DownloadFor returns the URL that serves roughly n bytes.
func (s Server) DownloadFor(n int) string {
	if s.Type == LibreSpeed {
		mib := (n + 1<<20 - 1) >> 20 // garbage.php counts whole MiB chunks
		mib = min(max(mib, 1), 1024)
		return s.base() + orDefault(s.DLPath, "garbage.php") + "?ckSize=" + strconv.Itoa(mib)
	}
	return strings.ReplaceAll(s.DownloadURL, "{bytes}", strconv.Itoa(n))
}

// Upload returns the URL that accepts upload POSTs.
func (s Server) Upload() string {
	if s.Type == LibreSpeed {
		return s.base() + orDefault(s.ULPath, "empty.php")
	}
	return s.UploadURL
}

// Ping returns a URL answering a tiny GET, used for reachability and
// latency checks.
func (s Server) Ping() string {
	if s.Type == LibreSpeed {
		return s.base() + orDefault(s.PingPath, "empty.php")
	}
	return s.DownloadFor(0)
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
	ServerListURL     string             `json:"serverListUrl"` // public LibreSpeed list, fetched only on request
	DNSServers        []PingTarget       `json:"dnsServers"`    // public resolvers compared in the DNS test
	PlanLowPct        float64            `json:"planLowPct"`    // below this % of the ISP plan a tip suggests action
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
		ServerListURL:     "https://librespeed.org/backend-servers/servers.php",
		DNSServers: []PingTarget{
			{Label: "Cloudflare", Host: "1.1.1.1"},
			{Label: "Google", Host: "8.8.8.8"},
			{Label: "Quad9", Host: "9.9.9.9"},
		},
		PlanLowPct:   50,
		HistoryLimit: 200,
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
			// Quick Ping: latency only, judged like the gaming profile.
			"quick": {
				Use:                []string{"ping", "jitter", "packetLoss"},
				PingMs:             Band{Excellent: 20, Good: 40, Fair: 70},
				JitterMs:           Band{Excellent: 3, Good: 8, Fair: 20},
				PacketLossPct:      Band{Excellent: 0, Good: 0, Fair: 1},
				DownloadMbps:       Band{Excellent: 50, Good: 15, Fair: 5},
				UploadMbps:         Band{Excellent: 10, Good: 3, Fair: 1},
				LoadedLatencyIncMs: Band{Excellent: 10, Good: 30, Fair: 80},
				UnstableLossPct:    2,
				UnstableJitterMs:   30,
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

// Validate checks that a server has the fields its type needs, with http(s)
// URLs.
func (s Server) Validate() error {
	check := func(field, v string) error {
		u, err := url.Parse(v)
		if err != nil || (u.Scheme != "https" && u.Scheme != "http") || u.Host == "" {
			return fmt.Errorf("server %q: %s must be an http(s) URL", s.Name, field)
		}
		return nil
	}
	switch s.Type {
	case LibreSpeed:
		return check("url", s.URL)
	case "":
		if err := check("downloadUrl", s.DownloadURL); err != nil {
			return err
		}
		if !strings.Contains(s.DownloadURL, "{bytes}") {
			return fmt.Errorf("server %q: downloadUrl must contain {bytes}", s.Name)
		}
		return check("uploadUrl", s.UploadURL)
	}
	return fmt.Errorf("server %q: unknown type %q", s.Name, s.Type)
}

// Validate rejects settings that would make tests meaningless.
func (c Config) Validate() error {
	if len(c.Servers) == 0 {
		return fmt.Errorf("at least one test server is required")
	}
	for _, s := range c.Servers {
		if err := s.Validate(); err != nil {
			return err
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
	for _, d := range c.DNSServers {
		if net.ParseIP(d.Host) == nil {
			return fmt.Errorf("dnsServers: %q must be an IP address", d.Host)
		}
	}
	if c.PlanLowPct < 0 || c.PlanLowPct > 100 {
		return fmt.Errorf("planLowPct must be between 0 and 100")
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

var hostRe = regexp.MustCompile(`^[A-Za-z0-9]([A-Za-z0-9-]{0,62}[A-Za-z0-9])?(\.[A-Za-z0-9]([A-Za-z0-9-]{0,62}[A-Za-z0-9])?)*$`)

// ValidTargetHost reports whether h is usable as a ping target: an IP
// address or a DNS name (the special "gateway" is handled separately).
func ValidTargetHost(h string) bool {
	if net.ParseIP(h) != nil {
		return true
	}
	return len(h) <= 253 && hostRe.MatchString(h) && !strings.EqualFold(h, "gateway")
}
