package intelmesh

import (
	"context"
	"net/http"
)

// RulesResource provides methods for rule operations.
type RulesResource struct {
	c *Client
}

// Create creates a new pipeline rule.
func (r *RulesResource) Create(ctx context.Context, req CreateRuleRequest) (*Rule, error) {
	var result Rule
	if err := r.c.doJSON(ctx, http.MethodPost, "/api/v1/rules", req, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// Get retrieves a rule by ID.
func (r *RulesResource) Get(ctx context.Context, id string) (*Rule, error) {
	var result Rule
	if err := r.c.doJSON(ctx, http.MethodGet, "/api/v1/rules/"+id, nil, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// List returns a paginated list of rules.
func (r *RulesResource) List(ctx context.Context, params ListParams) (*PaginatedResponse[Rule], error) {
	var result PaginatedResponse[Rule]
	if err := r.c.doPaginated(ctx, "/api/v1/rules", params, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// Update updates an existing rule.
func (r *RulesResource) Update(ctx context.Context, id string, req UpdateRuleRequest) (*Rule, error) {
	var result Rule
	if err := r.c.doJSON(ctx, http.MethodPut, "/api/v1/rules/"+id, req, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// Delete deletes a rule by ID.
func (r *RulesResource) Delete(ctx context.Context, id string) error {
	return r.c.doNoContent(ctx, http.MethodDelete, "/api/v1/rules/"+id, nil)
}

// Versions returns a paginated list of versions for a rule.
func (r *RulesResource) Versions(ctx context.Context, id string, params ListParams) (*PaginatedResponse[RuleVersion], error) {
	var result PaginatedResponse[RuleVersion]
	if err := r.c.doPaginated(ctx, "/api/v1/rules/"+id+"/versions", params, &result); err != nil {
		return nil, err
	}

	return &result, nil
}
