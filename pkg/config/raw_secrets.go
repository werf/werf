package config

import (
	"fmt"
)

// TODO (iapershin)
// rawOrigin is not suitable here since image stack contains `doc` property and coudn't be used along with doc()
// refactor to use common approach
type rawParent interface {
	getDoc() *doc
}

type rawSecret struct {
	Id         string `yaml:"id,omitempty"`
	Env        string `yaml:"env,omitempty"`
	Src        string `yaml:"src,omitempty"`
	PlainValue string `yaml:"value,omitempty"`

	parent rawParent `yaml:"-"` // parent

	UnsupportedAttributes map[string]interface{} `yaml:",inline"`
}

func (s *rawSecret) UnmarshalYAML(unmarshal func(interface{}) error) error {
	if parent, ok := parentStack.Peek().(rawParent); ok {
		s.parent = parent
	}

	type plain rawSecret
	if err := unmarshal((*plain)(s)); err != nil {
		return err
	}

	if err := s.validate(); err != nil {
		return newDetailedConfigError(fmt.Sprintf("secrets validation error: %s", err.Error()), s, s.parent.getDoc())
	}

	if err := checkOverflow(s.UnsupportedAttributes, nil, s.parent.getDoc()); err != nil {
		return err
	}

	return nil
}

func (s *rawSecret) validate() error {
	if !oneOrNone([]bool{s.Env != "", s.Src != "", s.PlainValue != ""}) {
		return fmt.Errorf("secret type could be ONLY `env`, `src` or `value`")
	}
	return nil
}

func (s *rawSecret) toDirective() (Secret, error) {
	switch {
	case s.Env != "":
		return newSecretFromEnv(s)
	case s.Src != "":
		return newSecretFromSrc(s)
	case s.PlainValue != "":
		return newSecretFromPlainValue(s)
	default:
		return nil, newDetailedConfigError("unknown secret type. only `env`, `src` and `value` is supported", s, s.parent.getDoc())
	}
}
