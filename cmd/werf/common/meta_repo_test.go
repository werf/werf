package common

import (
	"context"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/samber/lo"
	"github.com/spf13/cobra"

	"github.com/werf/werf/v2/pkg/container_backend"
	"github.com/werf/werf/v2/pkg/storage"
	"github.com/werf/werf/v2/pkg/werf"
)

var _ = ginkgo.Describe("meta repository routing", func() {
	ginkgo.DescribeTable("preserves the primary storage when no separate repository is selected",
		func(repo, meta string) {
			backend := &container_backend.DockerServerBackend{}
			stages := storage.NewLocalStagesStorage(backend)
			actual, err := GetOptionalMetaStorage(context.Background(), backend, &CmdData{
				Repo:     &RepoData{Address: lo.ToPtr(repo)},
				MetaRepo: &RepoData{Address: lo.ToPtr(meta)},
			}, stages, false)
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
			gomega.Expect(actual).To(gomega.BeIdenticalTo(stages))
		},
		ginkgo.Entry("local build", "", ""),
		ginkgo.Entry("unset meta repo", "example.com/project", ""),
		ginkgo.Entry("same address", "example.com/project", "example.com/project"),
		ginkgo.Entry("equivalent tagged address", "example.com/project:old", "example.com/project:new"),
	)

	ginkgo.DescribeTable("validates separate metadata repositories",
		func(repo, meta string, separate bool, errorSubstring string) {
			actual, err := hasSeparateMetaRepo(context.Background(), &CmdData{
				Repo:     &RepoData{Address: lo.ToPtr(repo)},
				MetaRepo: &RepoData{Address: lo.ToPtr(meta)},
			})
			if errorSubstring != "" {
				gomega.Expect(err).To(gomega.MatchError(gomega.ContainSubstring(errorSubstring)))
				return
			}
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
			gomega.Expect(actual).To(gomega.Equal(separate))
		},
		ginkgo.Entry("separate", "example.com/project", "example.com/metadata", true, ""),
		ginkgo.Entry("requires repo", "", "example.com/metadata", false, "--meta-repo requires --repo to be set"),
		ginkgo.Entry("rejects local repo", ":local", "example.com/metadata", false, "parse --repo"),
		ginkgo.Entry("rejects local metadata", "example.com/project", ":local", false, "parse --meta-repo"),
	)

	ginkgo.It("routes managed images and enforces the marker through real storage construction", func(ctx ginkgo.SpecContext) {
		gomega.Expect(werf.Init(ginkgo.GinkgoT().TempDir(), ginkgo.GinkgoT().TempDir())).To(gomega.Succeed())
		repo, _ := startRecordingRegistry("test/stages")
		metaRepo, _ := startRecordingRegistry("test/meta")
		backend := &container_backend.DockerServerBackend{}
		cmdData := storageManagerCmdData("--repo", repo, "--meta-repo", metaRepo, "--insecure-registry", "--check-built-images")
		manager, err := NewStorageManager(ctx, &NewStorageManagerConfig{ProjectName: "project", ContainerBackend: backend, CmdData: cmdData})
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		gomega.Expect(manager.GetMetaStorage(ctx).AddManagedImage(ctx, "project", "app")).To(gomega.Succeed())
		images, err := manager.GetMetaStorage(ctx).GetManagedImages(ctx, "project")
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		gomega.Expect(images).To(gomega.ConsistOf("app"))
		stagesImages, err := manager.GetStagesStorage().GetManagedImages(ctx, "project")
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		gomega.Expect(stagesImages).To(gomega.BeEmpty())

		cmdData = storageManagerCmdData("--repo", repo, "--insecure-registry", "--check-built-images")
		_, err = NewStorageManager(ctx, &NewStorageManagerConfig{ProjectName: "project", ContainerBackend: backend, CmdData: cmdData})
		gomega.Expect(err).To(gomega.MatchError(gomega.ContainSubstring(metaRepo)))
	})

	ginkgo.It("reads WERF_META_REPO and permits a flag override", func() {
		ginkgo.GinkgoT().Setenv("WERF_META_REPO", "example.com/environment")
		cmdData := &CmdData{}
		cmd := &cobra.Command{}
		SetupMetaRepo(context.Background(), cmdData, cmd)
		gomega.Expect(*cmdData.MetaRepo.Address).To(gomega.Equal("example.com/environment"))
		gomega.Expect(cmd.Flags().Parse([]string{"--meta-repo=example.com/flag"})).To(gomega.Succeed())
		gomega.Expect(*cmdData.MetaRepo.Address).To(gomega.Equal("example.com/flag"))
	})
})
