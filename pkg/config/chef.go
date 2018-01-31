package config

import "github.com/flant/dapp/pkg/config/ruby_marshal_config"

type Chef struct {
	Cookbook   []Cookbook
	Recipe     []string
	Attributes map[interface{}]interface{}

	Raw *RawChef
}

func (c *Chef) Validate() error {
	return nil
}

func (c *Chef) ToRuby() ruby_marshal_config.Chef {
	rubyChef := ruby_marshal_config.Chef{}
	rubyChef.Recipe = c.Recipe
	rubyChef.Attributes = c.Attributes

	rubyChef.Cookbook = make(map[string]map[ruby_marshal_config.Symbol]interface{})
	for _, cookbook := range c.Cookbook {
		rubyChef.Cookbook[cookbook.Name] = make(map[ruby_marshal_config.Symbol]interface{})
		rubyChef.Cookbook[cookbook.Name][ruby_marshal_config.Symbol("name")] = cookbook.Name

		if cookbook.VersionConstraint != "" {
			rubyChef.Cookbook[cookbook.Name][ruby_marshal_config.Symbol("version_constraint")] = cookbook.VersionConstraint
		}

		if cookbook.Path != "" {
			rubyChef.Cookbook[cookbook.Name][ruby_marshal_config.Symbol("path")] = cookbook.Path
		}

		for k, v := range symbolizeCookbookFields(cookbook.Fields) {
			rubyChef.Cookbook[cookbook.Name][k] = v
		}
	}

	return rubyChef
}

func symbolizeCookbookFields(hash map[string]interface{}) map[ruby_marshal_config.Symbol]interface{} {
	symbolizeHash := map[ruby_marshal_config.Symbol]interface{}{}
	for key, value := range hash {
		symbolizeHash[ruby_marshal_config.Symbol(key)] = value
	}
	return symbolizeHash
}
