package config

type ImageInterface interface {
	ImageBaseConfig() *ImageBase
	IsArtifact() bool
}
