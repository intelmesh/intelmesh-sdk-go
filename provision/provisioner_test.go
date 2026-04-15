package provision_test

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	sdk "github.com/intelmesh/intelmesh-sdk-go"
	"github.com/intelmesh/intelmesh-sdk-go/provision"
)

// mockAPI records created and deleted resources.
type mockAPI struct {
	mu       sync.Mutex
	created  map[string]int
	deleted  map[string]int
	idSeq    int
	teardown bool
	lastRule *sdk.CreateRuleRequest
}

func newMockAPI() *mockAPI {
	return &mockAPI{
		created: make(map[string]int),
		deleted: make(map[string]int),
	}
}

func (m *mockAPI) nextID() string {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.idSeq++
	return fmt.Sprintf("id-%d", m.idSeq)
}

func (m *mockAPI) record(resource string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.created[resource]++
}

func (m *mockAPI) recordDelete(resource string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.deleted[resource]++
	m.teardown = true
}

func (m *mockAPI) handler() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("POST /api/v1/phases", func(w http.ResponseWriter, _ *http.Request) {
		m.record("phases")
		respond(w, map[string]any{
			"id": m.nextID(), "name": "p", "position": 1,
			"created_at": "2025-01-01T00:00:00Z",
		})
	})

	mux.HandleFunc("POST /api/v1/scopes", func(w http.ResponseWriter, _ *http.Request) {
		m.record("scopes")
		respond(w, map[string]any{
			"id": m.nextID(), "name": "s", "json_path": "$.x",
			"created_at": "2025-01-01T00:00:00Z",
		})
	})

	mux.HandleFunc("POST /api/v1/lists", func(w http.ResponseWriter, _ *http.Request) {
		m.record("lists")
		respond(w, map[string]any{
			"id": m.nextID(), "name": "l",
			"created_at": "2025-01-01T00:00:00Z",
			"updated_at": "2025-01-01T00:00:00Z",
		})
	})

	mux.HandleFunc("POST /api/v1/rules", func(w http.ResponseWriter, r *http.Request) {
		m.record("rules")

		body, _ := io.ReadAll(r.Body)
		var req sdk.CreateRuleRequest
		_ = json.Unmarshal(body, &req)
		m.mu.Lock()
		m.lastRule = &req
		m.mu.Unlock()

		respond(w, map[string]any{
			"id": m.nextID(), "name": req.Name, "phase_id": req.PhaseID,
			"priority": req.Priority, "expression": req.Expression,
			"actions": map[string]any{}, "enabled": true,
			"created_at": "2025-01-01T00:00:00Z",
			"updated_at": "2025-01-01T00:00:00Z",
		})
	})

	mux.HandleFunc("DELETE /api/v1/phases/", func(w http.ResponseWriter, _ *http.Request) {
		m.recordDelete("phases")
		w.WriteHeader(http.StatusNoContent)
	})

	mux.HandleFunc("DELETE /api/v1/scopes/", func(w http.ResponseWriter, _ *http.Request) {
		m.recordDelete("scopes")
		w.WriteHeader(http.StatusNoContent)
	})

	mux.HandleFunc("DELETE /api/v1/lists/", func(w http.ResponseWriter, _ *http.Request) {
		m.recordDelete("lists")
		w.WriteHeader(http.StatusNoContent)
	})

	mux.HandleFunc("DELETE /api/v1/rules/", func(w http.ResponseWriter, _ *http.Request) {
		m.recordDelete("rules")
		w.WriteHeader(http.StatusNoContent)
	})

	return mux
}

func respond(w http.ResponseWriter, data any) {
	w.Header().Set("Content-Type", "application/json")
	envelope := sdk.Envelope[json.RawMessage]{}
	raw, _ := json.Marshal(data)
	envelope.Data = raw
	_ = json.NewEncoder(w).Encode(envelope)
}

func TestProvisioner_ApplyAndTeardown(t *testing.T) {
	t.Parallel()

	mock := newMockAPI()
	srv := httptest.NewServer(mock.handler())
	defer srv.Close()

	client := sdk.New(sdk.Config{BaseURL: srv.URL, APIKey: "test"})
	p := provision.New(client).
		Phase("screening", 1).
		Scope("client_uuid", "event.metadata.client_uuid").
		List("blocklist").
		Rule("block-rule").
		InPhase("screening").Priority(1).
		When("true").Decide("block", "critical").Halt().Done()

	ctx := context.Background()
	if err := p.Apply(ctx); err != nil {
		t.Fatalf("Apply failed: %v", err)
	}

	if p.PhaseID("screening") == "" {
		t.Error("expected PhaseID to be set")
	}
	if p.ScopeID("client_uuid") == "" {
		t.Error("expected ScopeID to be set")
	}
	if p.ListID("blocklist") == "" {
		t.Error("expected ListID to be set")
	}
	if p.RuleID("block-rule") == "" {
		t.Error("expected RuleID to be set")
	}

	if mock.created["phases"] != 1 {
		t.Errorf("expected 1 phase, got %d", mock.created["phases"])
	}
	if mock.created["scopes"] != 1 {
		t.Errorf("expected 1 scope, got %d", mock.created["scopes"])
	}
	if mock.created["lists"] != 1 {
		t.Errorf("expected 1 list, got %d", mock.created["lists"])
	}
	if mock.created["rules"] != 1 {
		t.Errorf("expected 1 rule, got %d", mock.created["rules"])
	}

	if err := p.Teardown(ctx); err != nil {
		t.Fatalf("Teardown failed: %v", err)
	}

	if !mock.teardown {
		t.Error("expected teardown to execute deletes")
	}
}

