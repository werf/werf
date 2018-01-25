package config

import (
	"fmt"
	"reflect"
	"strings"
)

func CheckOverflow(m map[string]interface{}, config interface{}) error {
	if len(m) > 0 {
		var keys []string
		for k := range m {
			keys = append(keys, k)
		}

		val := reflect.Indirect(reflect.ValueOf(config))
		return fmt.Errorf("Config `%s` doesn't support this keys: `%s`", val.Type().Name(), strings.Join(keys, "`, `")) // FIXME
	}
	return nil
}

func ArtifactByName(artifacts []*DimgArtifact, name string) *DimgArtifact {
	for _, artifact := range artifacts {
		if artifact.Artifact == name {
			return artifact
		}
	}
	return nil
}
