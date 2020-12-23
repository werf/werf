package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path/filepath"

	"github.com/werf/werf/pkg/util"
)

func PrepareConfig(projectPath string) (c GiterminismConfig, err error) {
	configPath := filepath.Join(projectPath, "werf-giterminism.yaml")

	var data []byte
	if exist, err := util.RegularFileExists(configPath); err != nil {
		return c, err
	} else if !exist {
		data = []byte("giterminismConfigVersion: \"1\"")
	} else {
		data, err = ioutil.ReadFile(configPath)
		if err != nil {
			return c, fmt.Errorf("unable to read file %s: %s", configPath, err)
		}
	}

	err = processWithOpenAPISchema(&data)
	if err != nil {
		return c, fmt.Errorf("%s validation failed: %s", configPath, err)
	}

	if err := json.Unmarshal(data, &c); err != nil {
		panic(fmt.Sprint("unexpected error: ", err))
	}

	return c, err
}
