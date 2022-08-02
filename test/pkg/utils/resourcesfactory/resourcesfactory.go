package resourcesfactory

import (
	"encoding/json"

	. "github.com/onsi/gomega"
	v1 "k8s.io/api/apps/v1"
	"sigs.k8s.io/yaml"
)

func unmarshalObject(manifestYaml string, obj interface{}) {
	manifestJson, err := yaml.YAMLToJSON([]byte(manifestYaml))
	Expect(err).To(Succeed())
	Expect(json.Unmarshal(manifestJson, obj)).To(Succeed())
}

func NewDeployment(manifestYaml string) *v1.Deployment {
	obj := &v1.Deployment{}
	unmarshalObject(manifestYaml, &obj)
	return obj
}
