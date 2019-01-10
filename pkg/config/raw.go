package config

type raw struct {
	Dimgs    interface{} `yaml:"dimg"`
	Artifact string      `yaml:"artifact,omitempty"`
	Project  string      `yaml:"project,omitempty"`
}

func (r *raw) IsRawDimg() bool {
	if r.Dimgs != nil || r.Artifact != "" {
		return true
	}
	return false
}

func (r *raw) IsRawMeta() bool {
	if r.Project != "" {
		return true
	}
	return false
}
