package stage

import (
	"context"
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/werf/werf/pkg/config"
	"github.com/werf/werf/pkg/container_runtime"
)

var _ = Describe("DependenciesStage", func() {
	When("using image dependencies", func() {
		type DepedenciesData struct {
			Dependencies   []*DependencyData
			ExpectedDigest string
		}

		DescribeTable("configuring images dependencies for dependencies stage",
			func(data DepedenciesData) {
				ctx := context.Background()

				conveyor := NewConveyorStubForDependencies(data.Dependencies)

				stage := newDependenciesStage(nil, GetConfigDependencies(data.Dependencies), "example-stage", &NewBaseStageOptions{
					ImageName:   "example-image",
					ProjectName: "example-project",
				})

				img := NewLegacyImageStub()

				digest, err := stage.GetDependencies(ctx, conveyor, nil, img)
				Expect(err).To(Succeed())
				Expect(digest).To(Equal(data.ExpectedDigest))

				err = stage.PrepareImage(ctx, conveyor, nil, img)
				Expect(err).To(Succeed())
				CheckImageDependenciesAfterPrepare(img, data.Dependencies)
			},

			Entry("should calculate basic stage digest when no dependencies are set",
				DepedenciesData{
					ExpectedDigest: "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
				}),

			Entry("should change stage digest and set configured environment variables when dependencies are set",
				DepedenciesData{
					ExpectedDigest: "84f7d49084ba98f8247feba78a217382c6801c7df27cce294566cac69c43d58d",
					Dependencies: []*DependencyData{
						{
							ImageName:          "one",
							TargetEnvImageName: "IMAGE_ONE_NAME",
							TargetEnvImageRepo: "IMAGE_ONE_REPO",
							TargetEnvImageTag:  "IMAGE_ONE_TAG",

							DockerImageRepo: "ONE_REPO",
							DockerImageTag:  "796e905d0cc975e718b3f8b3ea0199ea4d52668ecc12c4dbf85a136d-1638863657513",
							DockerImageID:   "sha256:d19deb06171086017db6aade408ce29592e7490f3b98d4da228ef6c771ddc6d5",
						},
						{
							ImageName:          "two",
							TargetEnvImageName: "TWO_NAME",
							TargetEnvImageRepo: "TWO_REPO",
							TargetEnvImageTag:  "TWO_TAG",
							TargetEnvImageID:   "TWO_ID",

							DockerImageRepo: "TWO_REPO",
							DockerImageTag:  "bc6db8dde5c051349b85dbb8f858f4c80a519a17723d2c67dc9f890c-1643039584147",
							DockerImageID:   "sha256:5a46fe1fe7f2867aeb0a74cfc5aea79b1003b8d6095e2350332d3c99d7e1df6b",
						},
						{
							ImageName:          "one",
							TargetEnvImageName: "ONE_NAME",
							TargetEnvImageRepo: "ONE_REPO",
							TargetEnvImageTag:  "ONE_TAG",
							TargetEnvImageID:   "ONE_ID",

							DockerImageRepo: "ONE_REPO",
							DockerImageTag:  "796e905d0cc975e718b3f8b3ea0199ea4d52668ecc12c4dbf85a136d-1638863657513",
							DockerImageID:   "sha256:d19deb06171086017db6aade408ce29592e7490f3b98d4da228ef6c771ddc6d5",
						},
					},
				}),

			Entry("new image added into dependencies should change stage digest and environment variables",
				DepedenciesData{
					ExpectedDigest: "d1af66208228e2be40cd861ac80d14d068f2c649d9fd345458efe3a48c2927b5",
					Dependencies: []*DependencyData{
						{
							ImageName:          "one",
							TargetEnvImageName: "IMAGE_ONE_NAME",
							TargetEnvImageRepo: "IMAGE_ONE_REPO",
							TargetEnvImageTag:  "IMAGE_ONE_TAG",

							DockerImageRepo: "ONE_REPO",
							DockerImageTag:  "796e905d0cc975e718b3f8b3ea0199ea4d52668ecc12c4dbf85a136d-1638863657513",
							DockerImageID:   "sha256:d19deb06171086017db6aade408ce29592e7490f3b98d4da228ef6c771ddc6d5",
						},
						{
							ImageName:          "two",
							TargetEnvImageName: "TWO_NAME",
							TargetEnvImageRepo: "TWO_REPO",
							TargetEnvImageTag:  "TWO_TAG",
							TargetEnvImageID:   "TWO_ID",

							DockerImageRepo: "TWO_REPO",
							DockerImageTag:  "bc6db8dde5c051349b85dbb8f858f4c80a519a17723d2c67dc9f890c-1643039584147",
							DockerImageID:   "sha256:5a46fe1fe7f2867aeb0a74cfc5aea79b1003b8d6095e2350332d3c99d7e1df6b",
						},
						{
							ImageName:          "one",
							TargetEnvImageName: "ONE_NAME",
							TargetEnvImageRepo: "ONE_REPO",
							TargetEnvImageTag:  "ONE_TAG",
							TargetEnvImageID:   "ONE_ID",

							DockerImageRepo: "ONE_REPO",
							DockerImageTag:  "796e905d0cc975e718b3f8b3ea0199ea4d52668ecc12c4dbf85a136d-1638863657513",
							DockerImageID:   "sha256:d19deb06171086017db6aade408ce29592e7490f3b98d4da228ef6c771ddc6d5",
						},
						{
							ImageName:          "three",
							TargetEnvImageName: "THREE_IMAGE_NAME",

							DockerImageRepo: "THREE_REPO",
							DockerImageTag:  "custom-tag",
							DockerImageID:   "sha256:6f510109a5ca7657babd6f3f48fd16c1b887d63857ac411f636967de5aa48d31",
						},
					},
				}),

			Entry("should change stage digest and environment variables when previously added image dependency params has been changed",
				DepedenciesData{
					ExpectedDigest: "d214e5d775ea7493e2fbe2f1d598d5c613a1c7fd605a55a4c4d98ab9d5161853",
					Dependencies: []*DependencyData{
						{
							ImageName:          "one",
							TargetEnvImageName: "IMAGE_ONE_NAME",
							TargetEnvImageRepo: "IMAGE_ONE_REPO",
							TargetEnvImageTag:  "IMAGE_ONE_TAG",

							DockerImageRepo: "ONE_REPO",
							DockerImageTag:  "b7aebf280be3fbb7d207d3b659bfc1a49338441ea933c1eac5766a5f-1638863693022",
							DockerImageID:   "sha256:c62467775792f47c1bb39ceb5dccdafa02db1734f12c8aa07dbb6d618c501166",
						},
						{
							ImageName:          "two",
							TargetEnvImageName: "TWO_NAME",
							TargetEnvImageRepo: "TWO_REPO",
							TargetEnvImageTag:  "TWO_TAG",
							TargetEnvImageID:   "TWO_ID",

							DockerImageRepo: "TWO_REPO",
							DockerImageTag:  "bc6db8dde5c051349b85dbb8f858f4c80a519a17723d2c67dc9f890c-1643039584147",
							DockerImageID:   "sha256:5a46fe1fe7f2867aeb0a74cfc5aea79b1003b8d6095e2350332d3c99d7e1df6b",
						},
						{
							ImageName:          "one",
							TargetEnvImageName: "ONE_NAME",
							TargetEnvImageRepo: "ONE_REPO",
							TargetEnvImageTag:  "ONE_TAG",
							TargetEnvImageID:   "ONE_ID",

							DockerImageRepo: "ONE_REPO",
							DockerImageTag:  "b7aebf280be3fbb7d207d3b659bfc1a49338441ea933c1eac5766a5f-1638863693022",
							DockerImageID:   "sha256:c62467775792f47c1bb39ceb5dccdafa02db1734f12c8aa07dbb6d618c501166",
						},
					},
				}),

			Entry("should change stage digest and set configured environment variables when dependant image environment variable has been changed",
				DepedenciesData{
					ExpectedDigest: "d0f6634579c776b6db5789d9c20e1f36a4c03bc7057a575d6965e4513fa27f8c",
					Dependencies: []*DependencyData{
						{
							ImageName:          "one",
							TargetEnvImageName: "IMAGE_ONE_NAME",
							TargetEnvImageRepo: "IMAGE_ONE_REPO",
							TargetEnvImageTag:  "IMAGE_ONE_TAG_VARIABLE",

							DockerImageRepo: "ONE_REPO",
							DockerImageTag:  "796e905d0cc975e718b3f8b3ea0199ea4d52668ecc12c4dbf85a136d-1638863657513",
							DockerImageID:   "sha256:d19deb06171086017db6aade408ce29592e7490f3b98d4da228ef6c771ddc6d5",
						},
						{
							ImageName:          "two",
							TargetEnvImageName: "TWO_NAME",
							TargetEnvImageRepo: "TWO_REPO",
							TargetEnvImageTag:  "TWO_TAG",
							TargetEnvImageID:   "TWO_ID",

							DockerImageRepo: "TWO_REPO",
							DockerImageTag:  "bc6db8dde5c051349b85dbb8f858f4c80a519a17723d2c67dc9f890c-1643039584147",
							DockerImageID:   "sha256:5a46fe1fe7f2867aeb0a74cfc5aea79b1003b8d6095e2350332d3c99d7e1df6b",
						},
						{
							ImageName:          "one",
							TargetEnvImageName: "ONE_NAME",
							TargetEnvImageRepo: "ONE_REPO",
							TargetEnvImageTag:  "ONE_TAG",
							TargetEnvImageID:   "ONE_ID",

							DockerImageRepo: "ONE_REPO",
							DockerImageTag:  "796e905d0cc975e718b3f8b3ea0199ea4d52668ecc12c4dbf85a136d-1638863657513",
							DockerImageID:   "sha256:d19deb06171086017db6aade408ce29592e7490f3b98d4da228ef6c771ddc6d5",
						},
					},
				}),
		)
	})
})

