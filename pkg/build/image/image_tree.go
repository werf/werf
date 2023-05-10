package image

import (
	"context"
	"fmt"
	"path"
	"path/filepath"
	"reflect"
	"sort"

	"github.com/gookit/color"

	"github.com/werf/logboek"
	"github.com/werf/logboek/pkg/types"
	"github.com/werf/werf/pkg/build/stage"
	"github.com/werf/werf/pkg/config"
	"github.com/werf/werf/pkg/git_repo"
	"github.com/werf/werf/pkg/giterminism_manager"
	"github.com/werf/werf/pkg/logging"
	"github.com/werf/werf/pkg/util"
)

type ImagesTree struct {
	ImagesTreeOptions

	werfConfig *config.WerfConfig

	allImages  []*Image
	imagesSets ImagesSets

	multiplatformImages []*MultiplatformImage
}

type ImagesTreeOptions struct {
	CommonImageOptions

	OnlyImages    []string
	WithoutImages bool
}

func NewImagesTree(werfConfig *config.WerfConfig, opts ImagesTreeOptions) *ImagesTree {
	return &ImagesTree{
		ImagesTreeOptions: opts,
		werfConfig:        werfConfig,
	}
}

func (tree *ImagesTree) Calculate(ctx context.Context) error {
	imageConfigSets, err := tree.werfConfig.GroupImagesByIndependentSets(tree.OnlyImages, tree.WithoutImages)
	if err != nil {
		return fmt.Errorf("unable to group werf config images by independent sets: %w", err)
	}

	forcedTargetPlatforms := tree.Conveyor.GetForcedTargetPlatforms()
	commonTargetPlatforms, err := tree.Conveyor.GetTargetPlatforms()
	if err != nil {
		return fmt.Errorf("invalid common target platforms: %w", err)
	}
	if len(commonTargetPlatforms) == 0 {
		commonTargetPlatforms = []string{tree.ContainerBackend.GetDefaultPlatform()}
	}

	commonImageOpts := tree.CommonImageOptions
	builder := NewImagesSetsBuilder()

	for _, iteration := range imageConfigSets {
		for _, imageConfigI := range iteration {
			var targetPlatforms []string
			if len(forcedTargetPlatforms) > 0 {
				targetPlatforms = forcedTargetPlatforms
			} else {
				imageTargetPlatforms, err := tree.Conveyor.GetImageTargetPlatforms(imageConfigI.GetName())
				if err != nil {
					return fmt.Errorf("invalid image %q target platforms: %w", imageConfigI.GetName(), err)
				}
				if len(imageTargetPlatforms) > 0 {
					targetPlatforms = imageTargetPlatforms
				} else {
					targetPlatforms = commonTargetPlatforms
				}
			}

			commonImageOpts.ForceTargetPlatformLogging = (len(targetPlatforms) > 1)

			for _, targetPlatform := range targetPlatforms {
				var imageLogName string
				var style color.Style

				switch imageConfig := imageConfigI.(type) {
				case config.StapelImageInterface:
					imageLogName = logging.ImageLogProcessName(imageConfig.ImageBaseConfig().Name, imageConfig.IsArtifact(), targetPlatform)
					style = ImageLogProcessStyle(imageConfig.IsArtifact())
				case *config.ImageFromDockerfile:
					imageLogName = logging.ImageLogProcessName(imageConfig.Name, false, targetPlatform)
					style = ImageLogProcessStyle(false)
				}

				err := logboek.Context(ctx).Info().LogProcess(imageLogName).
					Options(func(options types.LogProcessOptionsInterface) {
						options.Style(style)
					}).
					DoError(func() error {
						var err error
						var newImagesSets ImagesSets

						switch imageConfig := imageConfigI.(type) {
						case config.StapelImageInterface:
							newImagesSets, err = MapStapelConfigToImagesSets(ctx, tree.werfConfig.Meta, imageConfig, targetPlatform, commonImageOpts)
							if err != nil {
								return fmt.Errorf("unable to map stapel config to images sets: %w", err)
							}

						case *config.ImageFromDockerfile:
							newImagesSets, err = MapDockerfileConfigToImagesSets(ctx, imageConfig, targetPlatform, commonImageOpts)
							if err != nil {
								return fmt.Errorf("unable to map dockerfile to images sets: %w", err)
							}
						}

						builder.MergeImagesSets(newImagesSets)

						return nil
					})
				if err != nil {
					return err
				}
			}
		}

		builder.Next()
	}

	tree.imagesSets = builder.GetImagesSets()
	tree.allImages = builder.GetAllImages()

	return nil
}

