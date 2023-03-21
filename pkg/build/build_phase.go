package build

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/moby/buildkit/frontend/dockerfile/instructions"

	"github.com/werf/logboek"
	"github.com/werf/logboek/pkg/style"
	"github.com/werf/logboek/pkg/types"
	"github.com/werf/werf/pkg/build/image"
	"github.com/werf/werf/pkg/build/stage"
	"github.com/werf/werf/pkg/container_backend"
	backend_instruction "github.com/werf/werf/pkg/container_backend/instruction"
	"github.com/werf/werf/pkg/docker_registry"
	"github.com/werf/werf/pkg/git_repo"
	imagePkg "github.com/werf/werf/pkg/image"
	"github.com/werf/werf/pkg/stapel"
	"github.com/werf/werf/pkg/storage"
	"github.com/werf/werf/pkg/storage/manager"
	"github.com/werf/werf/pkg/util"
	"github.com/werf/werf/pkg/werf"
)

type BuildPhaseOptions struct {
	BuildOptions
	ShouldBeBuiltMode bool
}

type BuildOptions struct {
	ImageBuildOptions container_backend.BuildOptions
	IntrospectOptions

	ReportPath   string
	ReportFormat ReportFormat

	SkipImageMetadataPublication bool
	CustomTagFuncList            []imagePkg.CustomTagFunc
}

type IntrospectOptions struct {
	Targets []IntrospectTarget
}

type IntrospectTarget struct {
	ImageName string
	StageName string
}

func (opts *IntrospectOptions) ImageStageShouldBeIntrospected(imageName, stageName string) bool {
	for _, s := range opts.Targets {
		if (s.ImageName == "*" || s.ImageName == imageName) && s.StageName == stageName {
			return true
		}
	}

	return false
}

func NewBuildPhase(c *Conveyor, opts BuildPhaseOptions) *BuildPhase {
	return &BuildPhase{
		BasePhase:         BasePhase{c},
		BuildPhaseOptions: opts,
		ImagesReport:      &ImagesReport{Images: make(map[string]ReportImageRecord)},
	}
}

type BuildPhase struct {
	BasePhase
	BuildPhaseOptions

	StagesIterator              *StagesIterator
	ShouldAddManagedImageRecord bool

	ImagesReport *ImagesReport

	buildContextArchive container_backend.BuildContextArchiver
}

const (
	ReportJSON    ReportFormat = "json"
	ReportEnvFile ReportFormat = "envfile"
)

type ReportFormat string

type ImagesReport struct {
	mux    sync.Mutex
	Images map[string]ReportImageRecord
}

func (report *ImagesReport) SetImageRecord(name string, imageRecord ReportImageRecord) {
	report.mux.Lock()
	defer report.mux.Unlock()
	report.Images[name] = imageRecord
}

func (report *ImagesReport) ToJsonData() ([]byte, error) {
	report.mux.Lock()
	defer report.mux.Unlock()

	data, err := json.MarshalIndent(report, "", "\t")
	if err != nil {
		return nil, err
	}
	data = append(data, []byte("\n")...)

	return data, nil
}

func (report *ImagesReport) ToEnvFileData() []byte {
	report.mux.Lock()
	defer report.mux.Unlock()

	buf := bytes.NewBuffer([]byte{})
	for img, record := range report.Images {
		buf.WriteString(generateImageEnv(img, record.DockerImageName))
		buf.WriteString("\n")
	}

	return buf.Bytes()
}

func generateImageEnv(werfImageName, imageName string) string {
	var imageEnvName string
	if werfImageName == "" {
		imageEnvName = "WERF_DOCKER_IMAGE_NAME"
	} else {
		werfImageName := strings.ToUpper(werfImageName)
		for _, l := range []string{"/", "-"} {
			werfImageName = strings.ReplaceAll(werfImageName, l, "_")
		}

		imageEnvName = fmt.Sprintf("WERF_%s_DOCKER_IMAGE_NAME", werfImageName)
	}

	return fmt.Sprintf("%s=%s", imageEnvName, imageName)
}

type ReportImageRecord struct {
	WerfImageName     string
	DockerRepo        string
	DockerTag         string
	DockerImageID     string
	DockerImageDigest string
	DockerImageName   string
	Rebuilt           bool
}

func (phase *BuildPhase) Name() string {
	return "build"
}

func (phase *BuildPhase) BeforeImages(ctx context.Context) error {
	if err := phase.Conveyor.StorageManager.InitCache(ctx); err != nil {
		return fmt.Errorf("unable to init storage manager cache: %w", err)
	}
	return nil
}

func (phase *BuildPhase) AfterImages(ctx context.Context) error {
	return phase.createReport(ctx)
}

