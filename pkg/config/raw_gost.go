package config

import (
	"fmt"

	"github.com/samber/lo"

	"github.com/werf/werf/v2/pkg/sbom/cyclonedxutil/gost"
)

type rawGost struct {
	AttackSurface    *string `yaml:"attackSurface,omitempty"`
	SecurityFunction *string `yaml:"securityFunction,omitempty"`

	doc *doc `yaml:"-"`
}

func (g *rawGost) UnmarshalYAML(unmarshal func(interface{}) error) error {
	switch parent := parentStack.Peek().(type) {
	case *rawSbom:
		g.doc = parent.doc
	case *rawMetaBuildSbom:
		g.doc = parent.rawMetaBuild.rawMeta.doc
	}

	parentStack.Push(g)
	type plain rawGost
	err := unmarshal((*plain)(g))
	parentStack.Pop()
	if err != nil {
		return err
	}

	return g.validate()
}

func (g *rawGost) validate() error {
	if g.AttackSurface != nil && !gost.IsValidGostValue(*g.AttackSurface) {
		return newDetailedConfigError(fmt.Sprintf("invalid 'attackSurface' value %q: expected 'yes', 'no' or 'inherit'", *g.AttackSurface), nil, g.doc)
	}
	if g.SecurityFunction != nil && !gost.IsValidGostValue(*g.SecurityFunction) {
		return newDetailedConfigError(fmt.Sprintf("invalid 'securityFunction' value %q: expected 'yes', 'no' or 'inherit'", *g.SecurityFunction), nil, g.doc)
	}
	return nil
}

func (g *rawGost) toConfig() gost.Config {
	if g == nil {
		return gost.Config{}
	}

	return gost.Config{
		AttackSurface:    gost.GostValue(lo.FromPtr(g.AttackSurface)),
		SecurityFunction: gost.GostValue(lo.FromPtr(g.SecurityFunction)),
	}
}
