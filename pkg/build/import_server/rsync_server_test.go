package import_server

import (
	"reflect"
	"sort"
	"strings"
	"testing"
)

func TestPrepareRsyncFilters_NoSystemExcludesOnClient(t *testing.T) {
	got := PrepareRsyncFilters("", nil, nil)
	if got != "" {
		t.Fatalf("expected empty filters when no include/exclude paths, got %q", got)
	}
}

func TestPrepareRsyncFilters_IncludePathsNoSystemExcludes(t *testing.T) {
	got := PrepareRsyncFilters("", []string{"app/**"}, nil)

	for _, notExpected := range []string{
		"--filter='-/ dev'",
		"--filter='-/ proc'",
		"--filter='-/ run'",
		"--filter='-/ sys'",
	} {
		if strings.Contains(got, notExpected) {
			t.Fatalf("system exclude %q must not be present in client filters %q (handled server-side)", notExpected, got)
		}
	}

	for _, expected := range []string{
		"--filter='+/ app/'",
		"--filter='+/ app/**'",
		"--filter='-/ **'",
	} {
		if !strings.Contains(got, expected) {
			t.Fatalf("expected %q in %q", expected, got)
		}
	}
}

func TestRsyncdConfContainsSystemExcludes(t *testing.T) {
	conf := buildRsyncdConf("873", "testuser")

	for _, dir := range systemExcludeDirs {
		if !strings.Contains(conf, dir) {
			t.Errorf("rsyncd.conf must exclude system dir %q, got:\n%s", dir, conf)
		}
	}

	if !strings.Contains(conf, "exclude =") {
		t.Errorf("rsyncd.conf must contain exclude directive, got:\n%s", conf)
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
