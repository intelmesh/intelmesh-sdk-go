package intelmesh_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	intelmesh "github.com/intelmesh/intelmesh-sdk-go"
)

func TestEventsResource_Ingest(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}

		if r.URL.Path != "/api/v1/events" {
			t.Errorf("expected path /api/v1/events, got %s", r.URL.Path)
		}

		var req intelmesh.IngestRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("failed to decode request body: %v", err)
		}

		if req.EventType != "transaction" {
			t.Errorf("expected event_type 'transaction', got '%s'", req.EventType)
		}

		result := intelmesh.IngestResult{
			EventID:        "evt-123",
			TransientScore: 42,
			DurationMs:     15,
			Decision: &intelmesh.Decision{
				Action:   "allow",
				Severity: "low",
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(intelmesh.Envelope[intelmesh.IngestResult]{
			Data: result,
		})
	}))
	defer srv.Close()

	c := intelmesh.New(intelmesh.Config{
		BaseURL: srv.URL,
		APIKey:  "test-key",
	})

	result, err := c.Events.Ingest(context.Background(), intelmesh.IngestRequest{
		EventType: "transaction",
		Payload:   map[string]any{"amount": 100.50},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.EventID != "evt-123" {
		t.Errorf("expected event_id 'evt-123', got '%s'", result.EventID)
	}

	if result.TransientScore != 42 {
		t.Errorf("expected transient_score 42, got %d", result.TransientScore)
	}

	if result.Decision == nil {
		t.Fatal("expected decision, got nil")
	}

	if result.Decision.Action != "allow" {
		t.Errorf("expected decision action 'allow', got '%s'", result.Decision.Action)
	}
}

func TestEventsResource_IngestAsync(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}

		if r.URL.Query().Get("async") != "true" {
			t.Error("expected async=true query parameter")
		}

		w.WriteHeader(http.StatusAccepted)
	}))
	defer srv.Close()

	c := intelmesh.New(intelmesh.Config{
		BaseURL: srv.URL,
		APIKey:  "test-key",
	})

	err := c.Events.IngestAsync(context.Background(), intelmesh.IngestRequest{
		EventType: "transaction",
		Payload:   map[string]any{"amount": 50.00},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestEventsResource_Simulate(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}

		if r.URL.Path != "/api/v1/events/simulate" {
			t.Errorf("expected path /api/v1/events/simulate, got %s", r.URL.Path)
		}

		result := intelmesh.IngestResult{
			EventID:        "evt-sim-001",
			TransientScore: 85,
			DurationMs:     3,
			Decision: &intelmesh.Decision{
				Action:   "block",
				Severity: "high",
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(intelmesh.Envelope[intelmesh.IngestResult]{
			Data: result,
		})
	}))
	defer srv.Close()

	c := intelmesh.New(intelmesh.Config{
		BaseURL: srv.URL,
		APIKey:  "test-key",
	})

	result, err := c.Events.Simulate(context.Background(), intelmesh.IngestRequest{
		EventType: "login",
		Payload:   map[string]any{"ip": "1.2.3.4"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.EventID != "evt-sim-001" {
		t.Errorf("expected event_id 'evt-sim-001', got '%s'", result.EventID)
	}

	if result.Decision.Action != "block" {
		t.Errorf("expected decision action 'block', got '%s'", result.Decision.Action)
	}
}

func TestEventsResource_IngestOnly(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("evaluate") != "false" {
			t.Error("expected evaluate=false query parameter")
		}

		result := intelmesh.IngestResult{
			EventID:    "evt-ingest-001",
			DurationMs: 2,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(intelmesh.Envelope[intelmesh.IngestResult]{
			Data: result,
		})
	}))
	defer srv.Close()

	c := intelmesh.New(intelmesh.Config{
		BaseURL: srv.URL,
		APIKey:  "test-key",
	})

	result, err := c.Events.IngestOnly(context.Background(), intelmesh.IngestRequest{
		EventType: "pageview",
		Payload:   map[string]any{"url": "/checkout"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.EventID != "evt-ingest-001" {
		t.Errorf("expected event_id 'evt-ingest-001', got '%s'", result.EventID)
	}
}
