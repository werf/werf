package stage

import (
	"context"
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/werf/werf/pkg/config"
	"github.com/werf/werf/pkg/container_backend/stage_builder"
)

var _ = Describe("DependenciesStage", func() {
	DescribeTable("configuring images dependencies for dependencies stage",
		func(data TestDependencies) {
			ctx := context.Background()

			conveyor := NewConveyorStubForDependencies(NewGiterminismManagerStub(NewLocalGitRepoStub("9d8059842b6fde712c58315ca0ab4713d90761c0"), NewGiterminismInspectorStub()), data.Dependencies)
			containerBackend := NewContainerBackendStub()

			stage := newDependenciesStage(nil, GetConfigDependencies(data.Dependencies), "example-stage", &BaseStageOptions{
				ImageName:   "example-image",
				ProjectName: "example-project",
			})

			img := NewLegacyImageStub()
			stageBuilder := stage_builder.NewStageBuilder(containerBackend, "", img)
			stageImage := &StageImage{
				Image:   img,
				Builder: stageBuilder,
			}

			digest, err := stage.GetDependencies(ctx, conveyor, containerBackend, nil, stageImage, nil)
			Expect(err).To(Succeed())

			fmt.Printf("Calculated digest: %q\n", digest)
			fmt.Printf("Expected digest: %q\n", data.ExpectedDigest)
			Expect(digest).To(Equal(data.ExpectedDigest))

			err = stage.PrepareImage(ctx, conveyor, containerBackend, nil, stageImage, nil)
			Expect(err).To(Succeed())
			CheckImageDependenciesAfterPrepare(img, stageBuilder, data.Dependencies)
		},

		Entry("should calculate basic stage digest when no dependencies are set",
			TestDependencies{
				ExpectedDigest: "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
			}),

		Entry("should change stage digest and set configured environment variables when dependencies are set",
			TestDependencies{
				ExpectedDigest: "62b956afc89d0918164545c6336ecbcf3d29415cb42724368c6d331439bee956",
				Dependencies: []*TestDependency{
					{
						ImageName:          "one",
						TargetEnvImageName: "IMAGE_ONE_NAME",
						TargetEnvImageRepo: "IMAGE_ONE_REPO",
						TargetEnvImageTag:  "IMAGE_ONE_TAG",

						DockerImageRepo:   "ONE_REPO",
						DockerImageTag:    "796e905d0cc975e718b3f8b3ea0199ea4d52668ecc12c4dbf85a136d-1638863657513",
						DockerImageID:     "sha256:d19deb06171086017db6aade408ce29592e7490f3b98d4da228ef6c771ddc6d5",
						DockerImageDigest: "sha256:6a86a39f70f4dac3df671119ffe66a1d76958e7504e72b1ee9f893a152ef772b",
					},
					{
						ImageName:            "two",
						TargetEnvImageName:   "TWO_NAME",
						TargetEnvImageRepo:   "TWO_REPO",
						TargetEnvImageTag:    "TWO_TAG",
						TargetEnvImageID:     "TWO_ID",
						TargetEnvImageDigest: "TWO_DIGEST",

						DockerImageRepo:   "TWO_REPO",
						DockerImageTag:    "bc6db8dde5c051349b85dbb8f858f4c80a519a17723d2c67dc9f890c-1643039584147",
						DockerImageID:     "sha256:5a46fe1fe7f2867aeb0a74cfc5aea79b1003b8d6095e2350332d3c99d7e1df6b",
						DockerImageDigest: "sha256:0476c17a17b746284ea1622b4c97f8a9c986a1f1919ea3a9763cf06d8609b425",
					},
					{
						ImageName:            "one",
						TargetEnvImageName:   "ONE_NAME",
						TargetEnvImageRepo:   "ONE_REPO",
						TargetEnvImageTag:    "ONE_TAG",
						TargetEnvImageID:     "ONE_ID",
						TargetEnvImageDigest: "ONE_DIGEST",

						DockerImageRepo:   "ONE_REPO",
						DockerImageTag:    "796e905d0cc975e718b3f8b3ea0199ea4d52668ecc12c4dbf85a136d-1638863657513",
						DockerImageID:     "sha256:d19deb06171086017db6aade408ce29592e7490f3b98d4da228ef6c771ddc6d5",
						DockerImageDigest: "sha256:87f5ff85f0ff92c6185e6267a2039eff406337a5726c6b668831cdf1262b76e8",
					},
				},
			}),

		Entry("new image added into dependencies should change stage digest and environment variables",
			TestDependencies{
				ExpectedDigest: "69812782590820b507d9c8a2f74ed54d6544070ea28028e6fced4bf70f40112e",
				Dependencies: []*TestDependency{
					{
						ImageName:          "one",
						TargetEnvImageName: "IMAGE_ONE_NAME",
						TargetEnvImageRepo: "IMAGE_ONE_REPO",
						TargetEnvImageTag:  "IMAGE_ONE_TAG",

						DockerImageRepo:   "ONE_REPO",
						DockerImageTag:    "796e905d0cc975e718b3f8b3ea0199ea4d52668ecc12c4dbf85a136d-1638863657513",
						DockerImageID:     "sha256:d19deb06171086017db6aade408ce29592e7490f3b98d4da228ef6c771ddc6d5",
						DockerImageDigest: "sha256:6a86a39f70f4dac3df671119ffe66a1d76958e7504e72b1ee9f893a152ef772b",
					},
					{
						ImageName:            "two",
						TargetEnvImageName:   "TWO_NAME",
						TargetEnvImageRepo:   "TWO_REPO",
						TargetEnvImageTag:    "TWO_TAG",
						TargetEnvImageID:     "TWO_ID",
						TargetEnvImageDigest: "TWO_DIGEST",

						DockerImageRepo:   "TWO_REPO",
						DockerImageTag:    "bc6db8dde5c051349b85dbb8f858f4c80a519a17723d2c67dc9f890c-1643039584147",
						DockerImageID:     "sha256:5a46fe1fe7f2867aeb0a74cfc5aea79b1003b8d6095e2350332d3c99d7e1df6b",
						DockerImageDigest: "sha256:0476c17a17b746284ea1622b4c97f8a9c986a1f1919ea3a9763cf06d8609b425",
					},
					{
						ImageName:            "one",
						TargetEnvImageName:   "ONE_NAME",
						TargetEnvImageRepo:   "ONE_REPO",
						TargetEnvImageTag:    "ONE_TAG",
						TargetEnvImageID:     "ONE_ID",
						TargetEnvImageDigest: "ONE_DIGEST",

						DockerImageRepo:   "ONE_REPO",
						DockerImageTag:    "796e905d0cc975e718b3f8b3ea0199ea4d52668ecc12c4dbf85a136d-1638863657513",
						DockerImageID:     "sha256:d19deb06171086017db6aade408ce29592e7490f3b98d4da228ef6c771ddc6d5",
						DockerImageDigest: "sha256:6a86a39f70f4dac3df671119ffe66a1d76958e7504e72b1ee9f893a152ef772b",
					},
					{
						ImageName:          "three",
						TargetEnvImageName: "THREE_IMAGE_NAME",

						DockerImageRepo:   "THREE_REPO",
						DockerImageTag:    "custom-tag",
						DockerImageID:     "sha256:6f510109a5ca7657babd6f3f48fd16c1b887d63857ac411f636967de5aa48d31",
						DockerImageDigest: "sha256:bc18c2bf466481a9773822f3d29f681d866f7291895552609e75f2e7d76b9bcb",
					},
				},
			}),

		Entry("should change stage digest and environment variables when previously added image dependency params has been changed",
			TestDependencies{
				ExpectedDigest: "686672be02deb993edb8d41cd97eafdf26ebcfa5b878063ddd887ec623d84a1c",
				Dependencies: []*TestDependency{
					{
						ImageName:          "one",
						TargetEnvImageName: "IMAGE_ONE_NAME",
						TargetEnvImageRepo: "IMAGE_ONE_REPO",
						TargetEnvImageTag:  "IMAGE_ONE_TAG",

						DockerImageRepo:   "ONE_REPO",
						DockerImageTag:    "b7aebf280be3fbb7d207d3b659bfc1a49338441ea933c1eac5766a5f-1638863693022",
						DockerImageID:     "sha256:c62467775792f47c1bb39ceb5dccdafa02db1734f12c8aa07dbb6d618c501166",
						DockerImageDigest: "sha256:4d0e8f47643342b529b426aebcaac5c67d4744ee2ba54f967433e6b6fc075312",
					},
					{
						ImageName:            "two",
						TargetEnvImageName:   "TWO_NAME",
						TargetEnvImageRepo:   "TWO_REPO",
						TargetEnvImageTag:    "TWO_TAG",
						TargetEnvImageID:     "TWO_ID",
						TargetEnvImageDigest: "TWO_DIGEST",

						DockerImageRepo:   "TWO_REPO",
						DockerImageTag:    "bc6db8dde5c051349b85dbb8f858f4c80a519a17723d2c67dc9f890c-1643039584147",
						DockerImageID:     "sha256:5a46fe1fe7f2867aeb0a74cfc5aea79b1003b8d6095e2350332d3c99d7e1df6b",
						DockerImageDigest: "sha256:0476c17a17b746284ea1622b4c97f8a9c986a1f1919ea3a9763cf06d8609b425",
					},
					{
						ImageName:            "one",
						TargetEnvImageName:   "ONE_NAME",
						TargetEnvImageRepo:   "ONE_REPO",
						TargetEnvImageTag:    "ONE_TAG",
						TargetEnvImageID:     "ONE_ID",
						TargetEnvImageDigest: "ONE_DIGEST",

						DockerImageRepo:   "ONE_REPO",
						DockerImageTag:    "b7aebf280be3fbb7d207d3b659bfc1a49338441ea933c1eac5766a5f-1638863693022",
						DockerImageID:     "sha256:c62467775792f47c1bb39ceb5dccdafa02db1734f12c8aa07dbb6d618c501166",
						DockerImageDigest: "sha256:4d0e8f47643342b529b426aebcaac5c67d4744ee2ba54f967433e6b6fc075312",
					},
				},
			}),

		Entry("should change stage digest and set configured environment variables when dependant image environment variable has been changed",
			TestDependencies{
				ExpectedDigest: "1aa753aa608f9f19c85732d338e2f96236c2d2463de776b84936bedade37e9ce",
				Dependencies: []*TestDependency{
					{
						ImageName:          "one",
						TargetEnvImageName: "IMAGE_ONE_NAME",
						TargetEnvImageRepo: "IMAGE_ONE_REPO",
						TargetEnvImageTag:  "IMAGE_ONE_TAG_VARIABLE",

						DockerImageRepo:   "ONE_REPO",
						DockerImageTag:    "796e905d0cc975e718b3f8b3ea0199ea4d52668ecc12c4dbf85a136d-1638863657513",
						DockerImageID:     "sha256:d19deb06171086017db6aade408ce29592e7490f3b98d4da228ef6c771ddc6d5",
						DockerImageDigest: "sha256:6a86a39f70f4dac3df671119ffe66a1d76958e7504e72b1ee9f893a152ef772b",
					},
					{
						ImageName:            "two",
						TargetEnvImageName:   "TWO_NAME",
						TargetEnvImageRepo:   "TWO_REPO",
						TargetEnvImageTag:    "TWO_TAG",
						TargetEnvImageID:     "TWO_ID",
						TargetEnvImageDigest: "TWO_DIGEST",

						DockerImageRepo:   "TWO_REPO",
						DockerImageTag:    "bc6db8dde5c051349b85dbb8f858f4c80a519a17723d2c67dc9f890c-1643039584147",
						DockerImageID:     "sha256:5a46fe1fe7f2867aeb0a74cfc5aea79b1003b8d6095e2350332d3c99d7e1df6b",
						DockerImageDigest: "sha256:0476c17a17b746284ea1622b4c97f8a9c986a1f1919ea3a9763cf06d8609b425",
					},
					{
						ImageName:            "one",
						TargetEnvImageName:   "ONE_NAME",
						TargetEnvImageRepo:   "ONE_REPO",
						TargetEnvImageTag:    "ONE_TAG",
						TargetEnvImageID:     "ONE_ID",
						TargetEnvImageDigest: "ONE_DIGEST",

						DockerImageRepo:   "ONE_REPO",
						DockerImageTag:    "796e905d0cc975e718b3f8b3ea0199ea4d52668ecc12c4dbf85a136d-1638863657513",
						DockerImageID:     "sha256:d19deb06171086017db6aade408ce29592e7490f3b98d4da228ef6c771ddc6d5",
						DockerImageDigest: "sha256:6a86a39f70f4dac3df671119ffe66a1d76958e7504e72b1ee9f893a152ef772b",
					},
				},
			}),
	)
})

