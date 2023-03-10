package config

type Meta struct {
	ConfigVersion int
	Project       string
	Deploy        MetaDeploy
	Cleanup       MetaCleanup
	GitWorktree   MetaGitWorktree
	Build         MetaBuild
}
