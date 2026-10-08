// Package model holds the plain data types produced by the network tests.
// Measurement code fills these in; analysis, export, history and the UI only
// read them.
package model

import "time"

// ConnectionType is the kind of the active network link.
type ConnectionType string

const (
	ConnUnknown   ConnectionType = "Unknown"
	ConnWiFi      ConnectionType = "Wi-Fi"
	ConnEthernet  ConnectionType = "Ethernet"
	ConnTethering ConnectionType = "Hotspot / Tethering"
	ConnCellular  ConnectionType = "Cellular"
	ConnVPN       ConnectionType = "VPN / Virtual"
)

// NetworkInfo describes the active network interface.
type NetworkInfo struct {
	Connection  ConnectionType `json:"connection"`
	AdapterName string         `json:"adapterName"`
	Description string         `json:"description"`
	IPv4        []string       `json:"ipv4"`
	IPv6        []string       `json:"ipv6"`
	Gateway     string         `json:"gateway"`
	DNS         []string       `json:"dns"`
	WiFi        *WiFiInfo      `json:"wifi,omitempty"`
}

// WiFiInfo is extra detail for a Wi-Fi link, when the OS provides it.
// LinkSpeedMbps is the radio link rate, NOT the internet speed.
type WiFiInfo struct {
	SSID          string  `json:"ssid"`
	BSSID         string  `json:"bssid"`
	SignalPercent int     `json:"signalPercent"`
	RSSI          int     `json:"rssi"` // dBm, 0 when unknown
	Channel       int     `json:"channel"`
	Band          string  `json:"band"`
	FrequencyMHz  int     `json:"frequencyMHz"`
	PhyType       string  `json:"phyType"`
	RxRateMbps    float64 `json:"rxRateMbps"`
	TxRateMbps    float64 `json:"txRateMbps"`
	// NoteKey names an i18n message ("wifi.hotspot", "wifi.ssidHidden",
	// "wifi.unavailable"); Note holds its argument or raw detail.
	NoteKey string `json:"noteKey,omitempty"`
	Note    string `json:"note,omitempty"`
}

// CheckStatus is the state of one connectivity check step.
type CheckStatus string

const (
	CheckPending CheckStatus = "pending"
	CheckOK      CheckStatus = "ok"
	CheckWarn    CheckStatus = "warn"
	CheckFail    CheckStatus = "fail"
)

// CheckItem is one line in the "Internet Check" list.
type CheckItem struct {
	Key    string      `json:"key"` // adapter | gateway | dns | internet
	Name   string      `json:"name"`
	Status CheckStatus `json:"status"`
	Detail string      `json:"detail"`
}

// PingStats summarises an ICMP echo run against one target.
type PingStats struct {
	Label         string    `json:"label"`
	Target        string    `json:"target"`
	Sent          int       `json:"sent"`
	Received      int       `json:"received"`
	PacketLossPct float64   `json:"packetLossPct"`
	MinMs         float64   `json:"minMs"`
	AvgMs         float64   `json:"avgMs"`
	MaxMs         float64   `json:"maxMs"`
	JitterMs      float64   `json:"jitterMs"`
	SamplesMs     []float64 `json:"samplesMs"`
	Error         string    `json:"error,omitempty"`
	Method        string    `json:"method"` // "icmp" or "tcp"
	IsGateway     bool      `json:"isGateway"`
	PrimaryTarget bool      `json:"primaryTarget"`
}

// SpeedResult is the outcome of a download or upload test.
type SpeedResult struct {
	Mbps          float64   `json:"mbps"`
	Bytes         int64     `json:"bytes"`
	DurationSec   float64   `json:"durationSec"`
	Streams       int       `json:"streams"`
	Server        string    `json:"server"`
	SamplesMbps   []float64 `json:"samplesMbps"`
	LoadedPingMs  float64   `json:"loadedPingMs"` // median latency while loaded, 0 when not measured
	LoadedSamples int       `json:"loadedSamples"`
	Error         string    `json:"error,omitempty"`
}

// QualityLevel is the overall connection classification.
type QualityLevel string

const (
	QualityExcellent QualityLevel = "EXCELLENT"
	QualityGood      QualityLevel = "GOOD"
	QualityFair      QualityLevel = "FAIR"
	QualityPoor      QualityLevel = "POOR"
	QualityUnstable  QualityLevel = "UNSTABLE"
	QualityUnknown   QualityLevel = "UNKNOWN"
)

// MetricGrade is the grade of a single metric inside the quality summary.
type MetricGrade struct {
	Key    string       `json:"key"` // ping | jitter | packetLoss | download | upload | loadedLatency
	Metric string       `json:"metric"`
	Value  string       `json:"value"`
	Grade  QualityLevel `json:"grade"`
	// Counted is true when the metric affects the overall level in this
	// profile. Limits are the excellent/good/fair thresholds in Unit;
	// HigherBetter tells how to read them.
	Counted      bool       `json:"counted"`
	Limits       [3]float64 `json:"limits"`
	Unit         string     `json:"unit"`
	HigherBetter bool       `json:"higherBetter"`
}

