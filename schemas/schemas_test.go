package schemas

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/santhosh-tekuri/jsonschema/v6"
	"gopkg.in/yaml.v3"
)

var _ = Describe("published schemas", func() {
	It("are valid draft-07 schemas", func() {
		paths, err := filepath.Glob("*.json")
		Expect(err).NotTo(HaveOccurred())
		Expect(paths).NotTo(BeEmpty())

		for _, path := range paths {
			_, err := jsonschema.NewCompiler().Compile(path)
			Expect(err).NotTo(HaveOccurred(), path)
		}
	})

	It("reject unknown properties in every declared object", func() {
		paths, err := filepath.Glob("*.json")
		Expect(err).NotTo(HaveOccurred())

		for _, path := range paths {
			data, err := os.ReadFile(path)
			Expect(err).NotTo(HaveOccurred())

			var schema any
			Expect(json.Unmarshal(data, &schema)).To(Succeed())

			var nonStrictObjects []string
			var walk func(any, string)
			walk = func(value any, pointer string) {
				switch value := value.(type) {
				case map[string]any:
					if value["type"] == "object" && value["properties"] != nil && value["additionalProperties"] != false && pointer != "/definitions/dependencyBase" {
						nonStrictObjects = append(nonStrictObjects, pointer)
					}
					for key, child := range value {
						walk(child, pointer+"/"+key)
					}
				case []any:
					for index, child := range value {
						walk(child, fmt.Sprintf("%s/%d", pointer, index))
					}
				}
			}
			walk(schema, "")

			Expect(nonStrictObjects).To(BeEmpty(), path)
		}
	})

	It("accept every werf-giterminism.yaml fixture in the repository", func() {
		schema, err := jsonschema.NewCompiler().Compile("werf-giterminism.json")
		Expect(err).NotTo(HaveOccurred())

		fixtures := findFixtures("../test", "werf-giterminism.yaml")
		Expect(fixtures).NotTo(BeEmpty())

		for _, path := range fixtures {
			data, err := os.ReadFile(path)
			Expect(err).NotTo(HaveOccurred())

			if err := schema.Validate(yamlDocument(data)); err != nil {
				Fail(fmt.Sprintf("%s: %v", path, err))
			}
		}
	})
})

func findFixtures(root, name string) []string {
	var fixtures []string
	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !entry.IsDir() && entry.Name() == name {
			fixtures = append(fixtures, path)
		}
		return nil
	})
	Expect(err).NotTo(HaveOccurred())
	return fixtures
}

func yamlDocument(data []byte) any {
	var document any
	err := yaml.NewDecoder(bytes.NewReader(data)).Decode(&document)
	if errors.Is(err, io.EOF) {
		return nil
	}
	Expect(err).NotTo(HaveOccurred())
	return document
}
