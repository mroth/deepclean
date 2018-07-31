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

func main() {
	dirname := "."
	if len(os.Args) > 1 {
		dirname = os.Args[1]
	}

	scan(dirname)
}

var wg sync.WaitGroup

func scan(dirname string) {
	err := godirwalk.Walk(dirname, &godirwalk.Options{Unsorted: true, Callback: matcher})
	wg.Wait()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}

func matcher(path string, de *godirwalk.Dirent) error {
	var isMatch = isTrouble(path) && de.IsDir()
	if isMatch {
		wg.Add(1)
		go func() {
			files, size, _ := dirStats(path) // TODO: handle err
			fmt.Printf("%s (%d files, %s)\n", path, files, humanize.Bytes(size))
			wg.Done()
		}()
		return filepath.SkipDir
	}

	return nil
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
