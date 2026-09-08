package testtally

import (
	"sort"
	"time"
)

// TestResult is one finished test, extracted from a slice of Events.
type TestResult struct {
	Package string
	Test    string
	Elapsed time.Duration
}

// Summary is the outcome of tallying a run's Events.
type Summary struct {
	Passed   int
	Failed   int
	Skipped  int
	Duration time.Duration
	Failures []TestResult
	Slowest  []TestResult
}

// Total returns the number of tests accounted for in the summary.
func (s Summary) Total() int {
	return s.Passed + s.Failed + s.Skipped
}

// Summarize tallies a slice of Events into a Summary.
//
// It is pure: it only reads its argument and returns a new value, with
// no I/O, no clock access, and no shared state. Only events with a
// non-empty Test field and a terminal Action (pass, fail, or skip) are
// counted; package-level and "run"/"output" events are ignored. The
// Slowest field lists up to five tests, sorted by descending Elapsed.
func Summarize(events []Event) Summary {
	var s Summary
	var finished []TestResult

	for _, e := range events {
		if e.Test == "" {
			continue
		}
		elapsed := time.Duration(e.Elapsed * float64(time.Second))
		switch e.Action {
		case ActionPass:
			s.Passed++
			s.Duration += elapsed
			finished = append(finished, TestResult{e.Package, e.Test, elapsed})
		case ActionFail:
			s.Failed++
			s.Duration += elapsed
			r := TestResult{e.Package, e.Test, elapsed}
			finished = append(finished, r)
			s.Failures = append(s.Failures, r)
		case ActionSkip:
			s.Skipped++
			finished = append(finished, TestResult{e.Package, e.Test, elapsed})
		}
	}

	s.Slowest = slowest(finished, 5)
	return s
}

// slowest returns up to n results sorted by descending Elapsed, without
// mutating the input slice.
func slowest(results []TestResult, n int) []TestResult {
	sorted := make([]TestResult, len(results))
	copy(sorted, results)
	sort.SliceStable(sorted, func(i, j int) bool {
		return sorted[i].Elapsed > sorted[j].Elapsed
	})
	if len(sorted) > n {
		sorted = sorted[:n]
	}
	return sorted
}
