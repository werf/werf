package build

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"

	"github.com/docker/cli/cli/command/image/build"
	"github.com/docker/docker/pkg/fileutils"
	"github.com/moby/buildkit/frontend/dockerfile/instructions"
	"github.com/moby/buildkit/frontend/dockerfile/parser"
	"github.com/moby/buildkit/frontend/dockerfile/shell"

	"github.com/flant/logboek"

	"github.com/flant/werf/pkg/build/import_server"
	"github.com/flant/werf/pkg/build/stage"
	"github.com/flant/werf/pkg/config"
	"github.com/flant/werf/pkg/container_runtime"
	"github.com/flant/werf/pkg/git_repo"
	"github.com/flant/werf/pkg/images_manager"
	"github.com/flant/werf/pkg/logging"
	"github.com/flant/werf/pkg/path_matcher"
	"github.com/flant/werf/pkg/storage"
	"github.com/flant/werf/pkg/tag_strategy"
	"github.com/flant/werf/pkg/util"
)

type Conveyor struct {
	werfConfig          *config.WerfConfig
	imageNamesToProcess []string

	projectDir       string
	containerWerfDir string
	baseTmpDir       string

	baseImagesRepoIdsCache map[string]string
	baseImagesRepoErrCache map[string]error

	sshAuthSock string

	gitReposCaches map[string]*stage.GitRepoCache

	imagesInOrder []*Image

	stageImages                     map[string]*container_runtime.StageImage
	buildingGitStageNameByImageName map[string]stage.StageName
	localGitRepo                    *git_repo.Local
	remoteGitRepos                  map[string]*git_repo.Remote

	tmpDir string

	ContainerRuntime container_runtime.ContainerRuntime

	StagesStorage      storage.StagesStorage
	StagesStorageCache storage.StagesStorageCache
	ImagesRepo         storage.ImagesRepo
	StorageLockManager storage.LockManager

	onTerminateFuncs []func() error
	importServers    map[string]import_server.ImportServer
}

func NewConveyor(werfConfig *config.WerfConfig, imageNamesToProcess []string, projectDir, baseTmpDir, sshAuthSock string, containerRuntime container_runtime.ContainerRuntime, stagesStorage storage.StagesStorage, stagesStorageCache storage.StagesStorageCache, imagesRepo storage.ImagesRepo, storageLockManager storage.LockManager) *Conveyor {
	c := &Conveyor{
		werfConfig:          werfConfig,
		imageNamesToProcess: imageNamesToProcess,

		projectDir:       projectDir,
		containerWerfDir: "/.werf",
		baseTmpDir:       baseTmpDir,

		sshAuthSock: sshAuthSock,

		stageImages:                     make(map[string]*container_runtime.StageImage),
		gitReposCaches:                  make(map[string]*stage.GitRepoCache),
		baseImagesRepoIdsCache:          make(map[string]string),
		baseImagesRepoErrCache:          make(map[string]error),
		imagesInOrder:                   []*Image{},
		buildingGitStageNameByImageName: make(map[string]stage.StageName),
		remoteGitRepos:                  make(map[string]*git_repo.Remote),
		tmpDir:                          filepath.Join(baseTmpDir, string(util.GenerateConsistentRandomString(10))),
		importServers:                   make(map[string]import_server.ImportServer),

		ContainerRuntime:   containerRuntime,
		StagesStorage:      stagesStorage,
		StagesStorageCache: stagesStorageCache,
		ImagesRepo:         imagesRepo,
		StorageLockManager: storageLockManager,
	}

	return c
}

func (c *Conveyor) GetImportServer(imageName string) (import_server.ImportServer, error) {
	if srv, hasKey := c.importServers[imageName]; hasKey {
		return srv, nil
	}

	var srv *import_server.RsyncServer

	if err := logboek.Info.LogProcess(fmt.Sprintf("Firing up import rsync server for image %s", imageName), logboek.LevelLogProcessOptions{}, func() error {
		tmpDir := path.Join(c.tmpDir, "import-server", imageName)
		if err := os.MkdirAll(tmpDir, os.ModePerm); err != nil {
			return fmt.Errorf("unable to create dir %s: %s", tmpDir, err)
		}

		dockerImageName := c.GetImageLastStageImageName(imageName)
		var err error
		srv, err = import_server.RunRsyncServer(dockerImageName, tmpDir)
		if srv != nil {
			c.AppendOnTerminateFunc(func() error {
				if err := srv.Shutdown(); err != nil {
					return fmt.Errorf("unable to shutdown import server %s: %s", srv.DockerContainerName, err)
				}
				return nil
			})
		}
		if err != nil {
			return fmt.Errorf("unable to run rsync import server: %s", err)
		}
		return nil
	}); err != nil {
		return nil, err
	}

	c.importServers[imageName] = srv

	return srv, nil
}

func (c *Conveyor) AppendOnTerminateFunc(f func() error) {
	c.onTerminateFuncs = append(c.onTerminateFuncs, f)
}

