package werf_chart

import (
	"fmt"
	"io/ioutil"

	"github.com/werf/werf/pkg/deploy/secret"
	"sigs.k8s.io/yaml"
)

func DecodeSecretValuesFile(path string, m secret.Manager) (map[string]interface{}, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("cannot read file %q: %s", path, err)
	}

	decodedData, err := m.DecryptYamlData(data)
	if err != nil {
		return nil, fmt.Errorf("cannot decode file %q secret data: %s", path, err)
	}

	rawValues := map[string]interface{}{}
	if err := yaml.Unmarshal(decodedData, &rawValues); err != nil {
		return nil, fmt.Errorf("cannot unmarshal secret values file %s: %s", path, err)
	}

	return rawValues, nil
}
