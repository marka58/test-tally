// Command testtally reads "go test -json" output and prints a short
// pass/fail summary instead of the raw event stream.
//
// Usage:
//
//	go test -json ./... | testtally
//	testtally -in run.jsonl
//	testtally -json | jq .
//	testtally -flaky run1.jsonl run2.jsonl run3.jsonl
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/marka58/test-tally"
)

func main() {
	inPath := flag.String("in", "", "path to a go test -json log (default: stdin)")
	jsonOut := flag.Bool("json", false, "print the summary as JSON instead of plain text")
	flaky := flag.Bool("flaky", false, "report tests that both pass and fail across the log files given as arguments")
	flag.Parse()

	if *flaky {
		os.Exit(runFlaky(flag.Args(), *jsonOut))
	}

	in := os.Stdin
	if *inPath != "" {
		f, err := os.Open(*inPath)
		if err != nil {
			fmt.Fprintln(os.Stderr, "testtally:", err)
			os.Exit(1)
		}
		defer f.Close()
		in = f
	}

	events, err := testtally.ReadEvents(in)
	if err != nil {
		fmt.Fprintln(os.Stderr, "testtally:", err)
		os.Exit(1)
	}

	summary := testtally.Summarize(events)
	if *jsonOut {
		if err := printSummaryJSON(os.Stdout, summary); err != nil {
			fmt.Fprintln(os.Stderr, "testtally:", err)
			os.Exit(1)
		}
	} else {
		printSummary(os.Stdout, summary)
	}

	if summary.Failed > 0 {
		os.Exit(1)
	}
}

// runFlaky reads one go test -json log per path, treats each as a separate
// run, and prints the tests whose result differed between runs. It returns
// the process exit code: 1 if any flaky test was found or on error.
func runFlaky(paths []string, asJSON bool) int {
	if len(paths) < 2 {
		fmt.Fprintln(os.Stderr, "testtally: -flaky needs at least two log files")
		return 1
	}

	var runs [][]testtally.Event
	for _, p := range paths {
		f, err := os.Open(p)
		if err != nil {
			fmt.Fprintln(os.Stderr, "testtally:", err)
			return 1
		}
		events, err := testtally.ReadEvents(f)
		f.Close()
		if err != nil {
			fmt.Fprintf(os.Stderr, "testtally: %s: %v\n", p, err)
			return 1
		}
		runs = append(runs, events)
	}

	flaky := testtally.Flaky(runs)
	if asJSON {
		if flaky == nil {
			flaky = []testtally.FlakyTest{}
		}
		if err := json.NewEncoder(os.Stdout).Encode(flaky); err != nil {
			fmt.Fprintln(os.Stderr, "testtally:", err)
			return 1
		}
	} else if len(flaky) == 0 {
		fmt.Printf("no flaky tests in %d runs\n", len(runs))
	} else {
		fmt.Printf("%d flaky in %d runs:\n", len(flaky), len(runs))
		for _, t := range flaky {
			fmt.Printf("  %s %s (%d passed, %d failed)\n", t.Package, t.Test, t.Passed, t.Failed)
		}
	}

	if len(flaky) > 0 {
		return 1
	}
	return 0
}

// printSummaryJSON writes the summary as a single JSON object, so a CI
// step can parse it instead of scraping the plain-text report.
func printSummaryJSON(w *os.File, s testtally.Summary) error {
	return json.NewEncoder(w).Encode(s)
}

func printSummary(w *os.File, s testtally.Summary) {
	fmt.Fprintf(w, "%d passed, %d failed, %d skipped (%s)\n",
		s.Passed, s.Failed, s.Skipped, s.Duration.Round(time.Millisecond))

	if len(s.Failures) > 0 {
		fmt.Fprintln(w, "\nfailed:")
		for _, f := range s.Failures {
			fmt.Fprintf(w, "  %s %s\n", f.Package, f.Test)
		}
	}

	if len(s.Slowest) > 0 {
		fmt.Fprintln(w, "\nslowest:")
		for _, r := range s.Slowest {
			fmt.Fprintf(w, "  %-8s %s %s\n", r.Elapsed.Round(time.Millisecond), r.Package, r.Test)
		}
	}

	if len(s.Packages) > 1 {
		fmt.Fprintln(w, "\npackages:")
		for _, p := range s.Packages {
			fmt.Fprintf(w, "  %-8s %d passed, %d failed, %d skipped (%s)\n",
				p.Package, p.Passed, p.Failed, p.Skipped, p.Duration.Round(time.Millisecond))
		}
	}
}
