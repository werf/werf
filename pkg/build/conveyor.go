package build

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"strings"
	"sync"

	"github.com/gookit/color"
	"github.com/moby/buildkit/frontend/dockerfile/dockerignore"
	"github.com/moby/buildkit/frontend/dockerfile/instructions"
	"github.com/moby/buildkit/frontend/dockerfile/parser"

	"github.com/werf/logboek"
	stylePkg "github.com/werf/logboek/pkg/style"
	"github.com/werf/logboek/pkg/types"
	"github.com/werf/werf/pkg/build/dockerfile_helpers"
	"github.com/werf/werf/pkg/build/import_server"
	"github.com/werf/werf/pkg/build/stage"
	"github.com/werf/werf/pkg/config"
	"github.com/werf/werf/pkg/container_backend"
	"github.com/werf/werf/pkg/git_repo"
	"github.com/werf/werf/pkg/giterminism_manager"
	imagePkg "github.com/werf/werf/pkg/image"
	"github.com/werf/werf/pkg/logging"
	"github.com/werf/werf/pkg/path_matcher"
	"github.com/werf/werf/pkg/storage"
	"github.com/werf/werf/pkg/storage/manager"
	"github.com/werf/werf/pkg/util"
	"github.com/werf/werf/pkg/util/parallel"
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

	images    []*Image
	imageSets [][]*Image

	stageImages        map[string]*stage.StageImage
	giterminismManager giterminism_manager.Interface
	remoteGitRepos     map[string]*git_repo.Remote

	tmpDir string

	ContainerBackend container_backend.ContainerBackend

	StorageLockManager storage.LockManager
	StorageManager     manager.StorageManagerInterface

	onTerminateFuncs []func() error
	importServers    map[string]import_server.ImportServer

	ConveyorOptions

	mutex            sync.Mutex
	serviceRWMutex   map[string]*sync.RWMutex
	stageDigestMutex map[string]*sync.Mutex
}

type ConveyorOptions struct {
	Parallel                        bool
	ParallelTasksLimit              int64
	LocalGitRepoVirtualMergeOptions stage.VirtualMergeOptions
}

func NewConveyor(werfConfig *config.WerfConfig, giterminismManager giterminism_manager.Interface, imageNamesToProcess []string, projectDir, baseTmpDir, sshAuthSock string, containerBackend container_backend.ContainerBackend, storageManager manager.StorageManagerInterface, storageLockManager storage.LockManager, opts ConveyorOptions) *Conveyor {
	return &Conveyor{
		werfConfig:          werfConfig,
		imageNamesToProcess: imageNamesToProcess,

		projectDir:       projectDir,
		containerWerfDir: "/.werf",
		baseTmpDir:       baseTmpDir,

		sshAuthSock: sshAuthSock,

		giterminismManager: giterminismManager,

		stageImages:            make(map[string]*stage.StageImage),
		baseImagesRepoIdsCache: make(map[string]string),
		baseImagesRepoErrCache: make(map[string]error),
		images:                 []*Image{},
		imageSets:              [][]*Image{},
		remoteGitRepos:         make(map[string]*git_repo.Remote),
		tmpDir:                 filepath.Join(baseTmpDir, util.GenerateConsistentRandomString(10)),
		importServers:          make(map[string]import_server.ImportServer),

		ContainerBackend:   containerBackend,
		StorageLockManager: storageLockManager,
		StorageManager:     storageManager,

		ConveyorOptions: opts,

		serviceRWMutex:   map[string]*sync.RWMutex{},
		stageDigestMutex: map[string]*sync.Mutex{},
	}
}

func (c *Conveyor) getServiceRWMutex(service string) *sync.RWMutex {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	rwMutex, ok := c.serviceRWMutex[service]
	if !ok {
		rwMutex = &sync.RWMutex{}
		c.serviceRWMutex[service] = rwMutex
	}

	return rwMutex
}

func (c *Conveyor) UseLegacyStapelBuilder(cr container_backend.ContainerBackend) bool {
	return !cr.HasStapelBuildSupport()
}

func (c *Conveyor) IsBaseImagesRepoIdsCacheExist(key string) bool {
	c.getServiceRWMutex("BaseImagesRepoIdsCache").RLock()
	defer c.getServiceRWMutex("BaseImagesRepoIdsCache").RUnlock()

	_, exist := c.baseImagesRepoIdsCache[key]
	return exist
}

func (c *Conveyor) GetBaseImagesRepoIdsCache(key string) string {
	c.getServiceRWMutex("BaseImagesRepoIdsCache").RLock()
	defer c.getServiceRWMutex("BaseImagesRepoIdsCache").RUnlock()

	return c.baseImagesRepoIdsCache[key]
}

func (c *Conveyor) SetBaseImagesRepoIdsCache(key, value string) {
	c.getServiceRWMutex("BaseImagesRepoIdsCache").Lock()
	defer c.getServiceRWMutex("BaseImagesRepoIdsCache").Unlock()

	c.baseImagesRepoIdsCache[key] = value
}

func (c *Conveyor) IsBaseImagesRepoErrCacheExist(key string) bool {
	c.getServiceRWMutex("GetBaseImagesRepoErrCache").RLock()
	defer c.getServiceRWMutex("GetBaseImagesRepoErrCache").RUnlock()

	_, exist := c.baseImagesRepoErrCache[key]
	return exist
}

func (c *Conveyor) GetBaseImagesRepoErrCache(key string) error {
	c.getServiceRWMutex("GetBaseImagesRepoErrCache").RLock()
	defer c.getServiceRWMutex("GetBaseImagesRepoErrCache").RUnlock()

	return c.baseImagesRepoErrCache[key]
}

func (c *Conveyor) SetBaseImagesRepoErrCache(key string, err error) {
	c.getServiceRWMutex("BaseImagesRepoErrCache").Lock()
	defer c.getServiceRWMutex("BaseImagesRepoErrCache").Unlock()

	c.baseImagesRepoErrCache[key] = err
}

