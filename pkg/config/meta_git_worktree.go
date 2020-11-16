package config

type MetaGitWorktree struct {
	ForceShallowClone                  *bool
	AllowUnshallow                     *bool
	AllowFetchingOriginBranchesAndTags *bool
}

func (obj MetaGitWorktree) GetForceShallowClone() bool {
	if obj.ForceShallowClone != nil {
		return *obj.ForceShallowClone
	} else {
		return false
	}
}

func (obj MetaGitWorktree) GetAllowUnshallow() bool {
	if obj.AllowUnshallow != nil {
		return *obj.AllowUnshallow
	} else {
		return true
	}
}

func (obj MetaGitWorktree) GetAllowFetchingOriginBranchesAndTags() bool {
	if obj.AllowFetchingOriginBranchesAndTags != nil {
		return *obj.AllowFetchingOriginBranchesAndTags
	} else {
		return true
	}
}
