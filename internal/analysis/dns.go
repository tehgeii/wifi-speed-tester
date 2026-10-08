package analysis

import (
	"github.com/tehgeii/wifi-speed-tester/internal/i18n"
	"github.com/tehgeii/wifi-speed-tester/internal/model"
)

// DNSAdvice turns DNS test results (fastest first) into a recommendation.
func DNSAdvice(results []model.DNSResult, lang i18n.Lang) []string {
	T := func(key string, args ...any) string { return i18n.T(lang, key, args...) }
	if len(results) == 0 || results[0].OK == 0 {
		return []string{T("dns.allFailed")}
	}
	best := results[0]
	var sys *model.DNSResult
	for i := range results {
		r := &results[i]
		if r.System && r.OK > 0 && (sys == nil || r.MedianMs < sys.MedianMs) {
			sys = r
		}
	}
	if sys == nil {
		// The configured DNS answered nothing at all.
		for i := range results {
			if results[i].System {
				n := results[i].OK + results[i].Failed
				return []string{T("dns.systemFails", results[i].Failed, n, best.Label), T("dns.howto")}
			}
		}
		return nil
	}
	if n := sys.OK + sys.Failed; n > 0 && float64(sys.Failed)/float64(n) > 0.2 && !best.System {
		return []string{T("dns.systemFails", sys.Failed, n, best.Label), T("dns.howto")}
	}
	if !best.System && best.MedianMs < sys.MedianMs*0.7 && sys.MedianMs-best.MedianMs > 15 {
		return []string{T("dns.switch", best.Label, FormatMs(best.MedianMs), FormatMs(sys.MedianMs)), T("dns.howto")}
	}
	return []string{T("dns.keep")}
}
