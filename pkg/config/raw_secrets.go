package config

import (
	"fmt"
)

type rawSecret struct {
	Id         string `yaml:"id"`
	Env        string `yaml:"env,omitempty"`
	Src        string `yaml:"src,omitempty"`
	PlainValue string `yaml:"value,omitempty"`

	doc *doc `yaml:"-"` // parent

	UnsupportedAttributes map[string]interface{} `yaml:",inline"`
}

func (s *rawSecret) UnmarshalYAML(unmarshal func(interface{}) error) error {
	parentStack.Push(s)
	type plain rawSecret
	err := unmarshal((*plain)(s))
	parentStack.Pop()
	if err != nil {
		return fmt.Errorf("secrets parsing error: %w", err)
	}

	if err := s.validate(); err != nil {
		return fmt.Errorf("secrets validation error: %w", err)
	}

	if err := checkOverflow(s.UnsupportedAttributes, nil, s.doc); err != nil {
		return fmt.Errorf("secrets validation error: %w", err)
	}

	return nil
}

func (s *rawSecret) validate() error {
	if !oneOrNone([]bool{s.Env != "", s.Src != "", s.PlainValue != ""}) {
		return newDetailedConfigError("specify only env or src or value in secret", s, s.doc)
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
		return nil, newDetailedConfigError("secret type is not supported", s, s.doc)
	}
}
