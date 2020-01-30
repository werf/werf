package build

import (
	"fmt"
	"path/filepath"

	"github.com/flant/werf/pkg/util"

	"github.com/flant/werf/pkg/stages_storage"

	"github.com/flant/werf/pkg/build/stage"
	"github.com/flant/werf/pkg/config"
	"github.com/flant/werf/pkg/git_repo"
	"github.com/flant/werf/pkg/image"
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

	stageImages                     map[string]*image.StageImage
	buildingGitStageNameByImageName map[string]stage.StageName
	localGitRepo                    *git_repo.Local
	remoteGitRepos                  map[string]*git_repo.Remote
	imagesBySignature               map[string]image.ImageInterface

	tmpDir string

	StagesStorage            stages_storage.StagesStorage
	StagesStorageCache       stages_storage.Cache
	StagesStorageLockManager stages_storage.LockManager
}

func NewConveyor(werfConfig *config.WerfConfig, imageNamesToProcess []string, projectDir, baseTmpDir, sshAuthSock string) *Conveyor {
	c := &Conveyor{
		werfConfig:          werfConfig,
		imageNamesToProcess: imageNamesToProcess,

		projectDir:       projectDir,
		containerWerfDir: "/.werf",
		baseTmpDir:       baseTmpDir,

		sshAuthSock: sshAuthSock,

		stageImages:                     make(map[string]*image.StageImage),
		gitReposCaches:                  make(map[string]*stage.GitRepoCache),
		baseImagesRepoIdsCache:          make(map[string]string),
		baseImagesRepoErrCache:          make(map[string]error),
		imagesInOrder:                   []*Image{},
		imagesBySignature:               make(map[string]image.ImageInterface),
		buildingGitStageNameByImageName: make(map[string]stage.StageName),
		remoteGitRepos:                  make(map[string]*git_repo.Remote),
		tmpDir:                          filepath.Join(baseTmpDir, string(util.GenerateConsistentRandomString(10))),

		StagesStorage:            &stages_storage.LocalStagesStorage{},
		StagesStorageLockManager: &stages_storage.FileLockManager{},
		StagesStorageCache:       &stages_storage.MemoryCache{},
	}

	return c
}

