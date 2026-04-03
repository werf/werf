package frontend

import (
	"sort"

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
	return &shlexExpander{lex: shlex}
}

type shlexExpander struct {
	lex *shell.Lex
}

func (e *shlexExpander) ProcessWordWithMap(word string, env map[string]string) (string, error) {
	result, _, err := e.lex.ProcessWord(word, mapEnvGetter(env))
	if err != nil {
		return "", err
	}
	return result, nil
}

func (e *shlexExpander) ProcessWordsWithMap(word string, env map[string]string) ([]string, error) {
	return e.lex.ProcessWords(word, mapEnvGetter(env))
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
	sort.Strings(keys)
	return keys
}
