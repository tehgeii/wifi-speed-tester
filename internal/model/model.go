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
	Note          string  `json:"note,omitempty"`
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
	Metric string       `json:"metric"`
	Value  string       `json:"value"`
	Grade  QualityLevel `json:"grade"`
}

// Quality is the deterministic classification produced by analysis.
type Quality struct {
	Level   QualityLevel  `json:"level"`
	Profile string        `json:"profile"`
	Grades  []MetricGrade `json:"grades"`
	Notes   []string      `json:"notes"`
	Summary string        `json:"summary"`
}

// TestResult is everything one run produced.
type TestResult struct {
	ID         string       `json:"id"`
	StartedAt  time.Time    `json:"startedAt"`
	FinishedAt time.Time    `json:"finishedAt"`
	Mode       string       `json:"mode"` // "general" or "gaming"
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
}

// Failure is a user-facing explanation of why a run could not complete.
type Failure struct {
	Title      string `json:"title"`
	Reason     string `json:"reason"`
	Suggestion string `json:"suggestion"`
	Technical  string `json:"technical,omitempty"`
}
