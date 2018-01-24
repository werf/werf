package config

import (
	"gopkg.in/yaml.v2"
	"io/ioutil"
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

	return config, nil
}