func (c *Conveyor) Terminate() error {
	var terminateErrors []error

	for gitRepoName, gitRepoCache := range c.gitReposCaches {
		if err := gitRepoCache.Terminate(); err != nil {
			terminateErrors = append(terminateErrors, fmt.Errorf("unable to terminate cache of git repo '%s': %s", gitRepoName, err))
		}
	}

	for _, onTerminateFunc := range c.onTerminateFuncs {
		if err := onTerminateFunc(); err != nil {
			terminateErrors = append(terminateErrors, err)
		}
	}

	if len(terminateErrors) > 0 {
		errMsg := "Errors occurred during conveyor termination:\n"
		for _, err := range terminateErrors {
			errMsg += fmt.Sprintf(" - %s\n", err)
		}

		// NOTE: Errors printed here because conveyor termination should occur in defer,
		// NOTE: and errors in the defer will be silenced otherwise.
		logboek.LogErrorF("%s", errMsg)

		return errors.New(errMsg)
	}

	return nil
}

func (c *Conveyor) GetGitRepoCache(gitRepoName string) *stage.GitRepoCache {
	if _, hasKey := c.gitReposCaches[gitRepoName]; !hasKey {
		c.gitReposCaches[gitRepoName] = &stage.GitRepoCache{
			Archives:  make(map[string]git_repo.Archive),
			Patches:   make(map[string]git_repo.Patch),
			Checksums: make(map[string]git_repo.Checksum),
		}
	}
	return c.gitReposCaches[gitRepoName]
}

func (c *Conveyor) SetLocalGitRepo(repo *git_repo.Local) {
	c.localGitRepo = repo
}

func (c *Conveyor) GetLocalGitRepo() *git_repo.Local {
	return c.localGitRepo
}

type TagOptions struct {
	CustomTags           []string
	TagsByGitTag         []string
	TagsByGitBranch      []string
	TagsByGitCommit      []string
	TagByStagesSignature bool
}

func (c *Conveyor) ShouldBeBuilt() error {
	if err := c.determineStages(); err != nil {
		return err
	}

	phases := []Phase{
		NewBuildPhase(c, BuildPhaseOptions{SignaturesOnly: true}),
		NewShouldBeBuiltPhase(c),
	}

	return c.runPhases(phases, false)
}

func (c *Conveyor) GetImageInfoGetters(configImages []*config.StapelImage, configImagesFromDockerfile []*config.ImageFromDockerfile, commonTag string, tagStrategy tag_strategy.TagStrategy, withoutRegistry bool) []images_manager.ImageInfoGetter {
	var images []images_manager.ImageInfoGetter

	var imagesNames []string
	for _, imageConfig := range configImages {
		imagesNames = append(imagesNames, imageConfig.Name)
	}
	for _, imageConfig := range configImagesFromDockerfile {
		imagesNames = append(imagesNames, imageConfig.Name)
	}

	for _, imageName := range imagesNames {
		var tag string
		if tagStrategy == tag_strategy.StagesSignature {
			for _, img := range c.imagesInOrder {
				if img.GetName() == imageName {
					tag = img.GetStagesSignature()
					break
				}
			}
		} else {
			tag = commonTag
		}

		d := &images_manager.ImageInfo{
			ImagesRepo:      c.ImagesRepo,
			Name:            imageName,
			Tag:             tag,
			WithoutRegistry: withoutRegistry,
		}
		images = append(images, d)
	}

	return images
}

func (c *Conveyor) BuildStages(opts BuildStagesOptions) error {
	/*var phases []Phase
	phases = append(phases, NewInitializationPhase())
	phases = append(phases, NewSignaturesPhase())
	phases = append(phases, NewPrepareStagesPhase())
	phases = append(phases, NewBuildStagesPhase(stageRepo, opts))

	if err := c.StorageLockManager.LockAllImagesReadOnly(c.projectName()); err != nil {
		return fmt.Errorf("error locking all images read only: %s", err)
	}
	defer c.StorageLockManager.UnlockAllImages(c.projectName())

	return c.runPhases(phases)*/

	if err := c.determineStages(); err != nil {
		return err
	}

	phases := []Phase{NewBuildPhase(c, BuildPhaseOptions{
		IntrospectOptions: opts.IntrospectOptions,
		ImageBuildOptions: opts.ImageBuildOptions,
	})}

	return c.runPhases(phases, true)
}

type PublishImagesOptions struct {
	ImagesToPublish []string
	TagOptions
}

func (c *Conveyor) PublishImages(opts PublishImagesOptions) error {
	if err := c.determineStages(); err != nil {
		return err
	}

	phases := []Phase{
		NewBuildPhase(c, BuildPhaseOptions{SignaturesOnly: true}),
		NewShouldBeBuiltPhase(c),
		NewPublishImagesPhase(c, c.ImagesRepo, opts),
	}

	return c.runPhases(phases, true)

	/*
			var phases []Phase
			phases = append(phases, NewInitializationPhase())
			phases = append(phases, NewSignaturesPhase())
			phases = append(phases, NewShouldBeBuiltPhase())
			phases = append(phases, NewPublishImagesPhase(imagesRepoManager, opts))

		TODO: locks
			if err := c.StorageLockManager.LockAllImagesReadOnly(c.projectName()); err != nil {
				return fmt.Errorf("error locking all images read only: %s", err)
			}
			defer c.StorageLockManager.UnlockAllImages(c.projectName())

			return c.runPhases(phases)
	*/
}

type BuildAndPublishOptions struct {
	BuildStagesOptions
	PublishImagesOptions
}

