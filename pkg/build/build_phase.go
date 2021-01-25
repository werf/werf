package build

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"

	"github.com/docker/docker/pkg/stringid"

	"github.com/werf/logboek"
	"github.com/werf/logboek/pkg/style"
	"github.com/werf/logboek/pkg/types"

	"github.com/werf/werf/pkg/build/stage"
	"github.com/werf/werf/pkg/container_runtime"
	"github.com/werf/werf/pkg/image"
	imagePkg "github.com/werf/werf/pkg/image"
	"github.com/werf/werf/pkg/stapel"
	"github.com/werf/werf/pkg/storage"
	"github.com/werf/werf/pkg/util"
	"github.com/werf/werf/pkg/werf"
)

type BuildPhaseOptions struct {
	BuildOptions
	ShouldBeBuiltMode bool
}

type BuildOptions struct {
	ImageBuildOptions container_runtime.BuildOptions
	IntrospectOptions

	ReportPath   string
	ReportFormat ReportFormat

	DryRun bool
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
	WerfImageName   string
	DockerRepo      string
	DockerTag       string
	DockerImageID   string
	DockerImageName string
}

func (phase *BuildPhase) Name() string {
	return "build"
}

func (phase *BuildPhase) BeforeImages(_ context.Context) error {
	return nil
}

func (phase *BuildPhase) AfterImages(ctx context.Context) error {
	return phase.createReport(ctx)
}

func (phase *BuildPhase) createReport(ctx context.Context) error {
	for _, img := range phase.Conveyor.images {
		if img.isArtifact {
			continue
		}

		desc := img.GetLastNonEmptyStage().GetImage().GetStageDescription()
		phase.ImagesReport.SetImageRecord(img.GetName(), ReportImageRecord{
			WerfImageName:   img.GetName(),
			DockerRepo:      desc.Info.Repository,
			DockerTag:       desc.Info.Tag,
			DockerImageID:   desc.Info.ID,
			DockerImageName: desc.Info.Name,
		})
	}

	debugJsonData, err := phase.ImagesReport.ToJsonData()
	logboek.Context(ctx).Debug().LogF("ImagesReport: (err: %s)\n%s", err, debugJsonData)

	if phase.ReportPath != "" {
		var data []byte
		var err error
		switch phase.ReportFormat {
		case ReportJSON:
			if data, err = phase.ImagesReport.ToJsonData(); err != nil {
				return fmt.Errorf("unable to prepare report json: %s", err)
			}
			logboek.Context(ctx).Debug().LogF("Writing json report to the %q:\n%s", phase.ReportPath, data)
		case ReportEnvFile:
			data = phase.ImagesReport.ToEnvFileData()
			logboek.Context(ctx).Debug().LogF("Writing envfile report to the %q:\n%s", phase.ReportPath, data)
		default:
			panic(fmt.Sprintf("unknown report format %q", phase.ReportFormat))
		}

		if err := ioutil.WriteFile(phase.ReportPath, data, 0644); err != nil {
			return fmt.Errorf("unable to write report to %s: %s", phase.ReportPath, err)
		}
	}

	return nil
}

func (phase *BuildPhase) ImageProcessingShouldBeStopped(_ context.Context, img *Image) bool {
	return false
}

func (phase *BuildPhase) BeforeImageStages(_ context.Context, img *Image) error {
	phase.StagesIterator = NewStagesIterator(phase.Conveyor)

	img.SetupBaseImage(phase.Conveyor)

	return nil
}

func (phase *BuildPhase) AfterImageStages(ctx context.Context, img *Image) error {
	img.SetLastNonEmptyStage(phase.StagesIterator.PrevNonEmptyStage)
	img.SetContentDigest(phase.StagesIterator.PrevNonEmptyStage.GetContentDigest())

	if img.isArtifact {
		return nil
	}

	if err := phase.addManagedImage(ctx, img); err != nil {
		return err
	}

	if err := phase.publishImageMetadata(ctx, img); err != nil {
		return err
	}

	return nil
}

