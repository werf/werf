package build

import (
	"fmt"
	"strings"
	"time"

	"github.com/ghodss/yaml"
	"github.com/google/uuid"

	"github.com/docker/docker/pkg/stringid"

	"github.com/flant/logboek"

	"github.com/flant/werf/pkg/build/stage"
	"github.com/flant/werf/pkg/image"
	imagePkg "github.com/flant/werf/pkg/image"
	"github.com/flant/werf/pkg/stapel"
	"github.com/flant/werf/pkg/storage"
	"github.com/flant/werf/pkg/util"
	"github.com/flant/werf/pkg/werf"
)

const (
	MaxStageNameLength = 22
)

type BuildPhaseOptions struct {
	ShouldBeBuiltMode bool
	ImageBuildOptions imagePkg.BuildOptions
	IntrospectOptions IntrospectOptions
}

type BuildStagesOptions struct {
	ImageBuildOptions image.BuildOptions
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
	return &BuildPhase{BasePhase: BasePhase{c}, BuildPhaseOptions: opts}
}

type BuildPhase struct {
	BasePhase

	isBaseImagePrepared bool

	PrevStage                  stage.Interface
	PrevNonEmptyStage          stage.Interface
	PrevImage                  *image.StageImage
	PrevBuiltImage             image.ImageInterface
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
	phase.isBaseImagePrepared = false
	phase.PrevStage = nil
	phase.PrevNonEmptyStage = nil
	phase.PrevImage = nil
	phase.PrevBuiltImage = nil
	phase.PrevNonEmptyStageImageSize = 0

	if err := phase.Conveyor.StagesStorage.AddManagedImage(phase.Conveyor.projectName(), img.GetName()); err != nil {
		return fmt.Errorf("unable to add image %q to the managed images of project %q: %s", img.GetName(), phase.Conveyor.projectName(), err)
	}

	img.SetupBaseImage(phase.Conveyor)

	phase.PrevImage = img.GetBaseImage()
	if err := phase.PrevImage.SyncDockerState(); err != nil {
		return err
	}

	return nil
}

