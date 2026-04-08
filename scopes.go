package intelmesh

import (
	"context"
	"net/http"
)

// ScopesResource provides methods for scope operations.
type ScopesResource struct {
	c *Client
}

// Create creates a new scope.
func (r *ScopesResource) Create(ctx context.Context, req CreateScopeRequest) (*Scope, error) {
	var result Scope
	if err := r.c.doJSON(ctx, http.MethodPost, "/api/v1/scopes", req, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// Get retrieves a scope by ID.
func (r *ScopesResource) Get(ctx context.Context, id string) (*Scope, error) {
	var result Scope
	if err := r.c.doJSON(ctx, http.MethodGet, "/api/v1/scopes/"+id, nil, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// List returns a paginated list of scopes.
func (r *ScopesResource) List(ctx context.Context, params ListParams) (*PaginatedResponse[Scope], error) {
	var result PaginatedResponse[Scope]
	if err := r.c.doPaginated(ctx, "/api/v1/scopes", params, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// Update updates an existing scope.
func (r *ScopesResource) Update(ctx context.Context, id string, req UpdateScopeRequest) (*Scope, error) {
	var result Scope
	if err := r.c.doJSON(ctx, http.MethodPut, "/api/v1/scopes/"+id, req, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// Delete deletes a scope by ID.
func (r *ScopesResource) Delete(ctx context.Context, id string) error {
	return r.c.doNoContent(ctx, http.MethodDelete, "/api/v1/scopes/"+id, nil)
}
