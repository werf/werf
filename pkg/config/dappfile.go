package config

import (
	"io/ioutil"

	"gopkg.in/flant/yaml.v2"
)

func LoadDappfile(dappfilePath string) (interface{}, error) {
	data, err := ioutil.ReadFile(dappfilePath)
	if err != nil {
		return nil, err
	}

	config := &Config{}
	err = yaml.Unmarshal([]byte(data), &config)
	if err != nil {
		return nil, err
	}

	config = &Config{
		Dimg: []Dimg{
			Dimg{
				Name:    "assets",
				Builder: "shell",
				Shell:   nil,
				GitArtifact: GitArtifact{
					Local: []GitArtifactLocal{
						GitArtifactLocal{
							Export: []GitArtifactLocalExport{
								GitArtifactLocalExport{
									Cwd:   "/",
									To:    "/app",
									Owner: "app",
									ExcludePaths: []string{
										".helm",
										"Vagrantfile",
									},
									StageDependencies: StageDependencies{
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
