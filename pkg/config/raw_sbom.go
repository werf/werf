package config

import (
	"fmt"
	"strings"

	yamlv3 "gopkg.in/yaml.v3"
)

type rawSbom struct {
	doc *doc `yaml:"-"`

	Fragment *string  `yaml:"fragment,omitempty"`
	Gost     *rawGost `yaml:"gost,omitempty"`

	UnsupportedAttributes map[string]interface{} `yaml:",inline"`
}

func (s *rawSbom) UnmarshalYAML(unmarshal func(interface{}) error) error {
	// Bind to parent context for proper error reporting.
	// rawSbom can appear only under image sections (parent: *rawStapelImage or *rawImageFromDockerfile).
	switch parent := parentStack.Peek().(type) {
	case *rawStapelImage:
		s.doc = parent.doc
	case *rawImageFromDockerfile:
		s.doc = parent.doc
	case *rawSbom:
		// In case of nested parsing (shouldn't normally happen), inherit context.
		s.doc = parent.doc
	}

	parentStack.Push(s)
	type plain rawSbom
	err := unmarshal((*plain)(s))
	parentStack.Pop()
	if err != nil {
		return err
	}

	if err := checkOverflow(s.UnsupportedAttributes, nil, s.docForErrors()); err != nil {
		return err
	}

	// Soft validation (YAML-level only):
	// If `sbom:` section is present, fragment must be present and non-empty,
	// and must be valid YAML.
	if err := s.validateFragmentYAML(); err != nil {
		return err
	}

	return nil
}

func (s *rawSbom) docForErrors() *doc {
	if s != nil && s.doc != nil {
		return s.doc
	}
	// Fallback: avoid panics in error formatting in unexpected edge cases.
	return &doc{Content: []byte{}}
}

func (s *rawSbom) validateFragmentYAML() error {
	d := s.docForErrors()

	if s.Fragment == nil || strings.TrimSpace(*s.Fragment) == "" {
		return newDetailedConfigError("`sbom.fragment` is required when `sbom:` section is specified and must not be empty!", nil, d)
	}

	// Validate fragment YAML by parsing with yaml.v3.
	// We expect a YAML mapping at the root (e.g. `components: ...` or a full BOM document).
	var fragment map[string]any
	if err := yamlv3.Unmarshal([]byte(*s.Fragment), &fragment); err != nil {
		return newDetailedConfigError(fmt.Sprintf("`sbom.fragment` must be valid YAML: %s", err), nil, d)
	}
	if fragment == nil {
		return newDetailedConfigError("`sbom.fragment` must be a YAML mapping (e.g. `components: ...`)", nil, d)
	}

	return nil
}
