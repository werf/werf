package path_matcher

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("dockerfile ignore path matcher", func() {
	type entry struct {
		dockerignorePatterns        []string
		testPath                    string
		isPathMatched               bool
		shouldGoThrough             bool
		isDirOrSubmodulePathMatched bool
	}

	itBodyFunc := func(e entry) {
		matcher := newDockerfileIgnorePathMatcher(e.dockerignorePatterns)

		Ω(matcher.IsPathMatched(e.testPath)).Should(BeEquivalentTo(e.isPathMatched))
		Ω(matcher.ShouldGoThrough(e.testPath)).Should(BeEquivalentTo(e.shouldGoThrough))
		Ω(matcher.IsDirOrSubmodulePathMatched(e.testPath)).Should(BeEquivalentTo(e.isDirOrSubmodulePathMatched))
	}

	DescribeTable("empty pattern matcher", itBodyFunc,
		Entry("empty test path", entry{
			dockerignorePatterns:        []string{},
			testPath:                    "",
			isPathMatched:               true,
			shouldGoThrough:             false,
			isDirOrSubmodulePathMatched: true,
		}),
		Entry("any test path", entry{
			dockerignorePatterns:        []string{},
			testPath:                    "any",
			isPathMatched:               true,
			shouldGoThrough:             false,
			isDirOrSubmodulePathMatched: true,
		}),
	)

	DescribeTable("non-empty pattern matcher", itBodyFunc,
		Entry("empty test path (1)", entry{
			dockerignorePatterns:        []string{"*"},
			testPath:                    "",
			isPathMatched:               false,
			shouldGoThrough:             true,
			isDirOrSubmodulePathMatched: true,
		}),
		Entry("empty test path (2)", entry{
			dockerignorePatterns:        []string{"any"},
			testPath:                    "",
			isPathMatched:               true,
			shouldGoThrough:             true,
			isDirOrSubmodulePathMatched: true,
		}),
		Entry("matched test path (1)", entry{
			dockerignorePatterns:        []string{"dir1"},
			testPath:                    "dir2",
			isPathMatched:               true,
			shouldGoThrough:             false,
			isDirOrSubmodulePathMatched: true,
		}),
		Entry("matched test path (1)", entry{
			dockerignorePatterns:        []string{"dir", "!dir/file"},
			testPath:                    "dir/file",
			isPathMatched:               true,
			shouldGoThrough:             false,
			isDirOrSubmodulePathMatched: true,
		}),
		Entry("not matched test path (1)", entry{
			dockerignorePatterns:        []string{"dir"},
			testPath:                    "dir",
			isPathMatched:               false,
			shouldGoThrough:             false,
			isDirOrSubmodulePathMatched: false,
		}),
		Entry("not matched test path (1)", entry{
			dockerignorePatterns:        []string{"dir"},
			testPath:                    "dir/file",
			isPathMatched:               false,
			shouldGoThrough:             false,
			isDirOrSubmodulePathMatched: false,
		}),
		Entry("not matched test path (2)", entry{
			dockerignorePatterns:        []string{"dir", "!dir/file"},
			testPath:                    "dir",
			isPathMatched:               false,
			shouldGoThrough:             true,
			isDirOrSubmodulePathMatched: true,
		}),
	)
})
