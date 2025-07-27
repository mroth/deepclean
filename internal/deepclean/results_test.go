package deepclean

import (
	"reflect"
	"testing"
)

func TestAggregate(t *testing.T) {
	tests := []struct {
		name string
		rs   []Result
		want DirStats
	}{
		{
			name: "empty",
			rs:   []Result{},
			want: DirStats{Files: 0, Bytes: 0},
		},
		{
			name: "two results",
			rs: []Result{
				{Path: "a", Stats: DirStats{Files: 111, Bytes: 111}},
				{Path: "b", Stats: DirStats{Files: 222, Bytes: 222}},
			},
			want: DirStats{Files: 333, Bytes: 333},
		},
		{
			name: "three results",
			rs: []Result{
				{Path: "a", Stats: DirStats{Files: 111, Bytes: 111}},
				{Path: "b", Stats: DirStats{Files: 222, Bytes: 222}},
				{Path: "c", Stats: DirStats{Files: 333, Bytes: 333}},
			},
			want: DirStats{Files: 666, Bytes: 666},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Aggregate(tt.rs...); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}
