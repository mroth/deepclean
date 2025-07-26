package deepclean

import (
	"os"
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
		wantErr bool
	}{
		{
			name:    "case01 - no matched targets",
			path:    "case01",
			targets: []string{"fizzbuzz"},
			want:    []Result{},
		},
		{
			name:    "case01 - sample",
			path:    "case01",
			targets: []string{"node_modules", "vendor", "target"},
			want: []Result{
				{Path: "case01/build/vendor", Stats: DirStats{Files: 2, Bytes: 2}},
				{Path: "case01/node_modules", Stats: DirStats{Files: 8, Bytes: 29}},
			},
		},
		{
			name:    "case02 - empty",
			path:    "case02",
			targets: []string{"node_modules", "vendor", "target"},
			want:    []Result{},
		},
		{
			name:    "invalid path",
			path:    "XXXXXX",
			want:    []Result{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testdataFS := os.DirFS("testdata")
			var rs []Result
			scanner := Scan(testdataFS, tt.path, tt.targets)
			for r := range scanner.C {
				rs = append(rs, r)
			}

			var (
				want = newResultSet(t, tt.want)
				got  = newResultSet(t, rs)
			)
			if !reflect.DeepEqual(want, got) {
				t.Errorf("want %v, got %v", want, got)
			}
			if (scanner.Err() != nil) != tt.wantErr {
				t.Errorf("wantErr: %v, got: %v", tt.wantErr, scanner.Err())
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
