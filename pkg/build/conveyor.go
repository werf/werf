package build

import (
	"fmt"
	"path"
	"path/filepath"

	"github.com/flant/dapp/pkg/build/stage"
	"github.com/flant/dapp/pkg/config"
	"github.com/flant/dapp/pkg/git_repo"
	"github.com/flant/dapp/pkg/image"
	"github.com/flant/dapp/pkg/lock"
	"github.com/flant/dapp/pkg/util"
)

type Conveyor struct {
	*conveyorPermanentFields

	dimgsInOrder []*Dimg

	stageImages                   map[string]*image.Stage
	buildingGAStageNameByDimgName map[string]stage.StageName
	remoteGitRepos                map[string]*git_repo.Remote
	imagesBySignature             map[string]image.Image

	tmpDir string
}

type conveyorPermanentFields struct {
	dappfile           []*config.Dimg
	dimgNamesToProcess []string

	projectName string

	projectDir       string
	projectBuildDir  string
	containerDappDir string
	baseTmpDir       string

	dockerAuthorizer DockerAuthorizer

	sshAuthSock string
}

type DockerAuthorizer interface {
	LoginForPull(repo string) error
	LoginForPush(repo string) error
}

func NewConveyor(dappfile []*config.Dimg, dimgNamesToProcess []string, projectDir, projectName, buildDir, baseTmpDir, sshAuthSock string, authorizer DockerAuthorizer) *Conveyor {
	c := &Conveyor{
		conveyorPermanentFields: &conveyorPermanentFields{
			dappfile:           dappfile,
			dimgNamesToProcess: dimgNamesToProcess,

			projectName: projectName,

			projectDir:       projectDir,
			projectBuildDir:  buildDir,
			containerDappDir: "/.dapp",
			baseTmpDir:       baseTmpDir,

			dockerAuthorizer: authorizer,

			sshAuthSock: sshAuthSock,
		},
	}
	c.ReInitRuntimeFields()

	return c
}

func (c *Conveyor) ReInitRuntimeFields() {
	c.stageImages = make(map[string]*image.Stage)
	c.imagesBySignature = make(map[string]image.Image)

	c.buildingGAStageNameByDimgName = make(map[string]stage.StageName)

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
	for _, phase := range phases {
		err := phase.Run(c)
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *Conveyor) lockAllImagesReadOnly() (string, error) {
	lockName := fmt.Sprintf("%s.images", c.projectName)
	err := lock.Lock(lockName, lock.LockOptions{ReadOnly: true})
	if err != nil {
		return "", fmt.Errorf("error locking %s: %s", lockName, err)
	}
	return lockName, nil
}

func (c *Conveyor) GetImage(name string) *image.Stage {
	return c.stageImages[name]
}

func (c *Conveyor) GetOrCreateImage(fromImage *image.Stage, name string) *image.Stage {
	if img, ok := c.stageImages[name]; ok {
		return img
	}

	img := image.NewStageImage(fromImage, name)
	c.stageImages[name] = img
	return img
}

func (c *Conveyor) GetImageBySignature(signature string) image.Image {
	return c.imagesBySignature[signature]
}

func (c *Conveyor) SetImageBySignature(signature string, img image.Image) {
	c.imagesBySignature[signature] = img
}

func (c *Conveyor) GetDimg(name string) *Dimg {
	for _, dimg := range c.dimgsInOrder {
		if dimg.GetName() == name {
			return dimg
		}
	}

	panic(fmt.Sprintf("Dimg '%s' not found!", name))
}

func (c *Conveyor) GetDimgSignature(dimgName string) string {
	return c.GetDimg(dimgName).LatestStage().GetSignature()
}

func (c *Conveyor) GetDimgImageName(dimgName string) string {
	return c.GetDimg(dimgName).LatestStage().GetImage().Name()
}

func (c *Conveyor) GetDockerAuthorizer() DockerAuthorizer {
	return c.dockerAuthorizer
}

func (c *Conveyor) SetBuildingGAStage(dimgName string, stageName stage.StageName) {
	c.buildingGAStageNameByDimgName[dimgName] = stageName
}

func (c *Conveyor) GetBuildingGAStage(dimgName string) stage.StageName {
	stageName, ok := c.buildingGAStageNameByDimgName[dimgName]
	if !ok {
		return ""
	}

	return stageName
}

func (c *Conveyor) GetDimgTmpDir(dimgName string) string {
	return path.Join(c.tmpDir, "dimg", dimgName)
}