func (c *Conveyor) BuildAndPublish(opts BuildAndPublishOptions) error {
	if err := c.determineStages(); err != nil {
		return err
	}

	phases := []Phase{
		NewBuildPhase(c, BuildPhaseOptions{ImageBuildOptions: opts.ImageBuildOptions, IntrospectOptions: opts.IntrospectOptions}),
		NewShouldBeBuiltPhase(c),
		NewPublishImagesPhase(c, c.ImagesRepo, opts.PublishImagesOptions),
	}

	return c.runPhases(phases, true)

	/*
		var phases []Phase
		phases = append(phases, NewInitializationPhase())
		phases = append(phases, NewSignaturesPhase())
		phases = append(phases, NewPrepareStagesPhase())
		phases = append(phases, NewBuildStagesPhase(stagesRepo, opts.BuildStagesOptions))
		phases = append(phases, NewPublishImagesPhase(imagesRepoManager, opts.PublishImagesOptions))

		if err := c.StorageLockManager.LockAllImagesReadOnly(c.projectName()); err != nil {
			return fmt.Errorf("error locking all images read only: %s", err)
		}
		defer c.StorageLockManager.UnlockAllImages(c.projectName())

		return c.runPhases(phases)
	*/
}

func (c *Conveyor) determineStages() error {
	return logboek.Info.LogProcess(
		"Determining of stages",
		logboek.LevelLogProcessOptions{Style: logboek.HighlightStyle()},
		func() error {
			return c.doDetermineStages()
		},
	)
}

func (c *Conveyor) doDetermineStages() error {
	imagesInterfaces := getImageConfigsInOrder(c)
	for _, imageInterfaceConfig := range imagesInterfaces {
		var img *Image
		var imageLogName string
		var style *logboek.Style

		switch imageConfig := imageInterfaceConfig.(type) {
		case config.StapelImageInterface:
			imageLogName = logging.ImageLogProcessName(imageConfig.ImageBaseConfig().Name, imageConfig.IsArtifact())
			style = ImageLogProcessStyle(imageConfig.IsArtifact())
		case *config.ImageFromDockerfile:
			imageLogName = logging.ImageLogProcessName(imageConfig.Name, false)
			style = ImageLogProcessStyle(false)
		}

		err := logboek.Info.LogProcess(imageLogName, logboek.LevelLogProcessOptions{Style: style}, func() error {
			var err error

			switch imageConfig := imageInterfaceConfig.(type) {
			case config.StapelImageInterface:
				img, err = prepareImageBasedOnStapelImageConfig(imageConfig, c)
			case *config.ImageFromDockerfile:
				img, err = prepareImageBasedOnImageFromDockerfile(imageConfig, c)
			}

			if err != nil {
				return err
			}

			c.imagesInOrder = append(c.imagesInOrder, img)

			return nil
		})

		if err != nil {
			return err
		}
	}

	return nil
}

func (c *Conveyor) runPhases(phases []Phase, logImages bool) error {
	// TODO: Parallelize builds
	//images (по зависимостям), dependantImagesByStage
	//dependantImagesByStage строится в InitializationPhase, спросить у stage что она ждет.
	//Количество воркеров-goroutine ограничено.
	//Надо распределить images по воркерам.
	//
	//for img := range images {
	//	Goroutine {
	//		phases = append(phases, NewBuildStage())
	//
	//    	for phase = range phases {
	//            phase.OnStart()
	//
	//	    	for stage = range stages {
	//		    	for img = dependantImagesByStage[stage.name] {
	//			    	wait <-imgChanMap[img]
	//				}
	//				phase.HandleStage(stage)
	//    		}
	//		}
	//
	//        close(imgChanMap[img])
	//	} Goroutine
	//}

	var imagesLogger logboek.Level
	if logImages {
		imagesLogger = logboek.Default
	} else {
		imagesLogger = logboek.Info
	}

	for _, phase := range phases {
		logProcessMsg := fmt.Sprintf("Phase %s -- BeforeImages()", phase.Name())
		logboek.Debug.LogProcessStart(logProcessMsg, logboek.LevelLogProcessStartOptions{})
		if err := phase.BeforeImages(); err != nil {
			logboek.Debug.LogProcessFail(logboek.LevelLogProcessFailOptions{})
			return fmt.Errorf("phase %s before images handler failed: %s", phase.Name(), err)
		}
		logboek.Debug.LogProcessEnd(logboek.LevelLogProcessEndOptions{})
	}

	for _, img := range c.imagesInOrder {
		if err := imagesLogger.LogProcess(img.LogDetailedName(), logboek.LevelLogProcessOptions{Style: img.LogProcessStyle()}, func() error {
			for _, phase := range phases {
				logProcessMsg := fmt.Sprintf("Phase %s -- BeforeImageStages()", phase.Name())
				logboek.Debug.LogProcessStart(logProcessMsg, logboek.LevelLogProcessStartOptions{})
				if err := phase.BeforeImageStages(img); err != nil {
					logboek.Debug.LogProcessFail(logboek.LevelLogProcessFailOptions{})
					return fmt.Errorf("phase %s before image %s stages handler failed: %s", phase.Name(), img.GetLogName(), err)
				}
				logboek.Debug.LogProcessEnd(logboek.LevelLogProcessEndOptions{})

				logProcessMsg = fmt.Sprintf("Phase %s -- OnImageStage()", phase.Name())
				logboek.Debug.LogProcessStart(logProcessMsg, logboek.LevelLogProcessStartOptions{})
				var newStages []stage.Interface
				for _, stg := range img.GetStages() {
					logboek.Debug.LogF("Phase %s -- OnImageStage() %s %s\n", phase.Name(), img.GetLogName(), stg.LogDetailedName())
					if keepStage, err := phase.OnImageStage(img, stg); err != nil {
						logboek.Debug.LogProcessFail(logboek.LevelLogProcessFailOptions{})
						return fmt.Errorf("phase %s on image %s stage %s handler failed: %s", phase.Name(), img.GetLogName(), stg.Name(), err)
					} else if keepStage {
						newStages = append(newStages, stg)
					}
				}
				img.SetStages(newStages)
				logboek.Debug.LogProcessEnd(logboek.LevelLogProcessEndOptions{})

				logProcessMsg = fmt.Sprintf("Phase %s -- AfterImageStages()", phase.Name())
				logboek.Debug.LogProcessStart(logProcessMsg, logboek.LevelLogProcessStartOptions{})
				if err := phase.AfterImageStages(img); err != nil {
					logboek.Debug.LogProcessFail(logboek.LevelLogProcessFailOptions{})
					return fmt.Errorf("phase %s after image %s stages handler failed: %s", phase.Name(), img.GetLogName(), err)
				}
				logboek.Debug.LogProcessEnd(logboek.LevelLogProcessEndOptions{})

				logProcessMsg = fmt.Sprintf("Phase %s -- ImageProcessingShouldBeStopped()", phase.Name())
				logboek.Debug.LogProcessStart(logProcessMsg, logboek.LevelLogProcessStartOptions{})
				if phase.ImageProcessingShouldBeStopped(img) {
					logboek.Debug.LogProcessEnd(logboek.LevelLogProcessEndOptions{})
					return nil
				}
				logboek.Debug.LogProcessEnd(logboek.LevelLogProcessEndOptions{})
			}

			return nil
		}); err != nil {
			return err
		}
	}

	for _, phase := range phases {
		if err := logboek.Debug.LogProcess(fmt.Sprintf("Phase %s -- AfterImages()", phase.Name()), logboek.LevelLogProcessOptions{}, func() error {
			if err := phase.AfterImages(); err != nil {
				return fmt.Errorf("phase %s after images handler failed: %s", phase.Name(), err)
			}

			return nil
		}); err != nil {
			return err
		}
	}

	return nil
}

