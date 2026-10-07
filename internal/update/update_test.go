package update

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewer(t *testing.T) {
	cases := []struct {
		a, b string
		want bool
	}{
		{"1.1.0", "1.0.0", true},
		{"1.0.0", "1.0.0", false},
		{"1.0.0", "1.1.0", false},
		{"2.0.0", "1.9.9", true},
		{"1.10.0", "1.9.0", true},
		{"1.2.0", "1.2.0-beta.1", true},
		{"1.2.0-beta.1", "1.2.0", false},
		{"1.0.0", "dev", true},
		{"1.0.0", "0.0.0-abc1234", true},
		{"garbage", "1.0.0", false},
	}
	for _, c := range cases {
		if got := Newer(c.a, c.b); got != c.want {
			t.Errorf("Newer(%q, %q) = %v, want %v", c.a, c.b, got, c.want)
		}
	}
}

func TestCheck(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/repos/"+Repo+"/releases/latest" || r.Header.Get("User-Agent") == "" {
			http.NotFound(w, r)
			return
		}
		w.Write([]byte(`{"tag_name":"v1.1.0","html_url":"https://github.com/` + Repo + `/releases/tag/v1.1.0"}`))
	}))
	defer srv.Close()
	info, err := Check(context.Background(), srv.Client(), srv.URL, "1.0.0")
	if err != nil {
		t.Fatal(err)
	}
	if !info.Newer || info.Latest != "1.1.0" || info.URL != "https://github.com/"+Repo+"/releases/tag/v1.1.0" {
		t.Fatalf("info = %+v", info)
	}
	if info, _ := Check(context.Background(), srv.Client(), srv.URL, "1.1.0"); info.Newer {
		t.Fatal("same version must not be newer")
	}
}
