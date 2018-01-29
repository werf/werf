package config

import (
	"github.com/flant/dapp/pkg/config/directive"
	"github.com/flant/dapp/pkg/config/ruby_marshal_config"
)

func LoadDappfile(dappfilePath string) (interface{}, error) {
	dimgs, err := ParseDimgs(dappfilePath)
	if err != nil {
		return nil, err
	}
	return generateRubyConfig(dimgs), err
}

func generateRubyConfig(dimgs []*config.Dimg) *ruby_marshal_config.Config {
	c := &ruby_marshal_config.Config{}
	for _, dimg := range dimgs {
		c.Dimg = append(c.Dimg, dimg.ToRuby())
	}
	return c
}
