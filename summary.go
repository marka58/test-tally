package testtally

import (
	"encoding/json"
	"sort"
	"time"
)

// TestResult is one finished test, extracted from a slice of Events.
type TestResult struct {
	Package string        `json:"package"`
	Test    string        `json:"test"`
	Elapsed time.Duration `json:"elapsedNs"`
}

// Summary is the outcome of tallying a run's Events.
type Summary struct {
	Passed   int             `json:"passed"`
	Failed   int             `json:"failed"`
	Skipped  int             `json:"skipped"`
	Duration time.Duration   `json:"durationNs"`
	Failures []TestResult    `json:"failures"`
	Slowest  []TestResult    `json:"slowest"`
	Packages []PackageResult `json:"packages"`
}

// MarshalJSON encodes the summary with empty arrays in place of Go's zero
// value for an unset slice. Summarize leaves Failures, Slowest, and
// Packages nil on an empty run; a CI step decoding this JSON shouldn't
// have to special-case null versus [].
func (s Summary) MarshalJSON() ([]byte, error) {
	type alias Summary
	a := alias(s)
	if a.Failures == nil {
		a.Failures = []TestResult{}
	}
	if a.Slowest == nil {
		a.Slowest = []TestResult{}
	}
	if a.Packages == nil {
		a.Packages = []PackageResult{}
	}
	return json.Marshal(a)
}

// Total returns the number of tests accounted for in the summary.
func (s Summary) Total() int {
	return s.Passed + s.Failed + s.Skipped
}

// PackageResult is the pass/fail/skip tally for a single package, extracted
// from a slice of Events.
type PackageResult struct {
	Package  string        `json:"package"`
	Passed   int           `json:"passed"`
	Failed   int           `json:"failed"`
	Skipped  int           `json:"skipped"`
	Duration time.Duration `json:"durationNs"`
}

// Total returns the number of tests accounted for in the package.
func (p PackageResult) Total() int {
	return p.Passed + p.Failed + p.Skipped
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
	pkgIndex := make(map[string]int)

	pkg := func(name string) *PackageResult {
		i, ok := pkgIndex[name]
		if !ok {
			i = len(s.Packages)
			pkgIndex[name] = i
			s.Packages = append(s.Packages, PackageResult{Package: name})
		}
		return &s.Packages[i]
	}

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
			p := pkg(e.Package)
			p.Passed++
			p.Duration += elapsed
		case ActionFail:
			s.Failed++
			s.Duration += elapsed
			r := TestResult{e.Package, e.Test, elapsed}
			finished = append(finished, r)
			s.Failures = append(s.Failures, r)
			p := pkg(e.Package)
			p.Failed++
			p.Duration += elapsed
		case ActionSkip:
			s.Skipped++
			finished = append(finished, TestResult{e.Package, e.Test, elapsed})
			pkg(e.Package).Skipped++
		}
	}

	s.Slowest = slowest(finished, 5)
	sort.SliceStable(s.Packages, func(i, j int) bool {
		return s.Packages[i].Package < s.Packages[j].Package
	})
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