var _ = Describe("getDependencies helper", func() {
	When("using stapel image dependencies", func() {
		It("selects dependencies which are suitable for specified stage", func() {
			img := &config.StapelImageBase{
				Dependencies: []*config.Dependency{
					{
						ImageName: "one",
						Before:    "setup",
					},
					{
						ImageName: "two",
						Before:    "setup",
					},
					{
						ImageName: "three",
						Before:    "install",
					},
					{
						ImageName: "four",
						After:     "install",
					},
					{
						ImageName: "five",
						After:     "install",
					},
					{
						ImageName: "six",
						After:     "setup",
					},
				},
			}

			{
				deps := getDependencies(img, &getImportsOptions{Before: "install"})
				Expect(len(deps)).To(Equal(1))
				Expect(deps[0].ImageName).To(Equal("three"))
			}

			{
				deps := getDependencies(img, &getImportsOptions{After: "install"})
				Expect(len(deps)).To(Equal(2))
				Expect(deps[0].ImageName).To(Equal("four"))
				Expect(deps[1].ImageName).To(Equal("five"))
			}

			{
				deps := getDependencies(img, &getImportsOptions{Before: "setup"})
				Expect(len(deps)).To(Equal(2))
				Expect(deps[0].ImageName).To(Equal("one"))
				Expect(deps[1].ImageName).To(Equal("two"))
			}

			{
				deps := getDependencies(img, &getImportsOptions{After: "setup"})
				Expect(len(deps)).To(Equal(1))
				Expect(deps[0].ImageName).To(Equal("six"))
			}
		})
	})
})

func NewConveyorStubForDependencies(giterminismManager *GiterminismManagerStub, dependencies []*TestDependency) *ConveyorStub {
	lastStageImageNameByImageName := make(map[string]string)
	lastStageImageIDByImageName := make(map[string]string)
	lastStageImageDigestByImageName := make(map[string]string)

	for _, dep := range dependencies {
		lastStageImageNameByImageName[dep.ImageName] = dep.GetDockerImageName()
		lastStageImageIDByImageName[dep.ImageName] = dep.DockerImageID
		lastStageImageDigestByImageName[dep.ImageName] = dep.DockerImageDigest
	}

	return NewConveyorStub(giterminismManager, lastStageImageNameByImageName, lastStageImageIDByImageName, lastStageImageDigestByImageName)
}
