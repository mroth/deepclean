package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/dustin/go-humanize"
	"github.com/karrick/godirwalk"
)

var targets = [...]string{
	"node_modules",
	".bundle",
	"target",
}

func main() {
	dirname := "."
	if len(os.Args) > 1 {
		dirname = os.Args[1]
	}

	res := scan(dirname)

	var rs results
	for r := range res {
		fmt.Println(r)
		rs = append(rs, r)
	}

	total := rs.Sum()
	fmt.Fprintf(os.Stderr,
		"\nTotal cleanable discovered: %d files, %v\n",
		total.numFiles, humanize.Bytes(total.bytes),
	)
}

// TODO: consider use channel without subproc first, then switch to
// scanner than does new chan for each subdir? this would be uniformly parallel...
// but maybe too much overhead switching so often, look at how rust fd does it?
//
// using producer-consumer pattern would be cleaner conceptually than waitgroups..
// https://stackoverflow.com/questions/38170852/is-this-an-idiomatic-worker-thread-pool-in-go/38172204#38172204
//
// maybe wait until after tho so can compare benchmarks

func scan(dirname string) <-chan result {
	resultsChan := make(chan result)
	go func() {
		var wg sync.WaitGroup
		err := godirwalk.Walk(dirname, &godirwalk.Options{
			Unsorted: true, // do these in order? wont return in order so doesnt matter!
			Callback: func(path string, de *godirwalk.Dirent) error {
				var isMatch = isTarget(path) && de.IsDir()
				if isMatch {
					// spawn a sub-walker to get the dirstats for subtree
					wg.Add(1)
					go func() {
						defer wg.Done()
						r, _ := dirStats(path) // TODO: handle err
						resultsChan <- r
					}()
					// tell main walker to stop walking this subtree
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
	for i := range targets {
		if targets[i] == filepath.Base(path) {
			return true
		}
	}
	return false
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

// https://stackoverflow.com/questions/13422381/idiomatically-buffer-os-stdout

// When set to true, Walk skips sorting the list of immediate descendants
// for a directory, and simply visits each node in the order the operating
// system enumerated them. This will be more fast, but with the side effect
// that the traversal order may be different from one invocation to the
// next.
// Unsorted bool

// ScratchBuffer is an optional scratch buffer for Walk to use when reading
// directory entries, to reduce amount of garbage generation. Not all
// architectures take advantage of the scratch buffer.
// ScratchBuffer []byte

// http://flummox-engineering.blogspot.com/2015/05/how-to-check-if-file-is-in-git.html

// https://gobyexample.com/worker-pools
// https://stackoverflow.com/questions/44255814/concurrent-filesystem-scanning
