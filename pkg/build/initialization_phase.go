package build

type InitializationPhase struct {}

func (p *InitializationPhase) Run(c *Conveyor) error {
   // Определяем порядок на основе c.Dappfile (структура из pkg/config)
   // Инициализируем некоторый массив в Conveyor из условных GroupImage. Главное последовательность, поэтому массив.
}

?GroupImage? {
  Name
  []Image{StageName, Name, Signature}
}