func (phase *BuildPhase) createReport(ctx context.Context) error {
	for _, img := range phase.Conveyor.imagesTree.GetImages() {
		if img.IsArtifact {
			continue
		}

		stageImage := img.GetLastNonEmptyStage().GetStageImage().Image
		desc := stageImage.GetFinalStageDescription()
		if desc == nil {
			desc = stageImage.GetStageDescription()
		}

		phase.ImagesReport.SetImageRecord(img.GetName(), ReportImageRecord{
			WerfImageName:     img.GetName(),
			DockerRepo:        desc.Info.Repository,
			DockerTag:         desc.Info.Tag,
			DockerImageID:     desc.Info.ID,
			DockerImageDigest: desc.Info.RepoDigest,
			DockerImageName:   desc.Info.Name,
			Rebuilt:           img.GetRebuilt(),
		})
	}

	debugJsonData, err := phase.ImagesReport.ToJsonData()
	logboek.Context(ctx).Debug().LogF("ImagesReport: (err: %v)\n%s", err, debugJsonData)

	if phase.ReportPath != "" {
		var data []byte
		var err error
		switch phase.ReportFormat {
		case ReportJSON:
			if data, err = phase.ImagesReport.ToJsonData(); err != nil {
				return fmt.Errorf("unable to prepare report json: %w", err)
			}
			logboek.Context(ctx).Debug().LogF("Writing json report to the %q:\n%s", phase.ReportPath, data)
		case ReportEnvFile:
			data = phase.ImagesReport.ToEnvFileData()
			logboek.Context(ctx).Debug().LogF("Writing envfile report to the %q:\n%s", phase.ReportPath, data)
		default:
			panic(fmt.Sprintf("unknown report format %q", phase.ReportFormat))
		}

		if err := ioutil.WriteFile(phase.ReportPath, data, 0o644); err != nil {
			return fmt.Errorf("unable to write report to %s: %w", phase.ReportPath, err)
		}
	}

	return nil
}

func (phase *BuildPhase) ImageProcessingShouldBeStopped(_ context.Context, _ *image.Image) bool {
	return false
}

func (phase *BuildPhase) BeforeImageStages(ctx context.Context, img *image.Image) (deferFn func(), err error) {
	phase.StagesIterator = NewStagesIterator(phase.Conveyor)

	if err := img.SetupBaseImage(ctx, phase.Conveyor.StorageManager, manager.StorageOptions{
		ContainerBackend: phase.Conveyor.ContainerBackend,
		DockerRegistry:   docker_registry.API(),
	}); err != nil {
		return nil, fmt.Errorf("unable to setup base image: %w", err)
	}

	if img.UsesBuildContext() {
		phase.buildContextArchive = image.NewBuildContextArchive(phase.Conveyor.giterminismManager, img.TmpDir)
		if err := phase.buildContextArchive.Create(ctx, container_backend.BuildContextArchiveCreateOptions{
			DockerfileRelToContextPath: img.DockerfileImageConfig.Dockerfile,
			ContextGitSubDir:           img.DockerfileImageConfig.Context,
			ContextAddFiles:            img.DockerfileImageConfig.ContextAddFiles,
		}); err != nil {
			return nil, fmt.Errorf("unable to create build context archive: %w", err)
		}

		deferFn = func() {
			phase.buildContextArchive.CleanupExtractedDir(ctx)
		}
	}

	return deferFn, nil
}

func (phase *BuildPhase) AfterImageStages(ctx context.Context, img *image.Image) error {
	img.SetLastNonEmptyStage(phase.StagesIterator.PrevNonEmptyStage)
	img.SetContentDigest(phase.StagesIterator.PrevNonEmptyStage.GetContentDigest())

	if err := phase.addManagedImage(ctx, img); err != nil {
		return err
	}

	if !phase.BuildPhaseOptions.SkipImageMetadataPublication {
		if err := phase.publishImageMetadata(ctx, img); err != nil {
			return err
		}
	}

	if img.IsArtifact {
		return nil
	}
	if img.IsDockerfileImage && !img.IsDockerfileTargetStage {
		return nil
	}

	if phase.Conveyor.StorageManager.GetFinalStagesStorage() != nil {
		if err := phase.Conveyor.StorageManager.CopyStageIntoFinalStorage(ctx, img.GetLastNonEmptyStage(), phase.Conveyor.ContainerBackend, manager.CopyStageIntoFinalStorageOptions{ShouldBeBuiltMode: phase.ShouldBeBuiltMode}); err != nil {
			return err
		}
	}

	var customTagStorage storage.StagesStorage
	var customTagStage *imagePkg.StageDescription
	if phase.Conveyor.StorageManager.GetFinalStagesStorage() != nil {
		customTagStorage = phase.Conveyor.StorageManager.GetFinalStagesStorage()
		customTagStage = manager.ConvertStageDescriptionForStagesStorage(img.GetLastNonEmptyStage().GetStageImage().Image.GetStageDescription(), phase.Conveyor.StorageManager.GetFinalStagesStorage())
	} else {
		customTagStorage = phase.Conveyor.StorageManager.GetStagesStorage()
		customTagStage = img.GetLastNonEmptyStage().GetStageImage().Image.GetStageDescription()
	}

	if phase.ShouldBeBuiltMode {
		if err := phase.checkCustomImageTagsExistence(ctx, img.GetName(), customTagStage, customTagStorage); err != nil {
			return err
		}
	} else {
		if err := phase.addCustomImageTags(ctx, img.GetName(), customTagStage, customTagStorage, phase.Conveyor.StorageManager.GetStagesStorage(), phase.CustomTagFuncList); err != nil {
			return fmt.Errorf("unable to add custom image tags to stages storage: %w", err)
		}
	}

	return nil
}

func (phase *BuildPhase) addManagedImage(ctx context.Context, img *image.Image) error {
	if phase.ShouldAddManagedImageRecord {
		stagesStorage := phase.Conveyor.StorageManager.GetStagesStorage()
		exist, err := stagesStorage.IsManagedImageExist(ctx, phase.Conveyor.ProjectName(), img.GetName(), storage.WithCache())
		if err != nil {
			return fmt.Errorf("unable to check existence of managed image: %w", err)
		}

		if exist {
			return nil
		}

		if err := stagesStorage.AddManagedImage(ctx, phase.Conveyor.ProjectName(), img.GetName()); err != nil {
			return fmt.Errorf("unable to add image %q to the managed images of project %q: %w", img.GetName(), phase.Conveyor.ProjectName(), err)
		}
	}

	return nil
}