func (tree *ImagesTree) GetImage(name string) *Image {
	return nil
}

func (tree *ImagesTree) GetImagesByName(finalOnly bool) []util.Pair[string, []*Image] {
	images := make(map[string]map[string]*Image)
	var names []string

	appendImage := func(img *Image) {
		names = util.UniqAppendString(names, img.Name)
		if images[img.Name] == nil {
			images[img.Name] = make(map[string]*Image)
		}
		images[img.Name][img.TargetPlatform] = img
	}

	for _, img := range tree.GetImages() {
		if finalOnly {
			for _, finalImageName := range tree.werfConfig.GetAllImages() {
				if finalImageName.GetName() == img.Name {
					appendImage(img)
				}
			}
		} else {
			appendImage(img)
		}
	}

	var res []util.Pair[string, []*Image]

	sort.Strings(names)
	for _, name := range names {
		platforms := util.MapKeys(images[name])
		sort.Strings(platforms)

		allPlatformsImages := util.NewPair(name, make([]*Image, 0, len(platforms)))
		for _, platform := range platforms {
			allPlatformsImages.Second = append(allPlatformsImages.Second, images[name][platform])
		}
		res = append(res, allPlatformsImages)
	}

	return res
}

func (tree *ImagesTree) GetImagePlatformsByName(finalOnly bool) map[string][]string {
	res := make(map[string][]string)
	for _, img := range tree.GetImages() {
		if finalOnly {
			for _, finalImageName := range tree.werfConfig.GetAllImages() {
				if finalImageName.GetName() == img.Name {
					res[img.Name] = append(res[img.Name], img.TargetPlatform)
				}
			}
		} else {
			res[img.Name] = append(res[img.Name], img.TargetPlatform)
		}
	}
	return res
}

func (tree *ImagesTree) GetImagesNames() (res []string) {
	for _, img := range tree.allImages {
		res = util.UniqAppendString(res, img.Name)
	}
	return
}

func (tree *ImagesTree) GetImages() []*Image {
	return tree.allImages
}

func (tree *ImagesTree) GetImagesSets() ImagesSets {
	return tree.imagesSets
}

func (tree *ImagesTree) GetMultiplatformImage(name string) *MultiplatformImage {
	for _, img := range tree.multiplatformImages {
		if img.Name == name {
			return img
		}
	}
	return nil
}

func (tree *ImagesTree) SetMultiplatformImage(newImg *MultiplatformImage) {
	for _, img := range tree.multiplatformImages {
		if img.Name == newImg.Name {
			return
		}
	}
	tree.multiplatformImages = append(tree.multiplatformImages, newImg)
}

func (tree *ImagesTree) GetMultiplatformImages() []*MultiplatformImage {
	return tree.multiplatformImages
}

func getFromFields(imageBaseConfig *config.StapelImageBase) (string, string, bool) {
	var from string
	var fromImageName string

	switch {
	case imageBaseConfig.From != "":
		from = imageBaseConfig.From
	case imageBaseConfig.FromImageName != "":
		fromImageName = imageBaseConfig.FromImageName
	case imageBaseConfig.FromArtifactName != "":
		fromImageName = imageBaseConfig.FromArtifactName
	}

	return from, fromImageName, imageBaseConfig.FromLatest
}