/*
TODO: locks and log
func (c *Conveyor) runPhases(phases []Phase) error {
		logboek.LogOptionalLn()

		for _, phase := range phases {
			err := phase.Run(c)

			if err != nil {
				c.StorageLockManager.ReleaseAllStageLocks()
				return err
			}
		}
	return nil
}
*/

func (c *Conveyor) projectName() string {
	return c.werfConfig.Meta.Project
}

func (c *Conveyor) GetStageImage(name string) *container_runtime.StageImage {
	return c.stageImages[name]
}

func (c *Conveyor) UnsetStageImage(name string) {
	delete(c.stageImages, name)
}

func (c *Conveyor) SetStageImage(stageImage *container_runtime.StageImage) {
	c.stageImages[stageImage.Name()] = stageImage
}

func (c *Conveyor) GetOrCreateStageImage(fromImage *container_runtime.StageImage, name string) *container_runtime.StageImage {
	if img, ok := c.stageImages[name]; ok {
		return img
	}

	img := container_runtime.NewStageImage(fromImage, name, c.ContainerRuntime.(*container_runtime.LocalDockerServerRuntime))
	c.stageImages[name] = img
	return img
}

func (c *Conveyor) GetImage(name string) *Image {
	for _, img := range c.imagesInOrder {
		if img.GetName() == name {
			return img
		}
	}

	panic(fmt.Sprintf("Image '%s' not found!", name))
}

func (c *Conveyor) GetImageStagesSignature(imageName string) string {
	return c.GetImage(imageName).GetStagesSignature()
}

func (c *Conveyor) GetImageLastStageImageName(imageName string) string {
	return c.GetImage(imageName).GetLastNonEmptyStage().GetImage().Name()
}

func (c *Conveyor) GetImageLastStageImageID(imageName string) string {
	return c.GetImage(imageName).GetLastNonEmptyStage().GetImage().GetStagesStorageImageInfo().ID
}

func (c *Conveyor) SetBuildingGitStage(imageName string, stageName stage.StageName) {
	c.buildingGitStageNameByImageName[imageName] = stageName
}

func (c *Conveyor) GetBuildingGitStage(imageName string) stage.StageName {
	stageName, ok := c.buildingGitStageNameByImageName[imageName]
	if !ok {
		return ""
	}

	return stageName
}

func (c *Conveyor) GetImageTmpDir(imageName string) string {
	return filepath.Join(c.tmpDir, "image", imageName)
}

func prepareImageBasedOnStapelImageConfig(imageInterfaceConfig config.StapelImageInterface, c *Conveyor) (*Image, error) {
	image := &Image{}

	imageBaseConfig := imageInterfaceConfig.ImageBaseConfig()
	imageName := imageBaseConfig.Name
	imageArtifact := imageInterfaceConfig.IsArtifact()

	from, fromImageName, fromLatest := getFromFields(imageBaseConfig)

	image.name = imageName

	if from != "" {
		if err := handleImageFromName(from, fromLatest, image, c); err != nil {
			return nil, err
		}
	} else {
		image.baseImageImageName = fromImageName
	}

	image.isArtifact = imageArtifact

	err := initStages(image, imageInterfaceConfig, c)
	if err != nil {
		return nil, err
	}

	return image, nil
}

