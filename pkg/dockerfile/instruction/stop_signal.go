package instruction

type StopSignal struct {
	*Base

	Signal string
}

func NewStopSignal(raw, signal string) *StopSignal {
	return &StopSignal{Base: NewBase(raw), Signal: signal}
}

func (i *StopSignal) Name() string {
	return "STOPSIGNAL"
}
