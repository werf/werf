package secret

import (
	"fmt"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"gopkg.in/yaml.v2"
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

		Entry("simple yaml", `
a: one
bbb:
  url: http://service:3441/bbb
postgresql:
  password:
    _default: "1234"
`),

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

	// TODO: support restoring of original type during decode
	It("should encode integer, bool, float, timestamp and binary as string, then convert to string during decode", func() {
		originalData := `
mystring: value
mybool: !!bool true
myint: !!int 32
myfloat: !!float 64.5
mytime: !!timestamp 2022-07-15 20:33:23.34
mybinary: !binary |
  R0lGODlhDAAMAIQAAP//9/X17unp5WZmZgAAAOfn515eXvPz7Y6OjuDg4J+fn5
  OTk6enp56enmlpaWNjY6Ojo4SEhP/++f/++f/++f/++f/++f/++f/++f/++f/+
  +f/++f/++f/++f/++f/++SH+Dk1hZGUgd2l0aCBHSU1QACwAAAAADAAMAAAFLC
  AgjoEwnuNAFOhpEMTRiggcz4BNJHrv/zCFcLiwMWYNG84BwwEeECcgggoBADs=
`

		expectEncoded := `
mystring: 'encoded: value'
mybool: 'encoded: true'
myint: 'encoded: 32'
myfloat: 'encoded: 64.5'
mytime: 'encoded: 2022-07-15 20:33:23.34 +0000 UTC'
mybinary: |
  encoded: R0lGODlhDAAMAIQAAP//9/X17unp5WZmZgAAAOfn515eXvPz7Y6OjuDg4J+fn5
  OTk6enp56enmlpaWNjY6Ojo4SEhP/++f/++f/++f/++f/++f/++f/++f/++f/+
  +f/++f/++f/++f/++f/++SH+Dk1hZGUgd2l0aCBHSU1QACwAAAAADAAMAAAFLC
  AgjoEwnuNAFOhpEMTRiggcz4BNJHrv/zCFcLiwMWYNG84BwwEeECcgggoBADs=
`

		expectDecoded := `
mystring: value
mybool: "true"
myint: "32"
myfloat: "64.5"
mytime: 2022-07-15 20:33:23.34 +0000 UTC
mybinary: |
  R0lGODlhDAAMAIQAAP//9/X17unp5WZmZgAAAOfn515eXvPz7Y6OjuDg4J+fn5
  OTk6enp56enmlpaWNjY6Ojo4SEhP/++f/++f/++f/++f/++f/++f/++f/++f/+
  +f/++f/++f/++f/++f/++SH+Dk1hZGUgd2l0aCBHSU1QACwAAAAADAAMAAAFLC
  AgjoEwnuNAFOhpEMTRiggcz4BNJHrv/zCFcLiwMWYNG84BwwEeECcgggoBADs=
`

		enc := NewYamlEncoder(&EncoderMock{})

		fmt.Printf("Original data:\n%s\n---\n", strings.TrimSpace(string(originalData)))

		encodedData, err := enc.EncryptYamlData([]byte(originalData))
		Expect(err).To(Succeed())

		fmt.Printf("Encoded data:\n%s\n---\n", strings.TrimSpace(string(encodedData)))
		fmt.Printf("Expect encoded data:\n%s\n---\n", strings.TrimSpace(string(expectEncoded)))

		Expect(strings.TrimSpace(string(encodedData))).To(Equal(strings.TrimSpace(string(expectEncoded))))

		resultData, err := enc.DecryptYamlData(encodedData)
		Expect(err).To(Succeed())

		fmt.Printf("Decoded data:\n%s\n---\n", strings.TrimSpace(string(resultData)))
		fmt.Printf("Expect decoded data:\n%s\n---\n", strings.TrimSpace(string(expectDecoded)))

		Expect(strings.TrimSpace(expectDecoded)).To(Equal(strings.TrimSpace(string(resultData))), fmt.Sprintf("\n[EXPECTED]\n%q\n[GOT]\n%q\n", originalData, resultData))

		var resultDataMap map[string]interface{}
		Expect(yaml.Unmarshal(resultData, &resultDataMap)).To(Succeed())

		fmt.Printf("Decoded data map: %#v\n", resultDataMap)

		for _, k := range []string{"mystring", "myint", "myfloat", "mytime", "mybinary"} {
			_, isStr := resultDataMap[k].(string)
			Expect(isStr).To(BeTrue())
		}
	})
})

type EncoderMock struct{}

func (s *EncoderMock) Encrypt(data []byte) ([]byte, error) {
	return []byte(fmt.Sprintf("encoded: %s", data)), nil
}

func (s *EncoderMock) Decrypt(data []byte) ([]byte, error) {
	decodedData := []byte(strings.TrimPrefix(string(data), "encoded: "))
	return decodedData, nil
}