func handleImageFromName(from string, fromLatest bool, image *Image, c *Conveyor) error {
	image.baseImageName = from

	if fromLatest {
		if _, err := image.getFromBaseImageIdFromRegistry(c, image.baseImageName); err != nil {
			return err
		}
	}

	return nil
}

func getFromFields(imageBaseConfig *config.StapelImageBase) (string, string, bool) {
	var from string
	var fromImageName string

	if imageBaseConfig.From != "" {
		from = imageBaseConfig.From
	} else if imageBaseConfig.FromImageName != "" {
		fromImageName = imageBaseConfig.FromImageName
	} else if imageBaseConfig.FromImageArtifactName != "" {
		fromImageName = imageBaseConfig.FromImageArtifactName
	}

	return from, fromImageName, imageBaseConfig.FromLatest
}

func getImageConfigsInOrder(c *Conveyor) []config.ImageInterface {
	var images []config.ImageInterface
	for _, imageInterf := range getImageConfigsToProcess(c) {
		var imagesInBuildOrder []config.ImageInterface

		switch image := imageInterf.(type) {
		case *config.StapelImage:
			imagesInBuildOrder = c.werfConfig.ImageTree(image)
		case *config.ImageFromDockerfile:
			imagesInBuildOrder = append(imagesInBuildOrder, image)
		}

		for i := 0; i < len(imagesInBuildOrder); i++ {
			if isNotInArr(images, imagesInBuildOrder[i]) {
				images = append(images, imagesInBuildOrder[i])
			}
		}
	}

	return images
}

func getImageConfigsToProcess(c *Conveyor) []config.ImageInterface {
	var imageConfigsToProcess []config.ImageInterface

	if len(c.imageNamesToProcess) == 0 {
		imageConfigsToProcess = c.werfConfig.GetAllImages()
	} else {
		for _, imageName := range c.imageNamesToProcess {
			imageToProcess := c.werfConfig.GetImage(imageName)
			if imageToProcess == nil {
				logboek.LogWarnF("WARNING: Specified image %s isn't defined in werf.yaml!\n", imageName)
			} else {
				imageConfigsToProcess = append(imageConfigsToProcess, imageToProcess)
			}
		}
	}

	return imageConfigsToProcess
}

func isNotInArr(arr []config.ImageInterface, obj config.ImageInterface) bool {
	for _, elm := range arr {
		if reflect.DeepEqual(elm, obj) {
			return false
		}
	}

	return true
}

func initStages(image *Image, imageInterfaceConfig config.StapelImageInterface, c *Conveyor) error {
	var stages []stage.Interface

	imageBaseConfig := imageInterfaceConfig.ImageBaseConfig()
	imageName := imageBaseConfig.Name
	imageArtifact := imageInterfaceConfig.IsArtifact()

	baseStageOptions := &stage.NewBaseStageOptions{
		ImageName:        imageName,
		ConfigMounts:     imageBaseConfig.Mount,
		ImageTmpDir:      c.GetImageTmpDir(imageBaseConfig.Name),
		ContainerWerfDir: c.containerWerfDir,
		ProjectName:      c.werfConfig.Meta.Project,
	}

	gitArchiveStageOptions := &stage.NewGitArchiveStageOptions{
		ArchivesDir:          getImageArchivesDir(imageName, c),
		ScriptsDir:           getImageScriptsDir(imageName, c),
		ContainerArchivesDir: getImageArchivesContainerDir(c),
		ContainerScriptsDir:  getImageScriptsContainerDir(c),
	}

	gitPatchStageOptions := &stage.NewGitPatchStageOptions{
		PatchesDir:           getImagePatchesDir(imageName, c),
		ArchivesDir:          getImageArchivesDir(imageName, c),
		ScriptsDir:           getImageScriptsDir(imageName, c),
		ContainerPatchesDir:  getImagePatchesContainerDir(c),
		ContainerArchivesDir: getImageArchivesContainerDir(c),
		ContainerScriptsDir:  getImageScriptsContainerDir(c),
	}

	gitMappings, err := generateGitMappings(imageBaseConfig, c)
	if err != nil {
		return err
	}

	gitMappingsExist := len(gitMappings) != 0

	stages = appendIfExist(stages, stage.GenerateFromStage(imageBaseConfig, image.baseImageRepoId, baseStageOptions))
	stages = appendIfExist(stages, stage.GenerateBeforeInstallStage(imageBaseConfig, baseStageOptions))
	stages = appendIfExist(stages, stage.GenerateImportsBeforeInstallStage(imageBaseConfig, baseStageOptions))

	if gitMappingsExist {
		stages = append(stages, stage.NewGitArchiveStage(gitArchiveStageOptions, baseStageOptions))
	}

	stages = appendIfExist(stages, stage.GenerateInstallStage(imageBaseConfig, gitPatchStageOptions, baseStageOptions))
	stages = appendIfExist(stages, stage.GenerateImportsAfterInstallStage(imageBaseConfig, baseStageOptions))
	stages = appendIfExist(stages, stage.GenerateBeforeSetupStage(imageBaseConfig, gitPatchStageOptions, baseStageOptions))
	stages = appendIfExist(stages, stage.GenerateImportsBeforeSetupStage(imageBaseConfig, baseStageOptions))
	stages = appendIfExist(stages, stage.GenerateSetupStage(imageBaseConfig, gitPatchStageOptions, baseStageOptions))
	stages = appendIfExist(stages, stage.GenerateImportsAfterSetupStage(imageBaseConfig, baseStageOptions))

	if !imageArtifact {
		if gitMappingsExist {
			stages = append(stages, stage.NewGitCacheStage(gitPatchStageOptions, baseStageOptions))
			stages = append(stages, stage.NewGitLatestPatchStage(gitPatchStageOptions, baseStageOptions))
		}

		stages = appendIfExist(stages, stage.GenerateDockerInstructionsStage(imageInterfaceConfig.(*config.StapelImage), baseStageOptions))
	}

	if len(gitMappings) != 0 {
		logboek.Info.LogLnDetails("Using git stages")

		for _, s := range stages {
			s.SetGitMappings(gitMappings)
		}
	}

	image.SetStages(stages)

	return nil
}

