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
	for i := range targets {
		if targets[i] == filepath.Base(path) {
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
