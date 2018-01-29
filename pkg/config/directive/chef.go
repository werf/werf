package config

import "github.com/flant/dapp/pkg/config/ruby_marshal_config"

type Chef struct {
	Cookbook   string
	Recipe     []string
	Attributes map[interface{}]interface{}
}

func (c *Chef) ToRuby() ruby_marshal_config.Chef {
	rubyChef := ruby_marshal_config.Chef{}
	rubyChef.Recipe = c.Recipe
	rubyChef.Attributes = c.Attributes
	// TODO c.Cookbook
	return rubyChef
}