func (phase *BuildPhase) addManagedImage(ctx context.Context, img *Image) error {
	if phase.ShouldAddManagedImageRecord {
		if err := phase.Conveyor.StorageManager.StagesStorage.AddManagedImage(ctx, phase.Conveyor.projectName(), img.GetName()); err != nil {
			return fmt.Errorf("unable to add image %q to the managed images of project %q: %s", img.GetName(), phase.Conveyor.projectName(), err)
		}
	}

	return nil
}

func (phase *BuildPhase) publishImageMetadata(ctx context.Context, img *Image) error {
	return logboek.Context(ctx).Info().LogProcess(fmt.Sprintf("Processing image %s git metadata", img.GetName())).
		DoError(func() error {
			headCommit := phase.Conveyor.giterminismManager.HeadCommit()

			exists, err := phase.Conveyor.StorageManager.StagesStorage.IsImageMetadataExist(ctx, phase.Conveyor.projectName(), img.GetName(), headCommit, img.GetStageID())
			if err != nil {
				return fmt.Errorf("unable to get image %s metadata by commit %s and stage ID %s: %s", img.GetName(), headCommit, img.GetStageID(), err)
			}

			if !exists {
				return phase.Conveyor.StorageManager.StagesStorage.PutImageMetadata(ctx, phase.Conveyor.projectName(), img.GetName(), headCommit, img.GetStageID())
			}

			return nil
		})
}

func (phase *BuildPhase) getPrevNonEmptyStageImageSize() int64 {
	if phase.StagesIterator.PrevNonEmptyStage != nil {
		if phase.StagesIterator.PrevNonEmptyStage.GetImage().GetStageDescription() != nil {
			return phase.StagesIterator.PrevNonEmptyStage.GetImage().GetStageDescription().Info.Size
		}
	}
	return 0
}

func (phase *BuildPhase) OnImageStage(ctx context.Context, img *Image, stg stage.Interface) error {
	return phase.StagesIterator.OnImageStage(ctx, img, stg, func(img *Image, stg stage.Interface, isEmpty bool) error {
		return phase.onImageStage(ctx, img, stg, isEmpty)
	})
}

