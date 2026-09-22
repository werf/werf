package schemas

import (
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/santhosh-tekuri/jsonschema/v6"
)

var _ = Describe("werf.yaml JSON schema error reporting", func() {
	It("reports a typo in a stapel image against the stapel branch only", func() {
		schema, err := jsonschema.NewCompiler().Compile("werf.json")
		Expect(err).NotTo(HaveOccurred())

		err = schema.Validate(yamlDocument([]byte(`
image: app
from: alpine
shel:
  install: ls
`)))
		Expect(err).To(HaveOccurred())

		message := err.Error()
		Expect(message).To(ContainSubstring("'shel' not allowed"))
		Expect(message).NotTo(ContainSubstring("configVersion"))
		Expect(message).NotTo(ContainSubstring("dockerfile"))
		Expect(strings.Count(message, "additional propert")).To(Equal(1))
	})
})
