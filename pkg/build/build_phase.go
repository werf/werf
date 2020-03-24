package build

import (
	"fmt"
	"strings"
	"time"

	"github.com/flant/werf/pkg/container_runtime"

	"github.com/ghodss/yaml"
	"github.com/google/uuid"

	"github.com/docker/docker/pkg/stringid"

	"github.com/flant/logboek"

	"github.com/flant/werf/pkg/build/stage"
	"github.com/flant/werf/pkg/image"
	imagePkg "github.com/flant/werf/pkg/image"
	"github.com/flant/werf/pkg/stapel"
	"github.com/flant/werf/pkg/util"
	"github.com/flant/werf/pkg/werf"
)

const (
	MaxStageNameLength = 22
)

type BuildPhaseOptions struct {
	SignaturesOnly    bool
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

	PrevStage                  stage.Interface
	PrevNonEmptyStage          stage.Interface
	PrevBuiltStage             stage.Interface
	PrevNonEmptyStageImageSize int64

	BuildPhaseOptions
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
	phase.PrevStage = nil
	phase.PrevNonEmptyStage = nil
	phase.PrevBuiltStage = nil
	phase.PrevNonEmptyStageImageSize = 0

	if err := phase.Conveyor.StagesStorage.AddManagedImage(phase.Conveyor.projectName(), img.GetName()); err != nil {
		return fmt.Errorf("unable to add image %q to the managed images of project %q: %s", img.GetName(), phase.Conveyor.projectName(), err)
	}

	img.SetupBaseImage(phase.Conveyor)

	return nil
}

func (phase *BuildPhase) AfterImageStages(img *Image) error {
	if shouldCleanup, err := phase.Conveyor.StagesStorage.ShouldCleanupLocalImage(&container_runtime.DockerImage{Image: phase.PrevBuiltStage.GetImage()}); err == nil && shouldCleanup {
		if err := logboek.Default.LogProcess(
			fmt.Sprintf("Cleaning up stage %s local image", phase.PrevBuiltStage.LogDetailedName()),
			logboek.LevelLogProcessOptions{Style: logboek.HighlightStyle()},
			func() error {
				logboek.Info.LogF("Image name: %s\n", phase.PrevBuiltStage.GetImage().Name())
				if err := phase.Conveyor.StagesStorage.CleanupLocalImage(&container_runtime.DockerImage{Image: phase.PrevBuiltStage.GetImage()}); err != nil {
					return fmt.Errorf("unable to cleanup stage %s local image %s for stages storage %s: %s", phase.PrevBuiltStage.LogDetailedName(), phase.PrevBuiltStage.GetImage().Name(), phase.Conveyor.StagesStorage.String(), err)
				}
				return nil
			},
		); err != nil {
			return err
		}
	} else if err != nil {
		return err
	}

	img.SetLastNonEmptyStage(phase.PrevNonEmptyStage)

	stagesSig, err := calculateSignature("imageStages", "", phase.PrevNonEmptyStage, phase.Conveyor)
	if err != nil {
		return fmt.Errorf("unable to calculate image %s stages-signature: %s", img.GetName(), err)
	}
	img.SetContentSignature(stagesSig)

	return nil
}