func (c *Conveyor) Terminate() error {
	for gitRepoName, gitRepoCache := range c.gitReposCaches {
		if err := gitRepoCache.Terminate(); err != nil {
			return fmt.Errorf("unable to terminate cache of git repo '%s': %s", gitRepoName, err)
		}
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

func (c *Conveyor) BuildStages(stageRepo string, opts BuildStagesOptions) error {
	/*var phases []Phase
	phases = append(phases, NewInitializationPhase())
	phases = append(phases, NewSignaturesPhase())
	phases = append(phases, NewPrepareStagesPhase())
	phases = append(phases, NewBuildStagesPhase(stageRepo, opts))

	if err := c.StagesStorageLockManager.LockAllImagesReadOnly(c.projectName()); err != nil {
		return fmt.Errorf("error locking all images read only: %s", err)
	}
	defer c.StagesStorageLockManager.UnlockAllImages(c.projectName())

	return c.runPhases(phases)*/

	/*
				images (по зависимостям), dependantImagesByStage
				dependantImagesByStage строится в InitializationPhase, спросить у stage что она ждет.
				Количество воркеров-goroutine ограничено.
				Надо распределить images по воркерам.

				for img := range images {
					Goroutine {
			    		phases = append(phases, NewBuildStage())

				    	for phase = range phases {
		                    phase.OnStart()

					    	for stage = range stages {
						    	for img = dependantImagesByStage[stage.name] {
							    	wait <-imgChanMap[img]
			    				}
			    				phase.HandleStage(stage)
				    		}
						}

			            close(imgChanMap[img])
					} Goroutine
				}
	*/

	if err := NewInitializationPhase().Run(c); err != nil {
		return err
	}

	for _, img := range c.imagesInOrder {
		phases := []Phase{NewBuildPhase(c, BuildPhaseOptions{
			IntrospectOptions: opts.IntrospectOptions,
			ImageBuildOptions: opts.ImageBuildOptions,
		})}

		for _, phase := range phases {
			if err := phase.OnStart(img); err != nil {
				return fmt.Errorf("phase %s on start failed: %s", phase.Name(), err)
			}

			for _, stg := range img.GetStages() {
				if err := phase.HandleStage(img, stg); err != nil {
					return fmt.Errorf("phase %s failed to handle stage %s: %s", phase.Name(), stg.Name(), err)
				}
			}
		}
	}

	return nil
}

type TagOptions struct {
	CustomTags      []string
	TagsByGitTag    []string
	TagsByGitBranch []string
	TagsByGitCommit []string
}

type ImagesRepoManager interface {
	ImagesRepo() string
	ImageRepo(imageName string) string
	ImageRepoTag(imageName, tag string) string
	ImageRepoWithTag(imageName, tag string) string
}

type PublishImagesOptions struct {
	TagOptions
}

func (c *Conveyor) ShouldBeBuilt() error {
	/*
		var phases []Phase
		phases = append(phases, NewInitializationPhase())
		phases = append(phases, NewSignaturesPhase())
		phases = append(phases, NewShouldBeBuiltPhase())

		return c.runPhases(phases)
	*/

	return nil
}

func (c *Conveyor) PublishImages(imagesRepoManager ImagesRepoManager, opts PublishImagesOptions) error {
	/*
		var phases []Phase
		phases = append(phases, NewInitializationPhase())
		phases = append(phases, NewSignaturesPhase())
		phases = append(phases, NewShouldBeBuiltPhase())
		phases = append(phases, NewPublishImagesPhase(imagesRepoManager, opts))

		if err := c.StagesStorageLockManager.LockAllImagesReadOnly(c.projectName()); err != nil {
			return fmt.Errorf("error locking all images read only: %s", err)
		}
		defer c.StagesStorageLockManager.UnlockAllImages(c.projectName())

		return c.runPhases(phases)
	*/

	return nil
}

type BuildAndPublishOptions struct {
	BuildStagesOptions
	PublishImagesOptions
}

func (c *Conveyor) BuildAndPublish(stagesRepo string, imagesRepoManager ImagesRepoManager, opts BuildAndPublishOptions) error {
	/*
		var phases []Phase
		phases = append(phases, NewInitializationPhase())
		phases = append(phases, NewSignaturesPhase())
		phases = append(phases, NewPrepareStagesPhase())
		phases = append(phases, NewBuildStagesPhase(stagesRepo, opts.BuildStagesOptions))
		phases = append(phases, NewPublishImagesPhase(imagesRepoManager, opts.PublishImagesOptions))

		if err := c.StagesStorageLockManager.LockAllImagesReadOnly(c.projectName()); err != nil {
			return fmt.Errorf("error locking all images read only: %s", err)
		}
		defer c.StagesStorageLockManager.UnlockAllImages(c.projectName())

		return c.runPhases(phases)
	*/

	return nil
}

func (c *Conveyor) runPhases(phases []Phase) error {
	/*
		logboek.LogOptionalLn()

		for _, phase := range phases {
			err := phase.Run(c)

			if err != nil {
				c.StagesStorageLockManager.ReleaseAllStageLocks()
				return err
			}
		}
	*/
	return nil
}

func (c *Conveyor) projectName() string {
	return c.werfConfig.Meta.Project
}

func (c *Conveyor) GetStageImage(name string) *image.StageImage {
	return c.stageImages[name]
}

func (c *Conveyor) GetOrCreateImage(fromImage *image.StageImage, name string) *image.StageImage {
	if img, ok := c.stageImages[name]; ok {
		return img
	}

	img := image.NewStageImage(fromImage, name)
	c.stageImages[name] = img
	return img
}

func (c *Conveyor) GetImageBySignature(signature string) image.ImageInterface {
	return c.imagesBySignature[signature]
}

func (c *Conveyor) SetImageBySignature(signature string, img image.ImageInterface) {
	c.imagesBySignature[signature] = img
}

func (c *Conveyor) GetImage(name string) *Image {
	for _, img := range c.imagesInOrder {
		if img.GetName() == name {
			return img
		}
	}

	panic(fmt.Sprintf("Image '%s' not found!", name))
}

func (c *Conveyor) GetImageLatestStageSignature(imageName string) string {
	return c.GetImage(imageName).LatestStage().GetSignature()
}

func (c *Conveyor) GetImageLatestStageImageName(imageName string) string {
	return c.GetImage(imageName).LatestStage().GetImage().Name()
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
