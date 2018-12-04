package build

import (
	"fmt"
	"path"

	"github.com/flant/dapp/pkg/config"
	"github.com/flant/dapp/pkg/image"
	"github.com/flant/dapp/pkg/lock"
)

type Conveyor struct {
	Dappfile     []*config.Dimg
	DimgsInOrder []*Dimg

	// Все кеширование тут
	// Инициализируется конфигом dappfile (все dimgs, все artifacts)
	// Предоставляет интерфейс для получения инфы по образам связанным с dappfile. ???
	// SetEnabledDimgs(...)
	// defaultPhases() -> []Phase

	// Build()
	// Tag()
	// Push()
	// BP()

	stageImages      map[string]*image.Stage
	dockerAuthorizer DockerAuthorizer

	ProjectName string

	ProjectDir       string
	ProjectBuildDir  string
	ContainerDappDir string
	TmpDir           string

	SSHAuthSock string
}

type DockerAuthorizer interface {
	LoginBaseImage(repo string) error
}

func NewConveyor(dappfile []*config.Dimg, projectDir, projectName, buildDir, tmpDir, sshAuthSock string, authorizer DockerAuthorizer) *Conveyor {
	return &Conveyor{
		Dappfile:         dappfile,
		ProjectDir:       projectDir,
		ProjectName:      projectName,
		ProjectBuildDir:  buildDir,
		ContainerDappDir: "/.dapp",
		TmpDir:           tmpDir,
		SSHAuthSock:      sshAuthSock,
		stageImages:      make(map[string]*image.Stage),
		dockerAuthorizer: authorizer,
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

	lockImagesName := fmt.Sprintf("%s.images", c.ProjectName)
	err = lock.Lock(lockImagesName, lock.LockOptions{ReadOnly: true})
	if err != nil {
		return fmt.Errorf("error locking %s: %s", lockImagesName, err)
	}
	defer lock.Unlock(lockImagesName)

	for _, phase := range phases {
		err := phase.Run(c)
		if err != nil {
			return err
		}
	}

	return nil
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

func (c *Conveyor) GetDimgTmpDir(dimgName string) string {
	return path.Join(c.TmpDir, dimgName)
}

type stubDockerAuthorizer struct{}

func (a *stubDockerAuthorizer) LoginBaseImage(repo string) error {
	fmt.Printf("Called login for base image repo %s\n", repo)
	return nil
}
