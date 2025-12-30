package import_server

import (
	"slices"
	"sort"
	"strings"
	"testing"
)

func Test_globToRsyncFilterPaths(t *testing.T) {
	type args struct {
		glob             string
		finalSegmentOnly bool
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			name: "empty glob",
			args: args{glob: "", finalSegmentOnly: false},
			want: nil,
		},
		{
			name: "only slashes",
			args: args{glob: "///", finalSegmentOnly: false},
			want: nil,
		},
		{
			name: "simple glob (all paths)",
			args: args{glob: "a/b/c", finalSegmentOnly: false},
			want: []string{"a/", "a/b/", "a/b/c"},
		},
		{
			name: "simple glob (only final)",
			args: args{glob: "a/b/c", finalSegmentOnly: true},
			want: []string{"a/b/c"},
		},

		{
			name: "leading slash preserved (all paths)",
			args: args{glob: "/a/b/c", finalSegmentOnly: false},
			want: []string{"/a/", "/a/b/", "/a/b/c"},
		},
		{
			name: "leading slash preserved (only final)",
			args: args{glob: "/a/b/c", finalSegmentOnly: true},
			want: []string{"/a/b/c"},
		},
		{
			name: "leading slash with ** (all paths)",
			args: args{glob: "/src/**/file", finalSegmentOnly: false},
			want: []string{"/src/", "/src/**/", "/src/**/file", "/src/file"},
		},

		{
			name: "single segment (all paths)",
			args: args{glob: "file.txt", finalSegmentOnly: false},
			want: []string{"file.txt"},
		},
		{
			name: "single segment (only final)",
			args: args{glob: "file.txt", finalSegmentOnly: true},
			want: []string{"file.txt"},
		},
		{
			name: "single segment with leading slash",
			args: args{glob: "/file.txt", finalSegmentOnly: false},
			want: []string{"/file.txt"},
		},

		{
			name: "** recursion (all paths)",
			args: args{glob: "a/b/**/c", finalSegmentOnly: false},
			want: []string{"a/", "a/b/", "a/b/**/", "a/b/c", "a/b/**/c"},
		},
		{
			name: "** recursion (only final)",
			args: args{glob: "a/b/**/c", finalSegmentOnly: true},
			want: []string{"a/b/c", "a/b/**/c"},
		},
		{
			name: "** at the end (all paths)",
			args: args{glob: "a/b/**", finalSegmentOnly: false},
			want: []string{"a/", "a/b/"},
		},
		{
			name: "** at the end (only final)",
			args: args{glob: "a/b/**", finalSegmentOnly: true},
			want: nil,
		},
		{
			name: "** at the beginning (all paths)",
			args: args{glob: "**/a/b", finalSegmentOnly: false},
			want: []string{"**/", "**/a/", "**/a/b"},
		},
		{
			name: "** at the beginning (only final)",
			args: args{glob: "**/a/b", finalSegmentOnly: true},
			want: []string{"**/a/b"},
		},
		{
			name: "multiple ** (all paths)",
			args: args{glob: "a/**/b/**/c", finalSegmentOnly: false},
			want: []string{
				"a/", "a/**/", "a/b/", "a/**/b/",
				"a/b/**/", "a/**/b/**/",
				"a/b/c", "a/**/b/c", "a/b/**/c", "a/**/b/**/c",
			},
		},
		{
			name: "only ** (all paths)",
			args: args{glob: "**", finalSegmentOnly: false},
			want: nil,
		},
		{
			name: "only ** (only final)",
			args: args{glob: "**", finalSegmentOnly: true},
			want: nil,
		},

		{
			name: "mixed c* (all paths)",
			args: args{glob: "a/b/**/c*", finalSegmentOnly: false},
			want: []string{"a/", "a/b/", "a/b/**/", "a/b/c*", "a/b/**/c*"},
		},
		{
			name: "mixed c* (only final)",
			args: args{glob: "a/b/**/c*", finalSegmentOnly: true},
			want: []string{"a/b/c*", "a/b/**/c*"},
		},
		{
			name: "mixed c? (all paths)",
			args: args{glob: "a/b/**/c?", finalSegmentOnly: false},
			want: []string{"a/", "a/b/", "a/b/**/", "a/b/c?", "a/b/**/c?"},
		},
		{
			name: "mixed c? (only final)",
			args: args{glob: "a/b/**/c?", finalSegmentOnly: true},
			want: []string{"a/b/c?", "a/b/**/c?"},
		},
		{
			name: "char class (all paths)",
			args: args{glob: "a/b/**/c[abc]", finalSegmentOnly: false},
			want: []string{"a/", "a/b/", "a/b/**/", "a/b/c[abc]", "a/b/**/c[abc]"},
		},
		{
			name: "char class (only final)",
			args: args{glob: "a/b/**/c[abc]", finalSegmentOnly: true},
			want: []string{"a/b/c[abc]", "a/b/**/c[abc]"},
		},
		{
			name: "asterisk in middle segment (all paths)",
			args: args{glob: "a/b*/c", finalSegmentOnly: false},
			want: []string{"a/", "a/b*/", "a/b*/c"},
		},

		{
			name: "trailing slash is trimmed",
			args: args{glob: "a/b/c/", finalSegmentOnly: false},
			want: []string{"a/", "a/b/", "a/b/c"},
		},
		{
			name: "both leading and trailing slashes",
			args: args{glob: "/a/b/c/", finalSegmentOnly: false},
			want: []string{"/a/", "/a/b/", "/a/b/c"},
		},

		{
			name: "typical app pattern (all paths)",
			args: args{glob: "/app/src/**/*.go", finalSegmentOnly: false},
			want: []string{"/app/", "/app/src/", "/app/src/**/", "/app/src/*.go", "/app/src/**/*.go"},
		},
		{
			name: "typical app pattern (only final)",
			args: args{glob: "/app/src/**/*.go", finalSegmentOnly: true},
			want: []string{"/app/src/*.go", "/app/src/**/*.go"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := globToRsyncFilterPaths(tt.args.glob, tt.args.finalSegmentOnly)
			sort.Strings(got)
			sort.Strings(tt.want)
			if !slices.Equal(got, tt.want) {
				t.Errorf("globDescentPath(%q, %v) = %v, want %v",
					tt.args.glob, tt.args.finalSegmentOnly, got, tt.want)
			}
		})
	}
}

