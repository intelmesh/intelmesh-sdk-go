package intelmesh

import (
	"context"
	"net/http"
)

// EvaluationsResource provides methods for evaluation operations.
type EvaluationsResource struct {
	c *Client
}

// Get retrieves an evaluation by ID.
func (r *EvaluationsResource) Get(ctx context.Context, id string) (*Evaluation, error) {
	var result Evaluation
	if err := r.c.doJSON(ctx, http.MethodGet, "/api/v1/evaluations/"+id, nil, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// List returns a paginated list of evaluations.
func (r *EvaluationsResource) List(ctx context.Context, params ListParams) (*PaginatedResponse[Evaluation], error) {
	var result PaginatedResponse[Evaluation]
	if err := r.c.doPaginated(ctx, "/api/v1/evaluations", params, &result); err != nil {
		return nil, err
	}

	return &result, nil
}