func TestProvisioner_RuleMissingPhase(t *testing.T) {
	t.Parallel()

	mock := newMockAPI()
	srv := httptest.NewServer(mock.handler())
	defer srv.Close()

	client := sdk.New(sdk.Config{BaseURL: srv.URL, APIKey: "test"})
	p := provision.New(client).
		Rule("orphan-rule").
		InPhase("nonexistent").Priority(1).
		When("true").Done()

	err := p.Apply(context.Background())
	if err == nil {
		t.Fatal("expected error for missing phase, got nil")
	}
}

func TestProvisioner_MultiplePhases(t *testing.T) {
	t.Parallel()

	mock := newMockAPI()
	srv := httptest.NewServer(mock.handler())
	defer srv.Close()

	client := sdk.New(sdk.Config{BaseURL: srv.URL, APIKey: "test"})
	p := provision.New(client).
		Phase("screening", 1).
		Phase("scoring", 2).
		Phase("decision", 3)

	if err := p.Apply(context.Background()); err != nil {
		t.Fatalf("Apply failed: %v", err)
	}

	if mock.created["phases"] != 3 {
		t.Errorf("expected 3 phases, got %d", mock.created["phases"])
	}
}

func TestProvisioner_RuleWithMutations(t *testing.T) {
	t.Parallel()

	mock := newMockAPI()
	srv := httptest.NewServer(mock.handler())
	defer srv.Close()

	client := sdk.New(sdk.Config{BaseURL: srv.URL, APIKey: "test"})
	p := provision.New(client).
		Phase("scoring", 1).
		List("med_blocked").
		Rule("med-add").InPhase("scoring").Priority(1).
		ApplicableWhen("event.type == 'bacen.med.add'").
		When("true").
		MutateList("list.add", "med_blocked", "event.metadata.client_uuid").
		Continue().Done()

	ctx := context.Background()
	if err := p.Apply(ctx); err != nil {
		t.Fatalf("Apply failed: %v", err)
	}

	mock.mu.Lock()
	lastRule := mock.lastRule
	mock.mu.Unlock()

	if lastRule == nil {
		t.Fatal("expected rule request to be captured")
	}
	if lastRule.ApplicableWhen != "event.type == 'bacen.med.add'" {
		t.Errorf("applicable_when: got %q", lastRule.ApplicableWhen)
	}
	if len(lastRule.Actions.Mutations) != 1 {
		t.Fatalf("expected 1 mutation, got %d", len(lastRule.Actions.Mutations))
	}
	m := lastRule.Actions.Mutations[0]
	if m.Type != "list.add" {
		t.Errorf("mutation type: got %q, want %q", m.Type, "list.add")
	}
	if m.Target != "med_blocked" {
		t.Errorf("mutation target: got %q, want %q", m.Target, "med_blocked")
	}
	if m.ValuePath != "event.metadata.client_uuid" {
		t.Errorf("mutation value_path: got %q, want %q", m.ValuePath, "event.metadata.client_uuid")
	}
	if lastRule.Actions.Flow != "continue" {
		t.Errorf("flow: got %q, want %q", lastRule.Actions.Flow, "continue")
	}
}

func TestProvisioner_RuleWithScoreAndDryRun(t *testing.T) {
	t.Parallel()

	mock := newMockAPI()
	srv := httptest.NewServer(mock.handler())
	defer srv.Close()

	client := sdk.New(sdk.Config{BaseURL: srv.URL, APIKey: "test"})
	p := provision.New(client).
		Phase("scoring", 1).
		Rule("score-rule").InPhase("scoring").Priority(1).
		When("true").AddScore(100).DryRun().Continue().Done()

	ctx := context.Background()
	if err := p.Apply(ctx); err != nil {
		t.Fatalf("Apply failed: %v", err)
	}

	mock.mu.Lock()
	lastRule := mock.lastRule
	mock.mu.Unlock()

	if lastRule == nil {
		t.Fatal("expected rule request to be captured")
	}
	if lastRule.Actions.Score == nil {
		t.Fatal("expected score action to be set")
	}
	if lastRule.Actions.Score.Add != 100 {
		t.Errorf("score add: got %d, want %d", lastRule.Actions.Score.Add, 100)
	}
	if !lastRule.DryRun {
		t.Error("expected dry_run to be true")
	}
	if lastRule.Actions.Flow != "continue" {
		t.Errorf("flow: got %q, want %q", lastRule.Actions.Flow, "continue")
	}
}

