package testtally

import "sort"

// FlakyTest is a test that both passed and failed across a set of runs of
// the same code.
type FlakyTest struct {
	Package string `json:"package"`
	Test    string `json:"test"`
	Passed  int    `json:"passed"`
	Failed  int    `json:"failed"`
}

// Flaky reports the tests that have at least one pass and at least one
// failure across runs, where each element of runs is the Events of one
// complete run.
//
// Skips are ignored: a test that is skipped in one run and passes in
// another is not flaky, just conditional. A test that fails every time is
// a plain failure and is not reported either. Like Summarize it is pure,
// and the result is sorted by package and then test name so output is
// stable between invocations.
func Flaky(runs [][]Event) []FlakyTest {
	type key struct{ pkg, test string }
	counts := make(map[key]*FlakyTest)

	for _, run := range runs {
		for _, e := range run {
			if e.Test == "" || (e.Action != ActionPass && e.Action != ActionFail) {
				continue
			}
			k := key{e.Package, e.Test}
			t, ok := counts[k]
			if !ok {
				t = &FlakyTest{Package: e.Package, Test: e.Test}
				counts[k] = t
			}
			if e.Action == ActionPass {
				t.Passed++
			} else {
				t.Failed++
			}
		}
	}

	var flaky []FlakyTest
	for _, t := range counts {
		if t.Passed > 0 && t.Failed > 0 {
			flaky = append(flaky, *t)
		}
	}
	sort.Slice(flaky, func(i, j int) bool {
		if flaky[i].Package != flaky[j].Package {
			return flaky[i].Package < flaky[j].Package
		}
		return flaky[i].Test < flaky[j].Test
	})
	return flaky
}