func (phase *BuildPhase) onImageStage(ctx context.Context, img *Image, stg stage.Interface, isEmpty bool) error {
	if isEmpty {
		return nil
	}

	if err := stg.FetchDependencies(ctx, phase.Conveyor, phase.Conveyor.ContainerRuntime); err != nil {
		return fmt.Errorf("unable to fetch dependencies for stage %s: %s", stg.LogDetailedName(), err)
	}

	if stg.Name() != "from" && stg.Name() != "dockerfile" {
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

	// Stage is cached in the stages storage
	if foundSuitableStage {
		logboek.Context(ctx).Default().LogFHighlight("Use cache image for %s\n", stg.LogDetailedName())
		logImageInfo(ctx, stg.GetImage(), phase.getPrevNonEmptyStageImageSize(), true)

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
		i := phase.Conveyor.GetOrCreateStageImage(castToStageImage(phase.StagesIterator.GetPrevImage(img, stg)), uuid.New().String())
		stg.SetImage(i)

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

	if stg.GetImage().GetStageDescription() == nil {
		panic(fmt.Sprintf("expected stage %s image %q built image info (image name = %s) to be set!", stg.Name(), img.GetName(), stg.GetImage().Name()))
	}

	// Add managed image record only if there was at least one newly built stage
	phase.ShouldAddManagedImageRecord = true

	return nil
}

func (phase *BuildPhase) findAndFetchStageFromSecondaryStagesStorage(ctx context.Context, img *Image, stg stage.Interface) (bool, error) {
	foundSuitableStage := false

	atomicCopySuitableStageFromSecondaryStagesStorage := func(secondaryStageDesc *image.StageDescription, secondaryStagesStorage storage.StagesStorage) error {
		// Lock the primary stages storage
		if lock, err := phase.Conveyor.StorageLockManager.LockStage(ctx, phase.Conveyor.projectName(), stg.GetDigest()); err != nil {
			return fmt.Errorf("unable to lock project %s digest %s: %s", phase.Conveyor.projectName(), stg.GetDigest(), err)
		} else {
			defer phase.Conveyor.StorageLockManager.Unlock(ctx, lock)
		}

		// Query the primary stages storage for suitable stage again.
		// Suitable stage can be found this time and should be used in this case
		if stages, err := phase.Conveyor.StorageManager.GetStagesByDigest(ctx, stg.LogDetailedName(), stg.GetDigest()); err != nil {
			return err
		} else {
			if stageDesc, err := phase.Conveyor.StorageManager.SelectSuitableStage(ctx, phase.Conveyor, stg, stages); err != nil {
				return err
			} else if stageDesc != nil {
				i := phase.Conveyor.GetOrCreateStageImage(castToStageImage(phase.StagesIterator.GetPrevImage(img, stg)), stageDesc.Info.Name)
				i.SetStageDescription(stageDesc)
				stg.SetImage(i)

				logboek.Context(ctx).Default().LogFHighlight("Use cache image for %s\n", stg.LogDetailedName())
				logImageInfo(ctx, stg.GetImage(), phase.getPrevNonEmptyStageImageSize(), true)

				return nil
			}

			return logboek.Context(ctx).Default().LogProcess("Copy suitable stage from %s", secondaryStagesStorage.String()).DoError(func() error {
				// Copy suitable stage from a secondary stages storage to the primary stages storage
				// while primary stages storage lock for this digest is held
				if copiedStageDesc, err := phase.Conveyor.StorageManager.CopySuitableByDigestStage(ctx, secondaryStageDesc, secondaryStagesStorage, phase.Conveyor.StorageManager.StagesStorage, phase.Conveyor.ContainerRuntime); err != nil {
					return fmt.Errorf("unable to copy suitable stage %s from %s to %s: %s", secondaryStageDesc.StageID.String(), secondaryStagesStorage.String(), phase.Conveyor.StorageManager.StagesStorage.String(), err)
				} else {
					i := phase.Conveyor.GetOrCreateStageImage(castToStageImage(phase.StagesIterator.GetPrevImage(img, stg)), copiedStageDesc.Info.Name)
					i.SetStageDescription(copiedStageDesc)
					stg.SetImage(i)

					var stageIDs []image.StageID
					for _, stageDesc := range stages {
						stageIDs = append(stageIDs, *stageDesc.StageID)
					}
					stageIDs = append(stageIDs, *copiedStageDesc.StageID)

					if err := phase.Conveyor.StorageManager.AtomicStoreStagesByDigestToCache(ctx, string(stg.Name()), stg.GetDigest(), stageIDs); err != nil {
						return err
					}

					logboek.Context(ctx).Default().LogFHighlight("Use cache image for %s\n", stg.LogDetailedName())
					logImageInfo(ctx, stg.GetImage(), phase.getPrevNonEmptyStageImageSize(), true)

					return nil
				}
			})
		}
	}

ScanSecondaryStagesStorageList:
	for _, secondaryStagesStorage := range phase.Conveyor.StorageManager.SecondaryStagesStorageList {
		if secondaryStages, err := phase.Conveyor.StorageManager.GetStagesByDigestFromStagesStorage(ctx, stg.LogDetailedName(), stg.GetDigest(), secondaryStagesStorage); err != nil {
			return false, err
		} else {
			if secondaryStageDesc, err := phase.Conveyor.StorageManager.SelectSuitableStage(ctx, phase.Conveyor, stg, secondaryStages); err != nil {
				return false, err
			} else if secondaryStageDesc != nil {
				if err := atomicCopySuitableStageFromSecondaryStagesStorage(secondaryStageDesc, secondaryStagesStorage); err != nil {
					return false, fmt.Errorf("unable to copy suitable stage %s from secondary stages storage %s: %s", secondaryStageDesc.StageID.String(), secondaryStagesStorage.String(), err)
				}
				foundSuitableStage = true
				break ScanSecondaryStagesStorageList
			}
		}
	}

	return foundSuitableStage, nil
}

func (phase *BuildPhase) fetchBaseImageForStage(ctx context.Context, img *Image, stg stage.Interface) error {
	if stg.Name() == "from" {
		if err := img.FetchBaseImage(ctx, phase.Conveyor); err != nil {
			return fmt.Errorf("unable to fetch base image %s for stage %s: %s", img.GetBaseImage().Name(), stg.LogDetailedName(), err)
		}
	} else if stg.Name() == "dockerfile" {
		return nil
	} else {
		return phase.Conveyor.StorageManager.FetchStage(ctx, phase.StagesIterator.PrevBuiltStage)
	}

	return nil
}

func castToStageImage(img container_runtime.ImageInterface) *container_runtime.StageImage {
	if img == nil {
		return nil
	}
	return img.(*container_runtime.StageImage)
}

func (phase *BuildPhase) calculateStage(ctx context.Context, img *Image, stg stage.Interface) (bool, func(), error) {
	stageDependencies, err := stg.GetDependencies(ctx, phase.Conveyor, phase.StagesIterator.GetPrevImage(img, stg), phase.StagesIterator.GetPrevBuiltImage(img, stg))
	if err != nil {
		return false, nil, err
	}

	stageDigest, err := calculateDigest(ctx, string(stg.Name()), stageDependencies, phase.StagesIterator.PrevNonEmptyStage, phase.Conveyor)
	if err != nil {
		return false, nil, err
	}
	stg.SetDigest(stageDigest)

	logboek.Context(ctx).Info().LogProcessInline("Locking stage %s handling", stg.LogDetailedName()).
		Options(func(options types.LogProcessInlineOptionsInterface) {
			if !phase.Conveyor.Parallel {
				options.Mute()
			}
		}).
		Do(phase.Conveyor.GetStageDigestMutex(stg.GetDigest()).Lock)

	foundSuitableStage := false
	if stages, err := phase.Conveyor.StorageManager.GetStagesByDigest(ctx, stg.LogDetailedName(), stageDigest); err != nil {
		return false, phase.Conveyor.GetStageDigestMutex(stg.GetDigest()).Unlock, err
	} else {
		if stageDesc, err := phase.Conveyor.StorageManager.SelectSuitableStage(ctx, phase.Conveyor, stg, stages); err != nil {
			return false, phase.Conveyor.GetStageDigestMutex(stg.GetDigest()).Unlock, err
		} else if stageDesc != nil {
			i := phase.Conveyor.GetOrCreateStageImage(castToStageImage(phase.StagesIterator.GetPrevImage(img, stg)), stageDesc.Info.Name)
			i.SetStageDescription(stageDesc)
			stg.SetImage(i)
			foundSuitableStage = true
		}
	}

	stageContentSig, err := calculateDigest(ctx, fmt.Sprintf("%s-content", stg.Name()), "", stg, phase.Conveyor)
	if err != nil {
		return false, phase.Conveyor.GetStageDigestMutex(stg.GetDigest()).Unlock, fmt.Errorf("unable to calculate stage %s content digest: %s", stg.Name(), err)
	}
	stg.SetContentDigest(stageContentSig)

	logboek.Context(ctx).Info().LogF("Stage %s content digest: %s\n", stg.LogDetailedName(), stageContentSig)

	return foundSuitableStage, phase.Conveyor.GetStageDigestMutex(stg.GetDigest()).Unlock, nil
}

func (phase *BuildPhase) prepareStageInstructions(ctx context.Context, img *Image, stg stage.Interface) error {
	logboek.Context(ctx).Debug().LogF("-- BuildPhase.prepareStage %s %s\n", img.LogDetailedName(), stg.LogDetailedName())

	stageImage := stg.GetImage()

	serviceLabels := map[string]string{
		imagePkg.WerfDockerImageName:         stageImage.Name(),
		imagePkg.WerfLabel:                   phase.Conveyor.projectName(),
		imagePkg.WerfVersionLabel:            werf.Version,
		imagePkg.WerfCacheVersionLabel:       imagePkg.BuildCacheVersion,
		imagePkg.WerfImageLabel:              "false",
		imagePkg.WerfStageDigestLabel:        stg.GetDigest(),
		imagePkg.WerfStageContentDigestLabel: stg.GetContentDigest(),
	}

	switch stg.(type) {
	case *stage.DockerfileStage:
		var buildArgs []string

		for key, value := range serviceLabels {
			buildArgs = append(buildArgs, fmt.Sprintf("--label=%s=%s", key, value))
		}

		stageImage.DockerfileImageBuilder().AppendBuildArgs(buildArgs...)

		phase.Conveyor.AppendOnTerminateFunc(func() error {
			return stageImage.DockerfileImageBuilder().Cleanup(ctx)
		})

	default:
		imageServiceCommitChangeOptions := stageImage.Container().ServiceCommitChangeOptions()
		imageServiceCommitChangeOptions.AddLabel(serviceLabels)

		if phase.Conveyor.sshAuthSock != "" {
			imageRunOptions := stageImage.Container().RunOptions()

			if runtime.GOOS == "darwin" {
				imageRunOptions.AddVolume("/run/host-services/ssh-auth.sock:/run/host-services/ssh-auth.sock")
				imageRunOptions.AddEnv(map[string]string{"SSH_AUTH_SOCK": "/run/host-services/ssh-auth.sock"})
			} else {
				imageRunOptions.AddVolume(fmt.Sprintf("%s:/.werf/tmp/ssh-auth-sock", phase.Conveyor.sshAuthSock))
				imageRunOptions.AddEnv(map[string]string{"SSH_AUTH_SOCK": "/.werf/tmp/ssh-auth-sock"})
			}
		}
	}

	err := stg.PrepareImage(ctx, phase.Conveyor, phase.StagesIterator.GetPrevBuiltImage(img, stg), stageImage)
	if err != nil {
		return fmt.Errorf("error preparing stage %s: %s", stg.Name(), err)
	}

	return nil
}

func (phase *BuildPhase) buildStage(ctx context.Context, img *Image, stg stage.Interface) error {
	if !img.isDockerfileImage {
		_, err := stapel.GetOrCreateContainer(ctx)
		if err != nil {
			return fmt.Errorf("get or create stapel container failed: %s", err)
		}
	}

	infoSectionFunc := func(err error) {
		if err != nil {
			logboek.Context(ctx).Streams().DoWithIndent(func() {
				logImageCommands(ctx, stg.GetImage())
			})
			return
		}
		logImageInfo(ctx, stg.GetImage(), phase.getPrevNonEmptyStageImageSize(), false)
	}

	if err := logboek.Context(ctx).Default().LogProcess("Building stage %s", stg.LogDetailedName()).
		Options(func(options types.LogProcessOptionsInterface) {
			options.InfoSectionFunc(infoSectionFunc)
			options.Style(style.Highlight())
		}).
		DoError(func() (err error) {
			if err := stg.PreRunHook(ctx, phase.Conveyor); err != nil {
				return fmt.Errorf("%s preRunHook failed: %s", stg.LogDetailedName(), err)
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

func (phase *BuildPhase) atomicBuildStageImage(ctx context.Context, img *Image, stg stage.Interface) error {
	stageImage := stg.GetImage()

	if v := os.Getenv("WERF_TEST_ATOMIC_STAGE_BUILD__SLEEP_SECONDS_BEFORE_STAGE_BUILD"); v != "" {
		seconds := 0
		fmt.Sscanf(v, "%d", &seconds)
		fmt.Printf("Sleeping %d seconds before building new image by digest %s...\n", seconds, stg.GetDigest())
		time.Sleep(time.Duration(seconds) * time.Second)
	}

	if err := logboek.Context(ctx).Streams().DoErrorWithTag(fmt.Sprintf("%s/%s", img.LogName(), stg.Name()), img.LogTagStyle(), func() error {
		return stageImage.Build(ctx, phase.ImageBuildOptions)
	}); err != nil {
		return fmt.Errorf("failed to build image for stage %s with digest %s: %s", stg.Name(), stg.GetDigest(), err)
	}

	if v := os.Getenv("WERF_TEST_ATOMIC_STAGE_BUILD__SLEEP_SECONDS_BEFORE_STAGE_SAVE"); v != "" {
		seconds := 0
		fmt.Sscanf(v, "%d", &seconds)
		fmt.Printf("Sleeping %d seconds before saving newly built image %s into repo %s by digest %s...\n", seconds, stg.GetImage().GetBuiltId(), phase.Conveyor.StorageManager.StagesStorage.String(), stg.GetDigest())
		time.Sleep(time.Duration(seconds) * time.Second)
	}

	if lock, err := phase.Conveyor.StorageLockManager.LockStage(ctx, phase.Conveyor.projectName(), stg.GetDigest()); err != nil {
		return fmt.Errorf("unable to lock project %s digest %s: %s", phase.Conveyor.projectName(), stg.GetDigest(), err)
	} else {
		defer phase.Conveyor.StorageLockManager.Unlock(ctx, lock)
	}

	if stages, err := phase.Conveyor.StorageManager.GetStagesByDigest(ctx, stg.LogDetailedName(), stg.GetDigest()); err != nil {
		return err
	} else {
		if stageDesc, err := phase.Conveyor.StorageManager.SelectSuitableStage(ctx, phase.Conveyor, stg, stages); err != nil {
			return err
		} else if stageDesc != nil {
			logboek.Context(ctx).Default().LogF(
				"Discarding newly built image for stage %s by digest %s: detected already existing image %s in the repo\n",
				stg.LogDetailedName(), stg.GetDigest(), stageDesc.Info.Name,
			)

			i := phase.Conveyor.GetOrCreateStageImage(castToStageImage(phase.StagesIterator.GetPrevImage(img, stg)), stageDesc.Info.Name)
			i.SetStageDescription(stageDesc)
			stg.SetImage(i)
			return nil
		} else { // use newly built image
			newStageImageName, uniqueID := phase.Conveyor.StorageManager.GenerateStageUniqueID(stg.GetDigest(), stages)
			stageImageObj := phase.Conveyor.GetStageImage(stageImage.Name())
			phase.Conveyor.UnsetStageImage(stageImageObj.Name())
			stageImageObj.SetName(newStageImageName)
			phase.Conveyor.SetStageImage(stageImageObj)

			if err := logboek.Context(ctx).Info().LogProcess("Store stage").DoError(func() error {
				if err := phase.Conveyor.StorageManager.StagesStorage.StoreImage(ctx, &container_runtime.DockerImage{Image: stageImage}); err != nil {
					return fmt.Errorf("unable to store stage %s digest %s image %s into repo %s: %s", stg.LogDetailedName(), stg.GetDigest(), stageImage.Name(), phase.Conveyor.StorageManager.StagesStorage.String(), err)
				}
				if desc, err := phase.Conveyor.StorageManager.StagesStorage.GetStageDescription(ctx, phase.Conveyor.projectName(), stg.GetDigest(), uniqueID); err != nil {
					return fmt.Errorf("unable to get stage %s digest %s image %s description from repo %s after stages has been stored into repo: %s", stg.LogDetailedName(), stg.GetDigest(), stageImage.Name(), phase.Conveyor.StorageManager.StagesStorage.String(), err)
				} else {
					stageImageObj.SetStageDescription(desc)
				}
				return nil
			}); err != nil {
				return err
			}

			var stageIDs []image.StageID
			for _, stageDesc := range stages {
				stageIDs = append(stageIDs, *stageDesc.StageID)
			}
			stageIDs = append(stageIDs, *stageImage.GetStageDescription().StageID)

			return phase.Conveyor.StorageManager.AtomicStoreStagesByDigestToCache(ctx, string(stg.Name()), stg.GetDigest(), stageIDs)
		}
	}
}

func introspectStage(ctx context.Context, s stage.Interface) error {
	return logboek.Context(ctx).Info().LogProcess("Introspecting stage %s", s.Name()).
		Options(func(options types.LogProcessOptionsInterface) {
			options.Style(style.Highlight())
		}).
		DoError(func() error {
			if err := logboek.Context(ctx).Streams().DoErrorWithoutProxyStreamDataFormatting(func() error {
				return s.GetImage().Introspect(ctx)
			}); err != nil {
				return fmt.Errorf("introspect error failed: %s", err)
			}

			return nil
		})
}

var (
	logImageInfoLeftPartWidth = 8
	logImageInfoFormat        = fmt.Sprintf("  %%%ds: %%s\n", logImageInfoLeftPartWidth)
)

func logImageInfo(ctx context.Context, img container_runtime.ImageInterface, prevStageImageSize int64, isUsingCache bool) {
	repository, tag := image.ParseRepositoryAndTag(img.Name())
	logboek.Context(ctx).Default().LogFDetails(logImageInfoFormat, "name", fmt.Sprintf("%s:%s", repository, tag))
	logboek.Context(ctx).Info().LogFDetails(logImageInfoFormat, "image_id", stringid.TruncateID(img.GetStageDescription().Info.ID))
	logboek.Context(ctx).Info().LogFDetails(logImageInfoFormat, "created", img.GetStageDescription().Info.GetCreatedAt())

	if prevStageImageSize == 0 {
		logboek.Context(ctx).Default().LogFDetails(logImageInfoFormat, "size", byteCountBinary(img.GetStageDescription().Info.Size))
	} else {
		logboek.Context(ctx).Default().LogFDetails(logImageInfoFormat, "size", fmt.Sprintf("%s (+%s)", byteCountBinary(img.GetStageDescription().Info.Size), byteCountBinary(img.GetStageDescription().Info.Size-prevStageImageSize)))
	}

	if !isUsingCache {
		changes := img.Container().UserCommitChanges()
		if len(changes) != 0 {
			fitTextOptions := types.FitTextOptions{ExtraIndentWidth: logImageInfoLeftPartWidth + 4}
			formattedCommands := strings.TrimLeft(logboek.Context(ctx).FitText(strings.Join(changes, "\n"), fitTextOptions), " ")
			logboek.Context(ctx).Default().LogFDetails(logImageInfoFormat, "instructions", formattedCommands)
		}

		logImageCommands(ctx, img)
	}
}

func logImageCommands(ctx context.Context, img container_runtime.ImageInterface) {
	commands := img.Container().UserRunCommands()
	if len(commands) != 0 {
		fitTextOptions := types.FitTextOptions{ExtraIndentWidth: logImageInfoLeftPartWidth + 4}
		formattedCommands := strings.TrimLeft(logboek.Context(ctx).FitText(strings.Join(commands, "\n"), fitTextOptions), " ")
		logboek.Context(ctx).Default().LogFDetails(logImageInfoFormat, "commands", formattedCommands)
	}
}

func byteCountBinary(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %ciB", float64(b)/float64(div), "KMGTPE"[exp])
}

func calculateDigest(ctx context.Context, stageName, stageDependencies string, prevNonEmptyStage stage.Interface, conveyor *Conveyor) (string, error) {
	checksumArgs := []string{image.BuildCacheVersion, stageName, stageDependencies}
	if prevNonEmptyStage != nil {
		prevStageDependencies, err := prevNonEmptyStage.GetNextStageDependencies(ctx, conveyor)
		if err != nil {
			return "", fmt.Errorf("unable to get prev stage %s dependencies for the stage %s: %s", prevNonEmptyStage.Name(), stageName, err)
		}

		checksumArgs = append(checksumArgs, prevNonEmptyStage.GetDigest(), prevStageDependencies)
	}

	digest := util.Sha3_224Hash(checksumArgs...)

	blockMsg := fmt.Sprintf("Stage %s digest %s", stageName, digest)
	logboek.Context(ctx).Debug().LogBlock(blockMsg).Do(func() {
		checksumArgsNames := []string{
			"BuildCacheVersion",
			"stageName",
			"stageDependencies",
			"prevNonEmptyStage digest",
			"prevNonEmptyStage dependencies for next stage",
		}
		for ind, checksumArg := range checksumArgs {
			logboek.Context(ctx).Debug().LogF("%s => %q\n", checksumArgsNames[ind], checksumArg)
		}
	})

	return digest, nil
}

// TODO: move these prints to the after-images hook, print summary over all images
func (phase *BuildPhase) printShouldBeBuiltError(ctx context.Context, img *Image, stg stage.Interface) {
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

			if img.isDockerfileImage {
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
