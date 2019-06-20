package config

type ImageInterface interface {
	GetName() string
}

type WerfImageInterface interface {
	GetName() string
	ImageBaseConfig() *ImageBase
	IsArtifact() bool

	imports() []*Import
}
