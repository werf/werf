package stage

import (
	"bytes"
	"context"

	"github.com/moby/buildkit/frontend/dockerfile/instructions"
	"github.com/moby/buildkit/frontend/dockerfile/parser"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/werf/werf/pkg/build/dockerfile_helpers"
	"github.com/werf/werf/pkg/container_backend/stage_builder"
	"github.com/werf/werf/pkg/util"
)

func testDockerfileToDockerStages(dockerfile []byte) ([]instructions.Stage, []instructions.ArgCommand) {
	p, err := parser.Parse(bytes.NewReader(dockerfile))
	Expect(err).To(Succeed())

	dockerStages, dockerMetaArgs, err := instructions.Parse(p.AST)
	Expect(err).To(Succeed())

	dockerfile_helpers.ResolveDockerStagesFromValue(dockerStages)

	return dockerStages, dockerMetaArgs
}

func newTestDockerfileStage(dockerfile []byte, target string, buildArgs map[string]interface{}, dockerStages []instructions.Stage, dockerMetaArgs []instructions.ArgCommand, dependencies []*TestDependency) *DockerfileStage {
	dockerTargetIndex, err := dockerfile_helpers.GetDockerTargetStageIndex(dockerStages, target)
	Expect(err).To(Succeed())

	ds := NewDockerStages(
		dockerStages,
		util.MapStringInterfaceToMapStringString(buildArgs),
		dockerMetaArgs,
		dockerTargetIndex,
	)

	return newDockerfileStage(
		NewDockerRunArgs(
			dockerfile,
			"no-such-path",
			target,
			"",
			nil,
			buildArgs,
			nil,
			"",
			"",
		),
		ds,
		NewContextChecksum(nil),
		&NewBaseStageOptions{
			ImageName:   "example-image",
			ProjectName: "example-project",
		},
		GetConfigDependencies(dependencies),
	)
}

