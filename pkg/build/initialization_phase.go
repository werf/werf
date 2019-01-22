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
	imagesInOrder, err := generateImagesInOrder(c.werfConfig.Images, c)
	if err != nil {
		return err
	}

	c.imagesInOrder = imagesInOrder

	return nil
}

func generateImagesInOrder(imageConfigs []*config.Image, c *Conveyor) ([]*Image, error) {
	var images []*Image
	for _, imageConfig := range getImageConfigsInOrder(imageConfigs, c) {
		image := &Image{}

		imageBaseConfig, imageName, imageArtifact := processImageConfig(imageConfig)
		from, fromImageName := getFromAndFromImageName(imageBaseConfig)

		image.name = imageName
		image.baseImageName = from
		image.baseImageImageName = fromImageName
		image.isArtifact = imageArtifact

		stages, err := generateStages(imageConfig, c)
		if err != nil {
			return nil, err
		}

		image.SetStages(stages)

		images = append(images, image)
	}

	return images, nil
}

func getFromAndFromImageName(imageBaseConfig *config.ImageBase) (string, string) {
	var from string
	var fromImageName string

	if imageBaseConfig.From != "" {
		from = imageBaseConfig.From
	} else {
		fromImage := imageBaseConfig.FromImage
		fromImageArtifact := imageBaseConfig.FromImageArtifact

		if fromImage != nil {
			fromImageName = fromImage.Name
		} else {
			fromImageName = fromImageArtifact.Name
		}
	}

	return from, fromImageName
}

func getImageConfigsInOrder(imageConfigs []*config.Image, c *Conveyor) []config.ImageInterface {
	var images []config.ImageInterface
	for _, image := range getImageConfigToProcess(imageConfigs, c) {
		imagesInBuildOrder := image.ImageTree()
		for i := 0; i < len(imagesInBuildOrder); i++ {
			if isNotInArr(images, imagesInBuildOrder[i]) {
				images = append(images, imagesInBuildOrder[i])
			}
		}
	}

	return images
}

func getImageConfigToProcess(imageConfigs []*config.Image, c *Conveyor) []*config.Image {
	var imageConfigsToProcess []*config.Image

	if len(c.imageNamesToProcess) == 0 {
		imageConfigsToProcess = imageConfigs
	} else {
		for _, imageName := range c.imageNamesToProcess {
			imageToProcess := getImageConfigByName(imageConfigs, imageName)
			if imageToProcess == nil {
				logger.LogWarningF("WARNING: Specified image '%s' isn't defined in werf.yaml!\n", imageName)
			} else {
				imageConfigsToProcess = append(imageConfigsToProcess, imageToProcess)
			}
		}
	}

	return imageConfigsToProcess
}

func getImageConfigByName(imageConfigs []*config.Image, name string) *config.Image {
	for _, image := range imageConfigs {
		if image.Name == name {
			return image
		}
	}

	return nil
}

func isNotInArr(arr []config.ImageInterface, obj config.ImageInterface) bool {
	for _, elm := range arr {
		if reflect.DeepEqual(elm, obj) {
			return false
		}
	}

	return true
}

func generateStages(imageConfig config.ImageInterface, c *Conveyor) ([]stage.Interface, error) {
	var stages []stage.Interface

	imageBaseConfig, imageName, imageArtifact := processImageConfig(imageConfig)

	baseStageOptions := &stage.NewBaseStageOptions{
		ImageName:        imageName,
		ConfigMounts:     imageBaseConfig.Mount,
		ImageTmpDir:      c.GetImageTmpDir(imageBaseConfig.Name),
		ContainerWerfDir: c.containerWerfDir,
		ProjectBuildDir:  c.projectBuildDir,
	}

	gitArchiveStageOptions := &stage.NewGitArchiveStageOptions{
		ArchivesDir:          getImageArchivesDir(imageName, c),
		ContainerArchivesDir: getImageArchivesContainerDir(c),
	}

	gitPatchStageOptions := &stage.NewGitPatchStageOptions{
		PatchesDir:          getImagePatchesDir(imageName, c),
		ContainerPatchesDir: getImagePatchesContainerDir(c),
	}

	gitPaths, err := generateGitPaths(imageBaseConfig, c)
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
	stages = appendIfExist(stages, stage.GenerateFromStage(imageBaseConfig, baseStageOptions))

	// before_install
	stages = appendIfExist(stages, stage.GenerateBeforeInstallStage(imageBaseConfig, baseStageOptions))

	// before_install_artifact
	stages = appendIfExist(stages, stage.GenerateArtifactImportBeforeInstallStage(imageBaseConfig, baseStageOptions))

	// git_archive_stage
	stages = append(stages, stage.NewGitArchiveStage(gitArchiveStageOptions, baseStageOptions))

	// install
	stages = appendIfExist(stages, stage.GenerateInstallStage(imageBaseConfig, gitPatchStageOptions, baseStageOptions))

	// after_install_artifact
	stages = appendIfExist(stages, stage.GenerateArtifactImportAfterInstallStage(imageBaseConfig, baseStageOptions))

	// before_setup
	stages = appendIfExist(stages, stage.GenerateBeforeSetupStage(imageBaseConfig, gitPatchStageOptions, baseStageOptions))

	// before_setup_artifact
	stages = appendIfExist(stages, stage.GenerateArtifactImportBeforeSetupStage(imageBaseConfig, baseStageOptions))

	// setup
	stages = appendIfExist(stages, stage.GenerateSetupStage(imageBaseConfig, gitPatchStageOptions, baseStageOptions))

	// after_setup_artifact
	stages = appendIfExist(stages, stage.GenerateArtifactImportAfterSetupStage(imageBaseConfig, baseStageOptions))

	if !imageArtifact {
		// git_post_setup_patch
		stages = append(stages, stage.NewGitCacheStage(gitPatchStageOptions, baseStageOptions))

		// git_latest_patch
		stages = append(stages, stage.NewGitLatestPatchStage(gitPatchStageOptions, baseStageOptions))

		// docker_instructions
		stages = appendIfExist(stages, stage.GenerateDockerInstructionsStage(imageConfig.(*config.Image), baseStageOptions))
	}

	for _, s := range stages {
		s.SetGitPaths(gitPaths)
	}

	return stages, nil
}

