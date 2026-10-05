// Package history keeps past results in a local JSON file. Nothing leaves
// the machine.
package history

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sync"

	"github.com/tehgeii/wifi-speed-tester/internal/model"
)

// Store is a JSON-file backed list of results, newest first.
type Store struct {
	mu    sync.Mutex
	path  string
	limit int
}

// Open returns a store at path (the file is created on first save).
func Open(path string, limit int) *Store {
	if limit <= 0 {
		limit = 200
	}
	return &Store{path: path, limit: limit}
}

// Path is the backing file.
func (s *Store) Path() string { return s.path }

func (s *Store) load() ([]*model.TestResult, error) {
	data, err := os.ReadFile(s.path)
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var out []*model.TestResult
	if err := json.Unmarshal(data, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// List returns all stored results, newest first.
func (s *Store) List() ([]*model.TestResult, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.load()
}

// Get returns the result with the given ID, or nil.
func (s *Store) Get(id string) (*model.TestResult, error) {
	all, err := s.List()
	for _, r := range all {
		if r.ID == id {
			return r, err
		}
	}
	return nil, err
}

// Add prepends r, trims to the limit and writes the file atomically.
// A corrupt history file is replaced rather than blocking new results.
func (s *Store) Add(r *model.TestResult) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	all, _ := s.load()
	all = append([]*model.TestResult{r}, all...)
	if len(all) > s.limit {
		all = all[:s.limit]
	}
	return s.write(all)
}

// Clear removes every entry.
func (s *Store) Clear() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	err := os.Remove(s.path)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	return err
}

func (s *Store) write(all []*model.TestResult) error {
	if err := os.MkdirAll(filepath.Dir(s.path), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(all, "", " ")
	if err != nil {
		return err
	}
	tmp := s.path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, s.path)
}
