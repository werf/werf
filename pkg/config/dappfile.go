package config

import (
	"io/ioutil"

	"gopkg.in/flant/yaml.v2"
	"github.com/flant/dapp/pkg/config/ruby_marshal_config"
)

func LoadDappfile(dappfilePath string) (interface{}, error) {
	data, err := ioutil.ReadFile(dappfilePath)
	if err != nil {
		return nil, err
	}

	config := &ruby_marshal_config.Config{}
	err = yaml.Unmarshal([]byte(data), &config)
	if err != nil {
		return nil, err
	}

	config = &ruby_marshal_config.Config{
		Dimg: []ruby_marshal_config.Dimg{
			{
				Name:    "assets",
				Builder: "shell",
				GitArtifact: ruby_marshal_config.GitArtifact{
					Local: []ruby_marshal_config.GitArtifactLocal{
						{
							Export: []ruby_marshal_config.GitArtifactLocalExport{
								{
									ArtifactBaseExport: ruby_marshal_config.ArtifactBaseExport{
										Cwd:   "/",
										To:    "/app",
										Owner: "app",
										ExcludePaths: []string{
											".helm",
											"Vagrantfile",
										},
									},
									StageDependencies: ruby_marshal_config.StageDependencies{
										Install: []string{
											"package.json",
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	return config, nil
}