func generateGitMappings(imageBaseConfig *config.StapelImageBase, c *Conveyor) ([]*stage.GitMapping, error) {
	var gitMappings []*stage.GitMapping

	if len(imageBaseConfig.Git.Local) != 0 {
		localGitRepo := c.GetLocalGitRepo()
		if localGitRepo == nil {
			localGitRepo, err := git_repo.OpenLocalRepo("own", c.projectDir)
			if err != nil {
				return nil, fmt.Errorf("unable to open local repo %s: %s", c.projectDir, err)
			}

			if localGitRepo == nil {
				return nil, errors.New("local git mapping is used but project git repository is not found")
			}

			c.SetLocalGitRepo(localGitRepo)
		}
	}

	for _, localGitMappingConfig := range imageBaseConfig.Git.Local {
		gitMappings = append(gitMappings, gitLocalPathInit(localGitMappingConfig, c.GetLocalGitRepo(), imageBaseConfig.Name, c))
	}

	for _, remoteGitMappingConfig := range imageBaseConfig.Git.Remote {
		remoteGitRepo, exist := c.remoteGitRepos[remoteGitMappingConfig.Name]
		if !exist {
			remoteGitRepo = &git_repo.Remote{
				Base: git_repo.Base{Name: remoteGitMappingConfig.Name},
				Url:  remoteGitMappingConfig.Url,
			}

			if err := logboek.Info.LogProcess(fmt.Sprintf("Refreshing %s repository", remoteGitMappingConfig.Name), logboek.LevelLogProcessOptions{}, func() error {
				return remoteGitRepo.CloneAndFetch()
			}); err != nil {
				return nil, err
			}

			c.remoteGitRepos[remoteGitMappingConfig.Name] = remoteGitRepo
		}

		gitMappings = append(gitMappings, gitRemoteArtifactInit(remoteGitMappingConfig, remoteGitRepo, imageBaseConfig.Name, c))
	}

	var res []*stage.GitMapping

	if len(gitMappings) != 0 {
		err := logboek.Info.LogProcess(fmt.Sprintf("Initializing git mappings"), logboek.LevelLogProcessOptions{}, func() error {
			resGitMappings, err := filterAndLogGitMappings(gitMappings)
			if err != nil {
				return err
			}

			res = resGitMappings

			return nil
		})

		if err != nil {
			return nil, err
		}
	}

	return res, nil
}

func filterAndLogGitMappings(gitMappings []*stage.GitMapping) ([]*stage.GitMapping, error) {
	var res []*stage.GitMapping

	for ind, gitMapping := range gitMappings {
		if err := logboek.Info.LogProcess(fmt.Sprintf("[%d] git mapping from %s repository", ind, gitMapping.Name), logboek.LevelLogProcessOptions{}, func() error {
			withTripleIndent := func(f func()) {
				if logboek.Info.IsAccepted() {
					logboek.IndentUp()
					logboek.IndentUp()
					logboek.IndentUp()
				}

				f()

				if logboek.Info.IsAccepted() {
					logboek.IndentDown()
					logboek.IndentDown()
					logboek.IndentDown()
				}
			}

			withTripleIndent(func() {
				logboek.Info.LogFDetails("add: %s\n", gitMapping.Add)
				logboek.Info.LogFDetails("to: %s\n", gitMapping.To)

				if len(gitMapping.IncludePaths) != 0 {
					logboek.Info.LogFDetails("includePaths: %+v\n", gitMapping.IncludePaths)
				}

				if len(gitMapping.ExcludePaths) != 0 {
					logboek.Info.LogFDetails("excludePaths: %+v\n", gitMapping.ExcludePaths)
				}

				if gitMapping.Commit != "" {
					logboek.Info.LogFDetails("commit: %s\n", gitMapping.Commit)
				}

				if gitMapping.Branch != "" {
					logboek.Info.LogFDetails("branch: %s\n", gitMapping.Branch)
				}

				if gitMapping.Owner != "" {
					logboek.Info.LogFDetails("owner: %s\n", gitMapping.Owner)
				}

				if gitMapping.Group != "" {
					logboek.Info.LogFDetails("group: %s\n", gitMapping.Group)
				}

				if len(gitMapping.StagesDependencies) != 0 {
					logboek.Info.LogLnDetails("stageDependencies:")

					for s, values := range gitMapping.StagesDependencies {
						if len(values) != 0 {
							logboek.Info.LogFDetails("  %s: %v\n", s, values)
						}
					}

				}
			})

			logboek.Info.LogLn()

			commit, err := gitMapping.LatestCommit()
			if err != nil {
				return fmt.Errorf("unable to get commit of repo '%s': %s", gitMapping.GitRepo().GetName(), err)
			}

			logboek.Info.LogFDetails("Commit %s will be used\n", commit)

			res = append(res, gitMapping)

			return nil
		}); err != nil {
			return nil, err
		}
	}

	return res, nil
}

