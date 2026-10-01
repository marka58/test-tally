package testtally

import (
	"reflect"
	"testing"
)

func ev(action, pkg, test string) Event {
	return Event{Action: action, Package: pkg, Test: test}
}

func TestFlaky(t *testing.T) {
	tests := []struct {
		name string
		runs [][]Event
		want []FlakyTest
	}{
		{
			name: "no runs",
			runs: nil,
			want: nil,
		},
		{
			name: "stable pass and stable fail are not flaky",
			runs: [][]Event{
				{ev("pass", "p", "TestA"), ev("fail", "p", "TestB")},
				{ev("pass", "p", "TestA"), ev("fail", "p", "TestB")},
			},
			want: nil,
		},
		{
			name: "pass then fail is flaky",
			runs: [][]Event{
				{ev("pass", "p", "TestA")},
				{ev("fail", "p", "TestA")},
				{ev("pass", "p", "TestA")},
			},
			want: []FlakyTest{{Package: "p", Test: "TestA", Passed: 2, Failed: 1}},
		},
		{
			name: "skips are ignored",
			runs: [][]Event{
				{ev("skip", "p", "TestA")},
				{ev("pass", "p", "TestA")},
			},
			want: nil,
		},
		{
			name: "package-level events are ignored",
			runs: [][]Event{
				{ev("pass", "p", "")},
				{ev("fail", "p", "")},
			},
			want: nil,
		},
		{
			name: "same test name in different packages is tracked separately",
			runs: [][]Event{
				{ev("pass", "b", "TestX"), ev("pass", "a", "TestX")},
				{ev("fail", "b", "TestX"), ev("pass", "a", "TestX")},
			},
			want: []FlakyTest{{Package: "b", Test: "TestX", Passed: 1, Failed: 1}},
		},
		{
			name: "results are sorted by package then test",
			runs: [][]Event{
				{ev("pass", "b", "TestA"), ev("pass", "a", "TestZ"), ev("pass", "a", "TestY")},
				{ev("fail", "b", "TestA"), ev("fail", "a", "TestZ"), ev("fail", "a", "TestY")},
			},
			want: []FlakyTest{
				{Package: "a", Test: "TestY", Passed: 1, Failed: 1},
				{Package: "a", Test: "TestZ", Passed: 1, Failed: 1},
				{Package: "b", Test: "TestA", Passed: 1, Failed: 1},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Flaky(tt.runs)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Flaky() = %+v, want %+v", got, tt.want)
			}
		})
	}
}
