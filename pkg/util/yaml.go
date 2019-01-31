package util

import (
	"fmt"

	"github.com/ghodss/yaml"
)

func DumpYaml(v interface{}) string {
	data, err := yaml.Marshal(v)
	if err != nil {
		panic(fmt.Sprintf("Unable to dump yaml %#v: %s", v, err))
	}
	return string(data)
}