func TestProvisioner_FullPipeline(t *testing.T) {
	t.Parallel()

	mock := newMockAPI()
	srv := httptest.NewServer(mock.handler())
	defer srv.Close()

	client := sdk.New(sdk.Config{BaseURL: srv.URL, APIKey: "test"})
	p := provision.New(client).
		Phase("screening", 1).
		Phase("scoring", 2).
		Scope("client_uuid", "event.metadata.client_uuid").
		List("med_blocked").
		Rule("med-block").InPhase("screening").Priority(1).
		When("list.contains('med_blocked', event.metadata.client_uuid)").
		Decide("block", "critical").Halt().Done().
		Rule("med-add").InPhase("scoring").Priority(1).
		ApplicableWhen("event.type == 'bacen.med.add'").When("true").
		MutateList("list.add", "med_blocked", "event.metadata.client_uuid").
		Continue().Done()

	ctx := context.Background()
	if err := p.Apply(ctx); err != nil {
		t.Fatalf("Apply failed: %v", err)
	}

	if mock.created["phases"] != 2 {
		t.Errorf("expected 2 phases, got %d", mock.created["phases"])
	}
	if mock.created["scopes"] != 1 {
		t.Errorf("expected 1 scope, got %d", mock.created["scopes"])
	}
	if mock.created["lists"] != 1 {
		t.Errorf("expected 1 list, got %d", mock.created["lists"])
	}
	if mock.created["rules"] != 2 {
		t.Errorf("expected 2 rules, got %d", mock.created["rules"])
	}

	if err := p.Teardown(ctx); err != nil {
		t.Fatalf("Teardown failed: %v", err)
	}

	if mock.deleted["rules"] != 2 {
		t.Errorf("expected 2 rule deletes, got %d", mock.deleted["rules"])
	}
	if mock.deleted["lists"] != 1 {
		t.Errorf("expected 1 list delete, got %d", mock.deleted["lists"])
	}
	if mock.deleted["scopes"] != 1 {
		t.Errorf("expected 1 scope delete, got %d", mock.deleted["scopes"])
	}
	if mock.deleted["phases"] != 2 {
		t.Errorf("expected 2 phase deletes, got %d", mock.deleted["phases"])
	}
}

func TestProvisioner_RuleWithScoreMutations(t *testing.T) {
	t.Parallel()

	mock := newMockAPI()
	srv := httptest.NewServer(mock.handler())
	defer srv.Close()

	client := sdk.New(sdk.Config{BaseURL: srv.URL, APIKey: "test"})
	p := provision.New(client).
		Phase("scoring", 1).
		Rule("bump").InPhase("scoring").Priority(1).
		When("true").
		AddScoped("ip", "event.metadata.source_ip", 30).
		Done()

	if err := p.Apply(context.Background()); err != nil {
		t.Fatalf("Apply: %v", err)
	}

	mock.mu.Lock()
	lastRule := mock.lastRule
	mock.mu.Unlock()

	if lastRule == nil {
		t.Fatal("expected rule request to be captured")
	}
	if len(lastRule.Actions.Mutations) != 1 {
		t.Fatalf("expected 1 mutation, got %d", len(lastRule.Actions.Mutations))
	}
	m := lastRule.Actions.Mutations[0]
	if m.Type != "score.add" || m.Target != "ip" {
		t.Fatalf("unexpected mutation: %+v", m)
	}
	if m.ValuePath != "event.metadata.source_ip" {
		t.Errorf("value_path: got %q, want %q", m.ValuePath, "event.metadata.source_ip")
	}
	if m.Amount == nil || *m.Amount != 30 {
		t.Fatalf("expected amount=30, got %v", m.Amount)
	}
}

func TestProvisioner_RuleSkipPhase(t *testing.T) {
	t.Parallel()

	mock := newMockAPI()
	srv := httptest.NewServer(mock.handler())
	defer srv.Close()

	client := sdk.New(sdk.Config{BaseURL: srv.URL, APIKey: "test"})
	p := provision.New(client).
		Phase("pre", 1).
		Rule("skip").InPhase("pre").Priority(1).
		When("true").SkipPhase().Done()

	ctx := context.Background()
	if err := p.Apply(ctx); err != nil {
		t.Fatalf("Apply failed: %v", err)
	}

	mock.mu.Lock()
	lastRule := mock.lastRule
	mock.mu.Unlock()

	if lastRule == nil {
		t.Fatal("expected rule request")
	}
	if lastRule.Actions.Flow != "skip_phase" {
		t.Errorf("flow: got %q, want %q", lastRule.Actions.Flow, "skip_phase")
	}
}
