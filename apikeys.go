package intelmesh

import (
	"context"
	"net/http"
)

// APIKeysResource provides methods for API key operations.
type APIKeysResource struct {
	c *Client
}

// Create creates a new API key and returns it with the plain key (shown only once).
func (r *APIKeysResource) Create(ctx context.Context, req CreateAPIKeyRequest) (*CreateAPIKeyResult, error) {
	var result CreateAPIKeyResult
	if err := r.c.doJSON(ctx, http.MethodPost, "/api/v1/api-keys", req, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// Get retrieves an API key by ID.
func (r *APIKeysResource) Get(ctx context.Context, id string) (*APIKey, error) {
	var result APIKey
	if err := r.c.doJSON(ctx, http.MethodGet, "/api/v1/api-keys/"+id, nil, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// List returns a paginated list of API keys.
func (r *APIKeysResource) List(ctx context.Context, params ListParams) (*PaginatedResponse[APIKey], error) {
	var result PaginatedResponse[APIKey]
	if err := r.c.doPaginated(ctx, "/api/v1/api-keys", params, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// Update updates an existing API key.
func (r *APIKeysResource) Update(ctx context.Context, id string, req UpdateAPIKeyRequest) (*APIKey, error) {
	var result APIKey
	if err := r.c.doJSON(ctx, http.MethodPut, "/api/v1/api-keys/"+id, req, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// Delete deletes an API key by ID.
func (r *APIKeysResource) Delete(ctx context.Context, id string) error {
	return r.c.doNoContent(ctx, http.MethodDelete, "/api/v1/api-keys/"+id, nil)
}
