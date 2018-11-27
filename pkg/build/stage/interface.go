package stage

type Interface interface {
	GetDependencies() string     // dependencies + builder_checksum
	GetContext() string          // context
	GetRelatedStageName() string // -> related_stage.context должен влиять на сигнатуру стадии
	SetSignature(signature string)
	GetSignature()
}
