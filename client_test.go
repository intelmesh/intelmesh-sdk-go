package intelmesh_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	intelmesh "github.com/intelmesh/intelmesh-sdk-go"
)

func TestNew_DefaultTimeout(t *testing.T) {
	t.Parallel()

	c := intelmesh.New(intelmesh.Config{
		BaseURL: "https://api.example.com",
		APIKey:  "test-key",
	})

	if c == nil {
		t.Fatal("expected non-nil client")
	}

	if c.Events == nil {
		t.Fatal("expected Events resource to be initialized")
	}

	if c.Rules == nil {
		t.Fatal("expected Rules resource to be initialized")
	}
}

func TestClient_ErrorMapping_NotFound(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)

		resp := intelmesh.Envelope[any]{
			Error: &intelmesh.APIError{
				Code:    intelmesh.CodeNotFound,
				Message: "resource not found",
			},
		}

		json.NewEncoder(w).Encode(resp)
	}))
	defer srv.Close()

	c := intelmesh.New(intelmesh.Config{
		BaseURL: srv.URL,
		APIKey:  "test-key",
	})

	_, err := c.Phases.Get(context.Background(), "nonexistent")
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !intelmesh.IsNotFound(err) {
		t.Fatalf("expected IsNotFound=true, got error: %v", err)
	}
}

func TestClient_ErrorMapping_Unauthorized(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)

		resp := intelmesh.Envelope[any]{
			Error: &intelmesh.APIError{
				Code:    intelmesh.CodeUnauthorized,
				Message: "invalid api key",
			},
		}

		json.NewEncoder(w).Encode(resp)
	}))
	defer srv.Close()

	c := intelmesh.New(intelmesh.Config{
		BaseURL: srv.URL,
		APIKey:  "bad-key",
	})

	_, err := c.Phases.Get(context.Background(), "some-id")
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !intelmesh.IsUnauthorized(err) {
		t.Fatalf("expected IsUnauthorized=true, got error: %v", err)
	}
}

func TestClient_ErrorMapping_Validation(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)

		resp := intelmesh.Envelope[any]{
			Error: &intelmesh.APIError{
				Code:    intelmesh.CodeValidation,
				Message: "name is required",
			},
		}

		json.NewEncoder(w).Encode(resp)
	}))
	defer srv.Close()

	c := intelmesh.New(intelmesh.Config{
		BaseURL: srv.URL,
		APIKey:  "test-key",
	})

	_, err := c.Phases.Create(context.Background(), intelmesh.CreatePhaseRequest{})
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !intelmesh.IsValidation(err) {
		t.Fatalf("expected IsValidation=true, got error: %v", err)
	}
}

func TestClient_ErrorMapping_Forbidden(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)

		resp := intelmesh.Envelope[any]{
			Error: &intelmesh.APIError{
				Code:    intelmesh.CodeForbidden,
				Message: "insufficient permissions",
			},
		}

		json.NewEncoder(w).Encode(resp)
	}))
	defer srv.Close()

	c := intelmesh.New(intelmesh.Config{
		BaseURL: srv.URL,
		APIKey:  "test-key",
	})

	_, err := c.Phases.Get(context.Background(), "some-id")
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !intelmesh.IsForbidden(err) {
		t.Fatalf("expected IsForbidden=true, got error: %v", err)
	}
}

func TestClient_AuthHeader(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if auth != "Bearer my-secret-key" {
			t.Errorf("expected Authorization header 'Bearer my-secret-key', got '%s'", auth)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		resp := intelmesh.Envelope[json.RawMessage]{
			Data: json.RawMessage(`{"id":"p1","name":"test","position":1,"created_at":"2025-01-01T00:00:00Z"}`),
		}

		json.NewEncoder(w).Encode(resp)
	}))
	defer srv.Close()

	c := intelmesh.New(intelmesh.Config{
		BaseURL: srv.URL,
		APIKey:  "my-secret-key",
	})

	_, err := c.Phases.Get(context.Background(), "p1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestIntelMeshError_ErrorString(t *testing.T) {
	t.Parallel()

	err := &intelmesh.IntelMeshError{
		StatusCode: 404,
		Code:       intelmesh.CodeNotFound,
		Message:    "phase not found",
	}

	expected := "intelmesh: NOT_FOUND (HTTP 404): phase not found"
	if err.Error() != expected {
		t.Fatalf("expected %q, got %q", expected, err.Error())
	}
}
