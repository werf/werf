package config

type rawMetaGitWorktree struct {
	ForceShallowClone                  *bool `yaml:"forceShallowClone,omitempty"`
	AllowUnshallow                     *bool `yaml:"allowUnshallow,omitempty"`
	AllowFetchingOriginBranchesAndTags *bool `yaml:"allowFetchOriginBranchesAndTags,omitempty"`

	rawMeta *rawMeta

	UnsupportedAttributes map[string]interface{} `yaml:",inline"`
}

func (c *rawMetaGitWorktree) UnmarshalYAML(unmarshal func(interface{}) error) error {
	if parent, ok := parentStack.Peek().(*rawMeta); ok {
		c.rawMeta = parent
	}

	parentStack.Push(c)
	type plain rawMetaGitWorktree
	err := unmarshal((*plain)(c))
	parentStack.Pop()
	if err != nil {
		return err
	}

	if err := checkOverflow(c.UnsupportedAttributes, nil, c.rawMeta.doc); err != nil {
		return err
	}

	return nil
}

func (c *rawMetaGitWorktree) toMetaGitWorktree() MetaGitWorktree {
	obj := MetaGitWorktree{}
	obj.ForceShallowClone = c.ForceShallowClone
	obj.AllowUnshallow = c.AllowUnshallow
	obj.AllowFetchingOriginBranchesAndTags = c.AllowFetchingOriginBranchesAndTags
	return obj
}
