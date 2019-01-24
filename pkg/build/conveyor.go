package build

import (
	"fmt"
	"path"
	"path/filepath"

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

	tmpDir string
}

type conveyorPermanentFields struct {
	werfConfig          *config.WerfConfig
	imageNamesToProcess []string

	projectDir       string
	projectBuildDir  string
	containerWerfDir string
	baseTmpDir       string

	dockerAuthorizer DockerAuthorizer

	sshAuthSock string
}

type DockerAuthorizer interface {
	LoginForPull(repo string) error
	LoginForPush(repo string) error
}

func NewConveyor(werfConfig *config.WerfConfig, imageNamesToProcess []string, projectDir, buildDir, baseTmpDir, sshAuthSock string, authorizer DockerAuthorizer) *Conveyor {
	c := &Conveyor{
		conveyorPermanentFields: &conveyorPermanentFields{
			werfConfig:          werfConfig,
			imageNamesToProcess: imageNamesToProcess,

			projectDir:       projectDir,
			projectBuildDir:  buildDir,
			containerWerfDir: "/.werf",
			baseTmpDir:       baseTmpDir,

			dockerAuthorizer: authorizer,

			sshAuthSock: sshAuthSock,
		},
	}
	c.ReInitRuntimeFields()

	return c
}

func (c *Conveyor) ReInitRuntimeFields() {
	c.stageImages = make(map[string]*image.StageImage)
	c.imagesBySignature = make(map[string]image.ImageInterface)

	c.buildingGitStageNameByImageName = make(map[string]stage.StageName)

	c.remoteGitRepos = make(map[string]*git_repo.Remote)

	c.tmpDir = filepath.Join(c.baseTmpDir, string(util.GenerateConsistentRandomString(10)))
}

type Phase interface {
	Run(*Conveyor) error
}

func (c *Conveyor) Build(opts BuildOptions) error {
restart:
	if err := c.build(opts); err != nil {
		if isConveyorShouldBeResetError(err) {
			c.ReInitRuntimeFields()
			goto restart
		}

		return err
	}

	return nil
}

func (c *Conveyor) build(opts BuildOptions) error {
	var err error

	var phases []Phase
	phases = append(phases, NewInitializationPhase())
	phases = append(phases, NewSignaturesPhase())
	phases = append(phases, NewRenewPhase())
	phases = append(phases, NewPrepareImagesPhase())
	phases = append(phases, NewBuildPhase(opts))

	lockName, err := c.lockAllImagesReadOnly()
	if err != nil {
		return err
	}
	defer lock.Unlock(lockName)

	return c.runPhases(phases)
}

type TagOptions struct {
	Tags            []string
	TagsByGitTag    []string
	TagsByGitBranch []string
	TagsByGitCommit []string
	TagsByCI        []string
}

type PushOptions struct {
	TagOptions
	WithStages bool
}

func (c *Conveyor) Tag(repo string, opts TagOptions) error {
	var err error

	var phases []Phase
	phases = append(phases, NewInitializationPhase())
	phases = append(phases, NewSignaturesPhase())
	phases = append(phases, NewShouldBeBuiltPhase())
	phases = append(phases, NewTagPhase(repo, opts))

	lockName, err := c.lockAllImagesReadOnly()
	if err != nil {
		return err
	}
	defer lock.Unlock(lockName)

	return c.runPhases(phases)
}

func (c *Conveyor) Push(repo string, opts PushOptions) error {
	var err error

	var phases []Phase
	phases = append(phases, NewInitializationPhase())
	phases = append(phases, NewSignaturesPhase())
	phases = append(phases, NewShouldBeBuiltPhase())
	phases = append(phases, NewPushPhase(repo, opts))

	lockName, err := c.lockAllImagesReadOnly()
	if err != nil {
		return err
	}
	defer lock.Unlock(lockName)

	return c.runPhases(phases)
}

func (c *Conveyor) BP(repo string, buildOpts BuildOptions, pushOpts PushOptions) error {
restart:
	if err := c.bp(repo, buildOpts, pushOpts); err != nil {
		if isConveyorShouldBeResetError(err) {
			c.ReInitRuntimeFields()
			goto restart
		}

		return err
	}

	return nil
}

func (c *Conveyor) bp(repo string, buildOpts BuildOptions, pushOpts PushOptions) error {
	var err error

	var phases []Phase
	phases = append(phases, NewInitializationPhase())
	phases = append(phases, NewSignaturesPhase())
	phases = append(phases, NewRenewPhase())
	phases = append(phases, NewPrepareImagesPhase())
	phases = append(phases, NewBuildPhase(buildOpts))
	phases = append(phases, NewPushPhase(repo, pushOpts))

	lockName, err := c.lockAllImagesReadOnly()
	if err != nil {
		return err
	}
	defer lock.Unlock(lockName)

	return c.runPhases(phases)
}

func (c *Conveyor) runPhases(phases []Phase) error {
	for ind, phase := range phases {
		isLastPhase := ind == len(phases)-1

		err := phase.Run(c)
		if err != nil {
			return err
		}

		if !isLastPhase {
			fmt.Println()
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

func (c *Conveyor) GetDockerAuthorizer() DockerAuthorizer {
	return c.dockerAuthorizer
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
