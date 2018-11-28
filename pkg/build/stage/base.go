package stage

func newBaseStage() *BaseStage {
	return &BaseStage{}
}

type BaseStage struct {
	signature string
	image     Image
}

func (s *BaseStage) Name() string {
	panic("method must be implemented!")
}

func (s *BaseStage) GetDependencies(_ Cache) string {
	panic("method must be implemented!")
}

func (s *BaseStage) GetContext(_ Cache) string {
	return ""
}

func (s *BaseStage) GetRelatedStageName() string {
	return ""
}

func (s *BaseStage) SetSignature(signature string) {
	s.signature = signature
}

func (s *BaseStage) GetSignature() string {
	return s.signature
}

func (s *BaseStage) SetImage(image Image) {
	s.image = image
}

func (s *BaseStage) GetImage() Image {
	return s.image
}