/*
	TODO: calculating-signatures, prepare and build logs

SIGNATURES LOGS
func (p *SignaturesPhase) Run(c *Conveyor) error {
	logProcessOptions := logboek.LogProcessOptions{ColorizeMsgFunc: logboek.ColorizeHighlight}
	return logboek.LogProcess("Calculating signatures", logProcessOptions, func() error {
		return logboek.WithoutIndent(func() error { return p.run(c) })
	})
}
func (p *SignaturesPhase) run(c *Conveyor) error {
	for _, image := range c.imagesInOrder {
		if err := logboek.LogProcess(image.LogDetailedName(), logboek.LogProcessOptions{ColorizeMsgFunc: image.LogProcessColorizeFunc()}, func() error {
			return p.calculateImageSignatures(c, image)
		}); err != nil {
			return err
		}
	}

	return nil
}

PREPARE LOGS
	logProcessOptions := logboek.LogProcessOptions{ColorizeMsgFunc: logboek.ColorizeHighlight}
	return logboek.LogProcess("Preparing stages build instructions", logProcessOptions, func() error {
		return p.run(c)
	})
func (p *PrepareStagesPhase) run(c *Conveyor) (err error) {
	for _, image := range c.imagesInOrder {
		if err := logboek.LogProcess(image.LogDetailedName(), logboek.LogProcessOptions{ColorizeMsgFunc: image.LogProcessColorizeFunc()}, func() error {
			return p.runImage(image, c)
		}); err != nil {
			return err
		}
	}


BUILD LOGS
logProcessOptions := logboek.LogProcessOptions{ColorizeMsgFunc: logboek.ColorizeHighlight}
	return logboek.LogProcess("Building stages", logProcessOptions, func() error {
		return p.run(c)
	})

images := c.imagesInOrder
	for _, image := range images {
		if err := logboek.LogProcess(image.LogDetailedName(), logboek.LogProcessOptions{ColorizeMsgFunc: image.LogProcessColorizeFunc()}, func() error {
			return p.runImage(image, c)
		}); err != nil {
			return err
		}
	}

	return nil
}
*/

func (phase *BuildPhase) GetPrevImage(img *Image, stg stage.Interface) container_runtime.ImageInterface {
	if stg.Name() == "from" {
		return img.GetBaseImage()
	} else if phase.PrevNonEmptyStage != nil {
		return phase.PrevNonEmptyStage.GetImage()
	}
	return nil
}

func (phase *BuildPhase) GetPrevBuiltImage(img *Image, stg stage.Interface) container_runtime.ImageInterface {
	if stg.Name() == "from" {
		return img.GetBaseImage()
	} else if phase.PrevBuiltStage != nil {
		return phase.PrevBuiltStage.GetImage()
	}
	return nil
}

