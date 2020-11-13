package config

type Meta struct {
	ConfigVersion int
	Project       string
	MetaDeploy    MetaDeploy
	Cleanup       MetaCleanup
}