func Test_prepareRsyncDirsOnlyFilters(t *testing.T) {
	tests := []struct {
		name         string
		add          string
		includePaths []string
		excludePaths []string
		wantContains []string
		wantEndsWith string
	}{
		{
			name:         "with include paths",
			add:          "/src",
			includePaths: []string{"dir1", "dir2/subdir"},
			excludePaths: nil,
			wantContains: []string{
				"--filter='+/ /src/dir1/'",
				"--filter='+/ /src/dir2/'",
				"--filter='+/ /src/dir2/subdir/'",
			},
			wantEndsWith: "--filter='-/ **'",
		},
		{
			name:         "with exclude paths only",
			add:          "/src",
			includePaths: nil,
			excludePaths: []string{"excluded"},
			wantContains: []string{
				"--filter='-/ /src/excluded'",
				"--filter='+/ **/'",
			},
			wantEndsWith: "--filter='-/ **'",
		},
		{
			name:         "no filters",
			add:          "/src",
			includePaths: nil,
			excludePaths: nil,
			wantContains: []string{
				"--filter='+/ **/'",
			},
			wantEndsWith: "--filter='-/ **'",
		},
		{
			name:         "include with ** glob",
			add:          "/app",
			includePaths: []string{"src/**"},
			excludePaths: nil,
			wantContains: []string{
				"--filter='+/ /app/src/'",
				"--filter='+/ /app/src/**/'",
			},
			wantEndsWith: "--filter='-/ **'",
		},
		{
			name:         "multiple include and exclude paths",
			add:          "/data",
			includePaths: []string{"cache", "logs/app"},
			excludePaths: []string{"tmp", "cache/old"},
			wantContains: []string{
				"--filter='-/ /data/tmp'",
				"--filter='-/ /data/cache/old'",
				"--filter='+/ /data/cache/'",
				"--filter='+/ /data/logs/'",
				"--filter='+/ /data/logs/app/'",
			},
			wantEndsWith: "--filter='-/ **'",
		},
		{
			name:         "empty add path",
			add:          "",
			includePaths: []string{"dir"},
			excludePaths: nil,
			wantContains: []string{
				"--filter='+/ dir/'",
				"--filter='+/ dir/**/'",
			},
			wantEndsWith: "--filter='-/ **'",
		},
		{
			name:         "add path without leading slash",
			add:          "src",
			includePaths: []string{"app"},
			excludePaths: nil,
			wantContains: []string{
				"--filter='+/ src/app/'",
				"--filter='+/ src/app/**/'",
			},
			wantEndsWith: "--filter='-/ **'",
		},
		{
			name:         "deeply nested include paths",
			add:          "/project",
			includePaths: []string{"a/b/c/d/e"},
			excludePaths: nil,
			wantContains: []string{
				"--filter='+/ /project/a/'",
				"--filter='+/ /project/a/b/'",
				"--filter='+/ /project/a/b/c/'",
				"--filter='+/ /project/a/b/c/d/'",
				"--filter='+/ /project/a/b/c/d/e/'",
				"--filter='+/ /project/a/b/c/d/e/**/'",
			},
			wantEndsWith: "--filter='-/ **'",
		},
		{
			name:         "add path with trailing slash",
			add:          "/src/",
			includePaths: nil,
			excludePaths: nil,
			wantContains: []string{
				"--filter='+/ **/'",
			},
			wantEndsWith: "--filter='-/ **'",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := PrepareRsyncDirsOnlyFilters(tt.add, tt.includePaths, tt.excludePaths)
			for _, want := range tt.wantContains {
				if !strings.Contains(got, want) {
					t.Errorf("PrepareRsyncDirsOnlyFilters() = %v, want to contain %v", got, want)
				}
			}
			if !strings.HasSuffix(got, tt.wantEndsWith) {
				t.Errorf("PrepareRsyncDirsOnlyFilters() = %v, want to end with %v", got, tt.wantEndsWith)
			}
		})
	}
}

