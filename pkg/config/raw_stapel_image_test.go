package config

import (
	"errors"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"gopkg.in/yaml.v2"

	"github.com/werf/werf/pkg/util"
)

var _ = Describe("rawStapelImage", func() {
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
			rawStapelImage := &rawStapelImage{doc: doc}
			Expect(yaml.UnmarshalStrict(doc.Content, rawStapelImage)).To(Succeed())

			stapelImage, err := rawStapelImage.toStapelImageDirective(giterminismManager, "image1")
			Expect(err).To(Succeed())

			for i, expectedDep := range expected {
				Expect(expectedDep.ImageName).To(Equal(stapelImage.Dependencies[i].ImageName))
				Expect(expectedDep.After).To(Equal(stapelImage.Dependencies[i].After))
				Expect(expectedDep.Before).To(Equal(stapelImage.Dependencies[i].Before))

				for j, expectedDepImport := range expectedDep.Imports {
					Expect(expectedDepImport.Type).To(Equal(stapelImage.Dependencies[i].Imports[j].Type))
					Expect(expectedDepImport.TargetEnv).To(Equal(stapelImage.Dependencies[i].Imports[j].TargetEnv))
					Expect(expectedDepImport.TargetBuildArg).To(Equal(stapelImage.Dependencies[i].Imports[j].TargetBuildArg))
				}
			}
		},
		Entry(
			"with simple image dependency",
			map[string]interface{}{
				"image": "image1",
				"from":  "alpine",
				"dependencies": []map[string]interface{}{{
					"image":  "image2",
					"before": "install",
				}},
			},
			[]*Dependency{{
				ImageName: "image2",
				Before:    "install",
			}},
		),
		Entry(
			"with ImageTag dependency",
			map[string]interface{}{
				"image": "image1",
				"from":  "alpine",
				"dependencies": []map[string]interface{}{{
					"image":  "image2",
					"before": "install",
					"imports": []map[string]string{{
						"type":      string(ImageTagImport),
						"targetEnv": "IMAGE_TAG",
					}},
				}},
			},
			[]*Dependency{{
				ImageName: "image2",
				Before:    "install",
				Imports: []*DependencyImport{{
					Type:      ImageTagImport,
					TargetEnv: "IMAGE_TAG",
				}},
			}},
		),
		Entry(
			"in complex scenarios",
			map[string]interface{}{
				"image": "image1",
				"from":  "alpine",
				"dependencies": []map[string]interface{}{
					{
						"image":  "image2",
						"before": "install",
					},
					{
						"image": "image3",
						"after": "install",
						"imports": []map[string]string{
							{
								"type":      string(ImageTagImport),
								"targetEnv": "IMAGE_TAG_1",
							},
							{
								"type":      string(ImageNameImport),
								"targetEnv": "IMAGE_NAME_1",
							},
						},
					},
					{
						"image": "image4",
						"after": "setup",
						"imports": []map[string]string{
							{
								"type":      string(ImageTagImport),
								"targetEnv": "IMAGE_TAG_2",
							},
							{
								"type":      string(ImageNameImport),
								"targetEnv": "IMAGE_NAME_2",
							},
							{
								"type":      string(ImageIDImport),
								"targetEnv": "IMAGE_ID_2",
							},
							{
								"type":      string(ImageDigestImport),
								"targetEnv": "IMAGE_DIGEST_2",
							},
							{
								"type":      string(ImageRepoImport),
								"targetEnv": "IMAGE_REPO_2",
							},
						},
					},
				},
			},
			[]*Dependency{
				{
					ImageName: "image2",
					Before:    "install",
				},
				{
					ImageName: "image3",
					After:     "install",
					Imports: []*DependencyImport{
						{
							Type:      ImageTagImport,
							TargetEnv: "IMAGE_TAG_1",
						},
						{
							Type:      ImageNameImport,
							TargetEnv: "IMAGE_NAME_1",
						},
					},
				},
				{
					ImageName: "image4",
					After:     "setup",
					Imports: []*DependencyImport{
						{
							Type:      ImageTagImport,
							TargetEnv: "IMAGE_TAG_2",
						},
						{
							Type:      ImageNameImport,
							TargetEnv: "IMAGE_NAME_2",
						},
						{
							Type:      ImageIDImport,
							TargetEnv: "IMAGE_ID_2",
						},
						{
							Type:      ImageDigestImport,
							TargetEnv: "IMAGE_DIGEST_2",
						},
						{
							Type:      ImageRepoImport,
							TargetEnv: "IMAGE_REPO_2",
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
			rawStapelImage := &rawStapelImage{doc: doc}

			Expect(yaml.UnmarshalStrict(doc.Content, rawStapelImage)).To(Succeed())

			var errConf *configError
			_, err = rawStapelImage.toStapelImageDirective(giterminismManager, "image1")
			Expect(errors.As(err, &errConf)).To(BeTrue())
		},
		Entry(
			"with missing dependency image",
			map[string]interface{}{
				"image": "image1",
				"from":  "alpine",
				"dependencies": []map[string]interface{}{{
					"before": "install",
					"imports": []map[string]string{{
						"type":      string(ImageTagImport),
						"targetEnv": "IMAGE_TAG",
					}},
				}},
			},
		),
		Entry(
			"with missing before/after dependency directives",
			map[string]interface{}{
				"image": "image1",
				"from":  "alpine",
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
			"with missing type import directive",
			map[string]interface{}{
				"image": "image1",
				"from":  "alpine",
				"dependencies": []map[string]interface{}{{
					"image":  "image2",
					"before": "install",
					"imports": []map[string]string{{
						"targetEnv": "IMAGE_TAG",
					}},
				}},
			},
		),
		Entry(
			"with missing targetEnv import directive",
			map[string]interface{}{
				"image": "image1",
				"from":  "alpine",
				"dependencies": []map[string]interface{}{{
					"image":  "image2",
					"before": "install",
					"imports": []map[string]string{{
						"type": string(ImageTagImport),
					}},
				}},
			},
		),
		Entry(
			"with forbidden targetBuildArg import directive",
			map[string]interface{}{
				"image": "image1",
				"from":  "alpine",
				"dependencies": []map[string]interface{}{{
					"image":  "image2",
					"before": "install",
					"imports": []map[string]string{{
						"type":           string(ImageTagImport),
						"targetBuildArg": "IMAGE_TAG",
					}},
				}},
			},
		),
	)
})
