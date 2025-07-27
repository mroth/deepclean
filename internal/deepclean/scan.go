package deepclean

import (
	"io/fs"
	"path"
	"runtime"
	"slices"
	"sync"
)

// ScanTask contains fields to access the results of an ongoing Scan.
//
// If the underlying filepath Walk encounters a fatal error, the results
// channel will be closed and Err() will return a non-nil value. Always
// drain (*ScanTask).C prior to checking Err().
//
// If the underlying filepath Walk encounters a non-fatal error, the
// walked directory will be silently skipped (for now).
type ScanTask struct {
	C   <-chan Result
	err error
}

// Err returns the error status of the underlying filepath Walk performed by the
// scanner. This will only be set once the Walk has exited, indicated by the
// Results channel being closed.
func (s ScanTask) Err() error {
	return s.err
}

// Scan walks the filesystem searching for directories matching the targets strings,
// and then initiates a DirStats on the directory, returning the Results as they
// occur on ScanTask.C.
func Scan(fsys fs.FS, path string, targets []string) *ScanTask {
	resultsChan := make(chan Result)
	scanner := ScanTask{C: resultsChan}

	go func() {
		defer close(resultsChan)

		// spawn worker pool to perform stating of matched target directories
		matchedDirs := make(chan string)
		var wg sync.WaitGroup
		for range runtime.GOMAXPROCS(0) {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for fpath := range matchedDirs {
					s, err := Stat(fsys, fpath)
					if err != nil {
						continue
					}
					resultsChan <- Result{Path: fpath, Stats: s}
				}
			}()
		}

		// primary file system walk looking for target directories
		scanner.err = fs.WalkDir(fsys, path, func(fpath string, d fs.DirEntry, err error) error {
			if err != nil {
				// error on initial directory should be fatal
				if path == fpath {
					return err
				}
				// error anywhere else should skip the file, not abort the scan
				// TODO: log only if verbose
				// fmt.Fprintf(os.Stderr, "ERROR: %s\n", err)
				return nil
			}

			// Positives are directories matching any target string. For each
			// match, send to the worker pool to gather stats, but skip further
			// walking of the subdir in this primary scan thread.
			if d.IsDir() && inTargets(targets, fpath) {
				matchedDirs <- fpath
				return fs.SkipDir
			}
			return nil
		})
		close(matchedDirs)
		wg.Wait()
	}()

	return &scanner
}

func inTargets(targets []string, fpath string) bool {
	return slices.Contains(targets, path.Base(fpath))
}

// Stat walks the directory at path collecting the aggregate DirStats.
func Stat(fsys fs.FS, path string) (DirStats, error) {
	var t DirStats
	err := fs.WalkDir(fsys, path, func(fpath string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		fi, err := d.Info()
		if err != nil {
			return err
		}

		t.Files++
		if !d.IsDir() {
			t.Bytes += uint64(fi.Size())
		}
		return nil
	})
	return t, err
}