func (c *Conveyor) GetStageDigestMutex(stage string) *sync.Mutex {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	m, ok := c.stageDigestMutex[stage]
	if !ok {
		m = &sync.Mutex{}
		c.stageDigestMutex[stage] = m
	}

	return m
}

func (c *Conveyor) GetLocalGitRepoVirtualMergeOptions() stage.VirtualMergeOptions {
	return c.ConveyorOptions.LocalGitRepoVirtualMergeOptions
}

func (c *Conveyor) GetImportServer(ctx context.Context, imageName, stageName string) (import_server.ImportServer, error) {
	c.getServiceRWMutex("ImportServer").Lock()
	defer c.getServiceRWMutex("ImportServer").Unlock()

	importServerName := imageName
	if stageName != "" {
		importServerName += "/" + stageName
	}
	if srv, hasKey := c.importServers[importServerName]; hasKey {
		return srv, nil
	}

	var srv *import_server.RsyncServer

	var stg stage.Interface

	if stageName != "" {
		stg = c.getImageStage(imageName, stageName)
	} else {
		stg = c.GetImage(imageName).GetLastNonEmptyStage()
	}

	if err := c.StorageManager.FetchStage(ctx, c.ContainerBackend, stg); err != nil {
		return nil, fmt.Errorf("unable to fetch stage %s: %w", stg.GetStageImage().Image.Name(), err)
	}

	if err := logboek.Context(ctx).Info().LogProcess(fmt.Sprintf("Firing up import rsync server for image %s", imageName)).
		DoError(func() error {
			var tmpDir string
			if stageName == "" {
				tmpDir = filepath.Join(c.tmpDir, "import-server", imageName)
			} else {
				tmpDir = filepath.Join(c.tmpDir, "import-server", fmt.Sprintf("%s-%s", imageName, stageName))
			}

			if err := os.MkdirAll(tmpDir, os.ModePerm); err != nil {
				return fmt.Errorf("unable to create dir %s: %w", tmpDir, err)
			}

			var dockerImageName string
			if stageName == "" {
				dockerImageName = c.GetImageNameForLastImageStage(imageName)
			} else {
				dockerImageName = c.GetImageNameForImageStage(imageName, stageName)
			}

			var err error
			srv, err = import_server.RunRsyncServer(ctx, dockerImageName, tmpDir)
			if srv != nil {
				c.AppendOnTerminateFunc(func() error {
					if err := srv.Shutdown(ctx); err != nil {
						return fmt.Errorf("unable to shutdown import server %s: %w", srv.DockerContainerName, err)
					}
					return nil
				})
			}
			if err != nil {
				return fmt.Errorf("unable to run rsync import server: %w", err)
			}
			return nil
		}); err != nil {
		return nil, err
	}

	c.importServers[importServerName] = srv

	return srv, nil
}

func (c *Conveyor) AppendOnTerminateFunc(f func() error) {
	c.onTerminateFuncs = append(c.onTerminateFuncs, f)
}

func (c *Conveyor) Terminate(ctx context.Context) error {
	var terminateErrors []error

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
		logboek.Context(ctx).Warn().LogF("%s", errMsg)

		return errors.New(errMsg)
	}

	return nil
}

func (c *Conveyor) GiterminismManager() giterminism_manager.Interface {
	return c.giterminismManager
}

func (c *Conveyor) SetRemoteGitRepo(key string, repo *git_repo.Remote) {
	c.getServiceRWMutex("RemoteGitRepo").Lock()
	defer c.getServiceRWMutex("RemoteGitRepo").Unlock()

	c.remoteGitRepos[key] = repo
}

func (c *Conveyor) GetRemoteGitRepo(key string) *git_repo.Remote {
	c.getServiceRWMutex("RemoteGitRepo").RLock()
	defer c.getServiceRWMutex("RemoteGitRepo").RUnlock()

	return c.remoteGitRepos[key]
}

type ShouldBeBuiltOptions struct {
	CustomTagFuncList []CustomTagFunc
}

func (c *Conveyor) ShouldBeBuilt(ctx context.Context, opts ShouldBeBuiltOptions) error {
	if err := c.determineStages(ctx); err != nil {
		return err
	}

	phases := []Phase{
		NewBuildPhase(c, BuildPhaseOptions{
			ShouldBeBuiltMode: true,
			BuildOptions: BuildOptions{
				CustomTagFuncList: opts.CustomTagFuncList,
			},
		}),
	}

	if err := c.runPhases(ctx, phases, false); err != nil {
		return err
	}

	return nil
}

func (c *Conveyor) FetchLastImageStage(ctx context.Context, imageName string) error {
	lastImageStage := c.GetImage(imageName).GetLastNonEmptyStage()
	return c.StorageManager.FetchStage(ctx, c.ContainerBackend, lastImageStage)
}

func (c *Conveyor) GetImageInfoGetters() (images []*imagePkg.InfoGetter) {
	for _, img := range c.images {
		if img.isArtifact {
			continue
		}

		getter := c.StorageManager.GetImageInfoGetter(img.name, img.GetLastNonEmptyStage())
		images = append(images, getter)
	}

	return
}

func (c *Conveyor) GetExportedImagesNames() []string {
	var res []string

	for _, img := range c.images {
		if img.isArtifact {
			continue
		}

		res = append(res, img.name)
	}

	return res
}

func (c *Conveyor) GetImagesEnvArray() []string {
	var envArray []string
	for _, img := range c.images {
		if img.isArtifact {
			continue
		}

		envArray = append(envArray, generateImageEnv(img.name, c.GetImageNameForLastImageStage(img.name)))
	}

	return envArray
}

