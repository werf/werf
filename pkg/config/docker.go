package config

import "github.com/flant/dapp/pkg/config/ruby_marshal_config"

type Docker struct {
	Volume     []string
	Expose     []string
	Env        map[string]string
	Label      map[string]string
	Cmd        []string
	Onbuild    []string
	Workdir    string
	User       string
	Entrypoint []string

	Raw *RawDocker
}

func (c *Docker) Validate() error {
	return nil
}

func (c *Docker) ToRuby() ruby_marshal_config.DockerDimg {
	rubyDocker := ruby_marshal_config.DockerDimg{}
	rubyDocker.Volume = c.Volume
	rubyDocker.Expose = c.Expose
	rubyDocker.Env = symbolizeKeys(c.Env)
	rubyDocker.Label = symbolizeKeys(c.Label)
	rubyDocker.Cmd = c.Cmd
	rubyDocker.Onbuild = c.Onbuild
	rubyDocker.Workdir = c.Workdir
	rubyDocker.User = c.User
	rubyDocker.Entrypoint = c.Entrypoint
	return rubyDocker
}

func symbolizeKeys(hash map[string]string) map[ruby_marshal_config.Symbol]string {
	symbolizeHash := map[ruby_marshal_config.Symbol]string{}
	for key, value := range hash {
		symbolizeHash[ruby_marshal_config.Symbol(key)] = value
	}
	return symbolizeHash
}
