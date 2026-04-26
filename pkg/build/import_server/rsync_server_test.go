package import_server

import (
	"reflect"
	"sort"
	"strings"
	"testing"
)

func TestAI_PrepareRsyncIncludeFiltersForGlobs_NoAbsolutePathModifier(t *testing.T) {
	tests := []struct {
		name         string
		add          string
		includeGlobs []string
		wantContains []string
		wantAbsent   []string
	}{
		{
			name:         "add=/ includePaths=[usr]",
			add:          "/",
			includeGlobs: []string{"usr"},
			wantContains: []string{"--filter='+ /usr'", "--filter='+ /usr/**'", "--filter='- /**'"},
			wantAbsent:   []string{"+/", "-/"},
		},
		{
			name:         "add=/ includePaths=[usr/lib]",
			add:          "/",
			includeGlobs: []string{"usr/lib"},
			wantContains: []string{"--filter='+ /usr/lib'", "--filter='+ /usr/lib/**'", "--filter='+ usr/'", "--filter='- /**'"},
			wantAbsent:   []string{"+/", "-/"},
		},
		{
			name:         "empty includeGlobs returns empty",
			add:          "/",
			includeGlobs: nil,
			wantContains: nil,
			wantAbsent:   []string{"+/", "-/"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := PrepareRsyncIncludeFiltersForGlobs(tt.add, tt.includeGlobs)
			for _, want := range tt.wantContains {
				if !strings.Contains(got, want) {
					t.Errorf("PrepareRsyncIncludeFiltersForGlobs(%q, %v) = %q, want to contain %q", tt.add, tt.includeGlobs, got, want)
				}
			}
			for _, absent := range tt.wantAbsent {
				if strings.Contains(got, absent) {
					t.Errorf("PrepareRsyncIncludeFiltersForGlobs(%q, %v) = %q, must NOT contain %q (absolute path modifier)", tt.add, tt.includeGlobs, got, absent)
				}
			}
		})
	}
}

func TestAI_PrepareRsyncExcludeFiltersForGlobs_NoAbsolutePathModifier(t *testing.T) {
	tests := []struct {
		name         string
		add          string
		excludeGlobs []string
		wantContains []string
		wantAbsent   []string
	}{
		{
			name:         "add=/ excludePaths=[tmp]",
			add:          "/",
			excludeGlobs: []string{"tmp"},
			wantContains: []string{"--filter='- tmp'"},
			wantAbsent:   []string{"-/"},
		},
		{
			name:         "empty excludeGlobs returns empty",
			add:          "/",
			excludeGlobs: nil,
			wantContains: nil,
			wantAbsent:   []string{"-/"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := PrepareRsyncExcludeFiltersForGlobs(tt.add, tt.excludeGlobs)
			for _, want := range tt.wantContains {
				if !strings.Contains(got, want) {
					t.Errorf("PrepareRsyncExcludeFiltersForGlobs(%q, %v) = %q, want to contain %q", tt.add, tt.excludeGlobs, got, want)
				}
			}
			for _, absent := range tt.wantAbsent {
				if strings.Contains(got, absent) {
					t.Errorf("PrepareRsyncExcludeFiltersForGlobs(%q, %v) = %q, must NOT contain %q (absolute path modifier)", tt.add, tt.excludeGlobs, got, absent)
				}
			}
		})
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
