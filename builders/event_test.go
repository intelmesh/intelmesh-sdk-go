package builders_test

import (
	"testing"

	"github.com/intelmesh/intelmesh-sdk-go/builders"
)

func TestEventBuilder_Basic(t *testing.T) {
	t.Parallel()

	req := builders.Event("transaction").
		Set("amount", 100.50).
		Set("currency", "USD").
		Build()

	if req.EventType != "transaction" {
		t.Errorf("expected event_type 'transaction', got '%s'", req.EventType)
	}

	if req.Payload["amount"] != 100.50 {
		t.Errorf("expected payload amount 100.50, got %v", req.Payload["amount"])
	}

	if req.Payload["currency"] != "USD" {
		t.Errorf("expected payload currency 'USD', got %v", req.Payload["currency"])
	}

	if req.IdempotencyKey != "" {
		t.Errorf("expected empty idempotency key, got '%s'", req.IdempotencyKey)
	}
}

func TestEventBuilder_WithIdempotencyKey(t *testing.T) {
	t.Parallel()

	req := builders.Event("login").
		Set("ip", "192.168.1.1").
		IdempotencyKey("unique-key-123").
		Build()

	if req.EventType != "login" {
		t.Errorf("expected event_type 'login', got '%s'", req.EventType)
	}

	if req.IdempotencyKey != "unique-key-123" {
		t.Errorf("expected idempotency key 'unique-key-123', got '%s'", req.IdempotencyKey)
	}
}

func TestEventBuilder_OverwriteKey(t *testing.T) {
	t.Parallel()

	req := builders.Event("transaction").
		Set("amount", 50.00).
		Set("amount", 200.00).
		Build()

	if req.Payload["amount"] != 200.00 {
		t.Errorf("expected overwritten amount 200.00, got %v", req.Payload["amount"])
	}
}

func TestEventBuilder_EmptyPayload(t *testing.T) {
	t.Parallel()

	req := builders.Event("heartbeat").Build()

	if req.EventType != "heartbeat" {
		t.Errorf("expected event_type 'heartbeat', got '%s'", req.EventType)
	}

	if len(req.Payload) != 0 {
		t.Errorf("expected empty payload, got %v", req.Payload)
	}
}
