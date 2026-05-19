package visualbs

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"sync"
	"time"
)

//go:embed template.html
var indexHTML string

// Event represents a user interaction from the browser.
type Event struct {
	Type       string      `json:"type"`
	ID         string      `json:"id,omitempty"`
	Selected   bool        `json:"selected,omitempty"`
	Text       string      `json:"text,omitempty"`
	Selections []Selection `json:"selections,omitempty"`
}

// Selection represents a selected option.
type Selection struct {
	ID   string `json:"id"`
	Text string `json:"text"`
}

// Server is a lightweight HTTP server for visual brainstorming.
// Uses SSE for server→browser push and POST for browser→server events.
type Server struct {
	port        int
	events      chan Event
	sseClients  map[chan []byte]bool
	mu          sync.RWMutex
	srv         *http.Server
	idleTimer   *time.Timer
	idleTimeout time.Duration
	stopOnce    sync.Once
}

// NewServer creates a visual brainstorming server.
func NewServer(port int) *Server {
	return &Server{
		port:        port,
		events:      make(chan Event, 100),
		sseClients:  make(map[chan []byte]bool),
		idleTimeout: 30 * time.Minute,
	}
}

// Start begins serving. Returns the URL and a channel of user events.
func (s *Server) Start(ctx context.Context) (string, <-chan Event, error) {
	mux := http.NewServeMux()

	// Serve the HTML page
	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprint(w, indexHTML)
	})

	// SSE endpoint — browser subscribes for server-pushed content
	mux.HandleFunc("GET /sse", s.handleSSE)

	// Push endpoint — agent sends HTML content to display
	mux.HandleFunc("POST /push", func(w http.ResponseWriter, r *http.Request) {
		s.resetIdle()

		var msg struct {
			Type    string `json:"type"`    // "html", "append", "clear"
			Content string `json:"content"` // HTML fragment
		}
		if err := json.NewDecoder(r.Body).Decode(&msg); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		s.broadcast(msg.Type, msg.Content)
		w.WriteHeader(http.StatusOK)
	})

	// Events endpoint — agent polls for user interactions
	mux.HandleFunc("GET /events", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		var events []Event
		for {
			select {
			case e := <-s.events:
				events = append(events, e)
			default:
				goto done
			}
		}
	done:
		json.NewEncoder(w).Encode(events)
	})

	// Choice endpoint — browser sends user clicks
	mux.HandleFunc("POST /choice", func(w http.ResponseWriter, r *http.Request) {
		s.resetIdle()
		var evt Event
		if err := json.NewDecoder(r.Body).Decode(&evt); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		select {
		case s.events <- evt:
		default:
			// Drop if buffer full
		}
		w.WriteHeader(http.StatusOK)
	})

	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", s.port))
	if err != nil {
		return "", nil, fmt.Errorf("listen: %w", err)
	}

	s.port = listener.Addr().(*net.TCPAddr).Port
	s.srv = &http.Server{
		Handler: mux,
		// Bound header read so slow-loris clients can't tie up goroutines.
		// Body read and response write are unbounded because /sse is intentionally
		// long-lived; per-handler timeouts would be a larger refactor.
		ReadHeaderTimeout: 5 * time.Second,
	}

	s.idleTimer = time.AfterFunc(s.idleTimeout, func() {
		log.Printf("visualbs: idle timeout reached, shutting down")
		s.Stop()
	})

	go func() {
		if err := s.srv.Serve(listener); err != http.ErrServerClosed {
			log.Printf("visualbs: server error: %v", err)
		}
	}()

	go func() {
		<-ctx.Done()
		s.Stop()
	}()

	url := fmt.Sprintf("http://localhost:%d", s.port)
	return url, s.events, nil
}

// Stop shuts down the server. Safe to call multiple times from multiple
// goroutines (idle timer + ctx.Done watcher both race to call it).
func (s *Server) Stop() {
	s.stopOnce.Do(func() {
		if s.idleTimer != nil {
			s.idleTimer.Stop()
		}
		if s.srv != nil {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			s.srv.Shutdown(ctx)
		}
	})
}

func (s *Server) resetIdle() {
	if s.idleTimer != nil {
		s.idleTimer.Reset(s.idleTimeout)
	}
}

func (s *Server) handleSSE(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "SSE not supported", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	flusher.Flush()

	ch := make(chan []byte, 16)
	s.mu.Lock()
	s.sseClients[ch] = true
	s.mu.Unlock()

	defer func() {
		s.mu.Lock()
		delete(s.sseClients, ch)
		s.mu.Unlock()
	}()

	for {
		select {
		case data := <-ch:
			fmt.Fprintf(w, "data: %s\n\n", data)
			flusher.Flush()
		case <-r.Context().Done():
			return
		}
	}
}

func (s *Server) broadcast(msgType, content string) {
	data, _ := json.Marshal(map[string]string{
		"type":    msgType,
		"content": content,
	})

	s.mu.RLock()
	defer s.mu.RUnlock()
	for ch := range s.sseClients {
		select {
		case ch <- data:
		default:
			// Drop if client is slow
		}
	}
}
