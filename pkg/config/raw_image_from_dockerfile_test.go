package config

import (
	"errors"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"gopkg.in/yaml.v2"

	"github.com/werf/werf/pkg/util"
)

var _ = Describe("rawImageFromDockerfile", func() {
	var localGitRepo *LocalGitRepoStub
	var giterminismManager *GiterminismManagerStub

	BeforeEach(func() {
		// TODO: This might break sometimes since this is shared between all tests.
		//  This global variable needs to be factored out.
		parentStack = util.NewStack()

		localGitRepo = NewLocalGitRepoStub("9d8059842b6fde712c58315ca0ab4713d90761c0")
		giterminismManager = NewGiterminismManagerStub(localGitRepo)
	})

	DescribeTable("unmarshal and convert to directive succeed and produce expected Dependencies",
		func(yamlMap map[string]interface{}, expected []*Dependency) {
			switch {
			case len(yamlMap) == 0:
				Fail("yamlMap should not be empty")
			case len(expected) == 0:
				Fail("expected dependencies should not be empty")
			}

			rawYaml, err := yaml.Marshal(yamlMap)
			Expect(err).To(Succeed())

			doc := &doc{Content: rawYaml}
			rawDockerfileImage := &rawImageFromDockerfile{doc: doc}

			Expect(yaml.UnmarshalStrict(doc.Content, rawDockerfileImage)).To(Succeed())

			dockerfileImage, err := rawDockerfileImage.toImageFromDockerfileDirective(giterminismManager, "image1")
			Expect(err).To(Succeed())

			for i, expectedDep := range expected {
				Expect(expectedDep.ImageName).To(Equal(dockerfileImage.Dependencies[i].ImageName))
				Expect(expectedDep.After).To(Equal(dockerfileImage.Dependencies[i].After))
				Expect(expectedDep.Before).To(Equal(dockerfileImage.Dependencies[i].Before))

				for j, expectedDepImport := range expectedDep.Imports {
					Expect(expectedDepImport.Type).To(Equal(dockerfileImage.Dependencies[i].Imports[j].Type))
					Expect(expectedDepImport.TargetEnv).To(Equal(dockerfileImage.Dependencies[i].Imports[j].TargetEnv))
					Expect(expectedDepImport.TargetBuildArg).To(Equal(dockerfileImage.Dependencies[i].Imports[j].TargetBuildArg))
				}
			}
		},
		Entry(
			"with simple image dependency",
			map[string]interface{}{
				"image":      "image1",
				"dockerfile": "Dockerfile",
				"dependencies": []map[string]interface{}{{
					"image": "image2",
				}},
			},
			[]*Dependency{{
				ImageName: "image2",
			}},
		),
		Entry(
			"with ImageTag dependency",
			map[string]interface{}{
				"image":      "image1",
				"dockerfile": "Dockerfile",
				"dependencies": []map[string]interface{}{{
					"image": "image2",
					"imports": []map[string]string{{
						"type":           string(ImageTagImport),
						"targetBuildArg": "IMAGE_TAG",
					}},
				}},
			},
			[]*Dependency{{
				ImageName: "image2",
				Imports: []*DependencyImport{{
					Type:           ImageTagImport,
					TargetBuildArg: "IMAGE_TAG",
				}},
			}},
		),
		Entry(
			"in complex scenarios",
			map[string]interface{}{
				"image":      "image1",
				"dockerfile": "Dockerfile",
				"args": map[string]string{
					"ARG_1": "value",
				},
				"dependencies": []map[string]interface{}{
					{
						"image": "image2",
					},
					{
						"image": "image3",
						"imports": []map[string]string{
							{
								"type":           string(ImageTagImport),
								"targetBuildArg": "IMAGE_TAG_1",
							},
							{
								"type":           string(ImageNameImport),
								"targetBuildArg": "IMAGE_NAME_1",
							},
						},
					},
					{
						"image": "image4",
						"imports": []map[string]string{
							{
								"type":           string(ImageTagImport),
								"targetBuildArg": "IMAGE_TAG_2",
							},
							{
								"type":           string(ImageNameImport),
								"targetBuildArg": "IMAGE_NAME_2",
							},
							{
								"type":           string(ImageIDImport),
								"targetBuildArg": "IMAGE_ID_2",
							},
							{
								"type":           string(ImageDigestImport),
								"targetBuildArg": "IMAGE_DIGEST_2",
							},
							{
								"type":           string(ImageRepoImport),
								"targetBuildArg": "IMAGE_REPO_2",
							},
						},
					},
				},
			},
			[]*Dependency{
				{
					ImageName: "image2",
				},
				{
					ImageName: "image3",
					Imports: []*DependencyImport{
						{
							Type:           ImageTagImport,
							TargetBuildArg: "IMAGE_TAG_1",
						},
						{
							Type:           ImageNameImport,
							TargetBuildArg: "IMAGE_NAME_1",
						},
					},
				},
				{
					ImageName: "image4",
					Imports: []*DependencyImport{
						{
							Type:           ImageTagImport,
							TargetBuildArg: "IMAGE_TAG_2",
						},
						{
							Type:           ImageNameImport,
							TargetBuildArg: "IMAGE_NAME_2",
						},
						{
							Type:           ImageIDImport,
							TargetBuildArg: "IMAGE_ID_2",
						},
						{
							Type:           ImageDigestImport,
							TargetBuildArg: "IMAGE_DIGEST_2",
						},
						{
							Type:           ImageRepoImport,
							TargetBuildArg: "IMAGE_REPO_2",
						},
					},
				},
			},
		),
	)

	DescribeTable("unmarshal and convert to directive fail with configError",
		func(yamlMap map[string]interface{}) {
			if len(yamlMap) == 0 {
				Fail("yamlMap should not be empty")
			}

			rawYaml, err := yaml.Marshal(yamlMap)
			Expect(err).To(Succeed())

			doc := &doc{Content: rawYaml}
			rawDockerfileImage := &rawImageFromDockerfile{doc: doc}

			Expect(yaml.UnmarshalStrict(doc.Content, rawDockerfileImage)).To(Succeed())

			var errConf *configError
			_, err = rawDockerfileImage.toImageFromDockerfileDirective(giterminismManager, "image1")
			Expect(errors.As(err, &errConf)).To(BeTrue())
		},
		Entry(
			"with missing dependency image",
			map[string]interface{}{
				"image":      "image1",
				"dockerfile": "Dockerfile",
				"dependencies": []map[string]interface{}{{
					"imports": []map[string]string{{
						"type":           string(ImageTagImport),
						"targetBuildArg": "IMAGE_TAG",
					}},
				}},
			},
		),
		Entry(
			"with forbidden before directive",
			map[string]interface{}{
				"image":      "image1",
				"dockerfile": "Dockerfile",
				"dependencies": []map[string]interface{}{{
					"image":  "image2",
					"before": "install",
				}},
			},
		),
		Entry(
			"with forbidden after directive",
			map[string]interface{}{
				"image":      "image1",
				"dockerfile": "Dockerfile",
				"dependencies": []map[string]interface{}{{
					"image": "image2",
					"after": "install",
				}},
			},
		),
		Entry(
			"with missing type import directive",
			map[string]interface{}{
				"image":      "image1",
				"dockerfile": "Dockerfile",
				"dependencies": []map[string]interface{}{{
					"image": "image2",
					"imports": []map[string]string{{
						"targetBuildArg": "IMAGE_TAG",
					}},
				}},
			},
		),
		Entry(
			"with missing targetBuildArg import directive",
			map[string]interface{}{
				"image":      "image1",
				"dockerfile": "Dockerfile",
				"dependencies": []map[string]interface{}{{
					"image": "image2",
					"imports": []map[string]string{{
						"type": string(ImageTagImport),
					}},
				}},
			},
		),
		Entry(
			"with forbidden targetEnv import directive",
			map[string]interface{}{
				"image":      "image1",
				"dockerfile": "Dockerfile",
				"dependencies": []map[string]interface{}{{
					"image": "image2",
					"imports": []map[string]string{{
						"type":      string(ImageTagImport),
						"targetEnv": "IMAGE_TAG",
					}},
				}},
			},
		),
		Entry(
			"with dockerfile args pointing to the same build arg as import targetBuildArg",
			map[string]interface{}{
				"image":      "image1",
				"dockerfile": "Dockerfile",
				"args": map[string]string{
					"IMAGE_TAG": "tag",
				},
				"dependencies": []map[string]interface{}{{
					"image": "image2",
					"imports": []map[string]string{{
						"type":           string(ImageTagImport),
						"targetBuildArg": "IMAGE_TAG",
					}},
				}},
			},
		),
	)
})
