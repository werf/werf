package config

import (
	"fmt"
	"github.com/flant/dapp/pkg/config/ruby_marshal_config"
)

type Mount struct {
	From string
	// TODO: FromPath ?!
	To string
	// TODO: Type string

	Raw *RawMount
}

func (c *Mount) Validate() error {
	if c.From == "" {
		return fmt.Errorf("`from` required!") // FIXME
	} else if c.To == "" {
		return fmt.Errorf("`to` required!") // FIXME
	}
	// TODO: валидация `To` абсолютный путь
	// TODO: валидация `From` tmp_dir или build_dir
	// TODO: валидация `FromPath` относительный путь
	return nil
}

func (c *Mount) ToRuby() ruby_marshal_config.Mount {
	rubyMount := ruby_marshal_config.Mount{}
	rubyMount.From = c.From
	rubyMount.To = c.To
	// TODO: rubyMount.Type := c.Type
	return rubyMount
}
