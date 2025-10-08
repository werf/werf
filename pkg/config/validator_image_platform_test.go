package config

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/types"
)

var _ = Describe("imagePlatformValidator", func() {
	Describe("Validate", func() {
		DescribeTable("Stapel x Stapel with dependencies cases", testImagePlatformValidator,
			Entry("should return error when dependency is missing",
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
			Entry("should validate successfully when no dependencies",
				[]*rawStapelImage{
					{
						Images:   []string{"app"},
						Platform: []string{"linux/amd64"},
					},
				},
				[]*rawImageFromDockerfile{},
				BeNil(),
			),
			Entry("should validate successfully when dependencies are preset",
				[]*rawStapelImage{
					{
						Images:   []string{"app"},
						Platform: []string{"linux/amd64", "linux/arm64"},
					},
					{
						Images:   []string{"test-dependency"},
						Platform: []string{"linux/amd64", "linux/arm64"},
						RawDependencies: []*rawDependency{
							{Image: "app"},
						},
					},
				},
				[]*rawImageFromDockerfile{},
				BeNil(),
			),
		)

		DescribeTable("Stapel x Stapel with import cases", testImagePlatformValidator,
			Entry("should return error when import specified with import.image field and base image is missing",
				[]*rawStapelImage{
					{
						Images:   []string{"app"},
						Platform: []string{"linux/amd64"},
					},
					{
						Images:   []string{"test-import"},
						Platform: []string{"linux/amd64", "linux/arm64"},
						RawImport: []*rawImport{
							{ImageName: "app"},
						},
					},
				},
				[]*rawImageFromDockerfile{},
				MatchError(`image="test-import" platform="linux/arm64" requires import image="app" platform="linux/arm64" which is not present in configuration`),
			),

			Entry("should validate successfully when import specified with import.from field and base image is external one",
				[]*rawStapelImage{
					{
						Images:   []string{"app"},
						Platform: []string{"linux/amd64"},
						RawImport: []*rawImport{
							{From: "alpine:3.17"},
						},
					},
				},
				[]*rawImageFromDockerfile{},
				BeNil(),
			),
		)

		DescribeTable("Dockerfile x Dockerfile with dependencies cases", testImagePlatformValidator,
			Entry("should return error when dependency is missing",
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
			Entry("should validate successfully when no dependencies",
				[]*rawStapelImage{
					{
						Images:   []string{"app"},
						Platform: []string{"linux/amd64"},
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
			Entry("should validate successfully when all dependencies are present",
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
		)

		DescribeTable("Stapel x Dockerfile with dependencies cases", testImagePlatformValidator,
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
		)

		DescribeTable("Dockerfile x Stapel with dependencies cases", testImagePlatformValidator,
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
})

func testImagePlatformValidator(rawStapelImages []*rawStapelImage, rawImagesFromDockerfile []*rawImageFromDockerfile, errMatcher types.GomegaMatcher) {
	validator := newImagePlatformValidator()
	err := validator.Validate(rawStapelImages, rawImagesFromDockerfile)
	Expect(err).To(errMatcher)
}
