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

	targetStage, err := cfg.GetTargetStage()
	if err != nil {
		return nil, fmt.Errorf("unable to get target dockerfile stage: %w", err)
	}

	queue := []struct {
		Stage *dockerfile.DockerfileStage
		Level int
	}{
		{Stage: targetStage, Level: 0},
	}

	appendQueue := func(stage *dockerfile.DockerfileStage, level int) {
		queue = append(queue, struct {
			Stage *dockerfile.DockerfileStage
			Level int
		}{Stage: stage, Level: level})
	}

	for len(queue) > 0 {
		item := queue[0]
		queue = queue[1:]

		appendImageToCurrentSet := func(img *Image) {
			if item.Level == len(ret) {
				ret = append([][]*Image{nil}, ret...)
			}
			ret[len(ret)-item.Level-1] = append(ret[len(ret)-item.Level-1], img)
		}

		stg := item.Stage

		var img *Image
		var err error
		if baseStg := cfg.FindStage(stg.BaseName); baseStg != nil {
			img, err = NewImage(ctx, dockerfileImageConfig.Name, StageAsBaseImage, ImageOptions{
				IsDockerfileImage:     true,
				DockerfileImageConfig: dockerfileImageConfig,
				CommonImageOptions:    opts,
				BaseImageName:         baseStg.WerfImageName(),
			})
			if err != nil {
				return nil, fmt.Errorf("unable to map stage %s to werf image %q: %w", stg.LogName(), dockerfileImageConfig.Name, err)
			}

			appendQueue(baseStg, item.Level+1)
		} else {
			img, err = NewImage(ctx, dockerfileImageConfig.Name, ImageFromRegistryAsBaseImage, ImageOptions{
				IsDockerfileImage:     true,
				DockerfileImageConfig: dockerfileImageConfig,
				CommonImageOptions:    opts,
				BaseImageReference:    targetStage.BaseName,
			})
			if err != nil {
				return nil, fmt.Errorf("unable to map stage %s to werf image %q: %w", targetStage.LogName(), dockerfileImageConfig.Name, err)
			}
		}

		for ind, instr := range stg.Instructions {
			stageName := stage.StageName(fmt.Sprintf("%d-%s", ind, instr.GetInstructionData().Name()))
			isFirstStage := (len(img.stages) == 0)
			baseStageOptions := &stage.BaseStageOptions{
				ImageName:        img.Name,
				ImageTmpDir:      img.TmpDir,
				ContainerWerfDir: img.ContainerWerfDir,
				ProjectName:      opts.ProjectName,
			}

			var stg stage.Interface
			switch typedInstr := any(instr).(type) {
			case *dockerfile.DockerfileStageInstruction[*dockerfile_instruction.Add]:
				stg = stage_instruction.NewAdd(stageName, typedInstr, dockerfileImageConfig.Dependencies, !isFirstStage, baseStageOptions)
			case *dockerfile.DockerfileStageInstruction[*dockerfile_instruction.Cmd]:
				stg = stage_instruction.NewCmd(stageName, typedInstr, dockerfileImageConfig.Dependencies, !isFirstStage, baseStageOptions)
			case *dockerfile.DockerfileStageInstruction[*dockerfile_instruction.Copy]:
				stg = stage_instruction.NewCopy(stageName, typedInstr, dockerfileImageConfig.Dependencies, !isFirstStage, baseStageOptions)
			case *dockerfile.DockerfileStageInstruction[*dockerfile_instruction.Entrypoint]:
				stg = stage_instruction.NewEntrypoint(stageName, typedInstr, dockerfileImageConfig.Dependencies, !isFirstStage, baseStageOptions)
			case *dockerfile.DockerfileStageInstruction[*dockerfile_instruction.Env]:
				stg = stage_instruction.NewEnv(stageName, typedInstr, dockerfileImageConfig.Dependencies, !isFirstStage, baseStageOptions)
			case *dockerfile.DockerfileStageInstruction[*dockerfile_instruction.Expose]:
				stg = stage_instruction.NewExpose(stageName, typedInstr, dockerfileImageConfig.Dependencies, !isFirstStage, baseStageOptions)
			case *dockerfile.DockerfileStageInstruction[*dockerfile_instruction.Healthcheck]:
				stg = stage_instruction.NewHealthcheck(stageName, typedInstr, dockerfileImageConfig.Dependencies, !isFirstStage, baseStageOptions)
			case *dockerfile.DockerfileStageInstruction[*dockerfile_instruction.Label]:
				stg = stage_instruction.NewLabel(stageName, typedInstr, dockerfileImageConfig.Dependencies, !isFirstStage, baseStageOptions)
			case *dockerfile.DockerfileStageInstruction[*dockerfile_instruction.Maintainer]:
				stg = stage_instruction.NewMaintainer(stageName, typedInstr, dockerfileImageConfig.Dependencies, !isFirstStage, baseStageOptions)
			case *dockerfile.DockerfileStageInstruction[*dockerfile_instruction.OnBuild]:
				stg = stage_instruction.NewOnBuild(stageName, typedInstr, dockerfileImageConfig.Dependencies, !isFirstStage, baseStageOptions)
			case *dockerfile.DockerfileStageInstruction[*dockerfile_instruction.Run]:
				stg = stage_instruction.NewRun(stageName, typedInstr, dockerfileImageConfig.Dependencies, !isFirstStage, baseStageOptions)
			case *dockerfile.DockerfileStageInstruction[*dockerfile_instruction.Shell]:
				stg = stage_instruction.NewShell(stageName, typedInstr, dockerfileImageConfig.Dependencies, !isFirstStage, baseStageOptions)
			case *dockerfile.DockerfileStageInstruction[*dockerfile_instruction.StopSignal]:
				stg = stage_instruction.NewStopSignal(stageName, typedInstr, dockerfileImageConfig.Dependencies, !isFirstStage, baseStageOptions)
			case *dockerfile.DockerfileStageInstruction[*dockerfile_instruction.User]:
				stg = stage_instruction.NewUser(stageName, typedInstr, dockerfileImageConfig.Dependencies, !isFirstStage, baseStageOptions)
			case *dockerfile.DockerfileStageInstruction[*dockerfile_instruction.Volume]:
				stg = stage_instruction.NewVolume(stageName, typedInstr, dockerfileImageConfig.Dependencies, !isFirstStage, baseStageOptions)
			case *dockerfile.DockerfileStageInstruction[*dockerfile_instruction.Workdir]:
				stg = stage_instruction.NewWorkdir(stageName, typedInstr, dockerfileImageConfig.Dependencies, !isFirstStage, baseStageOptions)
			default:
				panic(fmt.Sprintf("unsupported instruction type %#v", instr))
			}

			img.stages = append(img.stages, stg)

			for _, dep := range instr.GetDependenciesByStageRef() {
				appendQueue(dep, item.Level+1)
			}
		}

		appendImageToCurrentSet(img)
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
