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
	yamlv2 "gopkg.in/yaml.v2"
	"gopkg.in/yaml.v3"

	"github.com/werf/common-go/pkg/util"
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

// parseWerfDocument runs one werf.yaml document through the same raw structs
// and directive conversion that GetWerfConfig applies, dispatching on the key
// splitByMetaAndRawImages dispatches on.
func parseWerfDocument(data string) error {
	parentStack = util.NewStack()
	giterminismManager := NewGiterminismManagerStub(NewLocalGitRepoStub("9d8059842b6fde712c58315ca0ab4713d90761c0"))

	var raw map[string]interface{}
	if err := yamlv2.UnmarshalStrict([]byte(data), &raw); err != nil {
		return err
	}

	document := &doc{Content: []byte(data)}
	switch {
	case isMetaDoc(raw):
		return yamlv2.UnmarshalStrict(document.Content, &rawMeta{doc: document})
	case isImageFromDockerfileDoc(raw):
		image := &rawImageFromDockerfile{doc: document}
		if err := yamlv2.UnmarshalStrict(document.Content, image); err != nil {
			return err
		}
		_, err := image.toImageFromDockerfileDirectives(giterminismManager)
		return err
	case isImageDoc(raw):
		image := &rawStapelImage{doc: document}
		if err := yamlv2.UnmarshalStrict(document.Content, image); err != nil {
			return err
		}
		_, err := image.toStapelImageDirectives(giterminismManager)
		return err
	default:
		return errors.New("cannot recognize config section type")
	}
}

func stapelImageWithGit(gitYaml string) string {
	return "image: image1\nfrom: alpine\nshell:\n  install:\n  - echo install\n  beforeSetup:\n  - echo beforeSetup\n  setup:\n  - echo setup\ngit:\n" + gitYaml
}

func parseStapelImageGitLocals(data string) ([]*GitLocal, error) {
	parentStack = util.NewStack()
	giterminismManager := NewGiterminismManagerStub(NewLocalGitRepoStub("9d8059842b6fde712c58315ca0ab4713d90761c0"))

	document := &doc{Content: []byte(data)}
	image := &rawStapelImage{doc: document}
	if err := yamlv2.UnmarshalStrict(document.Content, image); err != nil {
		return nil, err
	}

	stapelImage, err := image.toStapelImageDirective(giterminismManager, "image1")
	if err != nil {
		return nil, err
	}

	return stapelImage.Git.Local, nil
}

func nonTemplatedWerfYamlFixtures(root string) []string {
	var fixtures []string
	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			if strings.HasPrefix(entry.Name(), ".") && path != root {
				return filepath.SkipDir
			}
			return nil
		}
		if entry.Name() != "werf.yaml" {
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
