package build

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"path"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"

	"github.com/docker/cli/cli/command/image/build"
	"github.com/docker/docker/pkg/fileutils"
	"github.com/moby/buildkit/frontend/dockerfile/instructions"
	"github.com/moby/buildkit/frontend/dockerfile/parser"
	"github.com/moby/buildkit/frontend/dockerfile/shell"

	"github.com/flant/logboek"

	"github.com/flant/werf/pkg/build/stage"
	"github.com/flant/werf/pkg/config"
	"github.com/flant/werf/pkg/git_repo"
	"github.com/flant/werf/pkg/logging"
	"github.com/flant/werf/pkg/util"
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
	imagesInterfaces := getImageConfigsInOrder(c)
	for _, imageInterfaceConfig := range imagesInterfaces {
		var image *Image
		var imageLogName string
		var colorizeMsgFunc func(...interface{}) string

		switch imageConfig := imageInterfaceConfig.(type) {
		case config.StapelImageInterface:
			imageLogName = logging.ImageLogProcessName(imageConfig.ImageBaseConfig().Name, imageConfig.IsArtifact())
			colorizeMsgFunc = ImageLogProcessColorizeFunc(imageConfig.IsArtifact())
		case *config.ImageFromDockerfile:
			imageLogName = logging.ImageLogProcessName(imageConfig.Name, false)
			colorizeMsgFunc = ImageLogProcessColorizeFunc(false)
		}

		err := logboek.LogProcess(imageLogName, logboek.LogProcessOptions{ColorizeMsgFunc: colorizeMsgFunc}, func() error {
			var err error

			switch imageConfig := imageInterfaceConfig.(type) {
			case config.StapelImageInterface:
				image, err = prepareImageBasedOnStapelImageConfig(imageConfig, c)
			case *config.ImageFromDockerfile:
				image, err = prepareImageBasedOnImageFromDockerfile(imageConfig, c)
			}

			if err != nil {
				return err
			}

			c.imagesInOrder = append(c.imagesInOrder, image)

			return nil
		})

		if err != nil {
			return err
		}
	}

	return nil
}

func prepareImageBasedOnStapelImageConfig(imageInterfaceConfig config.StapelImageInterface, c *Conveyor) (*Image, error) {
	image := &Image{}

	imageBaseConfig := imageInterfaceConfig.ImageBaseConfig()
	imageName := imageBaseConfig.Name
	imageArtifact := imageInterfaceConfig.IsArtifact()

	from, fromImageName, fromLatest := getFromFields(imageBaseConfig)

	image.name = imageName

	if from != "" {
		if err := handleImageFromName(from, fromLatest, image, c); err != nil {
			return nil, err
		}
	} else {
		image.baseImageImageName = fromImageName
	}

	image.isArtifact = imageArtifact

	err := initStages(image, imageInterfaceConfig, c)
	if err != nil {
		return nil, err
	}

	return image, nil
}

func handleImageFromName(from string, fromLatest bool, image *Image, c *Conveyor) error {
	image.baseImageName = from

	if fromLatest {
		if _, err := image.getFromBaseImageIdFromRegistry(c, image.baseImageName); err != nil {
			return err
		}
	}

	return nil
}

func getFromFields(imageBaseConfig *config.StapelImageBase) (string, string, bool) {
	var from string
	var fromImageName string

	if imageBaseConfig.From != "" {
		from = imageBaseConfig.From
	} else if imageBaseConfig.FromImageName != "" {
		fromImageName = imageBaseConfig.FromImageName
	} else if imageBaseConfig.FromImageArtifactName != "" {
		fromImageName = imageBaseConfig.FromImageArtifactName
	}

	return from, fromImageName, imageBaseConfig.FromLatest
}

func getImageConfigsInOrder(c *Conveyor) []config.ImageInterface {
	var images []config.ImageInterface
	for _, imageInterf := range getImageConfigsToProcess(c) {
		var imagesInBuildOrder []config.ImageInterface

		switch image := imageInterf.(type) {
		case *config.StapelImage:
			imagesInBuildOrder = c.werfConfig.ImageTree(image)
		case *config.ImageFromDockerfile:
			imagesInBuildOrder = append(imagesInBuildOrder, image)
		}

		for i := 0; i < len(imagesInBuildOrder); i++ {
			if isNotInArr(images, imagesInBuildOrder[i]) {
				images = append(images, imagesInBuildOrder[i])
			}
		}
	}

	return images
}

