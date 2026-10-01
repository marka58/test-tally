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

packages:
  github.com/example/gadget  12 passed, 0 failed, 0 skipped (210ms)
  github.com/example/widget  30 passed, 1 failed, 2 skipped (1.074s)
```

The package breakdown is only printed when the run touched more than one
package; a single-package run already has the top line to summarize it.

The exit code is 1 if any test failed, 0 otherwise, so it works as a
drop-in replacement for `go test` in a CI step that only cares about
pass/fail.

You can also read a saved log instead of stdin:

```
go test -json ./... > run.jsonl
testtally -in run.jsonl
```

Pass `-json` to get the same summary as a single JSON object instead of
the plain-text report, for a CI step that wants to parse it rather than
scrape it:

```
go test -json ./... | testtally -json
```

```json
{"passed":42,"failed":1,"skipped":2,"durationNs":1284000000,"failures":[{"package":"github.com/example/widget","test":"TestParseConfig","elapsedNs":120000000}],"slowest":[...],"packages":[...]}
```

### Flaky tests

Run the suite several times, save each run, and pass the logs with
`-flaky`:

```
for i in 1 2 3; do go test -count=1 -json ./... > run$i.jsonl; done
testtally -flaky run1.jsonl run2.jsonl run3.jsonl
```

```
1 flaky in 3 runs:
  github.com/example/widget TestRetry (2 passed, 1 failed)
```

A test is flaky if it passed in at least one run and failed in at least
one. Tests that always fail, or that were only skipped in some runs, are
not reported. The exit code is 1 if any flaky test is found. `-json`
prints the list as a JSON array. In the library this is
`Flaky([][]Event) []FlakyTest`.

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

Early. The JSON event format, the pass/fail/slowest summary, the
per-package breakdown, JSON output, and flaky-test detection across
repeated runs work. Not yet covered: a quiet or fail-fast mode.

## License

MIT, see LICENSE.