func (phase *BuildPhase) OnImageStage(img *Image, stg stage.Interface) (bool, error) {
	isEmpty, err := stg.IsEmpty(phase.Conveyor, phase.GetPrevBuiltImage(img, stg))
	if err != nil {
		return false, fmt.Errorf("error checking stage %s is empty: %s", stg.Name(), err)
	}

	if stg.Name() != "from" {
		if phase.PrevStage == nil {
			panic(fmt.Sprintf("expected PrevStage to be set for image %q stage %s!", img.GetName(), stg.Name()))
		}
	}

	if !isEmpty {
		if phase.SignaturesOnly {
			if err := phase.calculateStageSignature(img, stg); err != nil {
				return false, err
			}
		} else {
			if stg.Name() != "from" {
				if phase.PrevNonEmptyStage == nil {
					panic(fmt.Sprintf("expected PrevNonEmptyStage to be set for image %q stage %s", img.GetName(), stg.Name()))
				}
				if phase.PrevBuiltStage == nil {
					panic(fmt.Sprintf("expected PrevBuiltStage to be set for image %q stage %s", img.GetName(), stg.Name()))
				}
				if phase.PrevBuiltStage != phase.PrevNonEmptyStage {
					panic(fmt.Sprintf("expected PrevBuiltStage (%q) to equal PrevNonEmptyStage (%q) for image %q stage %s", phase.PrevBuiltStage.Name(), phase.PrevNonEmptyStage.Name(), img.GetName(), stg.Name()))
				}
			}

			if err := phase.calculateStageSignature(img, stg); err != nil {
				return false, err
			}
			if err := phase.prepareStage(img, stg); err != nil {
				return false, err
			}
			if err := phase.buildStage(img, stg); err != nil {
				return false, err
			}

			if stg.GetImage().GetStagesStorageImageInfo() == nil {
				panic(fmt.Sprintf("expected stage %s image %q built image info (image name = %s) to be set!", stg.Name(), img.GetName(), stg.GetImage().Name()))
			}
		}
	}

	phase.PrevStage = stg
	logboek.Debug.LogF("Set prev stage = %q %s\n", phase.PrevStage.Name(), phase.PrevStage.GetSignature())

	if !isEmpty {
		phase.PrevNonEmptyStage = stg
		logboek.Debug.LogF("Set prev non empty stage = %q %s\n", phase.PrevNonEmptyStage.Name(), phase.PrevNonEmptyStage.GetSignature())

		if phase.PrevNonEmptyStage.GetImage().GetStagesStorageImageInfo() != nil {
			phase.PrevNonEmptyStageImageSize = phase.PrevNonEmptyStage.GetImage().GetStagesStorageImageInfo().Size
			logboek.Debug.LogF("Set prev non empty stage image size = %d %q %s\n", phase.PrevNonEmptyStageImageSize, phase.PrevNonEmptyStage.Name(), phase.PrevNonEmptyStage.GetSignature())
		}

		if stg.GetImage().GetStagesStorageImageInfo() != nil {
			phase.PrevBuiltStage = stg
			logboek.Debug.LogF("Set prev built stage = %q (image %s)\n", phase.PrevBuiltStage.Name(), phase.PrevBuiltStage.GetImage().Name())
		}

		stageContentSig, err := calculateSignature(fmt.Sprintf("%s-content", stg.Name()), "", phase.PrevNonEmptyStage, phase.Conveyor)
		if err != nil {
			return false, fmt.Errorf("unable to calculate stage %s content-signature: %s", stg.Name(), err)
		}
		stg.SetContentSignature(stageContentSig)

		return true, nil
	}

	return false, nil
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

func (phase *BuildPhase) calculateStageSignature(img *Image, stg stage.Interface) error {
	stageDependencies, err := stg.GetDependencies(phase.Conveyor, phase.GetPrevImage(img, stg), phase.GetPrevBuiltImage(img, stg))
	if err != nil {
		return err
	}

	stageSig, err := calculateSignature(string(stg.Name()), stageDependencies, phase.PrevNonEmptyStage, phase.Conveyor)
	if err != nil {
		return err
	}
	stg.SetSignature(stageSig)

	var i *container_runtime.StageImage
	var shouldResetCache bool
	var suitableImageFound bool

	cacheExists, cacheImagesDescs, err := phase.getImagesBySignatureFromCache(string(stg.Name()), stageSig)
	if err != nil {
		return err
	}

	if cacheExists {
		if imgInfo, err := phase.selectSuitableStagesStorageImage(stg, cacheImagesDescs); err != nil {
			return err
		} else if imgInfo != nil {
			if freshImgInfo, err := phase.Conveyor.StagesStorage.GetImageInfo(phase.Conveyor.projectName(), imgInfo.Signature, imgInfo.UniqueID); err != nil {
				return fmt.Errorf("unable to get image %q info from stages storage: %s", imgInfo.Name, err)
			} else if freshImgInfo == nil {
				logboek.Debug.LogF(
					"Stage %q image %s by signature %s from stages storage cache is not exists: resetting stages storage cache\n",
					stg.Name(), stageSig, imgInfo.Name,
				)
				shouldResetCache = true
			} else {
				suitableImageFound = true

				i = phase.Conveyor.GetOrCreateStageImage(phase.GetPrevImage(img, stg).(*container_runtime.StageImage), freshImgInfo.Name)
				i.SetStagesStorageImageInfo(freshImgInfo)
				stg.SetImage(i)
			}
		}
	} else {
		logboek.Debug.LogF(
			"Stage %q cache by signature %s is not exists in stages storage cache: resetting stages storage cache\n",
			stg.Name(), stageSig,
		)
		shouldResetCache = true
	}

	if shouldResetCache {
		imagesDescs, err := phase.atomicGetImagesBySignatureFromStagesStorageWithCacheReset(string(stg.Name()), stageSig)
		if err != nil {
			return err
		}

		if imgInfo, err := phase.selectSuitableStagesStorageImage(stg, imagesDescs); err != nil {
			return err
		} else if imgInfo != nil {
			suitableImageFound = true

			i = phase.Conveyor.GetOrCreateStageImage(phase.GetPrevImage(img, stg).(*container_runtime.StageImage), imgInfo.Name)
			i.SetStagesStorageImageInfo(imgInfo)
			stg.SetImage(i)
		}
	}

	if !suitableImageFound {
		// Will build a new image
		i = phase.Conveyor.GetOrCreateStageImage(phase.GetPrevImage(img, stg).(*container_runtime.StageImage), uuid.New().String())
		stg.SetImage(i)
	}

	if err = stg.AfterSignatureCalculated(phase.Conveyor); err != nil {
		return err
	}

	return nil
}

func (phase *BuildPhase) selectSuitableStagesStorageImage(stg stage.Interface, imagesDescs []*image.Info) (*image.Info, error) {
	if len(imagesDescs) == 0 {
		return nil, nil
	}

	var imgInfo *image.Info
	if err := logboek.Info.LogProcess(
		fmt.Sprintf("Selecting suitable image for stage %s by signature %s", stg.Name(), stg.GetSignature()),
		logboek.LevelLogProcessOptions{},
		func() error {
			var err error
			imgInfo, err = stg.SelectCacheImage(imagesDescs)
			return err
		},
	); err != nil {
		return nil, err
	}
	if imgInfo == nil {
		return nil, nil
	}

	imgInfoData, err := yaml.Marshal(imgInfo)
	if err != nil {
		panic(err)
	}

	_ = logboek.Debug.LogBlock("Selected cache image", logboek.LevelLogBlockOptions{Style: logboek.HighlightStyle()}, func() error {
		logboek.Debug.LogF(string(imgInfoData))
		return nil
	})

	return imgInfo, nil
}

func (phase *BuildPhase) getImagesBySignatureFromCache(stageName, stageSig string) (bool, []*image.Info, error) {
	var cacheExists bool
	var cacheImagesDescs []*image.Info

	err := logboek.Info.LogProcess(
		fmt.Sprintf("Getting stage %s images by signature %s from stages storage cache", stageName, stageSig),
		logboek.LevelLogProcessOptions{},
		func() error {
			var err error
			cacheExists, cacheImagesDescs, err = phase.Conveyor.StagesStorageCache.GetImagesBySignature(phase.Conveyor.projectName(), stageSig)
			if err != nil {
				return fmt.Errorf("error getting project %s stage %s images from stages storage cache: %s", phase.Conveyor.projectName(), stageSig, err)
			}

			return nil
		},
	)

	return cacheExists, cacheImagesDescs, err
}

func (phase *BuildPhase) atomicGetImagesBySignatureFromStagesStorageWithCacheReset(stageName, stageSig string) ([]*image.Info, error) {
	if err := phase.Conveyor.StorageLockManager.LockStageCache(phase.Conveyor.projectName(), stageSig); err != nil {
		return nil, fmt.Errorf("error locking project %s stage %s cache: %s", phase.Conveyor.projectName(), stageSig, err)
	}
	defer phase.Conveyor.StorageLockManager.UnlockStageCache(phase.Conveyor.projectName(), stageSig)

	var originImagesDescs []*image.Info
	var err error
	if err := logboek.Info.LogProcess(
		fmt.Sprintf("Getting stage %s images by signature %s from stages storage", stageName, stageSig),
		logboek.LevelLogProcessOptions{},
		func() error {
			originImagesDescs, err = phase.Conveyor.StagesStorage.GetRepoImagesBySignature(phase.Conveyor.projectName(), stageSig)
			if err != nil {
				return fmt.Errorf("error getting project %s stage %s images from stages storage: %s", phase.Conveyor.StagesStorage.String(), stageSig, err)
			}

			logboek.Debug.LogF("Images: %#v\n", originImagesDescs)

			return nil
		},
	); err != nil {
		return nil, err
	}

	if err := logboek.Info.LogProcess(
		fmt.Sprintf("Storing stage %s images by signature %s into stages storage cache", stageName, stageSig),
		logboek.LevelLogProcessOptions{},
		func() error {
			if err := phase.Conveyor.StagesStorageCache.StoreImagesBySignature(phase.Conveyor.projectName(), stageSig, originImagesDescs); err != nil {
				return fmt.Errorf("error storing stage %s images by signature %s into stages storage cache: %s", stageName, stageSig, err)
			}
			return nil
		},
	); err != nil {
		return nil, err
	}

	return originImagesDescs, nil
}

func (phase *BuildPhase) atomicStoreStageCache(stageName, stageSig string, imagesDescs []*image.Info) error {
	if err := phase.Conveyor.StorageLockManager.LockStageCache(phase.Conveyor.projectName(), stageSig); err != nil {
		return fmt.Errorf("error locking stage %s cache by signature %s: %s", stageName, stageSig, err)
	}
	defer phase.Conveyor.StorageLockManager.UnlockStageCache(phase.Conveyor.projectName(), stageSig)

	return logboek.Info.LogProcess(
		fmt.Sprintf("Storing stage %s images by signature %s into stages storage cache", stageName, stageSig),
		logboek.LevelLogProcessOptions{},
		func() error {
			if err := phase.Conveyor.StagesStorageCache.StoreImagesBySignature(phase.Conveyor.projectName(), stageSig, imagesDescs); err != nil {
				return fmt.Errorf("error storing stage %s images by signature %s into stages storage cache: %s", stageName, stageSig, err)
			}
			return nil
		},
	)
}

// FIXME: parallel builds should work together with common stages
func (phase *BuildPhase) fetchBaseImageForStage(img *Image, stg stage.Interface) error {
	if stg.Name() == "from" {
		if err := img.FetchBaseImage(phase.Conveyor); err != nil {
			return fmt.Errorf("unable to fetch base image %s for stage %s: %s", img.GetBaseImage().Name(), stg.LogDetailedName(), err)
		}
	} else {
		if shouldFetch, err := phase.Conveyor.StagesStorage.ShouldFetchImage(&container_runtime.DockerImage{Image: phase.PrevBuiltStage.GetImage()}); err == nil && shouldFetch {
			if err := logboek.Default.LogProcess(
				fmt.Sprintf("Fetching stage %s from stages storage", phase.PrevBuiltStage.LogDetailedName()),
				logboek.LevelLogProcessOptions{Style: logboek.HighlightStyle()},
				func() error {
					logboek.Info.LogF("Image name: %s\n", phase.PrevBuiltStage.GetImage().Name())
					if err := phase.Conveyor.StagesStorage.FetchImage(&container_runtime.DockerImage{Image: phase.PrevBuiltStage.GetImage()}); err != nil {
						return fmt.Errorf("unable to fetch stage %s image %s from stages storage %s: %s", phase.PrevBuiltStage.LogDetailedName(), phase.PrevBuiltStage.GetImage().Name(), phase.Conveyor.StagesStorage.String(), err)
					}
					return nil
				},
			); err != nil {
				return err
			}
		} else if err != nil {
			return err
		}
	}

	return nil
}

// FIXME: parallel builds should work together with common stages
func (phase *BuildPhase) cleanupBaseImageForStage(img *Image, stg stage.Interface) error {
	if stg.Name() == "from" {
		if err := img.CleanupBaseImage(phase.Conveyor); err != nil {
			return fmt.Errorf("unable to cleanup base image %s for stage %s: %s", img.GetBaseImage().Name(), stg.LogDetailedName(), err)
		}
	} else {
		if shouldCleanup, err := phase.Conveyor.StagesStorage.ShouldCleanupLocalImage(&container_runtime.DockerImage{Image: phase.PrevBuiltStage.GetImage()}); err == nil && shouldCleanup {
			if err := logboek.Default.LogProcess(
				fmt.Sprintf("Cleaning up stage %s local image", phase.PrevBuiltStage.LogDetailedName()),
				logboek.LevelLogProcessOptions{Style: logboek.HighlightStyle()},
				func() error {
					logboek.Info.LogF("Image name: %s\n", phase.PrevBuiltStage.GetImage().Name())
					if err := phase.Conveyor.StagesStorage.CleanupLocalImage(&container_runtime.DockerImage{Image: phase.PrevBuiltStage.GetImage()}); err != nil {
						return fmt.Errorf("unable to cleanup stage %s local image %s for stages storage %s: %s", phase.PrevBuiltStage.LogDetailedName(), phase.PrevBuiltStage.GetImage().Name(), phase.Conveyor.StagesStorage.String(), err)
					}
					return nil
				},
			); err != nil {
				return err
			}
		} else if err != nil {
			return err
		}
	}

	return nil
}

func (phase *BuildPhase) prepareStage(img *Image, stg stage.Interface) error {
	// Nothing to do if stage already exists in the stages storage
	if stg.GetImage().GetStagesStorageImageInfo() != nil {
		return nil
	}

	if err := phase.fetchBaseImageForStage(img, stg); err != nil {
		return err
	}

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

	err := stg.PrepareImage(phase.Conveyor, phase.GetPrevBuiltImage(img, stg), stageImage)
	if err != nil {
		return fmt.Errorf("error preparing stage %s: %s", stg.Name(), err)
	}

	return nil
}

func (phase *BuildPhase) buildStage(img *Image, stg stage.Interface) error {
	if stg.GetImage().GetStagesStorageImageInfo() != nil {
		logboek.Default.LogFHighlight("Use cache image for %s\n", stg.LogDetailedName())

		logImageInfo(stg.GetImage(), phase.PrevNonEmptyStageImageSize, true)

		logboek.LogOptionalLn()

		if phase.IntrospectOptions.ImageStageShouldBeIntrospected(img.GetName(), string(stg.Name())) {
			if err := introspectStage(stg); err != nil {
				return err
			}
		}

		return nil
	}

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
		logImageInfo(stg.GetImage(), phase.PrevNonEmptyStageImageSize, false)
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

	if err := phase.cleanupBaseImageForStage(img, stg); err != nil {
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

	if err := logboek.WithTag(fmt.Sprintf("%s/%s", img.LogName(), stg.Name()), img.LogTagStyle(), func() error {
		if err := stageImage.Build(phase.ImageBuildOptions); err != nil {
			return err
		}
		return nil
	}); err != nil {
		return fmt.Errorf("failed to build image for stage %s with signature %s: %s", stg.Name(), stg.GetSignature(), err)
	}

	if err := phase.Conveyor.StorageLockManager.LockStage(phase.Conveyor.projectName(), stg.GetSignature()); err != nil {
		return fmt.Errorf("unable to lock project %s signature %s: %s", phase.Conveyor.projectName(), stg.GetSignature(), err)
	}
	defer phase.Conveyor.StorageLockManager.UnlockStage(phase.Conveyor.projectName(), stg.GetSignature())

	imagesDescs, err := phase.atomicGetImagesBySignatureFromStagesStorageWithCacheReset(string(stg.Name()), stg.GetSignature())
	if err != nil {
		return err
	}

	if len(imagesDescs) > 0 {
		var imgInfo *image.Info
		if err := logboek.Info.LogProcess(
			fmt.Sprintf("Selecting suitable image for stage %s by signature %s", stg.Name(), stg.GetSignature()),
			logboek.LevelLogProcessOptions{},
			func() error {
				imgInfo, err = stg.SelectCacheImage(imagesDescs)
				return err
			},
		); err != nil {
			return err
		}

		if imgInfo != nil {
			logboek.Default.LogF(
				"Discarding newly built image for stage %s by signature %s: detected already existing image %s in the stages storage\n",
				stg.Name(), stg.GetSignature(), imgInfo.Name,
			)
			i := phase.Conveyor.GetOrCreateStageImage(phase.GetPrevImage(img, stg).(*container_runtime.StageImage), imgInfo.Name)
			i.SetStagesStorageImageInfo(imgInfo)
			stg.SetImage(i)

			return nil
		}
	}

	newStageImageName, uniqueID := phase.generateUniqueImageName(stg.GetSignature(), imagesDescs)
	repository, tag := image.ParseRepositoryAndTag(newStageImageName)

	stageImageObj := phase.Conveyor.GetStageImage(stageImage.Name())
	phase.Conveyor.UnsetStageImage(stageImageObj.Name())

	stageImageObj.SetName(newStageImageName)
	stageImageObj.GetStagesStorageImageInfo().Name = newStageImageName
	stageImageObj.GetStagesStorageImageInfo().Repository = repository
	stageImageObj.GetStagesStorageImageInfo().Tag = tag
	stageImageObj.GetStagesStorageImageInfo().Signature = stg.GetSignature()
	stageImageObj.GetStagesStorageImageInfo().UniqueID = uniqueID

	phase.Conveyor.SetStageImage(stageImageObj)

	if err := logboek.Info.LogProcess(
		fmt.Sprintf("Store stage %s signature %s image %s into stages storage", stageImage.Name(), stg.GetSignature(), stageImage.Name()),
		logboek.LevelLogProcessOptions{},
		func() error {
			if err := phase.Conveyor.StagesStorage.StoreImage(&container_runtime.DockerImage{stageImage}); err != nil {
				return fmt.Errorf("unable to store stage %s signature %s image %s into stages storage %s: %s", stg.Name(), stg.GetSignature(), stageImage.Name(), phase.Conveyor.StagesStorage.String(), err)
			}
			return nil
		},
	); err != nil {
		return err
	}

	imagesDescs = append(imagesDescs, stageImage.GetStagesStorageImageInfo())
	return phase.atomicStoreStageCache(string(stg.Name()), stg.GetSignature(), imagesDescs)
}

func (phase *BuildPhase) generateUniqueImageName(signature string, imagesDescs []*image.Info) (string, string) {
	var imageName string

	for {
		timeNow := time.Now().UTC()
		timeNowMicroseconds := timeNow.Unix()*1000 + int64(timeNow.Nanosecond()/1000000)
		uniqueID := fmt.Sprintf("%d", timeNowMicroseconds)
		imageName = phase.Conveyor.StagesStorage.ConstructStageImageName(phase.Conveyor.projectName(), signature, uniqueID)

		for _, imgInfo := range imagesDescs {
			if imgInfo.Name == imageName {
				continue
			}
		}
		return imageName, uniqueID
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
	logboek.Default.LogFDetails(logImageInfoFormat, "image_id", stringid.TruncateID(img.GetStagesStorageImageInfo().ID))
	logboek.Default.LogFDetails(logImageInfoFormat, "created", img.GetStagesStorageImageInfo().GetCreatedAt())
	logboek.Default.LogFDetails(logImageInfoFormat, "tag", tag)

	if prevStageImageSize == 0 {
		logboek.Default.LogFDetails(logImageInfoFormat, "size", byteCountBinary(img.GetStagesStorageImageInfo().Size))
	} else {
		logboek.Default.LogFDetails(logImageInfoFormat, "diff", byteCountBinary(img.GetStagesStorageImageInfo().Size-prevStageImageSize))
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
