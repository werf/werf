package build

import "fmt"

type InitializationPhase struct{}

func NewInitializationPhase() *InitializationPhase {
	return &InitializationPhase{}
}

func (p *InitializationPhase) Run(c *Conveyor) error {
	if debug() {
		fmt.Printf("InitializationPhase.Run\n")
	}
	return nil
	// Определяем порядок на основе c.Dappfile (структура из pkg/config)
	// Инициализируем некоторый массив в Conveyor из условных GroupImage. Главное последовательность, поэтому массив.
}

// ?GroupImage? {
//   Name
//   []Image{StageName, Name, Signature}
// }