// Quality is the deterministic classification produced by analysis.
type Quality struct {
	Level   QualityLevel  `json:"level"`
	Profile string        `json:"profile"`
	Grades  []MetricGrade `json:"grades"`
	Notes   []string      `json:"notes"` // what was observed
	Tips    []string      `json:"tips"`  // what the user can do about it
	Summary string        `json:"summary"`
	Lang    string        `json:"lang"`
}

// TestResult is everything one run produced.
type TestResult struct {
	ID         string       `json:"id"`
	StartedAt  time.Time    `json:"startedAt"`
	FinishedAt time.Time    `json:"finishedAt"`
	Mode       string       `json:"mode"` // "general", "gaming" or "quick"
	Network    *NetworkInfo `json:"network,omitempty"`
	Checks     []CheckItem  `json:"checks"`
	Pings      []PingStats  `json:"pings"`
	Download   *SpeedResult `json:"download,omitempty"`
	Upload     *SpeedResult `json:"upload,omitempty"`
	PingMs     float64      `json:"pingMs"`
	JitterMs   float64      `json:"jitterMs"`
	PacketLoss float64      `json:"packetLossPct"`
	Quality    *Quality     `json:"quality,omitempty"`
	Cancelled  bool         `json:"cancelled"`
	Failure    *Failure     `json:"failure,omitempty"`
	Plan       *Plan        `json:"plan,omitempty"` // the user's ISP plan at test time
}

// Plan is the internet package speed the user pays for, in Mbps (0 = not set).
type Plan struct {
	DownMbps float64 `json:"downMbps"`
	UpMbps   float64 `json:"upMbps"`
	LowPct   float64 `json:"lowPct"` // below this % a tip suggests action (config planLowPct)
}

// PlanPercent returns measured/plan in percent, or 0 when either is unknown.
func PlanPercent(measured, plan float64) float64 {
	if measured <= 0 || plan <= 0 {
		return 0
	}
	return measured / plan * 100
}

// WiFiNetwork is one access point seen in a Wi-Fi scan.
type WiFiNetwork struct {
	SSID          string `json:"ssid"` // "" for hidden networks
	BSSID         string `json:"bssid"`
	RSSI          int    `json:"rssi"`          // dBm
	SignalPercent int    `json:"signalPercent"` // 0..100
	Channel       int    `json:"channel"`
	Band          string `json:"band"`
	FrequencyMHz  int    `json:"frequencyMHz"`
	Connected     bool   `json:"connected"` // the access point this device uses
}

// DNSResult is the lookup speed of one DNS server.
type DNSResult struct {
	Label    string  `json:"label"`
	Server   string  `json:"server"`
	System   bool    `json:"system"` // configured on this computer
	MedianMs float64 `json:"medianMs"`
	MinMs    float64 `json:"minMs"`
	OK       int     `json:"ok"`
	Failed   int     `json:"failed"`
	Error    string  `json:"error,omitempty"`
}

// MonitorSample is one probe of the stability monitor.
type MonitorSample struct {
	T     float64 `json:"t"`     // seconds since start
	RttMs float64 `json:"rttMs"` // 0 when lost
	OK    bool    `json:"ok"`
	GwOK  *bool   `json:"gwOk,omitempty"` // router answered (nil when not probed)
}

// MonitorSummary describes a finished stability monitor run.
type MonitorSummary struct {
	StartedAt     time.Time    `json:"startedAt"`
	DurationSec   float64      `json:"durationSec"`
	Target        string       `json:"target"`
	TargetLabel   string       `json:"targetLabel"`
	Method        string       `json:"method"` // icmp | tcp
	Sent          int          `json:"sent"`
	Received      int          `json:"received"`
	LossPct       float64      `json:"lossPct"`
	AvgMs         float64      `json:"avgMs"`
	MinMs         float64      `json:"minMs"`
	MaxMs         float64      `json:"maxMs"`
	P95Ms         float64      `json:"p95Ms"`
	JitterMs      float64      `json:"jitterMs"`
	Spikes        int          `json:"spikes"`  // probes far above the median
	Outages       int          `json:"outages"` // runs of >= 3 lost probes
	LongestOutage float64      `json:"longestOutageSec"`
	GatewaySent   int          `json:"gatewaySent"`
	GatewayLost   int          `json:"gatewayLost"`
	Cancelled     bool         `json:"cancelled"`
	Verdict       string       `json:"verdict"` // localized one-line conclusion
	Notes         []string     `json:"notes"`
	Network       *NetworkInfo `json:"network,omitempty"`
}

// Failure is a user-facing explanation of why a run could not complete.
type Failure struct {
	Title      string `json:"title"`
	Reason     string `json:"reason"`
	Suggestion string `json:"suggestion"`
	Technical  string `json:"technical,omitempty"`
}
