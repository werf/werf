package config

import (
	"github.com/flant/dapp/pkg/config/ruby_marshal_config"
)

var (
	Warnings []string
)

func LoadDappfile(dappfilePath string) (interface{}, error, []string) {
	dimgs, err := ParseDimgs(dappfilePath)
	if err != nil {
		return nil, err, nil
	}
	return generateRubyConfig(dimgs), err, Warnings
}

func generateRubyConfig(dimgs []*Dimg) *ruby_marshal_config.Config {
	c := &ruby_marshal_config.Config{}
	for _, dimg := range dimgs {
		c.Dimg = append(c.Dimg, dimg.ToRuby())
	}
	return c
}
