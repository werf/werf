package config

import (
	"bytes"
	"errors"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	. "github.com/onsi/gomega"
	"github.com/santhosh-tekuri/jsonschema/v6"
	"gopkg.in/yaml.v3"
)

const werfSchemaPath = "../../schemas/werf.json"

func compileWerfSchema() *jsonschema.Schema {
	schema, err := jsonschema.NewCompiler().Compile(werfSchemaPath)
	Expect(err).NotTo(HaveOccurred())
	return schema
}

func yamlDocuments(data []byte) []any {
	decoder := yaml.NewDecoder(bytes.NewReader(data))

	var documents []any
	for {
		var document any
		err := decoder.Decode(&document)
		if errors.Is(err, io.EOF) {
			return documents
		}
		Expect(err).NotTo(HaveOccurred())

		if document != nil {
			documents = append(documents, document)
		}
	}
}

func yamlDocument(data string) any {
	documents := yamlDocuments([]byte(data))
	Expect(documents).To(HaveLen(1))
	return documents[0]
}

func nonTemplatedWerfYamlFixtures(root string) []string {
	var fixtures []string
	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() || entry.Name() != "werf.yaml" {
			return nil
		}

		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		if strings.Contains(string(data), "{{") {
			return nil
		}

		fixtures = append(fixtures, path)
		return nil
	})
	Expect(err).NotTo(HaveOccurred())
	return fixtures
}
