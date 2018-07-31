package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/dustin/go-humanize"
	"github.com/karrick/godirwalk"
)

var troubleMakers = [...]string{
	"node_modules",
	".bundle",
	"target",
}

type result struct {
	path            string
	numFiles, bytes uint64
}

func (r result) String() string {
	return fmt.Sprintf(
		"%s (%d files, %s)", r.path, r.numFiles, humanize.Bytes(r.bytes))
}

func main() {
	dirname := "."
	if len(os.Args) > 1 {
		dirname = os.Args[1]
	}

	res, err := scan(dirname)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
	for r := range res {
		fmt.Println(r)
	}
}

// TODO: consider use channel without subproc first, then switch to
// scanner than does new chan for each subdir? this would be uniformly parallel...
// but maybe too much overhead switching so often, look at how rust fd does it?
//
// using producer-consumer pattern would be cleaner conceptually than waitgroups..
// https://stackoverflow.com/questions/38170852/is-this-an-idiomatic-worker-thread-pool-in-go/38172204#38172204
//
// maybe wait until after tho so can compare benchmarks

func scan(dirname string) (<-chan result, error) {
	var wg sync.WaitGroup
	res := make(chan result)
	go func() {
		_ = godirwalk.Walk(dirname, &godirwalk.Options{
			Unsorted: true, // do these in order? wont retrurn in order so doesnt matter!
			Callback: func(path string, de *godirwalk.Dirent) error {
				var isMatch = isTrouble(path) && de.IsDir()
				if isMatch {
					// spawn a sub-walker to get the dirstats for subtree
					wg.Add(1)
					go func() {
						defer wg.Done()
						files, size, _ := dirStats(path) // TODO: handle err
						res <- result{
							path:     path,
							numFiles: files,
							bytes:    size,
						}
					}()
					// tell main walker to stop walking this subtree
					return filepath.SkipDir
				}
				return nil
			},
		})
		// TODO NEED TO DO SOMETHING WITH ERRRRRRS
		// if err != nil {
		// 	return nil, err
		// }
		wg.Wait()
		close(res)
	}()
	return res, nil
}

func isTrouble(path string) bool {
	for i := range troubleMakers {
		if troubleMakers[i] == filepath.Base(path) {
			return true
		}
	}
	return false
}

func dirStats(path string) (files, size uint64, err error) {
	err = godirwalk.Walk(path, &godirwalk.Options{
		Unsorted: true,
		Callback: func(path string, de *godirwalk.Dirent) error {
			fi, err := os.Stat(path)
			if err != nil {
				return err
			}
			files++
			size += uint64(fi.Size())
			return nil
		},
	})
	return files, size, err
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
