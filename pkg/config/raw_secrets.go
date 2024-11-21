package config

import (
	"fmt"
	"path/filepath"
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
		newDetailedConfigError("specify only env or src or value in secret", s, s.doc)
	}
	return nil
}

func (s *rawSecret) toDirective() (Secret, error) {
	switch {
	case s.Env != "":
		if s.Id == "" {
			s.Id = s.Env
		}
		return &SecretFromEnv{
			Id:    s.Id,
			Value: s.Env,
		}, nil
	case s.Src != "":
		if s.Id == "" {
			s.Id = filepath.Base(s.Src)
		}
		return &SecretFromSrc{
			Id:    s.Id,
			Value: s.Src,
		}, nil
	case s.PlainValue != "":
		if s.Id == "" {
			return nil, fmt.Errorf("type value should be used with id parameter")
		}
		return &SecretFromPlainValue{
			Id:    s.Id,
			Value: s.PlainValue,
		}, nil
	default:
		return nil, fmt.Errorf("secret type is not supported")
	}
}