func (c *Conveyor) checkContainerBackendSupported(_ context.Context) error {
	if _, isBuildah := c.ContainerBackend.(*container_backend.BuildahBackend); !isBuildah {
		return nil
	}

	imageConfigList, err := c.werfConfig.GetSpecificImages(c.imageNamesToProcess)
	if err != nil {
		return err
	}

	var nonDockerfileImages []string
	for _, processImage := range imageConfigList {
		for _, stapelImage := range c.werfConfig.StapelImages {
			if processImage.GetName() == stapelImage.Name {
				nonDockerfileImages = append(nonDockerfileImages, processImage.GetName())
				break
			}
		}

		for _, artifactImage := range c.werfConfig.Artifacts {
			if processImage.GetName() == artifactImage.Name {
				nonDockerfileImages = append(nonDockerfileImages, processImage.GetName())
				break
			}
		}
	}

	if len(nonDockerfileImages) > 0 && os.Getenv("WERF_BUILDAH_FOR_STAPEL_ENABLED") != "1" {
		return fmt.Errorf(`Unable to build stapel type images and artifacts with buildah container backend: %s

Please select only dockerfile images or delete all non-dockerfile images from your werf.yaml.

Or disable buildah runtime by unsetting WERF_BUILDAH_MODE environment variable.`, strings.Join(nonDockerfileImages, ", "))
	}

	return nil
}

func (c *Conveyor) Build(ctx context.Context, opts BuildOptions) error {
	if err := c.checkContainerBackendSupported(ctx); err != nil {
		return err
	}

	if err := c.determineStages(ctx); err != nil {
		return err
	}

	phases := []Phase{
		NewBuildPhase(c, BuildPhaseOptions{
			BuildOptions: opts,
		}),
	}

	return c.runPhases(ctx, phases, true)
}

type ExportOptions struct {
	BuildPhaseOptions  BuildPhaseOptions
	ExportPhaseOptions ExportPhaseOptions
}

func (c *Conveyor) Export(ctx context.Context, opts ExportOptions) error {
	if err := c.determineStages(ctx); err != nil {
		return err
	}

	phases := []Phase{
		NewBuildPhase(c, opts.BuildPhaseOptions),
		NewExportPhase(c, opts.ExportPhaseOptions),
	}

	return c.runPhases(ctx, phases, true)
}

func (c *Conveyor) determineStages(ctx context.Context) error {
	return logboek.Context(ctx).Info().LogProcess("Determining of stages").
		Options(func(options types.LogProcessOptionsInterface) {
			options.Style(stylePkg.Highlight())
		}).
		DoError(func() error {
			return c.doDetermineStages(ctx)
		})
}

func (c *Conveyor) doDetermineStages(ctx context.Context) error {
	imageConfigSets, err := c.werfConfig.GroupImagesByIndependentSets(c.imageNamesToProcess)
	if err != nil {
		return err
	}

	for _, iteration := range imageConfigSets {
		var imageSet []*Image

		for _, imageInterfaceConfig := range iteration {
			var img *Image
			var imageLogName string
			var style color.Style

			switch imageConfig := imageInterfaceConfig.(type) {
			case config.StapelImageInterface:
				imageLogName = logging.ImageLogProcessName(imageConfig.ImageBaseConfig().Name, imageConfig.IsArtifact())
				style = ImageLogProcessStyle(imageConfig.IsArtifact())
			case *config.ImageFromDockerfile:
				imageLogName = logging.ImageLogProcessName(imageConfig.Name, false)
				style = ImageLogProcessStyle(false)
			}

			err := logboek.Context(ctx).Info().LogProcess(imageLogName).
				Options(func(options types.LogProcessOptionsInterface) {
					options.Style(style)
				}).
				DoError(func() error {
					var err error

					switch imageConfig := imageInterfaceConfig.(type) {
					case config.StapelImageInterface:
						img, err = prepareImageBasedOnStapelImageConfig(ctx, imageConfig, c)
					case *config.ImageFromDockerfile:
						img, err = prepareImageBasedOnImageFromDockerfile(ctx, imageConfig, c)
					}

					if err != nil {
						return err
					}

					c.images = append(c.images, img)
					imageSet = append(imageSet, img)

					return nil
				})
			if err != nil {
				return err
			}
		}

		c.imageSets = append(c.imageSets, imageSet)
	}

	return nil
}

func (c *Conveyor) runPhases(ctx context.Context, phases []Phase, logImages bool) error {
	for _, phase := range phases {
		logProcess := logboek.Context(ctx).Debug().LogProcess("Phase %s -- BeforeImages()", phase.Name())
		logProcess.Start()
		if err := phase.BeforeImages(ctx); err != nil {
			logProcess.Fail()
			return fmt.Errorf("phase %s before images handler failed: %w", phase.Name(), err)
		}
		logProcess.End()
	}

	if err := c.doImages(ctx, phases, logImages); err != nil {
		return err
	}

	for _, phase := range phases {
		if err := logboek.Context(ctx).Debug().LogProcess(fmt.Sprintf("Phase %s -- AfterImages()", phase.Name())).
			DoError(func() error {
				if err := phase.AfterImages(ctx); err != nil {
					return fmt.Errorf("phase %s after images handler failed: %w", phase.Name(), err)
				}

				return nil
			}); err != nil {
			return err
		}
	}

	return nil
}

func (c *Conveyor) doImages(ctx context.Context, phases []Phase, logImages bool) error {
	if c.Parallel && len(c.images) > 1 {
		return c.doImagesInParallel(ctx, phases, logImages)
	} else {
		for _, img := range c.images {
			if err := c.doImage(ctx, img, phases, logImages); err != nil {
				return err
			}
		}
	}

	return nil
}