func gitRemoteArtifactInit(remoteGitMappingConfig *config.GitRemote, remoteGitRepo *git_repo.Remote, imageName string, c *Conveyor) *stage.GitMapping {
	gitMapping := baseGitMappingInit(remoteGitMappingConfig.GitLocalExport, imageName, c)

	gitMapping.Tag = remoteGitMappingConfig.Tag
	gitMapping.Commit = remoteGitMappingConfig.Commit
	gitMapping.Branch = remoteGitMappingConfig.Branch

	gitMapping.Name = remoteGitMappingConfig.Name

	gitMapping.GitRepoInterface = remoteGitRepo

	gitMapping.GitRepoCache = c.GetGitRepoCache(remoteGitRepo.GetName())

	return gitMapping
}

func gitLocalPathInit(localGitMappingConfig *config.GitLocal, localGitRepo *git_repo.Local, imageName string, c *Conveyor) *stage.GitMapping {
	gitMapping := baseGitMappingInit(localGitMappingConfig.GitLocalExport, imageName, c)

	gitMapping.Name = "own"

	gitMapping.GitRepoInterface = localGitRepo

	gitMapping.GitRepoCache = c.GetGitRepoCache(localGitRepo.GetName())

	return gitMapping
}

func baseGitMappingInit(local *config.GitLocalExport, imageName string, c *Conveyor) *stage.GitMapping {
	var stageDependencies map[stage.StageName][]string
	if local.StageDependencies != nil {
		stageDependencies = stageDependenciesToMap(local.GitMappingStageDependencies())
	}

	gitMapping := &stage.GitMapping{
		PatchesDir:           getImagePatchesDir(imageName, c),
		ContainerPatchesDir:  getImagePatchesContainerDir(c),
		ArchivesDir:          getImageArchivesDir(imageName, c),
		ContainerArchivesDir: getImageArchivesContainerDir(c),
		ScriptsDir:           getImageScriptsDir(imageName, c),
		ContainerScriptsDir:  getImageScriptsContainerDir(c),

		Add:                local.GitMappingAdd(),
		To:                 local.GitMappingTo(),
		ExcludePaths:       local.GitMappingExcludePath(),
		IncludePaths:       local.GitMappingIncludePaths(),
		Owner:              local.Owner,
		Group:              local.Group,
		StagesDependencies: stageDependencies,
	}

	return gitMapping
}

func getImagePatchesDir(imageName string, c *Conveyor) string {
	return filepath.Join(c.tmpDir, imageName, "patch")
}

func getImagePatchesContainerDir(c *Conveyor) string {
	return path.Join(c.containerWerfDir, "patch")
}

func getImageArchivesDir(imageName string, c *Conveyor) string {
	return filepath.Join(c.tmpDir, imageName, "archive")
}

func getImageArchivesContainerDir(c *Conveyor) string {
	return path.Join(c.containerWerfDir, "archive")
}

func getImageScriptsDir(imageName string, c *Conveyor) string {
	return filepath.Join(c.tmpDir, imageName, "scripts")
}

func getImageScriptsContainerDir(c *Conveyor) string {
	return path.Join(c.containerWerfDir, "scripts")
}

func stageDependenciesToMap(sd *config.StageDependencies) map[stage.StageName][]string {
	result := map[stage.StageName][]string{
		stage.Install:     sd.Install,
		stage.BeforeSetup: sd.BeforeSetup,
		stage.Setup:       sd.Setup,
	}

	return result
}

func appendIfExist(stages []stage.Interface, stage stage.Interface) []stage.Interface {
	if !reflect.ValueOf(stage).IsNil() {
		logboek.Info.LogFDetails("Using stage %s\n", stage.Name())
		return append(stages, stage)
	}

	return stages
}

