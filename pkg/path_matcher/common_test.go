package path_matcher

import (
	"testing"
)

func TestIsPathMatched(t *testing.T) {
	tests := []struct {
		name     string
		filePath string
		glob     string
		want     bool
	}{
		{
			name:     "exact match with special character",
			filePath: "usr/bin/[",
			glob:     "usr/bin/\\[",
			want:     true,
		},
		{
			name:     "file exists under glob with wildcard",
			filePath: "usr/bin/falco",
			glob:     "usr/bin/**",
			want:     true,
		},
		{
			name:     "file not matching unrelated glob",
			filePath: "etc/falco",
			glob:     "usr/bin/*",
			want:     false,
		},
		{
			name:     "deep path under glob",
			filePath: "usr/share/falco/plugins/libcontainer.so",
			glob:     "usr/share/**",
			want:     true,
		},
		{
			name:     "root lib64 path",
			filePath: "lib64",
			glob:     "lib64",
			want:     true,
		},
		{
			name:     "deep path with fallback form",
			filePath: "usr/bin/falco",
			glob:     "usr/bin",
			want:     true,
		},
		{
			name:     "glob with multiple patterns",
			filePath: "images/img/Dockerfile",
			glob:     "images/*/{Dockerfile,werf.inc.yaml}",
			want:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isPathMatched(tt.filePath, tt.glob)
			if got != tt.want {
				t.Errorf("isPathMatched(%q, %q) = %v; want %v", tt.filePath, tt.glob, got, tt.want)
			}
		})
	}
}

func TestFormatPath(t *testing.T) {
	tests := []struct {
		name string
		path string
		want string
	}{
		{
			name: "empty path",
			path: "",
			want: "",
		},
		{
			name: "dot path",
			path: ".",
			want: "",
		},
		{
			name: "root slash path",
			path: "/",
			want: "",
		},
		{
			name: "normal path",
			path: ".helm",
			want: ".helm",
		},
		{
			name: "path with subdirectory",
			path: ".helm/templates",
			want: ".helm/templates",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatPath(tt.path)
			if got != tt.want {
				t.Errorf("formatPath(%q) = %q; want %q", tt.path, got, tt.want)
			}
		})
	}
}
