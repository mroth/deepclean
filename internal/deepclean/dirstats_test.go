package deepclean

import (
	"context"
	"reflect"
	"testing"
	"testing/fstest"
)

func TestDirStats_Add(t *testing.T) {
	tests := []struct {
		name string
		a, b DirStats
		want DirStats
	}{
		{
			a:    DirStats{Files: 0, Bytes: 0},
			b:    DirStats{Files: 0, Bytes: 0},
			want: DirStats{Files: 0, Bytes: 0},
		},
		{
			a:    DirStats{Files: 101, Bytes: 123456789},
			b:    DirStats{Files: 123, Bytes: 777777777},
			want: DirStats{Files: 224, Bytes: 901234566},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.a.Add(tt.b); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("got %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestStatDir(t *testing.T) {
	// Create a single test filesystem to use for all test cases
	testFS := fstest.MapFS{
		"empty":                   {},                            // empty directory
		"single/file.txt":         {Data: []byte("hello world")}, // 11 bytes
		"multi/file1.txt":         {Data: []byte("content1")},    // 8 bytes
		"multi/file2.txt":         {Data: []byte("content22")},   // 9 bytes
		"multi/subdir/nested.txt": {Data: []byte("nested")},      // 6 bytes
		"mixed/file.txt":          {Data: []byte("data")},        // 4 bytes
		"mixed/subdir1/deep1.txt": {Data: []byte("deep1")},       // 5 bytes
		"mixed/subdir2/deep2.txt": {Data: []byte("deep22")},      // 6 bytes
		"mixed/subdir2/deep3.txt": {Data: []byte("deep333")},     // 7 bytes
	}

	tests := []struct {
		name    string
		path    string
		want    DirStats
		wantErr bool
	}{
		{
			path: "empty",
			want: DirStats{Files: 1, Bytes: 0}, // directory itself counts as 1 file
		},
		{
			path: "single",
			want: DirStats{Files: 2, Bytes: 11}, // directory + file
		},
		{
			path: "multi",
			want: DirStats{Files: 5, Bytes: 23}, // dir + file1 + file2 + subdir + nested.txt = 5 files, 8+9+6 = 23 bytes
		},
		{
			path: "mixed",
			want: DirStats{Files: 7, Bytes: 22}, // dir + file + subdir1 + subdir2 + deep1 + deep2 + deep3 = 7 files, 4+5+6+7 = 22 bytes
		},
		{
			path:    "nonexistent",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			got, err := StatDir(t.Context(), testFS, tt.path)

			if (err != nil) != tt.wantErr {
				t.Errorf("StatDir() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("StatDir() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestStatDir_ContextCancellation(t *testing.T) {
	testFS := fstest.MapFS{
		"test/file1.txt": {Data: []byte("content1")},
		"test/file2.txt": {Data: []byte("content2")},
	}

	// Create a cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	_, err := StatDir(ctx, testFS, "test")
	if err == nil {
		t.Error("StatDir() should return error when context is cancelled")
	}

	if err != context.Canceled {
		t.Errorf("StatDir() error = %v, want %v", err, context.Canceled)
	}
}
