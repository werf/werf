package stage

import (
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"

	"github.com/werf/common-go/pkg/util"
	"github.com/werf/werf/v3/pkg/path_matcher"
)

var _ = ginkgo.Describe("automatic LFS Dockerfile cache", func() {
	ginkgo.DescribeTable("does not reuse a legacy image built from pointer files",
		func(ctx ginkgo.SpecContext, command string, legacyDependencies []string) {
			manager := NewGiterminismManagerStub(NewLocalGitRepoStub("pointer-checksum"), NewGiterminismInspectorStub())
			conveyor := NewConveyorStubForDependencies(manager, []*TestDependency{})
			dockerfileData := []byte("FROM alpine\n" + command + "\n")
			dockerStages, dockerMetaArgs := testDockerfileToDockerStages(dockerfileData)
			stage := newTestFullDockerfileStage(dockerfileData, "", map[string]interface{}{}, dockerStages, dockerMetaArgs, []*TestDependency{}, "")
			stage.ContextChecksum = NewContextChecksum(path_matcher.NewPathMatcher(path_matcher.PathMatcherOptions{}))

			digest, err := stage.dependenciesDigest(ctx, conveyor, map[string]string{})
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
			gomega.Expect(digest).NotTo(gomega.Equal(util.Sha256Hash(legacyDependencies...)))
		},
		ginkgo.Entry("COPY", "COPY assets/ /app/", []string{
			"alpine", "COPY assets/ /app/", util.Sha256Hash("pointer-checksum"),
		}),
		ginkgo.Entry("RUN with a context bind mount and no COPY", "RUN --mount=type=bind,target=/src cat /src/payload > /out", []string{
			"alpine", "RUN --mount=type=bind,target=/src cat /src/payload > /out",
		}),
	)
})
