package stage

type Interface interface {
	GetDependencies(Cache) string // dependencies + builder_checksum
	GetContext(Cache) string      // context
	GetRelatedStageName() string  // -> related_stage.context должен влиять на сигнатуру стадии

	SetSignature(signature string)
	GetSignature() string

	SetImage(Image)
	GetImage() Image
}
