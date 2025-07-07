package config

//go:generate mockgen -source image_interface.go -package config -destination image_interface_mock.go

type ImageInterface interface {
	CacheVersion() string
	GetName() string
	IsStapel() bool
	IsFinal() bool
	Platform() []string
	Sbom() *Sbom

	dependsOn() DependsOn
	rawDoc() *doc
}
