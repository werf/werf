package config

type StapelImageInterface interface {
	ImageInterface
	ImageBaseConfig() *StapelImageBase
	IsArtifact() bool
}
