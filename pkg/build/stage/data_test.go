package stage

import (
	"fmt"

	. "github.com/onsi/gomega"

	"github.com/werf/werf/pkg/config"
	"github.com/werf/werf/pkg/container_backend/stage_builder"
	"github.com/werf/werf/pkg/util"
)

type TestDependencies struct {
	Dependencies   []*TestDependency
	ExpectedDigest string
}

type TestDependency struct {
	ImageName string

	TargetEnvImageName   string
	TargetEnvImageRepo   string
	TargetEnvImageTag    string
	TargetEnvImageID     string
	TargetEnvImageDigest string

	TargetBuildArgImageName   string
	TargetBuildArgImageRepo   string
	TargetBuildArgImageTag    string
	TargetBuildArgImageID     string
	TargetBuildArgImageDigest string

	DockerImageRepo   string
	DockerImageTag    string
	DockerImageID     string
	DockerImageDigest string
}

func (dep *TestDependency) GetDockerImageName() string {
	return fmt.Sprintf("%s:%s", dep.DockerImageRepo, dep.DockerImageTag)
}

func (dep *TestDependency) ToConfigDependency() *config.Dependency {
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
	if dep.TargetEnvImageDigest != "" {
		depCfg.Imports = append(depCfg.Imports, &config.DependencyImport{
			Type:      config.ImageDigestImport,
			TargetEnv: dep.TargetEnvImageDigest,
		})
	}

	if dep.TargetBuildArgImageName != "" {
		depCfg.Imports = append(depCfg.Imports, &config.DependencyImport{
			Type:           config.ImageNameImport,
			TargetBuildArg: dep.TargetBuildArgImageName,
		})
	}
	if dep.TargetBuildArgImageRepo != "" {
		depCfg.Imports = append(depCfg.Imports, &config.DependencyImport{
			Type:           config.ImageRepoImport,
			TargetBuildArg: dep.TargetBuildArgImageRepo,
		})
	}
	if dep.TargetBuildArgImageTag != "" {
		depCfg.Imports = append(depCfg.Imports, &config.DependencyImport{
			Type:           config.ImageTagImport,
			TargetBuildArg: dep.TargetBuildArgImageTag,
		})
	}
	if dep.TargetBuildArgImageID != "" {
		depCfg.Imports = append(depCfg.Imports, &config.DependencyImport{
			Type:           config.ImageIDImport,
			TargetBuildArg: dep.TargetBuildArgImageID,
		})
	}
	if dep.TargetBuildArgImageDigest != "" {
		depCfg.Imports = append(depCfg.Imports, &config.DependencyImport{
			Type:           config.ImageDigestImport,
			TargetBuildArg: dep.TargetBuildArgImageDigest,
		})
	}

	return depCfg
}

func GetConfigDependencies(dependencies []*TestDependency) (res []*config.Dependency) {
	for _, dep := range dependencies {
		res = append(res, dep.ToConfigDependency())
	}

	return
}

func CheckImageDependenciesAfterPrepare(img *LegacyImageStub, stageBuilder *stage_builder.StageBuilder, dependencies []*TestDependency) {
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
		if dep.TargetEnvImageDigest != "" {
			Expect(img._Container._ServiceCommitChangeOptions.Env[dep.TargetEnvImageDigest]).To(Equal(dep.DockerImageDigest))
		}

		if dep.TargetBuildArgImageName != "" {
			Expect(util.IsStringsContainValue(stageBuilder.GetDockerfileBuilderImplementation().BuildDockerfileOptions.BuildArgs, fmt.Sprintf("%s=%s", dep.TargetBuildArgImageName, dep.GetDockerImageName()))).To(BeTrue())
		}
		if dep.TargetBuildArgImageRepo != "" {
			Expect(util.IsStringsContainValue(stageBuilder.GetDockerfileBuilderImplementation().BuildDockerfileOptions.BuildArgs, fmt.Sprintf("%s=%s", dep.TargetBuildArgImageRepo, dep.DockerImageRepo))).To(BeTrue())
		}
		if dep.TargetBuildArgImageTag != "" {
			Expect(util.IsStringsContainValue(stageBuilder.GetDockerfileBuilderImplementation().BuildDockerfileOptions.BuildArgs, fmt.Sprintf("%s=%s", dep.TargetBuildArgImageTag, dep.DockerImageTag))).To(BeTrue())
		}
		if dep.TargetBuildArgImageID != "" {
			Expect(util.IsStringsContainValue(stageBuilder.GetDockerfileBuilderImplementation().BuildDockerfileOptions.BuildArgs, fmt.Sprintf("%s=%s", dep.TargetBuildArgImageID, dep.DockerImageID))).To(BeTrue())
		}
	}
}