type LegacyImageStub struct {
	container_runtime.LegacyImageInterface

	_Container *LegacyContainerStub
}

func NewLegacyImageStub() *LegacyImageStub {
	return &LegacyImageStub{
		_Container: NewLegacyContainerStub(),
	}
}

func (img *LegacyImageStub) Container() container_runtime.LegacyContainer {
	return img._Container
}

type LegacyContainerStub struct {
	container_runtime.LegacyContainer

	_ServiceCommitChangeOptions *LegacyContainerOptionsStub
}

func NewLegacyContainerStub() *LegacyContainerStub {
	return &LegacyContainerStub{
		_ServiceCommitChangeOptions: NewLegacyContainerOptionsStub(),
	}
}

func (c *LegacyContainerStub) ServiceCommitChangeOptions() container_runtime.LegacyContainerOptions {
	return c._ServiceCommitChangeOptions
}

type LegacyContainerOptionsStub struct {
	container_runtime.LegacyContainerOptions

	Env map[string]string
}

func NewLegacyContainerOptionsStub() *LegacyContainerOptionsStub {
	return &LegacyContainerOptionsStub{Env: make(map[string]string)}
}

func (opts *LegacyContainerOptionsStub) AddEnv(envs map[string]string) {
	for k, v := range envs {
		opts.Env[k] = v
	}
}