func (c *Conveyor) doImagesInParallel(ctx context.Context, phases []Phase, logImages bool) error {
	blockMsg := "Concurrent builds plan"
	if c.ParallelTasksLimit > 0 {
		blockMsg = fmt.Sprintf("%s (no more than %d images at the same time)", blockMsg, c.ParallelTasksLimit)
	}

	logboek.Context(ctx).LogBlock(blockMsg).
		Options(func(options types.LogBlockOptionsInterface) {
			options.Style(stylePkg.Highlight())
		}).
		Do(func() {
			for setId := range c.imageSets {
				logboek.Context(ctx).LogFHighlight("Set #%d:\n", setId)
				for _, img := range c.imageSets[setId] {
					logboek.Context(ctx).LogLnHighlight("-", img.LogDetailedName())
				}
				logboek.Context(ctx).LogOptionalLn()
			}
		})

	for setId := range c.imageSets {
		numberOfTasks := len(c.imageSets[setId])
		numberOfWorkers := int(c.ParallelTasksLimit)

		if err := parallel.DoTasks(ctx, numberOfTasks, parallel.DoTasksOptions{
			InitDockerCLIForEachWorker: true,
			MaxNumberOfWorkers:         numberOfWorkers,
			LiveOutput:                 true,
		}, func(ctx context.Context, taskId int) error {
			taskImage := c.imageSets[setId][taskId]

			var taskPhases []Phase
			for _, phase := range phases {
				taskPhases = append(taskPhases, phase.Clone())
			}

			return c.doImage(ctx, taskImage, taskPhases, logImages)
		}); err != nil {
			return err
		}
	}

	return nil
}

func (c *Conveyor) doImage(ctx context.Context, img *Image, phases []Phase, logImages bool) error {
	var imagesLogger types.ManagerInterface
	if logImages {
		imagesLogger = logboek.Context(ctx).Default()
	} else {
		imagesLogger = logboek.Context(ctx).Info()
	}

	return imagesLogger.LogProcess(img.LogDetailedName()).
		Options(func(options types.LogProcessOptionsInterface) {
			options.Style(img.LogProcessStyle())
		}).
		DoError(func() error {
			for _, phase := range phases {
				logProcess := logboek.Context(ctx).Debug().LogProcess("Phase %s -- BeforeImageStages()", phase.Name())
				logProcess.Start()
				if err := phase.BeforeImageStages(ctx, img); err != nil {
					logProcess.Fail()
					return fmt.Errorf("phase %s before image %s stages handler failed: %w", phase.Name(), img.GetLogName(), err)
				}
				logProcess.End()

				logProcess = logboek.Context(ctx).Debug().LogProcess("Phase %s -- OnImageStage()", phase.Name())
				logProcess.Start()
				for _, stg := range img.GetStages() {
					logboek.Context(ctx).Debug().LogF("Phase %s -- OnImageStage() %s %s\n", phase.Name(), img.GetLogName(), stg.LogDetailedName())
					if err := phase.OnImageStage(ctx, img, stg); err != nil {
						logProcess.Fail()
						return fmt.Errorf("phase %s on image %s stage %s handler failed: %w", phase.Name(), img.GetLogName(), stg.Name(), err)
					}
				}
				logProcess.End()

				logProcess = logboek.Context(ctx).Debug().LogProcess("Phase %s -- AfterImageStages()", phase.Name())
				logProcess.Start()
				if err := phase.AfterImageStages(ctx, img); err != nil {
					logProcess.Fail()
					return fmt.Errorf("phase %s after image %s stages handler failed: %w", phase.Name(), img.GetLogName(), err)
				}
				logProcess.End()

				logProcess = logboek.Context(ctx).Debug().LogProcess("Phase %s -- ImageProcessingShouldBeStopped()", phase.Name())
				logProcess.Start()
				if phase.ImageProcessingShouldBeStopped(ctx, img) {
					logProcess.End()
					return nil
				}
				logProcess.End()
			}

			return nil
		})
}

func (c *Conveyor) projectName() string {
	return c.werfConfig.Meta.Project
}

func (c *Conveyor) GetStageImage(name string) *stage.StageImage {
	c.getServiceRWMutex("StageImages").RLock()
	defer c.getServiceRWMutex("StageImages").RUnlock()

	return c.stageImages[name]
}

func (c *Conveyor) UnsetStageImage(name string) {
	c.getServiceRWMutex("StageImages").Lock()
	defer c.getServiceRWMutex("StageImages").Unlock()

	delete(c.stageImages, name)
}

func (c *Conveyor) SetStageImage(stageImage *stage.StageImage) {
	c.getServiceRWMutex("StageImages").Lock()
	defer c.getServiceRWMutex("StageImages").Unlock()

	c.stageImages[stageImage.Image.Name()] = stageImage
}

func (c *Conveyor) GetOrCreateStageImage(fromImage *container_backend.LegacyStageImage, name string) *stage.StageImage {
	if img := c.GetStageImage(name); img != nil {
		return img
	}

	i := container_backend.NewLegacyStageImage(fromImage, name, c.ContainerBackend)
	img := stage.NewStageImage(c.ContainerBackend, fromImage, i)

	c.SetStageImage(img)
	return img
}

func (c *Conveyor) GetImage(name string) *Image {
	for _, img := range c.images {
		if img.GetName() == name {
			return img
		}
	}

	panic(fmt.Sprintf("Image %q not found!", name))
}

func (c *Conveyor) GetImageStageContentDigest(imageName, stageName string) string {
	return c.getImageStage(imageName, stageName).GetContentDigest()
}

func (c *Conveyor) GetImageContentDigest(imageName string) string {
	return c.GetImage(imageName).GetContentDigest()
}

func (c *Conveyor) getImageStage(imageName, stageName string) stage.Interface {
	if stg := c.GetImage(imageName).GetStage(stage.StageName(stageName)); stg != nil {
		return stg
	} else {
		return c.getLastNonEmptyImageStage(imageName)
	}
}

