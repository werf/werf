package image

import (
	"bytes"
	"context"

	"github.com/moby/buildkit/frontend/dockerfile/instructions"
	"github.com/moby/buildkit/frontend/dockerfile/parser"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/werf/werf/v3/pkg/config"
)

var _ = DescribeTable("dependency build args reaching image content",
	func(dockerfile string, expected bool) {
		result, err := parser.Parse(bytes.NewBufferString(dockerfile))
		Expect(err).To(Succeed())

		stages, _, err := instructions.Parse(result.AST, nil)
		Expect(err).To(Succeed())

		reachesContent, err := dependencyArgsReachContent(stages, []string{"BASE_IMAGE"}, result.EscapeToken, true)
		Expect(err).To(Succeed())
		Expect(reachesContent).To(Equal(expected))
	},
	Entry("meta arg used only by FROM", `
ARG BASE_IMAGE
FROM ${BASE_IMAGE}
`, false),
	Entry("arg declared inside a stage", `
FROM scratch
ARG BASE_IMAGE
LABEL purpose=test
`, true),
	Entry("arg written to ENV", `
FROM scratch
ARG BASE_IMAGE
ENV DEPENDENCY=$BASE_IMAGE
`, true),
	Entry("RUN can inspect injected dependency args", `
ARG BASE_IMAGE
FROM ${BASE_IMAGE}
RUN env > /build-env.txt
`, true),
	Entry("lexer-unfriendly RUN remains valid", `
ARG BASE_IMAGE
FROM ${BASE_IMAGE}
RUN echo it's
`, true),
	Entry("lexer-unfriendly metadata remains valid", `
ARG BASE_IMAGE
FROM ${BASE_IMAGE}
LABEL text=it's
`, true),
	Entry("undeclared arg referenced by an instruction", `
FROM scratch
ENV DEPENDENCY=$BASE_IMAGE
`, true),
	Entry("similarly named arg", `
FROM scratch
ENV DEPENDENCY=$BASE_IMAGE_SUFFIX
`, false),
)

var _ = Describe("resolved dependency input mapping", func() {
	It("keeps legacy FROM-only dependency args stable across unrelated RUN commands", func() {
		result, err := parser.Parse(bytes.NewBufferString("ARG BASE_IMAGE\nFROM ${BASE_IMAGE}\nRUN true\n"))
		Expect(err).To(Succeed())
		stages, _, err := instructions.Parse(result.AST, nil)
		Expect(err).To(Succeed())

		reachesContent, err := dependencyArgsReachContent(stages, []string{"BASE_IMAGE"}, result.EscapeToken, false)
		Expect(err).To(Succeed())
		Expect(reachesContent).To(BeFalse())
	})

	It("marks every staged Dockerfile image produced by the mapper", func() {
		data := []byte("FROM scratch AS base\nRUN env\nFROM base\n")
		dockerfileConfig := &config.ImageFromDockerfile{
			Name:   "app",
			Staged: true,
			Dependencies: []*config.Dependency{{
				From:    "base-image",
				Imports: []*config.DependencyImport{{Type: config.ImageNameImport, TargetBuildArg: "BASE_IMAGE"}},
			}},
		}

		images, err := mapStagedDockerfileDataToImages(context.Background(), data, "Dockerfile", &config.Meta{}, dockerfileConfig, "linux/amd64", false, CommonImageOptions{})
		Expect(err).To(Succeed())
		Expect(images).NotTo(BeEmpty())
		for _, img := range images {
			Expect(img.RequiresResolvedDependencyInputs).To(BeTrue())
		}
	})

	It("marks Stapel images that import dependency values", func(ctx SpecContext) {
		stapelImage := &config.StapelImage{StapelImageBase: &config.StapelImageBase{
			Name: "app",
			From: "scratch",
			Git:  &config.GitManager{},
			Dependencies: []*config.Dependency{{
				From:   "base",
				Before: "install",
				Imports: []*config.DependencyImport{{
					Type:      config.ImageNameImport,
					TargetEnv: "BASE_IMAGE",
				}},
			}},
		}}

		img, err := mapStapelConfigToImage(ctx, &config.Meta{}, stapelImage, "linux/amd64", false, CommonImageOptions{})
		Expect(err).To(Succeed())
		Expect(img.RequiresResolvedDependencyInputs).To(BeTrue())
	})
})
