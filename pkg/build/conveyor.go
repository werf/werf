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
  Run(*Conveyor)
}
