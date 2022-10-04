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
	"github.com/werf/werf/pkg/config"
	"github.com/werf/werf/pkg/dockerfile"
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

		d, err := dockerfile.ParseDockerfile(dockerfileData, dockerfile.DockerfileOptions{
			Target:    dockerfileImageConfig.Target,
			BuildArgs: util.MapStringInterfaceToMapStringString(dockerfileImageConfig.Args),
			AddHost:   dockerfileImageConfig.AddHost,
			Network:   dockerfileImageConfig.Network,
			SSH:       dockerfileImageConfig.SSH,
		})
		if err != nil {
			return nil, fmt.Errorf("unable to parse dockerfile %s: %w", relDockerfilePath, err)
		}

		return mapDockerfileToImagesSets(ctx, d)
	}

	img, err := mapLegacyDockerfileToImage(ctx, dockerfileImageConfig, opts)
	if err != nil {
		return nil, err
	}

	var ret ImagesSets

	ret = append(ret, []*Image{img})

	return ret, nil
}

func mapDockerfileToImagesSets(ctx context.Context, cfg *dockerfile.Dockerfile) (ImagesSets, error) {
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

	return ret, nil
}

func mapLegacyDockerfileToImage(ctx context.Context, dockerfileImageConfig *config.ImageFromDockerfile, opts CommonImageOptions) (*Image, error) {
	img := NewImage(dockerfileImageConfig.Name, ImageOptions{
		CommonImageOptions: opts,
		IsDockerfileImage:  true,
	})

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

	relDockerfilePath := filepath.Join(dockerfileImageConfig.Context, dockerfileImageConfig.Dockerfile)
	dockerfileData, err := opts.GiterminismManager.FileReader().ReadDockerfile(ctx, relDockerfilePath)
	if err != nil {
		return nil, err
	}

	var relDockerignorePath string
	var dockerignorePatterns []string
	for _, relContextDockerignorePath := range []string{
		dockerfileImageConfig.Dockerfile + ".dockerignore",
		".dockerignore",
	} {
		relDockerignorePath = filepath.Join(dockerfileImageConfig.Context, relContextDockerignorePath)
		if exist, err := opts.GiterminismManager.FileReader().IsDockerignoreExistAnywhere(ctx, relDockerignorePath); err != nil {
			return nil, err
		} else if exist {
			dockerignoreData, err := opts.GiterminismManager.FileReader().ReadDockerignore(ctx, relDockerignorePath)
			if err != nil {
				return nil, err
			}

			r := bytes.NewReader(dockerignoreData)
			dockerignorePatterns, err = dockerignore.ReadAll(r)
			if err != nil {
				return nil, fmt.Errorf("unable to read %q file: %w", relContextDockerignorePath, err)
			}

			break
		}
	}

	dockerignorePathMatcher := path_matcher.NewPathMatcher(path_matcher.PathMatcherOptions{
		BasePath:             filepath.Join(opts.GiterminismManager.RelativeToGitProjectDir(), dockerfileImageConfig.Context),
		DockerignorePatterns: dockerignorePatterns,
	})

	if !dockerignorePathMatcher.IsPathMatched(relDockerfilePath) {
		exceptionRule := "!" + dockerfileImageConfig.Dockerfile
		logboek.Context(ctx).Warn().LogLn("WARNING: There is no way to ignore the Dockerfile due to docker limitation when building an image for a compressed context that reads from STDIN.")
		logboek.Context(ctx).Warn().LogF("WARNING: To hide this message, remove the Dockerfile ignore rule from the %q or add an exception rule %q.\n", relDockerignorePath, exceptionRule)

		dockerignorePatterns = append(dockerignorePatterns, exceptionRule)
		dockerignorePathMatcher = path_matcher.NewPathMatcher(path_matcher.PathMatcherOptions{
			BasePath:             filepath.Join(opts.GiterminismManager.RelativeToGitProjectDir(), dockerfileImageConfig.Context),
			DockerignorePatterns: dockerignorePatterns,
		})
	}

	p, err := parser.Parse(bytes.NewReader(dockerfileData))
	if err != nil {
		return nil, err
	}

	dockerStages, dockerMetaArgs, err := instructions.Parse(p.AST)
	if err != nil {
		return nil, err
	}

	dockerfile.ResolveDockerStagesFromValue(dockerStages)

	dockerTargetIndex, err := dockerfile.GetDockerTargetStageIndex(dockerStages, dockerfileImageConfig.Target)
	if err != nil {
		return nil, err
	}

	ds := stage.NewDockerStages(
		dockerStages,
		util.MapStringInterfaceToMapStringString(dockerfileImageConfig.Args),
		dockerMetaArgs,
		dockerTargetIndex,
	)

	baseStageOptions := &stage.NewBaseStageOptions{
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
		stage.NewContextChecksum(dockerignorePathMatcher),
		baseStageOptions,
		dockerfileImageConfig.Dependencies,
	)

	img.stages = append(img.stages, dockerfileStage)

	logboek.Context(ctx).Info().LogFDetails("Using stage %s\n", dockerfileStage.Name())

	return img, nil
}
