package testtally

import (
	"reflect"
	"testing"
	"time"
)

func TestSummarizeCounts(t *testing.T) {
	events := []Event{
		{Action: ActionPass, Package: "widget", Test: "TestA", Elapsed: 0.01},
		{Action: ActionFail, Package: "widget", Test: "TestB", Elapsed: 0.02},
		{Action: ActionSkip, Package: "widget", Test: "TestC", Elapsed: 0},
		{Action: "output", Package: "widget", Output: "=== RUN   TestA\n"},
		{Action: ActionPass, Package: "widget", Elapsed: 0.03}, // package-level, no Test
	}

	s := Summarize(events)

	if s.Passed != 1 {
		t.Errorf("Passed = %d, want 1", s.Passed)
	}
	if s.Failed != 1 {
		t.Errorf("Failed = %d, want 1", s.Failed)
	}
	if s.Skipped != 1 {
		t.Errorf("Skipped = %d, want 1", s.Skipped)
	}
	if want := 3; s.Total() != want {
		t.Errorf("Total() = %d, want %d", s.Total(), want)
	}
}

func TestSummarizeDuration(t *testing.T) {
	// Duration only accrues for pass/fail, and skip contributes nothing
	// even when the event carries a nonzero Elapsed.
	events := []Event{
		{Action: ActionPass, Package: "widget", Test: "TestA", Elapsed: 0.5},
		{Action: ActionFail, Package: "widget", Test: "TestB", Elapsed: 0.25},
		{Action: ActionSkip, Package: "widget", Test: "TestC", Elapsed: 0.75},
	}

	s := Summarize(events)

	want := 750 * time.Millisecond
	if s.Duration != want {
		t.Errorf("Duration = %s, want %s", s.Duration, want)
	}
}

func TestSummarizeFailures(t *testing.T) {
	events := []Event{
		{Action: ActionPass, Package: "widget", Test: "TestA", Elapsed: 0.01},
		{Action: ActionFail, Package: "widget", Test: "TestB", Elapsed: 0.02},
		{Action: ActionFail, Package: "other", Test: "TestZ", Elapsed: 0.03},
	}

	s := Summarize(events)

	want := []TestResult{
		{Package: "widget", Test: "TestB", Elapsed: 20 * time.Millisecond},
		{Package: "other", Test: "TestZ", Elapsed: 30 * time.Millisecond},
	}
	if !reflect.DeepEqual(s.Failures, want) {
		t.Errorf("Failures = %+v, want %+v", s.Failures, want)
	}
}

func TestSummarizeSlowestOrderedDescending(t *testing.T) {
	events := []Event{
		{Action: ActionPass, Package: "widget", Test: "TestFast", Elapsed: 0.01},
		{Action: ActionPass, Package: "widget", Test: "TestSlow", Elapsed: 0.30},
		{Action: ActionFail, Package: "widget", Test: "TestMid", Elapsed: 0.15},
	}

	s := Summarize(events)

	want := []TestResult{
		{Package: "widget", Test: "TestSlow", Elapsed: 300 * time.Millisecond},
		{Package: "widget", Test: "TestMid", Elapsed: 150 * time.Millisecond},
		{Package: "widget", Test: "TestFast", Elapsed: 10 * time.Millisecond},
	}
	if !reflect.DeepEqual(s.Slowest, want) {
		t.Errorf("Slowest = %+v, want %+v", s.Slowest, want)
	}
}

func TestSummarizeSlowestCapsAtFive(t *testing.T) {
	events := make([]Event, 0, 8)
	for i := 0; i < 8; i++ {
		events = append(events, Event{
			Action:  ActionPass,
			Package: "widget",
			Test:    "Test",
			Elapsed: float64(i) * 0.01,
		})
	}

	s := Summarize(events)

	if len(s.Slowest) != 5 {
		t.Fatalf("len(Slowest) = %d, want 5", len(s.Slowest))
	}
	if s.Slowest[0].Elapsed != 70*time.Millisecond {
		t.Errorf("Slowest[0].Elapsed = %s, want %s", s.Slowest[0].Elapsed, 70*time.Millisecond)
	}
}

func TestSummarizeIgnoresEventsWithoutTest(t *testing.T) {
	events := []Event{
		{Action: ActionPass, Package: "widget", Elapsed: 1.28}, // package-level pass
		{Action: "run", Package: "widget", Test: "TestA"},      // not a terminal action
		{Action: "output", Package: "widget", Output: "ok"},
	}

	s := Summarize(events)

	if s.Total() != 0 {
		t.Errorf("Total() = %d, want 0", s.Total())
	}
	if s.Duration != 0 {
		t.Errorf("Duration = %s, want 0", s.Duration)
	}
}

func TestSummarizeEmptyInput(t *testing.T) {
	s := Summarize(nil)

	if s.Total() != 0 {
		t.Errorf("Total() = %d, want 0", s.Total())
	}
	if s.Slowest != nil {
		t.Errorf("Slowest = %+v, want nil", s.Slowest)
	}
	if s.Failures != nil {
		t.Errorf("Failures = %+v, want nil", s.Failures)
	}
}
