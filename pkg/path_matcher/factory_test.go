package path_matcher

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("factory", func() {
	type entry struct {
		matcher                     PathMatcher
		testPath                    string
		isPathMatched               bool
		shouldGoThrough             bool
		isDirOrSubmodulePathMatched bool
	}

	itBodyFunc := func(e entry) {
		Ω(e.matcher.IsPathMatched(e.testPath)).Should(BeEquivalentTo(e.isPathMatched))
		Ω(e.matcher.ShouldGoThrough(e.testPath)).Should(BeEquivalentTo(e.shouldGoThrough))
		Ω(e.matcher.IsDirOrSubmodulePathMatched(e.testPath)).Should(BeEquivalentTo(e.isDirOrSubmodulePathMatched))
	}

	DescribeTable("each one", itBodyFunc,
		Entry("bath path (1)", entry{
			matcher:                     NewPathMatcher(PathMatcherOptions{BasePath: "dir"}),
			testPath:                    "",
			isPathMatched:               false,
			shouldGoThrough:             true,
			isDirOrSubmodulePathMatched: true,
		}),
		Entry("bath path (1)", entry{
			matcher:                     NewPathMatcher(PathMatcherOptions{BasePath: "dir"}),
			testPath:                    "dir",
			isPathMatched:               true,
			shouldGoThrough:             false,
			isDirOrSubmodulePathMatched: true,
		}),
		Entry("include (1)", entry{
			matcher:                     NewPathMatcher(PathMatcherOptions{IncludeGlobs: []string{"dir"}}),
			testPath:                    "",
			isPathMatched:               false,
			shouldGoThrough:             true,
			isDirOrSubmodulePathMatched: true,
		}),
		Entry("include (2)", entry{
			matcher:                     NewPathMatcher(PathMatcherOptions{BasePath: "dir"}),
			testPath:                    "dir",
			isPathMatched:               true,
			shouldGoThrough:             false,
			isDirOrSubmodulePathMatched: true,
		}),
		Entry("exclude (1)", entry{
			matcher:                     NewPathMatcher(PathMatcherOptions{ExcludeGlobs: []string{"dir1"}}),
			testPath:                    "dir1",
			isPathMatched:               false,
			shouldGoThrough:             false,
			isDirOrSubmodulePathMatched: false,
		}),
		Entry("exclude (2)", entry{
			matcher:                     NewPathMatcher(PathMatcherOptions{ExcludeGlobs: []string{"dir1"}}),
			testPath:                    "dir2",
			isPathMatched:               true,
			shouldGoThrough:             false,
			isDirOrSubmodulePathMatched: true,
		}),
		Entry("dockerfile ignore (1)", entry{
			matcher:                     NewPathMatcher(PathMatcherOptions{DockerignorePatterns: []string{"dir"}}),
			testPath:                    "dir1",
			isPathMatched:               true,
			shouldGoThrough:             false,
			isDirOrSubmodulePathMatched: true,
		}),
		Entry("matcher", entry{
			matcher:                     NewPathMatcher(PathMatcherOptions{Matchers: []PathMatcher{NewTruePathMatcher()}}),
			testPath:                    "any",
			isPathMatched:               true,
			shouldGoThrough:             false,
			isDirOrSubmodulePathMatched: true,
		}),
	)

	DescribeTable("complex", itBodyFunc,
		Entry("complex 1 (1)", entry{
			matcher: NewPathMatcher(PathMatcherOptions{
				BasePath:             "dir",
				IncludeGlobs:         []string{"sub-dir"},
				ExcludeGlobs:         []string{"sub-dir/file1"},
				DockerignorePatterns: []string{"*.tmp"},
			}),
			testPath:                    "",
			isPathMatched:               false,
			shouldGoThrough:             true,
			isDirOrSubmodulePathMatched: true,
		}),
		Entry("complex 1 (2)", entry{
			matcher: NewPathMatcher(PathMatcherOptions{
				BasePath:             "dir",
				IncludeGlobs:         []string{"sub-dir"},
				ExcludeGlobs:         []string{"sub-dir/file1"},
				DockerignorePatterns: []string{"*.tmp"},
			}),
			testPath:                    "sub-dir",
			isPathMatched:               false,
			shouldGoThrough:             false,
			isDirOrSubmodulePathMatched: false,
		}),
		Entry("complex 1 (3)", entry{
			matcher: NewPathMatcher(PathMatcherOptions{
				BasePath:             "dir",
				IncludeGlobs:         []string{"sub-dir"},
				ExcludeGlobs:         []string{"sub-dir/file1"},
				DockerignorePatterns: []string{"*.tmp"},
			}),
			testPath:                    "dir/sub-dir",
			isPathMatched:               true,
			shouldGoThrough:             true,
			isDirOrSubmodulePathMatched: true,
		}),
		Entry("complex 1 (4)", entry{
			matcher: NewPathMatcher(PathMatcherOptions{
				BasePath:             "dir",
				IncludeGlobs:         []string{"sub-dir"},
				ExcludeGlobs:         []string{"sub-dir/file1"},
				DockerignorePatterns: []string{"*.tmp"},
			}),
			testPath:                    "dir/sub-dir/file1",
			isPathMatched:               false,
			shouldGoThrough:             false,
			isDirOrSubmodulePathMatched: false,
		}),
		Entry("complex 1 (5)", entry{
			matcher: NewPathMatcher(PathMatcherOptions{
				BasePath:             "dir",
				IncludeGlobs:         []string{"sub-dir"},
				ExcludeGlobs:         []string{"sub-dir/file1"},
				DockerignorePatterns: []string{"*.tmp"},
			}),
			testPath:                    "dir/sub-dir/file2",
			isPathMatched:               true,
			shouldGoThrough:             false,
			isDirOrSubmodulePathMatched: true,
		}),
		Entry("complex 1 (6)", entry{
			matcher: NewPathMatcher(PathMatcherOptions{
				BasePath:             "dir",
				IncludeGlobs:         []string{"sub-dir"},
				ExcludeGlobs:         []string{"sub-dir/file1"},
				DockerignorePatterns: []string{"sub-dir/*.tmp"},
			}),
			testPath:                    "dir/sub-dir/file.tmp",
			isPathMatched:               false,
			shouldGoThrough:             false,
			isDirOrSubmodulePathMatched: false,
		}),
		Entry("complex 2 (1)", entry{
			matcher: NewPathMatcher(PathMatcherOptions{
				BasePath: "dir",
				Matchers: []PathMatcher{
					newBasePathMatcher("sub-dir", nil),
					newIncludePathMatcher([]string{"sub-dir/*.go"}),
				},
			}),
			testPath:                    "dir",
			isPathMatched:               false,
			shouldGoThrough:             true,
			isDirOrSubmodulePathMatched: true,
		}),
		Entry("complex 2 (2)", entry{
			matcher: NewPathMatcher(PathMatcherOptions{
				BasePath: "dir",
				Matchers: []PathMatcher{
					newBasePathMatcher("sub-dir", nil),
					newIncludePathMatcher([]string{"sub-dir/*.go"}),
				},
			}),
			testPath:                    "dir/sub-dir/file",
			isPathMatched:               false,
			shouldGoThrough:             false,
			isDirOrSubmodulePathMatched: false,
		}),
		Entry("complex 2 (3)", entry{
			matcher: NewPathMatcher(PathMatcherOptions{
				BasePath: "dir",
				Matchers: []PathMatcher{
					newBasePathMatcher("sub-dir", nil),
					newIncludePathMatcher([]string{"sub-dir/*.go"}),
				},
			}),
			testPath:                    "dir/sub-dir/file.go",
			isPathMatched:               true,
			shouldGoThrough:             false,
			isDirOrSubmodulePathMatched: true,
		}),
	)
})
