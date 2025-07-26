package deepclean

import (
	"reflect"
	"testing"
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
