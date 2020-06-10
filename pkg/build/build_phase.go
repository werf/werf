package build

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/werf/werf/pkg/container_runtime"
	"github.com/werf/werf/pkg/util"

	"github.com/docker/docker/pkg/stringid"

	"github.com/flant/logboek"

	"github.com/werf/werf/pkg/build/stage"
	"github.com/werf/werf/pkg/image"
	imagePkg "github.com/werf/werf/pkg/image"
	"github.com/werf/werf/pkg/stapel"
	"github.com/werf/werf/pkg/werf"
)

type BuildPhaseOptions struct {
	ShouldBeBuiltMode bool
	ImageBuildOptions container_runtime.BuildOptions
	IntrospectOptions IntrospectOptions
}

type BuildStagesOptions struct {
	ImageBuildOptions container_runtime.BuildOptions
	IntrospectOptions
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
	}
}

type BuildPhase struct {
	BasePhase
	BuildPhaseOptions

	StagesIterator              *StagesIterator
	ShouldAddManagedImageRecord bool
}

func (phase *BuildPhase) Name() string {
	return "build"
}

func (phase *BuildPhase) BeforeImages() error {
	return nil
}

func (phase *BuildPhase) AfterImages() error {
	return nil
}

func (phase *BuildPhase) ImageProcessingShouldBeStopped(img *Image) bool {
	return false
}

func (phase *BuildPhase) BeforeImageStages(img *Image) error {
	phase.StagesIterator = NewStagesIterator(phase.Conveyor)

	img.SetupBaseImage(phase.Conveyor)

	return nil
}

func (phase *BuildPhase) AfterImageStages(img *Image) error {
	img.SetLastNonEmptyStage(phase.StagesIterator.PrevNonEmptyStage)

	stagesSig, err := calculateSignature("imageStages", "", phase.StagesIterator.PrevNonEmptyStage, phase.Conveyor)
	if err != nil {
		return fmt.Errorf("unable to calculate image %s stages-signature: %s", img.GetName(), err)
	}
	img.SetContentSignature(stagesSig)

	if phase.ShouldAddManagedImageRecord {
		if err := phase.Conveyor.StagesManager.StagesStorage.AddManagedImage(phase.Conveyor.projectName(), img.GetName()); err != nil {
			return fmt.Errorf("unable to add image %q to the managed images of project %q: %s", img.GetName(), phase.Conveyor.projectName(), err)
		}
	}

	return nil
}

func (phase *BuildPhase) getPrevNonEmptyStageImageSize() int64 {
	if phase.StagesIterator.PrevNonEmptyStage != nil {
		if phase.StagesIterator.PrevNonEmptyStage.GetImage().GetStageDescription() != nil {
			return phase.StagesIterator.PrevNonEmptyStage.GetImage().GetStageDescription().Info.Size
		}
	}
	return 0
}

func (phase *BuildPhase) OnImageStage(img *Image, stg stage.Interface) error {
	return phase.StagesIterator.OnImageStage(img, stg, func(img *Image, stg stage.Interface, isEmpty bool) error {
		return phase.onImageStage(img, stg, isEmpty)
	})
}

func (phase *BuildPhase) onImageStage(img *Image, stg stage.Interface, isEmpty bool) error {
	if isEmpty {
		return nil
	}

	if phase.ShouldBeBuiltMode {
		return phase.calculateStage(img, stg, true)
	} else {
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

		if err := phase.calculateStage(img, stg, false); err != nil {
			return err
		}

		// Stage is cached in the stages storage
		if stg.GetImage().GetStageDescription() != nil {
			logboek.Default.LogFHighlight("Use cache image for %s\n", stg.LogDetailedName())

			logImageInfo(stg.GetImage(), phase.getPrevNonEmptyStageImageSize(), true)

			logboek.LogOptionalLn()

			if phase.IntrospectOptions.ImageStageShouldBeIntrospected(img.GetName(), string(stg.Name())) {
				if err := introspectStage(stg); err != nil {
					return err
				}
			}

			return nil
		}

		if err := phase.fetchBaseImageForStage(img, stg); err != nil {
			return err
		}

		if err := phase.prepareStageInstructions(img, stg); err != nil {
			return err
		}

		if err := phase.buildStage(img, stg); err != nil {
			return err
		}

		if stg.GetImage().GetStageDescription() == nil {
			panic(fmt.Sprintf("expected stage %s image %q built image info (image name = %s) to be set!", stg.Name(), img.GetName(), stg.GetImage().Name()))
		}

		// Add managed image record only if there was at least one newly built stage
		phase.ShouldAddManagedImageRecord = true

		return nil
	}
}

