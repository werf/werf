package path_matcher

import (
	"context"
	"fmt"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
)

var _ = ginkgo.Describe("dockerfile ignore path matcher", func() {
	type entry struct {
		dockerignorePatterns        []string
		testPath                    string
		isPathMatched               bool
		shouldGoThrough             bool
		isDirOrSubmodulePathMatched bool
	}

	itBodyFunc := func(e entry) {
		matcher, err := NewDockerfileIgnorePathMatcher(context.Background(), e.dockerignorePatterns)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())

		gomega.Expect(matcher.IsPathMatched(e.testPath)).Should(gomega.BeEquivalentTo(e.isPathMatched))
		gomega.Expect(matcher.ShouldGoThrough(e.testPath)).Should(gomega.BeEquivalentTo(e.shouldGoThrough))
		gomega.Expect(matcher.IsDirOrSubmodulePathMatched(e.testPath)).Should(gomega.BeEquivalentTo(e.isDirOrSubmodulePathMatched))
	}

	ginkgo.DescribeTable("invalid ignore patterns", func(pattern string) {
		matcher, err := NewDockerfileIgnorePathMatcher(context.Background(), []string{"Dockerfile", pattern})
		gomega.Expect(err).To(gomega.HaveOccurred())
		gomega.Expect(err.Error()).To(gomega.ContainSubstring(fmt.Sprintf("%q", pattern)))
		gomega.Expect(matcher).To(gomega.BeNil())
	},
		ginkgo.Entry("trailing escape", `archive-old\`),
		ginkgo.Entry("unclosed character class", "["),
		ginkgo.Entry("empty exclusion", "!"),
		ginkgo.Entry("invalid character range", "[z-a]"),
		ginkgo.Entry("invalid exclusion character range", "![z-a]"),
	)

	ginkgo.DescribeTable("empty pattern matcher", itBodyFunc,
		ginkgo.Entry("empty test path", entry{
			dockerignorePatterns:        []string{},
			testPath:                    "",
			isPathMatched:               true,
			shouldGoThrough:             false,
			isDirOrSubmodulePathMatched: true,
		}),
		ginkgo.Entry("any test path", entry{
			dockerignorePatterns:        []string{},
			testPath:                    "any",
			isPathMatched:               true,
			shouldGoThrough:             false,
			isDirOrSubmodulePathMatched: true,
		}),
	)

	ginkgo.DescribeTable("non-empty pattern matcher", itBodyFunc,
		ginkgo.Entry("empty test path (1)", entry{
			dockerignorePatterns:        []string{"*"},
			testPath:                    "",
			isPathMatched:               false,
			shouldGoThrough:             true,
			isDirOrSubmodulePathMatched: true,
		}),
		ginkgo.Entry("empty test path (2)", entry{
			dockerignorePatterns:        []string{"any"},
			testPath:                    "",
			isPathMatched:               true,
			shouldGoThrough:             true,
			isDirOrSubmodulePathMatched: true,
		}),
		ginkgo.Entry("matched test path (1)", entry{
			dockerignorePatterns:        []string{"dir1"},
			testPath:                    "dir2",
			isPathMatched:               true,
			shouldGoThrough:             false,
			isDirOrSubmodulePathMatched: true,
		}),
		ginkgo.Entry("matched test path (1)", entry{
			dockerignorePatterns:        []string{"dir", "!dir/file"},
			testPath:                    "dir/file",
			isPathMatched:               true,
			shouldGoThrough:             false,
			isDirOrSubmodulePathMatched: true,
		}),
		ginkgo.Entry("not matched test path (1)", entry{
			dockerignorePatterns:        []string{"dir"},
			testPath:                    "dir",
			isPathMatched:               false,
			shouldGoThrough:             false,
			isDirOrSubmodulePathMatched: false,
		}),
		ginkgo.Entry("not matched test path (1)", entry{
			dockerignorePatterns:        []string{"dir"},
			testPath:                    "dir/file",
			isPathMatched:               false,
			shouldGoThrough:             false,
			isDirOrSubmodulePathMatched: false,
		}),
		ginkgo.Entry("not matched test path (2)", entry{
			dockerignorePatterns:        []string{"dir", "!dir/file"},
			testPath:                    "dir",
			isPathMatched:               false,
			shouldGoThrough:             true,
			isDirOrSubmodulePathMatched: true,
		}),
	)
})
