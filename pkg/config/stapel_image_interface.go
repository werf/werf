package config

type StapelImageInterface interface {
	ImageInterface
	ImageBaseConfig() *StapelImageBase
	IsGitAfterPatchDisabled() bool
}
