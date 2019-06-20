package build

import (
	"fmt"
	"path"
	"path/filepath"

	"github.com/flant/logboek"
	"github.com/flant/werf/pkg/build/stage"
	"github.com/flant/werf/pkg/config"
	"github.com/flant/werf/pkg/git_repo"
	"github.com/flant/werf/pkg/image"
	"github.com/flant/werf/pkg/lock"
	"github.com/flant/werf/pkg/util"
)

type Conveyor struct {
	*conveyorPermanentFields

	imagesInOrder []*Image

	stageImages                     map[string]*image.StageImage
	buildingGitStageNameByImageName map[string]stage.StageName
	remoteGitRepos                  map[string]*git_repo.Remote
	imagesBySignature               map[string]image.ImageInterface
	baseImagesRepoIdsCache          map[string]string
	baseImagesRepoErrCache          map[string]error
	globalLocks                     []string

	tmpDir string
}

type conveyorPermanentFields struct {
	werfConfig          *config.WerfConfig
	imageNamesToProcess []string

	projectDir       string
	containerWerfDir string
	baseTmpDir       string

	sshAuthSock string
}

func NewConveyor(werfConfig *config.WerfConfig, imageNamesToProcess []string, projectDir, baseTmpDir, sshAuthSock string) *Conveyor {
	c := &Conveyor{
		conveyorPermanentFields: &conveyorPermanentFields{
			werfConfig:          werfConfig,
			imageNamesToProcess: imageNamesToProcess,

			projectDir:       projectDir,
			containerWerfDir: "/.werf",
			baseTmpDir:       baseTmpDir,

			sshAuthSock: sshAuthSock,
		},
	}
	c.ReInitRuntimeFields()

	return c
}

func (c *Conveyor) ReInitRuntimeFields() {
	c.stageImages = make(map[string]*image.StageImage)

	c.baseImagesRepoIdsCache = make(map[string]string)

	c.baseImagesRepoErrCache = make(map[string]error)

	c.imagesBySignature = make(map[string]image.ImageInterface)

	c.buildingGitStageNameByImageName = make(map[string]stage.StageName)

	c.remoteGitRepos = make(map[string]*git_repo.Remote)

	c.tmpDir = filepath.Join(c.baseTmpDir, string(util.GenerateConsistentRandomString(10)))

	c.globalLocks = nil
}

func (c *Conveyor) AcquireGlobalLock(name string, opts lock.LockOptions) error {
	for _, lockName := range c.globalLocks {
		if lockName == name {
			return nil
		}
	}

	if err := lock.Lock(name, opts); err != nil {
		return err
	}

	c.globalLocks = append(c.globalLocks, name)

	return nil
}

func (c *Conveyor) ReleaseGlobalLock(name string) error {
	ind := -1
	for i, lockName := range c.globalLocks {
		if lockName == name {
			ind = i
			break
		}
	}

	if ind >= 0 {
		if err := lock.Unlock(name); err != nil {
			return err
		}
		c.globalLocks = append(c.globalLocks[:ind], c.globalLocks[ind+1:]...)
	}

	return nil
}

func (c *Conveyor) ReleaseAllGlobalLocks() error {
	for len(c.globalLocks) > 0 {
		var lockName string
		lockName, c.globalLocks = c.globalLocks[0], c.globalLocks[1:]
		if err := lock.Unlock(lockName); err != nil {
			return err
		}
	}

	return nil
}

type Phase interface {
	Run(*Conveyor) error
}

func (c *Conveyor) BuildStages(stageRepo string, opts BuildStagesOptions) error {
restart:
	if err := c.buildStages(stageRepo, opts); err != nil {
		if isConveyorShouldBeResetError(err) {
			c.ReleaseAllGlobalLocks()
			c.ReInitRuntimeFields()
			goto restart
		}

		return err
	}

	return nil
}

func (c *Conveyor) buildStages(stageRepo string, opts BuildStagesOptions) error {
	var err error

	var phases []Phase
	phases = append(phases, NewInitializationPhase())
	phases = append(phases, NewSignaturesPhase(true))
	phases = append(phases, NewRenewPhase())
	phases = append(phases, NewPrepareStagesPhase())
	phases = append(phases, NewBuildStagesPhase(stageRepo, opts))

	lockName, err := c.lockAllImagesReadOnly()
	if err != nil {
		return err
	}
	defer lock.Unlock(lockName)

	return c.runPhases(phases)
}

type TagOptions struct {
	CustomTags      []string
	TagsByGitTag    []string
	TagsByGitBranch []string
	TagsByGitCommit []string
}

type PublishImagesOptions struct {
	TagOptions
}

func (c *Conveyor) ShouldBeBuilt() error {
	var phases []Phase
	phases = append(phases, NewInitializationPhase())
	phases = append(phases, NewSignaturesPhase(false))
	phases = append(phases, NewShouldBeBuiltPhase())

	return c.runPhases(phases)
}

func (c *Conveyor) PublishImages(imagesRepo string, opts PublishImagesOptions) error {
	var err error

	var phases []Phase
	phases = append(phases, NewInitializationPhase())
	phases = append(phases, NewSignaturesPhase(false))
	phases = append(phases, NewShouldBeBuiltPhase())
	phases = append(phases, NewPublishImagesPhase(imagesRepo, opts))

	lockName, err := c.lockAllImagesReadOnly()
	if err != nil {
		return err
	}
	defer lock.Unlock(lockName)

	return c.runPhases(phases)
}

type BuildAndPublishOptions struct {
	BuildStagesOptions
	PublishImagesOptions
}

func (c *Conveyor) BuildAndPublish(stagesRepo, imagesRepo string, opts BuildAndPublishOptions) error {
restart:
	if err := c.buildAndPublish(stagesRepo, imagesRepo, opts); err != nil {
		if isConveyorShouldBeResetError(err) {
			c.ReInitRuntimeFields()
			goto restart
		}

		return err
	}

	return nil
}

func (c *Conveyor) buildAndPublish(stagesRepo, imagesRepo string, opts BuildAndPublishOptions) error {
	var err error

	var phases []Phase
	phases = append(phases, NewInitializationPhase())
	phases = append(phases, NewSignaturesPhase(true))
	phases = append(phases, NewRenewPhase())
	phases = append(phases, NewPrepareStagesPhase())
	phases = append(phases, NewBuildStagesPhase(stagesRepo, opts.BuildStagesOptions))
	phases = append(phases, NewPublishImagesPhase(imagesRepo, opts.PublishImagesOptions))

	lockName, err := c.lockAllImagesReadOnly()
	if err != nil {
		return err
	}
	defer lock.Unlock(lockName)

	return c.runPhases(phases)
}

func (c *Conveyor) runPhases(phases []Phase) error {
	logboek.LogOptionalLn()

	for _, phase := range phases {
		err := phase.Run(c)

		if err != nil {
			c.ReleaseAllGlobalLocks()
			return err
		}
	}

	return nil
}

func (c *Conveyor) projectName() string {
	return c.werfConfig.Meta.Project
}

func (c *Conveyor) lockAllImagesReadOnly() (string, error) {
	lockName := fmt.Sprintf("%s.images", c.projectName())
	err := lock.Lock(lockName, lock.LockOptions{ReadOnly: true})
	if err != nil {
		return "", fmt.Errorf("error locking %s: %s", lockName, err)
	}
	return lockName, nil
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
	return path.Join(c.tmpDir, "image", imageName)
}