func filterAndLogGitMappings(ctx context.Context, gitMappings []*stage.GitMapping, conveyor Conveyor) ([]*stage.GitMapping, error) {
	var res []*stage.GitMapping

	for ind, gitMapping := range gitMappings {
		if err := logboek.Context(ctx).Info().LogProcess(fmt.Sprintf("[%d] git mapping from %s repository", ind, gitMapping.Name)).DoError(func() error {
			withTripleIndent := func(f func()) {
				if logboek.Context(ctx).Info().IsAccepted() {
					logboek.Context(ctx).Streams().IncreaseIndent()
					logboek.Context(ctx).Streams().IncreaseIndent()
					logboek.Context(ctx).Streams().IncreaseIndent()
				}

				f()

				if logboek.Context(ctx).Info().IsAccepted() {
					logboek.Context(ctx).Streams().DecreaseIndent()
					logboek.Context(ctx).Streams().DecreaseIndent()
					logboek.Context(ctx).Streams().DecreaseIndent()
				}
			}

			withTripleIndent(func() {
				logboek.Context(ctx).Info().LogFDetails("add: %s\n", gitMapping.Add)
				logboek.Context(ctx).Info().LogFDetails("to: %s\n", gitMapping.To)

				if len(gitMapping.IncludePaths) != 0 {
					logboek.Context(ctx).Info().LogFDetails("includePaths: %+v\n", gitMapping.IncludePaths)
				}

				if len(gitMapping.ExcludePaths) != 0 {
					logboek.Context(ctx).Info().LogFDetails("excludePaths: %+v\n", gitMapping.ExcludePaths)
				}

				if gitMapping.Commit != "" {
					logboek.Context(ctx).Info().LogFDetails("commit: %s\n", gitMapping.Commit)
				}

				if gitMapping.Branch != "" {
					logboek.Context(ctx).Info().LogFDetails("branch: %s\n", gitMapping.Branch)
				}

				if gitMapping.Owner != "" {
					logboek.Context(ctx).Info().LogFDetails("owner: %s\n", gitMapping.Owner)
				}

				if gitMapping.Group != "" {
					logboek.Context(ctx).Info().LogFDetails("group: %s\n", gitMapping.Group)
				}

				if len(gitMapping.StagesDependencies) != 0 {
					logboek.Context(ctx).Info().LogLnDetails("stageDependencies:")

					for s, values := range gitMapping.StagesDependencies {
						if len(values) != 0 {
							logboek.Context(ctx).Info().LogFDetails("  %s: %v\n", s, values)
						}
					}
				}
			})

			logboek.Context(ctx).Info().LogLn()

			commitInfo, err := gitMapping.GetLatestCommitInfo(ctx, conveyor)
			if err != nil {
				return fmt.Errorf("unable to get commit of repo %q: %w", gitMapping.GitRepo().GetName(), err)
			}

			if commitInfo.VirtualMerge {
				logboek.Context(ctx).Info().LogFDetails("Commit %s will be used (virtual merge of %s into %s)\n", commitInfo.Commit, commitInfo.VirtualMergeFromCommit, commitInfo.VirtualMergeIntoCommit)
			} else {
				logboek.Context(ctx).Info().LogFDetails("Commit %s will be used\n", commitInfo.Commit)
			}

			res = append(res, gitMapping)

			return nil
		}); err != nil {
			return nil, err
		}
	}

	return res, nil
}

func appendIfExist(ctx context.Context, stages []stage.Interface, stage stage.Interface) []stage.Interface {
	if !reflect.ValueOf(stage).IsNil() {
		logboek.Context(ctx).Info().LogFDetails("Using stage %s\n", stage.Name())
		return append(stages, stage)
	}

	return stages
}

