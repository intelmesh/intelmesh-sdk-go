// Package builders provides fluent builder APIs for constructing IntelMesh request types.
package builders

import intelmesh "github.com/intelmesh/intelmesh-sdk-go"

// EventBuilder builds IngestRequest payloads fluently.
type EventBuilder struct {
	eventType      string
	payload        map[string]any
	idempotencyKey string
}

// Event creates a new EventBuilder for the given event type.
func Event(eventType string) *EventBuilder {
	return &EventBuilder{
		eventType: eventType,
		payload:   make(map[string]any),
	}
}

// Set adds a key-value pair to the event payload.
func (b *EventBuilder) Set(key string, value any) *EventBuilder {
	b.payload[key] = value

	return b
}

// IdempotencyKey sets the idempotency key for the event.
func (b *EventBuilder) IdempotencyKey(key string) *EventBuilder {
	b.idempotencyKey = key

	return b
}

// Build constructs the final IngestRequest.
func (b *EventBuilder) Build() intelmesh.IngestRequest {
	return intelmesh.IngestRequest{
		EventType:      b.eventType,
		Payload:        b.payload,
		IdempotencyKey: b.idempotencyKey,
	}
}
