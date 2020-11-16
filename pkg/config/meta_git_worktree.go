package config

type MetaGitWorktree struct {
	AllowShallowClone              *bool
	AutoUnshallow                  *bool
	AutoFetchOriginBranchesAndTags *bool
}

func (obj MetaGitWorktree) GetAllowShallowClone() bool {
	if obj.AllowShallowClone != nil {
		return *obj.AllowShallowClone
	} else {
		return false
	}
}

func (obj MetaGitWorktree) GetAutoUnshallow() bool {
	if obj.AutoUnshallow != nil {
		return *obj.AutoUnshallow
	} else {
		return true
	}
}

func (obj MetaGitWorktree) GetAutoFetchOriginBranchesAndTags() bool {
	if obj.AutoFetchOriginBranchesAndTags != nil {
		return *obj.AutoFetchOriginBranchesAndTags
	} else {
		return true
	}
}
