package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/mroth/deepclean/internal/deepclean"
	"github.com/tj/go-spin"
)

const defaultTargets = "node_modules,.bundle,target"

var (
	targetStr = flag.String("target", defaultTargets, "dirs to scan for")
	sorted    = flag.Bool("sort", false, "sort output")
)

func main() {
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage: %s [options] [dir]\n", os.Args[0])
		fmt.Fprintf(flag.CommandLine.Output(), "\nOptions:\n")
		flag.PrintDefaults()
	}
	flag.Parse()

	targets := strings.Split(*targetStr, ",")
	dirname := flag.Arg(0)
	if dirname == "" {
		dirname = "."
	}

	scanner := deepclean.Scan(dirname, targets)
	displayResults(scanner)
}

func displayResults(scanner *deepclean.Scanner) {
	var rs []deepclean.Result

	// if going to display sorted results, we wont display until scan is
	// complete, so display a progress monitor.
	ctx, cancelSpinner := context.WithCancel(context.Background())
	if *sorted {
		go func() {
			s := spin.New()
			t := time.NewTicker(100 * time.Millisecond)
			defer t.Stop()
			for {
				select {
				case <-t.C:
					fmt.Fprintf(os.Stderr,
						"\r%v %s", s.Next(), strings.Repeat(".", len(rs)),
					)
				case <-ctx.Done():
					return
				}
			}
		}()
	}

	// Iterate over results.  If not sorted, display as they come in,
	// otherwise just collect them for later.
	for r := range scanner.C {
		rs = append(rs, r)
		if !*sorted {
			fmt.Println(formatResult(r))
		}
	}
	cancelSpinner()

	// handle fatal scan errors
	if err := scanner.Err(); err != nil {
		if *sorted {
			fmt.Fprintf(os.Stderr, "\r") // clear spinner status
		}
		fmt.Fprintf(os.Stderr, "Fatal: %v\n", err)
		os.Exit(1)
	}

	// display sorted results (if requested)
	if *sorted {
		sort.Slice(rs, func(i, j int) bool {
			return rs[i].Stats.Files > rs[j].Stats.Files
		})
		fmt.Fprintf(os.Stderr, "\r√\n") // spinner status
		for _, r := range rs {
			fmt.Println(formatResult(r))
		}
	}

	// display totals
	total := deepclean.Aggregate(rs...)
	fmt.Fprintf(os.Stderr,
		"\nTotal cleanable discovered: %d files, %v\n",
		total.Files, humanize.Bytes(total.Bytes),
	)
}

func formatResult(r deepclean.Result) string {
	return fmt.Sprintf(
		"%7d\t%7s\t%s", r.Stats.Files, humanize.Bytes(r.Stats.Bytes), r.Path)
}
