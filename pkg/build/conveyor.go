package build

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
}

type Phase interface {
	Run(*Conveyor) error
}

func (c *Conveyor) Build() error {
	phases := []Phase{}
	phases = append(phases, NewInitilizationPhase())
	phases = append(phases, NewSignaturesPhase())
	phases = append(phases, NewSetCachedStatePhase()) // Определение состояния кеша (есть в кеше / нет в кеше)
	phases = append(phases, NewRenewPhase()) // Сброс кеша отсутствующих коммитов из-за rebase
	phases = append(phases, NewBuildPhase()) // Определение состояния кеша (есть в кеше / нет в кеше)

	for _, phase := range phases {
		err := phase.Run(c)
		if err != nil {
			return err	
		}
	}
	
	return nil
}