func Test_prepareRsyncFilters(t *testing.T) {
	tests := []struct {
		name         string
		add          string
		includePaths []string
		excludePaths []string
		wantContains []string
		wantEmpty    bool
	}{
		{
			name:         "no filters",
			add:          "/src",
			includePaths: nil,
			excludePaths: nil,
			wantEmpty:    true,
		},
		{
			name:         "only include paths",
			add:          "/src",
			includePaths: []string{"dir1", "dir2"},
			excludePaths: nil,
			wantContains: []string{
				"--filter='+/ /src/dir1'",
				"--filter='+/ /src/dir2'",
				"--filter='-/ /src/**'",
			},
		},
		{
			name:         "only exclude paths",
			add:          "/src",
			includePaths: nil,
			excludePaths: []string{"excluded"},
			wantContains: []string{
				"--filter='-/ /src/excluded'",
			},
		},
		{
			name:         "both include and exclude paths",
			add:          "/app",
			includePaths: []string{"keep"},
			excludePaths: []string{"skip"},
			wantContains: []string{
				"--filter='-/ /app/skip'",
				"--filter='+/ /app/keep'",
				"--filter='-/ /app/**'",
			},
		},
		{
			name:         "include with ** glob",
			add:          "/data",
			includePaths: []string{"logs/**"},
			excludePaths: nil,
			wantContains: []string{
				"--filter='+/ /data/logs/'",
				"--filter='+/ /data/logs/**'",
				"--filter='-/ /data/**'",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := PrepareRsyncFilters(tt.add, tt.includePaths, tt.excludePaths)
			if tt.wantEmpty {
				if got != "" {
					t.Errorf("PrepareRsyncFilters() = %v, want empty string", got)
				}
				return
			}
			for _, want := range tt.wantContains {
				if !strings.Contains(got, want) {
					t.Errorf("PrepareRsyncFilters() = %v, want to contain %v", got, want)
				}
			}
		})
	}
}

func Test_prepareRsyncIncludeFiltersForGlobs(t *testing.T) {
	tests := []struct {
		name         string
		add          string
		includeGlobs []string
		wantContains []string
		wantEndsWith string
	}{
		{
			name:         "empty globs",
			add:          "/src",
			includeGlobs: nil,
			wantContains: nil,
			wantEndsWith: "",
		},
		{
			name:         "simple directory",
			add:          "/src",
			includeGlobs: []string{"mydir"},
			wantContains: []string{
				"--filter='+/ /src/mydir'",
				"--filter='+/ /src/mydir/**'",
			},
			wantEndsWith: "--filter='-/ /src/**'",
		},
		{
			name:         "nested directory",
			add:          "/app",
			includeGlobs: []string{"a/b/c"},
			wantContains: []string{
				"--filter='+/ /app/a/'",
				"--filter='+/ /app/a/b/'",
				"--filter='+/ /app/a/b/c'",
				"--filter='+/ /app/a/b/c/**'",
			},
			wantEndsWith: "--filter='-/ /app/**'",
		},
		{
			name:         "glob with ** pattern",
			add:          "/data",
			includeGlobs: []string{"logs/**/error.log"},
			wantContains: []string{
				"--filter='+/ /data/logs/'",
				"--filter='+/ /data/logs/**/'",
				"--filter='+/ /data/logs/error.log'",
				"--filter='+/ /data/logs/**/error.log'",
			},
			wantEndsWith: "--filter='-/ /data/**'",
		},
		{
			name:         "multiple globs",
			add:          "/project",
			includeGlobs: []string{"src", "lib"},
			wantContains: []string{
				"--filter='+/ /project/src'",
				"--filter='+/ /project/src/**'",
				"--filter='+/ /project/lib'",
				"--filter='+/ /project/lib/**'",
			},
			wantEndsWith: "--filter='-/ /project/**'",
		},
		{
			name:         "glob with wildcard",
			add:          "/src",
			includeGlobs: []string{"*.go"},
			wantContains: []string{
				"--filter='+/ /src/*.go'",
				"--filter='+/ /src/*.go/**'",
			},
			wantEndsWith: "--filter='-/ /src/**'",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := PrepareRsyncIncludeFiltersForGlobs(tt.add, tt.includeGlobs)
			if len(tt.includeGlobs) == 0 {
				if got != "" {
					t.Errorf("PrepareRsyncIncludeFiltersForGlobs() = %v, want empty string for empty globs", got)
				}
				return
			}
			for _, want := range tt.wantContains {
				if !strings.Contains(got, want) {
					t.Errorf("PrepareRsyncIncludeFiltersForGlobs() = %v, want to contain %v", got, want)
				}
			}
			if tt.wantEndsWith != "" && !strings.HasSuffix(got, tt.wantEndsWith) {
				t.Errorf("PrepareRsyncIncludeFiltersForGlobs() = %v, want to end with %v", got, tt.wantEndsWith)
			}
		})
	}
}

