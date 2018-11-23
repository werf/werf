package secret

import (
	"bytes"
	"testing"
)

type SecretMock struct{}

func (s *SecretMock) Generate(data []byte) ([]byte, error) {
	return []byte("encoded data"), nil
}

func (s *SecretMock) Extract(data []byte) ([]byte, error) {
	return []byte("data"), nil
}

func TestBaseSecret_doYamlData(t *testing.T) {
	valuesData := []byte(`image:
  repository: data
  tag: data
  pullPolicy: data
  output:
    host: data
    port: data
    buffer_chunk_limit: data
    buffer_queue_limit: data
  env: {}
  service:
    type: data
    externalPort: data
    ports:
    - name: data
      protocol: data
      containerPort: data
  ingress:
    enabled: data
    annotations: data
    tls: data
  configMaps:
    general.conf: data
    system.conf: data
    forward-input.conf: data
    output.conf: data
  resources: {}
  persistence:
    enabled: data
    accessMode: data
    size: data
  nodeSelector: {}
  tolerations: []
  affinity: {}
`)

	s, err := newBaseManager(&SecretMock{})
	if err != nil {
		t.Fatal(err)
	}

	encodedData, err := s.GenerateYamlData(valuesData)
	if err != nil {
		t.Fatal(err)
	}

	resultData, err := s.ExtractYamlData(encodedData)
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal(valuesData, resultData) {
		t.Errorf("\n[EXPECTED]\n%s\n[GOT]\n%s\n", string(valuesData), string(resultData))
	}
}
