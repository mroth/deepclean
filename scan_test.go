package deepclean

import (
	"reflect"
	"testing"
)

// godirwalk doesnt expose a fs interface, so we need to use the real file
// system to test Scans, making them effectively integration tests.

func TestScan(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode.")
	}

	tests := []struct {
		name    string
		path    string
		targets []string
		want    []Result // unordered!
	}{
		{
			name:    "case01 - no matched targets",
			path:    "testdata/case01",
			targets: []string{"fizzbuzz"},
			want:    []Result{},
		},
		{
			name:    "case01 - sample",
			path:    "testdata/case01",
			targets: []string{"node_modules", "vendor", "target"},
			want: []Result{
				{Path: "testdata/case01/build/vendor", Stats: DirStats{Files: 2, Bytes: 98}},
				{Path: "testdata/case01/node_modules", Stats: DirStats{Files: 8, Bytes: 509}},
			},
		},
		{
			name:    "case02 - empty",
			path:    "testdata/case02",
			targets: []string{"node_modules", "vendor", "target"},
			want:    []Result{},
		},
		// TODO error case, unknown path (currently exits)
	}
	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			var rs []Result
			for r := range Scan(tt.path, tt.targets) {
				rs = append(rs, r)
			}

			var (
				want = newResultSet(t, tt.want)
				got  = newResultSet(t, rs)
			)
			if !reflect.DeepEqual(want, got) {
				t.Errorf("want %v, got %v", want, got)
			}
		})
	}
}

// convert unordered result slice into a set-like data structure for comparison
func newResultSet(t *testing.T, rs []Result) map[Result]bool {
	t.Helper()
	set := make(map[Result]bool)
	for _, r := range rs {
		_, found := set[r]
		if found {
			t.Fatalf("duplicate result for path: %v", r)
		}
		set[r] = true
	}
	return set
}
