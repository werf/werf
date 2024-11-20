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
	switch {
	case s.Env != "":
		if s.Src != "" || s.PlainValue != "" {
			return newDetailedConfigError("env can't be specified with other types", s, s.doc)
		}
	case s.Src != "":
		if s.Env != "" || s.PlainValue != "" {
			return newDetailedConfigError("src can't be specified with other types", s, s.doc)
		}
	case s.PlainValue != "":
		if s.Id == "" {
			return newDetailedConfigError("id is required field for type value", s, s.doc)
		}
		if s.Env != "" || s.Src != "" {
			return newDetailedConfigError("value can't be specified with other types", s, s.doc)
		}
	default:
		return newDetailedConfigError("one of env, source, value should be specified", s, s.doc)
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
		return &SecretFromPlainValue{
			Id:    s.Id,
			Value: s.PlainValue,
		}, nil
	default:
		return nil, fmt.Errorf("secret type is not supported")
	}
}