func (phase *BuildPhase) fetchBaseImageForStage(img *Image, stg stage.Interface) error {
	if stg.Name() == "from" {
		if err := img.FetchBaseImage(phase.Conveyor); err != nil {
			return fmt.Errorf("unable to fetch base image %s for stage %s: %s", img.GetBaseImage().Name(), stg.LogDetailedName(), err)
		}
	} else if stg.Name() == "dockerfile" {
		return nil
	} else {
		return phase.Conveyor.StagesManager.FetchStage(phase.StagesIterator.PrevBuiltStage)
	}

	return nil
}

func castToStageImage(img container_runtime.ImageInterface) *container_runtime.StageImage {
	if img == nil {
		return nil
	}
	return img.(*container_runtime.StageImage)
}

func (phase *BuildPhase) calculateStage(img *Image, stg stage.Interface, shouldBeBuiltMode bool) error {
	stageDependencies, err := stg.GetDependencies(phase.Conveyor, phase.StagesIterator.GetPrevImage(img, stg), phase.StagesIterator.GetPrevBuiltImage(img, stg))
	if err != nil {
		return err
	}

	stageSig, err := calculateSignature(string(stg.Name()), stageDependencies, phase.StagesIterator.PrevNonEmptyStage, phase.Conveyor)
	if err != nil {
		return err
	}
	stg.SetSignature(stageSig)

	if stages, err := phase.Conveyor.StagesManager.GetStagesBySignature(stg.LogDetailedName(), stageSig); err != nil {
		return err
	} else {
		if stageDesc, err := phase.Conveyor.StagesManager.SelectSuitableStage(phase.Conveyor, stg, stages); err != nil {
			return err
		} else if stageDesc != nil {
			i := phase.Conveyor.GetOrCreateStageImage(castToStageImage(phase.StagesIterator.GetPrevImage(img, stg)), stageDesc.Info.Name)
			i.SetStageDescription(stageDesc)
			stg.SetImage(i)
		} else {
			if shouldBeBuiltMode {
				phase.printShouldBeBuiltError(img, stg)
				return fmt.Errorf("stages required")
			}

			// Will build a new image
			i := phase.Conveyor.GetOrCreateStageImage(castToStageImage(phase.StagesIterator.GetPrevImage(img, stg)), uuid.New().String())
			stg.SetImage(i)
		}
	}

	stageContentSig, err := calculateSignature(fmt.Sprintf("%s-content", stg.Name()), "", stg, phase.Conveyor)
	if err != nil {
		return fmt.Errorf("unable to calculate stage %s content-signature: %s", stg.Name(), err)
	}
	stg.SetContentSignature(stageContentSig)

	return nil
}

func (phase *BuildPhase) prepareStageInstructions(img *Image, stg stage.Interface) error {
	logboek.Debug.LogF("-- BuildPhase.prepareStage %s %s\n", img.LogDetailedName(), stg.LogDetailedName())

	stageImage := stg.GetImage()

	serviceLabels := map[string]string{
		imagePkg.WerfDockerImageName:     stageImage.Name(),
		imagePkg.WerfLabel:               phase.Conveyor.projectName(),
		imagePkg.WerfVersionLabel:        werf.Version,
		imagePkg.WerfCacheVersionLabel:   imagePkg.BuildCacheVersion,
		imagePkg.WerfImageLabel:          "false",
		imagePkg.WerfStageSignatureLabel: stg.GetSignature(),
	}

	switch stg.(type) {
	case *stage.DockerfileStage:
		var buildArgs []string

		for key, value := range serviceLabels {
			buildArgs = append(buildArgs, fmt.Sprintf("--label=%s=%s", key, value))
		}

		stageImage.DockerfileImageBuilder().AppendBuildArgs(buildArgs...)

		phase.Conveyor.AppendOnTerminateFunc(func() error {
			return stageImage.DockerfileImageBuilder().Cleanup()
		})

	default:
		imageServiceCommitChangeOptions := stageImage.Container().ServiceCommitChangeOptions()
		imageServiceCommitChangeOptions.AddLabel(serviceLabels)

		if phase.Conveyor.sshAuthSock != "" {
			imageRunOptions := stageImage.Container().RunOptions()
			imageRunOptions.AddVolume(fmt.Sprintf("%s:/.werf/tmp/ssh-auth-sock", phase.Conveyor.sshAuthSock))
			imageRunOptions.AddEnv(map[string]string{"SSH_AUTH_SOCK": "/.werf/tmp/ssh-auth-sock"})
		}
	}

	err := stg.PrepareImage(phase.Conveyor, phase.StagesIterator.GetPrevBuiltImage(img, stg), stageImage)
	if err != nil {
		return fmt.Errorf("error preparing stage %s: %s", stg.Name(), err)
	}

	return nil
}

