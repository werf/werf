package frontend

import (
	"fmt"

	"github.com/moby/buildkit/frontend/dockerfile/shell"
	"github.com/werf/werf/v2/pkg/dockerfile"
)

func NewShlexEnvGetterFromSlice(env []string) shell.EnvGetter {
	return shell.EnvsFromSlice(env)
}

type ShlexAdapter struct {
	shlex *shell.Lex
}

func NewShlexAdapter(shlex *shell.Lex) dockerfile.Expander {
	return &ShlexAdapter{shlex: shlex}
}

func (a *ShlexAdapter) ProcessWordWithMap(word string, env map[string]string) (string, error) {
	s, _, err := a.shlex.ProcessWord(word, NewShlexEnvGetterFromSlice(envToSlice(env)))
	return s, err
}
func (a *ShlexAdapter) ProcessWordsWithMap(word string, env map[string]string) ([]string, error) {
	return a.shlex.ProcessWords(word, NewShlexEnvGetterFromSlice(envToSlice(env)))
}

func envToSlice(env map[string]string) []string {
	s := make([]string, 0, len(env))
	for k, v := range env {
		s = append(s, fmt.Sprintf("%s=%s", k, v))
	}
	return s
}
