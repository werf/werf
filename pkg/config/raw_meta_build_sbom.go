package config

import (
	"fmt"

	"github.com/samber/lo"

	"github.com/werf/werf/v2/pkg/sbom"
	"github.com/werf/werf/v2/pkg/sbom/cyclonedxutil/gost"
	"github.com/werf/werf/v2/pkg/util/option"
)

type rawMetaBuildSbom struct {
	Enable   *bool    `yaml:"enable,omitempty"`
	Standard *string  `yaml:"standard,omitempty"`
	Gost     *rawGost `yaml:"gost,omitempty"`

	rawMetaBuild *rawMetaBuild `yaml:"-"`

	UnsupportedAttributes map[string]interface{} `yaml:",inline"`
}

func (s *rawMetaBuildSbom) UnmarshalYAML(unmarshal func(interface{}) error) error {
	if parent, ok := parentStack.Peek().(*rawMetaBuild); ok {
		s.rawMetaBuild = parent
	}

	parentStack.Push(s)
	type plain rawMetaBuildSbom
	err := unmarshal((*plain)(s))
	parentStack.Pop()
	if err != nil {
		return err
	}

	if err := checkOverflow(s.UnsupportedAttributes, nil, s.rawMetaBuild.rawMeta.doc); err != nil {
		return err
	}

	if err := s.applySbomAndValidate(); err != nil {
		return err
	}

	return nil
}

// applySbomAndValidate applies and validates SBOM.
//
// Rules:
// - if both enable and standard are omitted: enable=false, standard=cyclonedx@1.6
// - if standard is specified: require enable=true (and validate standard)
// - if enable=false is specified without standard: standard defaults to cyclonedx@1.6
// - if enable=true is specified: require standard (and it currently supports only cyclonedx@1.6)
func (s *rawMetaBuildSbom) applySbomAndValidate() error {
	const defaultStandard = sbom.StandardTypeCycloneDX16

	enableSpecified := s.Enable != nil
	standardSpecified := s.Standard != nil

	if !enableSpecified && !standardSpecified {
		s.Enable = lo.ToPtr(false)
		s.Standard = lo.ToPtr(defaultStandard.String())
		return nil
	}

	if standardSpecified && !enableSpecified {
		return fmt.Errorf("meta build sbom config: field 'enable' must be explicitly set to true when 'standard' is specified")
	}

	if enableSpecified {
		enable := option.PtrValueOrDefault(s.Enable, false)

		if enable && !standardSpecified {
			return fmt.Errorf("meta build sbom config: field 'standard' is required when 'enable' is true")
		}

		if !enable && !standardSpecified {
			s.Standard = lo.ToPtr(defaultStandard.String())
			return nil
		}
	}

	if s.Standard == nil || *s.Standard == "" {
		return fmt.Errorf("meta build sbom config: field 'standard' must not be empty")
	}

	parsed, err := sbom.StandardTypeString(*s.Standard)
	if err != nil {
		return fmt.Errorf("meta build sbom config: unsupported 'standard' value %q: %w", *s.Standard, err)
	}

	if parsed != defaultStandard {
		return fmt.Errorf(
			"meta build sbom config: unsupported 'standard' value %q (only %q is supported)",
			*s.Standard,
			defaultStandard.String(),
		)
	}

	return nil
}

func (s *rawMetaBuildSbom) toDirective() *MetaBuildSbom {
	if s == nil {
		return nil
	}

	return &MetaBuildSbom{
		Enable:   lo.FromPtr(s.Enable),
		Standard: sbom.StandardTypeCycloneDX16,
		Gost:     gost.DefaultConfig().Merge(s.Gost.toConfig()),
	}
}
