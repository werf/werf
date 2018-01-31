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
		return fmt.Errorf("`to` required absolute path") // FIXME
	} else if c.Type == "custom_dir" && c.From != "" && isRelativePath(c.From) {
		return fmt.Errorf("`fromPath` required absolute path") // FIXME
	} else if c.Type != "tmp_dir" && c.Type != "build_dir" {
		return fmt.Errorf("`From` should be `tmp_dir` or `build_dir`") // FIXME
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
