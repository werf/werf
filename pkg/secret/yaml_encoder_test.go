package secret

import (
	"fmt"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("YamlEncoder", func() {
	DescribeTable("encode then decode of yaml config should result in the same yaml config",
		func(originalData string) {
			enc := NewYamlEncoder(&EncoderMock{})

			fmt.Printf("Original data:\n%s\n---\n", strings.TrimSpace(string(originalData)))

			encodedData, err := enc.EncryptYamlData([]byte(originalData))
			Expect(err).To(Succeed())

			fmt.Printf("Encoded data:\n%s\n---\n", strings.TrimSpace(string(encodedData)))

			resultData, err := enc.DecryptYamlData(encodedData)
			Expect(err).To(Succeed())

			fmt.Printf("Decoded data:\n%s\n---\n", strings.TrimSpace(string(resultData)))

			Expect(strings.TrimSpace(originalData)).To(Equal(strings.TrimSpace(string(resultData))), fmt.Sprintf("\n[EXPECTED]\n%q\n[GOT]\n%q\n", originalData, resultData))
		},

		Entry("simple yaml", `a: one`),

		Entry("yaml with anchors", `
common_values: &common-values
  key1: value1
  key2:
    aa: bb
    cc: dd
  key3:
    - one
    - two
    - three
# Example list
common_list: &common-list
  - name: one
    value: xxx
  - name: two
    value: yyy
    # Some comment with indent
values:
  # Include common values
  !!merge <<: *common-values
  key4: value4
  # include common list
  key5: *common-list
`),

		Entry("null yaml", `null`),

		Entry("complex yaml", `
image:
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
`),
	)
})

type EncoderMock struct{}

func (s *EncoderMock) Encrypt(data []byte) ([]byte, error) {
	return []byte(fmt.Sprintf("encoded: %s", data)), nil
}

func (s *EncoderMock) Decrypt(data []byte) ([]byte, error) {
	decodedData := []byte(strings.TrimPrefix(string(data), "encoded: "))
	return decodedData, nil
}
