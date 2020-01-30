package build

import (
	"fmt"
	"strings"

	"github.com/docker/docker/pkg/stringid"

	imagePkg "github.com/flant/werf/pkg/image"
	"github.com/flant/werf/pkg/werf"

	"github.com/flant/logboek"
	"github.com/flant/werf/pkg/build/stage"
	"github.com/flant/werf/pkg/image"
	"github.com/flant/werf/pkg/util"
	"github.com/google/uuid"
)

const (
	MaxStageNameLength = 22
)

type BuildPhaseOptions struct {
	SignaturesOnly    bool
	ImageBuildOptions imagePkg.BuildOptions
	IntrospectOptions IntrospectOptions
}

func NewBuildPhase(conveyor *Conveyor, opts BuildPhaseOptions) *BuildPhase {
	return &BuildPhase{Conveyor: conveyor, BuildPhaseOptions: opts}
}

type BuildPhase struct {
	isBaseImagePrepared bool

	Conveyor *Conveyor

	PrevStage          stage.Interface
	PrevImage          *image.StageImage
	PrevBuiltImage     image.ImageInterface
	PrevStageImageSize int64

	BuildPhaseOptions
}

func (phase *BuildPhase) Name() string {
	return "build"
}

func (phase *BuildPhase) OnStart(img *Image) error {
	img.SetupBaseImage(phase.Conveyor)

	phase.PrevImage = img.GetBaseImage()
	if err := phase.PrevImage.SyncDockerState(); err != nil {
		return err
	}

	return nil
}

func (phase *BuildPhase) HandleStage(img *Image, stg stage.Interface) error {
	defer func() {
		phase.PrevStage = stg
	}()

	isEmpty, err := stg.IsEmpty(phase.Conveyor, phase.PrevBuiltImage)
	if err != nil {
		return fmt.Errorf("error checking stage %s is empty: %s", stg.Name(), err)
	}
	if isEmpty {
		logboek.LogInfoF("%s:%s <empty>\n", stg.Name(), strings.Repeat(" ", MaxStageNameLength-len(stg.Name())))
		return nil
	}

	if err := phase.calculateStageSignature(img, stg); err != nil {
		return err
	}
	if phase.SignaturesOnly {
		return nil
	}
	if err := phase.prepareStage(img, stg); err != nil {
		return err
	}
	if err := phase.buildStage(img, stg); err != nil {
		return err
	}

	return nil
}

func (phase *BuildPhase) calculateStageSignature(img *Image, stg stage.Interface) error {
	stageDependencies, err := stg.GetDependencies(phase.Conveyor, phase.PrevImage, phase.PrevBuiltImage)
	if err != nil {
		return err
	}

	checksumArgs := []string{stageDependencies, image.BuildCacheVersion}
	if phase.PrevStage != nil {
		prevStageDependencies, err := phase.PrevStage.GetNextStageDependencies(phase.Conveyor)
		if err != nil {
			return fmt.Errorf("unable to get prev stage %s dependencies for the stage %s: %s", phase.PrevStage.Name(), stg.Name(), err)
		}

		checksumArgs = append(checksumArgs, phase.PrevStage.GetSignature(), prevStageDependencies)
	}
	stageSig := util.Sha256Hash(checksumArgs...)
	stg.SetSignature(stageSig)

	logboek.LogInfoF("%s:%s %s\n", stg.Name(), strings.Repeat(" ", MaxStageNameLength-len(stg.Name())), stageSig)

	imagesDescs, err := phase.Conveyor.StagesStorage.GetImagesBySignature(phase.Conveyor.projectName(), stageSig)
	if err != nil {
		return fmt.Errorf("unable to get images from stages storage %s by signature %s: %s", phase.Conveyor.StagesStorage.String(), stageSig)
	}

	var imageExists bool
	var i *image.StageImage

	if len(imagesDescs) > 0 {
		imgInfo, err := stg.SelectCacheImage(imagesDescs)
		if err != nil {
			return err
		}

		if imgInfo != nil {
			fmt.Printf("-- SelectCacheImage => %v\n", imgInfo)
			imageExists = true

			i = phase.Conveyor.GetOrCreateImage(phase.PrevImage, imgInfo.ImageName)
			stg.SetImage(i)

			if err := phase.Conveyor.StagesStorage.SyncStageImage(i); err != nil {
				return fmt.Errorf("unable to fetch image %s from stages storage %s: %s", imgInfo.ImageName, phase.Conveyor.StagesStorage.String(), err)
			}
		}
	}

	if !imageExists {
		imageName := fmt.Sprintf(image.LocalImageStageImageFormat, phase.Conveyor.projectName(), stageSig, util.UUIDToShortString(uuid.New()))
		// TODO: timestamp and check timestamp is unique using stages-storage when building

		i = phase.Conveyor.GetOrCreateImage(phase.PrevImage, imageName)
		stg.SetImage(i)
	}

	if err = stg.AfterImageSyncDockerStateHook(phase.Conveyor); err != nil {
		return err
	}

	phase.PrevImage = i
	if phase.PrevImage.IsExists() {
		phase.PrevBuiltImage = phase.PrevImage
	}

	return nil
}

