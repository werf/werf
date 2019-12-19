package true_git

import (
	"fmt"
	"path/filepath"

	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

type entry struct {
	baseBase      string
	includePaths  []string
	excludePaths  []string
	validPaths    []string
	notValidPaths []string
}

var _ = DescribeTable("path filter",
	func(e entry) {
		pm := PathFilter{BasePath: e.baseBase, IncludePaths: e.includePaths, ExcludePaths: e.excludePaths}

		for _, path := range e.validPaths {
			Ω(pm.IsFilePathValid(path)).Should(BeTrue(), fmt.Sprintf("path %s should be valid", path))
		}

		for _, path := range e.notValidPaths {
			Ω(pm.IsFilePathValid(path)).Should(BeFalse(), fmt.Sprintf("path %s should not be valid", path))
		}
	},
	Entry("base path (empty)", entry{
		baseBase:      "",
		includePaths:  []string{},
		excludePaths:  []string{},
		validPaths:    []string{"a", filepath.Join("abc", "b")},
		notValidPaths: []string{},
	}),
	Entry("base path (abc)", entry{
		baseBase:      "abc",
		includePaths:  []string{},
		excludePaths:  []string{},
		validPaths:    []string{"abc", filepath.Join("abc", "a")},
		notValidPaths: []string{"a"},
	}),
	Entry("include paths (a)", entry{
		baseBase:      "",
		includePaths:  []string{"a"},
		excludePaths:  []string{},
		validPaths:    []string{"a", filepath.Join("a", "b", "c")},
		notValidPaths: []string{"b", "ab"},
	}),
	Entry("exclude paths (a)", entry{
		baseBase:      "",
		includePaths:  []string{},
		excludePaths:  []string{"a"},
		validPaths:    []string{"ab", "b", filepath.Join("b", "a")},
		notValidPaths: []string{"a", filepath.Join("a", "b", "c")},
	}),
	Entry("include paths (globs)", entry{
		baseBase: "",
		includePaths: []string{
			filepath.Join("a", "b", "*"),
			filepath.Join("c", "[dt]ata.txt"),
			filepath.Join(filepath.Join("d", "*.rb")),
			filepath.Join(filepath.Join("e", "**", "*.bin")),
		},
		excludePaths: []string{},
		validPaths: []string{
			filepath.Join("a", "b", "c"), filepath.Join("a", "b", "c", "d"),
			filepath.Join("c", "data.txt"),
			filepath.Join(filepath.Join("d", "file.rb")),
			filepath.Join(filepath.Join("e", "file.bin"), filepath.Join("e", "f", "file.bin")),
		},
		notValidPaths: []string{
			filepath.Join("a", "c"),
			filepath.Join("c", "gata.txt"),
			filepath.Join(filepath.Join("d", "e", "file.rb")),
			filepath.Join("file.bin"),
		},
	}),
	Entry("exclude paths (globs)", entry{
		baseBase:     "",
		includePaths: []string{},
		excludePaths: []string{
			filepath.Join("a", "b", "*"),
			filepath.Join("c", "[dt]ata.txt"),
			filepath.Join(filepath.Join("d", "*.rb")),
			filepath.Join(filepath.Join("e", "**", "*.bin")),
		},
		validPaths: []string{
			filepath.Join("a", "c"),
			filepath.Join("c", "gata.txt"),
			filepath.Join(filepath.Join("d", "e", "file.rb")),
			filepath.Join("file.bin"),
		},
		notValidPaths: []string{
			filepath.Join("a", "b", "c"), filepath.Join("a", "b", "c", "d"),
			filepath.Join("c", "data.txt"),
			filepath.Join(filepath.Join("d", "file.rb")),
			filepath.Join(filepath.Join("e", "file.bin"), filepath.Join("e", "f", "file.bin")),
		},
	}),
	Entry("base path (a), include paths (b)", entry{
		baseBase:      "a",
		includePaths:  []string{"b"},
		excludePaths:  []string{},
		validPaths:    []string{filepath.Join("a", "b")},
		notValidPaths: []string{filepath.Join("a", "c")},
	}),
	Entry("base path (a), exclude paths (b)", entry{
		baseBase:      "a",
		includePaths:  []string{},
		excludePaths:  []string{"b"},
		validPaths:    []string{filepath.Join("a", "c")},
		notValidPaths: []string{filepath.Join("a", "b")},
	}),
)
