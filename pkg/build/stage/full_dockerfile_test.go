package stage

import (
	"bytes"
	"context"
	"fmt"

	"github.com/moby/buildkit/frontend/dockerfile/instructions"
	"github.com/moby/buildkit/frontend/dockerfile/parser"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/werf/common-go/pkg/util"
	"github.com/werf/werf/v2/pkg/container_backend/stage_builder"
	"github.com/werf/werf/v2/pkg/dockerfile/frontend"
	"github.com/werf/werf/v2/pkg/image"
	"github.com/werf/werf/v2/pkg/logging"
)

func testDockerfileToDockerStages(dockerfileData []byte) ([]instructions.Stage, []instructions.ArgCommand) {
	p, err := parser.Parse(bytes.NewReader(dockerfileData))
	Expect(err).To(Succeed())

	dockerStages, dockerMetaArgs, err := instructions.Parse(p.AST, nil)
	Expect(err).To(Succeed())

	frontend.ResolveDockerStagesFromValue(dockerStages)

	return dockerStages, dockerMetaArgs
}

func newTestFullDockerfileStage(dockerfileData []byte, target string, buildArgs map[string]interface{}, dockerStages []instructions.Stage, dockerMetaArgs []instructions.ArgCommand, dependencies []*TestDependency, imageCacheVersion string) *FullDockerfileStage {
	dockerTargetIndex, err := frontend.GetDockerTargetStageIndex(dockerStages, target)
	Expect(err).To(Succeed())

	ds := NewDockerStages(
		dockerStages,
		util.MapStringInterfaceToMapStringString(buildArgs),
		dockerMetaArgs,
		dockerTargetIndex,
	)

	return newFullDockerfileStage(NewDockerRunArgs(
		dockerfileData,
		"no-such-path",
		target,
		"",
		nil,
		buildArgs,
		nil,
		"",
		nil,
	), ds, NewContextChecksum(nil), &BaseStageOptions{
		ImageName:   "example-image",
		ProjectName: "example-project",
	}, GetConfigDependencies(dependencies), imageCacheVersion)
}

