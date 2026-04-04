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
	return &shlexAdapter{lex: shlex}
}

type shlexAdapter struct {
	lex *shell.Lex
}

type mapEnvGetter map[string]string

func (m mapEnvGetter) Get(key string) (string, bool) {
	v, ok := m[key]
	return v, ok
}

func (m mapEnvGetter) Keys() []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

func (a *shlexAdapter) ProcessWordWithMap(word string, env map[string]string) (string, error) {
	result, _, err := a.lex.ProcessWord(word, mapEnvGetter(env))
	return result, err
}

func (a *shlexAdapter) ProcessWordsWithMap(word string, env map[string]string) ([]string, error) {
	return a.lex.ProcessWords(word, mapEnvGetter(env))
}
