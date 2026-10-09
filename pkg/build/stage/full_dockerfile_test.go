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
	"github.com/werf/werf/v3/pkg/container_backend/stage_builder"
	"github.com/werf/werf/v3/pkg/dockerfile/frontend"
	"github.com/werf/werf/v3/pkg/git_repo"
	"github.com/werf/werf/v3/pkg/image"
	"github.com/werf/werf/v3/pkg/logging"
	"github.com/werf/werf/v3/pkg/path_matcher"
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
					ExpectedDigest: "a72bb11e02d6c1fe6b27f7da25ce8aab93b0daf5f747a1176d23d78a8a70b29c",
				},
			}),

		Entry("should not change dockerfile stage digest when dependencies are defined, but build args not used",
			TestDockerfileDependencies{
				DockerfileData: []byte(`
FROM alpine:latest
RUN echo hello
`),
				TestDependencies: &TestDependencies{
					ExpectedDigest: "a72bb11e02d6c1fe6b27f7da25ce8aab93b0daf5f747a1176d23d78a8a70b29c",
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
					ExpectedDigest: "6c35a67bea90301ea5e3597333797fcd4e524c2472ac45ad28e76f24bac17be3",
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
					ExpectedDigest: "a72bb11e02d6c1fe6b27f7da25ce8aab93b0daf5f747a1176d23d78a8a70b29c",
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
					ExpectedDigest: "603c1774c2be380583f4a82867f22ab8881151d9c3ffb724a48418548fc17658",
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
					ExpectedDigest: "088b32459d71b7d64fc7553f7f4e93b4c68f1b8fc89e7f258944333743c69ad0",
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
					ExpectedDigest:    "a5b64163d3563f09df25d9595092b23ec12dfa6af47b5db93729b137d50b6819",
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

	When("COPY source depends on inherited base image ENV", func() {
		It("checksums the whole context until the source can be resolved", func(ctx SpecContext) {
			dockerfile := []byte(`
FROM alpine AS base
ENV SRC=payload
FROM base
COPY $SRC/file /payload
`)
			dockerStages, dockerMetaArgs := testDockerfileToDockerStages(dockerfile)
			stage := newTestFullDockerfileStage(dockerfile, "", nil, dockerStages, dockerMetaArgs, nil, "")
			stage.dockerignorePathMatcher = path_matcher.NewTruePathMatcher()
			gitRepo := &pathMatchingGitRepoStub{LocalGitRepoStub: NewLocalGitRepoStub("commit"), path: "payload"}
			conveyor := NewConveyorStubForDependencies(NewGiterminismManagerStub(gitRepo, NewGiterminismInspectorStub()), nil)

			_, err := stage.GetContentDependencies(ctx, conveyor, nil)
			Expect(err).To(Succeed())
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
				Expect(digest).To(Equal("5ba1cc3b688f4443fa48917fd7ac9528f3fde887d1f9e89945e9ba0673d16349"), "digest: %s", digest)
			}

			gitRepoStub.headCommitHash = "23a0884072c0d31b7c42dfaa7f0772cbfa33ec75"
			{
				digest, err := stage.GetDependencies(ctx, conveyor, containerBackend, nil, stageImage, nil)
				Expect(err).To(Succeed())
				Expect(digest).To(Equal("3b37d1ea006e26fc009e7251d605d3e6098850c04a47af06505e7cfd3b1c623d"), "digest: %s", digest)
			}
		})
	})
})

type pathMatchingGitRepoStub struct {
	*LocalGitRepoStub
	path string
}

func (repo *pathMatchingGitRepoStub) GetOrCreateChecksum(ctx context.Context, opts git_repo.ChecksumOptions) (string, error) {
	Expect(opts.LsTreeOptions.PathMatcher.IsPathMatched(repo.path)).To(BeTrue())
	return repo.LocalGitRepoStub.GetOrCreateChecksum(ctx, opts)
}

type TestDockerfileDependencies struct {
	DockerfileData []byte
	Target         string
	BuildArgs      map[string]interface{}

	TestDependencies  *TestDependencies
	ImageCacheVersion string
}
