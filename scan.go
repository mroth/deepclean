package deepclean

import (
	"os"
	"path/filepath"
	"sync"

	"github.com/karrick/godirwalk"
)

// Scanner contains fields to access the results of an ongoing Scan.
//
// If the underlying filepath Walk encounters a fatal error, the results
// channel will be closed and Err() will return a non-nil value. Always
// drain (*Scanner).C prior to checking Err().
//
// If the underlying filepath Walk encounters a non-fatal error, the
// walked directory will be silently skipped (for now).
type Scanner struct {
	C   <-chan Result
	err error
}

// Err returns the error status of the underlying filepath Walk performed by the
// scanner. This will only be set once the Walk has exited, indicated by the
// Results channel being closed.
func (s Scanner) Err() error {
	return s.err
}

// Scan walks the path searching for directories matching the targets strings,
// and then initiates a DirStats on the directory, returning the Results as they
// occur on the returned channel.
func Scan(path string, targets []string) *Scanner {
	resultsChan := make(chan Result)
	scanner := Scanner{C: resultsChan}

	go func() {
		defer close(resultsChan)
		var wg sync.WaitGroup
		scanner.err = godirwalk.Walk(path, &godirwalk.Options{
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
						if s, err := Stat(path); err == nil {
							resultsChan <- Result{Path: path, Stats: s}
						}
					}()
					return filepath.SkipDir
				}
				return nil
			},
			ErrorCallback: func(path string, err error) godirwalk.ErrorAction {
				// TODO: non-fatal error, log only if verbose
				// fmt.Fprintf(os.Stderr, "ERROR: %s\n", err)
				return godirwalk.SkipNode
			},
		})
		wg.Wait()
	}()
	return &scanner
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
