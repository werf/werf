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
			glob:     "usr/bin/[",
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