func (phase *BuildPhase) publishImageMetadata(ctx context.Context, img *image.Image) error {
	return logboek.Context(ctx).Info().LogProcess(fmt.Sprintf("Processing image %s git metadata", img.GetName())).
		DoError(func() error {
			var commits []string

			headCommit := phase.Conveyor.giterminismManager.HeadCommit()
			commits = append(commits, headCommit)

			if phase.Conveyor.GetLocalGitRepoVirtualMergeOptions().VirtualMerge {
				fromCommit, _, err := git_repo.GetVirtualMergeParents(ctx, phase.Conveyor.giterminismManager.LocalGitRepo(), headCommit)
				if err != nil {
					return fmt.Errorf("unable to get virtual merge commit %q parents: %w", headCommit, err)
				}

				commits = append(commits, fromCommit)
			}

			for _, commit := range commits {
				stagesStorage := phase.Conveyor.StorageManager.GetStagesStorage()
				exist, err := stagesStorage.IsImageMetadataExist(ctx, phase.Conveyor.ProjectName(), img.GetName(), commit, img.GetStageID(), storage.WithCache())
				if err != nil {
					return fmt.Errorf("unable to get image %s metadata by commit %s and stage ID %s: %w", img.GetName(), commit, img.GetStageID(), err)
				}

				if !exist {
					if err := stagesStorage.PutImageMetadata(ctx, phase.Conveyor.ProjectName(), img.GetName(), commit, img.GetStageID()); err != nil {
						return fmt.Errorf("unable to put image %s metadata by commit %s and stage ID %s: %w", img.GetName(), commit, img.GetStageID(), err)
					}
				}
			}

			return nil
		})
}

func (phase *BuildPhase) addCustomImageTags(ctx context.Context, imageName string, stageDesc *imagePkg.StageDescription, stagesStorage storage.StagesStorage, primaryStagesStorage storage.PrimaryStagesStorage, customTagFuncList []imagePkg.CustomTagFunc) error {
	if len(customTagFuncList) == 0 {
		return nil
	}

	return logboek.Context(ctx).Default().LogProcess("Adding custom tags").
		Options(func(options types.LogProcessOptionsInterface) {
			options.Style(style.Highlight())
		}).
		DoError(func() error {
			for _, tagFunc := range customTagFuncList {
				tag := tagFunc(imageName, stageDesc.Info.Tag)
				if err := addCustomImageTag(ctx, stagesStorage, primaryStagesStorage, stageDesc, tag); err != nil {
					return err
				}
			}

			return nil
		})
}

func addCustomImageTag(ctx context.Context, stagesStorage storage.StagesStorage, primaryStagesStorage storage.PrimaryStagesStorage, stageDesc *imagePkg.StageDescription, tag string) error {
	return logboek.Context(ctx).Default().LogProcess("tag %s", tag).
		DoError(func() error {
			if err := stagesStorage.AddStageCustomTag(ctx, stageDesc, tag); err != nil {
				return fmt.Errorf("unable to add stage %s custom tag %s in the storage %s: %w", stageDesc.StageID.String(), tag, stagesStorage.String(), err)
			}
			if err := primaryStagesStorage.RegisterStageCustomTag(ctx, stageDesc, tag); err != nil {
				return fmt.Errorf("unable to register stage %s custom tag %s in the primary storage %s: %w", stageDesc.StageID.String(), tag, primaryStagesStorage.String(), err)
			}

			logboek.Context(ctx).LogFDetails("  name: %s:%s\n", stageDesc.Info.Repository, tag)

			return nil
		})
}

func (phase *BuildPhase) checkCustomImageTagsExistence(ctx context.Context, imageName string, stageDesc *imagePkg.StageDescription, stagesStorage storage.StagesStorage) error {
	if len(phase.CustomTagFuncList) == 0 {
		return nil
	}

	for _, tagFunc := range phase.CustomTagFuncList {
		tag := tagFunc(imageName, stageDesc.Info.Tag)
		if err := stagesStorage.CheckStageCustomTag(ctx, stageDesc, tag); err != nil {
			return fmt.Errorf("check custom tag %q existence failed: %w", tag, err)
		}
	}

	return nil
}

func (phase *BuildPhase) getPrevNonEmptyStageImageSize() int64 {
	if phase.StagesIterator.PrevNonEmptyStage != nil {
		if phase.StagesIterator.PrevNonEmptyStage.GetStageImage().Image.GetStageDescription() != nil {
			return phase.StagesIterator.PrevNonEmptyStage.GetStageImage().Image.GetStageDescription().Info.Size
		}
	}
	return 0
}

func (phase *BuildPhase) OnImageStage(ctx context.Context, img *image.Image, stg stage.Interface) error {
	return phase.StagesIterator.OnImageStage(ctx, img, stg, func(img *image.Image, stg stage.Interface, isEmpty bool) error {
		return phase.onImageStage(ctx, img, stg, isEmpty)
	})
}

