package build

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/moby/buildkit/frontend/dockerfile/instructions"

	"github.com/werf/common-go/pkg/util"
	"github.com/werf/logboek"
	"github.com/werf/logboek/pkg/style"
	"github.com/werf/logboek/pkg/types"
	"github.com/werf/werf/v2/pkg/build/cleanup"
	"github.com/werf/werf/v2/pkg/build/image"
	"github.com/werf/werf/v2/pkg/build/stage"
	"github.com/werf/werf/v2/pkg/build/stage/instruction"
	"github.com/werf/werf/v2/pkg/container_backend"
	backend_instruction "github.com/werf/werf/v2/pkg/container_backend/instruction"
	"github.com/werf/werf/v2/pkg/docker_registry"
	"github.com/werf/werf/v2/pkg/git_repo"
	imagePkg "github.com/werf/werf/v2/pkg/image"
	"github.com/werf/werf/v2/pkg/logging"
	"github.com/werf/werf/v2/pkg/stapel"
	"github.com/werf/werf/v2/pkg/storage"
	"github.com/werf/werf/v2/pkg/storage/manager"
	"github.com/werf/werf/v2/pkg/util/parallel"
	"github.com/werf/werf/v2/pkg/werf"
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
	SkipAddManagedImagesRecords  bool
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
		ImagesReport:      NewImagesReport(),
	}
}

type BuildPhase struct {
	BasePhase
	BuildPhaseOptions

	StagesIterator *StagesIterator
	ImagesReport   *ImagesReport

	buildContextArchive container_backend.BuildContextArchiver
}

func GenerateImageEnv(werfImageName, imageName string) string {
	var imageEnvName string
	if werfImageName == "" {
		imageEnvName = "WERF_DOCKER_IMAGE_NAME"
	} else {
		werfImageName := strings.ToUpper(werfImageName)
		for _, l := range []string{"/", "-", "."} {
			werfImageName = strings.ReplaceAll(werfImageName, l, "_")
		}

		imageEnvName = fmt.Sprintf("WERF_%s_DOCKER_IMAGE_NAME", werfImageName)
	}

	return fmt.Sprintf("%s=%s", imageEnvName, imageName)
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
	forcedTargetPlatforms := phase.Conveyor.GetForcedTargetPlatforms()
	commonTargetPlatforms, err := phase.Conveyor.GetTargetPlatforms()
	if err != nil {
		return fmt.Errorf("invalid common target platforms: %w", err)
	}
	if len(commonTargetPlatforms) == 0 {
		commonTargetPlatforms = []string{phase.Conveyor.ContainerBackend.GetDefaultPlatform()}
	}

	imagesPairs := phase.Conveyor.imagesTree.GetImagesByName(false)
	if err := parallel.DoTasks(ctx, len(imagesPairs), parallel.DoTasksOptions{
		MaxNumberOfWorkers: int(phase.Conveyor.ParallelTasksLimit),
	}, func(ctx context.Context, taskId int) error {
		pair := imagesPairs[taskId]

		name, images := pair.Unpair()
		targetPlatforms, err := phase.targetPlatforms(ctx, forcedTargetPlatforms, commonTargetPlatforms, name, images)
		if err != nil {
			return err
		}

		if len(targetPlatforms) == 1 {
			img := images[0]

			if img.IsFinal && phase.Conveyor.StorageManager.GetFinalStagesStorage() != nil {
				if err := phase.publishFinalImage(
					ctx, name, img,
					phase.Conveyor.StorageManager.GetFinalStagesStorage(),
				); err != nil {
					return err
				}
				logboek.Context(ctx).LogOptionalLn()
			}

			// TODO: Separate LocalStagesStorage and RepoStagesStorage interfaces, local should not include metadata publishing methods at all
			if _, isLocal := phase.Conveyor.StorageManager.GetStagesStorage().(*storage.LocalStagesStorage); !isLocal {
				if err := phase.publishImageMetadata(ctx, name, img); err != nil {
					return fmt.Errorf("unable to publish image %q metadata: %w", name, err)
				}
			}
		} else {
			img := image.NewMultiplatformImage(name, images, taskId, len(imagesPairs))
			phase.Conveyor.imagesTree.SetMultiplatformImage(img)

			// TODO: Separate LocalStagesStorage and RepoStagesStorage interfaces, local should not include metadata publishing methods at all
			if _, isLocal := phase.Conveyor.StorageManager.GetStagesStorage().(*storage.LocalStagesStorage); !isLocal {
				if err := logboek.Context(ctx).LogProcess(img.LogDetailedName()).
					Options(func(options types.LogProcessOptionsInterface) {
						options.Style(logging.ImageMetadataStyle())
					}).
					DoError(func() error {
						if err := phase.publishMultiplatformImageMetadata(ctx, name, img, phase.Conveyor.StorageManager.GetFinalStagesStorage()); err != nil {
							return fmt.Errorf("unable to publish image %q multiplatform metadata: %w", name, err)
						}
						return nil
					}); err != nil {
					return err
				}
			}

			if img.IsFinal && phase.Conveyor.StorageManager.GetFinalStagesStorage() != nil {
				if _, isLocal := phase.Conveyor.StorageManager.GetStagesStorage().(*storage.LocalStagesStorage); !isLocal {
					if err := phase.publishMultiplatformFinalImage(ctx, name, img, phase.Conveyor.StorageManager.GetFinalStagesStorage()); err != nil {
						return err
					}
				}
			}
		}

		return nil
	}); err != nil {
		return err
	}

	return phase.createReport(ctx)
}

