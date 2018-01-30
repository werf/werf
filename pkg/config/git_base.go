package config

type GitBase struct {
	*ExportBase
	As                string
	StageDependencies *StageDependencies
}
