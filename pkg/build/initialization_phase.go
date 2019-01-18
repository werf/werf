package build

import (
	"fmt"
	"net/url"
	"path"
	"reflect"
	"strings"

	"github.com/flant/werf/pkg/build/stage"
	"github.com/flant/werf/pkg/config"
	"github.com/flant/werf/pkg/git_repo"
	"github.com/flant/werf/pkg/logger"
	"github.com/flant/werf/pkg/slug"
)

type InitializationPhase struct{}

func NewInitializationPhase() *InitializationPhase {
	return &InitializationPhase{}
}

func (p *InitializationPhase) Run(c *Conveyor) error {
	dimgsInOrder, err := generateDimgsInOrder(c.werfConfig.Dimgs, c)
	if err != nil {
		return err
	}

	c.dimgsInOrder = dimgsInOrder

	return nil
}

func generateDimgsInOrder(dimgConfigs []*config.Dimg, c *Conveyor) ([]*Dimg, error) {
	var dimgs []*Dimg
	for _, dimgConfig := range getDimgConfigsInOrder(dimgConfigs, c) {
		dimg := &Dimg{}

		dimgBaseConfig, dimgName, dimgArtifact := processDimgConfig(dimgConfig)
		from, fromDimgName := getFromAndFromDimgName(dimgBaseConfig)

		dimg.name = dimgName
		dimg.baseImageName = from
		dimg.baseImageDimgName = fromDimgName
		dimg.isArtifact = dimgArtifact

		stages, err := generateStages(dimgConfig, c)
		if err != nil {
			return nil, err
		}

		dimg.SetStages(stages)

		dimgs = append(dimgs, dimg)
	}

	return dimgs, nil
}

