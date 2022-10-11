package instruction

type Volume struct {
	Volumes []string
}

func NewVolume(volumes []string) *Volume {
	return &Volume{Volumes: volumes}
}

func (i *Volume) Name() string {
	return "VOLUME"
}
