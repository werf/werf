package common

import (
	ginkgo "github.com/onsi/ginkgo/v2"
	gomega "github.com/onsi/gomega"

	"github.com/werf/werf/v3/pkg/container_backend"
	"github.com/werf/werf/v3/pkg/storage"
	"github.com/werf/werf/v3/pkg/werf"
)

var _ = ginkgo.Describe("Meta repository with local storage", func() {
	ginkgo.BeforeEach(func() {
		clearListEnv()
		gomega.Expect(werf.Init(ginkgo.GinkgoT().TempDir(), ginkgo.GinkgoT().TempDir())).To(gomega.Succeed())
	})

	ginkgo.DescribeTable("ignores a configured meta repository", func(ctx ginkgo.SpecContext, fromEnvironment, explicitLocal bool, metaRepo string) {
		args := []string{}
		if explicitLocal {
			args = append(args, "--repo", ":local")
		}
		if fromEnvironment {
			ginkgo.GinkgoT().Setenv("WERF_META_REPO", metaRepo)
		} else {
			args = append(args, "--meta-repo", metaRepo)
		}
		manager, err := NewStorageManager(ctx, &NewStorageManagerConfig{
			ProjectName:      "project",
			ContainerBackend: &container_backend.DockerServerBackend{},
			CmdData:          storageManagerCmdData(args...),
		})
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		gomega.Expect(manager.StagesStorage.Address()).To(gomega.Equal(storage.LocalStorageAddress))
		gomega.Expect(manager.GetMetaStorage()).To(gomega.BeIdenticalTo(manager.StagesStorage))
	},
		ginkgo.Entry("registry flag", false, true, "registry.example/meta"),
		ginkgo.Entry("registry environment", true, true, "registry.example/meta"),
		ginkgo.Entry("local flag", false, true, ":local"),
		ginkgo.Entry("local environment", true, true, ":local"),
		ginkgo.Entry("invalid flag", false, true, "not a valid repository"),
		ginkgo.Entry("invalid environment", true, true, "not a valid repository"),
		ginkgo.Entry("implicit local repo with flag", false, false, "registry.example/meta"),
		ginkgo.Entry("implicit local repo with environment", true, false, "registry.example/meta"),
	)

	ginkgo.DescribeTable("keeps metadata local when no meta repository is configured", func(ctx ginkgo.SpecContext, args []string) {
		manager, err := NewStorageManager(ctx, &NewStorageManagerConfig{
			ProjectName:      "project",
			ContainerBackend: &container_backend.DockerServerBackend{},
			CmdData:          storageManagerCmdData(args...),
		})
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		gomega.Expect(manager.StagesStorage.Address()).To(gomega.Equal(storage.LocalStorageAddress))
		gomega.Expect(manager.GetMetaStorage()).To(gomega.BeIdenticalTo(manager.StagesStorage))
	},
		ginkgo.Entry("implicit local repo", []string{}),
		ginkgo.Entry("explicit local repo", []string{"--repo", ":local"}),
		ginkgo.Entry("empty meta repo", []string{"--repo", ":local", "--meta-repo="}),
	)
})
