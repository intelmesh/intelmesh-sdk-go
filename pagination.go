package intelmesh

import "context"

// PageFunc is a function that fetches a single page of results.
type PageFunc[T any] func(ctx context.Context, params ListParams) (*PaginatedResponse[T], error)

// Iterator provides cursor-based iteration over paginated results.
type Iterator[T any] struct {
	fetch  PageFunc[T]
	params ListParams
	items  []T
	index  int
	done   bool
}

// NewIterator creates a new paginated iterator.
func NewIterator[T any](fetch PageFunc[T], params ListParams) *Iterator[T] {
	return &Iterator[T]{
		fetch:  fetch,
		params: params,
	}
}

// Next returns the next item from the paginated results.
// It returns false when there are no more items.
func (it *Iterator[T]) Next(ctx context.Context) (T, bool, error) {
	if it.index < len(it.items) {
		item := it.items[it.index]
		it.index++

		return item, true, nil
	}

	if it.done {
		var zero T

		return zero, false, nil
	}

	page, err := it.fetch(ctx, it.params)
	if err != nil {
		var zero T

		return zero, false, err
	}

	it.items = page.Items
	it.index = 0

	if page.NextCursor == "" {
		it.done = true
	} else {
		it.params.Cursor = page.NextCursor
	}

	if len(it.items) == 0 {
		var zero T

		return zero, false, nil
	}

	item := it.items[it.index]
	it.index++

	return item, true, nil
}

// Collect fetches all pages and returns all items as a single slice.
func (it *Iterator[T]) Collect(ctx context.Context) ([]T, error) {
	var all []T

	for {
		item, ok, err := it.Next(ctx)
		if err != nil {
			return nil, err
		}

		if !ok {
			break
		}

		all = append(all, item)
	}

	return all, nil
}
