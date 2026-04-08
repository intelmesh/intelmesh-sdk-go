package testkit

import sdk "github.com/intelmesh/intelmesh-sdk-go"

// EventAssertion chains assertions on the result of an event send.
type EventAssertion struct {
	harness *Harness
	result  *sdk.IngestResult
	err     error
}

// ExpectDecision asserts the result has a specific decision.
func (a *EventAssertion) ExpectDecision(action, severity string) *EventAssertion {
	a.harness.t.Helper()

	if a.err != nil {
		a.harness.t.Fatalf("testkit: event error: %v", a.err)
	}
	if a.result.Decision == nil {
		a.harness.t.Errorf("testkit: expected decision %s/%s, got nil", action, severity)
		return a
	}
	if a.result.Decision.Action != action {
		a.harness.t.Errorf("testkit: decision action: got %q, want %q",
			a.result.Decision.Action, action)
	}
	if a.result.Decision.Severity != severity {
		a.harness.t.Errorf("testkit: decision severity: got %q, want %q",
			a.result.Decision.Severity, severity)
	}
	return a
}

// ExpectNoDecision asserts the result has no decision.
func (a *EventAssertion) ExpectNoDecision() *EventAssertion {
	a.harness.t.Helper()

	if a.err != nil {
		a.harness.t.Fatalf("testkit: event error: %v", a.err)
	}
	if a.result.Decision != nil {
		a.harness.t.Errorf("testkit: expected no decision, got %+v", a.result.Decision)
	}
	return a
}

// ExpectScore asserts the transient score value.
func (a *EventAssertion) ExpectScore(score int64) *EventAssertion {
	a.harness.t.Helper()

	if a.err != nil {
		a.harness.t.Fatalf("testkit: event error: %v", a.err)
	}
	if a.result.TransientScore != score {
		a.harness.t.Errorf("testkit: score: got %d, want %d",
			a.result.TransientScore, score)
	}
	return a
}

// Then returns the harness for chaining further steps.
func (a *EventAssertion) Then() *Harness {
	return a.harness
}
