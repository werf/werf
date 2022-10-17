package image

import (
	"bytes"
	"context"
	"fmt"
	"path/filepath"

	"github.com/docker/docker/builder/dockerignore"
	"github.com/moby/buildkit/frontend/dockerfile/instructions"
	"github.com/moby/buildkit/frontend/dockerfile/parser"

	"github.com/werf/logboek"
	"github.com/werf/werf/pkg/build/stage"
	stage_instruction "github.com/werf/werf/pkg/build/stage/instruction"
	"github.com/werf/werf/pkg/config"
	backend_instruction "github.com/werf/werf/pkg/container_backend/instruction"
	"github.com/werf/werf/pkg/dockerfile"
	"github.com/werf/werf/pkg/dockerfile/frontend"
	dockerfile_instruction "github.com/werf/werf/pkg/dockerfile/instruction"
	"github.com/werf/werf/pkg/giterminism_manager"
	"github.com/werf/werf/pkg/path_matcher"
	"github.com/werf/werf/pkg/util"
)

func MapDockerfileConfigToImagesSets(ctx context.Context, dockerfileImageConfig *config.ImageFromDockerfile, opts CommonImageOptions) (ImagesSets, error) {
	if dockerfileImageConfig.Staged {
		relDockerfilePath := filepath.Join(dockerfileImageConfig.Context, dockerfileImageConfig.Dockerfile)
		dockerfileData, err := opts.GiterminismManager.FileReader().ReadDockerfile(ctx, relDockerfilePath)
		if err != nil {
			return nil, fmt.Errorf("unable to read dockerfile %s: %w", relDockerfilePath, err)
		}

		d, err := frontend.ParseDockerfileWithBuildkit(dockerfileData, dockerfile.DockerfileOptions{
			Target:    dockerfileImageConfig.Target,
			BuildArgs: util.MapStringInterfaceToMapStringString(dockerfileImageConfig.Args),
			AddHost:   dockerfileImageConfig.AddHost,
			Network:   dockerfileImageConfig.Network,
			SSH:       dockerfileImageConfig.SSH,
		})
		if err != nil {
			return nil, fmt.Errorf("unable to parse dockerfile %s: %w", relDockerfilePath, err)
		}

		return mapDockerfileToImagesSets(ctx, d, dockerfileImageConfig, opts)
	}

	img, err := mapLegacyDockerfileToImage(ctx, dockerfileImageConfig, opts)
	if err != nil {
		return nil, err
	}

	var ret ImagesSets

	ret = append(ret, []*Image{img})

	return ret, nil
}

func mapDockerfileToImagesSets(ctx context.Context, cfg *dockerfile.Dockerfile, dockerfileImageConfig *config.ImageFromDockerfile, opts CommonImageOptions) (ImagesSets, error) {
	var ret ImagesSets

	stagesSets, err := cfg.GroupStagesByIndependentSets(ctx)
	if err != nil {
		return nil, fmt.Errorf("unable to group dockerfile stages by independent sets :%w", err)
	}

	for _, set := range stagesSets {
		for _, stg := range set {
			_ = stg
			// TODO(staged-dockerfile): map stages to *Image+build.Stage objects
			// ret = append(ret, stg.)
		}
	}

	{
		// TODO parse FROM instruction properly, set correct BaseImageReference here

		img, err := NewImage(ctx, "test", ImageFromRegistryAsBaseImage, ImageOptions{
			IsDockerfileImage:     true,
			DockerfileImageConfig: dockerfileImageConfig,
			CommonImageOptions:    opts,
			BaseImageReference:    "ubuntu:22.04",
		})
		if err != nil {
			return nil, fmt.Errorf("unable to create image %q: %w", "test", err)
		}

		img.stages = append(img.stages, stage_instruction.NewRun(backend_instruction.NewRun(*dockerfile_instruction.NewRun([]string{"ls", "/"}, false, nil, "", "")), nil, false, &stage.BaseStageOptions{
			ImageName:        img.Name,
			ImageTmpDir:      img.TmpDir,
			ContainerWerfDir: img.ContainerWerfDir,
			ProjectName:      opts.ProjectName,
		}))

		ret = append(ret, []*Image{img})
	}

	return ret, nil
}