func (phase *BuildPhase) buildStage(img *Image, stg stage.Interface) error {
	_, err := stapel.GetOrCreateContainer()
	if err != nil {
		return fmt.Errorf("get or create stapel container failed: %s", err)
	}

	infoSectionFunc := func(err error) {
		if err != nil {
			_ = logboek.WithIndent(func() error {
				logImageCommands(stg.GetImage())
				return nil
			})
			return
		}
		logImageInfo(stg.GetImage(), phase.getPrevNonEmptyStageImageSize(), false)
	}

	if err := logboek.Default.LogProcess(
		fmt.Sprintf("Building stage %s", stg.LogDetailedName()),
		logboek.LevelLogProcessOptions{
			InfoSectionFunc: infoSectionFunc,
			Style:           logboek.HighlightStyle(),
		},
		func() (err error) {
			if err := stg.PreRunHook(phase.Conveyor); err != nil {
				return fmt.Errorf("%s preRunHook failed: %s", stg.LogDetailedName(), err)
			}

			return phase.atomicBuildStageImage(img, stg)
		},
	); err != nil {
		return err
	}

	if phase.IntrospectOptions.ImageStageShouldBeIntrospected(img.GetName(), string(stg.Name())) {
		if err := introspectStage(stg); err != nil {
			return err
		}
	}

	return nil
}

