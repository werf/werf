package config

import (
	"github.com/romana/rlog"
	"io/ioutil"
	"gopkg.in/yaml.v2"
)

func LoadConfig(dappfilePath string) (interface{}, error) {
	rlog.Debugf("Parsing file `%s` ...", dappfilePath)
	data, err := ioutil.ReadFile(dappfilePath)
	if err != nil {
		rlog.Errorf("Error reading file `%s`: %v", dappfilePath, err)
		return nil, err
	}

	config := &Config{}
	err = yaml.Unmarshal([]byte(data), &config)
	if err != nil {
		rlog.Errorf("Error parsing yaml from `%s`: %v", dappfilePath, err)
		return nil, err
	}

	rlog.Infof("Config `%s` loaded", dappfilePath)

	return config, nil
}
