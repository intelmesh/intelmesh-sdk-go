package intelmesh

import (
	"context"
	"net/http"
	"net/url"
)

// ScoresResource provides methods for score operations.
type ScoresResource struct {
	c *Client
}

// Get retrieves a score by scope name and scope value.
func (r *ScoresResource) Get(ctx context.Context, scopeName string, scopeValue string) (*Score, error) {
	path := "/api/v1/scores?" + url.Values{
		"scope_name":  {scopeName},
		"scope_value": {scopeValue},
	}.Encode()

	var result Score
	if err := r.c.doJSON(ctx, http.MethodGet, path, nil, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// List returns a paginated list of scores.
func (r *ScoresResource) List(ctx context.Context, params ListParams) (*PaginatedResponse[Score], error) {
	var result PaginatedResponse[Score]
	if err := r.c.doPaginated(ctx, "/api/v1/scores", params, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// Set sets a score to a specific value.
func (r *ScoresResource) Set(ctx context.Context, req SetScoreRequest) (*Score, error) {
	var result Score
	if err := r.c.doJSON(ctx, http.MethodPut, "/api/v1/scores", req, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// Reset resets a score to zero by scope name and scope value.
func (r *ScoresResource) Reset(ctx context.Context, scopeName string, scopeValue string) error {
	path := "/api/v1/scores?" + url.Values{
		"scope_name":  {scopeName},
		"scope_value": {scopeValue},
	}.Encode()

	return r.c.doNoContent(ctx, http.MethodDelete, path, nil)
}