func (phase *BuildPhase) atomicBuildStageImage(img *Image, stg stage.Interface) error {
	stageImage := stg.GetImage()

	if v := os.Getenv("WERF_TEST_ATOMIC_STAGE_BUILD__SLEEP_SECONDS_BEFORE_STAGE_BUILD"); v != "" {
		seconds := 0
		fmt.Sscanf(v, "%d", &seconds)
		fmt.Printf("Sleeping %d seconds before building new image by signature %s...\n", seconds, stg.GetSignature())
		time.Sleep(time.Duration(seconds) * time.Second)
	}

	if err := logboek.WithTag(fmt.Sprintf("%s/%s", img.LogName(), stg.Name()), img.LogTagStyle(), func() error {
		if err := stageImage.Build(phase.ImageBuildOptions); err != nil {
			return err
		}
		return nil
	}); err != nil {
		return fmt.Errorf("failed to build image for stage %s with signature %s: %s", stg.Name(), stg.GetSignature(), err)
	}

	if v := os.Getenv("WERF_TEST_ATOMIC_STAGE_BUILD__SLEEP_SECONDS_BEFORE_STAGE_SAVE"); v != "" {
		seconds := 0
		fmt.Sscanf(v, "%d", &seconds)
		fmt.Printf("Sleeping %d seconds before saving newly built image %s into stages storage %s by signature %s...\n", seconds, stg.GetImage().GetBuiltId(), phase.Conveyor.StagesManager.StagesStorage.String(), stg.GetSignature())
		time.Sleep(time.Duration(seconds) * time.Second)
	}

	if lock, err := phase.Conveyor.StorageLockManager.LockStage(phase.Conveyor.projectName(), stg.GetSignature()); err != nil {
		return fmt.Errorf("unable to lock project %s signature %s: %s", phase.Conveyor.projectName(), stg.GetSignature(), err)
	} else {
		defer phase.Conveyor.StorageLockManager.Unlock(lock)
	}

	if stages, err := phase.Conveyor.StagesManager.GetStagesBySignature(stg.LogDetailedName(), stg.GetSignature()); err != nil {
		return err
	} else {
		if stageDesc, err := phase.Conveyor.StagesManager.SelectSuitableStage(phase.Conveyor, stg, stages); err != nil {
			return err
		} else if stageDesc != nil {
			logboek.Default.LogF(
				"Discarding newly built image for stage %s by signature %s: detected already existing image %s in the stages storage\n",
				stg.LogDetailedName(), stg.GetSignature(), stageDesc.Info.Name,
			)
			i := phase.Conveyor.GetOrCreateStageImage(phase.StagesIterator.GetPrevImage(img, stg).(*container_runtime.StageImage), stageDesc.Info.Name)
			i.SetStageDescription(stageDesc)
			stg.SetImage(i)
			return nil
		} else {
			newStageImageName, uniqueID := phase.Conveyor.StagesManager.GenerateStageUniqueID(stg.GetSignature(), stages)
			repository, tag := image.ParseRepositoryAndTag(newStageImageName)

			stageImageObj := phase.Conveyor.GetStageImage(stageImage.Name())
			phase.Conveyor.UnsetStageImage(stageImageObj.Name())

			stageImageObj.SetName(newStageImageName)
			stageImageObj.GetStageDescription().Info.Name = newStageImageName
			stageImageObj.GetStageDescription().Info.Repository = repository
			stageImageObj.GetStageDescription().Info.Tag = tag
			stageImageObj.GetStageDescription().StageID = &image.StageID{Signature: stg.GetSignature(), UniqueID: uniqueID}

			phase.Conveyor.SetStageImage(stageImageObj)

			if err := logboek.Default.LogProcess(
				fmt.Sprintf("Store into stages storage"),
				logboek.LevelLogProcessOptions{},
				func() error {
					if err := phase.Conveyor.StagesManager.StagesStorage.StoreImage(&container_runtime.DockerImage{Image: stageImage}); err != nil {
						return fmt.Errorf("unable to store stage %s signature %s image %s into stages storage %s: %s", stg.LogDetailedName(), stg.GetSignature(), stageImage.Name(), phase.Conveyor.StagesManager.StagesStorage.String(), err)
					}
					return nil
				},
			); err != nil {
				return err
			}

			var stageIDs []image.StageID
			for _, stageDesc := range stages {
				stageIDs = append(stageIDs, *stageDesc.StageID)
			}
			stageIDs = append(stageIDs, *stageImage.GetStageDescription().StageID)

			return phase.Conveyor.StagesManager.AtomicStoreStagesBySignatureToCache(string(stg.Name()), stg.GetSignature(), stageIDs)
		}
	}
}

func introspectStage(s stage.Interface) error {
	return logboek.Info.LogProcess(
		fmt.Sprintf("Introspecting stage %s", s.Name()),
		logboek.LevelLogProcessOptions{Style: logboek.HighlightStyle()},
		func() error {
			if err := logboek.WithRawStreamsOutputModeOn(s.GetImage().Introspect); err != nil {
				return fmt.Errorf("introspect error failed: %s", err)
			}

			return nil
		},
	)
}

var (
	logImageInfoLeftPartWidth = 12
	logImageInfoFormat        = fmt.Sprintf("  %%%ds: %%s\n", logImageInfoLeftPartWidth)
)

func logImageInfo(img container_runtime.ImageInterface, prevStageImageSize int64, isUsingCache bool) {
	repository, tag := image.ParseRepositoryAndTag(img.Name())
	logboek.Default.LogFDetails(logImageInfoFormat, "repository", repository)
	logboek.Default.LogFDetails(logImageInfoFormat, "image_id", stringid.TruncateID(img.GetStageDescription().Info.ID))
	logboek.Default.LogFDetails(logImageInfoFormat, "created", img.GetStageDescription().Info.GetCreatedAt())
	logboek.Default.LogFDetails(logImageInfoFormat, "tag", tag)

	if prevStageImageSize == 0 {
		logboek.Default.LogFDetails(logImageInfoFormat, "size", byteCountBinary(img.GetStageDescription().Info.Size))
	} else {
		logboek.Default.LogFDetails(logImageInfoFormat, "diff", byteCountBinary(img.GetStageDescription().Info.Size-prevStageImageSize))
	}

	if !isUsingCache {
		changes := img.Container().UserCommitChanges()
		if len(changes) != 0 {
			fitTextOptions := logboek.FitTextOptions{ExtraIndentWidth: logImageInfoLeftPartWidth + 4}
			formattedCommands := strings.TrimLeft(logboek.FitText(strings.Join(changes, "\n"), fitTextOptions), " ")
			logboek.Default.LogFDetails(logImageInfoFormat, "instructions", formattedCommands)
		}

		logImageCommands(img)
	}
}

