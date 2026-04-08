package intelmesh

import (
	"context"
	"net/http"
)

// ListsResource provides methods for list operations.
type ListsResource struct {
	c *Client
}

// Create creates a new list.
func (r *ListsResource) Create(ctx context.Context, req CreateListRequest) (*List, error) {
	var result List
	if err := r.c.doJSON(ctx, http.MethodPost, "/api/v1/lists", req, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// Get retrieves a list by ID.
func (r *ListsResource) Get(ctx context.Context, id string) (*List, error) {
	var result List
	if err := r.c.doJSON(ctx, http.MethodGet, "/api/v1/lists/"+id, nil, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// List returns a paginated list of lists.
func (r *ListsResource) List(ctx context.Context, params ListParams) (*PaginatedResponse[List], error) {
	var result PaginatedResponse[List]
	if err := r.c.doPaginated(ctx, "/api/v1/lists", params, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// Update updates an existing list.
func (r *ListsResource) Update(ctx context.Context, id string, req UpdateListRequest) (*List, error) {
	var result List
	if err := r.c.doJSON(ctx, http.MethodPut, "/api/v1/lists/"+id, req, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// Delete deletes a list by ID.
func (r *ListsResource) Delete(ctx context.Context, id string) error {
	return r.c.doNoContent(ctx, http.MethodDelete, "/api/v1/lists/"+id, nil)
}

// AddItem adds an item to a list.
func (r *ListsResource) AddItem(ctx context.Context, listID string, req AddListItemRequest) (*ListItem, error) {
	var result ListItem
	if err := r.c.doJSON(ctx, http.MethodPost, "/api/v1/lists/"+listID+"/items", req, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// RemoveItem removes an item from a list.
func (r *ListsResource) RemoveItem(ctx context.Context, listID string, itemID string) error {
	return r.c.doNoContent(ctx, http.MethodDelete, "/api/v1/lists/"+listID+"/items/"+itemID, nil)
}

// GetItems returns a paginated list of items in a list.
func (r *ListsResource) GetItems(ctx context.Context, listID string, params ListParams) (*PaginatedResponse[ListItem], error) {
	var result PaginatedResponse[ListItem]
	if err := r.c.doPaginated(ctx, "/api/v1/lists/"+listID+"/items", params, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// BulkImport imports multiple items into a list at once.
func (r *ListsResource) BulkImport(ctx context.Context, listID string, req BulkImportListRequest) (*BulkImportListResult, error) {
	var result BulkImportListResult
	if err := r.c.doJSON(ctx, http.MethodPost, "/api/v1/lists/"+listID+"/items/bulk", req, &result); err != nil {
		return nil, err
	}

	return &result, nil
}