func (phase *BuildPhase) AfterImageStages(img *Image) error {
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

func (phase *BuildPhase) OnImageStage(img *Image, stg stage.Interface) (bool, error) {
	defer func() {
		phase.PrevStage = stg
		logboek.Debug.LogF("Set prev stage = %q %s\n", phase.PrevStage.Name(), phase.PrevStage.GetSignature())
	}()

	isEmpty, err := stg.IsEmpty(phase.Conveyor, phase.PrevBuiltImage)
	if err != nil {
		return false, fmt.Errorf("error checking stage %s is empty: %s", stg.Name(), err)
	}
	if isEmpty {
		return false, nil
	}

	if err := phase.calculateStageSignature(img, stg, phase.ShouldBeBuiltMode); err != nil {
		return false, err
	}
	if phase.ShouldBeBuiltMode {
		return true, nil
	}
	if err := phase.prepareStage(img, stg); err != nil {
		return false, err
	}
	if err := phase.buildStage(img, stg); err != nil {
		return false, err
	}

	return true, nil
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

func (phase *BuildPhase) calculateStageSignature(img *Image, stg stage.Interface, shouldBeBuiltMode bool) error {
	stageDependencies, err := stg.GetDependencies(phase.Conveyor, phase.PrevImage, phase.PrevBuiltImage)
	if err != nil {
		return err
	}

	stageSig, err := calculateSignature(string(stg.Name()), stageDependencies, phase.PrevNonEmptyStage, phase.Conveyor)
	if err != nil {
		return err
	}
	stg.SetSignature(stageSig)

	var i *image.StageImage
	var shouldResetCache bool
	var suitableImageFound bool

	cacheExists, cacheImagesDescs, err := phase.getImagesBySignatureFromCache(string(stg.Name()), stageSig)
	if err != nil {
		return err
	}

	if cacheExists {
		suitableImageFound, i, err = phase.selectSuitableStagesStorageImage(stg, cacheImagesDescs)
		if err != nil {
			return err
		}

		if suitableImageFound && !i.IsExists() {
			logboek.Debug.LogF(
				"Stage %q image %s by signature %s from stages storage cache is not exists: resetting stages storage cache\n",
				stg.Name(), stageSig, i.Name(),
			)
			shouldResetCache = true
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

		suitableImageFound, i, err = phase.selectSuitableStagesStorageImage(stg, imagesDescs)
		if err != nil {
			return err
		}
	}

	if !suitableImageFound {
		if shouldBeBuiltMode {
			phase.printShouldBeBuiltError(img, stg)
			return fmt.Errorf("stages required")
		}

		// Will build a new image
		i = phase.Conveyor.GetOrCreateStageImage(phase.PrevImage, uuid.New().String())
		stg.SetImage(i)
	}

	phase.PrevNonEmptyStage = stg
	logboek.Debug.LogF("Set prev non empty stage = %q %s\n", phase.PrevNonEmptyStage.Name(), phase.PrevNonEmptyStage.GetSignature())
	phase.PrevImage = i
	logboek.Debug.LogF("Set prev image = %q\n", phase.PrevImage.Name())
	if i.IsExists() {
		phase.PrevBuiltImage = i
		logboek.Debug.LogF("Set prev built image = %q (image already exists)\n", phase.PrevBuiltImage.Name())
	}

	stageContentSig, err := calculateSignature(fmt.Sprintf("%s-content", stg.Name()), "", phase.PrevNonEmptyStage, phase.Conveyor)
	if err != nil {
		return fmt.Errorf("unable to calculate stage %s content-signature: %s", stg.Name(), err)
	}
	stg.SetContentSignature(stageContentSig)

	return nil
}

func (phase *BuildPhase) printShouldBeBuiltError(img *Image, stg stage.Interface) {
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
}

func (phase *BuildPhase) selectSuitableStagesStorageImage(stg stage.Interface, imagesDescs []*storage.ImageInfo) (bool, *image.StageImage, error) {
	if len(imagesDescs) == 0 {
		return false, nil, nil
	}

	var imgInfo *storage.ImageInfo
	if err := logboek.Info.LogProcess(
		fmt.Sprintf("Selecting suitable image for stage %s by signature %s", stg.Name(), stg.GetSignature()),
		logboek.LevelLogProcessOptions{},
		func() error {
			var err error
			imgInfo, err = stg.SelectCacheImage(imagesDescs)
			return err
		},
	); err != nil {
		return false, nil, err
	}
	if imgInfo == nil {
		return false, nil, nil
	}

	imgInfoData, err := yaml.Marshal(imgInfo)
	if err != nil {
		panic(err)
	}

	_ = logboek.Debug.LogBlock("Selected cache image", logboek.LevelLogBlockOptions{Style: logboek.HighlightStyle()}, func() error {
		logboek.Debug.LogF(string(imgInfoData))
		return nil
	})

	i := phase.Conveyor.GetOrCreateStageImage(phase.PrevImage, imgInfo.ImageName)
	stg.SetImage(i)

	if err := logboek.Info.LogProcess(
		fmt.Sprintf("Sync stage %s signature %s image %s from stages storage", stg.Name(), stg.GetSignature(), i.Name()),
		logboek.LevelLogProcessOptions{},
		func() error {
			if err := phase.Conveyor.StagesStorage.SyncStageImage(i); err != nil {
				return fmt.Errorf("unable to sync image %s from stages storage %s: %s", i.Name(), phase.Conveyor.StagesStorage.String(), err)
			}
			return nil
		},
	); err != nil {
		return false, nil, err
	}

	return true, i, nil
}

func (phase *BuildPhase) getImagesBySignatureFromCache(stageName, stageSig string) (bool, []*storage.ImageInfo, error) {
	var cacheExists bool
	var cacheImagesDescs []*storage.ImageInfo

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

func (phase *BuildPhase) atomicGetImagesBySignatureFromStagesStorageWithCacheReset(stageName, stageSig string) ([]*storage.ImageInfo, error) {
	if err := phase.Conveyor.StorageLockManager.LockStageCache(phase.Conveyor.projectName(), stageSig); err != nil {
		return nil, fmt.Errorf("error locking project %s stage %s cache: %s", phase.Conveyor.projectName(), stageSig, err)
	}
	defer phase.Conveyor.StorageLockManager.UnlockStageCache(phase.Conveyor.projectName(), stageSig)

	var originImagesDescs []*storage.ImageInfo
	var err error
	if err := logboek.Info.LogProcess(
		fmt.Sprintf("Getting stage %s images by signature %s from stages storage", stageName, stageSig),
		logboek.LevelLogProcessOptions{},
		func() error {
			originImagesDescs, err = phase.Conveyor.StagesStorage.GetImagesBySignature(phase.Conveyor.projectName(), stageSig)
			if err != nil {
				return fmt.Errorf("error getting project %s stage %s images from stages storage: %s", phase.Conveyor.StagesStorage.String(), stageSig, err)
			}

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

func (phase *BuildPhase) atomicStoreStageCache(stageName, stageSig string, imagesDescs []*storage.ImageInfo) error {
	if err := phase.Conveyor.StorageLockManager.LockStageCache(phase.Conveyor.projectName(), stageSig); err != nil {
		return fmt.Errorf("error locking stage %q cache by signature %s: %s", stageName, stageSig, err)
	}
	defer phase.Conveyor.StorageLockManager.UnlockStageCache(phase.Conveyor.projectName(), stageSig)

	return logboek.Info.LogProcess(
		fmt.Sprintf("Storing stage %q images by signature %s into stages storage cache", stageName, stageSig),
		logboek.LevelLogProcessOptions{},
		func() error {
			if err := phase.Conveyor.StagesStorageCache.StoreImagesBySignature(phase.Conveyor.projectName(), stageSig, imagesDescs); err != nil {
				return fmt.Errorf("error storing stage %q images by signature %s into stages storage cache: %s", stageName, stageSig, err)
			}
			return nil
		},
	)
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

	err := stg.PrepareImage(phase.Conveyor, phase.PrevBuiltImage, stageImage)
	if err != nil {
		return fmt.Errorf("error preparing stage %q: %s", stg.Name(), err)
	}

	phase.Conveyor.SetImageBySignature(stg.GetSignature(), stageImage)

	return nil
}

func (phase *BuildPhase) buildStage(img *Image, stg stage.Interface) error {
	isUsingCache := stg.GetImage().IsExists()

	if isUsingCache {
		logboek.Default.LogFHighlight("Use cache image for %s\n", stg.LogDetailedName())

		logImageInfo(stg.GetImage(), phase.PrevNonEmptyStageImageSize, isUsingCache)

		logboek.LogOptionalLn()

		phase.PrevNonEmptyStageImageSize = stg.GetImage().Inspect().Size

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
		logImageInfo(stg.GetImage(), phase.PrevNonEmptyStageImageSize, isUsingCache)
	}

	if err := logboek.Default.LogProcess(
		fmt.Sprintf("Building %s", stg.LogDetailedName()),
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

	phase.PrevNonEmptyStageImageSize = stg.GetImage().Inspect().Size

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
		return stageImage.Build(phase.ImageBuildOptions)
	}); err != nil {
		return fmt.Errorf("failed to build image for stage %q with signature %s: %s", stg.Name(), stg.GetSignature(), err)
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
		var imgInfo *storage.ImageInfo
		if err := logboek.Info.LogProcess(
			fmt.Sprintf("Selecting suitable image for stage %q by signature %s", stg.Name(), stg.GetSignature()),
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
				"Discarding newly built image for stage %q by signature %s: detected already existing image %s in the stages storage\n",
				stg.Name(), stg.GetSignature(), imgInfo.ImageName,
			)
			i := phase.Conveyor.GetOrCreateStageImage(phase.PrevImage, imgInfo.ImageName)
			stg.SetImage(i)
			phase.PrevImage = i
			logboek.Debug.LogF("Set prev image = %q\n", phase.PrevImage.Name())
			phase.PrevBuiltImage = i
			logboek.Debug.LogF("Set prev built image = %q (discard newly built image, set built image to already existing image)\n", phase.PrevBuiltImage.Name())

			if err := logboek.Info.LogProcess(
				fmt.Sprintf("Sync stage %q signature %s image %s from stages storage", stg.Name(), stg.GetSignature(), i.Name()),
				logboek.LevelLogProcessOptions{},
				func() error {
					if err := phase.Conveyor.StagesStorage.SyncStageImage(i); err != nil {
						return fmt.Errorf("unable to sync image %s from stages storage %s: %s", i.Name(), phase.Conveyor.StagesStorage.String(), err)
					}
					return nil
				},
			); err != nil {
				return err
			}

			return nil
		}
	}

	newStageImageName := phase.generateUniqStageImageName(stg.GetSignature(), imagesDescs)

	stageImageObj := phase.Conveyor.GetStageImage(stageImage.Name())
	phase.Conveyor.UnsetStageImage(stageImageObj.Name())

	stageImageObj.SetName(newStageImageName)
	phase.Conveyor.SetStageImage(stageImageObj)

	phase.PrevBuiltImage = stageImageObj
	logboek.Debug.LogF("Set prev built image = %q (newly built image)\n", phase.PrevBuiltImage.Name())

	if err := logboek.Info.LogProcess(
		fmt.Sprintf("Store stage %q signature %s image %s into stages storage", stageImage.Name(), stg.GetSignature(), stageImage.Name()),
		logboek.LevelLogProcessOptions{},
		func() error {
			if err := phase.Conveyor.StagesStorage.StoreStageImage(stageImage); err != nil {
				return fmt.Errorf("unable to store stage %q signature %s image %s into stages storage %s: %s", stg.Name(), stg.GetSignature(), stageImage.Name(), phase.Conveyor.StagesStorage.String(), err)
			}
			return nil
		},
	); err != nil {
		return err
	}

	imagesDescs = append(imagesDescs, &storage.ImageInfo{
		Signature:         stg.GetSignature(),
		ImageName:         stageImage.Name(),
		Labels:            stageImage.Labels(),
		CreatedAtUnixNano: stageImage.CreatedAtUnixNano(),
	})
	return phase.atomicStoreStageCache(string(stg.Name()), stg.GetSignature(), imagesDescs)
}

func (phase *BuildPhase) generateUniqStageImageName(signature string, imagesDescs []*storage.ImageInfo) string {
	var imageName string

	for {
		timeNow := time.Now().UTC()
		timeNowMicroseconds := timeNow.Unix()*1000 + int64(timeNow.Nanosecond()/1000000)
		uniqueID := fmt.Sprintf("%d", timeNowMicroseconds)
		imageName = fmt.Sprintf(image.LocalImageStageImageFormat, phase.Conveyor.projectName(), signature, uniqueID)

		for _, imgInfo := range imagesDescs {
			if imgInfo.ImageName == imageName {
				continue
			}
		}
		return imageName
	}
}

func introspectStage(s stage.Interface) error {
	return logboek.Info.LogProcess(
		fmt.Sprintf("Introspecting stage %q", s.Name()),
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

func logImageInfo(img imagePkg.ImageInterface, prevStageImageSize int64, isUsingCache bool) {
	parts := strings.Split(img.Name(), ":")
	repository, tag := parts[0], parts[1]

	logboek.Default.LogFDetails(logImageInfoFormat, "repository", repository)
	logboek.Default.LogFDetails(logImageInfoFormat, "image_id", stringid.TruncateID(img.ID()))
	logboek.Default.LogFDetails(logImageInfoFormat, "created", img.Inspect().Created)
	logboek.Default.LogFDetails(logImageInfoFormat, "tag", tag)

	if prevStageImageSize == 0 {
		logboek.Default.LogFDetails(logImageInfoFormat, "size", byteCountBinary(img.Inspect().Size))
	} else {
		logboek.Default.LogFDetails(logImageInfoFormat, "diff", byteCountBinary(img.Inspect().Size-prevStageImageSize))
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

func logImageCommands(img imagePkg.ImageInterface) {
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
