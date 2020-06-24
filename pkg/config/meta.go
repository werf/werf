package config

type Meta struct {
	ConfigVersion   int
	Project         string
	DeployTemplates MetaDeployTemplates
	Cleanup         MetaCleanup
}
