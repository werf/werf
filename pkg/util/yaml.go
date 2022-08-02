package util

import (
	"fmt"

	"sigs.k8s.io/yaml"
)

func DumpYaml(v interface{}) string {
	data, err := yaml.Marshal(v)
	if err != nil {
		panic(fmt.Sprintf("Unable to dump yaml %#v: %s", v, err))
	}
	return string(data)
}
