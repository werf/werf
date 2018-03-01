package config

import (
	"fmt"

	"github.com/flant/dapp/pkg/config/ruby_marshal_config"
)

type Mount struct {
	To   string
	From string
	Type string

	Raw *RawMount
}

func (c *Mount) Validate() error {
	if c.To == "" || !IsAbsolutePath(c.To) {
		return NewDetailedConfigError("`to: PATH` absolute path required for mount!", c.Raw, c.Raw.RawDimg.Doc)
	} else if c.Type == "custom_dir" {
		if (c.From != "" && isRelativePath(c.From)) || c.From == "" {
			return NewDetailedConfigError("`fromPath: PATH` should be absolute path for mount!", c.Raw, c.Raw.RawDimg.Doc)
		}
	} else if c.Type != "tmp_dir" && c.Type != "build_dir" {
		return NewDetailedConfigError(fmt.Sprintf("Invalid `from: %s` for mount: expected `tmp_dir` or `build_dir`!", c.Type), c.Raw, c.Raw.RawDimg.Doc)
	}
	return nil
}

func (c *Mount) ToRuby() ruby_marshal_config.Mount {
	rubyMount := ruby_marshal_config.Mount{}
	rubyMount.To = c.To
	rubyMount.From = c.From
	rubyMount.Type = ruby_marshal_config.Symbol(c.Type)
	return rubyMount
}
