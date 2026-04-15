package provision

import sdk "github.com/intelmesh/intelmesh-sdk-go"

// RuleBuilder provides a fluent API for constructing rules within a Provisioner.
type RuleBuilder struct {
	parent     *Provisioner
	name       string
	phaseName  string
	priority   int
	applicable string
	expression string
	decision   *sdk.Decision
	score      *sdk.ScoreOperation
	flow       string
	mutations  []mutationSpec
	dryRun     bool
}

type mutationSpec struct {
	mutationType string
	listName     string
	valuePath    string
	amount       *int64
}

func newRuleBuilder(parent *Provisioner, name string) *RuleBuilder {
	return &RuleBuilder{
		parent:     parent,
		name:       name,
		phaseName:  "",
		priority:   0,
		applicable: "",
		expression: "",
		decision:   nil,
		score:      nil,
		flow:       "",
		mutations:  nil,
		dryRun:     false,
	}
}

// InPhase sets the phase for the rule (by provisioner name).
func (r *RuleBuilder) InPhase(name string) *RuleBuilder {
	r.phaseName = name
	return r
}

// Priority sets the rule priority.
func (r *RuleBuilder) Priority(p int) *RuleBuilder {
	r.priority = p
	return r
}

// ApplicableWhen sets the applicability expression.
func (r *RuleBuilder) ApplicableWhen(expr string) *RuleBuilder {
	r.applicable = expr
	return r
}

// When sets the matching expression.
func (r *RuleBuilder) When(expr string) *RuleBuilder {
	r.expression = expr
	return r
}

// Decide sets the decision action and severity.
func (r *RuleBuilder) Decide(action, severity string) *RuleBuilder {
	r.decision = &sdk.Decision{
		Action:   action,
		Severity: severity,
		Metadata: nil,
	}
	return r
}

// AddScore sets a score delta for the rule.
func (r *RuleBuilder) AddScore(delta int64) *RuleBuilder {
	r.score = &sdk.ScoreOperation{Add: delta}
	return r
}

// Halt sets flow to halt.
func (r *RuleBuilder) Halt() *RuleBuilder {
	r.flow = "halt"
	return r
}

// Continue sets flow to continue.
func (r *RuleBuilder) Continue() *RuleBuilder {
	r.flow = "continue"
	return r
}

// SkipPhase sets flow to skip_phase.
func (r *RuleBuilder) SkipPhase() *RuleBuilder {
	r.flow = "skip_phase"
	return r
}

// MutateList adds a list mutation to the rule.
func (r *RuleBuilder) MutateList(mutType, listName, valuePath string) *RuleBuilder {
	r.mutations = append(r.mutations, mutationSpec{
		mutationType: mutType,
		listName:     listName,
		valuePath:    valuePath,
		amount:       nil,
	})
	return r
}

// AddScoped appends a score.add mutation targeting scopeName with the
// value resolved from valuePath and the given amount (negative allowed).
func (r *RuleBuilder) AddScoped(scopeName, valuePath string, amount int64) *RuleBuilder {
	a := amount
	r.mutations = append(r.mutations, mutationSpec{
		mutationType: "score.add",
		listName:     scopeName,
		valuePath:    valuePath,
		amount:       &a,
	})
	return r
}

// SetScoped appends a score.set mutation. Use amount=0 to reset.
func (r *RuleBuilder) SetScoped(scopeName, valuePath string, amount int64) *RuleBuilder {
	a := amount
	r.mutations = append(r.mutations, mutationSpec{
		mutationType: "score.set",
		listName:     scopeName,
		valuePath:    valuePath,
		amount:       &a,
	})
	return r
}

// DryRun enables dry-run mode for the rule.
func (r *RuleBuilder) DryRun() *RuleBuilder {
	r.dryRun = true
	return r
}

// Done returns to the parent provisioner after registering the rule.
func (r *RuleBuilder) Done() *Provisioner {
	r.parent.steps = append(r.parent.steps, step{
		kind: kindRule, name: r.name, position: 0,
		jsonPath: "", rule: r,
	})
	return r.parent
}

// buildActions constructs the SDK Actions from the builder state.
func (r *RuleBuilder) buildActions() sdk.Actions {
	actions := sdk.Actions{
		Decision:  r.decision,
		Score:     r.score,
		Flow:      r.flow,
		Mutations: nil,
	}

	if len(r.mutations) > 0 {
		muts := make([]sdk.Mutation, 0, len(r.mutations))
		for _, m := range r.mutations {
			muts = append(muts, sdk.Mutation{
				Type:      m.mutationType,
				Target:    m.listName,
				ValuePath: m.valuePath,
				Amount:    m.amount,
			})
		}
		actions.Mutations = muts
	}

	return actions
}