func prepareImageBasedOnImageFromDockerfile(imageFromDockerfileConfig *config.ImageFromDockerfile, c *Conveyor) (*Image, error) {
	img := &Image{}
	img.name = imageFromDockerfileConfig.Name
	img.isDockerfileImage = true

	contextDir := filepath.Join(c.projectDir, imageFromDockerfileConfig.Context)

	relContextDir, err := filepath.Rel(c.projectDir, contextDir)
	if err != nil || relContextDir == ".." || strings.HasPrefix(relContextDir, ".."+string(os.PathSeparator)) {
		return nil, fmt.Errorf("unsupported context folder %s.\nOnly context folder specified inside project directory %s supported", contextDir, c.projectDir)
	}

	exist, err := util.DirExists(contextDir)
	if err != nil {
		return nil, err
	} else if !exist {
		return nil, fmt.Errorf("context folder %s is not found", contextDir)
	}

	dockerfilePath := filepath.Join(c.projectDir, imageFromDockerfileConfig.Dockerfile)
	relDockerfilePath, err := filepath.Rel(c.projectDir, dockerfilePath)
	if err != nil || relDockerfilePath == "." || relDockerfilePath == ".." || strings.HasPrefix(relDockerfilePath, ".."+string(os.PathSeparator)) {
		return nil, fmt.Errorf("unsupported dockerfile %s.\n Only dockerfile specified inside project directory %s supported", dockerfilePath, c.projectDir)
	}

	exist, err = util.FileExists(dockerfilePath)
	if err != nil {
		return nil, err
	} else if !exist {
		return nil, fmt.Errorf("dockerfile %s is not found", dockerfilePath)
	}

	dockerignorePatterns, err := build.ReadDockerignore(contextDir)
	if err != nil {
		return nil, err
	}

	dockerignorePatternMatcher, err := fileutils.NewPatternMatcher(dockerignorePatterns)
	if err != nil {
		return nil, err
	}

	if relContextDir == "." {
		relContextDir = ""
	}
	dockerignorePathMatcher := path_matcher.NewDockerfileIgnorePathMatcher(relContextDir, dockerignorePatternMatcher, false)

	localGitRepo := c.GetLocalGitRepo()
	if localGitRepo == nil {
		localGitRepo, err = git_repo.OpenLocalRepo("own", c.projectDir)
		if err != nil {
			return nil, fmt.Errorf("unable to open local repo %s: %s", c.projectDir, err)
		}

		if localGitRepo != nil {
			c.SetLocalGitRepo(localGitRepo)
		}
	}

	localGitRepo = c.GetLocalGitRepo()
	if localGitRepo != nil {
		exist, err = localGitRepo.IsHeadReferenceExist()
		if err != nil {
			return nil, fmt.Errorf("git head reference failed: %s", err)
		}

		if !exist {
			logboek.Debug.LogLnWithCustomStyle(
				logboek.StyleByName(logboek.FailStyleName),
				"git repository reference is not found",
			)
			localGitRepo = nil
		}
	}

	data, err := ioutil.ReadFile(dockerfilePath)
	if err != nil {
		return nil, err
	}

	p, err := parser.Parse(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}

	dockerStages, dockerMetaArgs, err := instructions.Parse(p.AST)
	if err != nil {
		return nil, err
	}

	resolveDockerStagesFromValue(dockerStages)

	dockerTargetIndex, err := getDockerTargetStageIndex(dockerStages, imageFromDockerfileConfig.Target)
	if err != nil {
		return nil, err
	}

	dockerTargetStage := dockerStages[dockerTargetIndex]

	dockerArgsHash := map[string]string{}
	var dockerMetaArgsString []string
	for _, arg := range dockerMetaArgs {
		dockerArgsHash[arg.Key] = arg.ValueString()
	}

	for key, valueInterf := range imageFromDockerfileConfig.Args {
		value := fmt.Sprintf("%v", valueInterf)
		dockerArgsHash[key] = value
	}

	for key, value := range dockerArgsHash {
		dockerMetaArgsString = append(dockerMetaArgsString, fmt.Sprintf("%s=%v", key, value))
	}

	shlex := shell.NewLex(parser.DefaultEscapeToken)
	resolvedBaseName, err := shlex.ProcessWord(dockerTargetStage.BaseName, dockerMetaArgsString)
	if err != nil {
		return nil, err
	}

	if err := handleImageFromName(resolvedBaseName, false, img, c); err != nil {
		return nil, err
	}

	baseStageOptions := &stage.NewBaseStageOptions{
		ImageName:   imageFromDockerfileConfig.Name,
		ProjectName: c.werfConfig.Meta.Project,
	}

	dockerfileStage := stage.GenerateDockerfileStage(
		stage.NewDockerRunArgs(
			dockerfilePath,
			imageFromDockerfileConfig.Target,
			contextDir,
			imageFromDockerfileConfig.Args,
			imageFromDockerfileConfig.AddHost,
		),
		stage.NewDockerStages(dockerStages, dockerArgsHash, dockerTargetIndex),
		stage.NewContextChecksum(c.projectDir, dockerignorePathMatcher, localGitRepo),
		baseStageOptions,
	)

	img.stages = append(img.stages, dockerfileStage)

	logboek.Info.LogFDetails("Using stage %s\n", dockerfileStage.Name())

	return img, nil
}

func resolveDockerStagesFromValue(stages []instructions.Stage) {
	nameToIndex := make(map[string]string)
	for i, s := range stages {
		name := strings.ToLower(s.Name)
		index := strconv.Itoa(i)
		if name != index {
			nameToIndex[name] = index
		}

		for _, cmd := range s.Commands {
			switch c := cmd.(type) {
			case *instructions.CopyCommand:
				if c.From != "" {
					from := strings.ToLower(c.From)
					if val, ok := nameToIndex[from]; ok {
						c.From = val
					}
				}
			}
		}
	}
}

func getDockerTargetStageIndex(dockerStages []instructions.Stage, dockerTargetStage string) (int, error) {
	if dockerTargetStage == "" {
		return len(dockerStages) - 1, nil
	}

	for i, s := range dockerStages {
		if s.Name == dockerTargetStage {
			return i, nil
		}
	}

	return -1, fmt.Errorf("%s is not a valid target build stage", dockerTargetStage)
}