func gitRemoteArtifactInit(ctx context.Context, remoteGitMappingConfig *config.GitRemote, remoteGitRepo *git_repo.Remote, imageName string, conveyor Conveyor, containerWerfDir, tmpDir string) (*stage.GitMapping, error) {
	gitMapping := baseGitMappingInit(remoteGitMappingConfig.GitLocalExport, imageName, conveyor, containerWerfDir, tmpDir)

	gitMapping.Tag = remoteGitMappingConfig.Tag
	gitMapping.Commit = remoteGitMappingConfig.Commit
	gitMapping.Branch = remoteGitMappingConfig.Branch

	gitMapping.Name = remoteGitMappingConfig.Name
	gitMapping.SetGitRepo(remoteGitRepo)

	gitMappingTo, err := makeGitMappingTo(ctx, gitMapping, remoteGitMappingConfig.GitLocalExport.GitMappingTo(), conveyor)
	if err != nil {
		return nil, fmt.Errorf("unable to make remote git.to mapping for image %q: %w", imageName, err)
	}
	gitMapping.To = gitMappingTo

	return gitMapping, nil
}

func gitLocalPathInit(ctx context.Context, localGitMappingConfig *config.GitLocal, imageName string, conveyor Conveyor, giterminismManager giterminism_manager.Interface, containerWerfDir, tmpDir string) (*stage.GitMapping, error) {
	gitMapping := baseGitMappingInit(localGitMappingConfig.GitLocalExport, imageName, conveyor, containerWerfDir, tmpDir)

	gitMapping.Name = "own"
	gitMapping.SetGitRepo(giterminismManager.LocalGitRepo())

	gitMappingTo, err := makeGitMappingTo(ctx, gitMapping, localGitMappingConfig.GitLocalExport.GitMappingTo(), conveyor)
	if err != nil {
		return nil, fmt.Errorf("unable to make local git.to mapping for image %q: %w", imageName, err)
	}
	gitMapping.To = gitMappingTo

	return gitMapping, nil
}

func baseGitMappingInit(local *config.GitLocalExport, imageName string, conveyor Conveyor, containerWerfDir, tmpDir string) *stage.GitMapping {
	var stageDependencies map[stage.StageName][]string
	if local.StageDependencies != nil {
		stageDependencies = stageDependenciesToMap(local.StageDependencies)
	}

	gitMapping := stage.NewGitMapping()

	gitMapping.ContainerPatchesDir = path.Join(containerWerfDir, "patch")
	gitMapping.ContainerArchivesDir = path.Join(containerWerfDir, "archive")
	gitMapping.ScriptsDir = filepath.Join(tmpDir, imageName, "scripts")
	gitMapping.ContainerScriptsDir = path.Join(containerWerfDir, "scripts")

	gitMapping.Add = local.GitMappingAdd()
	gitMapping.ExcludePaths = local.ExcludePaths
	gitMapping.IncludePaths = local.IncludePaths
	gitMapping.Owner = local.Owner
	gitMapping.Group = local.Group
	gitMapping.StagesDependencies = stageDependencies

	return gitMapping
}

func makeGitMappingTo(ctx context.Context, gitMapping *stage.GitMapping, gitMappingTo string, conveyor Conveyor) (string, error) {
	if gitMappingTo != "/" {
		return gitMappingTo, nil
	}

	gitRepoName := gitMapping.GitRepo().GetName()
	commitInfo, err := gitMapping.GetLatestCommitInfo(ctx, conveyor)
	if err != nil {
		return "", fmt.Errorf("unable to get latest commit info for repo %q: %w", gitRepoName, err)
	}

	if gitMappingAddIsDir, err := gitMapping.GitRepo().IsCommitTreeEntryDirectory(ctx, commitInfo.Commit, gitMapping.Add); err != nil {
		return "", fmt.Errorf("unable to determine whether git `add: %s` is dir or file for repo %q: %w", gitMapping.Add, gitRepoName, err)
	} else if !gitMappingAddIsDir {
		return "", fmt.Errorf("for git repo %q specifying `to: /` when adding a single file from git with `add: %s` is not allowed. Fix this by changing `to: /` to `to: /%s`.", gitRepoName, gitMapping.Add, filepath.Base(gitMapping.Add))
	}

	return gitMappingTo, nil
}

func stageDependenciesToMap(sd *config.StageDependencies) map[stage.StageName][]string {
	result := map[stage.StageName][]string{
		stage.Install:     sd.Install,
		stage.BeforeSetup: sd.BeforeSetup,
		stage.Setup:       sd.Setup,
	}

	return result
}
