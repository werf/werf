package build

import (
	"fmt"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/flant/logboek"

	"github.com/flant/werf/pkg/build/stage"
	"github.com/flant/werf/pkg/config"
	"github.com/flant/werf/pkg/docker_registry"
	"github.com/flant/werf/pkg/git_repo"
	"github.com/flant/werf/pkg/logging"
	"github.com/flant/werf/pkg/slug"
	"github.com/flant/werf/pkg/werf"
)

type InitializationPhase struct{}

func NewInitializationPhase() *InitializationPhase {
	return &InitializationPhase{}
}

func (p *InitializationPhase) Run(c *Conveyor) (err error) {
	logProcessOptions := logboek.LogProcessOptions{ColorizeMsgFunc: logboek.ColorizeHighlight}
	return logboek.LogProcess("Determining of stages", logProcessOptions, func() error {
		return p.run(c)
	})
}

func (p *InitializationPhase) run(c *Conveyor) error {
	imagesInOrder, err := generateImagesInOrder(c.werfConfig.Images, c)
	if err != nil {
		return err
	}

	c.imagesInOrder = imagesInOrder

	return nil
}

func generateImagesInOrder(imageConfigs []*config.Image, c *Conveyor) ([]*Image, error) {
	var images []*Image

	imagesInterfaceConfigs := getImageConfigsInOrder(imageConfigs, c)
	for _, imageInterfaceConfig := range imagesInterfaceConfigs {
		imageName := logging.ImageLogProcessName(imageInterfaceConfig.ImageBaseConfig().Name, imageInterfaceConfig.IsArtifact())
		err := logboek.LogProcess(imageName, logboek.LogProcessOptions{ColorizeMsgFunc: ImageLogProcessColorizeFunc(imageInterfaceConfig.IsArtifact())}, func() error {
			image, err := generateImage(imageInterfaceConfig, c)
			if err != nil {
				return err
			}

			images = append(images, image)

			return nil
		})

		if err != nil {
			return nil, err
		}
	}

	return images, nil
}

func generateImage(imageInterfaceConfig config.ImageInterface, c *Conveyor) (*Image, error) {
	image := &Image{}

	imageBaseConfig := imageInterfaceConfig.ImageBaseConfig()
	imageName := imageBaseConfig.Name
	imageArtifact := imageInterfaceConfig.IsArtifact()

	from, fromImageName, fromLatest := getFromFields(imageBaseConfig)

	image.name = imageName

	if from != "" {
		image.baseImageName = from

		var baseImageRepoErr error
		baseImageRepoId, exist := c.baseImagesRepoIdsCache[from]
		if !exist {
			processMsg := fmt.Sprintf("Trying to get from base image id from registry (%s)", from)
			if err := logboek.LogProcessInline(processMsg, logboek.LogProcessInlineOptions{}, func() error {
				baseImageRepoId, baseImageRepoErr = docker_registry.ImageId(from)
				if fromLatest {
					return fmt.Errorf("can not get base image id from registry (%s): %s", from, baseImageRepoErr)
				}

				return nil
			}); err != nil {
				return nil, err
			}
		} else {
			baseImageRepoErr, _ = c.baseImagesRepoErrCache[from]
		}

		image.baseImageRepoId = baseImageRepoId
		image.baseImageRepoErr = baseImageRepoErr

		c.baseImagesRepoIdsCache[from] = baseImageRepoId
		c.baseImagesRepoErrCache[from] = baseImageRepoErr
	}

	image.baseImageLatest = fromLatest

	image.baseImageImageName = fromImageName
	image.isArtifact = imageArtifact

	err := initStages(image, imageInterfaceConfig, c)
	if err != nil {
		return nil, err
	}

	return image, nil
}

