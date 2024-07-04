package config

type ImageInterface interface {
	GetName() string
	IsStapel() bool
	IsFinal() bool
	Platform() []string

	dependsOn() DependsOn
	rawDoc() *doc
}
