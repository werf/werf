package schemas

import (
	"os"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/santhosh-tekuri/jsonschema/v6"
)

var _ = Describe("werf-includes.yaml JSON schema", func() {
	var schema *jsonschema.Schema

	BeforeEach(func() {
		var err error
		schema, err = jsonschema.NewCompiler().Compile("werf-includes.json")
		Expect(err).NotTo(HaveOccurred())
	})

	It("accepts the documentation example", func() {
		data, err := os.ReadFile("../docs/examples/includes/app/werf-includes.yaml")
		Expect(err).NotTo(HaveOccurred())
		Expect(schema.Validate(yamlDocument(data))).To(Succeed())
	})

	DescribeTable("accepts a document",
		func(document string) {
			Expect(schema.Validate(yamlDocument([]byte(document)))).To(Succeed())
		},
		Entry("include with every directive", `
includes:
- git: https://github.com/werf/werf
  basicAuth:
    username: bot
    password:
      env: GIT_PASSWORD
  branch: main
  add: /docs/examples/includes/werf-common
  to: /
  includePaths: [".werf/**", "*.yaml"]
  excludePaths: [".werf/local/**"]
`),
		Entry("include pinned to a tag", `
includes:
- git: https://github.com/werf/werf
  tag: v2.0.0
  add: /
  to: /vendor/werf
`),
		Entry("include pinned to a commit with password from file", `
includes:
- git: https://github.com/werf/werf
  basicAuth:
    username: bot
    password:
      src: ~/.git-password
  commit: 0123456789abcdef0123456789abcdef01234567
  add: /
  to: /
`),
		Entry("empty includes list", `
includes: []
`),
	)

	DescribeTable("rejects a document",
		func(document string) {
			Expect(schema.Validate(yamlDocument([]byte(document)))).NotTo(Succeed())
		},
		Entry("unknown root key", `
include:
- git: https://github.com/werf/werf
`),
		Entry("include without git", `
includes:
- branch: main
  add: /
  to: /
`),
		Entry("include without add", `
includes:
- git: https://github.com/werf/werf
  branch: main
  to: /
`),
		Entry("include without to", `
includes:
- git: https://github.com/werf/werf
  branch: main
  add: /
`),
		Entry("include without ref", `
includes:
- git: https://github.com/werf/werf
  add: /
  to: /
`),
		Entry("include with branch and tag", `
includes:
- git: https://github.com/werf/werf
  branch: main
  tag: v2.0.0
  add: /
  to: /
`),
		Entry("include with empty branch", `
includes:
- git: https://github.com/werf/werf
  branch: ""
  add: /
  to: /
`),
		Entry("relative add", `
includes:
- git: https://github.com/werf/werf
  branch: main
  add: docs
  to: /
`),
		Entry("relative to", `
includes:
- git: https://github.com/werf/werf
  branch: main
  add: /
  to: vendor
`),
		Entry("absolute includePaths", `
includes:
- git: https://github.com/werf/werf
  branch: main
  add: /
  to: /
  includePaths: ["/docs"]
`),
		Entry("basicAuth without username", `
includes:
- git: https://github.com/werf/werf
  basicAuth:
    password:
      env: GIT_PASSWORD
  branch: main
  add: /
  to: /
`),
		Entry("basicAuth password with two sources", `
includes:
- git: https://github.com/werf/werf
  basicAuth:
    username: bot
    password:
      env: GIT_PASSWORD
      value: secret
  branch: main
  add: /
  to: /
`),
		Entry("basicAuth password with empty source", `
includes:
- git: https://github.com/werf/werf
  basicAuth:
    username: bot
    password:
      env: ""
  branch: main
  add: /
  to: /
`),
		Entry("unknown include key", `
includes:
- git: https://github.com/werf/werf
  branch: main
  add: /
  to: /
  url: https://github.com/werf/werf
`),
	)
})