func (phase *BuildPhase) onImageStage(ctx context.Context, img *image.Image, stg stage.Interface, isEmpty bool) error {
	if isEmpty {
		return nil
	}

	if err := stg.FetchDependencies(ctx, phase.Conveyor, phase.Conveyor.ContainerBackend, docker_registry.API()); err != nil {
		return fmt.Errorf("unable to fetch dependencies for stage %s: %w", stg.LogDetailedName(), err)
	}

	if stg.HasPrevStage() {
		if phase.StagesIterator.PrevNonEmptyStage == nil {
			panic(fmt.Sprintf("expected PrevNonEmptyStage to be set for image %q stage %s", img.GetName(), stg.Name()))
		}
		if phase.StagesIterator.PrevBuiltStage == nil {
			panic(fmt.Sprintf("expected PrevBuiltStage to be set for image %q stage %s", img.GetName(), stg.Name()))
		}
		if phase.StagesIterator.PrevBuiltStage != phase.StagesIterator.PrevNonEmptyStage {
			panic(fmt.Sprintf("expected PrevBuiltStage (%q) to equal PrevNonEmptyStage (%q) for image %q stage %s", phase.StagesIterator.PrevBuiltStage.LogDetailedName(), phase.StagesIterator.PrevNonEmptyStage.LogDetailedName(), img.GetName(), stg.Name()))
		}
	}

	foundSuitableStage, cleanupFunc, err := phase.calculateStage(ctx, img, stg)
	if cleanupFunc != nil {
		defer cleanupFunc()
	}
	if err != nil {
		return err
	}

	if foundSuitableStage {
		logboek.Context(ctx).Default().LogFHighlight("Use cache image for %s\n", stg.LogDetailedName())
		container_backend.LogImageInfo(ctx, stg.GetStageImage().Image, phase.getPrevNonEmptyStageImageSize())

		logboek.Context(ctx).LogOptionalLn()

		if phase.IntrospectOptions.ImageStageShouldBeIntrospected(img.GetName(), string(stg.Name())) {
			if err := introspectStage(ctx, stg); err != nil {
				return err
			}
		}

		return nil
	}

	foundSuitableSecondaryStage, err := phase.findAndFetchStageFromSecondaryStagesStorage(ctx, img, stg)
	if err != nil {
		return err
	}

	if !foundSuitableSecondaryStage {
		if phase.ShouldBeBuiltMode {
			phase.printShouldBeBuiltError(ctx, img, stg)
			return fmt.Errorf("stages required")
		}

		// Will build a new stage
		i := phase.Conveyor.GetOrCreateStageImage(uuid.New().String(), phase.StagesIterator.GetPrevImage(img, stg), stg, img)
		stg.SetStageImage(i)

		if err := phase.fetchBaseImageForStage(ctx, img, stg); err != nil {
			return err
		}
		if err := phase.prepareStageInstructions(ctx, img, stg); err != nil {
			return err
		}
		if err := phase.buildStage(ctx, img, stg); err != nil {
			return err
		}
	}

	if stg.GetStageImage().Image.GetStageDescription() == nil {
		panic(fmt.Sprintf("expected stage %s image %q built image info (image name = %s) to be set!", stg.Name(), img.GetName(), stg.GetStageImage().Image.Name()))
	}

	// Add managed image record only if there was at least one newly built stage
	phase.ShouldAddManagedImageRecord = true

	return nil
}

func (phase *BuildPhase) findAndFetchStageFromSecondaryStagesStorage(ctx context.Context, img *image.Image, stg stage.Interface) (bool, error) {
	foundSuitableStage := false

	storageManager := phase.Conveyor.StorageManager
	atomicCopySuitableStageFromSecondaryStagesStorage := func(secondaryStageDesc *imagePkg.StageDescription, secondaryStagesStorage storage.StagesStorage) error {
		// Lock the primary stages storage
		var stageUnlocked bool
		var unlockStage func()
		if lock, err := phase.Conveyor.StorageLockManager.LockStage(ctx, phase.Conveyor.ProjectName(), stg.GetDigest()); err != nil {
			return fmt.Errorf("unable to lock project %s digest %s: %w", phase.Conveyor.ProjectName(), stg.GetDigest(), err)
		} else {
			unlockStage = func() {
				if stageUnlocked {
					return
				}
				phase.Conveyor.StorageLockManager.Unlock(ctx, lock)
				stageUnlocked = true
			}
			defer unlockStage()
		}

		err := logboek.Context(ctx).Default().LogProcess("Copy suitable stage from secondary %s", secondaryStagesStorage.String()).DoError(func() error {
			// Copy suitable stage from a secondary stages storage to the primary stages storage
			// while primary stages storage lock for this digest is held
			if copiedStageDesc, err := storageManager.CopySuitableByDigestStage(ctx, secondaryStageDesc, secondaryStagesStorage, storageManager.GetStagesStorage(), phase.Conveyor.ContainerBackend, img.TargetPlatform); err != nil {
				return fmt.Errorf("unable to copy suitable stage %s from %s to %s: %w", secondaryStageDesc.StageID.String(), secondaryStagesStorage.String(), storageManager.GetStagesStorage().String(), err)
			} else {
				i := phase.Conveyor.GetOrCreateStageImage(copiedStageDesc.Info.Name, phase.StagesIterator.GetPrevImage(img, stg), stg, img)
				i.Image.SetStageDescription(copiedStageDesc)
				stg.SetStageImage(i)

				logboek.Context(ctx).Default().LogFHighlight("Use cache image for %s\n", stg.LogDetailedName())
				container_backend.LogImageInfo(ctx, stg.GetStageImage().Image, phase.getPrevNonEmptyStageImageSize())

				return nil
			}
		})
		if err != nil {
			return err
		}

		unlockStage()

		if err := storageManager.CopyStageIntoCacheStorages(ctx, stg, phase.Conveyor.ContainerBackend); err != nil {
			return fmt.Errorf("unable to copy stage %s into cache storages: %w", stg.GetStageImage().Image.GetStageDescription().StageID.String(), err)
		}

		return nil
	}

ScanSecondaryStagesStorageList:
	for _, secondaryStagesStorage := range storageManager.GetSecondaryStagesStorageList() {
		secondaryStages, err := storageManager.GetStagesByDigestFromStagesStorageWithCache(ctx, stg.LogDetailedName(), stg.GetDigest(), secondaryStagesStorage)
		if err != nil {
			return false, err
		} else {
			if secondaryStageDesc, err := storageManager.SelectSuitableStage(ctx, phase.Conveyor, stg, secondaryStages); err != nil {
				return false, err
			} else if secondaryStageDesc != nil {
				if err := atomicCopySuitableStageFromSecondaryStagesStorage(secondaryStageDesc, secondaryStagesStorage); err != nil {
					return false, fmt.Errorf("unable to copy suitable stage %s from secondary stages storage %s: %w", secondaryStageDesc.StageID.String(), secondaryStagesStorage.String(), err)
				}
				foundSuitableStage = true
				break ScanSecondaryStagesStorageList
			}
		}
	}

	return foundSuitableStage, nil
}

