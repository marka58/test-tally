// Command testtally reads "go test -json" output and prints a short
// pass/fail summary instead of the raw event stream.
//
// Usage:
//
//	go test -json ./... | testtally
//	testtally -in run.jsonl
package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/marka58/test-tally"
)

func main() {
	inPath := flag.String("in", "", "path to a go test -json log (default: stdin)")
	flag.Parse()

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
	printSummary(os.Stdout, summary)

	if summary.Failed > 0 {
		os.Exit(1)
	}
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
}
