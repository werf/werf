package common

import (
	"context"
	"os"
	"slices"
	"strings"
)

type EnvValuesOptions struct {
	Name    string
	Exclude []string
}

// GetEnvValues reads a non-empty bare value followed by suffixed values in lexical order.
// Exclusions apply to both the exact name and its suffixed forms.
func GetEnvValues(_ context.Context, opts EnvValuesOptions) []string {
	var values []string
	env := os.Environ()
	slices.Sort(env)
	for _, entry := range env {
		name, value, _ := strings.Cut(entry, "=")
		if name != opts.Name && !strings.HasPrefix(name, opts.Name+"_") {
			continue
		}
		if name == opts.Name && value == "" {
			continue
		}
		if slices.ContainsFunc(opts.Exclude, func(excluded string) bool {
			return name == excluded || strings.HasPrefix(name, excluded+"_")
		}) {
			continue
		}
		values = append(values, value)
	}
	return values
}