var _ = Describe("DockerfileStage", func() {
	DescribeTable("configuring images dependencies for dockerfile stage",
		func(data TestDockerfileDependencies) {
			ctx := context.Background()

			conveyor := NewConveyorStubForDependencies(NewGiterminismManagerStub(NewLocalGitRepoStub("9d8059842b6fde712c58315ca0ab4713d90761c0")), data.TestDependencies.Dependencies)
			containerBackend := NewContainerBackendMock()

			dockerStages, dockerMetaArgs := testDockerfileToDockerStages(data.Dockerfile)

			stage := newTestDockerfileStage(data.Dockerfile, data.Target, data.BuildArgs, dockerStages, dockerMetaArgs, data.TestDependencies.Dependencies)

			img := NewLegacyImageStub()
			stageBuilder := stage_builder.NewStageBuilder(containerBackend, nil, img)
			stageImage := &StageImage{
				Image:   img,
				Builder: stageBuilder,
			}

			digest, err := stage.GetDependencies(ctx, conveyor, containerBackend, nil, stageImage)
			Expect(err).To(Succeed())
			Expect(digest).To(Equal(data.TestDependencies.ExpectedDigest))

			err = stage.PrepareImage(ctx, conveyor, containerBackend, nil, stageImage)
			Expect(err).To(Succeed())
			CheckImageDependenciesAfterPrepare(img, stageBuilder, data.TestDependencies.Dependencies)
		},

		Entry("should calculate dockerfile stage digest when no dependencies are set",
			TestDockerfileDependencies{
				Dockerfile: []byte(`
FROM alpine:latest
RUN echo hello
`),
				TestDependencies: &TestDependencies{
					ExpectedDigest: "b9d5527ee7a7047747bce5fb5fd1d7ab2b687f141a91151620098b60c2ad0eae",
				},
			}),

		Entry("should not change dockerfile stage digest when dependencies are defined, but build args not used",
			TestDockerfileDependencies{
				Dockerfile: []byte(`
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
							DockerImageID:   "sha256:d19deb06171086017db6aade408ce29592e7490f3b98d4da228ef6c771ddc6d5",
						},
					},
				},
			},
		),

		Entry("should change dockerfile stage digest when dependant image build args used in the Dockerfile",
			TestDockerfileDependencies{
				Dockerfile: []byte(`
FROM alpine:latest

ARG IMAGE_ONE_NAME
ARG IMAGE_ONE_REPO
ARG IMAGE_ONE_TAG
ARG IMAGE_ONE_ID

RUN echo hello
RUN echo {"name": "${IMAGE_ONE_NAME}", "repo": "${IMAGE_ONE_REPO}", "tag": "${IMAGE_ONE_TAG}", "id": "${IMAGE_ONE_ID}"} >> images.json
`),
				TestDependencies: &TestDependencies{
					ExpectedDigest: "b55701cfd33c5931e001e4d8ab24628df571afc7b7f647ca4083dab75aafff4d",
					Dependencies: []*TestDependency{
						{
							ImageName:               "one",
							TargetBuildArgImageName: "IMAGE_ONE_NAME",
							TargetBuildArgImageRepo: "IMAGE_ONE_REPO",
							TargetBuildArgImageTag:  "IMAGE_ONE_TAG",
							TargetBuildArgImageID:   "IMAGE_ONE_ID",

							DockerImageRepo: "ONE_REPO",
							DockerImageTag:  "796e905d0cc975e718b3f8b3ea0199ea4d52668ecc12c4dbf85a136d-1638863657513",
							DockerImageID:   "sha256:d19deb06171086017db6aade408ce29592e7490f3b98d4da228ef6c771ddc6d5",
						},
					},
				},
			},
		),

		Entry("should change dockerfile stage digest when dependant image name changed",
			TestDockerfileDependencies{
				Dockerfile: []byte(`
FROM alpine:latest

ARG IMAGE_ONE_NAME
ARG IMAGE_ONE_REPO
ARG IMAGE_ONE_TAG
ARG IMAGE_ONE_ID

RUN echo hello
RUN echo {"name": "${IMAGE_ONE_NAME}", "repo": "${IMAGE_ONE_REPO}", "tag": "${IMAGE_ONE_TAG}", "id": "${IMAGE_ONE_ID}"} >> images.json
`),
				TestDependencies: &TestDependencies{
					ExpectedDigest: "dac644aa25871d0d902581d9c2a901ef753267082adedc19a3bee23a18cfca17",
					Dependencies: []*TestDependency{
						{
							ImageName:               "one",
							TargetBuildArgImageName: "IMAGE_ONE_NAME",
							TargetBuildArgImageRepo: "IMAGE_ONE_REPO",
							TargetBuildArgImageTag:  "IMAGE_ONE_TAG",
							TargetBuildArgImageID:   "IMAGE_ONE_ID",

							DockerImageRepo: "ONE_REPO",
							DockerImageTag:  "b7aebf280be3fbb7d207d3b659bfc1a49338441ea933c1eac5766a5f-1638863693022",
							DockerImageID:   "sha256:d19deb06171086017db6aade408ce29592e7490f3b98d4da228ef6c771ddc6d5",
						},
					},
				},
			},
		),

		Entry("should change dockerfile stage digest when dependant image id changed",
			TestDockerfileDependencies{
				Dockerfile: []byte(`
FROM alpine:latest

ARG IMAGE_ONE_NAME
ARG IMAGE_ONE_REPO
ARG IMAGE_ONE_TAG
ARG IMAGE_ONE_ID

RUN echo hello
RUN echo {"name": "${IMAGE_ONE_NAME}", "repo": "${IMAGE_ONE_REPO}", "tag": "${IMAGE_ONE_TAG}", "id": "${IMAGE_ONE_ID}"} >> images.json
`),
				TestDependencies: &TestDependencies{
					ExpectedDigest: "914969761b92ec6a4a6eee5ad33c32c2d2c27b0b15fe4abf4f26c30755378ed4",
					Dependencies: []*TestDependency{
						{
							ImageName:               "one",
							TargetBuildArgImageName: "IMAGE_ONE_NAME",
							TargetBuildArgImageRepo: "IMAGE_ONE_REPO",
							TargetBuildArgImageTag:  "IMAGE_ONE_TAG",
							TargetBuildArgImageID:   "IMAGE_ONE_ID",

							DockerImageRepo: "ONE_REPO",
							DockerImageTag:  "b7aebf280be3fbb7d207d3b659bfc1a49338441ea933c1eac5766a5f-1638863693022",
							DockerImageID:   "sha256:44b14c266507626ec1e3f1eb22fcbd9b935595ead56800f77110fc4e1e95689c",
						},
					},
				},
			},
		),

		Entry("should calculate dockerfile stage digest when no dependencies are set",
			TestDockerfileDependencies{
				Dockerfile: []byte(`
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
				Dockerfile: []byte(`
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
							DockerImageID:   "sha256:d13c942271d66cb0954c3ba93e143cd253421fe0772b8bed32c4c0077a546d4d",
						},
					},
				},
			},
		),

		Entry("should change dockerfile stage digest when base dependency image has changed",
			TestDockerfileDependencies{
				Dockerfile: []byte(`
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
							DockerImageID:   "sha256:5d0da3dc976460b72c77d94c8a1ad043720b0416bfc16c52c45d4847e53fadb6",
						},
					},
				},
			},
		),
	)

	When("Dockerfile uses undefined build argument", func() {
		It("should report descriptive error when fetching dockerfile stage dependencies", func() {
			dockerfile := []byte(`
ARG BASE_NAME=alpine:latest

FROM ${BASE_NAME1}
RUN echo hello
`)

			ctx := context.Background()

			conveyor := NewConveyorStubForDependencies(NewGiterminismManagerStub(NewLocalGitRepoStub("9d8059842b6fde712c58315ca0ab4713d90761c0")), nil)

			dockerStages, dockerMetaArgs := testDockerfileToDockerStages(dockerfile)

			stage := newTestDockerfileStage(dockerfile, "", nil, dockerStages, dockerMetaArgs, nil)

			containerBackend := NewContainerBackendMock()

			dockerRegistry := NewDockerRegistryApiStub()

			err := stage.FetchDependencies(ctx, conveyor, containerBackend, dockerRegistry)
			Expect(IsErrInvalidBaseImage(err)).To(BeTrue())
		})
	})
})

type TestDockerfileDependencies struct {
	Dockerfile []byte
	Target     string
	BuildArgs  map[string]interface{}

	TestDependencies *TestDependencies
}