func (phase *BuildPhase) targetPlatforms(ctx context.Context, forcedTargetPlatforms, commonTargetPlatforms []string, name string, images []*image.Image) ([]string, error) {
	// TODO: this target platforms assertion could be removed in future versions and now exists only as a additional self-testing code
	var targetPlatforms []string
	if len(forcedTargetPlatforms) > 0 {
		targetPlatforms = forcedTargetPlatforms
	} else {
		targetName := name
		nameParts := strings.SplitN(name, "/", 3)
		if len(nameParts) == 3 && nameParts[1] == "stage" {
			targetName = nameParts[0]
		}

		imageTargetPlatforms, err := phase.Conveyor.GetImageTargetPlatforms(targetName)
		if err != nil {
			return []string{}, fmt.Errorf("invalid image %q target platforms: %w", name, err)
		}
		if len(imageTargetPlatforms) > 0 {
			targetPlatforms = imageTargetPlatforms
		} else {
			targetPlatforms = commonTargetPlatforms
		}
	}

	platforms := util.MapFuncToSlice(images, func(img *image.Image) string { return img.TargetPlatform })

AssertAllTargetPlatformsPresent:
	for _, targetPlatform := range targetPlatforms {
		for _, platform := range platforms {
			if targetPlatform == platform {
				logboek.Context(ctx).Debug().LogF("Found image %q built for target platform %q\n", name, targetPlatform)
				continue AssertAllTargetPlatformsPresent
			}
		}
		panic(fmt.Sprintf("There is no image %q built for target platform %q. Please report a bug.", name, targetPlatform))
	}

	if len(targetPlatforms) != len(platforms) {
		panic(fmt.Sprintf("We have built image %q for platforms %v, expected exactly these platforms: %v. Please report a bug.", name, platforms, targetPlatforms))
	}

	return targetPlatforms, nil
}

func (phase *BuildPhase) publishFinalImage(ctx context.Context, name string, img *image.Image, finalStagesStorage storage.StagesStorage) error {
	stg := img.GetLastNonEmptyStage()

	desc, err := phase.Conveyor.StorageManager.CopyStageIntoFinalStorage(
		ctx, *stg.GetStageImage().Image.GetStageDesc().StageID,
		phase.Conveyor.StorageManager.GetFinalStagesStorage(),
		manager.CopyStageIntoStorageOptions{
			ContainerBackend:  phase.Conveyor.ContainerBackend,
			FetchStage:        stg,
			ShouldBeBuiltMode: phase.ShouldBeBuiltMode,
			LogDetailedName:   stg.LogDetailedName(),
		},
	)
	if err != nil {
		return fmt.Errorf("unable to copy image into final repo: %w", err)
	}
	img.GetLastNonEmptyStage().GetStageImage().Image.SetFinalStageDesc(desc)

	return nil
}

func (phase *BuildPhase) publishMultiplatformFinalImage(ctx context.Context, name string, img *image.MultiplatformImage, finalStagesStorage storage.StagesStorage) error {
	desc, err := phase.Conveyor.StorageManager.CopyStageIntoFinalStorage(
		ctx, img.GetStageID(), finalStagesStorage,
		manager.CopyStageIntoStorageOptions{
			ShouldBeBuiltMode:    phase.ShouldBeBuiltMode,
			ContainerBackend:     phase.Conveyor.ContainerBackend,
			LogDetailedName:      img.Name,
			IsMultiplatformImage: true,
		},
	)
	if err != nil {
		return fmt.Errorf("unable to copy image into final repo: %w", err)
	}
	img.SetFinalStageDesc(desc)

	return nil
}

