package manager

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/empty"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"go.uber.org/mock/gomock"

	"github.com/werf/logboek"
	"github.com/werf/logboek/pkg/level"
	"github.com/werf/werf/v3/pkg/build/stage"
	"github.com/werf/werf/v3/pkg/container_backend"
	"github.com/werf/werf/v3/pkg/docker_registry"
	"github.com/werf/werf/v3/pkg/image"
	"github.com/werf/werf/v3/pkg/storage"
	"github.com/werf/werf/v3/pkg/werf"
	"github.com/werf/werf/v3/test/mock"
)

var _ = ginkgo.DescribeTable("FetchStageImage with a locally available cache image",
	func(ctx ginkgo.SpecContext, withFinalRepo bool) {
		var output bytes.Buffer
		logger := logboek.NewLogger(&output, &output)
		logger.SetAcceptedLevel(level.Warn)
		logger.Streams().DisableStyle()
		logger.Streams().DisableLineWrapping()
		logCtx := logboek.NewContext(ctx, logger)
		gomega.Expect(werf.Init(ginkgo.GinkgoT().TempDir(), "")).To(gomega.Succeed())
		previousCache := image.CommonManifestCache
		image.CommonManifestCache = image.NewManifestCache(ginkgo.GinkgoT().TempDir())
		ginkgo.DeferCleanup(func() { image.CommonManifestCache = previousCache })

		cacheServer := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			if request.URL.Path == "/v2/" {
				writer.WriteHeader(http.StatusOK)
				return
			}
			http.Error(writer, "cache descriptor access denied", http.StatusForbidden)
		}))
		ginkgo.DeferCleanup(cacheServer.Close)
		cacheAddress := strings.TrimPrefix(cacheServer.URL, "http://") + "/cache"
		cacheRegistry, err := docker_registry.NewDockerRegistry(logCtx, cacheAddress, docker_registry.DefaultImplementationName, docker_registry.DockerRegistryOptions{InsecureRegistry: true})
		gomega.Expect(err).To(gomega.Succeed())

		backend := mock.NewMockContainerBackend(gomock.NewController(ginkgo.GinkgoT()))
		cache := storage.NewRepoStagesStorage(&storage.NewRepoStagesStorageOptions{
			RepoAddress: cacheAddress, DockerRegistry: cacheRegistry, ContainerBackend: backend,
		})
		manager := newManifestStorageManager(logCtx)
		manager.StagesStorage.(*storage.RepoStagesStorage).ContainerBackend = backend
		manager.CacheStagesStorageList = []storage.StagesStorage{cache}
		if withFinalRepo {
			manager.FinalStagesStorageList = []storage.StagesStorage{storage.NewRepoStagesStorage(&storage.NewRepoStagesStorageOptions{RepoAddress: "registry.example/final"})}
		}
		ginkgo.DeferCleanup(func() {
			for _, lock := range manager.SharedHostImagesLocks {
				gomega.Expect(werf.HostLocker().ReleaseLock(lock)).To(gomega.Succeed())
			}
		})

		stageID := image.NewStageID(missingStageDigest, 100)
		primaryName := manager.StagesStorage.ConstructStageImageName(manager.ProjectName, stageID.Digest, stageID.CreationTs)
		primaryRef, err := name.NewTag(primaryName, name.Insecure)
		gomega.Expect(err).To(gomega.Succeed())
		gomega.Expect(remote.Write(primaryRef, empty.Image, remote.WithContext(ctx))).To(gomega.Succeed())
		primaryImage := container_backend.NewLegacyStageImage(nil, primaryName, backend, "linux/amd64")
		primaryImage.SetStageDesc(&image.StageDesc{StageID: stageID})
		cacheName := cache.ConstructStageImageName(manager.ProjectName, stageID.Digest, stageID.CreationTs)
		gomock.InOrder(
			backend.EXPECT().GetImageInfo(gomock.Any(), primaryName, container_backend.GetImageInfoOpts{TargetPlatform: "linux/amd64"}).Return(nil, nil),
			backend.EXPECT().GetImageInfo(gomock.Any(), cacheName, container_backend.GetImageInfoOpts{TargetPlatform: "linux/amd64"}).Return(&image.Info{Name: cacheName}, nil),
			backend.EXPECT().PullImageFromRegistry(gomock.Any(), primaryImage).Return(nil),
		)

		var info FetchStageInfo
		gomega.Expect(func() {
			info, err = manager.FetchStageImage(logCtx, backend, "stage", &stage.StageImage{Image: primaryImage})
		}).NotTo(gomega.Panic())
		gomega.Expect(err).To(gomega.Succeed())
		gomega.Expect(info.BaseImageSource).To(gomega.Equal(BaseImageSourceTypeRepo))
		gomega.Expect(primaryImage.GetStageDesc().Info.Name).To(gomega.Equal(primaryName))
		gomega.Expect(strings.Split(output.String(), "\n")).To(gomega.ContainElement(gomega.And(
			gomega.HavePrefix(fmt.Sprintf(
				"Unable to fetch stage %s from cache stages storage %s: error getting stage %s description from %s:",
				stageID.String(), cache.String(), stageID.String(), cache.String(),
			)),
			gomega.ContainSubstring("cache descriptor access denied"),
		)))
		gomega.Expect(output.String()).NotTo(gomega.ContainSubstring("registry.example/final"))
	},
	ginkgo.Entry("warns and falls back after a descriptor error without final-repo", false),
	ginkgo.Entry("reports the cache address rather than final-repo after a descriptor error", true),
)