func (phase *BuildPhase) fetchBaseImageForStage(ctx context.Context, img *image.Image, stg stage.Interface) error {
	if stg.HasPrevStage() {
		return phase.Conveyor.StorageManager.FetchStage(ctx, phase.Conveyor.ContainerBackend, phase.StagesIterator.PrevBuiltStage)
	} else {
		if err := img.FetchBaseImage(ctx); err != nil {
			return fmt.Errorf("unable to fetch base image %s for stage %s: %w", img.GetBaseStageImage().Image.Name(), stg.LogDetailedName(), err)
		}
	}
	return nil
}

func (phase *BuildPhase) calculateStage(ctx context.Context, img *image.Image, stg stage.Interface) (bool, func(), error) {
	// FIXME(stapel-to-buildah): store StageImage-s everywhere in stage and build pkgs
	stageDependencies, err := stg.GetDependencies(ctx, phase.Conveyor, phase.Conveyor.ContainerBackend, phase.StagesIterator.GetPrevImage(img, stg), phase.StagesIterator.GetPrevBuiltImage(img, stg), phase.buildContextArchive)
	if err != nil {
		return false, nil, err
	}

	var opts calculateDigestOption
	if img.IsDockerfileImage && img.DockerfileImageConfig.Staged {
		if !stg.HasPrevStage() {
			opts.BaseImage = img.GetBaseImageReference()
		}
	}
	opts.TargetPlatform = img.TargetPlatform

	stageDigest, err := calculateDigest(ctx, stage.GetLegacyCompatibleStageName(stg.Name()), stageDependencies, phase.StagesIterator.PrevNonEmptyStage, phase.Conveyor, opts)
	if err != nil {
		return false, nil, err
	}
	stg.SetDigest(stageDigest)

	logboek.Context(ctx).Info().LogProcessInline("Lock parallel conveyor tasks by stage digest %s", stg.LogDetailedName()).
		Options(func(options types.LogProcessInlineOptionsInterface) {
			if !phase.Conveyor.Parallel {
				options.Mute()
			}
		}).
		Do(phase.Conveyor.GetStageDigestMutex(stg.GetDigest()).Lock)

	foundSuitableStage := false
	storageManager := phase.Conveyor.StorageManager
	if stages, err := storageManager.GetStagesByDigestWithCache(ctx, stg.LogDetailedName(), stageDigest); err != nil {
		return false, phase.Conveyor.GetStageDigestMutex(stg.GetDigest()).Unlock, err
	} else {
		if stageDesc, err := storageManager.SelectSuitableStage(ctx, phase.Conveyor, stg, stages); err != nil {
			return false, phase.Conveyor.GetStageDigestMutex(stg.GetDigest()).Unlock, err
		} else if stageDesc != nil {
			i := phase.Conveyor.GetOrCreateStageImage(stageDesc.Info.Name, phase.StagesIterator.GetPrevImage(img, stg), stg, img)
			i.Image.SetStageDescription(stageDesc)
			stg.SetStageImage(i)
			foundSuitableStage = true
		}
	}

	stageContentSig, err := calculateDigest(ctx, fmt.Sprintf("%s-content", stg.Name()), "", stg, phase.Conveyor, calculateDigestOption{TargetPlatform: img.TargetPlatform})
	if err != nil {
		return false, phase.Conveyor.GetStageDigestMutex(stg.GetDigest()).Unlock, fmt.Errorf("unable to calculate stage %s content digest: %w", stg.Name(), err)
	}
	stg.SetContentDigest(stageContentSig)

	logboek.Context(ctx).Info().LogF("Stage %s content digest: %s\n", stg.LogDetailedName(), stageContentSig)

	return foundSuitableStage, phase.Conveyor.GetStageDigestMutex(stg.GetDigest()).Unlock, nil
}