func (phase *BuildPhase) publishImageMetadata(ctx context.Context, name string, img *image.Image) error {
	if err := phase.addManagedImage(ctx, name); err != nil {
		return err
	}

	if !phase.BuildPhaseOptions.SkipImageMetadataPublication {
		if err := logboek.Context(ctx).Info().
			LogProcess(fmt.Sprintf("Publish image %s git metadata", img.GetName())).
			DoError(func() error {
				return phase.publishImageGitMetadata(
					ctx, img.GetName(),
					*img.GetLastNonEmptyStage().GetStageImage().Image.GetStageDesc().StageID,
				)
			}); err != nil {
			return err
		}
	}

	if !img.IsFinal {
		return nil
	}

	var customTagStorage storage.StagesStorage
	var customTagStage *imagePkg.StageDesc
	if phase.Conveyor.StorageManager.GetFinalStagesStorage() != nil {
		customTagStorage = phase.Conveyor.StorageManager.GetFinalStagesStorage()
		customTagStage = manager.ConvertStageDescForStagesStorage(img.GetLastNonEmptyStage().GetStageImage().Image.GetStageDesc(), phase.Conveyor.StorageManager.GetFinalStagesStorage())
	} else {
		customTagStorage = phase.Conveyor.StorageManager.GetStagesStorage()
		customTagStage = img.GetLastNonEmptyStage().GetStageImage().Image.GetStageDesc()
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

func (phase *BuildPhase) publishMultiplatformImageMetadata(ctx context.Context, name string, img *image.MultiplatformImage, finalStagesStorage storage.StagesStorage) error {
	if err := phase.addManagedImage(ctx, name); err != nil {
		return err
	}

	primaryStagesStorage := phase.Conveyor.StorageManager.GetStagesStorage()

	fullImageName := primaryStagesStorage.ConstructStageImageName(phase.Conveyor.ProjectName(), img.GetStageID().Digest, img.GetStageID().CreationTs)
	platforms := img.GetPlatforms()

	container_backend.LogImageName(ctx, fullImageName)
	container_backend.LogMultiplatformImageInfo(ctx, platforms)

	if err := primaryStagesStorage.PostMultiplatformImage(ctx, phase.Conveyor.ProjectName(), img.GetStageID().String(), img.GetImagesInfoList(), platforms); err != nil {
		return fmt.Errorf("unable to post multiplatform image %s %s: %w", name, img.GetStageID(), err)
	}

	desc, err := primaryStagesStorage.GetStageDesc(ctx, phase.Conveyor.ProjectName(), img.GetStageID())
	if err != nil {
		return fmt.Errorf("unable to get image %s %s descriptor: %w", name, img.GetStageID(), err)
	}
	if desc == nil {
		return fmt.Errorf("unable to get image %s %s descriptor: no manifest found", name, img.GetStageID())
	}
	img.SetStageDesc(desc)

	if len(phase.CustomTagFuncList) > 0 {
		logboek.Context(ctx).Default().LogLn()
		logboek.Context(ctx).Default().LogProcess("Adding custom tags").
			Options(func(options types.LogProcessOptionsInterface) {
				options.Style(style.Highlight())
			}).
			DoError(func() error {
				for _, tagFunc := range phase.CustomTagFuncList {
					tag := tagFunc(name, img.GetStageID().String())

					var storage storage.StagesStorage = primaryStagesStorage
					if finalStagesStorage != nil {
						storage = finalStagesStorage
					}

					logboek.Context(ctx).Default().LogProcess("tag %s", tag).
						DoError(func() error {
							// final or primary
							if err := storage.AddStageCustomTag(ctx, desc, tag); err != nil {
								return fmt.Errorf("unable to add stage %s custom tag %s in the storage %s: %w", desc.StageID.String(), tag, storage.String(), err)
							}
							if err := primaryStagesStorage.RegisterStageCustomTag(ctx, phase.Conveyor.ProjectName(), desc, tag); err != nil {
								return fmt.Errorf("unable to register stage %s custom tag %s in the primary storage %s: %w", desc.StageID.String(), tag, primaryStagesStorage.String(), err)
							}
							logboek.Context(ctx).LogFDetails("  name: %s:%s\n", desc.Info.Repository, tag)
							return nil
						})
				}

				return nil
			})
	}

	if !phase.BuildPhaseOptions.SkipImageMetadataPublication {
		if err := logboek.Context(ctx).Info().
			LogProcess(fmt.Sprintf("Publish multiarch image %s git metadata", name)).
			DoError(func() error {
				return phase.publishImageGitMetadata(ctx, name, img.GetStageID())
			}); err != nil {
			return err
		}
	}
	return nil
}

func (phase *BuildPhase) createReport(ctx context.Context) error {
	return createBuildReport(ctx, phase)
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
	return nil
}

func (phase *BuildPhase) addManagedImage(ctx context.Context, name string) error {
	if phase.Conveyor.ShouldAddManagedImagesRecords() {
		stagesStorage := phase.Conveyor.StorageManager.GetStagesStorage()
		exist, err := stagesStorage.IsManagedImageExist(ctx, phase.Conveyor.ProjectName(), name, storage.WithCache())
		if err != nil {
			return fmt.Errorf("unable to check existence of managed image: %w", err)
		}

		if exist {
			return nil
		}

		if err := stagesStorage.AddManagedImage(ctx, phase.Conveyor.ProjectName(), name); err != nil {
			return fmt.Errorf("unable to add image %q to the managed images of project %q: %w", name, phase.Conveyor.ProjectName(), err)
		}
	}

	return nil
}

func (phase *BuildPhase) publishImageGitMetadata(ctx context.Context, imageName string, stageID imagePkg.StageID) error {
	var commits []string

	headCommit := phase.Conveyor.giterminismManager.HeadCommit(ctx)
	commits = append(commits, headCommit)

	if phase.Conveyor.GetLocalGitRepoVirtualMergeOptions().VirtualMerge {
		fromCommit, _, err := git_repo.GetVirtualMergeParents(ctx, phase.Conveyor.giterminismManager.LocalGitRepo(), headCommit)
		if err != nil {
			return fmt.Errorf("unable to get virtual merge commit %q parents: %w", headCommit, err)
		}

		commits = append(commits, fromCommit)
	}

	stagesStorage := phase.Conveyor.StorageManager.GetStagesStorage()

	fullImageName := stagesStorage.ConstructStageImageName(phase.Conveyor.ProjectName(), stageID.Digest, stageID.CreationTs)
	logboek.Context(ctx).Info().LogF("name: %s\n", fullImageName)
	logboek.Context(ctx).Info().LogF("commits:\n")

	for _, commit := range commits {
		logboek.Context(ctx).Info().LogF("  %s\n", commit)

		exist, err := stagesStorage.IsImageMetadataExist(ctx, phase.Conveyor.ProjectName(), imageName, commit, stageID.String(), storage.WithCache())
		if err != nil {
			return fmt.Errorf("unable to get image %s metadata by commit %s and stage ID %s: %w", imageName, commit, stageID.String(), err)
		}

		if !exist {
			if err := stagesStorage.PutImageMetadata(ctx, phase.Conveyor.ProjectName(), imageName, commit, stageID.String()); err != nil {
				return fmt.Errorf("unable to put image %s metadata by commit %s and stage ID %s: %w", imageName, commit, stageID.String(), err)
			}
		}
	}

	return nil
}

func (phase *BuildPhase) addCustomImageTags(ctx context.Context, imageName string, stageDesc *imagePkg.StageDesc, stagesStorage storage.StagesStorage, primaryStagesStorage storage.PrimaryStagesStorage, customTagFuncList []imagePkg.CustomTagFunc) error {
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
				if err := addCustomImageTag(ctx, phase.Conveyor.ProjectName(), stagesStorage, primaryStagesStorage, stageDesc, tag); err != nil {
					return err
				}
			}

			return nil
		})
}

