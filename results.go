package main

import (
	"fmt"

	humanize "github.com/dustin/go-humanize"
)

type totals struct {
	numFiles, bytes uint64
}

func (a totals) Add(b totals) totals {
	return totals{
		numFiles: a.numFiles + b.numFiles,
		bytes:    a.bytes + b.bytes,
	}
}

type result struct {
	path string
	totals
}

func (r result) String() string {
	return fmt.Sprintf(
		"%7d\t%7s\t%s", r.numFiles, humanize.Bytes(r.bytes), r.path)
}

type results []result

func (rs results) Sum() totals {
	var t totals
	for _, r := range rs {
		t = t.Add(r.totals)
	}
	return t
}
