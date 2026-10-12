package common

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/samber/lo"
	"github.com/spf13/cobra"

	"github.com/werf/logboek"
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
		ginkgo.Entry("local with ignored invalid metadata", ":local", "invalid://metadata address"),
		ginkgo.Entry("implicit local with ignored invalid metadata", "", "invalid://metadata address"),
		ginkgo.Entry("local metadata with local repo", ":local", ":local"),
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
		ginkgo.Entry("implicit local repo", "", "example.com/metadata", false, ""),
		ginkgo.Entry("ignores metadata with local repo", ":local", "example.com/metadata", false, ""),
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

var _ = ginkgo.DescribeTable("cleanup report metadata repository",
	func(selection string) {
		ctx := context.Background()
		gomega.Expect(werf.Init(ginkgo.GinkgoT().TempDir(), ginkgo.GinkgoT().TempDir())).To(gomega.Succeed())
		repo := "example.com/project"
		meta := ""
		switch selection {
		case "same":
			meta = repo
		case "equivalent":
			meta = repo + ":ignored"
		case "separate":
			meta = "example.com/metadata"
		}
		data := storageManagerCmdData("--repo", repo, "--meta-repo", meta, "--check-built-images")
		data.SaveCleanupReport = lo.ToPtr(true)
		data.CleanupReportPath = lo.ToPtr(filepath.Join(ginkgo.GinkgoT().TempDir(), "report.json"))
		manager, err := NewStorageManager(ctx, &NewStorageManagerConfig{ProjectName: "project", ContainerBackend: &container_backend.DockerServerBackend{}, CmdData: data, SkipMetaRepoSafeguard: true})
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		report, path, err := NewCleanupReport(ctx, data, "cleanup", false, manager)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		gomega.Expect(report.Save(ctx, path)).To(gomega.Succeed())
		raw, err := os.ReadFile(path)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		var fields map[string]any
		gomega.Expect(json.Unmarshal(raw, &fields)).To(gomega.Succeed())
		gomega.Expect(fields).To(gomega.HaveKeyWithValue("repo", repo))
		if selection == "separate" {
			gomega.Expect(fields).To(gomega.HaveKeyWithValue("metaRepo", meta))
		} else {
			gomega.Expect(fields).NotTo(gomega.HaveKey("metaRepo"))
		}
	},
	ginkgo.Entry("unset", "unset"), ginkgo.Entry("same", "same"), ginkgo.Entry("equivalent", "equivalent"), ginkgo.Entry("separate", "separate"),
)

var _ = ginkgo.DescribeTable("optional repository flags not registered by metadata commands", func(final *RepoData) {
	backend := &container_backend.DockerServerBackend{}
	result, err := GetOptionalFinalStagesStorage(context.Background(), backend, &CmdData{FinalRepo: final})
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	gomega.Expect(result).To(gomega.BeNil())
}, ginkgo.Entry("no flag group", (*RepoData)(nil)), ginkgo.Entry("no address flag", &RepoData{}), ginkgo.Entry("empty address", &RepoData{Address: lo.ToPtr("")}))

var _ = ginkgo.DescribeTable("optional cache and secondary repository flags", func(registered bool) {
	clearListEnv()
	ginkgo.GinkgoT().Setenv("WERF_CACHE_REPO", "example.com/bare-cache")
	ginkgo.GinkgoT().Setenv("WERF_CACHE_REPO_TEST", "example.com/cache")
	ginkgo.GinkgoT().Setenv("WERF_SECONDARY_REPO", "example.com/bare-secondary")
	ginkgo.GinkgoT().Setenv("WERF_SECONDARY_REPO_TEST", "example.com/secondary")
	data := &CmdData{}
	if registered {
		data.CacheStagesStorage = lo.ToPtr([]string{"example.com/flag-cache"})
		data.SecondaryStagesStorage = lo.ToPtr([]string{"example.com/flag-secondary"})
	}
	expectedCache := []string{"example.com/bare-cache", "example.com/cache"}
	expectedSecondary := []string{"example.com/bare-secondary", "example.com/secondary"}
	if registered {
		expectedCache = append(expectedCache, "example.com/flag-cache")
		expectedSecondary = append(expectedSecondary, "example.com/flag-secondary")
	}
	gomega.Expect(GetCacheStagesStorage(data)).To(gomega.Equal(expectedCache))
	gomega.Expect(GetSecondaryStagesStorage(data)).To(gomega.Equal(expectedSecondary))
}, ginkgo.Entry("unregistered", false), ginkgo.Entry("registered", true))

