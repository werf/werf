package common

import (
	"net/http"
	"net/http/httptest"
	"strings"

	ginkgo "github.com/onsi/ginkgo/v2"
	gomega "github.com/onsi/gomega"

	"github.com/werf/werf/v3/pkg/werf"
)

var _ = ginkgo.Describe("Meta repo autodetection", func() {
	ginkgo.BeforeEach(func() {
		clearListEnv()
		gomega.Expect(werf.Init(ginkgo.GinkgoT().TempDir(), ginkgo.GinkgoT().TempDir())).To(gomega.Succeed())
	})

	ginkgo.DescribeTable("resolves settings without overriding explicit addresses", func(ctx ginkgo.SpecContext, marker, configured string, fromEnv bool, expected, errorText string) {
		address, recorder := startRecordingRegistry("test/repo")
		metaAddress, metaRecorder := startRecordingRegistry("test/meta")
		addresses := map[string]string{"repo": address, "meta": metaAddress, "other": "registry.example/other"}
		if value, ok := addresses[marker]; ok {
			marker = value
		}
		if value, ok := addresses[configured]; ok {
			configured = value
		}
		if marker != "absent" {
			gomega.Expect(repoStagesStorageAt(ctx, address).PutMetaRepoMarker(ctx, "project", marker)).To(gomega.Succeed())
		}
		recorder.forgetWrites()
		args := []string{"--repo", address, "--insecure-registry", "--check-built-images"}
		if fromEnv {
			ginkgo.GinkgoT().Setenv("WERF_META_REPO", configured)
		} else {
			args = append(args, "--meta-repo", configured)
		}
		cmdData := storageManagerCmdData(args...)
		storageManager, err := NewStorageManager(ctx, &NewStorageManagerConfig{
			ProjectName: "project",
			CmdData:     cmdData,
		})
		if errorText != "" {
			gomega.Expect(err).To(gomega.MatchError(gomega.ContainSubstring(errorText)))
		} else {
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
			gomega.Expect(storageManager.GetMetaStorage().Address()).To(gomega.Equal(addresses[expected]))
		}
		gomega.Expect(*cmdData.MetaRepo.Address).To(gomega.Equal(configured))
		gomega.Expect(recorder.recordedWrites()).To(gomega.BeEmpty())
		gomega.Expect(metaRecorder.recordedWrites()).To(gomega.BeEmpty())
	},
		ginkgo.Entry("no marker or setting", "absent", "", false, "repo", ""),
		ginkgo.Entry("empty flag discovers marker", "meta", "", false, "meta", ""),
		ginkgo.Entry("empty environment discovers marker", "meta", "", true, "meta", ""),
		ginkgo.Entry("explicit flag without marker", "absent", "meta", false, "meta", ""),
		ginkgo.Entry("explicit environment without marker", "absent", "meta", true, "meta", ""),
		ginkgo.Entry("matching flag", "meta", "meta", false, "meta", ""),
		ginkgo.Entry("matching environment", "meta", "meta", true, "meta", ""),
		ginkgo.Entry("conflicting flag", "meta", "other", false, "", "they must match"),
		ginkgo.Entry("conflicting environment", "meta", "other", true, "", "they must match"),
		ginkgo.Entry("explicit primary without marker", "absent", "repo", false, "repo", ""),
		ginkgo.Entry("explicit primary conflicts with marker", "meta", "repo", false, "", "resolves to the same repository"),
		ginkgo.Entry("primary from environment conflicts with marker", "meta", "repo", true, "", "resolves to the same repository"),
		ginkgo.Entry("missing marker address", "", "", false, "", "malformed"),
		ginkgo.Entry("invalid marker address", "not a valid repository", "", false, "", "parse meta-repo marker"),
		ginkgo.Entry("local marker address", ":local", "", false, "", "parse meta-repo marker"),
		ginkgo.Entry("self-referencing marker", "repo", "", false, "", "points to the same repository"),
	)

	ginkgo.It("uses the marker for the requested project only", func(ctx ginkgo.SpecContext) {
		address, recorder := startRecordingRegistry("test/repo")
		gomega.Expect(repoStagesStorageAt(ctx, address).PutMetaRepoMarker(ctx, "other-project", "registry.example/meta")).To(gomega.Succeed())
		recorder.forgetWrites()
		storageManager, err := NewStorageManager(ctx, &NewStorageManagerConfig{
			ProjectName: "project",
			CmdData:     storageManagerCmdData("--repo", address, "--insecure-registry", "--check-built-images"),
		})
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		gomega.Expect(storageManager.GetMetaStorage()).To(gomega.BeIdenticalTo(storageManager.StagesStorage))
		gomega.Expect(recorder.recordedWrites()).To(gomega.BeEmpty())
	})

	ginkgo.It("lets an empty flag clear the environment and select autodetection", func(ctx ginkgo.SpecContext) {
		address, recorder := startRecordingRegistry("test/repo")
		metaAddress, metaRecorder := startRecordingRegistry("test/meta")
		gomega.Expect(repoStagesStorageAt(ctx, address).PutMetaRepoMarker(ctx, "project", metaAddress)).To(gomega.Succeed())
		recorder.forgetWrites()
		ginkgo.GinkgoT().Setenv("WERF_META_REPO", "registry.example/wrong")
		storageManager, err := NewStorageManager(ctx, &NewStorageManagerConfig{
			ProjectName: "project",
			CmdData:     storageManagerCmdData("--repo", address, "--meta-repo=", "--insecure-registry", "--check-built-images"),
		})
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		gomega.Expect(storageManager.GetMetaStorage().Address()).To(gomega.Equal(metaAddress))
		gomega.Expect(recorder.recordedWrites()).To(gomega.BeEmpty())
		gomega.Expect(metaRecorder.recordedWrites()).To(gomega.BeEmpty())
	})

	ginkgo.It("does not autodetect for marker maintenance commands", func(ctx ginkgo.SpecContext) {
		address, recorder := startRecordingRegistry("test/repo")
		gomega.Expect(repoStagesStorageAt(ctx, address).PutMetaRepoMarker(ctx, "project", "registry.example/meta")).To(gomega.Succeed())
		recorder.forgetWrites()
		storageManager, err := NewStorageManagerWithOptions(ctx, &NewStorageManagerConfig{
			ProjectName: "project",
			CmdData:     storageManagerCmdData("--repo", address, "--insecure-registry", "--check-built-images"),
		}, WithSkipMetaRepoSafeguard())
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		gomega.Expect(storageManager.GetMetaStorage()).To(gomega.BeIdenticalTo(storageManager.StagesStorage))
		gomega.Expect(recorder.recordedWrites()).To(gomega.BeEmpty())
	})

	ginkgo.It("propagates marker read errors instead of assuming no marker", func(ctx ginkgo.SpecContext) {
		server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			if request.URL.Path != "/v2/" {
				writer.WriteHeader(http.StatusForbidden)
			}
		}))
		ginkgo.DeferCleanup(server.Close)
		address := strings.TrimPrefix(server.URL, "http://") + "/test/repo"
		_, err := NewStorageManager(ctx, &NewStorageManagerConfig{
			ProjectName: "project",
			CmdData:     storageManagerCmdData("--repo", address, "--insecure-registry", "--check-built-images"),
		})
		gomega.Expect(err).To(gomega.MatchError(gomega.ContainSubstring("403")))
	})

	ginkgo.It("does not fall back when the detected meta repo denies access", func(ctx ginkgo.SpecContext) {
		server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			if request.URL.Path != "/v2/" {
				writer.WriteHeader(http.StatusForbidden)
			}
		}))
		ginkgo.DeferCleanup(server.Close)
		metaAddress := strings.TrimPrefix(server.URL, "http://") + "/test/meta"
		address, recorder := startRecordingRegistry("test/repo")
		gomega.Expect(repoStagesStorageAt(ctx, address).PutMetaRepoMarker(ctx, "project", metaAddress)).To(gomega.Succeed())
		recorder.forgetWrites()
		storageManager, err := NewStorageManager(ctx, &NewStorageManagerConfig{
			ProjectName: "project",
			CmdData:     storageManagerCmdData("--repo", address, "--insecure-registry", "--check-built-images"),
		})
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		gomega.Expect(storageManager.GetMetaStorage().Address()).To(gomega.Equal(metaAddress))
		_, err = storageManager.GetMetaStorage().GetManagedImages(ctx, "project")
		gomega.Expect(err).To(gomega.MatchError(gomega.ContainSubstring("403")))
		err = storageManager.GetMetaStorage().AddManagedImage(ctx, "project", "app")
		gomega.Expect(err).To(gomega.MatchError(gomega.ContainSubstring("403")))
		gomega.Expect(recorder.recordedWrites()).To(gomega.BeEmpty())
	})
})
