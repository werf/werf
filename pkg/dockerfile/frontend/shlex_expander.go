package frontend

import (
	"github.com/moby/buildkit/frontend/dockerfile/shell"

	"github.com/werf/werf/v2/pkg/dockerfile"
)

type ShlexExpanderFactory struct {
	EscapeToken rune
}

func NewShlexExpanderFactory(escapeToken rune) *ShlexExpanderFactory {
	return &ShlexExpanderFactory{EscapeToken: escapeToken}
}

func (factory *ShlexExpanderFactory) GetExpander(opts dockerfile.ExpandOptions) dockerfile.Expander {
	shlex := shell.NewLex(factory.EscapeToken)
	shlex.SkipUnsetEnv = opts.SkipUnsetEnv
	return shlex
}
