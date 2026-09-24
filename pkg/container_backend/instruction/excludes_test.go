package instruction

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestContextRelativeExcludes(t *testing.T) {
	for _, tt := range []struct {
		name            string
		sourcePaths     []string
		excludePatterns []string
		expected        []string
	}{
		{
			name:        "no patterns",
			sourcePaths: []string{"./aaa/"},
		},
		{
			name:            "pattern per source path",
			sourcePaths:     []string{"./aaa/", "./bbb"},
			excludePatterns: []string{"*.md", "node_modules"},
			expected:        []string{"aaa/*.md", "aaa/node_modules", "bbb/*.md", "bbb/node_modules"},
		},
		{
			name:            "context root source path",
			sourcePaths:     []string{"."},
			excludePatterns: []string{"*.md"},
			expected:        []string{"*.md"},
		},
		{
			name:            "absolute source path from another stage",
			sourcePaths:     []string{"/app/bin"},
			excludePatterns: []string{"**/*.debug"},
			expected:        []string{"app/bin/**/*.debug"},
		},
		{
			name:            "parents pivot point in source path",
			sourcePaths:     []string{"./aaa/./bbb"},
			excludePatterns: []string{"*.md"},
			expected:        []string{"aaa/bbb/*.md"},
		},
		{
			name:            "negated pattern",
			sourcePaths:     []string{"./aaa"},
			excludePatterns: []string{"*.md", "!keep.md"},
			expected:        []string{"aaa/*.md", "!aaa/keep.md"},
		},
		{
			name:            "remote source path skipped",
			sourcePaths:     []string{"https://example.com/x.tar", "./aaa"},
			excludePatterns: []string{"*.md"},
			expected:        []string{"aaa/*.md"},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.expected, contextRelativeExcludes(tt.sourcePaths, tt.excludePatterns))
		})
	}
}
