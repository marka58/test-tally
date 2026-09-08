# test-tally

`go test -json` is great for tooling and terrible for reading. A single
`./...` run against a mid-sized package can produce thousands of lines of
JSON, one per subtest event, with the actual failures buried somewhere in
the middle. test-tally reads that stream and prints the three things you
usually want: how many tests passed, which ones failed, and which ones
were slow.

## Install

```
go install github.com/marka58/test-tally/cmd/testtally@latest
```

## Use

Pipe `go test -json` straight into it:

```
go test -json ./... | testtally
```

```
42 passed, 1 failed, 2 skipped (1.284s)

failed:
  github.com/example/widget TestParseConfig

slowest:
  310ms    github.com/example/widget TestSlowIntegration
  120ms    github.com/example/widget TestParseConfig
  40ms     github.com/example/widget TestBasic
```

The exit code is 1 if any test failed, 0 otherwise, so it works as a
drop-in replacement for `go test` in a CI step that only cares about
pass/fail.

You can also read a saved log instead of stdin:

```
go test -json ./... > run.jsonl
testtally -in run.jsonl
```

## Library

The parsing and tallying logic lives in the root package and is meant to
be used on its own. The two functions that matter are pure: they take a
value and return a value, with no file or network access, so tests for
code built on top of this package don't need to shell out to `go test`
at all.

```go
events, err := testtally.ReadEvents(r) // the one function that does I/O
if err != nil {
    log.Fatal(err)
}

summary := testtally.Summarize(events) // pure: []Event in, Summary out
fmt.Println(summary.Passed, summary.Failed, summary.Skipped)
```

`ParseEvent([]byte) (Event, error)` and `Summarize([]Event) Summary` are
the two functions to reach for in a test: build the input by hand,
assert on the output, no fixtures or subprocesses needed.

## Status

Early. The JSON event format and the pass/fail/slowest summary work.
Not yet covered: flaky-test detection across repeated runs, package-level
breakdowns, and output formats other than plain text.

## License

MIT, see LICENSE.
