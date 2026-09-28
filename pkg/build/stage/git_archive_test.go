package stage_test

import (
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"

	"github.com/werf/common-go/pkg/util"
	"github.com/werf/werf/v3/pkg/build/stage"
)

var _ = ginkgo.DescribeTable("Git archive content dependencies",
	func(ctx ginkgo.SpecContext, contents [2]string, swapMappings, swapDestinations, sameDigest bool) {
		repos := []*gitArchiveRepoStub{
			{GitRepoStub: NewGitRepoStub("repo-a", false, "commit-a"), checksum: util.Sha256Hash("alpha")},
			{GitRepoStub: NewGitRepoStub("repo-b", false, "commit-b"), checksum: util.Sha256Hash("beta")},
		}
		mappings := []*stage.GitMapping{stage.NewGitMapping(), stage.NewGitMapping()}
		for i, mapping := range mappings {
			mapping.SetGitRepo(repos[i])
			mapping.Add = "/"
			mapping.To = "/" + repos[i].GetName()
		}

		stg := stage.NewGitArchiveStage(&stage.NewGitArchiveStageOptions{}, &stage.BaseStageOptions{})
		stg.SetGitMappings(mappings)
		conveyor := NewConveyorStub()
		before, err := stg.GetContentDependencies(ctx, conveyor, nil)
		gomega.Expect(err).To(gomega.Succeed())
		gomega.Expect(before).NotTo(gomega.BeEmpty())

		for i, repo := range repos {
			repo.headCommitHash += "-new"
			repo.checksum = util.Sha256Hash(contents[i])
		}
		if swapMappings {
			stg.SetGitMappings([]*stage.GitMapping{mappings[1], mappings[0]})
		}
		if swapDestinations {
			mappings[0].To, mappings[1].To = mappings[1].To, mappings[0].To
		}

		after, err := stg.GetContentDependencies(ctx, conveyor, nil)
		gomega.Expect(err).To(gomega.Succeed())
		if sameDigest {
			gomega.Expect(after).To(gomega.Equal(before))
		} else {
			gomega.Expect(after).NotTo(gomega.Equal(before))
		}
	},
	ginkgo.Entry("ignores new commits with unchanged content", [2]string{"alpha", "beta"}, false, false, true),
	ginkgo.Entry("detects content swapped between mappings", [2]string{"beta", "alpha"}, false, false, false),
	ginkgo.Entry("detects changed content", [2]string{"gamma", "beta"}, false, false, false),
	ginkgo.Entry("ignores mapping order", [2]string{"alpha", "beta"}, true, false, true),
	ginkgo.Entry("detects changed destinations", [2]string{"alpha", "beta"}, false, true, false),
)
