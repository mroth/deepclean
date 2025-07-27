package deepclean

import (
	"errors"
	"slices"
	"testing"
	"testing/fstest"
)

func TestSearcher_Walk(t *testing.T) {
	testFS := fstest.MapFS{
		"file.txt":                   {Data: []byte("content")},
		"node_modules/package.json":  {Data: []byte("{}")},
		"node_modules/lib/index.js":  {Data: []byte("code")},
		"src/main.go":                {Data: []byte("package main")},
		"vendor/pkg/lib.go":          {Data: []byte("package pkg")},
		"target/classes/Main.class":  {Data: []byte("bytecode")},
		"nested/node_modules/dep.js": {Data: []byte("dependency")},
		"nested/src/code.go":         {Data: []byte("code")},
		".git/config":                {Data: []byte("git config")},
		"docs/readme.md":             {Data: []byte("documentation")},
	}

	tests := []struct {
		name    string
		root    string
		targets []string
		want    []string
		wantErr bool
	}{
		{
			name:    "find single target",
			root:    ".",
			targets: []string{"node_modules"},
			want:    []string{"node_modules", "nested/node_modules"},
		},
		{
			name:    "find multiple targets",
			root:    ".",
			targets: []string{"node_modules", "vendor", "target"},
			want:    []string{"node_modules", "vendor", "target", "nested/node_modules"},
		},
		{
			name:    "no matches",
			root:    ".",
			targets: []string{"nonexistent"},
			want:    []string{},
		},
		{
			name:    "search subdirectory",
			root:    "nested",
			targets: []string{"node_modules"},
			want:    []string{"nested/node_modules"},
		},
		{
			name:    "error reading root path",
			root:    "nonexistent",
			targets: []string{"node_modules"},
			want:    []string{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			searcher := NewSearcher(testFS, defaultSearchOpts)
			matches := make([]string, 0)
			err := searcher.Walk(tt.root, tt.targets, func(matchedPath string) error {
				matches = append(matches, matchedPath)
				return nil
			})

			if (err != nil) != tt.wantErr {
				t.Errorf("Searcher.Walk() error = %v, wantErr %v", err, tt.wantErr)
			}

			// Sort both slices for comparison since order may vary
			got := slices.Sorted(slices.Values(matches))
			want := slices.Sorted(slices.Values(tt.want))
			if !slices.Equal(got, want) {
				t.Errorf("Searcher.Walk() sorted matches = %v, want %v", got, want)
			}
		})
	}
}

// Verify that if the callback returns an error, the walk is aborted and the error is propagated.
func TestSearcher_Walk_CallbackError(t *testing.T) {
	testFS := fstest.MapFS{
		"node_modules/package.json": {Data: []byte("{}")}, // need at least one match to invoke the match callback
		"vendor/lib.go":             {Data: []byte("code")},
	}

	var specificErr = errors.New("my specific error")
	searcher := NewSearcher(testFS, defaultSearchOpts)
	fn := func(matchedPath string) error {
		return specificErr // Return error to abort walk
	}
	err := searcher.Walk(".", []string{"node_modules", "vendor"}, fn)

	if err != specificErr {
		t.Errorf("Searcher.Walk() error = %v, want %v", err, specificErr)
	}
}
