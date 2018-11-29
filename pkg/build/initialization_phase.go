package build

import (
	"reflect"

	"github.com/flant/dapp/pkg/build/builder"
	"github.com/flant/dapp/pkg/build/stage"
	"github.com/flant/dapp/pkg/config"
	"github.com/flant/dapp/pkg/git_repo"
)

type InitializationPhase struct{}

func NewInitializationPhase() *InitializationPhase {
	return &InitializationPhase{}
}

func (p *InitializationPhase) Run(c *Conveyor) error {
	c.DimgsInOrder = generateDimgsInOrder(c.Dappfile, c)
	return nil
}

func generateDimgsInOrder(dappfile []*config.Dimg, c *Conveyor) []*Dimg {
	var dimgs []*Dimg
	for _, dimgConfig := range getDimgConfigsInOrder(dappfile) {
		dimg := &Dimg{}
		dimg.SetStages(generateStages(dimgConfig, c))
		dimgs = append(dimgs, dimg)
	}

	return dimgs
}

func getDimgConfigsInOrder(dappfile []*config.Dimg) []config.DimgInterface {
	var dimgConfigs []config.DimgInterface
	for _, dimg := range dappfile {
		relatedDimgs := dimg.RelatedDimgs()
		for i := len(relatedDimgs) - 1; i > 0; i-- {
			if isNotInArr(dimgConfigs, relatedDimgs[i]) {
				dimgConfigs = append(dimgConfigs, relatedDimgs[i])
			}
		}
	}

	return dimgConfigs
}

func isNotInArr(arr []config.DimgInterface, obj config.DimgInterface) bool {
	for _, elm := range arr {
		if reflect.DeepEqual(elm, obj) {
			return false
		}
	}

	return true
}

func generateStages(dimgConfig config.DimgInterface, c *Conveyor) []stage.Interface {
	var stages []stage.Interface

	dimgBaseConfig, dimgArtifact := processDimgConfig(dimgConfig)

	gitArtifacts := generateGitArtifacts(dimgBaseConfig, c)
	areGitArtifactsExist := len(gitArtifacts) != 0

	// from
	stages = appendIfExist(stages, stage.GenerateFromStage(dimgBaseConfig))

	// before_install
	stages = appendIfExist(stages, stage.GenerateBeforeInstallStage(dimgConfig, ansibleBuilderExtra(c)))

	// before_install_artifact
	stages = appendIfExist(stages, stage.GenerateArtifactImportBeforeInstallStage(dimgBaseConfig))

	if areGitArtifactsExist {
		// g_a_archive_stage
		stages = append(stages, stage.NewGAArchiveStage())
	}

	installStage := stage.GenerateInstallStage(dimgConfig, ansibleBuilderExtra(c))
	if installStage != nil {
		if areGitArtifactsExist {
			// g_a_pre_install_patch
			stages = append(stages, stage.NewGAPreInstallPatchStage())
		}

		// install
		stages = append(stages, installStage)
	}

	// after_install_artifact
	stages = appendIfExist(stages, stage.GenerateArtifactImportAfterInstallStage(dimgBaseConfig))

	beforeSetupStage := stage.GenerateBeforeSetupStage(dimgConfig, ansibleBuilderExtra(c))
	if beforeSetupStage != nil {
		if areGitArtifactsExist {
			// g_a_post_install_patch
			stages = append(stages, stage.NewGAPostInstallPatchStage())
		}

		// before_setup
		stages = append(stages, beforeSetupStage)
	}

	// before_setup_artifact
	stages = appendIfExist(stages, stage.GenerateArtifactImportBeforeSetupStage(dimgBaseConfig))

	setup := stage.GenerateSetupStage(dimgConfig, ansibleBuilderExtra(c))
	if beforeSetupStage != nil {
		if areGitArtifactsExist {
			// g_a_pre_setup_patch
			stages = append(stages, stage.NewGAPreSetupPatchStage())
		}

		// setup
		stages = append(stages, setup)
	}

	// after_setup_artifact
	stages = appendIfExist(stages, stage.GenerateArtifactImportAfterSetupStage(dimgBaseConfig))

	if dimgArtifact {
		buildArtifact := stage.GenerateBuildArtifactStage(dimgConfig, ansibleBuilderExtra(c))
		if beforeSetupStage != nil {
			if areGitArtifactsExist {
				// g_a_artifact_patch
				stages = append(stages, stage.NewGAArtifactPatchStage())
			}

			// build_artifact
			stages = append(stages, buildArtifact)
		}
	} else {
		if areGitArtifactsExist {
			// g_a_post_setup_patch
			stages = append(stages, stage.NewGAPostSetupPatchStage())

			// g_a_latest_patch
			stages = append(stages, stage.NewGALatestPatchStage())
		}

		// docker_instructions
		stages = appendIfExist(stages, stage.GenerateDockerInstructionsStage(dimgConfig.(*config.Dimg)))
	}

	if areGitArtifactsExist {
		for _, s := range stages {
			s.SetGitArtifacts(gitArtifacts)
		}
	}

	return stages
}