func (phase *BuildPhase) prepareStageInstructions(ctx context.Context, img *image.Image, stg stage.Interface) error {
	logboek.Context(ctx).Debug().LogF("-- BuildPhase.prepareStage %s %s\n", img.LogDetailedName(), stg.LogDetailedName())

	stageImage := stg.GetStageImage()

	serviceLabels := map[string]string{
		imagePkg.WerfDockerImageName:         stageImage.Image.Name(),
		imagePkg.WerfLabel:                   phase.Conveyor.ProjectName(),
		imagePkg.WerfVersionLabel:            werf.Version,
		imagePkg.WerfCacheVersionLabel:       imagePkg.BuildCacheVersion,
		imagePkg.WerfImageLabel:              "false",
		imagePkg.WerfStageDigestLabel:        stg.GetDigest(),
		imagePkg.WerfStageContentDigestLabel: stg.GetContentDigest(),
	}

	if stg.IsStapelStage() {
		if phase.Conveyor.UseLegacyStapelBuilder(phase.Conveyor.ContainerBackend) {
			stageImage.Builder.LegacyStapelStageBuilder().Container().ServiceCommitChangeOptions().AddLabel(serviceLabels)
		} else {
			stageImage.Builder.StapelStageBuilder().AddLabels(serviceLabels)
		}

		if phase.Conveyor.sshAuthSock != "" {
			if runtime.GOOS == "darwin" {
				if phase.Conveyor.UseLegacyStapelBuilder(phase.Conveyor.ContainerBackend) {
					stageImage.Builder.LegacyStapelStageBuilder().Container().RunOptions().AddVolume("/run/host-services/ssh-auth.sock:/run/host-services/ssh-auth.sock")
					stageImage.Builder.LegacyStapelStageBuilder().Container().RunOptions().AddEnv(map[string]string{"SSH_AUTH_SOCK": "/run/host-services/ssh-auth.sock"})
				} else {
					stageImage.Builder.StapelStageBuilder().AddBuildVolumes("/run/host-services/ssh-auth.sock:/run/host-services/ssh-auth.sock")
					stageImage.Builder.StapelStageBuilder().AddEnvs(map[string]string{"SSH_AUTH_SOCK": "/run/host-services/ssh-auth.sock"})
				}
			} else {
				if phase.Conveyor.UseLegacyStapelBuilder(phase.Conveyor.ContainerBackend) {
					stageImage.Builder.LegacyStapelStageBuilder().Container().RunOptions().AddVolume(fmt.Sprintf("%s:/.werf/tmp/ssh-auth-sock", phase.Conveyor.sshAuthSock))
					stageImage.Builder.LegacyStapelStageBuilder().Container().RunOptions().AddEnv(map[string]string{"SSH_AUTH_SOCK": "/.werf/tmp/ssh-auth-sock"})
				} else {
					stageImage.Builder.StapelStageBuilder().AddBuildVolumes(fmt.Sprintf("%s:/.werf/tmp/ssh-auth-sock", phase.Conveyor.sshAuthSock))
					stageImage.Builder.StapelStageBuilder().AddEnvs(map[string]string{"SSH_AUTH_SOCK": "/.werf/tmp/ssh-auth-sock"})
				}
			}
		}

		headHash, err := phase.Conveyor.GiterminismManager().LocalGitRepo().HeadCommitHash(ctx)
		if err != nil {
			return fmt.Errorf("error getting HEAD commit hash: %w", err)
		}
		headTime, err := phase.Conveyor.GiterminismManager().LocalGitRepo().HeadCommitTime(ctx)
		if err != nil {
			return fmt.Errorf("error getting HEAD commit time: %w", err)
		}
		commitEnvs := map[string]string{
			"WERF_COMMIT_HASH":       headHash,
			"WERF_COMMIT_TIME_HUMAN": headTime.String(),
			"WERF_COMMIT_TIME_UNIX":  strconv.FormatInt(headTime.Unix(), 10),
		}

		if phase.Conveyor.UseLegacyStapelBuilder(phase.Conveyor.ContainerBackend) {
			stageImage.Builder.LegacyStapelStageBuilder().Container().RunOptions().AddEnv(commitEnvs)
		} else {
			stageImage.Builder.StapelStageBuilder().AddEnvs(commitEnvs)
		}
	} else if _, ok := stg.(*stage.FullDockerfileStage); ok {
		var labels []string
		for key, value := range serviceLabels {
			labels = append(labels, fmt.Sprintf("%s=%v", key, value))
		}
		stageImage.Builder.DockerfileBuilder().AppendLabels(labels...)

		phase.Conveyor.AppendOnTerminateFunc(func() error {
			return stageImage.Builder.DockerfileBuilder().Cleanup(ctx)
		})
	} else {
		stageImage.Builder.DockerfileStageBuilder().SetBuildContextArchive(phase.buildContextArchive)

		for k, v := range serviceLabels {
			stageImage.Builder.DockerfileStageBuilder().AppendPostInstruction(
				backend_instruction.NewLabel(*instructions.NewLabelCommand(k, v, true)),
			)
		}
	}

	err := stg.PrepareImage(ctx, phase.Conveyor, phase.Conveyor.ContainerBackend, phase.StagesIterator.GetPrevBuiltImage(img, stg), stageImage, phase.buildContextArchive)
	if err != nil {
		return fmt.Errorf("error preparing stage %s: %w", stg.Name(), err)
	}

	return nil
}

func (phase *BuildPhase) buildStage(ctx context.Context, img *image.Image, stg stage.Interface) error {
	if !img.IsDockerfileImage && phase.Conveyor.UseLegacyStapelBuilder(phase.Conveyor.ContainerBackend) {
		_, err := stapel.GetOrCreateContainer(ctx)
		if err != nil {
			return fmt.Errorf("get or create stapel container failed: %w", err)
		}
	}

	infoSectionFunc := func(err error) {
		if err != nil {
			return
		}
		container_backend.LogImageInfo(ctx, stg.GetStageImage().Image, phase.getPrevNonEmptyStageImageSize())
	}

	if err := logboek.Context(ctx).Default().LogProcess("Building stage %s", stg.LogDetailedName()).
		Options(func(options types.LogProcessOptionsInterface) {
			options.InfoSectionFunc(infoSectionFunc)
			options.Style(style.Highlight())
		}).
		DoError(func() (err error) {
			if err := stg.PreRunHook(ctx, phase.Conveyor); err != nil {
				return fmt.Errorf("%s preRunHook failed: %w", stg.LogDetailedName(), err)
			}

			return phase.atomicBuildStageImage(ctx, img, stg)
		}); err != nil {
		return err
	}

	if phase.IntrospectOptions.ImageStageShouldBeIntrospected(img.GetName(), string(stg.Name())) {
		if err := introspectStage(ctx, stg); err != nil {
			return err
		}
	}

	return nil
}

