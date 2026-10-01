package image

import (
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
)

var _ = ginkgo.Describe("Dockerfile ignore errors", func() {
	ginkgo.DescribeTable("reports the selected ignore file and invalid pattern", func(ctx ginkgo.SpecContext, ignoreFile string) {
		projectDir := newProjectRepo(ctx, map[string]string{
			"app/Dockerfile":    "FROM scratch\n",
			"app/" + ignoreFile: "Dockerfile\narchive-old\\\n",
		})
		matcher, err := createDockerIgnorePathMatcher(ctx, *giterminismManagerOf(ctx, projectDir), "app", "Dockerfile")
		gomega.Expect(err).To(gomega.MatchError(`read ignore file "app/` + ignoreFile + `": parse ignore pattern "archive-old\\": syntax error in pattern`))
		gomega.Expect(matcher).To(gomega.BeNil())
	},
		ginkgo.Entry("context dockerignore", ".dockerignore"),
		ginkgo.Entry("Dockerfile dockerignore", "Dockerfile.dockerignore"),
		ginkgo.Entry("context containerignore", ".containerignore"),
		ginkgo.Entry("Dockerfile containerignore", "Dockerfile.containerignore"),
	)
})