func getFromAndFromDimgName(dimgBaseConfig *config.DimgBase) (string, string) {
	var from string
	var fromDimgName string

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

func getDimgConfigsInOrder(dimgConfigs []*config.Dimg, c *Conveyor) []config.DimgInterface {
	var dimgs []config.DimgInterface
	for _, dimg := range getDimgConfigToProcess(dimgConfigs, c) {
		dimgsInBuildOrder := dimg.DimgTree()
		for i := 0; i < len(dimgsInBuildOrder); i++ {
			if isNotInArr(dimgs, dimgsInBuildOrder[i]) {
				dimgs = append(dimgs, dimgsInBuildOrder[i])
			}
		}
	}

	return dimgs
}

func getDimgConfigToProcess(dimgConfigs []*config.Dimg, c *Conveyor) []*config.Dimg {
	var dimgConfigsToProcess []*config.Dimg

	if len(c.dimgNamesToProcess) == 0 {
		dimgConfigsToProcess = dimgConfigs
	} else {
		for _, dimgName := range c.dimgNamesToProcess {
			dimgToProcess := getDimgConfigByName(dimgConfigs, dimgName)
			if dimgToProcess == nil {
				logger.LogWarningF("WARNING: Specified dimg '%s' isn't defined in werf.yaml!\n", dimgName)
			} else {
				dimgConfigsToProcess = append(dimgConfigsToProcess, dimgToProcess)
			}
		}
	}

	return dimgConfigsToProcess
}

func getDimgConfigByName(dimgConfigs []*config.Dimg, name string) *config.Dimg {
	for _, dimg := range dimgConfigs {
		if dimg.Name == name {
			return dimg
		}
	}

	return nil
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

	dimgBaseConfig, dimgName, dimgArtifact := processDimgConfig(dimgConfig)

	baseStageOptions := &stage.NewBaseStageOptions{
		DimgName:         dimgName,
		ConfigMounts:     dimgBaseConfig.Mount,
		DimgTmpDir:       c.GetDimgTmpDir(dimgBaseConfig.Name),
		ContainerWerfDir: c.containerWerfDir,
		ProjectBuildDir:  c.projectBuildDir,
	}

	gitArchiveStageOptions := &stage.NewGitArchiveStageOptions{
		ArchivesDir:          getDimgArchivesDir(dimgName, c),
		ContainerArchivesDir: getDimgArchivesContainerDir(c),
	}

	gitPatchStageOptions := &stage.NewGitPatchStageOptions{
		PatchesDir:          getDimgPatchesDir(dimgName, c),
		ContainerPatchesDir: getDimgPatchesContainerDir(c),
	}

	gitPaths, err := generateGitPaths(dimgBaseConfig, c)
	if err != nil {
		return nil, err
	}

	for _, gitPath := range gitPaths {
		commit, err := gitPath.LatestCommit()
		if err != nil {
			return nil, fmt.Errorf("unable to get commit of repo '%s': %s", gitPath.GitRepo().String(), err)
		}

		fmt.Printf("Using commit '%s' of repo '%s'\n", commit, gitPath.GitRepo().String())
	}

	// from
	stages = appendIfExist(stages, stage.GenerateFromStage(dimgBaseConfig, baseStageOptions))

	// before_install
	stages = appendIfExist(stages, stage.GenerateBeforeInstallStage(dimgBaseConfig, baseStageOptions))

	// before_install_artifact
	stages = appendIfExist(stages, stage.GenerateArtifactImportBeforeInstallStage(dimgBaseConfig, baseStageOptions))

	// git_archive_stage
	stages = append(stages, stage.NewGitArchiveStage(gitArchiveStageOptions, baseStageOptions))

	// install
	stages = appendIfExist(stages, stage.GenerateInstallStage(dimgBaseConfig, gitPatchStageOptions, baseStageOptions))

	// after_install_artifact
	stages = appendIfExist(stages, stage.GenerateArtifactImportAfterInstallStage(dimgBaseConfig, baseStageOptions))

	// before_setup
	stages = appendIfExist(stages, stage.GenerateBeforeSetupStage(dimgBaseConfig, gitPatchStageOptions, baseStageOptions))

	// before_setup_artifact
	stages = appendIfExist(stages, stage.GenerateArtifactImportBeforeSetupStage(dimgBaseConfig, baseStageOptions))

	// setup
	stages = appendIfExist(stages, stage.GenerateSetupStage(dimgBaseConfig, gitPatchStageOptions, baseStageOptions))

	// after_setup_artifact
	stages = appendIfExist(stages, stage.GenerateArtifactImportAfterSetupStage(dimgBaseConfig, baseStageOptions))

	if !dimgArtifact {
		// git_post_setup_patch
		stages = append(stages, stage.NewGitCacheStage(gitPatchStageOptions, baseStageOptions))

		// git_latest_patch
		stages = append(stages, stage.NewGitLatestPatchStage(gitPatchStageOptions, baseStageOptions))

		// docker_instructions
		stages = appendIfExist(stages, stage.GenerateDockerInstructionsStage(dimgConfig.(*config.Dimg), baseStageOptions))
	}

	for _, s := range stages {
		s.SetGitPaths(gitPaths)
	}

	return stages, nil
}

func generateGitPaths(dimgBaseConfig *config.DimgBase, c *Conveyor) ([]*stage.GitPath, error) {
	var gitPaths, nonEmptyGitPaths []*stage.GitPath

	var localGitRepo *git_repo.Local
	if len(dimgBaseConfig.Git.Local) != 0 {
		localGitRepo = &git_repo.Local{
			Base:   git_repo.Base{Name: "own"},
			Path:   c.projectDir,
			GitDir: path.Join(c.projectDir, ".git"),
		}
	}

	for _, localGitPathConfig := range dimgBaseConfig.Git.Local {
		gitPaths = append(gitPaths, gitLocalPathInit(localGitPathConfig, localGitRepo, dimgBaseConfig.Name, c))
	}

	for _, remoteGitPathConfig := range dimgBaseConfig.Git.Remote {
		remoteGitRepo, exist := c.remoteGitRepos[remoteGitPathConfig.Name]
		if !exist {
			clonePath, err := getRemoteGitRepoClonePath(remoteGitPathConfig, c)
			if err != nil {
				return nil, err
			}

			remoteGitRepo = &git_repo.Remote{
				Base:      git_repo.Base{Name: remoteGitPathConfig.Name},
				Url:       remoteGitPathConfig.Url,
				ClonePath: clonePath,
			}

			if err := remoteGitRepo.CloneAndFetch(); err != nil {
				return nil, err
			}

			c.remoteGitRepos[remoteGitPathConfig.Name] = remoteGitRepo
		}

		gitPaths = append(gitPaths, gitRemoteArtifactInit(remoteGitPathConfig, remoteGitRepo, dimgBaseConfig.Name, c))
	}

	for _, gitPath := range gitPaths {
		if empty, err := gitPath.IsEmpty(); err != nil {
			return nil, err
		} else if !empty {
			nonEmptyGitPaths = append(nonEmptyGitPaths, gitPath)
		}
	}

	return nonEmptyGitPaths, nil
}

func getRemoteGitRepoClonePath(remoteGitPathConfig *config.GitRemote, c *Conveyor) (string, error) {
	scheme, err := urlScheme(remoteGitPathConfig.Url)
	if err != nil {
		return "", err
	}

	clonePath := path.Join(
		c.projectBuildDir,
		"remote_git_repo",
		fmt.Sprintf("%v", git_repo.RemoteGitRepoCacheVersion),
		slug.Slug(remoteGitPathConfig.Name),
		scheme,
	)

	return clonePath, nil
}

func urlScheme(urlString string) (string, error) {
	u, err := url.Parse(urlString)
	if err != nil {
		if strings.HasSuffix(err.Error(), "first path segment in URL cannot contain colon") {
			for _, protocol := range []string{"git", "ssh"} {
				if strings.HasPrefix(urlString, fmt.Sprintf("%s@", protocol)) {
					return "ssh", nil
				}
			}
		}
		return "", err
	}

	return u.Scheme, nil
}

func gitRemoteArtifactInit(remoteGitPathConfig *config.GitRemote, remoteGitRepo *git_repo.Remote, dimgName string, c *Conveyor) *stage.GitPath {
	gitPath := baseGitPathInit(remoteGitPathConfig.GitLocalExport, dimgName, c)

	gitPath.Tag = remoteGitPathConfig.Tag
	gitPath.Commit = remoteGitPathConfig.Commit
	gitPath.Branch = remoteGitPathConfig.Branch

	gitPath.Name = remoteGitPathConfig.Name

	gitPath.GitRepoInterface = remoteGitRepo

	return gitPath
}

func gitLocalPathInit(localGitPathConfig *config.GitLocal, localGitRepo *git_repo.Local, dimgName string, c *Conveyor) *stage.GitPath {
	gitPath := baseGitPathInit(localGitPathConfig.GitLocalExport, dimgName, c)

	gitPath.As = localGitPathConfig.As

	gitPath.Name = "own"

	gitPath.GitRepoInterface = localGitRepo

	return gitPath
}

func baseGitPathInit(local *config.GitLocalExport, dimgName string, c *Conveyor) *stage.GitPath {
	var stageDependencies map[stage.StageName][]string
	if local.StageDependencies != nil {
		stageDependencies = stageDependenciesToMap(local.StageDependencies)
	}

	gitPath := &stage.GitPath{
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

	return gitPath
}

func getDimgPatchesDir(dimgName string, c *Conveyor) string {
	return path.Join(c.tmpDir, dimgName, "patch")
}

func getDimgPatchesContainerDir(c *Conveyor) string {
	return path.Join(c.containerWerfDir, "patch")
}

func getDimgArchivesDir(dimgName string, c *Conveyor) string {
	return path.Join(c.tmpDir, dimgName, "archive")
}

func getDimgArchivesContainerDir(c *Conveyor) string {
	return path.Join(c.containerWerfDir, "archive")
}

func stageDependenciesToMap(sd *config.StageDependencies) map[stage.StageName][]string {
	result := map[stage.StageName][]string{
		stage.Install:     sd.Install,
		stage.BeforeSetup: sd.BeforeSetup,
		stage.Setup:       sd.Setup,
	}

	return result
}

func processDimgConfig(dimgConfig config.DimgInterface) (*config.DimgBase, string, bool) {
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

	return dimgBase, dimgBase.Name, dimgArtifact
}

func appendIfExist(stages []stage.Interface, stage stage.Interface) []stage.Interface {
	if !reflect.ValueOf(stage).IsNil() {
		return append(stages, stage)
	}

	return stages
}
