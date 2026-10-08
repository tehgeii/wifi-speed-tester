package analysis

import (
	"math"
	"sort"

	"github.com/tehgeii/wifi-speed-tester/internal/i18n"
	"github.com/tehgeii/wifi-speed-tester/internal/model"
)

// ChannelLoad is how busy one Wi-Fi channel is.
type ChannelLoad struct {
	Band        string  `json:"band"`
	Channel     int     `json:"channel"`
	Networks    int     `json:"networks"` // other networks on exactly this channel
	Score       float64 `json:"score"`    // interference estimate, higher = busier
	Current     bool    `json:"current"`
	Recommended bool    `json:"recommended"`
}

// ScanReport is the analysed result of a Wi-Fi scan.
type ScanReport struct {
	Networks    []model.WiFiNetwork `json:"networks"` // strongest first
	Channels    []ChannelLoad       `json:"channels"`
	Current     *model.WiFiNetwork  `json:"current,omitempty"`
	Recommended int                 `json:"recommended"` // 0 = keep the current channel
	Advice      []string            `json:"advice"`
}

var (
	candidates24 = []int{1, 6, 11}
	candidates5  = []int{36, 40, 44, 48, 149, 153, 157, 161}
)

// weight is how much a neighbour disturbs: strong signals count fully,
// faint ones barely (-100 dBm → 0.05, -50 dBm or stronger → 1).
func weight(n model.WiFiNetwork) float64 {
	rssi := n.RSSI
	if rssi == 0 && n.SignalPercent > 0 {
		rssi = n.SignalPercent/2 - 100
	}
	return math.Min(1, math.Max(0.05, float64(rssi+100)/50))
}

// overlap is the share of a 20 MHz 2.4 GHz channel c that ch covers:
// channels 5 or more apart do not overlap.
func overlap(c, ch int) float64 { return math.Max(0, 1-math.Abs(float64(c-ch))/5) }

// AnalyzeScan scores channels and recommends a less crowded one. Networks
// with the same name as the connected one are treated as the user's own
// (mesh or dual-band) and not counted as interference.
func AnalyzeScan(nets []model.WiFiNetwork, lang i18n.Lang) ScanReport {
	T := func(key string, args ...any) string { return i18n.T(lang, key, args...) }
	rep := ScanReport{Networks: append([]model.WiFiNetwork(nil), nets...)}
	sort.SliceStable(rep.Networks, func(a, b int) bool { return weight(rep.Networks[a]) > weight(rep.Networks[b]) })
	for i := range rep.Networks {
		if rep.Networks[i].Connected {
			rep.Current = &rep.Networks[i]
			break
		}
	}
	if len(nets) == 0 {
		rep.Advice = []string{T("scan.none")}
		return rep
	}
	own := func(n model.WiFiNetwork) bool {
		return rep.Current != nil && (n.BSSID == rep.Current.BSSID || (n.SSID != "" && n.SSID == rep.Current.SSID))
	}
	score := func(band string, c int) (float64, int) {
		s, count := 0.0, 0
		for _, n := range nets {
			if n.Band != band || own(n) {
				continue
			}
			if n.Channel == c {
				count++
			}
			if band == "2.4 GHz" {
				s += weight(n) * overlap(c, n.Channel)
			} else if n.Channel == c {
				s += weight(n)
			}
		}
		return s, count
	}
	seen := map[string]map[int]bool{}
	for _, n := range nets {
		if n.Band == "" || n.Channel == 0 {
			continue
		}
		if seen[n.Band] == nil {
			seen[n.Band] = map[int]bool{}
		}
		seen[n.Band][n.Channel] = true
	}
	bestOf := map[string]int{}
	for _, band := range []string{"2.4 GHz", "5 GHz", "6 GHz"} {
		chans := map[int]bool{}
		for c := range seen[band] {
			chans[c] = true
		}
		var cands []int
		switch band {
		case "2.4 GHz":
			cands = candidates24
		case "5 GHz":
			cands = candidates5
		}
		if len(seen[band]) == 0 && band != "2.4 GHz" && band != "5 GHz" {
			continue
		}
		for _, c := range cands {
			chans[c] = true
		}
		if band == "6 GHz" && len(chans) == 0 {
			continue
		}
		list := make([]int, 0, len(chans))
		for c := range chans {
			list = append(list, c)
		}
		sort.Ints(list)
		best, bestScore := 0, math.Inf(1)
		for _, c := range list {
			s, count := score(band, c)
			cl := ChannelLoad{Band: band, Channel: c, Networks: count, Score: math.Round(s*100) / 100}
			if rep.Current != nil && rep.Current.Band == band && rep.Current.Channel == c {
				cl.Current = true
			}
			if len(seen[band]) > 0 || cl.Current {
				rep.Channels = append(rep.Channels, cl)
			}
			isCand := false
			for _, x := range cands {
				isCand = isCand || x == c
			}
			if isCand && s < bestScore {
				best, bestScore = c, s
			}
		}
		if best != 0 && len(seen[band]) > 0 {
			bestOf[band] = best
		}
	}

	if cur := rep.Current; cur != nil && cur.Channel > 0 {
		curScore, curCount := score(cur.Band, cur.Channel)
		if best, ok := bestOf[cur.Band]; ok && best != cur.Channel {
			bestScore, _ := score(cur.Band, best)
			if curScore > bestScore+0.5 && curScore > 1.3*bestScore {
				rep.Recommended = best
				rep.Advice = append(rep.Advice, T("scan.recommend", cur.Channel, cur.Band, curCount, best))
			}
		}
		if rep.Recommended == 0 {
			rep.Advice = append(rep.Advice, T("scan.fine", cur.Channel, cur.Band))
		}
		if cur.Band == "2.4 GHz" {
			if n24 := countBand(nets, "2.4 GHz"); n24 >= 8 {
				rep.Advice = append(rep.Advice, T("scan.try5", n24))
			}
		}
	} else {
		if b, ok := bestOf["2.4 GHz"]; ok {
			rep.Advice = append(rep.Advice, T("scan.best24", b))
		}
		if b, ok := bestOf["5 GHz"]; ok {
			rep.Advice = append(rep.Advice, T("scan.best5", b))
		}
	}
	for i := range rep.Channels {
		c := &rep.Channels[i]
		c.Recommended = rep.Recommended != 0 && rep.Current != nil && c.Band == rep.Current.Band && c.Channel == rep.Recommended
	}
	return rep
}

func countBand(nets []model.WiFiNetwork, band string) int {
	n := 0
	for _, x := range nets {
		if x.Band == band {
			n++
		}
	}
	return n
}
