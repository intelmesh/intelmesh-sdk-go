package intelmesh

import "context"

// AuditResource provides methods for audit log operations.
type AuditResource struct {
	c *Client
}

// List returns a paginated list of audit log entries.
func (r *AuditResource) List(ctx context.Context, params ListParams) (*PaginatedResponse[AuditEntry], error) {
	var result PaginatedResponse[AuditEntry]
	if err := r.c.doPaginated(ctx, "/api/v1/audit", params, &result); err != nil {
		return nil, err
	}

	return &result, nil
}
