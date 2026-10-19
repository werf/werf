package manager

import (
	"strings"

	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/empty"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"

	"github.com/werf/werf/v3/pkg/image"
	"github.com/werf/werf/v3/pkg/storage"
	"github.com/werf/werf/v3/pkg/werf"
)

var _ = ginkgo.Describe("Multiple final storages", func() {
	ginkgo.DescribeTable("copies and checks the requested destination independently",
		func(ctx ginkgo.SpecContext, populatedDestination int, checkOnly bool) {
			gomega.Expect(werf.Init(ginkgo.GinkgoT().TempDir(), "")).To(gomega.Succeed())
			manager := newManifestStorageManager(ctx)
			finalStorages := []storage.StagesStorage{
				newManifestStorageManager(ctx).StagesStorage,
				newManifestStorageManager(ctx).StagesStorage,
			}
			manager.FinalStagesStorageList = finalStorages
			stageID := *image.NewStageID(strings.Repeat("a", 56), 1611836746968)
			for _, repo := range []storage.StagesStorage{manager.StagesStorage, finalStorages[populatedDestination]} {
				tag, err := name.NewTag(repo.ConstructStageImageName(manager.ProjectName, stageID.Digest, stageID.CreationTs))
				gomega.Expect(err).To(gomega.Succeed())
				gomega.Expect(remote.Write(tag, empty.Image, remote.WithContext(ctx))).To(gomega.Succeed())
			}
			gomega.Expect(manager.InitCache(ctx)).To(gomega.Succeed())

			for i, repo := range finalStorages {
				desc, err := manager.CopyStageIntoFinalStorage(ctx, stageID, repo, CopyStageIntoStorageOptions{ShouldBeBuiltMode: checkOnly})
				if checkOnly && i != populatedDestination {
					gomega.Expect(err).To(gomega.MatchError(gomega.ContainSubstring("is not exist in the final repo")))
					_, err = repo.GetStageDesc(ctx, manager.ProjectName, stageID)
					gomega.Expect(storage.IsErrStageNotFound(err)).To(gomega.BeTrue())
					continue
				}
				gomega.Expect(err).To(gomega.Succeed())
				gomega.Expect(desc.Info.Name).To(gomega.Equal(repo.ConstructStageImageName(manager.ProjectName, stageID.Digest, stageID.CreationTs)))
				published, err := repo.GetStageDesc(ctx, manager.ProjectName, stageID)
				gomega.Expect(err).To(gomega.Succeed())
				gomega.Expect(published.Info.GetDigest()).To(gomega.Equal(desc.Info.GetDigest()))
			}
		},
		ginkgo.Entry("first destination already populated", 0, false),
		ginkgo.Entry("second destination already populated", 1, false),
		ginkgo.Entry("check fails when second destination is empty", 0, true),
		ginkgo.Entry("check finds an image present only in second destination", 1, true),
	)
})