func mapLegacyDockerfileToImage(ctx context.Context, dockerfileImageConfig *config.ImageFromDockerfile, opts CommonImageOptions) (*Image, error) {
	img, err := NewImage(ctx, dockerfileImageConfig.Name, NoBaseImage, ImageOptions{
		CommonImageOptions:    opts,
		IsDockerfileImage:     true,
		DockerfileImageConfig: dockerfileImageConfig,
	})
	if err != nil {
		return nil, fmt.Errorf("unable to create image %q: %w", dockerfileImageConfig.Name, err)
	}

	for _, contextAddFile := range dockerfileImageConfig.ContextAddFiles {
		relContextAddFile := filepath.Join(dockerfileImageConfig.Context, contextAddFile)
		absContextAddFile := filepath.Join(opts.ProjectDir, relContextAddFile)
		exist, err := util.FileExists(absContextAddFile)
		if err != nil {
			return nil, fmt.Errorf("unable to check existence of file %s: %w", absContextAddFile, err)
		}

		if !exist {
			return nil, fmt.Errorf("contextAddFile %q was not found (the path must be relative to the context %q)", contextAddFile, dockerfileImageConfig.Context)
		}
	}

	dockerIgnorePathMatcher, err := createDockerIgnorePathMatcher(ctx, opts.GiterminismManager, dockerfileImageConfig.Context, dockerfileImageConfig.Dockerfile)
	if err != nil {
		return nil, fmt.Errorf("unable to create dockerignore path matcher: %w", err)
	}

	relDockerfilePath := filepath.Join(dockerfileImageConfig.Context, dockerfileImageConfig.Dockerfile)
	dockerfileData, err := opts.GiterminismManager.FileReader().ReadDockerfile(ctx, relDockerfilePath)
	if err != nil {
		return nil, err
	}

	p, err := parser.Parse(bytes.NewReader(dockerfileData))
	if err != nil {
		return nil, err
	}

	dockerStages, dockerMetaArgs, err := instructions.Parse(p.AST)
	if err != nil {
		return nil, err
	}

	frontend.ResolveDockerStagesFromValue(dockerStages)

	dockerTargetIndex, err := frontend.GetDockerTargetStageIndex(dockerStages, dockerfileImageConfig.Target)
	if err != nil {
		return nil, err
	}

	ds := stage.NewDockerStages(
		dockerStages,
		util.MapStringInterfaceToMapStringString(dockerfileImageConfig.Args),
		dockerMetaArgs,
		dockerTargetIndex,
	)

	baseStageOptions := &stage.BaseStageOptions{
		ImageName:   dockerfileImageConfig.Name,
		ProjectName: opts.ProjectName,
	}

	dockerfileStage := stage.GenerateFullDockerfileStage(
		stage.NewDockerRunArgs(
			dockerfileData,
			dockerfileImageConfig.Dockerfile,
			dockerfileImageConfig.Target,
			dockerfileImageConfig.Context,
			dockerfileImageConfig.ContextAddFiles,
			dockerfileImageConfig.Args,
			dockerfileImageConfig.AddHost,
			dockerfileImageConfig.Network,
			dockerfileImageConfig.SSH,
		),
		ds,
		stage.NewContextChecksum(dockerIgnorePathMatcher),
		baseStageOptions,
		dockerfileImageConfig.Dependencies,
	)

	img.stages = append(img.stages, dockerfileStage)

	logboek.Context(ctx).Info().LogFDetails("Using stage %s\n", dockerfileStage.Name())

	return img, nil
}

func createDockerIgnorePathMatcher(ctx context.Context, giterminismMgr giterminism_manager.Interface, contextGitSubDir, dockerfileRelToContextPath string) (path_matcher.PathMatcher, error) {
	dockerfileRelToGitPath := filepath.Join(contextGitSubDir, dockerfileRelToContextPath)

	var dockerIgnorePatterns []string
	for _, dockerIgnoreRelToContextPath := range []string{
		dockerfileRelToContextPath + ".dockerignore",
		".dockerignore",
	} {
		dockerIgnoreRelToGitPath := filepath.Join(contextGitSubDir, dockerIgnoreRelToContextPath)
		if exist, err := giterminismMgr.FileReader().IsDockerignoreExistAnywhere(ctx, dockerIgnoreRelToGitPath); err != nil {
			return nil, err
		} else if !exist {
			continue
		}

		dockerIgnore, err := giterminismMgr.FileReader().ReadDockerignore(ctx, dockerIgnoreRelToGitPath)
		if err != nil {
			return nil, err
		}

		r := bytes.NewReader(dockerIgnore)
		dockerIgnorePatterns, err = dockerignore.ReadAll(r)
		if err != nil {
			return nil, fmt.Errorf("unable to read %q file: %w", dockerIgnoreRelToContextPath, err)
		}

		break
	}

	dockerIgnorePathMatcher := path_matcher.NewPathMatcher(path_matcher.PathMatcherOptions{
		BasePath:             filepath.Join(giterminismMgr.RelativeToGitProjectDir(), contextGitSubDir),
		DockerignorePatterns: dockerIgnorePatterns,
	})

	if !dockerIgnorePathMatcher.IsPathMatched(dockerfileRelToGitPath) {
		logboek.Context(ctx).Warn().LogLn("WARNING: There is no way to ignore the Dockerfile due to docker limitation when building an image for a compressed context that reads from STDIN.")
		logboek.Context(ctx).Warn().LogF("WARNING: To hide this message, remove the Dockerfile ignore rule or add an exception rule.\n")

		exceptionRule := "!" + dockerfileRelToContextPath
		dockerIgnorePatterns = append(dockerIgnorePatterns, exceptionRule)
		dockerIgnorePathMatcher = path_matcher.NewPathMatcher(path_matcher.PathMatcherOptions{
			BasePath:             filepath.Join(giterminismMgr.RelativeToGitProjectDir(), contextGitSubDir),
			DockerignorePatterns: dockerIgnorePatterns,
		})
	}

	return dockerIgnorePathMatcher, nil
}