func logImageCommands(img container_runtime.ImageInterface) {
	commands := img.Container().UserRunCommands()
	if len(commands) != 0 {
		fitTextOptions := logboek.FitTextOptions{ExtraIndentWidth: logImageInfoLeftPartWidth + 4}
		formattedCommands := strings.TrimLeft(logboek.FitText(strings.Join(commands, "\n"), fitTextOptions), " ")
		logboek.Default.LogFDetails(logImageInfoFormat, "commands", formattedCommands)
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

func calculateSignature(stageName, stageDependencies string, prevNonEmptyStage stage.Interface, conveyor *Conveyor) (string, error) {
	checksumArgs := []string{image.BuildCacheVersion, stageName, stageDependencies}
	if prevNonEmptyStage != nil {
		prevStageDependencies, err := prevNonEmptyStage.GetNextStageDependencies(conveyor)
		if err != nil {
			return "", fmt.Errorf("unable to get prev stage %s dependencies for the stage %s: %s", prevNonEmptyStage.Name(), stageName, err)
		}

		checksumArgs = append(checksumArgs, prevNonEmptyStage.GetSignature(), prevStageDependencies)
	}

	signature := util.Sha3_224Hash(checksumArgs...)

	blockMsg := fmt.Sprintf("Stage %s signature %s", stageName, signature)
	_ = logboek.Debug.LogBlock(blockMsg, logboek.LevelLogBlockOptions{}, func() error {
		checksumArgsNames := []string{
			"BuildCacheVersion",
			"stageName",
			"stageDependencies",
			"prevNonEmptyStage signature",
			"prevNonEmptyStage dependencies for next stage",
		}
		for ind, checksumArg := range checksumArgs {
			logboek.Debug.LogF("%s => %q\n", checksumArgsNames[ind], checksumArg)
		}
		return nil
	})

	return signature, nil
}

// TODO: move these prints to the after-images hook, print summary over all images
func (phase *BuildPhase) printShouldBeBuiltError(img *Image, stg stage.Interface) {
	logProcessOptions := logboek.LevelLogProcessOptions{Style: logboek.HighlightStyle()}
	logboek.Default.LogProcess("Built stages cache check", logProcessOptions, func() error {
		logboek.LogWarnF("%s with signature %s is not exist in stages storage\n", stg.LogDetailedName(), stg.GetSignature())

		var reasonNumber int
		reasonNumberFunc := func() string {
			reasonNumber++
			return fmt.Sprintf("(%d) ", reasonNumber)
		}

		logboek.LogWarnLn()
		logboek.LogWarnLn("There are some possible reasons:")
		logboek.LogWarnLn()

		if img.isDockerfileImage {
			logboek.LogWarnLn(reasonNumberFunc() + `Dockerfile has COPY or ADD instruction which uses non-permanent data that affects stage signature:
- .git directory which should be excluded with .dockerignore file (https://docs.docker.com/engine/reference/builder/#dockerignore-file)
- auto-generated file`)
			logboek.LogWarnLn()
		}

		logboek.LogWarnLn(reasonNumberFunc() + `werf.yaml has non-permanent data that affects stage signature:
- environment variable (e.g. {{ env "JOB_ID" }})
- dynamic go template function (e.g. one of sprig date functions http://masterminds.github.io/sprig/date.html)
- auto-generated file content (e.g. {{ .Files.Get "hash_sum_of_something" }})`)
		logboek.LogWarnLn()

		logboek.LogWarnLn(`Stage signature dependencies can be found here, https://werf.io/documentation/reference/stages_and_images.html#stage-dependencies.

To quickly find the problem compare current and previous rendered werf configurations.
Get the path at the beginning of command output by the following prefix 'Using werf config render file: '.
E.g.:

  diff /tmp/werf-config-render-502883762 /tmp/werf-config-render-837625028`)
		logboek.LogWarnLn()

		logboek.LogWarnLn(reasonNumberFunc() + `Stages have not been built yet or stages have been removed:
- automatically with werf cleanup command
- manually with werf purge, werf stages purge or werf host purge commands`)
		logboek.LogWarnLn()

		return nil
	})
}
