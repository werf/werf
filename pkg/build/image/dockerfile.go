package image

import (
	"bytes"
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/docker/docker/builder/dockerignore"
	"github.com/moby/buildkit/frontend/dockerfile/instructions"
	"github.com/moby/buildkit/frontend/dockerfile/parser"

	"github.com/werf/logboek"
	"github.com/werf/werf/pkg/build/stage"
	stage_instruction "github.com/werf/werf/pkg/build/stage/instruction"
	"github.com/werf/werf/pkg/config"
	"github.com/werf/werf/pkg/dockerfile"
	"github.com/werf/werf/pkg/dockerfile/frontend"
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

		dockerfileID := util.Sha256Hash(filepath.Clean(relDockerfilePath))

		d, err := frontend.ParseDockerfileWithBuildkit(dockerfileID, dockerfileData, dockerfileImageConfig.Name, dockerfile.DockerfileOptions{
			Target:               dockerfileImageConfig.Target,
			BuildArgs:            util.MapStringInterfaceToMapStringString(dockerfileImageConfig.Args),
			AddHost:              dockerfileImageConfig.AddHost,
			Network:              dockerfileImageConfig.Network,
			SSH:                  dockerfileImageConfig.SSH,
			DependenciesArgsKeys: stage.GetDependenciesArgsKeys(dockerfileImageConfig.Dependencies),
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
		WerfImageName string
		Stage         *dockerfile.DockerfileStage
		Level         int
		IsTargetStage bool
	}{
		{WerfImageName: dockerfileImageConfig.Name, Stage: targetStage, Level: 0, IsTargetStage: true},
	}

	appendQueue := func(werfImageName string, stage *dockerfile.DockerfileStage, level int) {
		queue = append(queue, struct {
			WerfImageName string
			Stage         *dockerfile.DockerfileStage
			Level         int
			IsTargetStage bool
		}{WerfImageName: werfImageName, Stage: stage, Level: level})
	}

	for len(queue) > 0 {
		item := queue[0]
		queue = queue[1:]

		appendImageToCurrentSet := func(newImg *Image) {
			if item.Level == len(ret) {
				ret = append([][]*Image{nil}, ret...)
			}
			for _, img := range ret[len(ret)-item.Level-1] {
				if img.Name == newImg.Name {
					return
				}
			}
			ret[len(ret)-item.Level-1] = append(ret[len(ret)-item.Level-1], newImg)
		}

		stg := item.Stage

		var img *Image
		var err error
		if baseStg := cfg.FindStage(stg.BaseName); baseStg != nil {
			img, err = NewImage(ctx, item.WerfImageName, StageAsBaseImage, ImageOptions{
				IsDockerfileImage:         true,
				IsDockerfileTargetStage:   item.IsTargetStage,
				DockerfileImageConfig:     dockerfileImageConfig,
				CommonImageOptions:        opts,
				BaseImageName:             baseStg.GetWerfImageName(),
				DockerfileExpanderFactory: stg.ExpanderFactory,
			})
			if err != nil {
				return nil, fmt.Errorf("unable to map stage %s to werf image %q: %w", stg.LogName(), dockerfileImageConfig.Name, err)
			}

			appendQueue(baseStg.GetWerfImageName(), baseStg, item.Level+1)
		} else {
			img, err = NewImage(ctx, item.WerfImageName, ImageFromRegistryAsBaseImage, ImageOptions{
				IsDockerfileImage:         true,
				IsDockerfileTargetStage:   item.IsTargetStage,
				DockerfileImageConfig:     dockerfileImageConfig,
				CommonImageOptions:        opts,
				BaseImageReference:        stg.BaseName,
				DockerfileExpanderFactory: stg.ExpanderFactory,
			})
			if err != nil {
				return nil, fmt.Errorf("unable to map stage %s to werf image %q: %w", stg.LogName(), dockerfileImageConfig.Name, err)
			}
		}

		for ind, instr := range stg.Instructions {
			stageLogName := fmt.Sprintf("%s%d", strings.ToUpper(instr.GetInstructionData().Name()), ind+1)
			isFirstStage := (len(img.stages) == 0)
			baseStageOptions := &stage.BaseStageOptions{
				ImageName:        img.Name,
				LogName:          stageLogName,
				ImageTmpDir:      img.TmpDir,
				ContainerWerfDir: img.ContainerWerfDir,
				ProjectName:      opts.ProjectName,
			}

			var stg stage.Interface
			switch typedInstr := any(instr).(type) {
			case *dockerfile.DockerfileStageInstruction[*instructions.ArgCommand]:
				continue
			case *dockerfile.DockerfileStageInstruction[*instructions.AddCommand]:
				stg = stage_instruction.NewAdd(typedInstr, dockerfileImageConfig.Dependencies, !isFirstStage, baseStageOptions)
			case *dockerfile.DockerfileStageInstruction[*instructions.CmdCommand]:
				stg = stage_instruction.NewCmd(typedInstr, dockerfileImageConfig.Dependencies, !isFirstStage, baseStageOptions)
			case *dockerfile.DockerfileStageInstruction[*instructions.CopyCommand]:
				stg = stage_instruction.NewCopy(typedInstr, dockerfileImageConfig.Dependencies, !isFirstStage, baseStageOptions)
			case *dockerfile.DockerfileStageInstruction[*instructions.EntrypointCommand]:
				stg = stage_instruction.NewEntrypoint(typedInstr, dockerfileImageConfig.Dependencies, !isFirstStage, baseStageOptions)
			case *dockerfile.DockerfileStageInstruction[*instructions.EnvCommand]:
				stg = stage_instruction.NewEnv(typedInstr, dockerfileImageConfig.Dependencies, !isFirstStage, baseStageOptions)
			case *dockerfile.DockerfileStageInstruction[*instructions.ExposeCommand]:
				stg = stage_instruction.NewExpose(typedInstr, dockerfileImageConfig.Dependencies, !isFirstStage, baseStageOptions)
			case *dockerfile.DockerfileStageInstruction[*instructions.HealthCheckCommand]:
				stg = stage_instruction.NewHealthcheck(typedInstr, dockerfileImageConfig.Dependencies, !isFirstStage, baseStageOptions)
			case *dockerfile.DockerfileStageInstruction[*instructions.LabelCommand]:
				stg = stage_instruction.NewLabel(typedInstr, dockerfileImageConfig.Dependencies, !isFirstStage, baseStageOptions)
			case *dockerfile.DockerfileStageInstruction[*instructions.MaintainerCommand]:
				stg = stage_instruction.NewMaintainer(typedInstr, dockerfileImageConfig.Dependencies, !isFirstStage, baseStageOptions)
			case *dockerfile.DockerfileStageInstruction[*instructions.OnbuildCommand]:
				stg = stage_instruction.NewOnBuild(typedInstr, dockerfileImageConfig.Dependencies, !isFirstStage, baseStageOptions)
			case *dockerfile.DockerfileStageInstruction[*instructions.RunCommand]:
				stg = stage_instruction.NewRun(typedInstr, dockerfileImageConfig.Dependencies, !isFirstStage, baseStageOptions)
			case *dockerfile.DockerfileStageInstruction[*instructions.ShellCommand]:
				stg = stage_instruction.NewShell(typedInstr, dockerfileImageConfig.Dependencies, !isFirstStage, baseStageOptions)
			case *dockerfile.DockerfileStageInstruction[*instructions.StopSignalCommand]:
				stg = stage_instruction.NewStopSignal(typedInstr, dockerfileImageConfig.Dependencies, !isFirstStage, baseStageOptions)
			case *dockerfile.DockerfileStageInstruction[*instructions.UserCommand]:
				stg = stage_instruction.NewUser(typedInstr, dockerfileImageConfig.Dependencies, !isFirstStage, baseStageOptions)
			case *dockerfile.DockerfileStageInstruction[*instructions.VolumeCommand]:
				stg = stage_instruction.NewVolume(typedInstr, dockerfileImageConfig.Dependencies, !isFirstStage, baseStageOptions)
			case *dockerfile.DockerfileStageInstruction[*instructions.WorkdirCommand]:
				stg = stage_instruction.NewWorkdir(typedInstr, dockerfileImageConfig.Dependencies, !isFirstStage, baseStageOptions)
			default:
				panic(fmt.Sprintf("unsupported instruction type %#v", instr))
			}

			img.stages = append(img.stages, stg)

			for _, dep := range instr.GetDependenciesByStageRef() {
				appendQueue(dep.GetWerfImageName(), dep, item.Level+1)
			}
		}

		appendImageToCurrentSet(img)
	}

	return ret, nil
}

func mapLegacyDockerfileToImage(ctx context.Context, dockerfileImageConfig *config.ImageFromDockerfile, opts CommonImageOptions) (*Image, error) {
	img, err := NewImage(ctx, dockerfileImageConfig.Name, NoBaseImage, ImageOptions{
		CommonImageOptions:      opts,
		IsDockerfileImage:       true,
		IsDockerfileTargetStage: true,
		DockerfileImageConfig:   dockerfileImageConfig,
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
	var dockerIgnorePatterns []string
	for _, dockerIgnoreRelToContextPath := range []string{
		dockerfileRelToContextPath + ".dockerignore",
		".dockerignore",
	} {
		relDockerIgnorePath := filepath.Join(contextGitSubDir, dockerIgnoreRelToContextPath)
		if exist, err := giterminismMgr.FileReader().IsDockerignoreExistAnywhere(ctx, relDockerIgnorePath); err != nil {
			return nil, err
		} else if !exist {
			continue
		}

		dockerIgnore, err := giterminismMgr.FileReader().ReadDockerignore(ctx, relDockerIgnorePath)
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

	dockerfileRelToGitPath := filepath.Join(giterminismMgr.RelativeToGitProjectDir(), contextGitSubDir, dockerfileRelToContextPath)
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