var _ = Describe("FullDockerfileStage", func() {
	DescribeTable("configuring images dependencies for dockerfile stage",
		func(ctx SpecContext, data TestDockerfileDependencies) {
			conveyor := NewConveyorStubForDependencies(NewGiterminismManagerStub(NewLocalGitRepoStub("9d8059842b6fde712c58315ca0ab4713d90761c0"), NewGiterminismInspectorStub()), data.TestDependencies.Dependencies)
			containerBackend := NewContainerBackendStub()

			dockerStages, dockerMetaArgs := testDockerfileToDockerStages(data.DockerfileData)

			stage := newTestFullDockerfileStage(data.DockerfileData, data.Target, data.BuildArgs, dockerStages, dockerMetaArgs, data.TestDependencies.Dependencies, data.TestDependencies.ImageCacheVersion)

			img := NewLegacyImageStub()
			stageBuilder := stage_builder.NewStageBuilder(containerBackend, "", img)
			stageImage := &StageImage{
				Image:   img,
				Builder: stageBuilder,
			}

			digest, err := stage.GetDependencies(ctx, conveyor, containerBackend, nil, stageImage, nil)
			Expect(err).To(Succeed())
			fmt.Printf("calculated digest: %s\n", digest)
			fmt.Printf("expected digest: %s\n", data.TestDependencies.ExpectedDigest)
			Expect(digest).To(Equal(data.TestDependencies.ExpectedDigest))

			err = stage.PrepareImage(ctx, conveyor, containerBackend, nil, stageImage, nil)
			Expect(err).To(Succeed())
			CheckImageDependenciesAfterPrepare(img, stageBuilder, data.TestDependencies.Dependencies)
		},

		Entry("should calculate dockerfile stage digest when no dependencies are set",
			TestDockerfileDependencies{
				DockerfileData: []byte(`
FROM alpine:latest
RUN echo hello
`),
				TestDependencies: &TestDependencies{
					ExpectedDigest: "b9d5527ee7a7047747bce5fb5fd1d7ab2b687f141a91151620098b60c2ad0eae",
				},
			}),

		Entry("should not change dockerfile stage digest when dependencies are defined, but build args not used",
			TestDockerfileDependencies{
				DockerfileData: []byte(`
FROM alpine:latest
RUN echo hello
`),
				TestDependencies: &TestDependencies{
					ExpectedDigest: "b9d5527ee7a7047747bce5fb5fd1d7ab2b687f141a91151620098b60c2ad0eae",
					Dependencies: []*TestDependency{
						{
							ImageName:               "one",
							TargetBuildArgImageName: "IMAGE_ONE_NAME",

							DockerImageRepo: "ONE_REPO",
							DockerImageTag:  "796e905d0cc975e718b3f8b3ea0199ea4d52668ecc12c4dbf85a136d-1638863657513",
						},
					},
				},
			},
		),

		Entry("should change dockerfile stage digest when dependant image build args used in the Dockerfile",
			TestDockerfileDependencies{
				DockerfileData: []byte(`
FROM alpine:latest

ARG IMAGE_ONE_NAME
ARG IMAGE_ONE_REPO
ARG IMAGE_ONE_TAG

RUN echo hello
RUN echo {"name": "${IMAGE_ONE_NAME}", "repo": "${IMAGE_ONE_REPO}", "tag": "${IMAGE_ONE_TAG}"} >> images.json
`),
				TestDependencies: &TestDependencies{
					ExpectedDigest: "6ab9de52e1aa389b7e0f684b052c8047dc6c1e01709e1daacc305e4c23211941",
					Dependencies: []*TestDependency{
						{
							ImageName:               "one",
							TargetBuildArgImageName: "IMAGE_ONE_NAME",
							TargetBuildArgImageRepo: "IMAGE_ONE_REPO",
							TargetBuildArgImageTag:  "IMAGE_ONE_TAG",

							DockerImageRepo: "ONE_REPO",
							DockerImageTag:  "796e905d0cc975e718b3f8b3ea0199ea4d52668ecc12c4dbf85a136d-1638863657513",
						},
					},
				},
			},
		),

		Entry("should calculate dockerfile stage digest when no dependencies are set",
			TestDockerfileDependencies{
				DockerfileData: []byte(`
ARG BASE_IMAGE=alpine:latest

FROM ${BASE_IMAGE}
RUN echo hello
`),
				TestDependencies: &TestDependencies{
					ExpectedDigest: "b9d5527ee7a7047747bce5fb5fd1d7ab2b687f141a91151620098b60c2ad0eae",
				},
			}),

		Entry("should allow usage of dependency image as a base image in the dockerfile",
			TestDockerfileDependencies{
				DockerfileData: []byte(`
ARG BASE_IMAGE=alpine:latest

FROM ${BASE_IMAGE}
RUN echo hello
`),
				TestDependencies: &TestDependencies{
					ExpectedDigest: "573c4bd0f7480e27c266d55d3a020c7ec4acaebebf897d29cad78fded3b725c7",
					Dependencies: []*TestDependency{
						{
							ImageName:               "two",
							TargetBuildArgImageName: "BASE_IMAGE",

							DockerImageRepo: "ubuntu",
							DockerImageTag:  "latest",
						},
					},
				},
			},
		),

		Entry("should change dockerfile stage digest when base dependency image has changed",
			TestDockerfileDependencies{
				DockerfileData: []byte(`
ARG BASE_IMAGE=alpine:latest

FROM ${BASE_IMAGE}
RUN echo hello
`),
				TestDependencies: &TestDependencies{
					ExpectedDigest: "5b66aa2c1c9f0bf3a9089c52f04ddf4e47af055c4d1fe69d272cba24df372121",
					Dependencies: []*TestDependency{
						{
							ImageName:               "two",
							TargetBuildArgImageName: "BASE_IMAGE",

							DockerImageRepo: "centos",
							DockerImageTag:  "latest",
						},
					},
				},
			},
		),

		Entry("should change dockerfile stage digest when image cache version specified",
			TestDockerfileDependencies{
				DockerfileData: []byte(`
ARG BASE_IMAGE=alpine:latest

FROM ${BASE_IMAGE}
RUN echo hello
`),
				TestDependencies: &TestDependencies{
					ExpectedDigest:    "08298d918e51f6572692ca642027870539a71c46beafd403719360928ffe11de",
					ImageCacheVersion: "image-cache-version",
					Dependencies:      []*TestDependency{},
				},
			},
		),
	)

	When("Dockerfile uses undefined build argument", func() {
		It("should report descriptive error when fetching dockerfile stage dependencies", func(ctx SpecContext) {
			dockerfile := []byte(`
ARG BASE_NAME=alpine:latest

FROM ${BASE_NAME1}
RUN echo hello
`)

			conveyor := NewConveyorStubForDependencies(NewGiterminismManagerStub(NewLocalGitRepoStub("9d8059842b6fde712c58315ca0ab4713d90761c0"), NewGiterminismInspectorStub()), nil)

			dockerStages, dockerMetaArgs := testDockerfileToDockerStages(dockerfile)

			stage := newTestFullDockerfileStage(dockerfile, "", nil, dockerStages, dockerMetaArgs, nil, "")

			containerBackend := NewContainerBackendStub()

			_, err := stage.GetDependencies(ctx, conveyor, containerBackend, nil, nil, nil)
			Expect(IsErrInvalidBaseImage(err)).To(BeTrue())
		})
	})

	When("head commit is empty", func() {
		It("should not append project repo commit label", func(ctx SpecContext) {
			dockerfile := []byte(`
FROM alpine:latest
RUN echo hello
`)

			conveyor := NewConveyorStubForDependencies(NewGiterminismManagerStub(NewLocalGitRepoStub(""), NewGiterminismInspectorStub()), nil)
			containerBackend := NewContainerBackendStub()
			dockerStages, dockerMetaArgs := testDockerfileToDockerStages(dockerfile)
			stage := newTestFullDockerfileStage(dockerfile, "", nil, dockerStages, dockerMetaArgs, nil, "")
			img := NewLegacyImageStub()
			stageBuilder := stage_builder.NewStageBuilder(containerBackend, "", img)
			stageImage := &StageImage{Image: img, Builder: stageBuilder}

			Expect(stage.PrepareImage(ctx, conveyor, containerBackend, nil, stageImage, nil)).To(Succeed())
			Expect(stageBuilder.GetDockerfileBuilderImplementation().BuildDockerfileOptions.Labels).NotTo(ContainElement(HavePrefix(fmt.Sprintf("%s=", image.WerfProjectRepoCommitLabel))))
		})
	})

	When("Dockerfile uses run with mount from another stage", func() {
		It("should change dockerfile stage digest when base stage context has changed", func(ctx context.Context) {
			dockerfile := []byte(`
FROM alpnie:latest AS build
WORKDIR /usr/local/test_project
COPY . .
RUN mkdir -p dist && \
    cp -v main.py dist/prog.py

FROM alpine:latest
RUN --mount=type=bind,from=build,source=/usr/local/test_project/dist,target=/usr/test_project/dist \
    cp -v /usr/test_project/dist/prog.py /usr/local/bin/prog
`)

			ctx = logging.WithLogger(ctx)

			gitRepoStub := NewLocalGitRepoStub("9d8059842b6fde712c58315ca0ab4713d90761c0")

			conveyor := NewConveyorStubForDependencies(NewGiterminismManagerStub(gitRepoStub, NewGiterminismInspectorStub()), nil)

			dockerStages, dockerMetaArgs := testDockerfileToDockerStages(dockerfile)

			stage := newTestFullDockerfileStage(dockerfile, "", nil, dockerStages, dockerMetaArgs, nil, "")

			containerBackend := NewContainerBackendStub()

			img := NewLegacyImageStub()
			stageBuilder := stage_builder.NewStageBuilder(containerBackend, "", img)
			stageImage := &StageImage{
				Image:   img,
				Builder: stageBuilder,
			}

			{
				digest, err := stage.GetDependencies(ctx, conveyor, containerBackend, nil, stageImage, nil)
				Expect(err).To(Succeed())
				Expect(digest).To(Equal("65d219096bc3718c101995b00584d700de791027f2e2ca00635e428932478a1c"))
			}

			gitRepoStub.headCommitHash = "23a0884072c0d31b7c42dfaa7f0772cbfa33ec75"
			{
				digest, err := stage.GetDependencies(ctx, conveyor, containerBackend, nil, stageImage, nil)
				Expect(err).To(Succeed())
				Expect(digest).To(Equal("beb818f2c49f6501194c72449aff59e80be61b405ef39581b01dbf68da927609"))
			}
		})
	})
})

type TestDockerfileDependencies struct {
	DockerfileData []byte
	Target         string
	BuildArgs      map[string]interface{}

	TestDependencies  *TestDependencies
	ImageCacheVersion string
}
