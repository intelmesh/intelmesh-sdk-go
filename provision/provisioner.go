// Package provision provides a fluent builder for declaratively provisioning
// IntelMesh phases, scopes, lists, and rules via the HTTP API.
package provision

import (
	"context"
	"errors"
	"fmt"

	sdk "github.com/intelmesh/intelmesh-sdk-go"
)

// Sentinel errors for provisioning operations.
var ( //nolint:gochecknoglobals // sentinel errors are idiomatic
	// ErrPhaseNotFound indicates a rule references a phase that was not provisioned.
	ErrPhaseNotFound = errors.New("phase not found")

	// ErrUnknownStep indicates an internal step kind that is not recognized.
	ErrUnknownStep = errors.New("unknown step kind")
)

// stepKind identifies the resource type for a provisioning step.
type stepKind int

const (
	kindPhase stepKind = iota
	kindScope
	kindList
	kindRule
)

// step represents a single resource to create during Apply.
type step struct {
	kind           stepKind
	name           string
	position       int    // phases only
	applicableWhen string // phases only — optional CEL expression
	jsonPath       string // scopes only
	rule           *RuleBuilder
}

// Provisioner builds and executes a declarative resource plan against
// the IntelMesh API. Resources are created in dependency order
// (phases, scopes, lists, rules) and torn down in reverse.
type Provisioner struct {
	client *sdk.Client
	steps  []step

	// Resolved IDs keyed by user-given name.
	phases map[string]string
	lists  map[string]string
	scopes map[string]string
	rules  map[string]string
}

// New creates a new Provisioner backed by the given SDK client.
func New(client *sdk.Client) *Provisioner {
	return &Provisioner{
		client: client,
		steps:  nil,
		phases: make(map[string]string),
		lists:  make(map[string]string),
		scopes: make(map[string]string),
		rules:  make(map[string]string),
	}
}

// Phase registers a pipeline phase to be created.
func (p *Provisioner) Phase(name string, position int) *Provisioner {
	p.steps = append(p.steps, step{
		kind: kindPhase, name: name, position: position,
		applicableWhen: "", jsonPath: "", rule: nil,
	})
	return p
}

// PhaseWithFilter registers a pipeline phase with an applicable_when CEL expression.
func (p *Provisioner) PhaseWithFilter(name string, position int, applicableWhen string) *Provisioner {
	p.steps = append(p.steps, step{
		kind: kindPhase, name: name, position: position,
		applicableWhen: applicableWhen, jsonPath: "", rule: nil,
	})
	return p
}

// Scope registers a scope to be created.
func (p *Provisioner) Scope(name, jsonPath string) *Provisioner {
	p.steps = append(p.steps, step{
		kind: kindScope, name: name, position: 0,
		jsonPath: jsonPath, rule: nil,
	})
	return p
}

// List registers a named list to be created.
func (p *Provisioner) List(name string) *Provisioner {
	p.steps = append(p.steps, step{
		kind: kindList, name: name, position: 0,
		jsonPath: "", rule: nil,
	})
	return p
}

// Rule starts building a rule and returns a RuleBuilder.
func (p *Provisioner) Rule(name string) *RuleBuilder {
	return newRuleBuilder(p, name)
}

// Apply creates all registered resources via the API in registration order.
// Returns an error on the first failure.
func (p *Provisioner) Apply(ctx context.Context) error {
	for _, s := range p.steps {
		if err := p.applyStep(ctx, s); err != nil {
			return err
		}
	}
	return nil
}

// Teardown deletes all created resources in reverse dependency order:
// rules, lists, scopes, phases.
func (p *Provisioner) Teardown(ctx context.Context) error {
	var firstErr error
	record := func(err error) {
		if firstErr == nil {
			firstErr = err
		}
	}

	deleteAll(ctx, p.client.Rules.Delete, p.rules, record)
	deleteAll(ctx, p.client.Lists.Delete, p.lists, record)
	deleteAll(ctx, p.client.Scopes.Delete, p.scopes, record)
	deleteAll(ctx, p.client.Phases.Delete, p.phases, record)

	return firstErr
}

// PhaseID returns the provisioned ID for the named phase.
func (p *Provisioner) PhaseID(name string) string { return p.phases[name] }

// ListID returns the provisioned ID for the named list.
func (p *Provisioner) ListID(name string) string { return p.lists[name] }

// ScopeID returns the provisioned ID for the named scope.
func (p *Provisioner) ScopeID(name string) string { return p.scopes[name] }

// RuleID returns the provisioned ID for the named rule.
func (p *Provisioner) RuleID(name string) string { return p.rules[name] }

func (p *Provisioner) applyStep(ctx context.Context, s step) error {
	switch s.kind {
	case kindPhase:
		return p.createPhase(ctx, s)
	case kindScope:
		return p.createScope(ctx, s)
	case kindList:
		return p.createList(ctx, s)
	case kindRule:
		return p.createRule(ctx, s)
	default:
		return fmt.Errorf("provision: %w: %d", ErrUnknownStep, s.kind)
	}
}

func (p *Provisioner) createPhase(ctx context.Context, s step) error {
	phase, err := p.client.Phases.Create(ctx, sdk.CreatePhaseRequest{
		Name:           s.name,
		Position:       s.position,
		ApplicableWhen: s.applicableWhen,
	})
	if err != nil {
		return fmt.Errorf("provision phase %q: %w", s.name, err)
	}
	p.phases[s.name] = phase.ID
	return nil
}

func (p *Provisioner) createScope(ctx context.Context, s step) error {
	scope, err := p.client.Scopes.Create(ctx, sdk.CreateScopeRequest{
		Name:     s.name,
		JSONPath: s.jsonPath,
	})
	if err != nil {
		return fmt.Errorf("provision scope %q: %w", s.name, err)
	}
	p.scopes[s.name] = scope.ID
	return nil
}

func (p *Provisioner) createList(ctx context.Context, s step) error {
	list, err := p.client.Lists.Create(ctx, sdk.CreateListRequest{
		Name:        s.name,
		Description: "",
	})
	if err != nil {
		return fmt.Errorf("provision list %q: %w", s.name, err)
	}
	p.lists[s.name] = list.ID
	return nil
}

func (p *Provisioner) createRule(ctx context.Context, s step) error {
	rb := s.rule
	phaseID, ok := p.phases[rb.phaseName]
	if !ok {
		return fmt.Errorf(
			"provision rule %q: %w: %s", rb.name, ErrPhaseNotFound, rb.phaseName,
		)
	}

	req := sdk.CreateRuleRequest{
		Name:           rb.name,
		PhaseID:        phaseID,
		Priority:       rb.priority,
		ApplicableWhen: rb.applicable,
		Expression:     rb.expression,
		Actions:        rb.buildActions(),
		Enabled:        true,
		DryRun:         rb.dryRun,
	}

	rule, err := p.client.Rules.Create(ctx, req)
	if err != nil {
		return fmt.Errorf("provision rule %q: %w", rb.name, err)
	}
	p.rules[rb.name] = rule.ID
	return nil
}

// deleteAll deletes resources by ID using the given function.
func deleteAll(
	ctx context.Context,
	deleteFn func(context.Context, string) error,
	ids map[string]string,
	record func(error),
) {
	for name, id := range ids {
		if err := deleteFn(ctx, id); err != nil {
			record(fmt.Errorf("deleting %s (%s): %w", name, id, err))
		}
	}
}
