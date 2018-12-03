package build

import (
	"net/url"
	"path"
	"reflect"

	"github.com/flant/dapp/pkg/build/builder"
	"github.com/flant/dapp/pkg/build/stage"
	"github.com/flant/dapp/pkg/config"
	"github.com/flant/dapp/pkg/git_repo"
	"github.com/flant/dapp/pkg/slug"
)

type InitializationPhase struct{}

func NewInitializationPhase() *InitializationPhase {
	return &InitializationPhase{}
}

func (p *InitializationPhase) Run(c *Conveyor) error {
	dimgsInOrder, err := generateDimgsInOrder(c.Dappfile, c)
	if err != nil {
		return err
	}

	c.DimgsInOrder = dimgsInOrder

	return nil
}

func generateDimgsInOrder(dappfile []*config.Dimg, c *Conveyor) ([]*Dimg, error) {
	var dimgs []*Dimg
	for _, dimgConfig := range getDimgConfigsInOrder(dappfile) {
		dimg := &Dimg{}

		from, fromDimgName := getFromAndFromDimgName(dimgConfig)
		dimg.baseImageName = from
		dimg.baseImageDimgName = fromDimgName

		stages, err := generateStages(dimgConfig, c)
		if err != nil {
			return nil, err
		}

		dimg.SetStages(stages)

		dimgs = append(dimgs, dimg)
	}

	return dimgs, nil
}

func getFromAndFromDimgName(dimgConfig config.DimgInterface) (string, string) {
	var from string
	var fromDimgName string

	dimgBaseConfig, _ := processDimgConfig(dimgConfig)

	if dimgBaseConfig.From != "" {
		from = dimgBaseConfig.From
	} else {
		fromDimg := dimgBaseConfig.FromDimg
		fromDimgArtifact := dimgBaseConfig.FromDimgArtifact

		if fromDimg != nil {
			fromDimgName = fromDimg.Name
		} else {
			fromDimgName = fromDimgArtifact.Name
		}
	}

	return from, fromDimgName
}

