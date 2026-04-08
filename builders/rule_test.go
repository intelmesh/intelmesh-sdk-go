package builders_test

import (
	"testing"

	"github.com/intelmesh/intelmesh-sdk-go/builders"
)

func TestRuleBuilder_Basic(t *testing.T) {
	t.Parallel()

	req := builders.Rule("block-high-risk").
		Phase("phase-1").
		Priority(10).
		Expression("score > 80").
		Decide("block", "high").
		Build()

	if req.Name != "block-high-risk" {
		t.Errorf("expected name 'block-high-risk', got '%s'", req.Name)
	}

	if req.PhaseID != "phase-1" {
		t.Errorf("expected phase_id 'phase-1', got '%s'", req.PhaseID)
	}

	if req.Priority != 10 {
		t.Errorf("expected priority 10, got %d", req.Priority)
	}

	if req.Expression != "score > 80" {
		t.Errorf("expected expression 'score > 80', got '%s'", req.Expression)
	}

	if req.Actions.Decision == nil {
		t.Fatal("expected decision, got nil")
	}

	if req.Actions.Decision.Action != "block" {
		t.Errorf("expected decision action 'block', got '%s'", req.Actions.Decision.Action)
	}

	if !req.Enabled {
		t.Error("expected enabled=true by default")
	}
}

func TestRuleBuilder_WithScore(t *testing.T) {
	t.Parallel()

	req := builders.Rule("add-score").
		Phase("phase-2").
		Expression("true").
		AddScore(25).
		Build()

	if req.Actions.Score == nil {
		t.Fatal("expected score action, got nil")
	}

	if req.Actions.Score.Add != 25 {
		t.Errorf("expected score add 25, got %d", req.Actions.Score.Add)
	}
}

func TestRuleBuilder_WithMutations(t *testing.T) {
	t.Parallel()

	req := builders.Rule("enrich-event").
		Phase("phase-1").
		Expression("true").
		Mutate("add_to_list", "blocklist", "$.payload.ip").
		Mutate("add_to_list", "watchlist", "$.payload.email").
		Build()

	if len(req.Actions.Mutations) != 2 {
		t.Fatalf("expected 2 mutations, got %d", len(req.Actions.Mutations))
	}

	if req.Actions.Mutations[0].Target != "blocklist" {
		t.Errorf("expected first mutation target 'blocklist', got '%s'", req.Actions.Mutations[0].Target)
	}
}

func TestRuleBuilder_DryRun(t *testing.T) {
	t.Parallel()

	req := builders.Rule("test-rule").
		Phase("phase-1").
		Expression("true").
		DryRun(true).
		Enabled(false).
		Build()

	if !req.DryRun {
		t.Error("expected dry_run=true")
	}

	if req.Enabled {
		t.Error("expected enabled=false")
	}
}

func TestRuleBuilder_WithFlow(t *testing.T) {
	t.Parallel()

	req := builders.Rule("halt-rule").
		Phase("phase-1").
		Expression("score > 100").
		Flow("halt").
		Build()

	if req.Actions.Flow != "halt" {
		t.Errorf("expected flow 'halt', got '%s'", req.Actions.Flow)
	}
}

func TestRuleBuilder_ApplicableWhen(t *testing.T) {
	t.Parallel()

	req := builders.Rule("conditional").
		Phase("phase-1").
		ApplicableWhen("event_type == 'transaction'").
		Expression("amount > 1000").
		Decide("review", "medium").
		Build()

	if req.ApplicableWhen != "event_type == 'transaction'" {
		t.Errorf("expected applicable_when, got '%s'", req.ApplicableWhen)
	}
}
