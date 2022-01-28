package config

type StapelImageInterface interface {
	ImageInterface

	ImageBaseConfig() *StapelImageBase
	IsArtifact() bool

	imports() []*Import
	dependencies() []*Dependency
}