func addCustomImageTag(ctx context.Context, projectName string, stagesStorage storage.StagesStorage, primaryStagesStorage storage.PrimaryStagesStorage, stageDesc *imagePkg.StageDesc, tag string) error {
	return logboek.Context(ctx).Default().LogProcess("tag %s", tag).
		DoError(func() error {
			if err := stagesStorage.AddStageCustomTag(ctx, stageDesc, tag); err != nil {
				return fmt.Errorf("unable to add stage %s custom tag %s in the storage %s: %w", stageDesc.StageID.String(), tag, stagesStorage.String(), err)
			}
			if err := primaryStagesStorage.RegisterStageCustomTag(ctx, projectName, stageDesc, tag); err != nil {
				return fmt.Errorf("unable to register stage %s custom tag %s in the primary storage %s: %w", stageDesc.StageID.String(), tag, primaryStagesStorage.String(), err)
			}

			logboek.Context(ctx).LogFDetails("  name: %s:%s\n", stageDesc.Info.Repository, tag)

			return nil
		})
}

func (phase *BuildPhase) checkCustomImageTagsExistence(ctx context.Context, imageName string, stageDesc *imagePkg.StageDesc, stagesStorage storage.StagesStorage) error {
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
		if phase.StagesIterator.PrevNonEmptyStage.GetStageImage().Image.GetStageDesc() != nil {
			return phase.StagesIterator.PrevNonEmptyStage.GetStageImage().Image.GetStageDesc().Info.Size
		}
	}
	return 0
}

func (phase *BuildPhase) getPrevNonEmptyStageCreationTs() int64 {
	if phase.StagesIterator.PrevNonEmptyStage != nil {
		if phase.StagesIterator.PrevNonEmptyStage.GetStageImage().Image.GetStageDesc() != nil {
			return phase.StagesIterator.PrevNonEmptyStage.GetStageImage().Image.GetStageDesc().StageID.CreationTs
		}
	}
	return 0
}

func (phase *BuildPhase) OnImageStage(ctx context.Context, img *image.Image, stg stage.Interface) error {
	return phase.StagesIterator.OnImageStage(ctx, img, stg, func(img *image.Image, stg stage.Interface, isEmpty bool) error {
		if isEmpty {
			return nil
		}
		if err := phase.onImageStage(ctx, img, stg); err != nil {
			return err
		}
		if err := phase.afterImageStage(ctx, img, stg); err != nil {
			return err
		}
		return nil
	})
}

func (phase *BuildPhase) onImageStage(ctx context.Context, img *image.Image, stg stage.Interface) error {
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

	promise := cleanup.NewPromise()
	defer promise.Give()

	var foundSuitableStage bool

	if err := logboek.Context(ctx).Info().LogProcess("Try to find suitable stage for %s", stg.LogDetailedName()).
		DoError(func() error {
			found, cleanupFunc, err := phase.calculateStage(ctx, img, stg)
			if err != nil {
				return err
			}
			promise.Add(cleanupFunc)
			foundSuitableStage = found
			return nil
		}); err != nil {
		return err
	}

	if foundSuitableStage {
		logboek.Context(ctx).Default().LogFHighlight("Use previously built image for %s\n", stg.LogDetailedName())
		container_backend.LogImageInfo(ctx, stg.GetStageImage().Image, phase.getPrevNonEmptyStageImageSize(), img.ShouldLogPlatform())

		logboek.Context(ctx).LogOptionalLn()

		if phase.IntrospectOptions.ImageStageShouldBeIntrospected(img.GetName(), string(stg.Name())) {
			if err := introspectStage(ctx, stg); err != nil {
				return err
			}
		}

		stg.SetMeta(&stage.StageMeta{
			Rebuilt: false,
		})

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

		start := time.Now()

		// Will build a new stage
		i := phase.Conveyor.GetOrCreateStageImage(uuid.New().String(), phase.StagesIterator.GetPrevImage(img, stg), stg, img)
		stg.SetStageImage(i)

		var fetchInfo fetchBaseImageForStageInfo
		if stg.IsBuildable() {
			info, err := phase.fetchBaseImageForStage(ctx, img, stg)
			if err != nil {
				return err
			}
			fetchInfo = info
		} else {
			fetchInfo = fetchBaseImageForStageInfo{
				BaseImagePulled: false,
				BaseImageSource: BaseImageSourceTypeRepo,
			}
		}

		if cleanupFunc, err := phase.prepareStageInstructions(ctx, img, stg); err != nil {
			return err
		} else {
			promise.Add(cleanupFunc)
		}

		if err := phase.buildStage(ctx, img, stg); err != nil {
			return err
		}
		duration := time.Since(start).Seconds()

		stg.SetMeta(&stage.StageMeta{
			Rebuilt:             true,
			BaseImagePulled:     fetchInfo.BaseImagePulled,
			BaseImageSourceType: fetchInfo.BaseImageSource,
			BuildTime:           fmt.Sprintf("%.2f", duration),
		})
	}

	// debug assertion
	if stg.GetStageImage().Image.GetStageDesc() == nil {
		panic(fmt.Sprintf("expected stage %s image %q built image info (image name = %s) to be set!", stg.Name(), img.GetName(), stg.GetStageImage().Image.Name()))
	}

	if foundSuitableSecondaryStage {
		stg.SetMeta(&stage.StageMeta{
			BaseImageSourceType: BaseImageSourceTypeSecondary,
		})
	}

	// Add managed image record only if there was at least one newly built stage
	if !phase.BuildOptions.SkipAddManagedImagesRecords {
		phase.Conveyor.SetShouldAddManagedImagesRecords()
	}

	return nil
}