func (phase *BuildPhase) atomicBuildStageImage(ctx context.Context, img *image.Image, stg stage.Interface) error {
	stageImage := stg.GetStageImage()

	if v := os.Getenv("WERF_TEST_ATOMIC_STAGE_BUILD__SLEEP_SECONDS_BEFORE_STAGE_BUILD"); v != "" {
		seconds := 0
		fmt.Sscanf(v, "%d", &seconds)
		fmt.Printf("Sleeping %d seconds before building new image by digest %s...\n", seconds, stg.GetDigest())
		time.Sleep(time.Duration(seconds) * time.Second)
	}

	if err := logboek.Context(ctx).Streams().DoErrorWithTag(fmt.Sprintf("%s/%s", img.LogName(), stg.Name()), img.LogTagStyle(), func() error {
		opts := phase.ImageBuildOptions
		opts.TargetPlatform = img.TargetPlatform
		return stageImage.Builder.Build(ctx, opts)
	}); err != nil {
		return fmt.Errorf("failed to build image for stage %s with digest %s: %w", stg.Name(), stg.GetDigest(), err)
	}

	if v := os.Getenv("WERF_TEST_ATOMIC_STAGE_BUILD__SLEEP_SECONDS_BEFORE_STAGE_SAVE"); v != "" {
		seconds := 0
		fmt.Sscanf(v, "%d", &seconds)
		fmt.Printf("Sleeping %d seconds before saving newly built image %s into repo %s by digest %s...\n", seconds, stg.GetStageImage().Image.GetBuiltID(), phase.Conveyor.StorageManager.GetStagesStorage().String(), stg.GetDigest())
		time.Sleep(time.Duration(seconds) * time.Second)
	}

	var stageUnlocked bool
	var unlockStage func()
	if lock, err := phase.Conveyor.StorageLockManager.LockStage(ctx, phase.Conveyor.ProjectName(), stg.GetDigest()); err != nil {
		return fmt.Errorf("unable to lock project %s digest %s: %w", phase.Conveyor.ProjectName(), stg.GetDigest(), err)
	} else {
		unlockStage = func() {
			if stageUnlocked {
				return
			}
			phase.Conveyor.StorageLockManager.Unlock(ctx, lock)
			stageUnlocked = true
		}
		defer unlockStage()
	}

	if stages, err := phase.Conveyor.StorageManager.GetStagesByDigest(ctx, stg.LogDetailedName(), stg.GetDigest()); err != nil {
		return err
	} else {
		stageDesc, err := phase.Conveyor.StorageManager.SelectSuitableStage(ctx, phase.Conveyor, stg, stages)
		if err != nil {
			return err
		}

		if stageDesc != nil {
			logboek.Context(ctx).Default().LogF(
				"Discarding newly built image for stage %s by digest %s: detected already existing image %s in the repo\n",
				stg.LogDetailedName(), stg.GetDigest(), stageDesc.Info.Name,
			)

			i := phase.Conveyor.GetOrCreateStageImage(stageDesc.Info.Name, phase.StagesIterator.GetPrevImage(img, stg), stg, img)
			i.Image.SetStageDescription(stageDesc)
			stg.SetStageImage(i)

			return nil
		}

		// use newly built image
		newStageImageName, uniqueID := phase.Conveyor.StorageManager.GenerateStageUniqueID(stg.GetDigest(), stages)
		phase.Conveyor.UnsetStageImage(stageImage.Image.Name())
		stageImage.Image.SetName(newStageImageName)
		phase.Conveyor.SetStageImage(stageImage)

		if err := logboek.Context(ctx).Default().LogProcess("Store stage into %s", phase.Conveyor.StorageManager.GetStagesStorage().String()).DoError(func() error {
			if err := phase.Conveyor.StorageManager.GetStagesStorage().StoreImage(ctx, stageImage.Image); err != nil {
				return fmt.Errorf("unable to store stage %s digest %s image %s into repo %s: %w", stg.LogDetailedName(), stg.GetDigest(), stageImage.Image.Name(), phase.Conveyor.StorageManager.GetStagesStorage().String(), err)
			}

			if desc, err := phase.Conveyor.StorageManager.GetStagesStorage().GetStageDescription(ctx, phase.Conveyor.ProjectName(), stg.GetDigest(), uniqueID); err != nil {
				return fmt.Errorf("unable to get stage %s digest %s image %s description from repo %s after stages has been stored into repo: %w", stg.LogDetailedName(), stg.GetDigest(), stageImage.Image.Name(), phase.Conveyor.StorageManager.GetStagesStorage().String(), err)
			} else {
				stageImage.Image.SetStageDescription(desc)
			}

			img.SetRebuilt(true)

			return nil
		}); err != nil {
			return err
		}

		unlockStage()

		if err := phase.Conveyor.StorageManager.CopyStageIntoCacheStorages(ctx, stg, phase.Conveyor.ContainerBackend); err != nil {
			return fmt.Errorf("unable to copy stage %s into cache storages: %w", stageImage.Image.GetStageDescription().StageID.String(), err)
		}
		return nil
	}
}

