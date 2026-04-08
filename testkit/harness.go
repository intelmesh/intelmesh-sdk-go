// Package testkit provides a declarative test harness for IntelMesh
// end-to-end testing via the HTTP API. It manages ephemeral API keys,
// resource provisioning, event sending, and assertion chaining.
package testkit

import (
	"context"
	"fmt"
	"testing"
	"time"

	sdk "github.com/intelmesh/intelmesh-sdk-go"
	"github.com/intelmesh/intelmesh-sdk-go/provision"
)

// allPermissions lists every permission needed for full test coverage.
var allPermissions = []string{ //nolint:gochecknoglobals // test constant
	"events:write", "events:simulate",
	"rules:read", "rules:write",
	"scopes:read", "scopes:write",
	"lists:read", "lists:write",
	"scores:read", "scores:write",
	"api_keys:manage",
	"evaluations:read",
	"audit:read",
}

// projectorWait is the default pause for async projectors.
const projectorWait = 500 * time.Millisecond

// Config holds harness configuration.
type Config struct {
	// BaseURL is the IntelMesh API root (e.g. "http://localhost:8080").
	BaseURL string
	// AdminKey is an API key with api_keys:manage permission.
	AdminKey string
}

// Harness manages test lifecycle: ephemeral API key, provisioning, and assertions.
type Harness struct {
	t           testing.TB
	client      *sdk.Client
	adminClient *sdk.Client
	testKeyID   string
	provisioner *provision.Provisioner
}

// New creates a test harness. It uses AdminKey to create an ephemeral
// API key with all permissions for testing, and deletes it on cleanup.
func New(t testing.TB, cfg Config) *Harness {
	t.Helper()

	adminClient := sdk.New(sdk.Config{
		BaseURL:    cfg.BaseURL,
		APIKey:     cfg.AdminKey,
		Timeout:    0,
		HTTPClient: nil,
	})

	ctx := context.Background()
	result, err := adminClient.APIKeys.Create(ctx, sdk.CreateAPIKeyRequest{
		Name:        fmt.Sprintf("testkit-%d", time.Now().UnixNano()),
		Permissions: allPermissions,
	})
	if err != nil {
		t.Fatalf("testkit: creating ephemeral API key: %v", err)
	}

	testClient := sdk.New(sdk.Config{
		BaseURL:    cfg.BaseURL,
		APIKey:     result.PlainKey,
		Timeout:    0,
		HTTPClient: nil,
	})

	h := &Harness{
		t:           t,
		client:      testClient,
		adminClient: adminClient,
		testKeyID:   result.APIKey.ID,
		provisioner: nil,
	}

	t.Cleanup(func() {
		delErr := adminClient.APIKeys.Delete(context.Background(), h.testKeyID)
		if delErr != nil {
			t.Logf("testkit: warning: failed to delete ephemeral key %s: %v",
				h.testKeyID, delErr)
		}
	})

	return h
}

// Client returns the test-scoped SDK client.
func (h *Harness) Client() *sdk.Client {
	return h.client
}

// Provision applies a provisioner and registers teardown on t.Cleanup.
func (h *Harness) Provision(p *provision.Provisioner) *Harness {
	h.t.Helper()
	h.provisioner = p

	ctx := context.Background()
	if err := p.Apply(ctx); err != nil {
		h.t.Fatalf("testkit: provision failed: %v", err)
	}

	h.t.Cleanup(func() {
		if err := p.Teardown(context.Background()); err != nil {
			h.t.Logf("testkit: teardown warning: %v", err)
		}
	})

	return h
}

// Send sends an event for synchronous evaluation.
func (h *Harness) Send(eventType string, payload map[string]any) *EventAssertion {
	h.t.Helper()

	ctx := context.Background()
	result, err := h.client.Events.Ingest(ctx, sdk.IngestRequest{
		EventType:      eventType,
		Payload:        payload,
		IdempotencyKey: "",
	})

	return &EventAssertion{harness: h, result: result, err: err}
}

// SendSimulate sends to /events/simulate.
func (h *Harness) SendSimulate(eventType string, payload map[string]any) *EventAssertion {
	h.t.Helper()

	ctx := context.Background()
	result, err := h.client.Events.Simulate(ctx, sdk.IngestRequest{
		EventType:      eventType,
		Payload:        payload,
		IdempotencyKey: "",
	})

	return &EventAssertion{harness: h, result: result, err: err}
}

// SendIngestOnly sends with evaluate=false.
func (h *Harness) SendIngestOnly(eventType string, payload map[string]any) *EventAssertion {
	h.t.Helper()

	ctx := context.Background()
	result, err := h.client.Events.IngestOnly(ctx, sdk.IngestRequest{
		EventType:      eventType,
		Payload:        payload,
		IdempotencyKey: "",
	})

	return &EventAssertion{harness: h, result: result, err: err}
}

// WaitForProjectors pauses for async projectors to complete.
func (h *Harness) WaitForProjectors() *Harness {
	time.Sleep(projectorWait)
	return h
}

// VerifyListContains asserts a value is in a list by provisioner name.
func (h *Harness) VerifyListContains(listName, value string) *Harness {
	h.t.Helper()
	h.verifyList(listName, value, true)
	return h
}

// VerifyListNotContains asserts a value is NOT in a list.
func (h *Harness) VerifyListNotContains(listName, value string) *Harness {
	h.t.Helper()
	h.verifyList(listName, value, false)
	return h
}

func (h *Harness) verifyList(listName, value string, shouldContain bool) {
	h.t.Helper()

	if h.provisioner == nil {
		h.t.Fatalf("testkit: no provisioner set, cannot resolve list %q", listName)
	}

	listID := h.provisioner.ListID(listName)
	if listID == "" {
		h.t.Fatalf("testkit: list %q not found in provisioner", listName)
	}

	ctx := context.Background()
	page, err := h.client.Lists.GetItems(ctx, listID, sdk.ListParams{
		Cursor: "",
		Limit:  0,
	})
	if err != nil {
		h.t.Fatalf("testkit: listing items for %q: %v", listName, err)
	}

	found := false
	for _, item := range page.Items {
		if item.Value == value {
			found = true
			break
		}
	}

	if shouldContain && !found {
		h.t.Errorf("testkit: list %q should contain %q but does not (%d items)",
			listName, value, len(page.Items))
	}
	if !shouldContain && found {
		h.t.Errorf("testkit: list %q should NOT contain %q but does",
			listName, value)
	}
}
