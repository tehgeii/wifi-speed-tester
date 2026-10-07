package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultIsValid(t *testing.T) {
	if err := Default().Validate(); err != nil {
		t.Fatal(err)
	}
}

func TestLoadMergesOverDefaults(t *testing.T) {
	dir := t.TempDir()
	if _, path, err := Load(dir); err != nil || path != "" {
		t.Fatalf("missing file: path=%q err=%v", path, err)
	}
	data := `{"pingCount": 50, "servers": [{"name":"Mine","downloadUrl":"https://x/d?b={bytes}","uploadUrl":"https://x/u"}],
	          "profiles": {"gaming": {"use":["ping"],"pingMs":{"excellent":10,"good":20,"fair":30}}}}`
	if err := os.WriteFile(filepath.Join(dir, FileName), []byte(data), 0o644); err != nil {
		t.Fatal(err)
	}
	cfg, path, err := Load(dir)
	if err != nil || path == "" {
		t.Fatalf("load: %v", err)
	}
	if cfg.PingCount != 50 || cfg.Streams != 4 || len(cfg.Servers) != 1 || cfg.Servers[0].Name != "Mine" {
		t.Fatalf("merge failed: %+v", cfg)
	}
	if cfg.Profile("gaming").PingMs.Excellent != 10 || cfg.Profile("general").PingMs.Excellent != 20 {
		t.Fatal("profile merge failed")
	}
}

func TestInvalidConfigFallsBack(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, FileName), []byte(`{"streams": 0}`), 0o644)
	cfg, _, err := Load(dir)
	if err == nil {
		t.Fatal("expected validation error")
	}
	if cfg.Streams != 4 {
		t.Fatal("defaults should be used after an invalid config")
	}
	os.WriteFile(filepath.Join(dir, FileName), []byte(`{not json`), 0o644)
	if _, _, err := Load(dir); err == nil {
		t.Fatal("expected parse error")
	}
}

func TestLibreSpeedServer(t *testing.T) {
	s := Server{Name: "L", Type: LibreSpeed, URL: "https://host/backend"}
	if err := s.Validate(); err != nil {
		t.Fatal(err)
	}
	if got := s.DownloadFor(25_000_000); got != "https://host/backend/garbage.php?ckSize=24" {
		t.Fatalf("download = %s", got)
	}
	if got := s.DownloadFor(0); got != "https://host/backend/garbage.php?ckSize=1" {
		t.Fatalf("download(0) = %s", got)
	}
	if got := s.DownloadFor(5 << 30); got != "https://host/backend/garbage.php?ckSize=1024" {
		t.Fatalf("download cap = %s", got)
	}
	if s.Upload() != "https://host/backend/empty.php" || s.Ping() != "https://host/backend/empty.php" {
		t.Fatal("upload/ping urls")
	}
	s.DLPath, s.ULPath = "dl.php", "ul.php"
	if s.DownloadFor(1) != "https://host/backend/dl.php?ckSize=1" || s.Upload() != "https://host/backend/ul.php" {
		t.Fatal("custom paths")
	}
}

func TestServerValidate(t *testing.T) {
	bad := []Server{
		{Name: "a", Type: LibreSpeed},
		{Name: "b", Type: LibreSpeed, URL: "ftp://x/"},
		{Name: "c", DownloadURL: "https://x/d", UploadURL: "https://x/u"},
		{Name: "d", DownloadURL: "https://x/d?b={bytes}"},
		{Name: "e", Type: "other", URL: "https://x/"},
	}
	for _, s := range bad {
		if s.Validate() == nil {
			t.Errorf("%s should be invalid", s.Name)
		}
	}
	if err := Default().Servers[0].Validate(); err != nil {
		t.Fatal(err)
	}
}