func introspectStage(ctx context.Context, s stage.Interface) error {
	return logboek.Context(ctx).Info().LogProcess("Introspecting stage %s", s.Name()).
		Options(func(options types.LogProcessOptionsInterface) {
			options.Style(style.Highlight())
		}).
		DoError(func() error {
			if err := logboek.Context(ctx).Streams().DoErrorWithoutProxyStreamDataFormatting(func() error {
				return s.GetStageImage().Image.Introspect(ctx) // FIXME(stapel-to-buildah): use container backend operation
			}); err != nil {
				return fmt.Errorf("introspect error failed: %w", err)
			}

			return nil
		})
}

type calculateDigestOption struct {
	BaseImage      string
	TargetPlatform string
}

func calculateDigest(ctx context.Context, stageName, stageDependencies string, prevNonEmptyStage stage.Interface, conveyor *Conveyor, opts calculateDigestOption) (string, error) {
	var checksumArgs []string
	var checksumArgsNames []string

	// linux/amd64 not affects digest for compatibility with currently built stages
	if opts.TargetPlatform != "" && opts.TargetPlatform != "linux/amd64" {
		checksumArgs = append(checksumArgs, opts.TargetPlatform)
		checksumArgsNames = append(checksumArgsNames, "TargetPlatform")
	}

	checksumArgs = append(checksumArgs, imagePkg.BuildCacheVersion, stageName, stageDependencies)
	checksumArgsNames = append(checksumArgsNames,
		"BuildCacheVersion",
		"StageName",
		"StageDependencies",
	)

	if prevNonEmptyStage != nil {
		prevStageDependencies, err := prevNonEmptyStage.GetNextStageDependencies(ctx, conveyor)
		if err != nil {
			return "", fmt.Errorf("unable to get prev stage %s dependencies for the stage %s: %w", prevNonEmptyStage.Name(), stageName, err)
		}

		checksumArgs = append(checksumArgs, prevNonEmptyStage.GetDigest(), prevStageDependencies)
		checksumArgsNames = append(checksumArgsNames,
			"PrevNonEmptyStage digest",
			"PrevNonEmptyStage dependencies for next stage",
		)
	}

	if opts.BaseImage != "" {
		checksumArgs = append(checksumArgs, opts.BaseImage)
		checksumArgsNames = append(checksumArgsNames, "BaseImage")
	}

	digest := util.Sha3_224Hash(checksumArgs...)

	blockMsg := fmt.Sprintf("Stage %s digest %s", stageName, digest)
	logboek.Context(ctx).Debug().LogBlock(blockMsg).Do(func() {
		for ind, checksumArg := range checksumArgs {
			logboek.Context(ctx).Debug().LogF("%s => %q\n", checksumArgsNames[ind], checksumArg)
		}
	})

	return digest, nil
}

// TODO: move these prints to the after-images hook, print summary over all images
func (phase *BuildPhase) printShouldBeBuiltError(ctx context.Context, img *image.Image, stg stage.Interface) {
	logboek.Context(ctx).Default().LogProcess("Built stages cache check").
		Options(func(options types.LogProcessOptionsInterface) {
			options.Style(style.Highlight())
		}).
		Do(func() {
			logboek.Context(ctx).Warn().LogF("%s with digest %s is not exist in repo\n", stg.LogDetailedName(), stg.GetDigest())

			var reasonNumber int
			reasonNumberFunc := func() string {
				reasonNumber++
				return fmt.Sprintf("(%d) ", reasonNumber)
			}

			logboek.Context(ctx).Warn().LogLn()
			logboek.Context(ctx).Warn().LogLn("There are some possible reasons:")
			logboek.Context(ctx).Warn().LogLn()

			if img.IsDockerfileImage {
				logboek.Context(ctx).Warn().LogLn(reasonNumberFunc() + `Dockerfile has COPY or ADD instruction which uses non-permanent data that affects stage digest:
- .git directory which should be excluded with .dockerignore file (https://docs.docker.com/engine/reference/builder/#dockerignore-file)
- auto-generated file`)
				logboek.Context(ctx).Warn().LogLn()
			}

			logboek.Context(ctx).Warn().LogLn(reasonNumberFunc() + `werf.yaml has non-permanent data that affects stage digest:
- environment variable (e.g. {{ env "JOB_ID" }})
- dynamic go template function (e.g. one of sprig date functions http://masterminds.github.io/sprig/date.html)
- auto-generated file content (e.g. {{ .Files.Get "hash_sum_of_something" }})`)
			logboek.Context(ctx).Warn().LogLn()

			logboek.Context(ctx).Warn().LogLn(`Stage digest dependencies can be found here, https://werf.io/documentation/reference/stages_and_images.html#stage-dependencies.

To quickly find the problem compare current and previous rendered werf configurations.
Get the path at the beginning of command output by the following prefix 'Using werf config render file: '.
E.g.:

  diff /tmp/werf-config-render-502883762 /tmp/werf-config-render-837625028`)
			logboek.Context(ctx).Warn().LogLn()

			logboek.Context(ctx).Warn().LogLn(reasonNumberFunc() + `Stages have not been built yet or stages have been removed:
- automatically with werf cleanup command
- manually with werf purge or werf host purge commands`)
			logboek.Context(ctx).Warn().LogLn()
		})
}

func (phase *BuildPhase) Clone() Phase {
	u := *phase
	return &u
}
