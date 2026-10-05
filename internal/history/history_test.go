package history

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/tehgeii/wifi-speed-tester/internal/model"
)

func TestStore(t *testing.T) {
	p := filepath.Join(t.TempDir(), "sub", "history.json")
	s := Open(p, 2)
	if l, err := s.List(); err != nil || len(l) != 0 {
		t.Fatalf("empty list: %v %v", l, err)
	}
	for _, id := range []string{"a", "b", "c"} {
		if err := s.Add(&model.TestResult{ID: id}); err != nil {
			t.Fatal(err)
		}
	}
	l, _ := s.List()
	if len(l) != 2 || l[0].ID != "c" || l[1].ID != "b" {
		t.Fatalf("list = %v", l)
	}
	if r, _ := s.Get("b"); r == nil {
		t.Fatal("get b")
	}
	if err := s.Clear(); err != nil {
		t.Fatal(err)
	}
	if l, _ := s.List(); len(l) != 0 {
		t.Fatal("clear failed")
	}
}

func TestCorruptFileIsReplaced(t *testing.T) {
	p := filepath.Join(t.TempDir(), "history.json")
	os.WriteFile(p, []byte("garbage"), 0o644)
	s := Open(p, 10)
	if _, err := s.List(); err == nil {
		t.Fatal("expected error reading corrupt file")
	}
	if err := s.Add(&model.TestResult{ID: "x"}); err != nil {
		t.Fatal(err)
	}
	if l, err := s.List(); err != nil || len(l) != 1 {
		t.Fatalf("after add: %v %v", l, err)
	}
}