func (c *Conveyor) getLastNonEmptyImageStage(imageName string) stage.Interface {
	// FIXME: find first existing stage after specified unexisting
	return c.GetImage(imageName).GetLastNonEmptyStage()
}

func (c *Conveyor) FetchImageStage(ctx context.Context, imageName, stageName string) error {
	return c.StorageManager.FetchStage(ctx, c.ContainerBackend, c.getImageStage(imageName, stageName))
}

func (c *Conveyor) FetchLastNonEmptyImageStage(ctx context.Context, imageName string) error {
	return c.StorageManager.FetchStage(ctx, c.ContainerBackend, c.getLastNonEmptyImageStage(imageName))
}

func (c *Conveyor) GetImageNameForLastImageStage(imageName string) string {
	return c.GetImage(imageName).GetLastNonEmptyStage().GetStageImage().Image.Name()
}

func (c *Conveyor) GetImageNameForImageStage(imageName, stageName string) string {
	return c.getImageStage(imageName, stageName).GetStageImage().Image.Name()
}

func (c *Conveyor) GetStageID(imageName string) string {
	return c.GetImage(imageName).GetStageID()
}

func (c *Conveyor) GetImageIDForLastImageStage(imageName string) string {
	return c.GetImage(imageName).GetLastNonEmptyStage().GetStageImage().Image.GetStageDescription().Info.ID
}

func (c *Conveyor) GetImageIDForImageStage(imageName, stageName string) string {
	return c.getImageStage(imageName, stageName).GetStageImage().Image.GetStageDescription().Info.ID
}

func (c *Conveyor) GetImageTmpDir(imageName string) string {
	return filepath.Join(c.tmpDir, "image", imageName)
}

func (c *Conveyor) GetImportMetadata(ctx context.Context, projectName, id string) (*storage.ImportMetadata, error) {
	return c.StorageManager.GetStagesStorage().GetImportMetadata(ctx, projectName, id)
}

func (c *Conveyor) PutImportMetadata(ctx context.Context, projectName string, metadata *storage.ImportMetadata) error {
	return c.StorageManager.GetStagesStorage().PutImportMetadata(ctx, projectName, metadata)
}

func (c *Conveyor) RmImportMetadata(ctx context.Context, projectName, id string) error {
	return c.StorageManager.GetStagesStorage().RmImportMetadata(ctx, projectName, id)
}

func prepareImageBasedOnStapelImageConfig(ctx context.Context, imageInterfaceConfig config.StapelImageInterface, c *Conveyor) (*Image, error) {
	image := &Image{}

	imageBaseConfig := imageInterfaceConfig.ImageBaseConfig()
	imageName := imageBaseConfig.Name
	imageArtifact := imageInterfaceConfig.IsArtifact()

	from, fromImageName, fromLatest := getFromFields(imageBaseConfig)

	image.name = imageName

	if from != "" {
		if err := handleImageFromName(ctx, from, fromLatest, image, c); err != nil {
			return nil, err
		}
	} else {
		image.baseImageImageName = fromImageName
	}

	image.isArtifact = imageArtifact

	err := initStages(ctx, image, imageInterfaceConfig, c)
	if err != nil {
		return nil, err
	}

	return image, nil
}

func handleImageFromName(ctx context.Context, from string, fromLatest bool, image *Image, c *Conveyor) error {
	image.baseImageName = from

	if fromLatest {
		if _, err := image.getFromBaseImageIdFromRegistry(ctx, c, image.baseImageName); err != nil {
			return err
		}
	}

	return nil
}

func getFromFields(imageBaseConfig *config.StapelImageBase) (string, string, bool) {
	var from string
	var fromImageName string

	switch {
	case imageBaseConfig.From != "":
		from = imageBaseConfig.From
	case imageBaseConfig.FromImageName != "":
		fromImageName = imageBaseConfig.FromImageName
	case imageBaseConfig.FromArtifactName != "":
		fromImageName = imageBaseConfig.FromArtifactName
	}

	return from, fromImageName, imageBaseConfig.FromLatest
}

func initStages(ctx context.Context, image *Image, imageInterfaceConfig config.StapelImageInterface, c *Conveyor) error {
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
		ScriptsDir:           getImageScriptsDir(imageName, c),
		ContainerArchivesDir: getImageArchivesContainerDir(c),
		ContainerScriptsDir:  getImageScriptsContainerDir(c),
	}

	gitPatchStageOptions := &stage.NewGitPatchStageOptions{
		ScriptsDir:           getImageScriptsDir(imageName, c),
		ContainerPatchesDir:  getImagePatchesContainerDir(c),
		ContainerArchivesDir: getImageArchivesContainerDir(c),
		ContainerScriptsDir:  getImageScriptsContainerDir(c),
	}

	gitMappings, err := generateGitMappings(ctx, imageBaseConfig, c)
	if err != nil {
		return err
	}

	gitMappingsExist := len(gitMappings) != 0

	stages = appendIfExist(ctx, stages, stage.GenerateFromStage(imageBaseConfig, image.baseImageRepoId, baseStageOptions))
	stages = appendIfExist(ctx, stages, stage.GenerateBeforeInstallStage(ctx, imageBaseConfig, baseStageOptions))
	stages = appendIfExist(ctx, stages, stage.GenerateDependenciesBeforeInstallStage(imageBaseConfig, baseStageOptions))

	if gitMappingsExist {
		stages = append(stages, stage.NewGitArchiveStage(gitArchiveStageOptions, baseStageOptions))
	}

	stages = appendIfExist(ctx, stages, stage.GenerateInstallStage(ctx, imageBaseConfig, gitPatchStageOptions, baseStageOptions))
	stages = appendIfExist(ctx, stages, stage.GenerateDependenciesAfterInstallStage(imageBaseConfig, baseStageOptions))
	stages = appendIfExist(ctx, stages, stage.GenerateBeforeSetupStage(ctx, imageBaseConfig, gitPatchStageOptions, baseStageOptions))
	stages = appendIfExist(ctx, stages, stage.GenerateDependenciesBeforeSetupStage(imageBaseConfig, baseStageOptions))
	stages = appendIfExist(ctx, stages, stage.GenerateSetupStage(ctx, imageBaseConfig, gitPatchStageOptions, baseStageOptions))
	stages = appendIfExist(ctx, stages, stage.GenerateDependenciesAfterSetupStage(imageBaseConfig, baseStageOptions))

	if !imageArtifact {
		if gitMappingsExist {
			stages = append(stages, stage.NewGitCacheStage(gitPatchStageOptions, baseStageOptions))
			stages = append(stages, stage.NewGitLatestPatchStage(gitPatchStageOptions, baseStageOptions))
		}

		stages = appendIfExist(ctx, stages, stage.GenerateDockerInstructionsStage(imageInterfaceConfig.(*config.StapelImage), baseStageOptions))
	}

	if len(gitMappings) != 0 {
		logboek.Context(ctx).Info().LogLnDetails("Using git stages")

		for _, s := range stages {
			s.SetGitMappings(gitMappings)
		}
	}

	image.SetStages(stages)

	return nil
}