func generateGitArtifacts(dimgBaseConfig *config.DimgBase, _ *Conveyor) []*stage.GitArtifact {
	var gitArtifacts []*stage.GitArtifact

	var localGitRepo *git_repo.Local
	if len(dimgBaseConfig.Git.Local) != 0 {
		localGitRepo = &git_repo.Local{
			Path:   "",
			GitDir: "",
		} // TODO
	}

	for _, localGA := range dimgBaseConfig.Git.Local {
		ga := &stage.GitArtifact{
			GitRepoInterface:     localGitRepo,
			PatchesDir:           "", // TODO
			ContainerPatchesDir:  "", // TODO
			ArchivesDir:          "", // TODO
			ContainerArchivesDir: "", // TODO

			Cwd:                localGA.Add,
			To:                 localGA.To,
			ExcludePaths:       localGA.ExcludePaths,
			IncludePaths:       localGA.IncludePaths,
			Owner:              localGA.Owner,
			Group:              localGA.Group,
			StagesDependencies: stageDependenciesToMap(localGA.StageDependencies),
		}

		if empty, err := ga.IsEmpty(); err != nil {
			panic(err)
		} else if !empty {
			gitArtifacts = append(gitArtifacts, ga)
		}
	}

	remoteGitRepos := map[string]*git_repo.Remote{}
	for _, remoteGA := range dimgBaseConfig.Git.Remote {
		if len(dimgBaseConfig.Git.Remote) != 0 {
			_, exist := remoteGitRepos[remoteGA.Name]
			if !exist {
				remoteGitRepos[remoteGA.Name] = &git_repo.Remote{
					Url: remoteGA.Url,
					//	... TODO
				}
			}
		}

		ga := &stage.GitArtifact{
			GitRepoInterface:     remoteGitRepos[remoteGA.Name],
			PatchesDir:           "", // TODO
			ContainerPatchesDir:  "", // TODO
			ArchivesDir:          "", // TODO
			ContainerArchivesDir: "", // TODO

			As:                 remoteGA.As,
			Cwd:                remoteGA.Add,
			To:                 remoteGA.To,
			ExcludePaths:       remoteGA.ExcludePaths,
			IncludePaths:       remoteGA.IncludePaths,
			Owner:              remoteGA.Owner,
			StagesDependencies: stageDependenciesToMap(remoteGA.StageDependencies),
			Tag:                remoteGA.Tag,
			Commit:             remoteGA.Commit,
			Branch:             remoteGA.Branch,
		}

		if empty, err := ga.IsEmpty(); err != nil {
			panic(err)
		} else if !empty {
			gitArtifacts = append(gitArtifacts, ga)
		}
	}

	return gitArtifacts
}

func stageDependenciesToMap(sd *config.StageDependencies) map[string][]string {
	result := map[string][]string{
		"install":        sd.Install,
		"before_setup":   sd.BeforeSetup,
		"setup":          sd.Setup,
		"build_artifact": sd.BuildArtifact,
	}

	return result
}

func processDimgConfig(dimgConfig config.DimgInterface) (*config.DimgBase, bool) {
	var dimgBase *config.DimgBase
	var dimgArtifact bool
	switch dimgConfig.(type) {
	case config.Dimg:
		dimgBase = dimgConfig.(*config.Dimg).DimgBase
		dimgArtifact = false
	case config.DimgArtifact:
		dimgBase = dimgConfig.(*config.DimgArtifact).DimgBase
		dimgArtifact = true
	}

	return dimgBase, dimgArtifact
}

func ansibleBuilderExtra(c *Conveyor) *builder.Extra {
	ansibleBuilderExtra := &builder.Extra{
		TmpPath:           c.TmpDir,
		ContainerDappPath: c.ContainerDappPath,
	}

	return ansibleBuilderExtra
}

func appendIfExist(stages []stage.Interface, stage stage.Interface) []stage.Interface {
	if stage != nil {
		return append(stages, stage)
	}

	return stages
}
