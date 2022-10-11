package instruction

type StopSignal struct {
	Signal string
}

func NewStopSignal(signal string) *StopSignal {
	return &StopSignal{Signal: signal}
}

func (i *StopSignal) Name() string {
	return "STOPSIGNAL"
}