func (phase *BuildPhase) prepareStage(img *Image, stg stage.Interface) error {
	if !phase.isBaseImagePrepared {
		if !img.isDockerfileImage {
			if err := img.PrepareBaseImage(phase.Conveyor); err != nil {
				return fmt.Errorf("prepare base image %s failed: %s", img.GetBaseImage().Name(), err)
			}
		}
		phase.isBaseImagePrepared = true
	}

	stageImage := stg.GetImage()

	if phase.Conveyor.GetImageBySignature(stg.GetSignature()) != nil || stageImage.IsExists() {
		// Do not prepare this image second time, because it has been already prepared for this conveyor instance
		return nil
	}

	switch certainStage := stg.(type) {
	case *stage.DockerfileStage:
		var buildArgs []string

		for key, value := range map[string]string{
			imagePkg.WerfDockerImageName:   stageImage.Name(),
			imagePkg.WerfLabel:             phase.Conveyor.projectName(),
			imagePkg.WerfVersionLabel:      werf.Version,
			imagePkg.WerfCacheVersionLabel: imagePkg.BuildCacheVersion,
			imagePkg.WerfImageLabel:        "false",
		} {
			buildArgs = append(buildArgs, fmt.Sprintf("--label=%s=%s", key, value))
		}

		buildArgs = append(buildArgs, certainStage.DockerBuildArgs()...)
		stageImage.DockerfileImageBuilder().AppendBuildArgs(buildArgs...)

	default:
		imageServiceCommitChangeOptions := stageImage.Container().ServiceCommitChangeOptions()
		imageServiceCommitChangeOptions.AddLabel(map[string]string{
			imagePkg.WerfDockerImageName:   stageImage.Name(),
			imagePkg.WerfLabel:             phase.Conveyor.projectName(),
			imagePkg.WerfVersionLabel:      werf.Version,
			imagePkg.WerfCacheVersionLabel: imagePkg.BuildCacheVersion,
			imagePkg.WerfImageLabel:        "false",
		})

		if phase.Conveyor.sshAuthSock != "" {
			imageRunOptions := stageImage.Container().RunOptions()
			imageRunOptions.AddVolume(fmt.Sprintf("%s:/.werf/tmp/ssh-auth-sock", phase.Conveyor.sshAuthSock))
			imageRunOptions.AddEnv(map[string]string{"SSH_AUTH_SOCK": "/.werf/tmp/ssh-auth-sock"})
		}
	}

	err := stg.PrepareImage(phase.Conveyor, phase.PrevBuiltImage, stageImage)
	if err != nil {
		return fmt.Errorf("error preparing stage %s: %s", stg.Name(), err)
	}

	phase.Conveyor.SetImageBySignature(stg.GetSignature(), stageImage)

	return nil
}

