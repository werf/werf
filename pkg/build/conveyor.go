package build

import "github.com/flant/dapp/pkg/build/stage"

type Conveyor struct {
	// Все кеширование тут
	// Инициализируется конфигом dappfile (все dimgs, все artifacts)
	// Предоставляет интерфейс для получения инфы по образам связанным с dappfile. ???
	// SetEnabledDimgs(...)
	// defaultPhases() -> []Phase

	// Build()
	// Tag()
	// Push()
	// BP()

	TmpDir string
}

type Phase interface {
	Run(*Conveyor) error
}

func (c *Conveyor) Build() error {
	phases := []Phase{}
	phases = append(phases, NewInitializationPhase())
	phases = append(phases, NewSignaturesPhase())
	phases = append(phases, NewDockerStatePhase())   // Определение состояния кеша (есть в кеше / нет в кеше)
	phases = append(phases, NewRenewPhase())         // Сброс кеша отсутствующих коммитов из-за rebase
	phases = append(phases, NewPrepareImagesPhase()) // Определение состояния кеша (есть в кеше / нет в кеше)
	phases = append(phases, NewBuildPhase())         // Определение состояния кеша (есть в кеше / нет в кеше)

	for _, phase := range phases {
		err := phase.Run(c)
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *Conveyor) GetOrCreateImage(fromImage stage.Interface, name string) stage.Image {
	return nil
}

func (c *Conveyor) GetDimg(name string) *stage.Dimg {
	return nil
}

func (c *Conveyor) GetImage(imageName string) stage.Image {
	return nil
}

func (c *Conveyor) GetDimgsInOrder() []*stage.Dimg {
	return nil
}