func generateGitMappings(ctx context.Context, imageBaseConfig *config.StapelImageBase, c *Conveyor) ([]*stage.GitMapping, error) {
	var gitMappings []*stage.GitMapping

	if len(imageBaseConfig.Git.Local) != 0 {
		localGitRepo := c.giterminismManager.LocalGitRepo()

		if !c.werfConfig.Meta.GitWorktree.GetForceShallowClone() {
			isShallowClone, err := localGitRepo.IsShallowClone(ctx)
			if err != nil {
				return nil, fmt.Errorf("check shallow clone failed: %w", err)
			}

			if isShallowClone {
				if c.werfConfig.Meta.GitWorktree.GetAllowUnshallow() {
					if err := localGitRepo.FetchOrigin(ctx); err != nil {
						return nil, err
					}
				} else {
					logboek.Context(ctx).Warn().LogLn("The usage of shallow git clone may break reproducibility and slow down incremental rebuilds.")
					logboek.Context(ctx).Warn().LogLn("It is recommended to enable automatic unshallow of the git worktree with gitWorktree.allowUnshallow=true werf.yaml directive (which is enabled by default, http://werf.io/documentation/reference/werf_yaml.html#git-worktree).")
					logboek.Context(ctx).Warn().LogLn("If you still want to use shallow clone, then add gitWorktree.forceShallowClone=true werf.yaml directive (http://werf.io/documentation/reference/werf_yaml.html#git-worktree).")

					return nil, fmt.Errorf("shallow git clone is not allowed")
				}
			}
		}

		for _, localGitMappingConfig := range imageBaseConfig.Git.Local {
			gitMapping, err := gitLocalPathInit(ctx, localGitMappingConfig, imageBaseConfig.Name, c)
			if err != nil {
				return nil, err
			}
			gitMappings = append(gitMappings, gitMapping)
		}
	}

	for _, remoteGitMappingConfig := range imageBaseConfig.Git.Remote {
		remoteGitRepo := c.GetRemoteGitRepo(remoteGitMappingConfig.Name)
		if remoteGitRepo == nil {
			var err error
			remoteGitRepo, err = git_repo.OpenRemoteRepo(remoteGitMappingConfig.Name, remoteGitMappingConfig.Url)
			if err != nil {
				return nil, fmt.Errorf("unable to open remote git repo %s by url %s: %w", remoteGitMappingConfig.Name, remoteGitMappingConfig.Url, err)
			}

			if err := logboek.Context(ctx).Info().LogProcess(fmt.Sprintf("Refreshing %s repository", remoteGitMappingConfig.Name)).
				DoError(func() error {
					return remoteGitRepo.CloneAndFetch(ctx)
				}); err != nil {
				return nil, err
			}

			c.SetRemoteGitRepo(remoteGitMappingConfig.Name, remoteGitRepo)
		}

		gitMapping, err := gitRemoteArtifactInit(ctx, remoteGitMappingConfig, remoteGitRepo, imageBaseConfig.Name, c)
		if err != nil {
			return nil, err
		}
		gitMappings = append(gitMappings, gitMapping)
	}

	var res []*stage.GitMapping

	if len(gitMappings) != 0 {
		err := logboek.Context(ctx).Info().LogProcess("Initializing git mappings").DoError(func() error {
			resGitMappings, err := filterAndLogGitMappings(ctx, c, gitMappings)
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

func filterAndLogGitMappings(ctx context.Context, c *Conveyor, gitMappings []*stage.GitMapping) ([]*stage.GitMapping, error) {
	var res []*stage.GitMapping

	for ind, gitMapping := range gitMappings {
		if err := logboek.Context(ctx).Info().LogProcess(fmt.Sprintf("[%d] git mapping from %s repository", ind, gitMapping.Name)).DoError(func() error {
			withTripleIndent := func(f func()) {
				if logboek.Context(ctx).Info().IsAccepted() {
					logboek.Context(ctx).Streams().IncreaseIndent()
					logboek.Context(ctx).Streams().IncreaseIndent()
					logboek.Context(ctx).Streams().IncreaseIndent()
				}

				f()

				if logboek.Context(ctx).Info().IsAccepted() {
					logboek.Context(ctx).Streams().DecreaseIndent()
					logboek.Context(ctx).Streams().DecreaseIndent()
					logboek.Context(ctx).Streams().DecreaseIndent()
				}
			}

			withTripleIndent(func() {
				logboek.Context(ctx).Info().LogFDetails("add: %s\n", gitMapping.Add)
				logboek.Context(ctx).Info().LogFDetails("to: %s\n", gitMapping.To)

				if len(gitMapping.IncludePaths) != 0 {
					logboek.Context(ctx).Info().LogFDetails("includePaths: %+v\n", gitMapping.IncludePaths)
				}

				if len(gitMapping.ExcludePaths) != 0 {
					logboek.Context(ctx).Info().LogFDetails("excludePaths: %+v\n", gitMapping.ExcludePaths)
				}

				if gitMapping.Commit != "" {
					logboek.Context(ctx).Info().LogFDetails("commit: %s\n", gitMapping.Commit)
				}

				if gitMapping.Branch != "" {
					logboek.Context(ctx).Info().LogFDetails("branch: %s\n", gitMapping.Branch)
				}

				if gitMapping.Owner != "" {
					logboek.Context(ctx).Info().LogFDetails("owner: %s\n", gitMapping.Owner)
				}

				if gitMapping.Group != "" {
					logboek.Context(ctx).Info().LogFDetails("group: %s\n", gitMapping.Group)
				}

				if len(gitMapping.StagesDependencies) != 0 {
					logboek.Context(ctx).Info().LogLnDetails("stageDependencies:")

					for s, values := range gitMapping.StagesDependencies {
						if len(values) != 0 {
							logboek.Context(ctx).Info().LogFDetails("  %s: %v\n", s, values)
						}
					}
				}
			})

			logboek.Context(ctx).Info().LogLn()

			commitInfo, err := gitMapping.GetLatestCommitInfo(ctx, c)
			if err != nil {
				return fmt.Errorf("unable to get commit of repo %q: %w", gitMapping.GitRepo().GetName(), err)
			}

			if commitInfo.VirtualMerge {
				logboek.Context(ctx).Info().LogFDetails("Commit %s will be used (virtual merge of %s into %s)\n", commitInfo.Commit, commitInfo.VirtualMergeFromCommit, commitInfo.VirtualMergeIntoCommit)
			} else {
				logboek.Context(ctx).Info().LogFDetails("Commit %s will be used\n", commitInfo.Commit)
			}

			res = append(res, gitMapping)

			return nil
		}); err != nil {
			return nil, err
		}
	}

	return res, nil
}

func gitRemoteArtifactInit(ctx context.Context, remoteGitMappingConfig *config.GitRemote, remoteGitRepo *git_repo.Remote, imageName string, c *Conveyor) (*stage.GitMapping, error) {
	gitMapping := baseGitMappingInit(remoteGitMappingConfig.GitLocalExport, imageName, c)

	gitMapping.Tag = remoteGitMappingConfig.Tag
	gitMapping.Commit = remoteGitMappingConfig.Commit
	gitMapping.Branch = remoteGitMappingConfig.Branch

	gitMapping.Name = remoteGitMappingConfig.Name
	gitMapping.SetGitRepo(remoteGitRepo)

	gitMappingTo, err := makeGitMappingTo(ctx, gitMapping, remoteGitMappingConfig.GitLocalExport.GitMappingTo(), c)
	if err != nil {
		return nil, fmt.Errorf("unable to make remote git.to mapping for image %q: %w", imageName, err)
	}
	gitMapping.To = gitMappingTo

	return gitMapping, nil
}

func gitLocalPathInit(ctx context.Context, localGitMappingConfig *config.GitLocal, imageName string, c *Conveyor) (*stage.GitMapping, error) {
	gitMapping := baseGitMappingInit(localGitMappingConfig.GitLocalExport, imageName, c)

	gitMapping.Name = "own"
	gitMapping.SetGitRepo(c.giterminismManager.LocalGitRepo())

	gitMappingTo, err := makeGitMappingTo(ctx, gitMapping, localGitMappingConfig.GitLocalExport.GitMappingTo(), c)
	if err != nil {
		return nil, fmt.Errorf("unable to make local git.to mapping for image %q: %w", imageName, err)
	}
	gitMapping.To = gitMappingTo

	return gitMapping, nil
}

func baseGitMappingInit(local *config.GitLocalExport, imageName string, c *Conveyor) *stage.GitMapping {
	var stageDependencies map[stage.StageName][]string
	if local.StageDependencies != nil {
		stageDependencies = stageDependenciesToMap(local.StageDependencies)
	}

	gitMapping := stage.NewGitMapping()

	gitMapping.ContainerPatchesDir = getImagePatchesContainerDir(c)
	gitMapping.ContainerArchivesDir = getImageArchivesContainerDir(c)
	gitMapping.ScriptsDir = getImageScriptsDir(imageName, c)
	gitMapping.ContainerScriptsDir = getImageScriptsContainerDir(c)

	gitMapping.Add = local.GitMappingAdd()
	gitMapping.ExcludePaths = local.ExcludePaths
	gitMapping.IncludePaths = local.IncludePaths
	gitMapping.Owner = local.Owner
	gitMapping.Group = local.Group
	gitMapping.StagesDependencies = stageDependencies

	return gitMapping
}

func makeGitMappingTo(ctx context.Context, gitMapping *stage.GitMapping, gitMappingTo string, c *Conveyor) (string, error) {
	if gitMappingTo != "/" {
		return gitMappingTo, nil
	}

	gitRepoName := gitMapping.GitRepo().GetName()
	commitInfo, err := gitMapping.GetLatestCommitInfo(ctx, c)
	if err != nil {
		return "", fmt.Errorf("unable to get latest commit info for repo %q: %w", gitRepoName, err)
	}

	if gitMappingAddIsDir, err := gitMapping.GitRepo().IsCommitTreeEntryDirectory(ctx, commitInfo.Commit, gitMapping.Add); err != nil {
		return "", fmt.Errorf("unable to determine whether git `add: %s` is dir or file for repo %q: %w", gitMapping.Add, gitRepoName, err)
	} else if !gitMappingAddIsDir {
		return "", fmt.Errorf("for git repo %q specifying `to: /` when adding a single file from git with `add: %s` is not allowed. Fix this by changing `to: /` to `to: /%s`.", gitRepoName, gitMapping.Add, filepath.Base(gitMapping.Add))
	}

	return gitMappingTo, nil
}

func getImagePatchesContainerDir(c *Conveyor) string {
	return path.Join(c.containerWerfDir, "patch")
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

func appendIfExist(ctx context.Context, stages []stage.Interface, stage stage.Interface) []stage.Interface {
	if !reflect.ValueOf(stage).IsNil() {
		logboek.Context(ctx).Info().LogFDetails("Using stage %s\n", stage.Name())
		return append(stages, stage)
	}

	return stages
}

func prepareImageBasedOnImageFromDockerfile(ctx context.Context, imageFromDockerfileConfig *config.ImageFromDockerfile, c *Conveyor) (*Image, error) {
	img := &Image{}
	img.name = imageFromDockerfileConfig.Name
	img.isDockerfileImage = true

	for _, contextAddFile := range imageFromDockerfileConfig.ContextAddFiles {
		relContextAddFile := filepath.Join(imageFromDockerfileConfig.Context, contextAddFile)
		absContextAddFile := filepath.Join(c.projectDir, relContextAddFile)
		exist, err := util.FileExists(absContextAddFile)
		if err != nil {
			return nil, fmt.Errorf("unable to check existence of file %s: %w", absContextAddFile, err)
		}

		if !exist {
			return nil, fmt.Errorf("contextAddFile %q was not found (the path must be relative to the context %q)", contextAddFile, imageFromDockerfileConfig.Context)
		}
	}

	relDockerfilePath := filepath.Join(imageFromDockerfileConfig.Context, imageFromDockerfileConfig.Dockerfile)
	dockerfileData, err := c.giterminismManager.FileReader().ReadDockerfile(ctx, relDockerfilePath)
	if err != nil {
		return nil, err
	}

	var relDockerignorePath string
	var dockerignorePatterns []string
	for _, relContextDockerignorePath := range []string{
		imageFromDockerfileConfig.Dockerfile + ".dockerignore",
		".dockerignore",
	} {
		relDockerignorePath = filepath.Join(imageFromDockerfileConfig.Context, relContextDockerignorePath)
		if exist, err := c.giterminismManager.FileReader().IsDockerignoreExistAnywhere(ctx, relDockerignorePath); err != nil {
			return nil, err
		} else if exist {
			dockerignoreData, err := c.giterminismManager.FileReader().ReadDockerignore(ctx, relDockerignorePath)
			if err != nil {
				return nil, err
			}

			r := bytes.NewReader(dockerignoreData)
			dockerignorePatterns, err = dockerignore.ReadAll(r)
			if err != nil {
				return nil, fmt.Errorf("unable to read %q file: %w", relContextDockerignorePath, err)
			}

			break
		}
	}

	dockerignorePathMatcher := path_matcher.NewPathMatcher(path_matcher.PathMatcherOptions{
		BasePath:             filepath.Join(c.GiterminismManager().RelativeToGitProjectDir(), imageFromDockerfileConfig.Context),
		DockerignorePatterns: dockerignorePatterns,
	})

	if !dockerignorePathMatcher.IsPathMatched(relDockerfilePath) {
		exceptionRule := "!" + imageFromDockerfileConfig.Dockerfile
		logboek.Context(ctx).Warn().LogLn("WARNING: There is no way to ignore the Dockerfile due to docker limitation when building an image for a compressed context that reads from STDIN.")
		logboek.Context(ctx).Warn().LogF("WARNING: To hide this message, remove the Dockerfile ignore rule from the %q or add an exception rule %q.\n", relDockerignorePath, exceptionRule)

		dockerignorePatterns = append(dockerignorePatterns, exceptionRule)
		dockerignorePathMatcher = path_matcher.NewPathMatcher(path_matcher.PathMatcherOptions{
			BasePath:             filepath.Join(c.GiterminismManager().RelativeToGitProjectDir(), imageFromDockerfileConfig.Context),
			DockerignorePatterns: dockerignorePatterns,
		})
	}

	p, err := parser.Parse(bytes.NewReader(dockerfileData))
	if err != nil {
		return nil, err
	}

	dockerStages, dockerMetaArgs, err := instructions.Parse(p.AST)
	if err != nil {
		return nil, err
	}

	dockerfile_helpers.ResolveDockerStagesFromValue(dockerStages)

	dockerTargetIndex, err := dockerfile_helpers.GetDockerTargetStageIndex(dockerStages, imageFromDockerfileConfig.Target)
	if err != nil {
		return nil, err
	}

	ds := stage.NewDockerStages(
		dockerStages,
		util.MapStringInterfaceToMapStringString(imageFromDockerfileConfig.Args),
		dockerMetaArgs,
		dockerTargetIndex,
	)

	baseStageOptions := &stage.NewBaseStageOptions{
		ImageName:   imageFromDockerfileConfig.Name,
		ProjectName: c.werfConfig.Meta.Project,
	}

	dockerfileStage := stage.GenerateDockerfileStage(
		stage.NewDockerRunArgs(
			dockerfileData,
			imageFromDockerfileConfig.Dockerfile,
			imageFromDockerfileConfig.Target,
			imageFromDockerfileConfig.Context,
			imageFromDockerfileConfig.ContextAddFiles,
			imageFromDockerfileConfig.Args,
			imageFromDockerfileConfig.AddHost,
			imageFromDockerfileConfig.Network,
			imageFromDockerfileConfig.SSH,
		),
		ds,
		stage.NewContextChecksum(dockerignorePathMatcher),
		baseStageOptions,
		imageFromDockerfileConfig.Dependencies,
	)

	img.stages = append(img.stages, dockerfileStage)

	logboek.Context(ctx).Info().LogFDetails("Using stage %s\n", dockerfileStage.Name())

	return img, nil
}
