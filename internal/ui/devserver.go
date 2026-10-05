package ui

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"sync"

	"github.com/tehgeii/wifi-speed-tester/internal/app"
)

// DevServer serves the UI over HTTP with the same API the desktop window
// exposes. It is meant for development and UI testing, and only listens on
// the loopback interface.
type DevServer struct {
	mu      sync.Mutex
	clients map[chan []byte]struct{}
	App     *app.App
}

// NewDevServer returns a server; call Emit as the app's Host.Emit.
func NewDevServer() *DevServer {
	return &DevServer{clients: map[chan []byte]struct{}{}}
}

// Emit broadcasts an event to every connected page.
func (s *DevServer) Emit(name string, payload any) {
	b, err := json.Marshal(map[string]any{"name": name, "payload": payload})
	if err != nil {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	for c := range s.clients {
		select {
		case c <- b:
		default: // slow client; drop rather than block the engine
		}
	}
}

// Handler returns the HTTP handler.
func (s *DevServer) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.Handle("/", http.FileServerFS(Assets()))
	mux.HandleFunc("POST /api/{method}", func(w http.ResponseWriter, r *http.Request) {
		var params []json.RawMessage
		if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
			params = nil
		}
		res, err := s.App.Call(r.PathValue("method"), params)
		out := map[string]any{"result": res}
		if err != nil {
			out = map[string]any{"error": err.Error()}
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(out)
	})
	mux.HandleFunc("GET /events", func(w http.ResponseWriter, r *http.Request) {
		fl, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, "streaming unsupported", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		c := make(chan []byte, 256)
		s.mu.Lock()
		s.clients[c] = struct{}{}
		s.mu.Unlock()
		defer func() {
			s.mu.Lock()
			delete(s.clients, c)
			s.mu.Unlock()
		}()
		fmt.Fprint(w, ": connected\n\n")
		fl.Flush()
		for {
			select {
			case <-r.Context().Done():
				return
			case b := <-c:
				fmt.Fprintf(w, "data: %s\n\n", b)
				fl.Flush()
			}
		}
	})
	return mux
}

// ListenLoopback starts listening on 127.0.0.1:port.
func ListenLoopback(port int) (net.Listener, error) {
	return net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", port))
}