var _ = ginkgo.DescribeTable("deferred metadata warning", func(selection string, warning bool) {
	var output bytes.Buffer
	ctx := logboek.NewContext(context.Background(), logboek.NewLogger(&output, &output))
	gomega.Expect(werf.Init(ginkgo.GinkgoT().TempDir(), ginkgo.GinkgoT().TempDir())).To(gomega.Succeed())
	repo, registry := startRecordingRegistry("test/stages")
	handler := registry.handler
	registry.handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer ginkgo.GinkgoRecover()
		if strings.HasSuffix(r.URL.Path, "/tags/list") {
			tags := make([]string, 100)
			for i := range tags {
				tags[i] = fmt.Sprintf("stage-%d", i)
			}
			w.Header().Set("Content-Type", "application/json")
			gomega.Expect(json.NewEncoder(w).Encode(map[string]any{"name": "test/stages", "tags": tags})).To(gomega.Succeed())
			return
		}
		handler.ServeHTTP(w, r)
	})
	meta := ""
	switch selection {
	case "equivalent":
		meta = repo + ":ignored"
	case "separate":
		meta = "example.com/meta"
	}
	data := storageManagerCmdData("--repo", repo, "--meta-repo", meta, "--insecure-registry", "--check-built-images")
	manager, err := NewStorageManager(ctx, &NewStorageManagerConfig{ProjectName: "project", ContainerBackend: &container_backend.DockerServerBackend{}, CmdData: data, SkipMetaRepoSafeguard: true})
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	_, err = manager.GetStagesStorage().GetStagesIDs(ctx, "project")
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	gomega.Expect(strings.Contains(output.String(), "has not been run for a long time")).To(gomega.Equal(warning))
}, ginkgo.Entry("default retains warning", "unset", true), ginkgo.Entry("equivalent retains warning", "equivalent", true), ginkgo.Entry("separate suppresses primary warning", "separate", false))

var _ = ginkgo.DescribeTable("local storage ignores metadata repository configuration", func(fromEnv, invalid, explicitLocal bool) {
	ctx := context.Background()
	gomega.Expect(werf.Init(ginkgo.GinkgoT().TempDir(), ginkgo.GinkgoT().TempDir())).To(gomega.Succeed())
	meta, recorder := startRecordingRegistry("test/ignored-meta")
	var requests atomic.Int64
	handler := recorder.handler
	recorder.handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests.Add(1)
		handler.ServeHTTP(w, r)
	})
	if invalid {
		meta += "/invalid address"
	}
	ginkgo.GinkgoT().Setenv("WERF_REPO", "")
	args := []string{"--insecure-registry"}
	if explicitLocal {
		args = append(args, "--repo", ":local")
	}
	if fromEnv {
		ginkgo.GinkgoT().Setenv("WERF_META_REPO", meta)
	} else {
		args = append(args, "--meta-repo", meta)
	}
	data := storageManagerCmdData(args...)
	manager, err := NewStorageManager(ctx, &NewStorageManagerConfig{ProjectName: "project", ContainerBackend: &container_backend.DockerServerBackend{}, CmdData: data})
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	gomega.Expect(manager.GetStagesStorage()).To(gomega.BeAssignableToTypeOf(&storage.LocalStagesStorage{}))
	gomega.Expect(manager.GetMetaStorage(ctx)).To(gomega.BeIdenticalTo(manager.GetStagesStorage()))
	gomega.Expect(requests.Load()).To(gomega.BeZero())
}, ginkgo.Entry("CLI", false, false, true), ginkgo.Entry("CLI invalid address", false, true, true), ginkgo.Entry("environment", true, false, true), ginkgo.Entry("environment invalid address", true, true, true),
	ginkgo.Entry("implicit local CLI", false, false, false), ginkgo.Entry("implicit local CLI invalid address", false, true, false), ginkgo.Entry("implicit local environment", true, false, false), ginkgo.Entry("implicit local environment invalid address", true, true, false))
