package tag_strategy

type TagStrategy string

const (
	Custom    TagStrategy = "custom"
	GitTag    TagStrategy = "git-tag"
	GitBranch TagStrategy = "git-branch"
	GitCommit TagStrategy = "git-commit"
)
