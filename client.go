package intelmesh

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

const defaultTimeout = 30 * time.Second

// Config holds client configuration.
type Config struct {
	// BaseURL is the root URL of the IntelMesh API (e.g. "https://api.intelmesh.io").
	BaseURL string
	// APIKey is the authentication key for the API.
	APIKey string
	// Timeout is the HTTP request timeout. Defaults to 30 seconds.
	Timeout time.Duration
	// HTTPClient is an optional custom HTTP client. If nil, a default client is created.
	HTTPClient *http.Client
}

// Client is the IntelMesh API client.
type Client struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client

	// Events provides methods for event operations.
	Events *EventsResource
	// Rules provides methods for rule operations.
	Rules *RulesResource
	// Phases provides methods for phase operations.
	Phases *PhasesResource
	// Scopes provides methods for scope operations.
	Scopes *ScopesResource
	// Lists provides methods for list operations.
	Lists *ListsResource
	// Scores provides methods for score operations.
	Scores *ScoresResource
	// APIKeys provides methods for API key operations.
	APIKeys *APIKeysResource
	// Evaluations provides methods for evaluation operations.
	Evaluations *EvaluationsResource
	// Audit provides methods for audit log operations.
	Audit *AuditResource
}

// New creates a new IntelMesh client.
func New(cfg Config) *Client {
	if cfg.Timeout == 0 {
		cfg.Timeout = defaultTimeout
	}

	httpClient := cfg.HTTPClient
	if httpClient == nil {
		httpClient = &http.Client{Timeout: cfg.Timeout}
	}

	c := &Client{
		baseURL:    cfg.BaseURL,
		apiKey:     cfg.APIKey,
		httpClient: httpClient,
	}

	c.Events = &EventsResource{c: c}
	c.Rules = &RulesResource{c: c}
	c.Phases = &PhasesResource{c: c}
	c.Scopes = &ScopesResource{c: c}
	c.Lists = &ListsResource{c: c}
	c.Scores = &ScoresResource{c: c}
	c.APIKeys = &APIKeysResource{c: c}
	c.Evaluations = &EvaluationsResource{c: c}
	c.Audit = &AuditResource{c: c}

	return c
}

// do executes an HTTP request and returns the raw response.
func (c *Client) do(
	ctx context.Context,
	method string,
	path string,
	body any,
) (*http.Response, error) {
	reqURL := c.baseURL + path

	var bodyReader io.Reader
	if body != nil {
		encoded, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("intelmesh: failed to encode request body: %w", err)
		}

		bodyReader = bytes.NewReader(encoded)
	}

	req, err := http.NewRequestWithContext(ctx, method, reqURL, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("intelmesh: failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("intelmesh: request failed: %w", err)
	}

	return resp, nil
}

// doJSON executes a request and decodes the JSON response into result.
func (c *Client) doJSON(
	ctx context.Context,
	method string,
	path string,
	body any,
	result any,
) error {
	resp, err := c.do(ctx, method, path, body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= http.StatusBadRequest {
		return parseErrorResponse(resp)
	}

	var envelope Envelope[json.RawMessage]

	if err := json.NewDecoder(resp.Body).Decode(&envelope); err != nil {
		return fmt.Errorf("intelmesh: failed to decode response: %w", err)
	}

	if envelope.Error != nil {
		return &IntelMeshError{
			StatusCode: resp.StatusCode,
			Code:       envelope.Error.Code,
			Message:    envelope.Error.Message,
		}
	}

	if result != nil {
		if err := json.Unmarshal(envelope.Data, result); err != nil {
			return fmt.Errorf("intelmesh: failed to decode data: %w", err)
		}
	}

	return nil
}

// doNoContent executes a request expecting no response body (e.g. 202, 204).
func (c *Client) doNoContent(
	ctx context.Context,
	method string,
	path string,
	body any,
) error {
	resp, err := c.do(ctx, method, path, body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= http.StatusBadRequest {
		return parseErrorResponse(resp)
	}

	return nil
}

// doPaginated executes a paginated list request.
func (c *Client) doPaginated(
	ctx context.Context,
	path string,
	params ListParams,
	result any,
) error {
	u, err := url.Parse(path)
	if err != nil {
		return fmt.Errorf("intelmesh: invalid path: %w", err)
	}

	q := u.Query()
	if params.Cursor != "" {
		q.Set("cursor", params.Cursor)
	}

	if params.Limit > 0 {
		q.Set("limit", strconv.Itoa(params.Limit))
	}

	u.RawQuery = q.Encode()

	return c.doJSON(ctx, http.MethodGet, u.String(), nil, result)
}

// parseErrorResponse reads an error response body and returns a typed error.
func parseErrorResponse(resp *http.Response) error {
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return &IntelMeshError{
			StatusCode: resp.StatusCode,
			Code:       CodeInternal,
			Message:    "failed to read error response",
		}
	}

	var envelope Envelope[json.RawMessage]
	if err := json.Unmarshal(bodyBytes, &envelope); err != nil {
		return &IntelMeshError{
			StatusCode: resp.StatusCode,
			Code:       CodeInternal,
			Message:    string(bodyBytes),
		}
	}

	if envelope.Error != nil {
		return &IntelMeshError{
			StatusCode: resp.StatusCode,
			Code:       envelope.Error.Code,
			Message:    envelope.Error.Message,
		}
	}

	return &IntelMeshError{
		StatusCode: resp.StatusCode,
		Code:       CodeInternal,
		Message:    "unknown error",
	}
}
