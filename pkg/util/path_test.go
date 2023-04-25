package util

import (
	"testing"
)

func TestSafeTrimGlobsAndSlashes(t *testing.T) {
	tests := []struct {
		name string
		path string
		want string
	}{
		{
			name: "globs",
			path: "**/*",
			want: "",
		},
		{
			name: "empty path",
			path: "",
			want: "",
		},
		{
			name: "path",
			path: "path",
			want: "path",
		},
		{
			name: "path with trailing slash",
			path: "path/",
			want: "path",
		},
		{
			name: "path with globs and slashes",
			path: "path/**/*",
			want: "path",
		},
		{
			name: "path with glob as part of a directory or file name 1",
			path: "path/name-*",
			want: "path/name-*",
		},
		{
			name: "path with glob as part of a directory or file name 2",
			path: "path/*.tmp",
			want: "path/*.tmp",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := SafeTrimGlobsAndSlashesFromPath(tt.path); got != tt.want {
				t.Errorf("SafeTrimGlobsAndSlashesFromPath() = %v, want %v", got, tt.want)
			}
		})
	}
}
