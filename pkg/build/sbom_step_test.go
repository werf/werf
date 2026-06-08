package build

import (
	"archive/tar"
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/mock/gomock"

	"github.com/werf/common-go/pkg/util"
	"github.com/werf/werf/v2/pkg/container_backend"
	"github.com/werf/werf/v2/pkg/container_backend/filter"
	"github.com/werf/werf/v2/pkg/container_backend/label"
	"github.com/werf/werf/v2/pkg/image"
	"github.com/werf/werf/v2/pkg/logging"
	"github.com/werf/werf/v2/pkg/sbom/cyclonedxutil"
	"github.com/werf/werf/v2/pkg/sbom/gomod"
	sbomImage "github.com/werf/werf/v2/pkg/sbom/image"
	"github.com/werf/werf/v2/pkg/sbom/scanner"
	"github.com/werf/werf/v2/test/mock"
)

// createEmptyBOMStream creates a mock tar stream containing a minimal valid Docker image structure
// that contains an SBOM in the expected format.
func createEmptyBOMStream() io.ReadCloser {
	var buf bytes.Buffer
	tw := tar.NewWriter(&buf)

	layerContent := createLayerTarGz()
	layerDigest := sha256.Sum256(layerContent)
	layerDigestHex := hex.EncodeToString(layerDigest[:])

	configJSON := fmt.Sprintf(`{
		"architecture": "amd64",
		"os": "linux",
		"rootfs": {
			"type": "layers",
			"diff_ids": ["sha256:%s"]
		}
	}`, layerDigestHex)
	configDigest := sha256.Sum256([]byte(configJSON))
	configDigestHex := hex.EncodeToString(configDigest[:])
	configFileName := configDigestHex + ".json"
	layerFileName := layerDigestHex + "/layer.tar"
	manifest := []map[string]interface{}{
		{
			"Config":   configFileName,
			"RepoTags": []string{"test:latest"},
			"Layers":   []string{layerFileName},
		},
	}
	manifestJSON, _ := json.Marshal(manifest)

	writeFileToTar(tw, "manifest.json", manifestJSON)

	writeFileToTar(tw, configFileName, []byte(configJSON))

	tw.WriteHeader(&tar.Header{
		Name:     layerDigestHex + "/",
		Mode:     0o755,
		Typeflag: tar.TypeDir,
	})
	writeFileToTar(tw, layerFileName, layerContent)

	tw.Close()
	return io.NopCloser(&buf)
}

func createLayerTarGz() []byte {
	var buf bytes.Buffer
	tw := tar.NewWriter(&buf)

	scanOpts := scanner.DefaultSyftScanOptions()
	billName := scanner.BillNameFromCommand(scanOpts.Commands[0])

	bomJSON := []byte(`{"bomFormat":"CycloneDX","specVersion":"1.6","version":1,"components":[]}`)
	bomFilePath := filepath.Join("sbom", billName)
	writeFileToTar(tw, bomFilePath, bomJSON)

	tw.Close()
	return buf.Bytes()
}

func writeFileToTar(tw *tar.Writer, name string, content []byte) {
	tw.WriteHeader(&tar.Header{
		Name: name,
		Mode: 0o644,
		Size: int64(len(content)),
	})
	tw.Write(content)
}

