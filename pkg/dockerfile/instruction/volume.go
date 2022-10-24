package instruction

type Volume struct {
	*Base

	Volumes []string
}

func NewVolume(raw string, volumes []string) *Volume {
	return &Volume{Base: NewBase(raw), Volumes: volumes}
}

func (i *Volume) Name() string {
	return "VOLUME"
}