func (phase *BuildPhase) buildStage(img *Image, stg stage.Interface) error {
	stageImage := stg.GetImage()

	isUsingCache := stageImage.IsExists()

	if isUsingCache {
		logboek.LogHighlightF("Use cache image for %s\n", stg.LogDetailedName())

		logImageInfo(stageImage, phase.PrevStageImageSize, isUsingCache)

		logboek.LogOptionalLn()

		phase.PrevStageImageSize = stageImage.Inspect().Size

		if phase.IntrospectOptions.ImageStageShouldBeIntrospected(img.GetName(), string(stg.Name())) {
			if err := introspectStage(stg); err != nil {
				return err
			}
		}

		return nil
	}

	infoSectionFunc := func(err error) {
		if err != nil {
			_ = logboek.WithIndent(func() error {
				logImageCommands(stageImage)
				return nil
			})

			return
		}

		logImageInfo(stageImage, phase.PrevStageImageSize, isUsingCache)
	}

	logProcessOptions := logboek.LogProcessOptions{InfoSectionFunc: infoSectionFunc, ColorizeMsgFunc: logboek.ColorizeHighlight}
	err := logboek.LogProcess(fmt.Sprintf("Building %s", stg.LogDetailedName()), logProcessOptions, func() (err error) {
		if err := stg.PreRunHook(phase.Conveyor); err != nil {
			return fmt.Errorf("%s preRunHook failed: %s", stg.LogDetailedName(), err)
		}

		if err := logboek.WithTag(fmt.Sprintf("%s/%s", img.LogName(), stg.Name()), img.LogTagColorizeFunc(), func() error {
			if err := stageImage.Build(phase.ImageBuildOptions); err != nil {
				return fmt.Errorf("failed to build %s: %s", stageImage.Name(), err)
			}

			if err := phase.Conveyor.StagesStorageLockManager.LockStage(phase.Conveyor.projectName(), stg.GetSignature()); err != nil {
				return fmt.Errorf("unable to lock project %s signature %s: %s", phase.Conveyor.projectName(), stg.GetSignature(), err)
			}
			defer phase.Conveyor.StagesStorageLockManager.UnlockStage(phase.Conveyor.projectName(), stg.GetSignature())

			var imageExists = false
			imagesDescs, err := phase.Conveyor.StagesStorage.GetImagesBySignature(phase.Conveyor.projectName(), stg.GetSignature())
			if err != nil {
				return fmt.Errorf("unable to get images from stages storage %s by signature %s: %s", phase.Conveyor.StagesStorage.String(), stg.GetSignature())
			}
			if len(imagesDescs) > 0 {
				imgInfo, err := stg.SelectCacheImage(imagesDescs)
				if err != nil {
					return err
				}

				if imgInfo != nil {
					imageExists = true
					panic("Suitable image by signature already exists!")
					// recalculate signatures, restart build
				}
			}

			if !imageExists {
				if err := phase.Conveyor.StagesStorage.StoreStageImage(stageImage); err != nil {
					return fmt.Errorf("unable to store image %s into stages storage %s: %s", stageImage.Name(), phase.Conveyor.StagesStorage.String(), err)
				}
			}

			return nil
		}); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return err
	}

	phase.PrevStageImageSize = stageImage.Inspect().Size

	if phase.IntrospectOptions.ImageStageShouldBeIntrospected(img.GetName(), string(stg.Name())) {
		if err := introspectStage(stg); err != nil {
			return err
		}
	}

	return nil
}

func introspectStage(s stage.Interface) error {
	logProcessMessage := fmt.Sprintf("Introspecting stage %s", s.Name())
	logProcessOptions := logboek.LogProcessOptions{ColorizeMsgFunc: logboek.ColorizeHighlight}
	return logboek.LogProcess(logProcessMessage, logProcessOptions, func() error {
		if err := logboek.WithRawStreamsOutputModeOn(s.GetImage().Introspect); err != nil {
			return fmt.Errorf("introspect error failed: %s", err)
		}

		return nil
	})
}

var (
	logImageInfoLeftPartWidth = 12
	logImageInfoFormat        = fmt.Sprintf("  %%%ds: %%s\n", logImageInfoLeftPartWidth)
)

func logImageInfo(img imagePkg.ImageInterface, prevStageImageSize int64, isUsingCache bool) {
	parts := strings.Split(img.Name(), ":")
	repository, tag := parts[0], parts[1]

	logboek.LogInfoF(logImageInfoFormat, "repository", repository)
	logboek.LogInfoF(logImageInfoFormat, "image_id", stringid.TruncateID(img.ID()))
	logboek.LogInfoF(logImageInfoFormat, "created", img.Inspect().Created)
	logboek.LogInfoF(logImageInfoFormat, "tag", tag)

	if prevStageImageSize == 0 {
		logboek.LogInfoF(logImageInfoFormat, "size", byteCountBinary(img.Inspect().Size))
	} else {
		logboek.LogInfoF(logImageInfoFormat, "diff", byteCountBinary(img.Inspect().Size-prevStageImageSize))
	}

	if !isUsingCache {
		changes := img.Container().UserCommitChanges()
		if len(changes) != 0 {
			fitTextOptions := logboek.FitTextOptions{ExtraIndentWidth: logImageInfoLeftPartWidth + 4}
			formattedCommands := strings.TrimLeft(logboek.FitText(strings.Join(changes, "\n"), fitTextOptions), " ")
			logboek.LogInfoF(logImageInfoFormat, "instructions", formattedCommands)
		}

		logImageCommands(img)
	}
}

func logImageCommands(img imagePkg.ImageInterface) {
	commands := img.Container().UserRunCommands()
	if len(commands) != 0 {
		fitTextOptions := logboek.FitTextOptions{ExtraIndentWidth: logImageInfoLeftPartWidth + 4}
		formattedCommands := strings.TrimLeft(logboek.FitText(strings.Join(commands, "\n"), fitTextOptions), " ")
		logboek.LogInfoF(logImageInfoFormat, "commands", formattedCommands)
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
