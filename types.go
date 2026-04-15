// Package intelmesh provides a Go client for the IntelMesh Risk Intelligence Engine API.
package intelmesh

import "time"

// Decision represents an evaluation decision.
type Decision struct {
	Action   string         `json:"action"`
	Severity string         `json:"severity"`
	Metadata map[string]any `json:"metadata,omitempty"`
}

// IngestRequest is the request body for event ingestion.
type IngestRequest struct {
	EventType      string         `json:"event_type"`
	Payload        map[string]any `json:"payload"`
	IdempotencyKey string         `json:"idempotency_key,omitempty"`
}

// IngestResult is the response from event ingestion.
type IngestResult struct {
	EventID        string    `json:"event_id"`
	Decision       *Decision `json:"decision,omitempty"`
	TransientScore int64     `json:"transient_score"`
	DurationMs     int       `json:"duration_ms"`
}

// Phase represents a pipeline execution phase.
type Phase struct {
	ID             string    `json:"id"`
	Name           string    `json:"name"`
	Position       int       `json:"position"`
	ApplicableWhen string    `json:"applicable_when,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
}

// CreatePhaseRequest is the request body for creating a phase.
type CreatePhaseRequest struct {
	Name           string `json:"name"`
	Position       int    `json:"position"`
	ApplicableWhen string `json:"applicable_when,omitempty"`
}

// UpdatePhaseRequest is the request body for updating a phase.
type UpdatePhaseRequest struct {
	Name           string `json:"name,omitempty"`
	Position       *int   `json:"position,omitempty"`
	ApplicableWhen string `json:"applicable_when,omitempty"`
}

// Scope represents a scope for tracking event dimensions.
type Scope struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	JSONPath  string    `json:"json_path"`
	CreatedAt time.Time `json:"created_at"`
}

// CreateScopeRequest is the request body for creating a scope.
type CreateScopeRequest struct {
	Name     string `json:"name"`
	JSONPath string `json:"json_path"`
}

// UpdateScopeRequest is the request body for updating a scope.
type UpdateScopeRequest struct {
	Name     string `json:"name,omitempty"`
	JSONPath string `json:"json_path,omitempty"`
}

// List represents a named list (allowlist/denylist).
type List struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// CreateListRequest is the request body for creating a list.
type CreateListRequest struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}

// UpdateListRequest is the request body for updating a list.
type UpdateListRequest struct {
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
}

// ListItem represents an item in a list.
type ListItem struct {
	ID        string    `json:"id"`
	ListID    string    `json:"list_id"`
	Value     string    `json:"value"`
	CreatedAt time.Time `json:"created_at"`
}

// AddListItemRequest is the request body for adding an item to a list.
type AddListItemRequest struct {
	Value string `json:"value"`
}

// BulkImportListRequest is the request body for bulk importing items to a list.
type BulkImportListRequest struct {
	Values []string `json:"values"`
}

// BulkImportListResult is the response from a bulk import operation.
type BulkImportListResult struct {
	Imported int `json:"imported"`
	Skipped  int `json:"skipped"`
}

// ScoreOperation defines a score delta for rule actions.
type ScoreOperation struct {
	Add int64 `json:"add"`
}

// Mutation describes a side-effect produced by a matching rule.
type Mutation struct {
	Type      string `json:"type"`
	Target    string `json:"target"`
	ValuePath string `json:"value_path"`
	Amount    *int64 `json:"amount,omitempty"`
}

// Actions defines what happens when a rule matches.
type Actions struct {
	Decision  *Decision       `json:"decision,omitempty"`
	Score     *ScoreOperation `json:"score,omitempty"`
	Flow      string          `json:"flow,omitempty"`
	Mutations []Mutation      `json:"mutations,omitempty"`
}

// Rule represents a pipeline rule.
type Rule struct {
	ID               string    `json:"id"`
	Name             string    `json:"name"`
	PhaseID          string    `json:"phase_id"`
	Priority         int       `json:"priority"`
	ApplicableWhen   string    `json:"applicable_when,omitempty"`
	Expression       string    `json:"expression"`
	Actions          Actions   `json:"actions"`
	Enabled          bool      `json:"enabled"`
	DryRun           bool      `json:"dry_run"`
	CurrentVersionID string    `json:"current_version_id,omitempty"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

// CreateRuleRequest is the request body for creating a rule.
type CreateRuleRequest struct {
	Name           string  `json:"name"`
	PhaseID        string  `json:"phase_id"`
	Priority       int     `json:"priority"`
	ApplicableWhen string  `json:"applicable_when,omitempty"`
	Expression     string  `json:"expression"`
	Actions        Actions `json:"actions"`
	Enabled        bool    `json:"enabled"`
	DryRun         bool    `json:"dry_run"`
}

// UpdateRuleRequest is the request body for updating a rule.
type UpdateRuleRequest struct {
	Name           string   `json:"name,omitempty"`
	Priority       *int     `json:"priority,omitempty"`
	ApplicableWhen string   `json:"applicable_when,omitempty"`
	Expression     string   `json:"expression,omitempty"`
	Actions        *Actions `json:"actions,omitempty"`
	Enabled        *bool    `json:"enabled,omitempty"`
	DryRun         *bool    `json:"dry_run,omitempty"`
}

// RuleVersion represents a historical version of a rule.
type RuleVersion struct {
	ID         string    `json:"id"`
	RuleID     string    `json:"rule_id"`
	Version    int       `json:"version"`
	Expression string    `json:"expression"`
	Actions    Actions   `json:"actions"`
	CreatedAt  time.Time `json:"created_at"`
}

// Score represents an accumulated score.
type Score struct {
	ScopeName  string `json:"scope_name"`
	ScopeValue string `json:"scope_value"`
	Score      int64  `json:"score"`
}

// SetScoreRequest is the request body for setting a score.
type SetScoreRequest struct {
	ScopeName  string `json:"scope_name"`
	ScopeValue string `json:"scope_value"`
	Score      int64  `json:"score"`
}

// APIKey represents an API key (key hash is never exposed).
type APIKey struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Permissions []string  `json:"permissions"`
	Enabled     bool      `json:"enabled"`
	CreatedAt   time.Time `json:"created_at"`
}

