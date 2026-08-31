package stage

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/werf/werf/v2/pkg/git_repo"
	"github.com/werf/werf/v2/pkg/image"
)

type gitRepoStringStub struct {
	git_repo.GitRepo
}

func (r *gitRepoStringStub) String() string {
	return "test"
}

func (r *gitRepoStringStub) GetName() string {
	return "test"
}

var _ = Describe("GitStage", func() {
	It("selects the oldest cached content anchor without its git commit", func() {
		gitMapping := NewGitMapping()
		gitMapping.SetGitRepo(&gitRepoStringStub{})

		stg := newGitStage(GitLatestPatch, &BaseStageOptions{})
		stg.SetGitMappings([]*GitMapping{gitMapping})
		stg.SetContentAnchor(true)

		oldest := &image.StageDesc{
			StageID: image.NewStageID("anchor", 1),
			Info:    &image.Info{Name: "repo:oldest"},
		}
		newest := &image.StageDesc{
			StageID: image.NewStageID("anchor", 2),
			Info:    &image.Info{Name: "repo:newest"},
		}

		selected, err := stg.SelectSuitableStageDesc(context.Background(), nil, image.NewStageDescSet(newest, oldest))

		Expect(err).NotTo(HaveOccurred())
		Expect(selected).To(BeIdenticalTo(oldest))
	})
})