func Test_prepareRsyncExcludeFiltersForGlobs(t *testing.T) {
	tests := []struct {
		name         string
		add          string
		excludeGlobs []string
		wantContains []string
	}{
		{
			name:         "empty globs",
			add:          "/src",
			excludeGlobs: nil,
			wantContains: nil,
		},
		{
			name:         "simple exclude",
			add:          "/src",
			excludeGlobs: []string{"tmp"},
			wantContains: []string{
				"--filter='-/ /src/tmp'",
			},
		},
		{
			name:         "nested exclude",
			add:          "/app",
			excludeGlobs: []string{"cache/old"},
			wantContains: []string{
				"--filter='-/ /app/cache/old'",
			},
		},
		{
			name:         "exclude with ** pattern returns empty",
			add:          "/data",
			excludeGlobs: []string{"logs/**"},
			wantContains: nil,
		},
		{
			name:         "multiple excludes",
			add:          "/project",
			excludeGlobs: []string{"tmp", "cache", "*.log"},
			wantContains: []string{
				"--filter='-/ /project/tmp'",
				"--filter='-/ /project/cache'",
				"--filter='-/ /project/*.log'",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := PrepareRsyncExcludeFiltersForGlobs(tt.add, tt.excludeGlobs)
			if len(tt.wantContains) == 0 {
				if got != "" {
					t.Errorf("PrepareRsyncExcludeFiltersForGlobs() = %v, want empty string", got)
				}
			}
			for _, want := range tt.wantContains {
				if !strings.Contains(got, want) {
					t.Errorf("PrepareRsyncExcludeFiltersForGlobs() = %v, want to contain %v", got, want)
				}
			}
		})
	}
}

func Test_isDirectoryGlob(t *testing.T) {
	tests := []struct {
		glob string
		want bool
	}{
		// Directory globs (should return true)
		{"cache", true},
		{"cache/", true},
		{"app/data", true},
		{"app/data/", true},
		{"logs/**", true},
		{"app/logs/**", true},
		{"**", true},
		{"", true},
		{".hidden", true},
		{"Makefile", true},

		// File globs (should return false)
		{"*.txt", false},
		{"**/*.txt", false},
		{"**/*.go", false},
		{"file?.log", false},
		{"data[0-9].json", false},
		{"app/*.txt", false},
		{"logs/**/*.log", false},
		{"*", false},
		{"app/**/x.txt", false},
		{"x.txt", false},
		{"config.json", false},
		{"Makefile.bak", false},
		{"src/main.go", false},
		{"app/a/b/c.txt", false},
	}
	for _, tt := range tests {
		t.Run(tt.glob, func(t *testing.T) {
			got := isDirectoryGlob(tt.glob)
			if got != tt.want {
				t.Errorf("isDirectoryGlob(%q) = %v, want %v", tt.glob, got, tt.want)
			}
		})
	}
}

func Test_hasDirectoryGlobs(t *testing.T) {
	tests := []struct {
		name  string
		globs []string
		want  bool
	}{
		{
			name:  "empty list",
			globs: nil,
			want:  false,
		},
		{
			name:  "only file globs",
			globs: []string{"*.txt", "**/*.go"},
			want:  false,
		},
		{
			name:  "only directory globs",
			globs: []string{"cache", "logs/**"},
			want:  true,
		},
		{
			name:  "mixed globs - has directory",
			globs: []string{"*.txt", "cache", "**/*.go"},
			want:  true,
		},
		{
			name:  "single directory glob",
			globs: []string{"app/data"},
			want:  true,
		},
		{
			name:  "single file glob",
			globs: []string{"**/*.txt"},
			want:  false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := hasDirectoryGlobs(tt.globs)
			if got != tt.want {
				t.Errorf("hasDirectoryGlobs(%v) = %v, want %v", tt.globs, got, tt.want)
			}
		})
	}
}