func (phase *BuildPhase) afterImageStage(ctx context.Context, img *image.Image, stg stage.Interface) error {
	// TODO(staged-dockerfile): Expand possible ONBUILD instruction into specified intructions,
	// TODO(staged-dockerfile):  proxying ONBUILD instruction to chain of arbitrary instructions.

	if img.IsDockerfileImage && img.DockerfileImageConfig.Staged {
		if werf.GetStagedDockerfileVersion() == werf.StagedDockerfileV2 {
			if _, isFromStage := stg.(*instruction.From); isFromStage {
				if err := img.ExpandDependencies(ctx, image.EnvToMap(stg.GetStageImage().Image.GetStageDesc().Info.Env)); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func (phase *BuildPhase) findAndFetchStageFromSecondaryStagesStorage(ctx context.Context, img *image.Image, stg stage.Interface) (bool, error) {
	foundSuitableStage := false

	storageManager := phase.Conveyor.StorageManager
	atomicCopySuitableStageFromSecondaryStagesStorage := func(secondaryStageDesc *imagePkg.StageDesc, secondaryStagesStorage storage.StagesStorage) error {
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
			if stageDescCopy, err := storageManager.CopySuitableStageDescByDigest(ctx, secondaryStageDesc, secondaryStagesStorage, storageManager.GetStagesStorage(), phase.Conveyor.ContainerBackend, img.TargetPlatform); err != nil {
				return fmt.Errorf("unable to copy suitable stage %s from %s to %s: %w", secondaryStageDesc.StageID.String(), secondaryStagesStorage.String(), storageManager.GetStagesStorage().String(), err)
			} else {
				i := phase.Conveyor.GetOrCreateStageImage(stageDescCopy.Info.Name, phase.StagesIterator.GetPrevImage(img, stg), stg, img)
				i.Image.SetStageDesc(stageDescCopy)
				stg.SetStageImage(i)

				// The stage digest remains the same, but the content digest may differ (e.g., the content digest of git and some user stages depends on the git commit).
				contentDigest, exist := stageDescCopy.Info.Labels[imagePkg.WerfStageContentDigestLabel]
				if exist {
					stg.SetContentDigest(contentDigest)
				} else {
					panic(fmt.Sprintf("expected stage %q content digest label to be set!", stg.Name()))
				}

				logboek.Context(ctx).Default().LogFHighlight("Use previously built image for %s\n", stg.LogDetailedName())
				container_backend.LogImageInfo(ctx, stg.GetStageImage().Image, phase.getPrevNonEmptyStageImageSize(), img.ShouldLogPlatform())

				return nil
			}
		})
		if err != nil {
			return err
		}

		unlockStage()

		if err := storageManager.CopyStageIntoCacheStorages(
			ctx, *stg.GetStageImage().Image.GetStageDesc().StageID,
			storageManager.GetCacheStagesStorageList(),
			manager.CopyStageIntoStorageOptions{
				FetchStage:       stg,
				ContainerBackend: phase.Conveyor.ContainerBackend,
				LogDetailedName:  stg.LogDetailedName(),
			},
		); err != nil {
			return fmt.Errorf("unable to copy stage %s into cache storages: %w", stg.GetStageImage().Image.GetStageDesc().StageID.String(), err)
		}

		return nil
	}

ScanSecondaryStagesStorageList:
	for _, secondaryStagesStorage := range storageManager.GetSecondaryStagesStorageList() {
		secondaryStages, err := storageManager.GetStageDescSetByDigestFromStagesStorageWithCache(ctx, stg.LogDetailedName(), stg.GetDigest(), phase.getPrevNonEmptyStageCreationTs(), secondaryStagesStorage)
		if err != nil {
			return false, err
		} else {
			if secondaryStageDesc, err := storageManager.SelectSuitableStageDesc(ctx, phase.Conveyor, stg, secondaryStages); err != nil {
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

type fetchBaseImageForStageInfo struct {
	BaseImagePulled bool
	BaseImageSource string
}

func (phase *BuildPhase) fetchBaseImageForStage(ctx context.Context, img *image.Image, stg stage.Interface) (fetchBaseImageForStageInfo, error) {
	if stg.HasPrevStage() {
		info, err := phase.Conveyor.StorageManager.FetchStage(ctx, phase.Conveyor.ContainerBackend, phase.StagesIterator.PrevBuiltStage)
		return fetchBaseImageForStageInfo{
			BaseImagePulled: info.BaseImagePulled,
			BaseImageSource: info.BaseImageSource,
		}, err
	} else {
		info, err := img.FetchBaseImage(ctx)
		if err != nil {
			return fetchBaseImageForStageInfo{}, fmt.Errorf("unable to fetch base image %q for stage %s: %w", img.GetBaseStageImage().Image.Name(), stg.LogDetailedName(), err)
		}
		return fetchBaseImageForStageInfo{
			BaseImagePulled: info.BaseImagePulled,
			BaseImageSource: info.BaseImageSource,
		}, nil
	}
}

func (phase *BuildPhase) calculateStage(ctx context.Context, img *image.Image, stg stage.Interface) (bool, cleanup.Func, error) {
	// FIXME(stapel-to-buildah): store StageImage-s everywhere in stage and build pkgs
	stageDependencies, err := stg.GetDependencies(ctx, phase.Conveyor, phase.Conveyor.ContainerBackend, phase.StagesIterator.GetPrevImage(img, stg), phase.StagesIterator.GetPrevBuiltImage(img, stg), phase.buildContextArchive)
	if err != nil {
		return false, nil, err
	}

	var opts calculateDigestOptions
	opts.TargetPlatform = img.TargetPlatform

	if img.IsDockerfileImage && img.DockerfileImageConfig.Staged {
		if !stg.HasPrevStage() {
			// FIXME: For werf.StagedDockerfileV2, this logic should also be the default.
			// Currently, to avoid breaking tag reproducibility, this logic is only enabled for multi-stage cases.
			// Eventually, this behavior should be default for all versions without the extra if condition.
			if img.IsBasedOnStage() || werf.GetStagedDockerfileVersion() == werf.StagedDockerfileV1 {
				opts.BaseImage = img.GetBaseImageReference()
			}
		}
	}

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

	storageManager := phase.Conveyor.StorageManager
	stageDescSet, err := storageManager.GetStageDescSetByDigestWithCache(ctx, stg.LogDetailedName(), stageDigest, phase.getPrevNonEmptyStageCreationTs())
	if err != nil {
		return false, phase.Conveyor.GetStageDigestMutex(stg.GetDigest()).Unlock, err
	}

	stageDesc, err := storageManager.SelectSuitableStageDesc(ctx, phase.Conveyor, stg, stageDescSet)
	if err != nil {
		return false, phase.Conveyor.GetStageDigestMutex(stg.GetDigest()).Unlock, err
	}

	var stageContentSig string
	foundSuitableStage := false
	if stageDesc != nil {
		i := phase.Conveyor.GetOrCreateStageImage(stageDesc.Info.Name, phase.StagesIterator.GetPrevImage(img, stg), stg, img)
		i.Image.SetStageDesc(stageDesc)
		stg.SetStageImage(i)
		foundSuitableStage = true

		// The stage digest remains the same, but the content digest may differ (e.g., the content digest of git and some user stages depends on the git commit).
		contentDigest, exist := stageDesc.Info.Labels[imagePkg.WerfStageContentDigestLabel]
		if exist {
			stageContentSig = contentDigest
		} else {
			panic(fmt.Sprintf("expected stage %q content digest label to be set!", stg.Name()))
		}
	} else {
		stageContentSig, err = calculateDigest(ctx, fmt.Sprintf("%s-content", stg.Name()), "", stg, phase.Conveyor, calculateDigestOptions{TargetPlatform: img.TargetPlatform})
		if err != nil {
			return false, phase.Conveyor.GetStageDigestMutex(stg.GetDigest()).Unlock, fmt.Errorf("unable to calculate stage %s content digest: %w", stg.Name(), err)
		}
	}

	stg.SetContentDigest(stageContentSig)
	logboek.Context(ctx).Info().LogF("Stage %s content digest: %s\n", stg.LogDetailedName(), stageContentSig)

	return foundSuitableStage, phase.Conveyor.GetStageDigestMutex(stg.GetDigest()).Unlock, nil
}

func (phase *BuildPhase) prepareStageInstructions(ctx context.Context, img *image.Image, stg stage.Interface) (cleanup.Func, error) {
	logboek.Context(ctx).Debug().LogF("-- BuildPhase.prepareStage %s %s\n", img.LogDetailedName(), stg.LogDetailedName())

	stageImage := stg.GetStageImage()

	serviceLabels := map[string]string{
		imagePkg.WerfLabel:                   phase.Conveyor.ProjectName(),
		imagePkg.WerfVersionLabel:            werf.Version,
		imagePkg.WerfStageContentDigestLabel: stg.GetContentDigest(),
	}

	prevBuiltImage := phase.StagesIterator.GetPrevBuiltImage(img, stg)
	if stg.HasPrevStage() {
		if prevBuiltImage == nil {
			panic(fmt.Sprintf("expected prevBuiltImage to be set for stage %s", stg.Name()))
		}

		serviceLabels[imagePkg.WerfParentStageID] = prevBuiltImage.Image.GetStageDesc().StageID.String()

		// TODO: remove this legacy logic in v3.
		serviceLabels[imagePkg.WerfBaseImageIDLabel] = prevBuiltImage.Image.GetStageDesc().Info.ID
	} else if img.IsBasedOnStage() {
		baseStageImage := img.GetBaseStageImage()
		serviceLabels[imagePkg.WerfParentStageID] = baseStageImage.Image.GetStageDesc().StageID.String()
	}

	// TODO: refactor this workaround required for the image spec stage.
	stageImage.Image.SetBuildServiceLabels(serviceLabels)

	if stg.IsStapelStage() {
		if phase.Conveyor.UseLegacyStapelBuilder(phase.Conveyor.ContainerBackend) {
			stageImage.Builder.LegacyStapelStageBuilder().Container().ServiceCommitChangeOptions().AddLabel(serviceLabels)
		} else {
			stageImage.Builder.StapelStageBuilder().AddLabels(serviceLabels)
		}

		headHash, err := phase.Conveyor.GiterminismManager().LocalGitRepo().HeadCommitHash(ctx)
		if err != nil {
			return nil, fmt.Errorf("error getting HEAD commit hash: %w", err)
		}
		headTime, err := phase.Conveyor.GiterminismManager().LocalGitRepo().HeadCommitTime(ctx)
		if err != nil {
			return nil, fmt.Errorf("error getting HEAD commit time: %w", err)
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
	} else {
		stageImage.Builder.DockerfileStageBuilder().SetBuildContextArchive(phase.buildContextArchive)

		for k, v := range serviceLabels {
			stageImage.Builder.DockerfileStageBuilder().AppendPostInstruction(
				backend_instruction.NewLabel(*instructions.NewLabelCommand(k, v, true)),
			)
		}
	}

	promise := cleanup.NewPromise()
	defer promise.Give()

	promise.Add(func() {
		if err := stageImage.Builder.Cleanup(ctx); err != nil {
			logboek.Context(ctx).Warn().LogF("Error cleaning up stage: %s\n", err)
		}
	})

	if cleanupPrepareImageFunc, err := stg.PrepareImage(ctx, phase.Conveyor, phase.Conveyor.ContainerBackend, prevBuiltImage, stageImage, phase.buildContextArchive); err != nil {
		return nil, fmt.Errorf("error preparing stage %s: %w", stg.Name(), err)
	} else {
		promise.Add(cleanupPrepareImageFunc)
	}

	return promise.Forget(), nil
}

func (phase *BuildPhase) buildStage(ctx context.Context, img *image.Image, stg stage.Interface) error {
	if stg.IsBuildable() {
		if !img.IsDockerfileImage && phase.Conveyor.UseLegacyStapelBuilder(phase.Conveyor.ContainerBackend) {
			_, err := stapel.GetOrCreateContainer(ctx)
			if err != nil {
				return fmt.Errorf("get or create stapel container failed: %w", err)
			}
		}
	}

	infoSectionFunc := func(err error) {
		if err != nil {
			return
		}
		container_backend.LogImageInfo(ctx, stg.GetStageImage().Image, phase.getPrevNonEmptyStageImageSize(), img.ShouldLogPlatform())
	}

	if err := logboek.Context(ctx).Default().LogProcess("Building stage %s", stg.LogDetailedName()).
		Options(func(options types.LogProcessOptionsInterface) {
			options.InfoSectionFunc(infoSectionFunc)
			options.Style(style.Highlight())
		}).
		DoError(func() (err error) {
			if err := stg.PreRun(ctx, phase.Conveyor); err != nil {
				return fmt.Errorf("%s preRun failed: %w", stg.LogDetailedName(), err)
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

	if stg.IsBuildable() {
		if v := os.Getenv("WERF_TEST_ATOMIC_STAGE_BUILD__SLEEP_SECONDS_BEFORE_STAGE_BUILD"); v != "" {
			seconds := 0
			fmt.Sscanf(v, "%d", &seconds)
			fmt.Printf("Sleeping %d seconds before building new image by digest %s...\n", seconds, stg.GetDigest())
			time.Sleep(time.Duration(seconds) * time.Second)
		}

		if err := logboek.Context(ctx).Streams().DoErrorWithTag(fmt.Sprintf("%s/%s", img.LogName(), stg.Name()), img.LogTagStyle(), func() error {
			opts := phase.ImageBuildOptions
			opts.TargetPlatform = img.TargetPlatform
			if err := stageImage.Builder.Build(ctx, opts); err != nil {
				return fmt.Errorf("error building stage %s: %w", stg.Name(), err)
			}
			return nil
		}); err != nil {
			return fmt.Errorf("failed to build image for stage %s with digest %s: %w", stg.Name(), stg.GetDigest(), err)
		}

		if v := os.Getenv("WERF_TEST_ATOMIC_STAGE_BUILD__SLEEP_SECONDS_BEFORE_STAGE_SAVE"); v != "" {
			seconds := 0
			fmt.Sscanf(v, "%d", &seconds)
			fmt.Printf("Sleeping %d seconds before saving newly built image %s into repo %s by digest %s...\n", seconds, stg.GetStageImage().Image.BuiltID(), phase.Conveyor.StorageManager.GetStagesStorage().String(), stg.GetDigest())
			time.Sleep(time.Duration(seconds) * time.Second)
		}
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

	var stageDescSet imagePkg.StageDescSet
	if os.Getenv("WERF_DISABLE_PUBLISH_TAG_CACHE_SYNC") == "1" {
		stageDescSet = imagePkg.NewStageDescSet()
	} else {
		var err error
		stageDescSet, err = phase.Conveyor.StorageManager.GetStageDescSetByDigest(ctx, stg.LogDetailedName(), stg.GetDigest(), phase.getPrevNonEmptyStageCreationTs())
		if err != nil {
			return err
		}
		stageDesc, err := phase.Conveyor.StorageManager.SelectSuitableStageDesc(ctx, phase.Conveyor, stg, stageDescSet)
		if err != nil {
			return err
		}

		if stageDesc != nil {
			logboek.Context(ctx).Default().LogF(
				"Discarding newly built image for stage %s by digest %s: detected already existing image %s in the repo\n",
				stg.LogDetailedName(), stg.GetDigest(), stageDesc.Info.Name,
			)

			i := phase.Conveyor.GetOrCreateStageImage(stageDesc.Info.Name, phase.StagesIterator.GetPrevImage(img, stg), stg, img)
			i.Image.SetStageDesc(stageDesc)
			stg.SetStageImage(i)

			// The stage digest remains the same, but the content digest may differ (e.g., the content digest of git and some user stages depends on the git commit).
			contentDigest, exist := stageDesc.Info.Labels[imagePkg.WerfStageContentDigestLabel]
			if exist {
				stg.SetContentDigest(contentDigest)
			} else {
				panic(fmt.Sprintf("expected stage %q content digest label to be set!", stg.Name()))
			}
			return nil
		}
	}

	// use newly built image
	newStageImageName, stageCreationTs := phase.Conveyor.StorageManager.GenerateStageDescCreationTs(stg.GetDigest(), stageDescSet)
	phase.Conveyor.UnsetStageImage(stageImage.Image.Name())
	stageImage.Image.SetName(newStageImageName)
	phase.Conveyor.SetStageImage(stageImage)

	if err := logboek.Context(ctx).Default().LogProcess("Store stage into %s", phase.Conveyor.StorageManager.GetStagesStorage().String()).DoError(func() error {
		if stg.IsMutable() {
			switch phase.Conveyor.StorageManager.GetStagesStorage().(type) {
			case *storage.RepoStagesStorage:
			default:
				err := ErrMutableStageLocalStorage
				if stg.Name() == stage.ImageSpec {
					err = ErrMutableStageLocalStorageImageSpec
				}
				return fmt.Errorf("unable to build stage %q: %w", stg.Name(), err)
			}

			if err := stg.MutateImage(ctx, phase.Conveyor.StorageManager.GetStagesStorage().(*storage.RepoStagesStorage).DockerRegistry, phase.StagesIterator.PrevBuiltStage.GetStageImage(), stageImage); err != nil {
				return fmt.Errorf("unable to mutate %s: %w", stg.Name(), err)
			}
		} else {
			if err := phase.Conveyor.StorageManager.GetStagesStorage().StoreImage(ctx, stageImage.Image); err != nil {
				return fmt.Errorf("unable to store stage %s digest %s image %s into repo %s: %w", stg.LogDetailedName(), stg.GetDigest(), stageImage.Image.Name(), phase.Conveyor.StorageManager.GetStagesStorage().String(), err)
			}
		}

		desc, err := phase.Conveyor.StorageManager.GetStagesStorage().GetStageDesc(ctx, phase.Conveyor.ProjectName(), *imagePkg.NewStageID(stg.GetDigest(), stageCreationTs))
		if err != nil || desc == nil {
			return fmt.Errorf("unable to get stage %s digest %s image %s description from repo %s after stages has been stored into repo: %w", stg.LogDetailedName(), stg.GetDigest(), stageImage.Image.Name(), phase.Conveyor.StorageManager.GetStagesStorage().String(), err)
		}

		stageImage.Image.SetStageDesc(desc)
		img.SetRebuilt(true)

		return nil
	}); err != nil {
		return err
	}

	unlockStage()

	if err := phase.Conveyor.StorageManager.CopyStageIntoCacheStorages(
		ctx, *stg.GetStageImage().Image.GetStageDesc().StageID,
		phase.Conveyor.StorageManager.GetCacheStagesStorageList(),
		manager.CopyStageIntoStorageOptions{
			FetchStage:       stg,
			ContainerBackend: phase.Conveyor.ContainerBackend,
			LogDetailedName:  stg.LogDetailedName(),
		},
	); err != nil {
		return fmt.Errorf("unable to copy stage %s into cache storages: %w", stageImage.Image.GetStageDesc().StageID.String(), err)
	}
	return nil
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

type calculateDigestOptions struct {
	TargetPlatform string
	BaseImage      string // TODO(staged-dockerfile): legacy compatibility field
}

func calculateDigest(ctx context.Context, stageName, stageDependencies string, prevNonEmptyStage stage.Interface, conveyor *Conveyor, opts calculateDigestOptions) (string, error) {
	var checksumArgs []string
	var checksumArgsNames []string

	// TODO: linux/amd64 not affects digest for compatibility with currently built stages.
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

			logboek.Context(ctx).Warn().LogLn(`To quickly find the problem compare current and previous rendered werf configurations.
Get the path at the beginning of command output by the following prefix 'Using werf config render file: '.
E.g.:

  diff /tmp/werf-config-render-502883762 /tmp/werf-config-render-837625028`)
			logboek.Context(ctx).Warn().LogLn()

			logboek.Context(ctx).Warn().LogLn(reasonNumberFunc() + `Stages have not been built yet or stages have been removed:
- automatically with werf cleanup command
- manually with werf purge or werf host purge commands`)
			logboek.Context(ctx).Warn().LogLn()
			logboek.Context(ctx).Warn().LogLn(reasonNumberFunc() + `You are using --require-built-images flag (or WERF_REQUIRE_BUILT_IMAGES env) which requires images to be already built:
- If you expect images to be built and available in the registry, check the reasons above
- If you want to build images instead of requiring them to be already built, remove --require-built-images flag / WERF_REQUIRE_BUILT_IMAGES env`)
			logboek.Context(ctx).Warn().LogLn()
		})
}

func (phase *BuildPhase) Clone() Phase {
	u := *phase
	return &u
}
