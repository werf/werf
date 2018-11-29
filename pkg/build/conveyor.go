package build

import (
	"path"

	"github.com/flant/dapp/pkg/build/stage"
	"github.com/flant/dapp/pkg/config"
	"github.com/flant/dapp/pkg/dapp"
	"github.com/flant/dapp/pkg/image"
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

	stageImages map[string]*image.Stage

	ProjectName       string
	ProjectPath       string
	TmpDir            string
	ContainerDappPath string
	SshAuthSock       string
}

func NewConveyor(projectName, tmpDir string) *Conveyor {
	return &Conveyor{
		ProjectName: projectName,
		TmpDir:      tmpDir,
		stageImages: make(map[string]*image.Stage),
	}
}

type Phase interface {
	Run(*Conveyor) error
}

func (c *Conveyor) Build() error {
	phases := []Phase{}
	phases = append(phases, NewInitializationPhase())
	phases = append(phases, NewSignaturesPhase())
	phases = append(phases, NewRenewPhase())
	phases = append(phases, NewPrepareImagesPhase())
	phases = append(phases, NewBuildPhase())

	for _, phase := range phases {
		err := phase.Run(c)
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *Conveyor) GetOrCreateImage(fromImage *image.Stage, name string) *image.Stage {
	if image, ok := c.stageImages[name]; ok {
		return image
	}

	image := image.NewStageImage(fromImage, name)
	c.stageImages[name] = image
	return image
}

func (c *Conveyor) GetDimg(name string) *Dimg {
	return nil
}

func (c *Conveyor) GetImage(imageName string) stage.Image {
	return nil
}

func (c *Conveyor) GetDimgsInOrder() []*Dimg {
	return nil
}

func (c *Conveyor) GetProjectName() string {
	return c.ProjectName
}

func (c *Conveyor) GetDimgSignature(dimgName string) string {
	return c.GetDimg(dimgName).LatestStage().GetSignature()
}

func (c *Conveyor) GetProjectBuildDir() string {
	return path.Join(dapp.GetHomeDir(), "build", c.ProjectName)
}

func getDimgPatchesDir(dimgName string, c *Conveyor) string {
	return path.Join(c.TmpDir, dimgName, "patch")
}

func getDimgPatchesContainerDir(c *Conveyor) string {
	return path.Join(c.ContainerDappPath, "patch")
}

func getDimgArchivesDir(dimgName string, c *Conveyor) string {
	return path.Join(c.TmpDir, dimgName, "archive")
}

func getDimgArchivesContainerDir(c *Conveyor) string {
	return path.Join(c.ContainerDappPath, "archive")
}
