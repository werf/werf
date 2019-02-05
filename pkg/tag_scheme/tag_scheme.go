package tag_scheme

type TagScheme string

const (
	CustomScheme    TagScheme = "custom"
	GitTagScheme    TagScheme = "git_tag"
	GitBranchScheme TagScheme = "git_branch"
	GitCommitScheme TagScheme = "git_commit"
)
