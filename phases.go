package intelmesh

import (
	"context"
	"net/http"
)

// PhasesResource provides methods for phase operations.
type PhasesResource struct {
	c *Client
}

// Create creates a new pipeline phase.
func (r *PhasesResource) Create(ctx context.Context, req CreatePhaseRequest) (*Phase, error) {
	var result Phase
	if err := r.c.doJSON(ctx, http.MethodPost, "/api/v1/phases", req, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// Get retrieves a phase by ID.
func (r *PhasesResource) Get(ctx context.Context, id string) (*Phase, error) {
	var result Phase
	if err := r.c.doJSON(ctx, http.MethodGet, "/api/v1/phases/"+id, nil, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// List returns a paginated list of phases.
func (r *PhasesResource) List(ctx context.Context, params ListParams) (*PaginatedResponse[Phase], error) {
	var result PaginatedResponse[Phase]
	if err := r.c.doPaginated(ctx, "/api/v1/phases", params, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// Update updates an existing phase.
func (r *PhasesResource) Update(ctx context.Context, id string, req UpdatePhaseRequest) (*Phase, error) {
	var result Phase
	if err := r.c.doJSON(ctx, http.MethodPut, "/api/v1/phases/"+id, req, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// Delete deletes a phase by ID.
func (r *PhasesResource) Delete(ctx context.Context, id string) error {
	return r.c.doNoContent(ctx, http.MethodDelete, "/api/v1/phases/"+id, nil)
}
