package builders

import intelmesh "github.com/intelmesh/intelmesh-sdk-go"

// RuleBuilder builds CreateRuleRequest payloads fluently.
type RuleBuilder struct {
	name           string
	phaseID        string
	priority       int
	applicableWhen string
	expression     string
	actions        intelmesh.Actions
	enabled        bool
	dryRun         bool
}

// Rule creates a new RuleBuilder with the given rule name.
func Rule(name string) *RuleBuilder {
	return &RuleBuilder{
		name:    name,
		enabled: true,
	}
}

// Phase sets the phase ID for the rule.
func (b *RuleBuilder) Phase(phaseID string) *RuleBuilder {
	b.phaseID = phaseID

	return b
}

// Priority sets the execution priority for the rule.
func (b *RuleBuilder) Priority(priority int) *RuleBuilder {
	b.priority = priority

	return b
}

// ApplicableWhen sets the applicability expression for the rule.
func (b *RuleBuilder) ApplicableWhen(expr string) *RuleBuilder {
	b.applicableWhen = expr

	return b
}

// Expression sets the condition expression for the rule.
func (b *RuleBuilder) Expression(expr string) *RuleBuilder {
	b.expression = expr

	return b
}

// Decide sets a decision action on the rule.
func (b *RuleBuilder) Decide(action string, severity string) *RuleBuilder {
	b.actions.Decision = &intelmesh.Decision{
		Action:   action,
		Severity: severity,
	}

	return b
}

// AddScore sets a score increment action on the rule.
func (b *RuleBuilder) AddScore(delta int64) *RuleBuilder {
	b.actions.Score = &intelmesh.ScoreOperation{Add: delta}

	return b
}

// Flow sets the flow control action on the rule.
func (b *RuleBuilder) Flow(flow string) *RuleBuilder {
	b.actions.Flow = flow

	return b
}

// Mutate adds a mutation action to the rule.
func (b *RuleBuilder) Mutate(mutationType string, target string, valuePath string) *RuleBuilder {
	b.actions.Mutations = append(b.actions.Mutations, intelmesh.Mutation{
		Type:      mutationType,
		Target:    target,
		ValuePath: valuePath,
	})

	return b
}

// Enabled sets whether the rule is enabled.
func (b *RuleBuilder) Enabled(enabled bool) *RuleBuilder {
	b.enabled = enabled

	return b
}

// DryRun sets whether the rule runs in dry-run mode.
func (b *RuleBuilder) DryRun(dryRun bool) *RuleBuilder {
	b.dryRun = dryRun

	return b
}

// Build constructs the final CreateRuleRequest.
func (b *RuleBuilder) Build() intelmesh.CreateRuleRequest {
	return intelmesh.CreateRuleRequest{
		Name:           b.name,
		PhaseID:        b.phaseID,
		Priority:       b.priority,
		ApplicableWhen: b.applicableWhen,
		Expression:     b.expression,
		Actions:        b.actions,
		Enabled:        b.enabled,
		DryRun:         b.dryRun,
	}
}
