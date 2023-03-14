package image

import (
	"context"
	"fmt"
	"path"
	"path/filepath"

	"github.com/werf/logboek"
	"github.com/werf/werf/pkg/build/stage"
	"github.com/werf/werf/pkg/config"
	"github.com/werf/werf/pkg/git_repo"
)

func MapStapelConfigToImagesSets(ctx context.Context, metaConfig *config.Meta, stapelImageConfig config.StapelImageInterface, targetPlatform string, opts CommonImageOptions) (ImagesSets, error) {
	img, err := mapStapelConfigToImage(ctx, metaConfig, stapelImageConfig, targetPlatform, opts)
	if err != nil {
		return nil, err
	}

	var ret ImagesSets

	ret = append(ret, []*Image{img})

	return ret, nil
}

func mapStapelConfigToImage(ctx context.Context, metaConfig *config.Meta, stapelImageConfig config.StapelImageInterface, targetPlatform string, opts CommonImageOptions) (*Image, error) {
	imageBaseConfig := stapelImageConfig.ImageBaseConfig()
	imageName := imageBaseConfig.Name
	imageArtifact := stapelImageConfig.IsArtifact()
	from, fromImageName, fromLatest := getFromFields(imageBaseConfig)

	imageOpts := ImageOptions{
		CommonImageOptions: opts,
		IsArtifact:         imageArtifact,
	}

	var baseImageType BaseImageType
	if from != "" {
		baseImageType = ImageFromRegistryAsBaseImage
		imageOpts.BaseImageReference = from
		imageOpts.FetchLatestBaseImage = fromLatest
	} else {
		baseImageType = StageAsBaseImage
		imageOpts.BaseImageName = fromImageName
	}

	image, err := NewImage(ctx, targetPlatform, imageName, baseImageType, imageOpts)
	if err != nil {
		return nil, fmt.Errorf("unable to create image %q: %w", imageName, err)
	}

	if err := initStages(ctx, image, metaConfig, stapelImageConfig, opts); err != nil {
		return nil, err
	}

	return image, nil
}

func initStages(ctx context.Context, image *Image, metaConfig *config.Meta, stapelImageConfig config.StapelImageInterface, opts CommonImageOptions) error {
	var stages []stage.Interface

	imageBaseConfig := stapelImageConfig.ImageBaseConfig()
	imageName := imageBaseConfig.Name
	imageArtifact := stapelImageConfig.IsArtifact()

	baseStageOptions := &stage.BaseStageOptions{
		TargetPlatform:   image.TargetPlatform,
		ImageName:        imageName,
		ConfigMounts:     imageBaseConfig.Mount,
		ImageTmpDir:      filepath.Join(opts.TmpDir, "image", imageBaseConfig.Name),
		ContainerWerfDir: opts.ContainerWerfDir,
		ProjectName:      opts.ProjectName,
	}

	gitArchiveStageOptions := &stage.NewGitArchiveStageOptions{
		ScriptsDir:           filepath.Join(opts.TmpDir, imageName, "scripts"),
		ContainerArchivesDir: path.Join(opts.ContainerWerfDir, "archive"),
		ContainerScriptsDir:  path.Join(opts.ContainerWerfDir, "scripts"),
	}

	gitPatchStageOptions := &stage.NewGitPatchStageOptions{
		ScriptsDir:           filepath.Join(opts.TmpDir, imageName, "scripts"),
		ContainerPatchesDir:  path.Join(opts.ContainerWerfDir, "patch"),
		ContainerArchivesDir: path.Join(opts.ContainerWerfDir, "archive"),
		ContainerScriptsDir:  path.Join(opts.ContainerWerfDir, "scripts"),
	}

	gitMappings, err := generateGitMappings(ctx, metaConfig, imageBaseConfig, opts)
	if err != nil {
		return err
	}

	gitMappingsExist := len(gitMappings) != 0

	stages = appendIfExist(ctx, stages, stage.GenerateFromStage(imageBaseConfig, image.baseImageRepoId, baseStageOptions))
	stages = appendIfExist(ctx, stages, stage.GenerateBeforeInstallStage(ctx, imageBaseConfig, baseStageOptions))
	stages = appendIfExist(ctx, stages, stage.GenerateDependenciesBeforeInstallStage(imageBaseConfig, baseStageOptions))

	if gitMappingsExist {
		stages = append(stages, stage.NewGitArchiveStage(gitArchiveStageOptions, baseStageOptions))
	}

	stages = appendIfExist(ctx, stages, stage.GenerateInstallStage(ctx, imageBaseConfig, gitPatchStageOptions, baseStageOptions))
	stages = appendIfExist(ctx, stages, stage.GenerateDependenciesAfterInstallStage(imageBaseConfig, baseStageOptions))
	stages = appendIfExist(ctx, stages, stage.GenerateBeforeSetupStage(ctx, imageBaseConfig, gitPatchStageOptions, baseStageOptions))
	stages = appendIfExist(ctx, stages, stage.GenerateDependenciesBeforeSetupStage(imageBaseConfig, baseStageOptions))
	stages = appendIfExist(ctx, stages, stage.GenerateSetupStage(ctx, imageBaseConfig, gitPatchStageOptions, baseStageOptions))
	stages = appendIfExist(ctx, stages, stage.GenerateDependenciesAfterSetupStage(imageBaseConfig, baseStageOptions))

	if !imageArtifact {
		if gitMappingsExist {
			stages = append(stages, stage.NewGitCacheStage(gitPatchStageOptions, baseStageOptions))
			stages = append(stages, stage.NewGitLatestPatchStage(gitPatchStageOptions, baseStageOptions))
		}

		stages = appendIfExist(ctx, stages, stage.GenerateStapelDockerInstructionsStage(stapelImageConfig.(*config.StapelImage), baseStageOptions))
	}

	if len(gitMappings) != 0 {
		logboek.Context(ctx).Info().LogLnDetails("Using git stages")

		for _, s := range stages {
			s.SetGitMappings(gitMappings)
		}
	}

	image.SetStages(stages)

	return nil
}

