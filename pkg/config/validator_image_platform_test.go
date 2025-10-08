package config

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/types"
)

var _ = Describe("imagePlatformValidator", func() {
	DescribeTable("Validate",
		func(rawStapelImages []*rawStapelImage, rawImagesFromDockerfile []*rawImageFromDockerfile, errMatcher types.GomegaMatcher) {
			validator := newImagePlatformValidator()
			err := validator.Validate(rawStapelImages, rawImagesFromDockerfile)
			Expect(err).To(errMatcher)
		},

		Entry("should validate successfully when all dependencies are present (Stapel x Stapel)",
			[]*rawStapelImage{
				{
					Images:   []string{"app"},
					Platform: []string{"linux/amd64"},
				},
			},
			[]*rawImageFromDockerfile{},
			BeNil(),
		),

		Entry("should return error when dependency is missing (Stapel x Stapel)",
			[]*rawStapelImage{
				{
					Images:   []string{"app"},
					Platform: []string{"linux/amd64"},
					RawDependencies: []*rawDependency{
						{Image: "missing-base"},
					},
				},
			},
			[]*rawImageFromDockerfile{},
			MatchError(`image="app" platform="linux/amd64" requires dependency image="missing-base" platform="linux/amd64" which is not present in configuration`),
		),

		Entry("should return error when import is missing (Stapel x Stapel)",
			[]*rawStapelImage{
				{
					Images:   []string{"app"},
					Platform: []string{"linux/amd64"},
					RawImport: []*rawImport{
						{From: "missing-base"},
					},
				},
			},
			[]*rawImageFromDockerfile{},
			MatchError(`image="app" platform="linux/amd64" requires import image="missing-base" platform="linux/amd64" which is not present in configuration`),
		),

		Entry("should validate successfully when all dependencies are present (Dockerfile x Dockerfile)",
			[]*rawStapelImage{},
			[]*rawImageFromDockerfile{
				{
					Images:   []string{"app"},
					Platform: []string{"linux/amd64"},
					RawDependencies: []*rawDependency{
						{Image: "base"},
					},
				},
				{
					Images:          []string{"base"},
					Platform:        []string{"linux/amd64"},
					RawDependencies: []*rawDependency{},
				},
			},
			BeNil(),
		),

		Entry("should return error when dependency is missing (Dockerfile x Dockerfile)",
			[]*rawStapelImage{},
			[]*rawImageFromDockerfile{
				{
					Images:   []string{"app"},
					Platform: []string{"linux/amd64"},
					RawDependencies: []*rawDependency{
						{Image: "missing-base"},
					},
				},
				{
					Images:          []string{"base"},
					Platform:        []string{"linux/amd64"},
					RawDependencies: []*rawDependency{},
				},
			},
			MatchError(`image="app" platform="linux/amd64" requires dependency image="missing-base" platform="linux/amd64" which is not present in configuration`),
		),

		Entry("should validate successfully when Stapel depends on Dockerfile",
			[]*rawStapelImage{
				{
					Images:   []string{"app"},
					Platform: []string{"linux/amd64"},
					RawDependencies: []*rawDependency{
						{Image: "base"},
					},
				},
			},
			[]*rawImageFromDockerfile{
				{
					Images:          []string{"base"},
					Platform:        []string{"linux/amd64"},
					RawDependencies: []*rawDependency{},
				},
			},
			BeNil(),
		),

		Entry("should return error when Stapel depends on Dockerfile with missing image",
			[]*rawStapelImage{
				{
					Images:   []string{"app"},
					Platform: []string{"linux/amd64"},
					RawDependencies: []*rawDependency{
						{Image: "missing-base"},
					},
				},
			},
			[]*rawImageFromDockerfile{
				{
					Images:          []string{"base"},
					Platform:        []string{"linux/amd64"},
					RawDependencies: []*rawDependency{},
				},
			},
			MatchError(`image="app" platform="linux/amd64" requires dependency image="missing-base" platform="linux/amd64" which is not present in configuration`),
		),

		Entry("should validate successfully when Dockerfile depends on Stapel",
			[]*rawStapelImage{
				{
					Images:          []string{"base"},
					Platform:        []string{"linux/amd64"},
					RawDependencies: []*rawDependency{},
				},
			},
			[]*rawImageFromDockerfile{
				{
					Images:   []string{"app"},
					Platform: []string{"linux/amd64"},
					RawDependencies: []*rawDependency{
						{Image: "base"},
					},
				},
			},
			BeNil(),
		),

		Entry("should return error when Dockerfile depends on Stapel with missing image",
			[]*rawStapelImage{
				{
					Images:          []string{"base"},
					Platform:        []string{"linux/amd64"},
					RawDependencies: []*rawDependency{},
				},
			},
			[]*rawImageFromDockerfile{
				{
					Images:   []string{"app"},
					Platform: []string{"linux/amd64"},
					RawDependencies: []*rawDependency{
						{Image: "missing-base"},
					},
				},
			},
			MatchError(`image="app" platform="linux/amd64" requires dependency image="missing-base" platform="linux/amd64" which is not present in configuration`),
		),
	)
})
