package path_matcher

import (
	"context"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/samber/lo"
)

var _ = ginkgo.Describe("factory", func() {
	type entry struct {
		matcher                     PathMatcher
		testPath                    string
		isPathMatched               bool
		shouldGoThrough             bool
		isDirOrSubmodulePathMatched bool
	}

	itBodyFunc := func(e entry) {
		gomega.Expect(e.matcher.IsPathMatched(e.testPath)).Should(gomega.BeEquivalentTo(e.isPathMatched))
		gomega.Expect(e.matcher.ShouldGoThrough(e.testPath)).Should(gomega.BeEquivalentTo(e.shouldGoThrough))
		gomega.Expect(e.matcher.IsDirOrSubmodulePathMatched(e.testPath)).Should(gomega.BeEquivalentTo(e.isDirOrSubmodulePathMatched))
	}

	ginkgo.DescribeTable("each one", itBodyFunc,
		ginkgo.Entry("bath path (1)", entry{
			matcher:                     NewPathMatcher(PathMatcherOptions{BasePath: "dir"}),
			testPath:                    "",
			isPathMatched:               false,
			shouldGoThrough:             true,
			isDirOrSubmodulePathMatched: true,
		}),
		ginkgo.Entry("bath path (1)", entry{
			matcher:                     NewPathMatcher(PathMatcherOptions{BasePath: "dir"}),
			testPath:                    "dir",
			isPathMatched:               true,
			shouldGoThrough:             false,
			isDirOrSubmodulePathMatched: true,
		}),
		ginkgo.Entry("include (1)", entry{
			matcher:                     NewPathMatcher(PathMatcherOptions{IncludeGlobs: []string{"dir"}}),
			testPath:                    "",
			isPathMatched:               false,
			shouldGoThrough:             true,
			isDirOrSubmodulePathMatched: true,
		}),
		ginkgo.Entry("include (2)", entry{
			matcher:                     NewPathMatcher(PathMatcherOptions{BasePath: "dir"}),
			testPath:                    "dir",
			isPathMatched:               true,
			shouldGoThrough:             false,
			isDirOrSubmodulePathMatched: true,
		}),
		ginkgo.Entry("exclude (1)", entry{
			matcher:                     NewPathMatcher(PathMatcherOptions{ExcludeGlobs: []string{"dir1"}}),
			testPath:                    "dir1",
			isPathMatched:               false,
			shouldGoThrough:             false,
			isDirOrSubmodulePathMatched: false,
		}),
		ginkgo.Entry("exclude (2)", entry{
			matcher:                     NewPathMatcher(PathMatcherOptions{ExcludeGlobs: []string{"dir1"}}),
			testPath:                    "dir2",
			isPathMatched:               true,
			shouldGoThrough:             false,
			isDirOrSubmodulePathMatched: true,
		}),
		ginkgo.Entry("dockerfile ignore (1)", entry{
			matcher:                     NewPathMatcher(PathMatcherOptions{Matchers: []PathMatcher{lo.Must(NewDockerfileIgnorePathMatcher(context.Background(), []string{"dir"}))}}),
			testPath:                    "dir1",
			isPathMatched:               true,
			shouldGoThrough:             false,
			isDirOrSubmodulePathMatched: true,
		}),
		ginkgo.Entry("matcher", entry{
			matcher:                     NewPathMatcher(PathMatcherOptions{Matchers: []PathMatcher{NewTruePathMatcher()}}),
			testPath:                    "any",
			isPathMatched:               true,
			shouldGoThrough:             false,
			isDirOrSubmodulePathMatched: true,
		}),
	)

	ginkgo.DescribeTable("complex", itBodyFunc,
		ginkgo.Entry("complex 1 (1)", entry{
			matcher: NewPathMatcher(PathMatcherOptions{
				BasePath:     "dir",
				IncludeGlobs: []string{"sub-dir"},
				ExcludeGlobs: []string{"sub-dir/file1"},
				Matchers:     []PathMatcher{lo.Must(NewDockerfileIgnorePathMatcher(context.Background(), []string{"*.tmp"}))},
			}),
			testPath:                    "",
			isPathMatched:               false,
			shouldGoThrough:             true,
			isDirOrSubmodulePathMatched: true,
		}),
		ginkgo.Entry("complex 1 (2)", entry{
			matcher: NewPathMatcher(PathMatcherOptions{
				BasePath:     "dir",
				IncludeGlobs: []string{"sub-dir"},
				ExcludeGlobs: []string{"sub-dir/file1"},
				Matchers:     []PathMatcher{lo.Must(NewDockerfileIgnorePathMatcher(context.Background(), []string{"*.tmp"}))},
			}),
			testPath:                    "sub-dir",
			isPathMatched:               false,
			shouldGoThrough:             false,
			isDirOrSubmodulePathMatched: false,
		}),
		ginkgo.Entry("complex 1 (3)", entry{
			matcher: NewPathMatcher(PathMatcherOptions{
				BasePath:     "dir",
				IncludeGlobs: []string{"sub-dir"},
				ExcludeGlobs: []string{"sub-dir/file1"},
				Matchers:     []PathMatcher{lo.Must(NewDockerfileIgnorePathMatcher(context.Background(), []string{"*.tmp"}))},
			}),
			testPath:                    "dir/sub-dir",
			isPathMatched:               true,
			shouldGoThrough:             true,
			isDirOrSubmodulePathMatched: true,
		}),
		ginkgo.Entry("complex 1 (4)", entry{
			matcher: NewPathMatcher(PathMatcherOptions{
				BasePath:     "dir",
				IncludeGlobs: []string{"sub-dir"},
				ExcludeGlobs: []string{"sub-dir/file1"},
				Matchers:     []PathMatcher{lo.Must(NewDockerfileIgnorePathMatcher(context.Background(), []string{"*.tmp"}))},
			}),
			testPath:                    "dir/sub-dir/file1",
			isPathMatched:               false,
			shouldGoThrough:             false,
			isDirOrSubmodulePathMatched: false,
		}),
		ginkgo.Entry("complex 1 (5)", entry{
			matcher: NewPathMatcher(PathMatcherOptions{
				BasePath:     "dir",
				IncludeGlobs: []string{"sub-dir"},
				ExcludeGlobs: []string{"sub-dir/file1"},
				Matchers:     []PathMatcher{lo.Must(NewDockerfileIgnorePathMatcher(context.Background(), []string{"*.tmp"}))},
			}),
			testPath:                    "dir/sub-dir/file2",
			isPathMatched:               true,
			shouldGoThrough:             false,
			isDirOrSubmodulePathMatched: true,
		}),
		ginkgo.Entry("complex 1 (6)", entry{
			matcher: NewPathMatcher(PathMatcherOptions{
				BasePath:     "dir",
				IncludeGlobs: []string{"sub-dir"},
				ExcludeGlobs: []string{"sub-dir/file1"},
				Matchers:     []PathMatcher{lo.Must(NewDockerfileIgnorePathMatcher(context.Background(), []string{"sub-dir/*.tmp"}))},
			}),
			testPath:                    "dir/sub-dir/file.tmp",
			isPathMatched:               false,
			shouldGoThrough:             false,
			isDirOrSubmodulePathMatched: false,
		}),
		ginkgo.Entry("complex 2 (1)", entry{
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
		ginkgo.Entry("complex 2 (2)", entry{
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
		ginkgo.Entry("complex 2 (3)", entry{
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
