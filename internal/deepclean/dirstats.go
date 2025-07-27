package deepclean

import (
	"context"
	"io/fs"
)

// DirStats contains metadata for disk usage of a parent directory.
type DirStats struct {
	Files uint64
	Bytes uint64
}

// Add combines two DirStats together to total their results.
func (a DirStats) Add(b DirStats) DirStats {
	return DirStats{
		Files: a.Files + b.Files,
		Bytes: a.Bytes + b.Bytes,
	}
}

// StatDir walks the directory at path collecting the aggregate DirStats.
// The operation respects context cancellation and will abort early if the context is cancelled.
func StatDir(ctx context.Context, fsys fs.FS, path string) (DirStats, error) {
	var t DirStats
	err := fs.WalkDir(fsys, path, func(fpath string, d fs.DirEntry, err error) error {
		// failed to read directory entry, abort
		if err != nil {
			return err
		}

		// if context is done, abort early
		if ctx.Err() != nil {
			return ctx.Err()
		}

		// get file info to read size
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