type ConveyorStub struct {
	Conveyor

	lastStageImageNameByImageName map[string]string
	lastStageImageIDByImageName   map[string]string
}

func NewConveyorStub(lastStageImageNameByImageName, lastStageImageIDByImageName map[string]string) *ConveyorStub {
	return &ConveyorStub{
		lastStageImageNameByImageName: lastStageImageNameByImageName,
		lastStageImageIDByImageName:   lastStageImageIDByImageName,
	}
}

func NewConveyorStubForDependencies(dependencies []*DependencyData) *ConveyorStub {
	lastStageImageNameByImageName := make(map[string]string)
	lastStageImageIDByImageName := make(map[string]string)

	for _, dep := range dependencies {
		lastStageImageNameByImageName[dep.ImageName] = dep.GetDockerImageName()
		lastStageImageIDByImageName[dep.ImageName] = dep.DockerImageID
	}

	return NewConveyorStub(lastStageImageNameByImageName, lastStageImageIDByImageName)
}

func (c *ConveyorStub) GetImageNameForLastImageStage(imageName string) string {
	return c.lastStageImageNameByImageName[imageName]
}

func (c *ConveyorStub) GetImageIDForLastImageStage(imageName string) string {
	return c.lastStageImageIDByImageName[imageName]
}

type DependencyData struct {
	ImageName string

	TargetEnvImageName string
	TargetEnvImageRepo string
	TargetEnvImageTag  string
	TargetEnvImageID   string

	DockerImageRepo string
	DockerImageTag  string
	DockerImageID   string
}

func (dep *DependencyData) GetDockerImageName() string {
	return fmt.Sprintf("%s:%s", dep.DockerImageRepo, dep.DockerImageTag)
}

func (dep *DependencyData) ToConfigDependency() *config.Dependency {
	depCfg := &config.Dependency{ImageName: dep.ImageName}

	if dep.TargetEnvImageName != "" {
		depCfg.Imports = append(depCfg.Imports, &config.DependencyImport{
			Type:      config.ImageNameImport,
			TargetEnv: dep.TargetEnvImageName,
		})
	}
	if dep.TargetEnvImageRepo != "" {
		depCfg.Imports = append(depCfg.Imports, &config.DependencyImport{
			Type:      config.ImageRepoImport,
			TargetEnv: dep.TargetEnvImageRepo,
		})
	}
	if dep.TargetEnvImageTag != "" {
		depCfg.Imports = append(depCfg.Imports, &config.DependencyImport{
			Type:      config.ImageTagImport,
			TargetEnv: dep.TargetEnvImageTag,
		})
	}
	if dep.TargetEnvImageID != "" {
		depCfg.Imports = append(depCfg.Imports, &config.DependencyImport{
			Type:      config.ImageIDImport,
			TargetEnv: dep.TargetEnvImageID,
		})
	}

	return depCfg
}

func GetConfigDependencies(dependencies []*DependencyData) (res []*config.Dependency) {
	for _, dep := range dependencies {
		res = append(res, dep.ToConfigDependency())
	}

	return
}

func CheckImageDependenciesAfterPrepare(img *LegacyImageStub, dependencies []*DependencyData) {
	for _, dep := range dependencies {
		if dep.TargetEnvImageName != "" {
			Expect(img._Container._ServiceCommitChangeOptions.Env[dep.TargetEnvImageName]).To(Equal(dep.GetDockerImageName()))
		}
		if dep.TargetEnvImageRepo != "" {
			Expect(img._Container._ServiceCommitChangeOptions.Env[dep.TargetEnvImageRepo]).To(Equal(dep.DockerImageRepo))
		}
		if dep.TargetEnvImageTag != "" {
			Expect(img._Container._ServiceCommitChangeOptions.Env[dep.TargetEnvImageTag]).To(Equal(dep.DockerImageTag))
		}
		if dep.TargetEnvImageID != "" {
			Expect(img._Container._ServiceCommitChangeOptions.Env[dep.TargetEnvImageID]).To(Equal(dep.DockerImageID))
		}
	}
}