// CreateAPIKeyRequest is the request body for creating an API key.
type CreateAPIKeyRequest struct {
	Name        string   `json:"name"`
	Permissions []string `json:"permissions"`
}

// UpdateAPIKeyRequest is the request body for updating an API key.
type UpdateAPIKeyRequest struct {
	Name        string   `json:"name,omitempty"`
	Permissions []string `json:"permissions,omitempty"`
	Enabled     *bool    `json:"enabled,omitempty"`
}

// CreateAPIKeyResult includes the plain key (shown only once).
type CreateAPIKeyResult struct {
	APIKey   APIKey `json:"api_key"`
	PlainKey string `json:"plain_key"`
}

// Evaluation represents an event evaluation result.
type Evaluation struct {
	ID             string    `json:"id"`
	EventID        string    `json:"event_id"`
	EventType      string    `json:"event_type"`
	Decision       *Decision `json:"decision,omitempty"`
	TransientScore int64     `json:"transient_score"`
	DurationMs     int       `json:"duration_ms"`
	CreatedAt      time.Time `json:"created_at"`
}

// AuditEntry represents an audit log entry.
type AuditEntry struct {
	ID         string         `json:"id"`
	Action     string         `json:"action"`
	Resource   string         `json:"resource"`
	ResourceID string         `json:"resource_id"`
	ActorID    string         `json:"actor_id"`
	Details    map[string]any `json:"details,omitempty"`
	CreatedAt  time.Time      `json:"created_at"`
}

// PaginatedResponse wraps paginated results.
type PaginatedResponse[T any] struct {
	Items      []T    `json:"items"`
	NextCursor string `json:"next_cursor,omitempty"`
	Count      int    `json:"count"`
}

// Envelope is the standard API response wrapper.
type Envelope[T any] struct {
	Data  T         `json:"data,omitempty"`
	Error *APIError `json:"error,omitempty"`
}

// APIError is the structured error from the API.
type APIError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// ListParams holds common pagination parameters.
type ListParams struct {
	Cursor string `json:"cursor,omitempty"`
	Limit  int    `json:"limit,omitempty"`
}
