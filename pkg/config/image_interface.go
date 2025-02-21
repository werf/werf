package config

type ImageInterface interface {
	CacheVersion() string
	GetName() string
	IsStapel() bool
	IsFinal() bool
	Platform() []string

	dependsOn() DependsOn
	rawDoc() *doc
}
