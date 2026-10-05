package common

import (
	"net/http"
	"net/http/httptest"

	ginkgo "github.com/onsi/ginkgo/v2"
	gomega "github.com/onsi/gomega"
	"github.com/spf13/cobra"

	"github.com/werf/werf/v3/pkg/werf"
)

var _ = ginkgo.Describe("Storage manager initialization in check mode", func() {
	ginkgo.BeforeEach(func() {
		gomega.Expect(werf.Init(ginkgo.GinkgoT().TempDir(), ginkgo.GinkgoT().TempDir())).To(gomega.Succeed())
	})

	ginkgo.DescribeTable("writes nothing into an empty repo and never contacts the sync server", func(ctx ginkgo.SpecContext, flag string) {
		if flag == "" {
			ginkgo.GinkgoT().Setenv("WERF_CHECK_BUILT_IMAGES", "true")
		}
		address, recorder := startRecordingRegistry("test/repo")
		args := []string{"--repo", address, "--insecure-registry", "--synchronization", "https://synchronization.invalid"}
		if flag != "" {
			args = append(args, flag)
		}

		storageManager, err := NewStorageManager(ctx, &NewStorageManagerConfig{
			ProjectName: "project",
			CmdData:     storageManagerCmdData(args...),
		})

		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		gomega.Expect(storageManager.StorageLockManager).NotTo(gomega.BeNil())
		gomega.Expect(recorder.recordedWrites()).To(gomega.BeEmpty())
	},
		ginkgo.Entry("check flag", "--check-built-images"),
		ginkgo.Entry("legacy flag", "--require-built-images"),
		ginkgo.Entry("short flag", "-Z"),
		ginkgo.Entry("environment", ""),
	)

	ginkgo.It("still validates the synchronization address", func(ctx ginkgo.SpecContext) {
		address, recorder := startRecordingRegistry("test/repo")

		_, err := NewStorageManager(ctx, &NewStorageManagerConfig{
			ProjectName: "project",
			CmdData: storageManagerCmdData(
				"--repo", address,
				"--insecure-registry",
				"--synchronization", "ftp://sync.example",
				"--check-built-images",
			),
		})

		gomega.Expect(err).To(gomega.MatchError(gomega.ContainSubstring("unsupported synchronization address")))
		gomega.Expect(recorder.recordedWrites()).To(gomega.BeEmpty())
	})

	ginkgo.It("still rejects a repo whose meta-repo marker points elsewhere, without replanting it", func(ctx ginkgo.SpecContext) {
		address, recorder := startRecordingRegistry("test/repo")
		otherAddress, _ := startRecordingRegistry("test/meta")

		gomega.Expect(repoStagesStorageAt(ctx, address).PutMetaRepoMarker(ctx, "project", "registry.example/planted")).To(gomega.Succeed())
		recorder.forgetWrites()

		_, err := NewStorageManager(ctx, &NewStorageManagerConfig{
			ProjectName: "project",
			CmdData: storageManagerCmdData(
				"--repo", address,
				"--meta-repo", otherAddress,
				"--insecure-registry",
				"--synchronization", "https://synchronization.invalid",
				"--check-built-images",
			),
		})

		gomega.Expect(err).To(gomega.MatchError(gomega.ContainSubstring("registry.example/planted")))
		gomega.Expect(recorder.recordedWrites()).To(gomega.BeEmpty())
	})

	ginkgo.DescribeTable("registers the synchronization identity during ordinary initialization", func(ctx ginkgo.SpecContext, requireBuiltImages bool) {
		address, recorder := startRecordingRegistry("test/repo")
		syncServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			defer ginkgo.GinkgoRecover()
			_, err := w.Write([]byte(`{"clientID":"test-client"}`))
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
		}))
		ginkgo.DeferCleanup(syncServer.Close)

		cmdData := storageManagerCmdData("--repo", address, "--insecure-registry", "--synchronization", syncServer.URL)
		if requireBuiltImages {
			command := &cobra.Command{}
			SetupRequireBuiltImages(cmdData, command)
			gomega.Expect(command.ParseFlags([]string{"--require-built-images"})).To(gomega.Succeed())
		}

		storageManager, err := NewStorageManager(ctx, &NewStorageManagerConfig{
			ProjectName: "project",
			CmdData:     cmdData,
		})

		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		gomega.Expect(storageManager.StorageLockManager).NotTo(gomega.BeNil())
		gomega.Expect(recorder.recordedWrites()).NotTo(gomega.BeEmpty())
	},
		ginkgo.Entry("build command", false),
		ginkgo.Entry("require-built-images command", true),
	)
})