func getDimgConfigsInOrder(dappfile []*config.Dimg) []config.DimgInterface {
	var dimgConfigs []config.DimgInterface
	for _, dimg := range dappfile {
		dimgsInBuildOrder := dimg.DimgTree()
		for i := 0; i < len(dimgsInBuildOrder); i++ {
			if isNotInArr(dimgConfigs, dimgsInBuildOrder[i]) {
				dimgConfigs = append(dimgConfigs, dimgsInBuildOrder[i])
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

func generateStages(dimgConfig config.DimgInterface, c *Conveyor) ([]stage.Interface, error) {
	var stages []stage.Interface

	dimgBaseConfig, dimgArtifact := processDimgConfig(dimgConfig)

	baseStageOptions := &stage.NewBaseStageOptions{
		DimgTmpDir:          c.GetDimgTmpDir(dimgBaseConfig.Name),
		DimgContainerTmpDir: c.GetDimgContainerTmpDir(dimgBaseConfig.Name),
		ProjectBuildDir:     c.GetProjectBuildDir(),
	}

	gitArtifacts, err := generateGitArtifacts(dimgBaseConfig, c)
	if err != nil {
		return nil, err
	}

	areGitArtifactsExist := len(gitArtifacts) != 0

	// from
	stages = appendIfExist(stages, stage.GenerateFromStage(dimgBaseConfig, baseStageOptions))

	// before_install
	stages = appendIfExist(stages, stage.GenerateBeforeInstallStage(dimgConfig, ansibleBuilderExtra(c), baseStageOptions))

	// before_install_artifact
	stages = appendIfExist(stages, stage.GenerateArtifactImportBeforeInstallStage(dimgBaseConfig, baseStageOptions))

	if areGitArtifactsExist {
		// g_a_archive_stage
		stages = append(stages, stage.NewGAArchiveStage(baseStageOptions))
	}

	installStage := stage.GenerateInstallStage(dimgConfig, ansibleBuilderExtra(c), baseStageOptions)
	if installStage != nil {
		if areGitArtifactsExist {
			// g_a_pre_install_patch
			stages = append(stages, stage.NewGAPreInstallPatchStage(baseStageOptions))
		}

		// install
		stages = append(stages, installStage)
	}

	// after_install_artifact
	stages = appendIfExist(stages, stage.GenerateArtifactImportAfterInstallStage(dimgBaseConfig, baseStageOptions))

	beforeSetupStage := stage.GenerateBeforeSetupStage(dimgConfig, ansibleBuilderExtra(c), baseStageOptions)
	if beforeSetupStage != nil {
		if areGitArtifactsExist {
			// g_a_post_install_patch
			stages = append(stages, stage.NewGAPostInstallPatchStage(baseStageOptions))
		}

		// before_setup
		stages = append(stages, beforeSetupStage)
	}

	// before_setup_artifact
	stages = appendIfExist(stages, stage.GenerateArtifactImportBeforeSetupStage(dimgBaseConfig, baseStageOptions))

	setup := stage.GenerateSetupStage(dimgConfig, ansibleBuilderExtra(c), baseStageOptions)
	if setup != nil {
		if areGitArtifactsExist {
			// g_a_pre_setup_patch
			stages = append(stages, stage.NewGAPreSetupPatchStage(baseStageOptions))
		}

		// setup
		stages = append(stages, setup)
	}

	// after_setup_artifact
	stages = appendIfExist(stages, stage.GenerateArtifactImportAfterSetupStage(dimgBaseConfig, baseStageOptions))

	if dimgArtifact {
		buildArtifact := stage.GenerateBuildArtifactStage(dimgConfig, ansibleBuilderExtra(c), baseStageOptions)
		if buildArtifact != nil {
			if areGitArtifactsExist {
				// g_a_artifact_patch
				stages = append(stages, stage.NewGAArtifactPatchStage(baseStageOptions))
			}

			// build_artifact
			stages = append(stages, buildArtifact)
		}
	} else {
		if areGitArtifactsExist {
			// g_a_post_setup_patch
			stages = append(stages, stage.NewGAPostSetupPatchStage(baseStageOptions))

			// g_a_latest_patch
			stages = append(stages, stage.NewGALatestPatchStage(baseStageOptions))
		}

		// docker_instructions
		stages = appendIfExist(stages, stage.GenerateDockerInstructionsStage(dimgConfig.(*config.Dimg), baseStageOptions))
	}

	if areGitArtifactsExist {
		for _, s := range stages {
			s.SetGitArtifacts(gitArtifacts)
		}
	}

	return stages, nil
}

func generateGitArtifacts(dimgBaseConfig *config.DimgBase, c *Conveyor) ([]*stage.GitArtifact, error) {
	var gitArtifacts, nonEmptyGitArtifacts []*stage.GitArtifact

	var localGitRepo *git_repo.Local
	if len(dimgBaseConfig.Git.Local) != 0 {
		localGitRepo = &git_repo.Local{
			Path:   c.ProjectPath,
			GitDir: path.Join(c.ProjectPath, ".git"),
		}
	}

	for _, localGAConfig := range dimgBaseConfig.Git.Local {
		gitArtifacts = append(gitArtifacts, gitLocalArtifactInit(localGAConfig, localGitRepo, dimgBaseConfig.Name, c))
	}

	remoteGitRepos := map[string]*git_repo.Remote{}
	for _, remoteGAConfig := range dimgBaseConfig.Git.Remote {
		var remoteGitRepo *git_repo.Remote
		if len(dimgBaseConfig.Git.Remote) != 0 {
			_, exist := remoteGitRepos[remoteGAConfig.Name]
			if !exist {
				clonePath, err := getRemoteGitRepoClonePath(remoteGAConfig, c)
				if err != nil {
					return nil, err
				}

				remoteGitRepo = &git_repo.Remote{
					Url:       remoteGAConfig.Url,
					ClonePath: clonePath,
				}
				remoteGitRepos[remoteGAConfig.Name] = remoteGitRepo
			}
		}

		gitArtifacts = append(gitArtifacts, gitRemoteArtifactInit(remoteGAConfig, remoteGitRepo, dimgBaseConfig.Name, c))
	}

	for _, ga := range gitArtifacts {
		if empty, err := ga.IsEmpty(); err != nil {
			return nil, err
		} else if !empty {
			nonEmptyGitArtifacts = append(nonEmptyGitArtifacts, ga)
		}
	}

	return nonEmptyGitArtifacts, nil
}

func getRemoteGitRepoClonePath(remoteGaConfig *config.GitRemote, c *Conveyor) (string, error) {
	scheme, err := urlScheme(remoteGaConfig.Url)
	if err != nil {
		return "", err
	}

	clonePath := path.Join(
		c.GetProjectBuildDir(),
		"remote_git_repo",
		string(git_repo.RemoteGitRepoCacheVersion),
		slug.Slug(remoteGaConfig.Name),
		scheme,
	)

	return clonePath, nil
}

func urlScheme(urlString string) (string, error) {
	u, err := url.Parse(urlString)
	if err != nil {
		return "", err
	}

	return u.Scheme, nil
}

func gitRemoteArtifactInit(remoteGAConfig *config.GitRemote, remoteGitRepo *git_repo.Remote, dimgName string, c *Conveyor) *stage.GitArtifact {
	ga := baseGitArtifactInit(remoteGAConfig.GitLocalExport, dimgName, c)

	ga.Tag = remoteGAConfig.Tag
	ga.Commit = remoteGAConfig.Commit
	ga.Branch = remoteGAConfig.Branch

	ga.Name = remoteGAConfig.Name

	ga.GitRepoInterface = remoteGitRepo

	return ga
}

func gitLocalArtifactInit(localGAConfig *config.GitLocal, localGitRepo *git_repo.Local, dimgName string, c *Conveyor) *stage.GitArtifact {
	ga := baseGitArtifactInit(localGAConfig.GitLocalExport, dimgName, c)

	ga.As = localGAConfig.As

	ga.Name = "own"

	ga.GitRepoInterface = localGitRepo

	return ga
}

func baseGitArtifactInit(local *config.GitLocalExport, dimgName string, c *Conveyor) *stage.GitArtifact {
	var stageDependencies map[string][]string
	if local.StageDependencies != nil {
		stageDependencies = stageDependenciesToMap(local.StageDependencies)
	}

	ga := &stage.GitArtifact{
		PatchesDir:           getDimgPatchesDir(dimgName, c),
		ContainerPatchesDir:  getDimgPatchesContainerDir(c),
		ArchivesDir:          getDimgArchivesDir(dimgName, c),
		ContainerArchivesDir: getDimgArchivesContainerDir(c),

		RepoPath: path.Join("/", local.Add),

		Cwd:                local.Add,
		To:                 local.To,
		ExcludePaths:       local.ExcludePaths,
		IncludePaths:       local.IncludePaths,
		Owner:              local.Owner,
		Group:              local.Group,
		StagesDependencies: stageDependencies,
	}

	return ga
}

func getDimgPatchesDir(dimgName string, c *Conveyor) string {
	return path.Join(c.TmpDir, dimgName, "patch")
}

func getDimgPatchesContainerDir(c *Conveyor) string {
	return path.Join(c.ContainerDappPath, "patch")
}

func getDimgArchivesDir(dimgName string, c *Conveyor) string {
	return path.Join(c.TmpDir, dimgName, "archive")
}

func getDimgArchivesContainerDir(c *Conveyor) string {
	return path.Join(c.ContainerDappPath, "archive")
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
	case *config.Dimg:
		dimgBase = dimgConfig.(*config.Dimg).DimgBase
		dimgArtifact = false
	case *config.DimgArtifact:
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
	if !reflect.ValueOf(stage).IsNil() {
		return append(stages, stage)
	}

	return stages
}
