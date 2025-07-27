package deepclean

import (
	"io/fs"
	"slices"
)

// A Searcher performs walks of the filesystem looking for candidate directories.
type Searcher struct {
	fsys fs.FS
	opts SearchOpts
}

// SearchOpts configures the behavior of a Searcher.
type SearchOpts struct {
	ExcludeDirs []string // Paths to always ignore (exact name match)
}

var defaultSearchOpts = SearchOpts{
	ExcludeDirs: []string{".git", ".hg", ".svn", ".jj"},
}

// NewSearcher creates a new Searcher with the given filesystem and options.
func NewSearcher(fsys fs.FS, opts SearchOpts) *Searcher {
	return &Searcher{
		fsys: fsys,
		opts: opts,
	}
}

// Walk begins the Searcher walking the filesystem at the given root path,
// efficiently looking for directories matching any of the targets.
//
// It invokes the provided callback fn for each matched directory path, which
// may return an error to abort the walk early.
//
// If the root path cannot be read, this function will return that error. Other
// errors encountered while walking the filesystem will be ignored.
func (s *Searcher) Walk(root string, targets []string, fn SearchWalkFunc) error {
	walker := func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			// error on initial directory should be fatal
			if root == p {
				return err
			}
			// error anywhere else should skip the file, not abort the scan
			// TODO: log only if verbose
			// fmt.Fprintf(os.Stderr, "ERROR: %s\n", err)
			return nil
		}

		if d.IsDir() {
			dirname := d.Name()

			// skip unwanted directories, do not descend into them
			if slices.Contains(s.opts.ExcludeDirs, dirname) {
				return fs.SkipDir
			}

			// check if this directory matches any of the targets.
			// If so, apply the callback and skip descending
			// further into this directory.
			if slices.Contains(targets, dirname) {
				if err := fn(p); err != nil {
					return err // callback can abort the walk
				}
				return fs.SkipDir
			}
		}
		return nil
	}

	return fs.WalkDir(s.fsys, root, walker)
}

// SearchWalkFunc is the callback function type used by Searcher.Walk.
// It is called for each directory that matches the search targets.
// Returning an error will abort the filesystem walk.
type SearchWalkFunc func(matchedPath string) error
