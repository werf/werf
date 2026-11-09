package schemas

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/santhosh-tekuri/jsonschema/v6"
)

var _ = Describe("Chart.yaml JSON schema", func() {
	var schema *jsonschema.Schema

	BeforeEach(func() {
		var err error
		schema, err = jsonschema.NewCompiler().Compile("chart.json")
		Expect(err).NotTo(HaveOccurred())
	})

	DescribeTable("accepts a document",
		func(document string) {
			Expect(schema.Validate(yamlDocument([]byte(document)))).To(Succeed())
		},
		Entry("chart with the required Helm metadata", `
apiVersion: v2
name: mychart
version: 1.2.3
`),
		Entry("main chart without the metadata werf fills in", `
type: library
`),
		Entry("main chart with kubeVersion only", `
kubeVersion: "~1.20.3"
`),
		Entry("chart with every informational directive", `
apiVersion: v2
name: mychart
version: 1.2.3
kubeVersion: ">=1.20.0-0"
description: This is My Chart
type: application
keywords:
- apps
home: https://example.org
sources:
- https://github.com/my/chart
icon: https://example.org/mychart-icon.svg
appVersion: "1.0"
deprecated: false
annotations:
  anyAdditionalInfo: here
maintainers:
- name: John Doe
  email: john@example.org
  url: https://example.org/john
`),
		Entry("apiVersion v1", `
apiVersion: v1
name: mychart
version: 1.2.3
`),
		Entry("apiVersion v3", `
apiVersion: v3
name: mychart
version: 1.2.3
`),
		Entry("dependency with every standard directive", `
apiVersion: v2
name: myparent
version: 1.2.3
dependencies:
- name: mychild
  version: 1.0.0
  repository: https://charts.example.org
  condition: mychild.enabled
  tags:
  - backend
  alias: mychildalias
`),
		Entry("dependency with a repository alias", `
apiVersion: v2
name: myparent
version: 1.2.3
dependencies:
- name: mychild
  version: 1.0.0
  repository: "@myrepo"
`),
		Entry("dependency with a repository alias in the alias form", `
apiVersion: v2
dependencies:
- name: mychild
  version: 1.0.0
  repository: alias:myrepo
`),
		Entry("local dependency with an empty repository", `
apiVersion: v2
dependencies:
- name: redis
  repository: ""
`),
		Entry("informational metadata without a URL scheme", `
name: mychart
version: 1.2.3
icon: myicon
home: myhome
sources:
- mysource
maintainers:
- name: John Doe
  email: john
  url: myurl
`),
		Entry("optional directives left empty", `
apiVersion: v2
name: ""
version: ""
kubeVersion: ""
dependencies:
- name: mychild
  version: ""
  condition: ""
  alias: ""
`),
		Entry("local dependency without a repository", `
apiVersion: v2
dependencies:
- name: redis
  condition: redis.enabled
`),
		Entry("import-values in both forms", `
apiVersion: v2
name: myparent
version: 1.2.3
dependencies:
- name: mychild
  import-values:
  - mydict
  - child: mydict
    parent: myotherdict
`),
		Entry("export-values mapping", `
apiVersion: v2
name: myparent
version: 1.2.3
dependencies:
- name: mychild
  export-values:
  - parent: myparam
    child: myparam
  - parent: global.werf.images.pg.registry
    child: image.registry
`),
		Entry("export-values mapping to the root of the dependent chart", `
apiVersion: v2
dependencies:
- name: mychild
  export-values:
  - parent: mydict
    child: ""
  - parent: myotherdict
    child: .
`),
		Entry("export-values in the string form", `
apiVersion: v2
dependencies:
- name: mychild
  export-values:
  - mydict
  - ""
  - .
`),
		Entry("empty dependencies list", `
apiVersion: v2
dependencies: []
`),
		Entry("empty export-values list", `
apiVersion: v2
dependencies:
- name: mychild
  export-values: []
`),
	)

	DescribeTable("rejects a document",
		func(document string) {
			Expect(schema.Validate(yamlDocument([]byte(document)))).NotTo(Succeed())
		},
		Entry("unknown root key", `
apiVersions: v2
name: mychart
version: 1.2.3
`),
		Entry("misspelled maintainers", `
maintainters:
- name: John Doe
`),
		Entry("unknown apiVersion", `
apiVersion: v4
name: mychart
version: 1.2.3
`),
		Entry("unknown chart type", `
type: application-library
`),
		Entry("non-string name", `
name: 123
`),
		Entry("non-boolean deprecated", `
deprecated: "yes"
`),
		Entry("non-string annotation value", `
annotations:
  replicas: 3
`),
		Entry("maintainer without a name", `
maintainers:
- email: john@example.org
`),
		Entry("dependencies as a mapping", `
dependencies:
  mychild:
    version: 1.0.0
`),
		Entry("dependency without a name", `
dependencies:
- version: 1.0.0
`),
		Entry("unknown dependency key", `
dependencies:
- name: mychild
  depends-on:
  - myotherchild
`),
		Entry("import-values mapping without child", `
dependencies:
- name: mychild
  import-values:
  - parent: myotherdict
`),
		Entry("export-values mapping without child", `
dependencies:
- name: mychild
  export-values:
  - parent: myparam
`),
		Entry("export-values mapping without parent", `
dependencies:
- name: mychild
  export-values:
  - child: myparam
`),
		Entry("export-values mapping with an unknown key", `
dependencies:
- name: mychild
  export-values:
  - parent: myparam
    child: myparam
    merge: true
`),
		Entry("export-values mapping with a non-string parent", `
dependencies:
- name: mychild
  export-values:
  - parent: 1
    child: myparam
`),
		Entry("export-values mapping with a non-string child", `
dependencies:
- name: mychild
  export-values:
  - parent: myparam
    child: ["myparam"]
`),
		Entry("export-values mapping from the root of the parent chart", `
dependencies:
- name: mychild
  export-values:
  - parent: ""
    child: myparam
`),
		Entry("export-values mapping from the dot path of the parent chart", `
dependencies:
- name: mychild
  export-values:
  - parent: .
    child: myparam
`),
		Entry("export-values mapping from a blank parent path", `
dependencies:
- name: mychild
  export-values:
  - parent: " "
    child: myparam
`),
		Entry("export-values item of an unsupported type", `
dependencies:
- name: mychild
  export-values:
  - 1
`),
		Entry("export-values as a mapping", `
dependencies:
- name: mychild
  export-values:
    parent: myparam
    child: myparam
`),
	)
})
