package visualbs

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"
)

// startTestServer launches a server on a random port and returns its URL.
// Cleanup is registered with t.Cleanup so each test isolates its goroutines.
func startTestServer(t *testing.T) (string, *Server, <-chan Event) {
	t.Helper()
	srv := NewServer(0) // 0 = OS-assigned free port
	srv.idleTimeout = 200 * time.Millisecond
	ctx, cancel := context.WithCancel(context.Background())
	url, events, err := srv.Start(ctx)
	if err != nil {
		t.Fatalf("Start: %v", err)
	}
	t.Cleanup(func() {
		cancel()
		srv.Stop()
	})
	return url, srv, events
}

func TestStopIsIdempotent(t *testing.T) {
	_, srv, _ := startTestServer(t)

	// Calling Stop concurrently from many goroutines must not panic or block.
	done := make(chan struct{}, 10)
	for i := 0; i < 10; i++ {
		go func() {
			srv.Stop()
			done <- struct{}{}
		}()
	}
	for i := 0; i < 10; i++ {
		select {
		case <-done:
		case <-time.After(2 * time.Second):
			t.Fatal("Stop blocked on concurrent invocation — sync.Once guard missing")
		}
	}
}

func TestPushBroadcastsToEventsEndpoint_NoOpOnEmptyClients(t *testing.T) {
	url, _, _ := startTestServer(t)

	// /push with no SSE clients connected must still succeed (200), not error.
	body := bytes.NewBufferString(`{"type":"html","content":"<div>hi</div>"}`)
	resp, err := http.Post(url+"/push", "application/json", body)
	if err != nil {
		t.Fatalf("POST /push: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
}

func TestPushRejectsMalformedJSON(t *testing.T) {
	url, _, _ := startTestServer(t)

	resp, err := http.Post(url+"/push", "application/json", strings.NewReader("not-json"))
	if err != nil {
		t.Fatalf("POST /push: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400 for malformed JSON, got %d", resp.StatusCode)
	}
}

func TestChoiceEnqueuesUserEvent(t *testing.T) {
	url, _, events := startTestServer(t)

	choice := Event{Type: "click", ID: "option-a", Selected: true}
	payload, _ := json.Marshal(choice)
	resp, err := http.Post(url+"/choice", "application/json", bytes.NewReader(payload))
	if err != nil {
		t.Fatalf("POST /choice: %v", err)
	}
	resp.Body.Close()

	select {
	case e := <-events:
		if e.ID != "option-a" {
			t.Errorf("expected ID=option-a, got %q", e.ID)
		}
	case <-time.After(time.Second):
		t.Fatal("event was not enqueued on /choice POST")
	}
}

func TestEventsEndpointReturnsBufferedEvents(t *testing.T) {
	url, _, _ := startTestServer(t)

	// Push a choice
	payload, _ := json.Marshal(Event{Type: "click", ID: "x"})
	http.Post(url+"/choice", "application/json", bytes.NewReader(payload))

	// Drain via /events — the buffered channel goes through /choice into s.events,
	// /events pulls from s.events (not from SSE broadcasts). The handler is
	// non-blocking on the channel so we have to give it a beat.
	time.Sleep(50 * time.Millisecond)

	resp, err := http.Get(url + "/events")
	if err != nil {
		t.Fatalf("GET /events: %v", err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	var got []Event
	if err := json.Unmarshal(body, &got); err != nil {
		t.Fatalf("decode events response: %v (body=%s)", err, body)
	}
	if len(got) != 1 || got[0].ID != "x" {
		t.Errorf("expected 1 event with ID=x, got %+v", got)
	}
}
