package config

type raw struct {
	Dimgs    interface{} `yaml:"dimg"`
	Artifact *string     `yaml:"artifact,omitempty"`
	Project  *string     `yaml:"project,omitempty"`
}

func (c *raw) UnmarshalYAML(unmarshal func(interface{}) error) error {
	type plain raw
	err := unmarshal((*plain)(c))
	if err != nil {
		return err
	}

	return nil
}

func (r *raw) IsRawDimg() bool {
	if r.Dimgs != nil || r.Artifact != nil {
		return true
	}
	return false
}

func (r *raw) IsRawMeta() bool {
	if r.Project != nil {
		return true
	}
	return false
}