func generateGitPaths(imageBaseConfig *config.ImageBase, c *Conveyor) ([]*stage.GitPath, error) {
	var gitPaths, nonEmptyGitPaths []*stage.GitPath

	var localGitRepo *git_repo.Local
	if len(imageBaseConfig.Git.Local) != 0 {
		localGitRepo = &git_repo.Local{
			Base:   git_repo.Base{Name: "own"},
			Path:   c.projectDir,
			GitDir: path.Join(c.projectDir, ".git"),
		}
	}

	for _, localGitPathConfig := range imageBaseConfig.Git.Local {
		gitPaths = append(gitPaths, gitLocalPathInit(localGitPathConfig, localGitRepo, imageBaseConfig.Name, c))
	}

	for _, remoteGitPathConfig := range imageBaseConfig.Git.Remote {
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

		gitPaths = append(gitPaths, gitRemoteArtifactInit(remoteGitPathConfig, remoteGitRepo, imageBaseConfig.Name, c))
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

func gitRemoteArtifactInit(remoteGitPathConfig *config.GitRemote, remoteGitRepo *git_repo.Remote, imageName string, c *Conveyor) *stage.GitPath {
	gitPath := baseGitPathInit(remoteGitPathConfig.GitLocalExport, imageName, c)

	gitPath.Tag = remoteGitPathConfig.Tag
	gitPath.Commit = remoteGitPathConfig.Commit
	gitPath.Branch = remoteGitPathConfig.Branch

	gitPath.Name = remoteGitPathConfig.Name

	gitPath.GitRepoInterface = remoteGitRepo

	return gitPath
}

func gitLocalPathInit(localGitPathConfig *config.GitLocal, localGitRepo *git_repo.Local, imageName string, c *Conveyor) *stage.GitPath {
	gitPath := baseGitPathInit(localGitPathConfig.GitLocalExport, imageName, c)

	gitPath.As = localGitPathConfig.As

	gitPath.Name = "own"

	gitPath.GitRepoInterface = localGitRepo

	return gitPath
}

func baseGitPathInit(local *config.GitLocalExport, imageName string, c *Conveyor) *stage.GitPath {
	var stageDependencies map[stage.StageName][]string
	if local.StageDependencies != nil {
		stageDependencies = stageDependenciesToMap(local.StageDependencies)
	}

	gitPath := &stage.GitPath{
		PatchesDir:           getImagePatchesDir(imageName, c),
		ContainerPatchesDir:  getImagePatchesContainerDir(c),
		ArchivesDir:          getImageArchivesDir(imageName, c),
		ContainerArchivesDir: getImageArchivesContainerDir(c),

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

func getImagePatchesDir(imageName string, c *Conveyor) string {
	return path.Join(c.tmpDir, imageName, "patch")
}

func getImagePatchesContainerDir(c *Conveyor) string {
	return path.Join(c.containerWerfDir, "patch")
}

func getImageArchivesDir(imageName string, c *Conveyor) string {
	return path.Join(c.tmpDir, imageName, "archive")
}

func getImageArchivesContainerDir(c *Conveyor) string {
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

func processImageConfig(imageConfig config.ImageInterface) (*config.ImageBase, string, bool) {
	var imageBase *config.ImageBase
	var imageArtifact bool
	switch imageConfig.(type) {
	case *config.Image:
		imageBase = imageConfig.(*config.Image).ImageBase
		imageArtifact = false
	case *config.ImageArtifact:
		imageBase = imageConfig.(*config.ImageArtifact).ImageBase
		imageArtifact = true
	}

	return imageBase, imageBase.Name, imageArtifact
}

func appendIfExist(stages []stage.Interface, stage stage.Interface) []stage.Interface {
	if !reflect.ValueOf(stage).IsNil() {
		return append(stages, stage)
	}

	return stages
}
