package deepclean

import (
	"context"
	"io/fs"
	"runtime"
	"sync"
)

// ScanTask contains fields to access the results of an ongoing Scan.
//
// If the underlying filesystem Walk encounters a fatal error, the results
// channel will be closed and Err() will return a non-nil value. Always
// drain ScanTask.C prior to checking Err().
//
// If the underlying filesystem Walk encounters a non-fatal error, the
// walked directory will be silently skipped.
type ScanTask struct {
	C   <-chan Result
	err error
}

// Err returns the error status of the underlying filesystem Walk performed by the
// scanner. This will only be set once the Walk has exited, indicated by the
// Results channel being closed.
func (s ScanTask) Err() error {
	return s.err
}

// Scan walks the filesystem searching for directories matching the targets strings,
// and then initiates a DirStats on the directory, returning the Results as they
// occur on ScanTask.C.
func Scan(ctx context.Context, fsys fs.FS, path string, targets []string) *ScanTask {
	resultsChan := make(chan Result)
	task := ScanTask{C: resultsChan}
	maxWorkers := min(8, runtime.GOMAXPROCS(0))
	searcher := NewSearcher(fsys, defaultSearchOpts)

	go func() {
		defer close(resultsChan)
		var wg sync.WaitGroup
		sem := make(semaphore, maxWorkers)

		// Launch StatDir goroutines directly from callback to avoid buffering overhead.
		// The semaphore provides backpressure: if StatDir operations are slow,
		// the filesystem search will naturally slow down, preventing memory pressure.
		task.err = searcher.Walk(path, targets, func(matchedPath string) error {
			if err := ctx.Err(); err != nil {
				return err
			}

			sem.Acquire()
			wg.Add(1)
			go func() {
				defer sem.Release()
				defer wg.Done()
				s, err := StatDir(ctx, fsys, matchedPath)
				if err != nil {
					// TODO: optional StatDir err logging goes here
					return
				}
				resultsChan <- Result{Path: matchedPath, Stats: s}
			}()
			return nil
		})

		wg.Wait()
	}()

	return &task
}

// A semaphore is a counting semaphore implemented via a buffered channel.
// To create a semaphore of capacity n, use make(semaphore, n).
type semaphore chan struct{}

func (s semaphore) Acquire() { s <- struct{}{} }
func (s semaphore) Release() { <-s }