func getImageConfigsToProcess(c *Conveyor) []config.ImageInterface {
	var imageConfigsToProcess []config.ImageInterface

	if len(c.imageNamesToProcess) == 0 {
		imageConfigsToProcess = c.werfConfig.GetAllImages()
	} else {
		for _, imageName := range c.imageNamesToProcess {
			imageToProcess := c.werfConfig.GetImage(imageName)
			if imageToProcess == nil {
				logboek.LogErrorF("WARNING: Specified image %s isn't defined in werf.yaml!\n", imageName)
			} else {
				imageConfigsToProcess = append(imageConfigsToProcess, imageToProcess)
			}
		}
	}

	return imageConfigsToProcess
}

func isNotInArr(arr []config.ImageInterface, obj config.ImageInterface) bool {
	for _, elm := range arr {
		if reflect.DeepEqual(elm, obj) {
			return false
		}
	}

	return true
}

func initStages(image *Image, imageInterfaceConfig config.StapelImageInterface, c *Conveyor) error {
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
		ScriptsDir:           getImageScriptsDir(imageName, c),
		ContainerArchivesDir: getImageArchivesContainerDir(c),
		ContainerScriptsDir:  getImageScriptsContainerDir(c),
	}

	gitPatchStageOptions := &stage.NewGitPatchStageOptions{
		PatchesDir:           getImagePatchesDir(imageName, c),
		ArchivesDir:          getImageArchivesDir(imageName, c),
		ScriptsDir:           getImageScriptsDir(imageName, c),
		ContainerPatchesDir:  getImagePatchesContainerDir(c),
		ContainerArchivesDir: getImageArchivesContainerDir(c),
		ContainerScriptsDir:  getImageScriptsContainerDir(c),
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

		stages = appendIfExist(stages, stage.GenerateDockerInstructionsStage(imageInterfaceConfig.(*config.StapelImage), baseStageOptions))
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

func generateGitMappings(imageBaseConfig *config.StapelImageBase, c *Conveyor) ([]*stage.GitMapping, error) {
	var gitMappings []*stage.GitMapping

	var localGitRepo *git_repo.Local
	if len(imageBaseConfig.Git.Local) != 0 {
		localGitRepo = &git_repo.Local{
			Base:   git_repo.Base{Name: "own"},
			Path:   c.projectDir,
			GitDir: filepath.Join(c.projectDir, ".git"),
		}
	}

	for _, localGitMappingConfig := range imageBaseConfig.Git.Local {
		gitMappings = append(gitMappings, gitLocalPathInit(localGitMappingConfig, localGitRepo, imageBaseConfig.Name, c))
	}

	for _, remoteGitMappingConfig := range imageBaseConfig.Git.Remote {
		remoteGitRepo, exist := c.remoteGitRepos[remoteGitMappingConfig.Name]
		if !exist {
			remoteGitRepo = &git_repo.Remote{
				Base: git_repo.Base{Name: remoteGitMappingConfig.Name},
				Url:  remoteGitMappingConfig.Url,
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

func gitRemoteArtifactInit(remoteGitMappingConfig *config.GitRemote, remoteGitRepo *git_repo.Remote, imageName string, c *Conveyor) *stage.GitMapping {
	gitMapping := baseGitMappingInit(remoteGitMappingConfig.GitLocalExport, imageName, c)

	gitMapping.Tag = remoteGitMappingConfig.Tag
	gitMapping.Commit = remoteGitMappingConfig.Commit
	gitMapping.Branch = remoteGitMappingConfig.Branch

	gitMapping.Name = remoteGitMappingConfig.Name

	gitMapping.GitRepoInterface = remoteGitRepo

	gitMapping.GitRepoCache = c.GetGitRepoCache(remoteGitRepo.GetName())

	return gitMapping
}

func gitLocalPathInit(localGitMappingConfig *config.GitLocal, localGitRepo *git_repo.Local, imageName string, c *Conveyor) *stage.GitMapping {
	gitMapping := baseGitMappingInit(localGitMappingConfig.GitLocalExport, imageName, c)

	gitMapping.Name = "own"

	gitMapping.GitRepoInterface = localGitRepo

	gitMapping.GitRepoCache = c.GetGitRepoCache(localGitRepo.GetName())

	return gitMapping
}

func baseGitMappingInit(local *config.GitLocalExport, imageName string, c *Conveyor) *stage.GitMapping {
	var stageDependencies map[stage.StageName][]string
	if local.StageDependencies != nil {
		stageDependencies = stageDependenciesToMap(local.GitMappingStageDependencies())
	}

	gitMapping := &stage.GitMapping{
		PatchesDir:           getImagePatchesDir(imageName, c),
		ContainerPatchesDir:  getImagePatchesContainerDir(c),
		ArchivesDir:          getImageArchivesDir(imageName, c),
		ContainerArchivesDir: getImageArchivesContainerDir(c),
		ScriptsDir:           getImageScriptsDir(imageName, c),
		ContainerScriptsDir:  getImageScriptsContainerDir(c),

		RepoPath: local.GitMappingAdd(),

		Cwd:                local.GitMappingAdd(),
		To:                 local.GitMappingTo(),
		ExcludePaths:       local.GitMappingExcludePath(),
		IncludePaths:       local.GitMappingIncludePaths(),
		Owner:              local.Owner,
		Group:              local.Group,
		StagesDependencies: stageDependencies,
	}

	return gitMapping
}

func getImagePatchesDir(imageName string, c *Conveyor) string {
	return filepath.Join(c.tmpDir, imageName, "patch")
}

func getImagePatchesContainerDir(c *Conveyor) string {
	return path.Join(c.containerWerfDir, "patch")
}

func getImageArchivesDir(imageName string, c *Conveyor) string {
	return filepath.Join(c.tmpDir, imageName, "archive")
}

func getImageArchivesContainerDir(c *Conveyor) string {
	return path.Join(c.containerWerfDir, "archive")
}

func getImageScriptsDir(imageName string, c *Conveyor) string {
	return filepath.Join(c.tmpDir, imageName, "scripts")
}

func getImageScriptsContainerDir(c *Conveyor) string {
	return path.Join(c.containerWerfDir, "scripts")
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

func prepareImageBasedOnImageFromDockerfile(imageFromDockerfileConfig *config.ImageFromDockerfile, c *Conveyor) (*Image, error) {
	image := &Image{}
	image.name = imageFromDockerfileConfig.Name

	contextDir := filepath.Join(c.projectDir, imageFromDockerfileConfig.Context)

	rel, err := filepath.Rel(c.projectDir, contextDir)
	if err != nil || strings.HasPrefix(rel, "../") {
		return nil, fmt.Errorf("unsupported context folder %s.\nOnly context folder specified inside project directory %s supported", contextDir, c.projectDir)
	}

	exist, err := util.DirExists(contextDir)
	if err != nil {
		return nil, err
	} else if !exist {
		return nil, fmt.Errorf("context folder %s is not found", contextDir)
	}

	dockerfilePath := filepath.Join(c.projectDir, imageFromDockerfileConfig.Dockerfile)
	rel, err = filepath.Rel(c.projectDir, dockerfilePath)
	if err != nil || strings.HasPrefix(rel, "../") {
		return nil, fmt.Errorf("unsupported dockerfile %s.\n Only dockerfile specified inside project directory %s supported", dockerfilePath, c.projectDir)
	}

	exist, err = util.FileExists(dockerfilePath)
	if err != nil {
		return nil, err
	} else if !exist {
		return nil, fmt.Errorf("dockerfile %s is not found", dockerfilePath)
	}

	dockerignorePatterns, err := build.ReadDockerignore(contextDir)
	if err != nil {
		return nil, err
	}

	var dockerignorePatternsWithContextPrefix []string
	for _, dockerignorePattern := range dockerignorePatterns {
		patterns := []string{dockerignorePattern}
		specialPrefixes := []string{
			"**/",
			"/**/",
			"!**/",
			"!/**/",
		}

		for _, prefix := range specialPrefixes {
			if strings.HasPrefix(dockerignorePattern, prefix) {
				patterns = append(patterns, strings.Replace(dockerignorePattern, "**/", "", 1))
				break
			}
		}

		for _, pattern := range patterns {
			var resultPattern string
			if strings.HasPrefix(pattern, "!") {
				resultPattern = "!" + path.Join(contextDir, pattern[1:])
			} else {
				resultPattern = path.Join(contextDir, pattern)
			}

			dockerignorePatternsWithContextPrefix = append(dockerignorePatternsWithContextPrefix, resultPattern)
		}
	}

	dockerignorePatternMatcher, err := fileutils.NewPatternMatcher(dockerignorePatternsWithContextPrefix)
	if err != nil {
		return nil, err
	}

	data, err := ioutil.ReadFile(dockerfilePath)
	if err != nil {
		return nil, err
	}

	p, err := parser.Parse(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}

	dockerStages, dockerMetaArgs, err := instructions.Parse(p.AST)
	if err != nil {
		return nil, err
	}

	resolveDockerStagesFromValue(dockerStages)

	dockerTargetIndex, err := getDockerTargetStageIndex(dockerStages, imageFromDockerfileConfig.Target)
	if err != nil {
		return nil, err
	}

	dockerTargetStage := dockerStages[dockerTargetIndex]

	dockerArgsHash := map[string]string{}
	var dockerMetaArgsString []string
	for _, arg := range dockerMetaArgs {
		dockerArgsHash[arg.Key] = arg.ValueString()
	}

	for key, valueInterf := range imageFromDockerfileConfig.Args {
		value := fmt.Sprintf("%v", valueInterf)
		dockerArgsHash[key] = value
	}

	for key, value := range dockerArgsHash {
		dockerMetaArgsString = append(dockerMetaArgsString, fmt.Sprintf("%s=%v", key, value))
	}

	shlex := shell.NewLex(parser.DefaultEscapeToken)
	resolvedBaseName, err := shlex.ProcessWord(dockerTargetStage.BaseName, dockerMetaArgsString)
	if err != nil {
		return nil, err
	}

	if err := handleImageFromName(resolvedBaseName, false, image, c); err != nil {
		return nil, err
	}

	baseStageOptions := &stage.NewBaseStageOptions{
		ImageName:   imageFromDockerfileConfig.Name,
		ProjectName: c.werfConfig.Meta.Project,
	}

	dockerfileStage := stage.GenerateDockerfileStage(
		dockerfilePath,
		imageFromDockerfileConfig.Target,
		contextDir,
		dockerignorePatternMatcher,
		imageFromDockerfileConfig.Args,
		imageFromDockerfileConfig.AddHost,
		dockerStages,
		dockerArgsHash,
		dockerTargetIndex,
		baseStageOptions)

	image.stages = append(image.stages, dockerfileStage)

	logboek.LogInfoF("Using stage %s\n", dockerfileStage.Name())

	return image, nil
}

func resolveDockerStagesFromValue(stages []instructions.Stage) {
	nameToIndex := make(map[string]string)
	for i, s := range stages {
		index := strconv.Itoa(i)
		if s.Name != index {
			nameToIndex[s.Name] = index
		}
		for _, cmd := range s.Commands {
			switch c := cmd.(type) {
			case *instructions.CopyCommand:
				if c.From != "" {
					if val, ok := nameToIndex[c.From]; ok {
						c.From = val
					}

				}
			}
		}
	}
}

func getDockerTargetStageIndex(dockerStages []instructions.Stage, dockerTargetStage string) (int, error) {
	if dockerTargetStage == "" {
		return len(dockerStages) - 1, nil
	}

	for i, s := range dockerStages {
		if s.Name == dockerTargetStage {
			return i, nil
		}
	}

	return -1, fmt.Errorf("%s is not a valid target build stage", dockerTargetStage)
}