func generateGitMappings(ctx context.Context, metaConfig *config.Meta, imageBaseConfig *config.StapelImageBase, opts CommonImageOptions) ([]*stage.GitMapping, error) {
	var gitMappings []*stage.GitMapping

	if len(imageBaseConfig.Git.Local) != 0 {
		localGitRepo := opts.GiterminismManager.LocalGitRepo()

		if !metaConfig.GitWorktree.GetForceShallowClone() {
			isShallowClone, err := localGitRepo.IsShallowClone(ctx)
			if err != nil {
				return nil, fmt.Errorf("check shallow clone failed: %w", err)
			}

			if isShallowClone {
				if metaConfig.GitWorktree.GetAllowUnshallow() {
					if err := localGitRepo.Unshallow(ctx); err != nil {
						return nil, fmt.Errorf("unable to fetch local git repo: %w", err)
					}
				} else {
					logboek.Context(ctx).Warn().LogLn("The usage of shallow git clone may break reproducibility and slow down incremental rebuilds.")
					logboek.Context(ctx).Warn().LogLn("It is recommended to enable automatic unshallow of the git worktree with gitWorktree.allowUnshallow=true werf.yaml directive (which is enabled by default, http://werf.io/documentation/reference/werf_yaml.html#git-worktree).")
					logboek.Context(ctx).Warn().LogLn("If you still want to use shallow clone, then add gitWorktree.forceShallowClone=true werf.yaml directive (http://werf.io/documentation/reference/werf_yaml.html#git-worktree).")

					return nil, fmt.Errorf("shallow git clone is not allowed")
				}
			}
		}

		for _, localGitMappingConfig := range imageBaseConfig.Git.Local {
			gitMapping, err := gitLocalPathInit(ctx, localGitMappingConfig, imageBaseConfig.Name, opts.Conveyor, opts.GiterminismManager, opts.ContainerWerfDir, opts.TmpDir)
			if err != nil {
				return nil, err
			}
			gitMappings = append(gitMappings, gitMapping)
		}
	}

	for _, remoteGitMappingConfig := range imageBaseConfig.Git.Remote {
		remoteGitRepo := opts.Conveyor.GetRemoteGitRepo(remoteGitMappingConfig.Name)
		if remoteGitRepo == nil {
			var err error
			remoteGitRepo, err = git_repo.OpenRemoteRepo(remoteGitMappingConfig.Name, remoteGitMappingConfig.Url)
			if err != nil {
				return nil, fmt.Errorf("unable to open remote git repo %s by url %s: %w", remoteGitMappingConfig.Name, remoteGitMappingConfig.Url, err)
			}

			if err := logboek.Context(ctx).Info().LogProcess(fmt.Sprintf("Refreshing %s repository", remoteGitMappingConfig.Name)).
				DoError(func() error {
					return remoteGitRepo.CloneAndFetch(ctx)
				}); err != nil {
				return nil, err
			}

			opts.Conveyor.SetRemoteGitRepo(remoteGitMappingConfig.Name, remoteGitRepo)
		}

		gitMapping, err := gitRemoteArtifactInit(ctx, remoteGitMappingConfig, remoteGitRepo, imageBaseConfig.Name, opts.Conveyor, opts.ContainerWerfDir, opts.TmpDir)
		if err != nil {
			return nil, err
		}
		gitMappings = append(gitMappings, gitMapping)
	}

	var res []*stage.GitMapping

	if len(gitMappings) != 0 {
		err := logboek.Context(ctx).Info().LogProcess("Initializing git mappings").DoError(func() error {
			resGitMappings, err := filterAndLogGitMappings(ctx, gitMappings, opts.Conveyor)
			if err != nil {
				return err
			}

			res = resGitMappings

			return nil
		})
		if err != nil {
			return nil, err
		}
	}

	return res, nil
}
