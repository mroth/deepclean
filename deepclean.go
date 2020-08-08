package deepclean

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/karrick/godirwalk"
)

// Scan walks the path searching for directories matching the targets strings,
// and then initiates a DirStats on the directory, returning the Results as they
// occur on the returned channel.
func Scan(path string, targets []string) <-chan Result {
	resultsChan := make(chan Result)
	go func() {
		var wg sync.WaitGroup
		err := godirwalk.Walk(path, &godirwalk.Options{
			Unsorted: true,
			Callback: func(path string, de *godirwalk.Dirent) error {
				// Positives are directories matching any target string. For
				// each match, spawn background goroutine to gather stats, but
				// skip further walking of the subdir in this primary scan
				// thread.
				if de.IsDir() && inTargets(targets, path) {
					wg.Add(1)
					go func() {
						defer wg.Done()
						s, _ := Stat(path) // TODO: handle err
						resultsChan <- Result{Path: path, Stats: s}
					}()
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

func inTargets(targets []string, path string) bool {
	for _, t := range targets {
		if t == filepath.Base(path) {
			return true
		}
	}
	return false
}

// Stat walks the directory at path collecting the aggregate DirStats.
func Stat(path string) (DirStats, error) {
	var t DirStats
	err := godirwalk.Walk(path, &godirwalk.Options{
		Unsorted: true,
		Callback: func(path string, de *godirwalk.Dirent) error {
			fi, err := os.Stat(path)
			if err != nil {
				return err
			}
			t.Files++
			t.Bytes += uint64(fi.Size())
			return nil
		},
	})
	return t, err
}
