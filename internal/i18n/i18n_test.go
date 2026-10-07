package i18n

import (
	"regexp"
	"testing"
)

var verb = regexp.MustCompile(`%[-+# 0-9.]*[a-zA-Z%]`)

// Every message needs both languages with the same format verbs, or a
// translated message would print garbage like %!d(MISSING).
func TestCatalogComplete(t *testing.T) {
	for key, m := range catalog {
		if m[0] == "" || m[1] == "" {
			t.Errorf("%s: missing translation", key)
			continue
		}
		en, id := verb.FindAllString(m[0], -1), verb.FindAllString(m[1], -1)
		if len(en) != len(id) {
			t.Errorf("%s: verbs differ: %v vs %v", key, en, id)
			continue
		}
		for i := range en {
			if en[i] != id[i] {
				t.Errorf("%s: verb %d differs: %s vs %s", key, i, en[i], id[i])
			}
		}
	}
}

func TestParseAndT(t *testing.T) {
	if Parse("id-ID") != ID || Parse("in") != ID || Parse("en-US") != EN || Parse("") != EN || Parse("de") != EN {
		t.Fatal("Parse")
	}
	if T(ID, "check.gatewayOK", "1.2.3.4") != "1.2.3.4 terjangkau" {
		t.Fatal(T(ID, "check.gatewayOK", "1.2.3.4"))
	}
	if T(EN, "no.such.key") != "no.such.key" {
		t.Fatal("unknown keys should be visible")
	}
}
