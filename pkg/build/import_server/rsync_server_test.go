package import_server

import (
	"reflect"
	"sort"
	"strings"
	"testing"
)

func TestPrepareRsyncFilters_DefaultSystemExcludes(t *testing.T) {
	got := PrepareRsyncFilters("", nil, nil)

	want := " --filter='-/ dev' --filter='-/ proc' --filter='-/ run' --filter='-/ sys'"
	if got != want {
		t.Fatalf("unexpected filters: got %q, want %q", got, want)
	}
}

func TestPrepareRsyncFilters_IncludePathsKeepsSystemExcludes(t *testing.T) {
	got := PrepareRsyncFilters("", []string{"app/**"}, nil)

	for _, expected := range []string{
		"--filter='-/ dev'",
		"--filter='-/ proc'",
		"--filter='-/ run'",
		"--filter='-/ sys'",
		"--filter='+/ app/'",
		"--filter='+/ app/**'",
		"--filter='+/ app/**/**'",
		"--filter='-/ **'",
	} {
		if !strings.Contains(got, expected) {
			t.Fatalf("expected %q in %q", expected, got)
		}
	}
}

func TestPrepareRsyncFilters_RootIncludeAllKeepsSystemExcludesFirst(t *testing.T) {
	got := PrepareRsyncFilters("/", []string{"**/*"}, nil)

	for _, expected := range []string{
		"--filter='-/ dev'",
		"--filter='-/ proc'",
		"--filter='-/ run'",
		"--filter='-/ sys'",
		"--filter='+/ **/'",
		"--filter='+/ **/*'",
	} {
		if !strings.Contains(got, expected) {
			t.Fatalf("expected %q in %q", expected, got)
		}
	}

	if !strings.Contains(got, "--filter='-/ **'") && !strings.Contains(got, "--filter='-/ /**'") {
		t.Fatalf("expected root catch-all exclude in %q", got)
	}

	if strings.Index(got, "--filter='-/ proc'") > strings.Index(got, "--filter='+/ **/*'") {
		t.Fatalf("exclude rules must be placed before include rules: %q", got)
	}
}

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
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := globToRsyncFilterPaths(tt.args.glob, tt.args.finalSegmentOnly)
			sort.Strings(got)
			sort.Strings(tt.want)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("globDescentPath(%q, %v) = %v, want %v",
					tt.args.glob, tt.args.finalSegmentOnly, got, tt.want)
			}
		})
	}
}
