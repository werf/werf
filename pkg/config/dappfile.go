package config

import (
	"github.com/flant/dapp/pkg/config/directive"
	"github.com/flant/dapp/pkg/config/ruby_marshal_config"
)

func LoadDappfile(dappfilePath string) (interface{}, error) {
	dimgs, artifacts, err := ParseDimgs(dappfilePath)
	if err != nil {
		return nil, err
	}
	return generateRubyConfig(dimgs, artifacts), err
}

func generateRubyConfig(dimgs []*config.Dimg, artifacts []*config.DimgArtifact) ruby_marshal_config.Config {
	c := ruby_marshal_config.Config{}
	for _, dimg := range dimgs {
		c.Dimg = append(c.Dimg, dimg.ToRuby(artifacts)...)
	}
	return c
}
