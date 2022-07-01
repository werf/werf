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

		Entry("added elements into map and array", MergeEncodedYamlTest{
			OldData: []byte(`
database:
  user: vasya
  password: gfhjkm
hosts:
  - name: local1
    address: a1
  - name: local2
    address: a2
`),
			OldEncodedData: []byte(`
database:
  user: enc1
  password: enc2
hosts:
  - name: enc3
    address: enc4
  - name: enc5
    address: enc6
`),
			NewData: []byte(`
database:
  user: vasya
  password: gfhjkm
  admin: petya
  adminPassword:
    _default: "1234"
    production: "#@$213sldfzjxXFLKJSdf1233s"
hosts:
  - name: local1
    address: a1
    options:
      timeout: 5s
  - name: local2
    address: a2
`),
			NewEncodedData: []byte(`
database:
  user: enc1-1
  password: enc2-1
  admin: enc3-1
  adminPassword:
    _default: enc4-1
    production: enc5-1
hosts:
  - name: enc6-1
    address: enc7-1
    options:
      timeout: enc8-1
  - name: enc9-1
    address: enc10-1
`),
			ExpectedResult: []byte(`
database:
  user: enc1
  password: enc2
  admin: enc3-1
  adminPassword:
    _default: enc4-1
    production: enc5-1
hosts:
  - name: enc3
    address: enc4
    options:
      timeout: enc8-1
  - name: enc5
    address: enc6
`),
		}),

		Entry("removed elements from map and array", MergeEncodedYamlTest{
			OldData: []byte(`
database:
  user: vasya
  password: gfhjkm
  admin: petya
  adminPassword:
    _default: "1234"
    production: "#@$213sldfzjxXFLKJSdf1233s"
hosts:
  - name: local1
    address: a1
    options:
      timeout: 5s
  - name: local2
    address: a2
`),
			OldEncodedData: []byte(`
database:
  user: enc1
  password: enc2
  admin: enc3
  adminPassword:
    _default: enc4
    production: enc5
hosts:
  - name: enc6
    address: enc7
    options:
      timeout: enc8
  - name: enc9
    address: enc10
`),
			NewData: []byte(`
database:
  user: vasya
  adminPassword:
    _default: "1234"
    production: "#@$213sldfzjxXFLKJSdf1233s"
hosts:
  - name: local2
    address: a2
`),
			NewEncodedData: []byte(`
database:
  user: enc1-1
  adminPassword:
    _default: enc2-1
    production: enc3-1
hosts:
  - name: enc4-1
    address: enc5-1
`),
			ExpectedResult: []byte(`
database:
  user: enc1
  adminPassword:
    _default: enc4
    production: enc5
hosts:
  - name: enc4-1
    address: enc5-1
`),
		}),

		Entry("no changes in array", MergeEncodedYamlTest{
			OldData: []byte(`
arr:
  - k1: v1
    k2: v2
  - k1: v3
    k2: v4
  - k1: v5
    k2: v6
`),
			OldEncodedData: []byte(`
arr:
  - k1: enc1
    k2: enc2
  - k1: enc3
    k2: enc4
  - k1: enc5
    k2: enc6
`),
			NewData: []byte(`
arr:
  - k1: v1
    k2: v2
  - k1: v3
    k2: v4
  - k1: v5
    k2: v6
`),
			NewEncodedData: []byte(`
arr:
  - k1: enc1-1
    k2: enc2-1
  - k1: enc3-1
    k2: enc4-1
  - k1: enc5-1
    k2: enc6-1
`),
			ExpectedResult: []byte(`
arr:
  - k1: enc1
    k2: enc2
  - k1: enc3
    k2: enc4
  - k1: enc5
    k2: enc6
`),
		}),

		Entry("random crazy mindless changes: change order of array elements, swap map keys with different values, change order of map values", MergeEncodedYamlTest{
			OldData: []byte(`
database:
  user: vasya
  password: gfhjkm
  admin: petya
  adminPassword:
    _default: "1234"
    production: "#@$213sldfzjxXFLKJSdf1233s"
hosts:
  - name: local0
    address: a0
  - name: local1
    address: a1
    options:
      timeout: 5s
  - name: local2
    address: a2
`),
			OldEncodedData: []byte(`
database:
  user: enc1
  password: enc2
  admin: enc3
  adminPassword:
    _default: enc4
    production: enc5
hosts:
  - name: enc6
    address: enc7
  - name: enc8
    address: enc9
    options:
      timeout: enc10
  - name: enc11
    address: enc12
`),
			NewData: []byte(`
database:
  password: gfhjkm
  user: vasya
  adminPassword: petya
  admin:
    _default: "1234"
    production: "#@$213sldfzjxXFLKJSdf1233s"
hosts:
  - name: local0
    address: new-addr
  - name: local2
    address: a2
  - name: local1
    address: a1
    options:
      timeout: 5s
`),
			NewEncodedData: []byte(`
database:
  password: enc1-1
  user: enc2-1
  adminPassword: enc3-1
  admin:
    _default: enc4-1
    production: enc5-1
hosts:
  - name: enc6-1
    address: enc7-1
  - name: enc8-1
    address: enc9-1
  - name: enc10-1
    address: enc11-1
    options:
      timeout: enc12-1
`),
			ExpectedResult: []byte(`
database:
  password: enc2
  user: enc1
  adminPassword: enc3-1
  admin:
    _default: enc4-1
    production: enc5-1
hosts:
  - name: enc6
    address: enc7-1
  - name: enc8-1
    address: enc9-1
  - name: enc10-1
    address: enc11-1
    options:
      timeout: enc12-1
`),
		}),
	)
})
