package deepclean

type DirStats struct {
	Files uint64
	Bytes uint64
}

// Add two DirStats together to aggregate the results.
func (a DirStats) Add(b DirStats) DirStats {
	return DirStats{
		Files: a.Files + b.Files,
		Bytes: a.Bytes + b.Bytes,
	}
}

type Result struct {
	Path  string
	Stats DirStats
}

func Aggregate(rs ...Result) DirStats {
	var t DirStats
	for _, r := range rs {
		t = t.Add(r.Stats)
	}
	return t
}