func getFromFields(imageBaseConfig *config.ImageBase) (string, string, bool) {
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

	return from, fromImageName, imageBaseConfig.FromLatest
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
				logboek.LogErrorF("WARNING: Specified image '%s' isn't defined in werf.yaml!\n", imageName)
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

func initStages(image *Image, imageInterfaceConfig config.ImageInterface, c *Conveyor) error {
	var stages []stage.Interface

	imageBaseConfig := imageInterfaceConfig.ImageBaseConfig()
	imageName := imageBaseConfig.Name
	imageArtifact := imageInterfaceConfig.IsArtifact()

	baseStageOptions := &stage.NewBaseStageOptions{
		ImageName:        imageName,
		ConfigMounts:     imageBaseConfig.Mount,
		ImageTmpDir:      c.GetImageTmpDir(imageBaseConfig.Name),
		ContainerWerfDir: c.containerWerfDir,
		ProjectName:      c.werfConfig.Meta.Project,
	}

	gitArchiveStageOptions := &stage.NewGitArchiveStageOptions{
		ArchivesDir:          getImageArchivesDir(imageName, c),
		ContainerArchivesDir: getImageArchivesContainerDir(c),
	}

	gitPatchStageOptions := &stage.NewGitPatchStageOptions{
		PatchesDir:           getImagePatchesDir(imageName, c),
		ArchivesDir:          getImageArchivesDir(imageName, c),
		ContainerPatchesDir:  getImagePatchesContainerDir(c),
		ContainerArchivesDir: getImageArchivesContainerDir(c),
	}

	gitMappings, err := generateGitMappings(imageBaseConfig, c)
	if err != nil {
		return err
	}

	gitMappingsExist := len(gitMappings) != 0

	stages = appendIfExist(stages, stage.GenerateFromStage(imageBaseConfig, image.baseImageRepoId, baseStageOptions))
	stages = appendIfExist(stages, stage.GenerateBeforeInstallStage(imageBaseConfig, baseStageOptions))
	stages = appendIfExist(stages, stage.GenerateImportsBeforeInstallStage(imageBaseConfig, baseStageOptions))

	if gitMappingsExist {
		stages = append(stages, stage.NewGitArchiveStage(gitArchiveStageOptions, baseStageOptions))
	}

	stages = appendIfExist(stages, stage.GenerateInstallStage(imageBaseConfig, gitPatchStageOptions, baseStageOptions))
	stages = appendIfExist(stages, stage.GenerateImportsAfterInstallStage(imageBaseConfig, baseStageOptions))
	stages = appendIfExist(stages, stage.GenerateBeforeSetupStage(imageBaseConfig, gitPatchStageOptions, baseStageOptions))
	stages = appendIfExist(stages, stage.GenerateImportsBeforeSetupStage(imageBaseConfig, baseStageOptions))
	stages = appendIfExist(stages, stage.GenerateSetupStage(imageBaseConfig, gitPatchStageOptions, baseStageOptions))
	stages = appendIfExist(stages, stage.GenerateImportsAfterSetupStage(imageBaseConfig, baseStageOptions))

	if !imageArtifact {
		if gitMappingsExist {
			stages = append(stages, stage.NewGitCacheStage(gitPatchStageOptions, baseStageOptions))
			stages = append(stages, stage.NewGitLatestPatchStage(gitPatchStageOptions, baseStageOptions))
		}

		stages = appendIfExist(stages, stage.GenerateDockerInstructionsStage(imageInterfaceConfig.(*config.Image), baseStageOptions))
	}

	if len(gitMappings) != 0 {
		logboek.LogInfoLn("Using git stages")

		for _, s := range stages {
			s.SetGitMappings(gitMappings)
		}
	}

	image.SetStages(stages)

	return nil
}

func generateGitMappings(imageBaseConfig *config.ImageBase, c *Conveyor) ([]*stage.GitMapping, error) {
	var gitMappings []*stage.GitMapping

	var localGitRepo *git_repo.Local
	if len(imageBaseConfig.Git.Local) != 0 {
		localGitRepo = &git_repo.Local{
			Base:   git_repo.Base{Name: "own"},
			Path:   c.projectDir,
			GitDir: path.Join(c.projectDir, ".git"),
		}
	}

	for _, localGitMappingConfig := range imageBaseConfig.Git.Local {
		gitMappings = append(gitMappings, gitLocalPathInit(localGitMappingConfig, localGitRepo, imageBaseConfig.Name, c))
	}

	for _, remoteGitMappingConfig := range imageBaseConfig.Git.Remote {
		remoteGitRepo, exist := c.remoteGitRepos[remoteGitMappingConfig.Name]
		if !exist {
			clonePath, err := getRemoteGitRepoClonePath(remoteGitMappingConfig, c)
			if err != nil {
				return nil, err
			}

			if err := os.MkdirAll(filepath.Dir(clonePath), os.ModePerm); err != nil {
				return nil, fmt.Errorf("unable to mkdir %s: %s", filepath.Dir(clonePath), err)
			}

			remoteGitRepo = &git_repo.Remote{
				Base:      git_repo.Base{Name: remoteGitMappingConfig.Name},
				Url:       remoteGitMappingConfig.Url,
				ClonePath: clonePath,
			}

			if err := logboek.LogProcess(fmt.Sprintf("Refreshing %s repository", remoteGitMappingConfig.Name), logboek.LogProcessOptions{}, func() error {
				return remoteGitRepo.CloneAndFetch()
			}); err != nil {
				return nil, err
			}

			c.remoteGitRepos[remoteGitMappingConfig.Name] = remoteGitRepo
		}

		gitMappings = append(gitMappings, gitRemoteArtifactInit(remoteGitMappingConfig, remoteGitRepo, imageBaseConfig.Name, c))
	}

	var res []*stage.GitMapping

	if len(gitMappings) != 0 {
		err := logboek.LogProcess(fmt.Sprintf("Initializing git mappings"), logboek.LogProcessOptions{}, func() error {
			nonEmptyGitMappings, err := getNonEmptyGitMappings(gitMappings)
			if err != nil {
				return err
			}

			res = nonEmptyGitMappings

			return nil
		})

		if err != nil {
			return nil, err
		}
	}

	return res, nil
}

func getNonEmptyGitMappings(gitMappings []*stage.GitMapping) ([]*stage.GitMapping, error) {
	var nonEmptyGitMappings []*stage.GitMapping

	for ind, gitMapping := range gitMappings {
		if err := logboek.LogProcess(fmt.Sprintf("[%d] git mapping from %s repository", ind, gitMapping.Name), logboek.LogProcessOptions{}, func() error {
			withTripleIndent := func(f func()) {
				logboek.IndentUp()
				logboek.IndentUp()
				logboek.IndentUp()
				f()
				logboek.IndentDown()
				logboek.IndentDown()
				logboek.IndentDown()
			}

			withTripleIndent(func() {
				logboek.LogInfoF("add: %s\n", gitMapping.Cwd)
				logboek.LogInfoF("to: %s\n", gitMapping.To)

				if len(gitMapping.IncludePaths) != 0 {
					logboek.LogInfoF("includePaths: %+v\n", gitMapping.IncludePaths)
				}

				if len(gitMapping.ExcludePaths) != 0 {
					logboek.LogInfoF("excludePaths: %+v\n", gitMapping.ExcludePaths)
				}

				if gitMapping.Commit != "" {
					logboek.LogInfoF("commit: %s\n", gitMapping.Commit)
				}

				if gitMapping.Branch != "" {
					logboek.LogInfoF("branch: %s\n", gitMapping.Branch)
				}

				if gitMapping.Owner != "" {
					logboek.LogInfoF("owner: %s\n", gitMapping.Owner)
				}

				if gitMapping.Group != "" {
					logboek.LogInfoF("group: %s\n", gitMapping.Group)
				}

				if len(gitMapping.StagesDependencies) != 0 {
					logboek.LogInfoLn("stageDependencies:")
					for s, values := range gitMapping.StagesDependencies {
						if len(values) != 0 {
							logboek.LogInfoF("  %s: %v\n", s, values)
						}
					}

				}
			})

			logboek.LogLn()

			commit, err := gitMapping.LatestCommit()
			if err != nil {
				return fmt.Errorf("unable to get commit of repo '%s': %s", gitMapping.GitRepo().GetName(), err)
			}

			cwd := gitMapping.Cwd
			if cwd == "" {
				cwd = "/"
			}

			if empty, err := gitMapping.IsEmpty(); err != nil {
				return err
			} else if !empty {
				logboek.LogInfoF("Commit %s will be used\n", commit)
				nonEmptyGitMappings = append(nonEmptyGitMappings, gitMapping)
			} else {
				logboek.LogErrorF("WARNING: Empty git mapping will be ignored (commit %s)\n", commit)
			}

			return nil
		}); err != nil {
			return nil, err
		}
	}

	return nonEmptyGitMappings, nil
}

func getRemoteGitRepoClonePath(remoteGitMappingConfig *config.GitRemote, c *Conveyor) (string, error) {
	scheme, err := urlScheme(remoteGitMappingConfig.Url)
	if err != nil {
		return "", err
	}

	clonePath := path.Join(
		werf.GetLocalCacheDir(),
		"remote_git_repos",
		"projects",
		c.werfConfig.Meta.Project,
		fmt.Sprintf("%v", git_repo.RemoteGitRepoCacheVersion),
		slug.Slug(remoteGitMappingConfig.Name),
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

func gitRemoteArtifactInit(remoteGitMappingConfig *config.GitRemote, remoteGitRepo *git_repo.Remote, imageName string, c *Conveyor) *stage.GitMapping {
	gitMapping := baseGitMappingInit(remoteGitMappingConfig.GitLocalExport, imageName, c)

	gitMapping.Tag = remoteGitMappingConfig.Tag
	gitMapping.Commit = remoteGitMappingConfig.Commit
	gitMapping.Branch = remoteGitMappingConfig.Branch

	gitMapping.Name = remoteGitMappingConfig.Name

	gitMapping.GitRepoInterface = remoteGitRepo

	return gitMapping
}

func gitLocalPathInit(localGitMappingConfig *config.GitLocal, localGitRepo *git_repo.Local, imageName string, c *Conveyor) *stage.GitMapping {
	gitMapping := baseGitMappingInit(localGitMappingConfig.GitLocalExport, imageName, c)

	gitMapping.As = localGitMappingConfig.As

	gitMapping.Name = "own"

	gitMapping.GitRepoInterface = localGitRepo

	return gitMapping
}

func baseGitMappingInit(local *config.GitLocalExport, imageName string, c *Conveyor) *stage.GitMapping {
	var stageDependencies map[stage.StageName][]string
	if local.StageDependencies != nil {
		stageDependencies = stageDependenciesToMap(local.StageDependencies)
	}

	gitMapping := &stage.GitMapping{
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

	return gitMapping
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

func appendIfExist(stages []stage.Interface, stage stage.Interface) []stage.Interface {
	if !reflect.ValueOf(stage).IsNil() {
		logboek.LogInfoF("Using stage %s\n", stage.Name())
		return append(stages, stage)
	}

	return stages
}