var _ = Describe("SbomStep", func() {
	DescribeTable("ConvergeWithMerge()",
		func(
			ctx context.Context,
			isLocalStorage bool,
			mergeOpts cyclonedxutil.MergeOpts,
			setupGitRepo func(ctx context.Context, repo *mock.MockGitRepo, commit, imageContext string) []byte,
			setupMocks func(
				ctx context.Context,
				backend *mock.MockContainerBackend,
				stagesStorage *mock.MockStagesStorage,
				stageDesc *image.StageDesc,
				scanOpts scanner.ScanOptions,
				sbomImgLabels label.LabelList,
				imgFilters []util.Pair[string, string],
			),
		) {
			backend := mock.NewMockContainerBackend(gomock.NewController(GinkgoT()))
			stagesStorage := mock.NewMockStagesStorage(gomock.NewController(GinkgoT()))

			step := newSbomStep(backend, stagesStorage)
			step.isLocalStorage = isLocalStorage

			ctx = logging.WithLogger(ctx)
			var patchers []BOMPatcherInterface
			if setupGitRepo != nil {
				repo := mock.NewMockGitRepo(gomock.NewController(GinkgoT()))
				commit := "0123456789abcdef0123456789abcdef01234567"
				imageContext := "app"
				setupGitRepo(ctx, repo, commit, imageContext)
				patchers = []BOMPatcherInterface{gomod.NewBOMPatcher(repo, commit, imageContext)}
			}
			stageDesc := &image.StageDesc{
				Info: &image.Info{
					Name: "docker.io/namespace/repo:e5c6ebcd2718ccfe74d01069a0d758e03d5a2554155ccdc01be0daff-1747987463184",
					Labels: map[string]string{
						image.WerfLabel:                   "werf-test-app",
						image.WerfProjectRepoCommitLabel:  "fdf911003d06d63460172a0e663bf492cb9b8160",
						image.WerfStageContentDigestLabel: "7072c9c90d5b76981e4ef55ab07c2d935039b3eb73570354aaf4c2da",
						image.WerfVersionLabel:            "dev",
					},
				},
			}
			scanOpts := scanner.DefaultSyftScanOptions()

			sbomBaseImgLabels := step.prepareSbomBaseLabelsWithMerge(ctx, stageDesc.Info.Labels, scanOpts, mergeOpts)
			imgFilters := filter.NewFilterListFromLabelList(sbomBaseImgLabels).ToPairs()

			sbomImgLabels := step.prepareSbomLabelsWithMerge(ctx, stageDesc.Info.Labels, scanOpts, mergeOpts)
			setupMocks(ctx, backend, stagesStorage, stageDesc, scanOpts, sbomImgLabels, imgFilters)

			Expect(step.ConvergeWithMerge(ctx, "some-name", stageDesc, scanOpts, mergeOpts, patchers)).To(Succeed())
		},
		Entry(
			"[local storage]: should not scan source image if sbom image already exists",
			true,
			cyclonedxutil.MergeOpts{},
			nil,
			func(
				ctx context.Context,
				backend *mock.MockContainerBackend,
				stagesStorage *mock.MockStagesStorage,
				stageDesc *image.StageDesc,
				scanOpts scanner.ScanOptions,
				sbomImgLabels label.LabelList,
				imgFilters []util.Pair[string, string],
			) {
				backend.EXPECT().Images(ctx, container_backend.ImagesOptions{Filters: imgFilters}).Return(image.ImagesList{
					{RepoTags: []string{"namespace/repo:e5c6ebcd2718ccfe74d01069a0d758e03d5a2554155ccdc01be0daff-1747987463184-sbom"}},
				}, nil)
			},
		),
		Entry(
			"[local storage]: should scan source image if sbom image does not exist",
			true,
			cyclonedxutil.MergeOpts{},
			nil,
			func(
				ctx context.Context,
				backend *mock.MockContainerBackend,
				stagesStorage *mock.MockStagesStorage,
				stageDesc *image.StageDesc,
				scanOpts scanner.ScanOptions,
				sbomImgLabels label.LabelList,
				imgFilters []util.Pair[string, string],
			) {
				backend.EXPECT().Images(ctx, container_backend.ImagesOptions{Filters: imgFilters}).Return(image.ImagesList{}, nil)

				tmpImgId := "tmp-sbom-img-id"
				backend.EXPECT().GenerateSBOM(ctx, scanOpts, gomock.Any()).Return(tmpImgId, nil)
				// Return a new stream each time SaveImageToStream is called
				backend.EXPECT().SaveImageToStream(ctx, tmpImgId).DoAndReturn(func(_ context.Context, _ string) (io.ReadCloser, error) {
					return createEmptyBOMStream(), nil
				}).AnyTimes()
				backend.EXPECT().Rmi(ctx, tmpImgId, container_backend.RmiOpts{Force: true}).Return(nil)
				backend.EXPECT().BuildDockerfile(ctx, gomock.Any(), gomock.Any()).Return("final-sbom-img-id", nil)
				backend.EXPECT().Tag(ctx, "final-sbom-img-id", sbomImage.ImageName(stageDesc.Info.Name), container_backend.TagOpts{}).Return(nil)
			},
		),
		Entry(
			"[remote storage]: should push sbom image if it exists locally",
			false,
			cyclonedxutil.MergeOpts{},
			nil,
			func(
				ctx context.Context,
				backend *mock.MockContainerBackend,
				stagesStorage *mock.MockStagesStorage,
				stageDesc *image.StageDesc,
				scanOpts scanner.ScanOptions,
				sbomImgLabels label.LabelList,
				imgFilters []util.Pair[string, string],
			) {
				backend.EXPECT().Images(ctx, container_backend.ImagesOptions{Filters: imgFilters}).Return(image.ImagesList{
					{RepoTags: []string{"namespace/repo:e5c6ebcd2718ccfe74d01069a0d758e03d5a2554155ccdc01be0daff-1747987463184-sbom"}},
				}, nil)
				stagesStorage.EXPECT().PushIfNotExistSbomImage(ctx, sbomImage.ImageName(stageDesc.Info.Name)).Return(true, nil)
			},
		),
		Entry(
			"[remote storage]: should not scan if sbom image is pulled from registry",
			false,
			cyclonedxutil.MergeOpts{},
			nil,
			func(
				ctx context.Context,
				backend *mock.MockContainerBackend,
				stagesStorage *mock.MockStagesStorage,
				stageDesc *image.StageDesc,
				scanOpts scanner.ScanOptions,
				sbomImgLabels label.LabelList,
				imgFilters []util.Pair[string, string],
			) {
				backend.EXPECT().Images(ctx, container_backend.ImagesOptions{Filters: imgFilters}).Return(image.ImagesList{}, nil)
				stagesStorage.EXPECT().PullIfExistSbomImage(ctx, sbomImage.ImageName(stageDesc.Info.Name)).Return(true, nil)
			},
		),
		Entry(
			"[remote storage]: should scan, build and push sbom image if not found",
			false,
			cyclonedxutil.MergeOpts{},
			nil,
			func(
				ctx context.Context,
				backend *mock.MockContainerBackend,
				stagesStorage *mock.MockStagesStorage,
				stageDesc *image.StageDesc,
				scanOpts scanner.ScanOptions,
				sbomImgLabels label.LabelList,
				imgFilters []util.Pair[string, string],
			) {
				backend.EXPECT().Images(ctx, container_backend.ImagesOptions{Filters: imgFilters}).Return(image.ImagesList{}, nil)
				stagesStorage.EXPECT().PullIfExistSbomImage(ctx, sbomImage.ImageName(stageDesc.Info.Name)).Return(false, nil)

				backend.EXPECT().Pull(ctx, stageDesc.Info.Name, container_backend.PullOpts{}).Return(nil)

				tmpImgId := "tmp-sbom-img-id"
				backend.EXPECT().GenerateSBOM(ctx, scanOpts, gomock.Any()).Return(tmpImgId, nil)
				backend.EXPECT().SaveImageToStream(ctx, tmpImgId).DoAndReturn(func(_ context.Context, _ string) (io.ReadCloser, error) {
					return createEmptyBOMStream(), nil
				}).AnyTimes()
				backend.EXPECT().Rmi(ctx, tmpImgId, container_backend.RmiOpts{Force: true}).Return(nil)
				backend.EXPECT().BuildDockerfile(ctx, gomock.Any(), gomock.Any()).Return("final-sbom-img-id", nil)
				backend.EXPECT().Tag(ctx, "final-sbom-img-id", sbomImage.ImageName(stageDesc.Info.Name), container_backend.TagOpts{}).Return(nil)

				stagesStorage.EXPECT().PushIfNotExistSbomImage(ctx, sbomImage.ImageName(stageDesc.Info.Name)).Return(true, nil)
			},
		),
		Entry(
			"[go.mod]: should skip version resolution when go.mod is missing",
			true,
			cyclonedxutil.MergeOpts{},
			func(ctx context.Context, repo *mock.MockGitRepo, commit, imageContext string) []byte {
				repo.EXPECT().IsCommitFileExist(ctx, commit, filepath.Join(imageContext, "go.mod")).Return(false, nil)

				return nil
			},
			func(
				ctx context.Context,
				backend *mock.MockContainerBackend,
				stagesStorage *mock.MockStagesStorage,
				stageDesc *image.StageDesc,
				scanOpts scanner.ScanOptions,
				sbomImgLabels label.LabelList,
				imgFilters []util.Pair[string, string],
			) {
				backend.EXPECT().Images(ctx, container_backend.ImagesOptions{Filters: imgFilters}).Return(image.ImagesList{}, nil)

				tmpImgId := "tmp-sbom-img-id"
				backend.EXPECT().GenerateSBOM(ctx, scanOpts, gomock.Any()).Return(tmpImgId, nil)
				backend.EXPECT().SaveImageToStream(ctx, tmpImgId).DoAndReturn(func(_ context.Context, _ string) (io.ReadCloser, error) {
					return createEmptyBOMStream(), nil
				}).AnyTimes()
				backend.EXPECT().Rmi(ctx, tmpImgId, container_backend.RmiOpts{Force: true}).Return(nil)
				backend.EXPECT().BuildDockerfile(ctx, gomock.Any(), gomock.Any()).Return("final-sbom-img-id", nil)
				backend.EXPECT().Tag(ctx, "final-sbom-img-id", sbomImage.ImageName(stageDesc.Info.Name), container_backend.TagOpts{}).Return(nil)
			},
		),
		Entry(
			"[go.mod]: should use tag version when tag matches commit",
			true,
			cyclonedxutil.MergeOpts{},
			func(ctx context.Context, repo *mock.MockGitRepo, commit, imageContext string) []byte {
				goModPath := filepath.Join(imageContext, "go.mod")
				goModContent := []byte("module example.com/app\n")
				repo.EXPECT().IsCommitFileExist(ctx, commit, goModPath).Return(true, nil)
				repo.EXPECT().ReadCommitFile(ctx, commit, goModPath).Return(goModContent, nil)
				repo.EXPECT().TagsList(ctx).Return([]string{"v1.2.3"}, nil)
				repo.EXPECT().TagCommit(ctx, "v1.2.3").Return(commit, nil)

				return goModContent
			},
			func(
				ctx context.Context,
				backend *mock.MockContainerBackend,
				stagesStorage *mock.MockStagesStorage,
				stageDesc *image.StageDesc,
				scanOpts scanner.ScanOptions,
				sbomImgLabels label.LabelList,
				imgFilters []util.Pair[string, string],
			) {
				backend.EXPECT().Images(ctx, container_backend.ImagesOptions{Filters: imgFilters}).Return(image.ImagesList{}, nil)

				tmpImgId := "tmp-sbom-img-id"
				backend.EXPECT().GenerateSBOM(ctx, scanOpts, gomock.Any()).Return(tmpImgId, nil)
				backend.EXPECT().SaveImageToStream(ctx, tmpImgId).DoAndReturn(func(_ context.Context, _ string) (io.ReadCloser, error) {
					return createEmptyBOMStream(), nil
				}).AnyTimes()
				backend.EXPECT().Rmi(ctx, tmpImgId, container_backend.RmiOpts{Force: true}).Return(nil)
				backend.EXPECT().BuildDockerfile(ctx, gomock.Any(), gomock.Any()).Return("final-sbom-img-id", nil)
				backend.EXPECT().Tag(ctx, "final-sbom-img-id", sbomImage.ImageName(stageDesc.Info.Name), container_backend.TagOpts{}).Return(nil)
			},
		),
		Entry(
			"[go.mod]: should fallback to pseudo version when tag mismatch",
			true,
			cyclonedxutil.MergeOpts{},
			func(ctx context.Context, repo *mock.MockGitRepo, commit, imageContext string) []byte {
				goModPath := filepath.Join(imageContext, "go.mod")
				goModContent := []byte("module example.com/app\n")
				repo.EXPECT().IsCommitFileExist(ctx, commit, goModPath).Return(true, nil)
				repo.EXPECT().ReadCommitFile(ctx, commit, goModPath).Return(goModContent, nil)
				repo.EXPECT().TagsList(ctx).Return([]string{"v1.2.3"}, nil)
				repo.EXPECT().TagCommit(ctx, "v1.2.3").Return("deadbeefdeadbeefdeadbeefdeadbeefdeadbeef", nil)

				return goModContent
			},
			func(ctx context.Context, backend *mock.MockContainerBackend, stagesStorage *mock.MockStagesStorage, stageDesc *image.StageDesc, scanOpts scanner.ScanOptions, sbomImgLabels label.LabelList, imgFilters []util.Pair[string, string]) {
				backend.EXPECT().Images(ctx, container_backend.ImagesOptions{Filters: imgFilters}).Return(image.ImagesList{}, nil)

				tmpImgId := "tmp-sbom-img-id"
				backend.EXPECT().GenerateSBOM(ctx, scanOpts, gomock.Any()).Return(tmpImgId, nil)
				backend.EXPECT().SaveImageToStream(ctx, tmpImgId).DoAndReturn(func(_ context.Context, _ string) (io.ReadCloser, error) {
					return createEmptyBOMStream(), nil
				}).AnyTimes()
				backend.EXPECT().Rmi(ctx, tmpImgId, container_backend.RmiOpts{Force: true}).Return(nil)
				backend.EXPECT().BuildDockerfile(ctx, gomock.Any(), gomock.Any()).Return("final-sbom-img-id", nil)
				backend.EXPECT().Tag(ctx, "final-sbom-img-id", sbomImage.ImageName(stageDesc.Info.Name), container_backend.TagOpts{}).Return(nil)
			},
		),
	)

	DescribeTable("pullImageSbom()",
		func(ctx context.Context, baseImageInfo *image.Info, expectError bool, expectedErrorMsg string, setupMocks func(ctx context.Context, backend *mock.MockContainerBackend, baseImageInfo *image.Info),
		) {
			backend := mock.NewMockContainerBackend(gomock.NewController(GinkgoT()))
			stagesStorage := mock.NewMockStagesStorage(gomock.NewController(GinkgoT()))

			step := newSbomStep(backend, stagesStorage)

			ctx = logging.WithLogger(ctx)

			setupMocks(ctx, backend, baseImageInfo)

			_, err := step.pullImageSbom(ctx, "some-name", baseImageInfo)
			if expectError {
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring(expectedErrorMsg))
			} else {
				Expect(err).ToNot(HaveOccurred())
			}
		},
		Entry(
			"should fail if sbom not found in registry",
			&image.Info{
				Name:       "docker.io/namespace/repo:e5c6ebcd2718ccfe74d01069a0d758e03d5a2554155ccdc01be0daff-1747987463184",
				Repository: "docker.io/namespace/repo",
				Labels: map[string]string{
					image.WerfStageContentDigestLabel: "e5c6ebcd2718ccfe74d01069a0d758e03d5a2554155ccdc01be0daff",
				},
			},
			true,
			"SBOM for image",
			func(
				ctx context.Context,
				backend *mock.MockContainerBackend,
				baseImageInfo *image.Info,
			) {
				_, tag := image.ParseRepositoryAndTag(baseImageInfo.Name)
				sbomImageName := sbomImage.BaseImageName(baseImageInfo.Repository, tag)
				backend.EXPECT().GetImageInfo(ctx, sbomImageName, container_backend.GetImageInfoOpts{}).Return(nil, nil)
				backend.EXPECT().Pull(ctx, sbomImageName, container_backend.PullOpts{}).Return(fmt.Errorf("not found"))
			},
		),
		Entry(
			"should fail if no werf labels available",
			&image.Info{
				Name:       "ubuntu:22.04",
				Repository: "ubuntu",
				Labels:     map[string]string{},
			},
			true,
			"required werf stage content digest label is missing",
			func(
				ctx context.Context,
				backend *mock.MockContainerBackend,
				baseImageInfo *image.Info,
			) {
				// No mocks needed - should fail before any backend calls
			},
		),
	)

	DescribeTable("pullImageSbom() with local storage",
		func(
			ctx context.Context,
			baseImageInfo *image.Info,
			expectError bool,
			expectedErrorMsg string,
			setupMocks func(
				ctx context.Context,
				backend *mock.MockContainerBackend,
				baseImageInfo *image.Info,
			),
		) {
			backend := mock.NewMockContainerBackend(gomock.NewController(GinkgoT()))
			stagesStorage := mock.NewMockStagesStorage(gomock.NewController(GinkgoT()))

			step := &sbomStep{
				containerBackend: backend,
				stagesStorage:    stagesStorage,
				isLocalStorage:   true,
			}

			ctx = logging.WithLogger(ctx)

			setupMocks(ctx, backend, baseImageInfo)

			_, err := step.pullImageSbom(ctx, "some-name", baseImageInfo)
			if expectError {
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring(expectedErrorMsg))
			} else {
				Expect(err).ToNot(HaveOccurred())
			}
		},
		Entry(
			"should fail if sbom not found locally (no pull for local storage)",
			&image.Info{
				Name:       "docker.io/namespace/repo:e5c6ebcd2718ccfe74d01069a0d758e03d5a2554155ccdc01be0daff-1747987463184",
				Repository: "docker.io/namespace/repo",
				Labels: map[string]string{
					image.WerfStageContentDigestLabel: "e5c6ebcd2718ccfe74d01069a0d758e03d5a2554155ccdc01be0daff",
				},
			},
			true,
			"not found locally",
			func(
				ctx context.Context,
				backend *mock.MockContainerBackend,
				baseImageInfo *image.Info,
			) {
				_, tag := image.ParseRepositoryAndTag(baseImageInfo.Name)
				sbomImageName := sbomImage.BaseImageName(baseImageInfo.Repository, tag)
				backend.EXPECT().GetImageInfo(ctx, sbomImageName, container_backend.GetImageInfoOpts{}).Return(nil, nil)
			},
		),
	)
})
