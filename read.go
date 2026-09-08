package testtally

import (
	"bufio"
	"fmt"
	"io"
)

// ReadEvents reads newline-delimited "go test -json" output from r and
// parses each line with ParseEvent.
//
// This is the one place in the package that touches I/O; it exists so
// that ParseEvent and Summarize can stay pure and get tested directly
// against []byte and []Event values instead of readers. Blank lines are
// skipped. A malformed line stops parsing and returns the events found
// so far alongside the error, so a caller can still report partial
// progress on a truncated log.
func ReadEvents(r io.Reader) ([]Event, error) {
	var events []Event
	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)

	lineNum := 0
	for scanner.Scan() {
		lineNum++
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}
		e, err := ParseEvent(line)
		if err != nil {
			return events, fmt.Errorf("testtally: line %d: %w", lineNum, err)
		}
		events = append(events, e)
	}
	if err := scanner.Err(); err != nil {
		return events, fmt.Errorf("testtally: read: %w", err)
	}
	return events, nil
}
