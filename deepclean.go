package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/karrick/godirwalk"
	"github.com/tj/go-spin"
)

const defaultTargets = "node_modules,.bundle,target"

var targets = flag.String("target", defaultTargets, "dirs to scan for")
var sorted = flag.Bool("sort", false, "sort output")
var _targets []string

func main() {
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage: %s [options] [dir]\n", os.Args[0])
		fmt.Fprintf(flag.CommandLine.Output(), "\nOptions:\n")
		flag.PrintDefaults()
	}
	flag.Parse()
	_targets = strings.Split(*targets, ",")

	dirname := "."
	if len(flag.Args()) >= 1 {
		dirname = flag.Arg(0)
	}

	res := scan(dirname)
	printResults(res)
}

func scan(dirname string) <-chan result {
	resultsChan := make(chan result)
	go func() {
		var wg sync.WaitGroup
		err := godirwalk.Walk(dirname, &godirwalk.Options{
			Unsorted: true,
			Callback: func(path string, de *godirwalk.Dirent) error {
				var isMatch = isTarget(path) && de.IsDir()
				if isMatch {
					wg.Add(1)
					go dirStatter(path, resultsChan, &wg)
					return filepath.SkipDir
				}
				return nil
			},
			ErrorCallback: func(path string, err error) godirwalk.ErrorAction {
				fmt.Fprintf(os.Stderr, "ERROR: %s\n", err)
				return godirwalk.SkipNode
			},
		})

		if err != nil {
			fmt.Fprintf(os.Stderr, "FATAL ERROR: %s\n", err)
			os.Exit(1)
		}
		wg.Wait()
		close(resultsChan)
	}()
	return resultsChan
}

func isTarget(path string) bool {
	for i := range _targets {
		if _targets[i] == filepath.Base(path) {
			return true
		}
	}
	return false
}

func dirStatter(path string, resultsChan chan result, wg *sync.WaitGroup) {
	defer wg.Done()
	r, _ := dirStats(path) // TODO: handle err
	resultsChan <- r
}

func dirStats(path string) (result, error) {
	var t totals
	err := godirwalk.Walk(path, &godirwalk.Options{
		Unsorted: true,
		Callback: func(path string, de *godirwalk.Dirent) error {
			fi, err := os.Stat(path)
			if err != nil {
				return err
			}
			t.numFiles++
			t.bytes += uint64(fi.Size())
			return nil
		},
	})
	return result{path: path, totals: t}, err
}

// http://flummox-engineering.blogspot.com/2015/05/how-to-check-if-file-is-in-git.html

func printResults(res <-chan result) {
	var rs results
	var done = false
	if *sorted {
		go func() {
			s := spin.New()
			for !done {
				fmt.Fprintf(
					os.Stderr,
					"\r%v %s", s.Next(), strings.Repeat(".", len(rs)),
				)
				time.Sleep(100 * time.Millisecond)
			}
		}()
	}

	for r := range res {
		if !*sorted {
			fmt.Println(r)
		}
		rs = append(rs, r)
	}
	done = true

	if *sorted {
		sort.Slice(rs, func(i, j int) bool {
			return rs[i].numFiles > rs[j].numFiles
		})
		fmt.Fprintf(os.Stderr, "\r√\n")
		for _, r := range rs {
			fmt.Println(r)
		}
	}

	total := rs.Sum()
	fmt.Fprintf(os.Stderr,
		"\nTotal cleanable discovered: %d files, %v\n",
		total.numFiles, humanize.Bytes(total.bytes),
	)
}
