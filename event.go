// Package testtally turns the line-delimited JSON that "go test -json"
// produces into a small summary: how many tests ran, which ones failed,
// and which ones were slow. The parsing and summarizing steps are kept
// as pure functions so they can be tested against fixed input without
// ever invoking the go tool.
package testtally

import (
	"encoding/json"
	"fmt"
)

// Event mirrors one line of "go test -json" output.
type Event struct {
	Action  string  `json:"Action"`
	Package string  `json:"Package"`
	Test    string  `json:"Test"`
	Elapsed float64 `json:"Elapsed"`
	Output  string  `json:"Output"`
}

// Recognized values of Event.Action that mark a test as finished.
const (
	ActionPass = "pass"
	ActionFail = "fail"
	ActionSkip = "skip"
)

// ParseEvent decodes a single line of "go test -json" output.
//
// It is pure: given the same bytes it always returns the same Event or
// the same error, and it does no I/O of its own. Callers that read from
// a file or pipe should split on newlines first and call this per line.
func ParseEvent(line []byte) (Event, error) {
	var e Event
	if err := json.Unmarshal(line, &e); err != nil {
		return Event{}, fmt.Errorf("testtally: parse event: %w", err)
	}
	return e, nil
}
