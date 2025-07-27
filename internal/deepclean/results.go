package deepclean

// Result wraps DirStats with the Path that was used to stat the data.
//
// It is used whenever a task may stat multiple directories and return the
// results out-of-order and/or asynchronously.
type Result struct {
	Path  string
	Stats DirStats
}

// Aggregate totals the DirStats from multiple Results.
func Aggregate(rs ...Result) DirStats {
	var t DirStats
	for _, r := range rs {
		t = t.Add(r.Stats)
	}
	return t
}
