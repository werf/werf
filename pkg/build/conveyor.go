package build

import (
	"fmt"
	"path"

	"github.com/flant/dapp/pkg/build/stage"
	"github.com/flant/dapp/pkg/config"
	"github.com/flant/dapp/pkg/git_repo"
	"github.com/flant/dapp/pkg/image"
	"github.com/flant/dapp/pkg/lock"
)

type Conveyor struct {
	Dappfile           []*config.Dimg
	DimgsInOrder       []*Dimg
	DimgNamesToProcess []string

	stageImages                   map[string]*image.Stage
	buildingGAStageNameByDimgName map[string]stage.StageName
	remoteGitRepos                map[string]*git_repo.Remote
	dockerAuthorizer              DockerAuthorizer

	ProjectName string

	ProjectDir       string
	ProjectBuildDir  string
	ContainerDappDir string
	TmpDir           string

	SSHAuthSock string
}

type DockerAuthorizer interface {
	LoginForPull(repo string) error
	LoginForPush(repo string) error
}

func NewConveyor(dappfile []*config.Dimg, dimgNamesToProcess []string, projectDir, projectName, buildDir, tmpDir, sshAuthSock string, authorizer DockerAuthorizer) *Conveyor {
	return &Conveyor{
		Dappfile:                      dappfile,
		DimgNamesToProcess:            dimgNamesToProcess,
		ProjectDir:                    projectDir,
		ProjectName:                   projectName,
		ProjectBuildDir:               buildDir,
		ContainerDappDir:              "/.dapp",
		TmpDir:                        tmpDir,
		SSHAuthSock:                   sshAuthSock,
		stageImages:                   make(map[string]*image.Stage),
		dockerAuthorizer:              authorizer,
		buildingGAStageNameByDimgName: make(map[string]stage.StageName),
		remoteGitRepos:                make(map[string]*git_repo.Remote),
	}
}

type Phase interface {
	Run(*Conveyor) error
}

func (c *Conveyor) Build() error {
	var err error

	var phases []Phase
	phases = append(phases, NewInitializationPhase())
	phases = append(phases, NewSignaturesPhase())
	phases = append(phases, NewRenewPhase())
	phases = append(phases, NewPrepareImagesPhase())
	phases = append(phases, NewBuildPhase())

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

func (c *Conveyor) BP(repo string, opts PushOptions) error {
	var err error

	var phases []Phase
	phases = append(phases, NewInitializationPhase())
	phases = append(phases, NewSignaturesPhase())
	phases = append(phases, NewRenewPhase())
	phases = append(phases, NewPrepareImagesPhase())
	phases = append(phases, NewBuildPhase())
	phases = append(phases, NewPushPhase(repo, opts))

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
	lockName := fmt.Sprintf("%s.images", c.ProjectName)
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

func (c *Conveyor) GetDimg(name string) *Dimg {
	for _, dimg := range c.DimgsInOrder {
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
	return path.Join(c.TmpDir, dimgName)
}

type stubDockerAuthorizer struct{}

func (a *stubDockerAuthorizer) LoginBaseImage(repo string) error {
	fmt.Printf("Called login for base image repo %s\n", repo)
	return nil
}
