package secret

import (
	"fmt"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

type MergeEncodedYamlTest struct {
	OldData, OldEncodedData []byte
	NewData, NewEncodedData []byte
	ExpectedResult          []byte
}

var _ = Describe("MergeEncodedYaml", func() {
	DescribeTable("change some keys of decoded yaml and create new version of encoded yaml preserving same encoded value for unchanged decoded values",
		func(tst MergeEncodedYamlTest) {
			oldData := strings.TrimSpace(string(tst.OldData)) + "\n"
			newData := strings.TrimSpace(string(tst.NewData)) + "\n"
			oldEncodedData := strings.TrimSpace(string(tst.OldEncodedData)) + "\n"
			newEncodedData := strings.TrimSpace(string(tst.NewEncodedData)) + "\n"
			expectedResult := strings.TrimSpace(string(tst.ExpectedResult)) + "\n"

			fmt.Printf("Old data:\n%s---\n", oldData)
			fmt.Printf("Old encoded data:\n%s---\n", oldEncodedData)
			fmt.Printf("New data:\n%s---\n", newData)
			fmt.Printf("New encoded data:\n%s---\n", newEncodedData)

			res, err := MergeEncodedYaml([]byte(oldData), []byte(newData), []byte(oldEncodedData), []byte(newEncodedData))
			Expect(err).To(Succeed())

			fmt.Printf("Merge result:\n%s---\n", res)
			fmt.Printf("Expected result:\n%s---\n", expectedResult)
			Expect(string(res)).To(Equal(expectedResult))
		},

		Entry("case 1", MergeEncodedYamlTest{
			OldData: []byte(`
database:
  user: vasya
  password: gfhjkm
mailbox:
  address: vasya@myorg.org
  password: gfhjkm
`),
			OldEncodedData: []byte(`
database:
  user: enc1
  password: enc2
mailbox:
  address: enc3
  password: enc4
`),
			NewData: []byte(`
database:
  user: vasya
  password: gfhjkm1
mailbox:
  address: vasya@myorg.org
  password: gfhjkm
`),
			NewEncodedData: []byte(`
database:
  user: enc1-1
  password: enc2-1
mailbox:
  address: enc3-1
  password: enc4-1
`),
			ExpectedResult: []byte(`
database:
  user: enc1
  password: enc2-1
mailbox:
  address: enc3
  password: enc4
`),
		}),

		Entry("complex yaml with anchors and comments", MergeEncodedYamlTest{
			OldData: []byte(`
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
			OldEncodedData: []byte(`
common_values: &common-values
  key1: value1-encoded
  key2:
    aa: bb-encoded
    cc: dd-encoded
  key3:
    - one-encoded
    - two-encoded
    - three-encoded
# Example list
common_list: &common-list
  - name: one-encoded
    value: xxx-encoded
  - name: two-encoded
    value: yyy-encoded
    # Some comment with indent
values:
  # Include common values
  !!merge <<: *common-values
  key4: value4-encoded
  # include common list
  key5: *common-list
`),
			NewData: []byte(`
common_values: &common-values
  key1: value1
  key2:
    aa: bb
    cc: dd-changed
  key3:
    - one
    - two-changed
    - three
# Example list
common_list: &common-list
  - name: one
    value: xxx
  - name: two
    value: yyy-changed
    # Some comment with indent
values:
  # Include common values
  !!merge <<: *common-values
  key4: value4-changed
  # include common list
  key5: *common-list
`),
			NewEncodedData: []byte(`
common_values: &common-values
  key1: value1
  key2:
    aa: bb-encoded1
    cc: dd-changed-encoded1
  key3:
    - one-encoded1
    - two-changed-encoded1
    - three-encoded1
# Example list
common_list: &common-list
  - name: one-encoded1
    value: xxx-encoded1
  - name: two-encoded1
    value: yyy-changed-encoded1
    # Some comment with indent
values:
  # Include common values
  !!merge <<: *common-values
  key4: value4-changed-encoded1
  # include common list
  key5: *common-list
`),
			ExpectedResult: []byte(`
common_values: &common-values
  key1: value1-encoded
  key2:
    aa: bb-encoded
    cc: dd-changed-encoded1
  key3:
    - one-encoded
    - two-changed-encoded1
    - three-encoded
# Example list
common_list: &common-list
  - name: one-encoded
    value: xxx-encoded
  - name: two-encoded
    value: yyy-changed-encoded1
    # Some comment with indent
values:
  # Include common values
  !!merge <<: *common-values
  key4: value4-changed-encoded1
  # include common list
  key5: *common-list
`),
		}),
	)
})
