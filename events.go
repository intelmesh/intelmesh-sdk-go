package intelmesh

import "context"

// EventsResource provides methods for event operations.
type EventsResource struct {
	c *Client
}

// Ingest sends an event for synchronous evaluation.
func (r *EventsResource) Ingest(ctx context.Context, req IngestRequest) (*IngestResult, error) {
	var result IngestResult
	if err := r.c.doJSON(ctx, "POST", "/api/v1/events", req, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// IngestAsync sends an event for asynchronous evaluation (202 Accepted).
func (r *EventsResource) IngestAsync(ctx context.Context, req IngestRequest) error {
	return r.c.doNoContent(ctx, "POST", "/api/v1/events?async=true", req)
}

// IngestOnly sends an event without evaluation (ingest-only mode).
func (r *EventsResource) IngestOnly(ctx context.Context, req IngestRequest) (*IngestResult, error) {
	var result IngestResult
	if err := r.c.doJSON(ctx, "POST", "/api/v1/events?evaluate=false", req, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// Simulate runs the pipeline without persisting or triggering side effects.
func (r *EventsResource) Simulate(ctx context.Context, req IngestRequest) (*IngestResult, error) {
	var result IngestResult
	if err := r.c.doJSON(ctx, "POST", "/api/v1/events/simulate", req, &result); err != nil {
		return nil, err
	}

	return &result, nil
}
