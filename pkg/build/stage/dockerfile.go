package stage

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/moby/buildkit/frontend/dockerfile/instructions"
	"github.com/moby/buildkit/frontend/dockerfile/parser"
	"github.com/moby/buildkit/frontend/dockerfile/shell"

	"github.com/werf/logboek"
	"github.com/werf/logboek/pkg/style"

	"github.com/werf/werf/pkg/container_runtime"
	"github.com/werf/werf/pkg/context_manager"
	"github.com/werf/werf/pkg/docker_registry"
	"github.com/werf/werf/pkg/git_repo"
	"github.com/werf/werf/pkg/giterminism_manager"
	"github.com/werf/werf/pkg/image"
	"github.com/werf/werf/pkg/path_matcher"
	"github.com/werf/werf/pkg/true_git/ls_tree"
	"github.com/werf/werf/pkg/util"
)

func GenerateDockerfileStage(dockerRunArgs *DockerRunArgs, dockerStages *DockerStages, contextChecksum *ContextChecksum, baseStageOptions *NewBaseStageOptions) *DockerfileStage {
	return newDockerfileStage(dockerRunArgs, dockerStages, contextChecksum, baseStageOptions)
}

func newDockerfileStage(dockerRunArgs *DockerRunArgs, dockerStages *DockerStages, contextChecksum *ContextChecksum, baseStageOptions *NewBaseStageOptions) *DockerfileStage {
	s := &DockerfileStage{}
	s.DockerRunArgs = dockerRunArgs
	s.DockerStages = dockerStages
	s.ContextChecksum = contextChecksum
	s.BaseStage = newBaseStage(Dockerfile, baseStageOptions)

	return s
}

type DockerfileStage struct {
	*DockerRunArgs
	*DockerStages
	*ContextChecksum
	*BaseStage
}

func NewDockerRunArgs(dockerfilePath, target, context string, contextAddFile []string, buildArgs map[string]interface{}, addHost []string, network, ssh string) *DockerRunArgs {
	return &DockerRunArgs{
		dockerfilePath: dockerfilePath,
		target:         target,
		context:        context,
		contextAddFile: contextAddFile,
		buildArgs:      buildArgs,
		addHost:        addHost,
		network:        network,
		ssh:            ssh,
	}
}

type DockerRunArgs struct {
	dockerfilePath string
	target         string
	context        string
	contextAddFile []string
	buildArgs      map[string]interface{}
	addHost        []string
	network        string
	ssh            string
}

func (d *DockerRunArgs) contextAddFileRelativeToProject() []string {
	var result []string
	for _, addFile := range d.contextAddFile {
		result = append(result, filepath.Join(d.context, addFile))
	}

	return result
}

type DockerStages struct {
	dockerStages           []instructions.Stage
	dockerTargetStageIndex int
	dockerBuildArgsHash    map[string]string
	dockerMetaArgsHash     map[string]string
	dockerStageArgsHash    map[int]map[string]string
	dockerStageEnvs        map[int]map[string]string

	imageOnBuildInstructions map[string][]string
}

func NewDockerStages(dockerStages []instructions.Stage, dockerBuildArgsHash map[string]string, dockerMetaArgs []instructions.ArgCommand, dockerTargetStageIndex int) (*DockerStages, error) {
	ds := &DockerStages{
		dockerStages:             dockerStages,
		dockerTargetStageIndex:   dockerTargetStageIndex,
		dockerBuildArgsHash:      dockerBuildArgsHash,
		dockerStageArgsHash:      map[int]map[string]string{},
		dockerStageEnvs:          map[int]map[string]string{},
		imageOnBuildInstructions: map[string][]string{},
	}

	ds.dockerMetaArgsHash = map[string]string{}
	for _, arg := range dockerMetaArgs {
		if _, _, err := ds.addDockerMetaArg(arg.Key, arg.ValueString()); err != nil {
			return nil, err
		}
	}

	return ds, nil
}

// addDockerMetaArg function sets --build-arg value or resolved meta ARG value
func (ds *DockerStages) addDockerMetaArg(key, value string) (string, string, error) {
	resolvedKey, err := ds.ShlexProcessWordWithMetaArgs(key)
	if err != nil {
		return "", "", err
	}

	var resolvedValue string
	if buildArgValue, ok := ds.dockerBuildArgsHash[resolvedKey]; ok {
		resolvedValue = buildArgValue
	} else {
		rValue, err := ds.ShlexProcessWordWithMetaArgs(value)
		if err != nil {
			return "", "", err
		}

		resolvedValue = rValue
	}

	ds.dockerMetaArgsHash[resolvedKey] = resolvedValue
	return resolvedKey, resolvedValue, err
}

// AddDockerStageArg function sets --build-arg value or resolved dockerfile stage ARG value or resolved meta ARG value (if stage ARG value is empty)
func (ds *DockerStages) AddDockerStageArg(dockerStageID int, key, value string) (string, string, error) {
	resolvedKey, err := ds.ShlexProcessWordWithStageArgsAndEnvs(dockerStageID, key)
	if err != nil {
		return "", "", err
	}

	var resolvedValue string
	if buildArgValue, ok := ds.dockerBuildArgsHash[resolvedKey]; ok {
		resolvedValue = buildArgValue
	} else if value == "" {
		resolvedValue = ds.dockerMetaArgsHash[resolvedKey]
	} else {
		rValue, err := ds.ShlexProcessWordWithStageArgsAndEnvs(dockerStageID, value)
		if err != nil {
			return "", "", err
		}

		resolvedValue = rValue
	}

	ds.DockerStageArgsHash(dockerStageID)[resolvedKey] = resolvedValue
	return resolvedKey, resolvedValue, nil
}

func (ds *DockerStages) AddDockerStageEnv(dockerStageID int, key, value string) (string, string, error) {
	resolvedKey, err := ds.ShlexProcessWordWithStageArgsAndEnvs(dockerStageID, key)
	if err != nil {
		return "", "", err
	}

	resolvedValue, err := ds.ShlexProcessWordWithStageArgsAndEnvs(dockerStageID, value)
	if err != nil {
		return "", "", err
	}

	ds.DockerStageEnvs(dockerStageID)[resolvedKey] = resolvedValue
	return resolvedKey, resolvedValue, nil
}

func (ds *DockerStages) ShlexProcessWordWithMetaArgs(value string) (string, error) {
	return shlexProcessWord(value, toArgsArray(ds.dockerMetaArgsHash))
}

func (ds *DockerStages) ShlexProcessWordWithStageArgsAndEnvs(dockerStageID int, value string) (string, error) {
	return shlexProcessWord(value, toArgsArray(ds.DockerStageArgsHash(dockerStageID), ds.DockerStageEnvs(dockerStageID)))
}

func (ds *DockerStages) ShlexProcessWordWithStageEnvs(dockerStageID int, value string) (string, error) {
	return shlexProcessWord(value, toArgsArray(ds.DockerStageEnvs(dockerStageID)))
}

func (ds *DockerStages) DockerStageArgsHash(dockerStageID int) map[string]string {
	_, ok := ds.dockerStageArgsHash[dockerStageID]
	if !ok {
		ds.dockerStageArgsHash[dockerStageID] = map[string]string{}
	}

	return ds.dockerStageArgsHash[dockerStageID]
}

func (ds *DockerStages) DockerStageEnvs(dockerStageID int) map[string]string {
	_, ok := ds.dockerStageEnvs[dockerStageID]
	if !ok {
		ds.dockerStageEnvs[dockerStageID] = map[string]string{}
	}

	return ds.dockerStageEnvs[dockerStageID]
}

func toArgsArray(argsHashes ...map[string]string) []string {
	var argsArray []string

	isAddedKey := map[string]bool{}
	for i := len(argsHashes) - 1; i >= 0; i-- {
		for _, argsHash := range argsHashes {
			for key, value := range argsHash {
				if _, ok := isAddedKey[key]; ok {
					continue
				}

				argsArray = append(argsArray, fmt.Sprintf("%s=%s", key, value))
				isAddedKey[key] = true
			}
		}
	}

	return argsArray
}

func shlexProcessWord(value string, argsArray []string) (string, error) {
	shlex := shell.NewLex(parser.DefaultEscapeToken)
	resolvedValue, err := shlex.ProcessWord(value, argsArray)
	if err != nil {
		return "", err
	}

	return resolvedValue, nil
}

func NewContextChecksum(dockerignorePathMatcher *path_matcher.DockerfileIgnorePathMatcher) *ContextChecksum {
	return &ContextChecksum{
		dockerignorePathMatcher: dockerignorePathMatcher,
	}
}

type ContextChecksum struct {
	dockerignorePathMatcher *path_matcher.DockerfileIgnorePathMatcher

	mainLsTreeResult *ls_tree.Result
}

type dockerfileInstructionInterface interface {
	String() string
	Name() string
}

func (s *DockerfileStage) FetchDependencies(ctx context.Context, _ Conveyor, cr container_runtime.ContainerRuntime) error {
	containerRuntime := cr.(*container_runtime.LocalDockerServerRuntime)

outerLoop:
	for ind, stage := range s.dockerStages {
		for relatedStageIndex, relatedStage := range s.dockerStages {
			if ind == relatedStageIndex {
				continue
			}

			if stage.BaseName == relatedStage.Name {
				continue outerLoop
			}
		}

		resolvedBaseName, err := s.ShlexProcessWordWithMetaArgs(stage.BaseName)
		if err != nil {
			return err
		}

		_, ok := s.imageOnBuildInstructions[resolvedBaseName]
		if ok || resolvedBaseName == "scratch" {
			continue
		}

		getBaseImageOnBuildLocally := func() ([]string, error) {
			inspect, err := containerRuntime.GetImageInspect(ctx, resolvedBaseName)
			if err != nil {
				return nil, err
			}

			if inspect == nil {
				return nil, imageNotExistLocally
			}

			return inspect.Config.OnBuild, nil
		}

		getBaseImageOnBuildRemotely := func() ([]string, error) {
			configFile, err := docker_registry.API().GetRepoImageConfigFile(ctx, resolvedBaseName)
			if err != nil {
				return nil, fmt.Errorf("get repo image %s config file failed: %s", resolvedBaseName, err)
			}

			return configFile.Config.OnBuild, nil
		}

		var onBuild []string
		if onBuild, err = getBaseImageOnBuildLocally(); err != nil && err != imageNotExistLocally {
			return err
		} else if err == imageNotExistLocally {
			var getRemotelyErr error
			if onBuild, getRemotelyErr = getBaseImageOnBuildRemotely(); getRemotelyErr != nil {
				if isUnsupportedMediaTypeError(getRemotelyErr) {
					logboek.Context(ctx).Warn().LogF("WARNING: Could not get base image manifest from local docker and from docker registry: %s\n", getRemotelyErr)
					logboek.Context(ctx).Warn().LogLn("WARNING: The base image pulling is necessary for calculating digest of image correctly\n")
					if err := logboek.Context(ctx).Default().LogProcess("Pulling base image %s", resolvedBaseName).DoError(func() error {
						return containerRuntime.PullImage(ctx, resolvedBaseName)
					}); err != nil {
						return err
					}

					if onBuild, err = getBaseImageOnBuildLocally(); err != nil {
						return err
					}
				} else {
					return getRemotelyErr
				}
			}
		}

		s.imageOnBuildInstructions[resolvedBaseName] = onBuild
	}

	return nil
}

func isUnsupportedMediaTypeError(err error) bool {
	return strings.Contains(err.Error(), "unsupported MediaType")
}

var imageNotExistLocally = errors.New("IMAGE_NOT_EXIST_LOCALLY")

func (s *DockerfileStage) GetDependencies(ctx context.Context, c Conveyor, _, _ container_runtime.ImageInterface) (string, error) {
	var stagesDependencies [][]string
	var stagesOnBuildDependencies [][]string

	for ind, stage := range s.dockerStages {
		var dependencies []string
		var onBuildDependencies []string

		dependencies = append(dependencies, s.addHost...)

		resolvedBaseName, err := s.ShlexProcessWordWithMetaArgs(stage.BaseName)
		if err != nil {
			return "", err
		}

		dependencies = append(dependencies, resolvedBaseName)

		onBuildInstructions, ok := s.imageOnBuildInstructions[resolvedBaseName]
		if ok {
			for _, instruction := range onBuildInstructions {
				_, iOnBuildDependencies, err := s.dockerfileOnBuildInstructionDependencies(ctx, c.GiterminismManager(), ind, instruction, true)
				if err != nil {
					return "", err
				}

				dependencies = append(dependencies, iOnBuildDependencies...)
			}
		}

		for _, cmd := range stage.Commands {
			cmdDependencies, cmdOnBuildDependencies, err := s.dockerfileInstructionDependencies(ctx, c.GiterminismManager(), ind, cmd, false, false)
			if err != nil {
				return "", err
			}

			dependencies = append(dependencies, cmdDependencies...)
			onBuildDependencies = append(onBuildDependencies, cmdOnBuildDependencies...)
		}

		stagesDependencies = append(stagesDependencies, dependencies)
		stagesOnBuildDependencies = append(stagesOnBuildDependencies, onBuildDependencies)
	}

	for ind, stage := range s.dockerStages {
		for relatedStageIndex, relatedStage := range s.dockerStages {
			if ind == relatedStageIndex {
				continue
			}

			if stage.BaseName == relatedStage.Name {
				stagesDependencies[ind] = append(stagesDependencies[ind], stagesDependencies[relatedStageIndex]...)
				stagesDependencies[ind] = append(stagesDependencies[ind], stagesOnBuildDependencies[relatedStageIndex]...)
			}
		}

		for _, cmd := range stage.Commands {
			switch c := cmd.(type) {
			case *instructions.CopyCommand:
				if c.From != "" {
					relatedStageIndex, err := strconv.Atoi(c.From)
					if err == nil && relatedStageIndex < len(stagesDependencies) {
						stagesDependencies[ind] = append(stagesDependencies[ind], stagesDependencies[relatedStageIndex]...)
					}
				}
			}
		}
	}

	dockerfileStageDependencies := stagesDependencies[s.dockerTargetStageIndex]

	if dockerfileStageDependenciesDebug() {
		logboek.Context(ctx).LogLn(dockerfileStageDependencies)
	}

	return util.Sha256Hash(dockerfileStageDependencies...), nil
}

func (s *DockerfileStage) dockerfileInstructionDependencies(ctx context.Context, giterminismManager giterminism_manager.Interface, dockerStageID int, cmd interface{}, isOnbuildInstruction bool, isBaseImageOnbuildInstruction bool) ([]string, []string, error) {
	var dependencies []string
	var onBuildDependencies []string

	resolveValueFunc := func(value string) (string, error) {
		if isBaseImageOnbuildInstruction {
			return value, nil
		}

		var shlexProcessWordFunc func(int, string) (string, error)
		if isOnbuildInstruction {
			shlexProcessWordFunc = s.ShlexProcessWordWithStageEnvs
		} else {
			shlexProcessWordFunc = s.ShlexProcessWordWithStageArgsAndEnvs
		}

		resolvedValue, err := shlexProcessWordFunc(dockerStageID, value)
		if err != nil {
			return "", err
		}

		return resolvedValue, nil
	}

	resolveKeyAndValueFunc := func(key, value string) (string, string, error) {
		resolvedKey, err := resolveValueFunc(key)
		if err != nil {
			return "", "", err
		}

		resolvedValue, err := resolveValueFunc(value)
		if err != nil {
			return "", "", err
		}

		return resolvedKey, resolvedValue, nil
	}

	processArgFunc := func(key, value string) (string, string, error) {
		var resolvedKey, resolvedValue string
		var err error
		if !isOnbuildInstruction {
			resolvedKey, resolvedValue, err = s.AddDockerStageArg(dockerStageID, key, value)
			if err != nil {
				return "", "", err
			}
		} else {
			resolvedKey, resolvedValue, err = resolveKeyAndValueFunc(key, value)
			if err != nil {
				return "", "", err
			}
		}

		return resolvedKey, resolvedValue, nil
	}

	processEnvFunc := func(key, value string) (string, string, error) {
		var resolvedKey, resolvedValue string
		var err error
		if !isOnbuildInstruction {
			resolvedKey, resolvedValue, err = s.AddDockerStageEnv(dockerStageID, key, value)
			if err != nil {
				return "", "", err
			}
		} else {
			resolvedKey, resolvedValue, err = resolveKeyAndValueFunc(key, value)
			if err != nil {
				return "", "", err
			}
		}

		return resolvedKey, resolvedValue, nil
	}

	resolveSourcesFunc := func(sources []string) ([]string, error) {
		var resolvedSources []string
		for _, source := range sources {
			resolvedSource, err := resolveValueFunc(source)
			if err != nil {
				return nil, err
			}

			resolvedSources = append(resolvedSources, resolvedSource)
		}

		return resolvedSources, nil
	}

	switch c := cmd.(type) {
	case *instructions.ArgCommand:
		resolvedKey, resolvedValue, err := processArgFunc(c.Key, c.ValueString())
		if err != nil {
			return nil, nil, err
		}

		dependencies = append(dependencies, fmt.Sprintf("ARG %s=%s", resolvedKey, resolvedValue))
	case *instructions.EnvCommand:
		for _, keyValuePair := range c.Env {
			resolvedKey, resolvedValue, err := processEnvFunc(keyValuePair.Key, keyValuePair.Value)
			if err != nil {
				return nil, nil, err
			}

			dependencies = append(dependencies, fmt.Sprintf("ENV %s=%s", resolvedKey, resolvedValue))
		}
	case *instructions.AddCommand:
		dependencies = append(dependencies, c.String())

		resolvedSources, err := resolveSourcesFunc(c.SourcesAndDest.Sources())
		if err != nil {
			return nil, nil, err
		}

		checksum, err := s.calculateFilesChecksum(ctx, giterminismManager, resolvedSources, c.String())
		if err != nil {
			return nil, nil, err
		}
		dependencies = append(dependencies, checksum)
	case *instructions.CopyCommand:
		dependencies = append(dependencies, c.String())
		if c.From == "" {
			resolvedSources, err := resolveSourcesFunc(c.SourcesAndDest.Sources())
			if err != nil {
				return nil, nil, err
			}

			checksum, err := s.calculateFilesChecksum(ctx, giterminismManager, resolvedSources, c.String())
			if err != nil {
				return nil, nil, err
			}
			dependencies = append(dependencies, checksum)
		}
	case *instructions.OnbuildCommand:
		cDependencies, cOnBuildDependencies, err := s.dockerfileOnBuildInstructionDependencies(ctx, giterminismManager, dockerStageID, c.Expression, false)
		if err != nil {
			return nil, nil, err
		}

		dependencies = append(dependencies, cDependencies...)
		onBuildDependencies = append(onBuildDependencies, cOnBuildDependencies...)
	case dockerfileInstructionInterface:
		resolvedValue, err := resolveValueFunc(c.String())
		if err != nil {
			return nil, nil, err
		}

		dependencies = append(dependencies, resolvedValue)
	default:
		panic("runtime error")
	}

	return dependencies, onBuildDependencies, nil
}

func (s *DockerfileStage) dockerfileOnBuildInstructionDependencies(ctx context.Context, giterminismManager giterminism_manager.Interface, dockerStageID int, expression string, isBaseImageOnbuildInstruction bool) ([]string, []string, error) {
	p, err := parser.Parse(bytes.NewReader([]byte(expression)))
	if err != nil {
		return nil, nil, err
	}

	if len(p.AST.Children) != 1 {
		panic(fmt.Sprintf("unexpected condition: %s (%d children)", expression, len(p.AST.Children)))
	}

	instruction := p.AST.Children[0]
	cmd, err := instructions.ParseInstruction(instruction)
	if err != nil {
		return nil, nil, err
	}

	onBuildDependencies, _, err := s.dockerfileInstructionDependencies(ctx, giterminismManager, dockerStageID, cmd, true, isBaseImageOnbuildInstruction)
	if err != nil {
		return nil, nil, err
	}

	return []string{expression}, onBuildDependencies, nil
}

func (s *DockerfileStage) PrepareImage(ctx context.Context, c Conveyor, _, img container_runtime.ImageInterface) error {
	archivePath, err := s.prepareContextArchive(ctx, c.GiterminismManager())
	if err != nil {
		return err
	}

	img.DockerfileImageBuilder().AppendBuildArgs(s.DockerBuildArgs()...)
	img.DockerfileImageBuilder().AppendBuildArgs(fmt.Sprintf("--label=%s=%s", image.WerfProjectRepoCommitLabel, c.GiterminismManager().HeadCommit()))
	img.DockerfileImageBuilder().SetFilePathToStdin(archivePath)

	if c.GiterminismManager().Dev() {
		img.DockerfileImageBuilder().AppendBuildArgs(fmt.Sprintf("--label=%s=true", image.WerfDevLabel))
	}

	return nil
}

func (s *DockerfileStage) prepareContextArchive(ctx context.Context, giterminismManager giterminism_manager.Interface) (string, error) {
	archive, err := giterminismManager.LocalGitRepo().CreateArchive(ctx, git_repo.ArchiveOptions{
		FilterOptions: git_repo.FilterOptions{
			BasePath: s.context,
		},
		Commit: giterminismManager.HeadCommit(),
	})
	if err != nil {
		return "", fmt.Errorf("unable to create archive: %s", err)
	}

	archivePath := archive.GetFilePath()
	if len(s.contextAddFile) != 0 {
		if err := logboek.Context(ctx).Debug().LogProcess("Add contextAddFile to build context archive %s", archivePath).DoError(func() error {
			var sourceArchivePath = archivePath
			destinationArchivePath, err := context_manager.AddContextAddFileToContextArchive(ctx, sourceArchivePath, giterminismManager.ProjectDir(), s.context, s.contextAddFile)
			if err != nil {
				return err
			}

			archivePath = destinationArchivePath
			return nil
		}); err != nil {
			return "", err
		}
	}

	return archivePath, nil
}

func (s *DockerfileStage) DockerBuildArgs() []string {
	var result []string

	if s.dockerfilePath != "" {
		result = append(result, fmt.Sprintf("--file=%s", s.dockerfilePath))
	}

	if s.target != "" {
		result = append(result, fmt.Sprintf("--target=%s", s.target))
	}

	if len(s.buildArgs) != 0 {
		for key, value := range s.buildArgs {
			result = append(result, fmt.Sprintf("--build-arg=%s=%v", key, value))
		}
	}

	for _, addHost := range s.addHost {
		result = append(result, fmt.Sprintf("--add-host=%s", addHost))
	}

	if s.network != "" {
		result = append(result, fmt.Sprintf("--network=%s", s.network))
	}

	if s.ssh != "" {
		result = append(result, fmt.Sprintf("--ssh=%s", s.ssh))
	}

	return result
}

func (s *DockerfileStage) calculateFilesChecksum(ctx context.Context, giterminismManager giterminism_manager.Interface, wildcards []string, dockerfileLine string) (string, error) {
	var checksum string
	var err error

	normalizedWildcards := normalizeCopyAddSources(wildcards)

	logProcess := logboek.Context(ctx).Debug().LogProcess("Calculating files checksum (%v) from local git repo", normalizedWildcards)
	logProcess.Start()

	checksum, err = s.calculateFilesChecksumWithGit(ctx, giterminismManager, normalizedWildcards, dockerfileLine)
	if err != nil {
		logProcess.Fail()
		return "", err
	} else {
		logProcess.End()
	}

	if len(s.contextAddFile) > 0 {
		logProcess = logboek.Context(ctx).Debug().LogProcess("Calculating contextAddFile checksum")
		logProcess.Start()

		wildcardsPathMatcher := path_matcher.NewSimplePathMatcher(s.dockerignorePathMatcher.BaseFilepath(), wildcards, false)
		if contextAddChecksum, err := context_manager.ContextAddFileChecksum(ctx, giterminismManager.ProjectDir(), s.context, s.contextAddFile, wildcardsPathMatcher); err != nil {
			logProcess.Fail()
			return "", fmt.Errorf("unable to calculate checksum for contextAddFile files list: %s", err)
		} else {
			if contextAddChecksum != "" {
				logboek.Context(ctx).Debug().LogLn()
				logboek.Context(ctx).Debug().LogLn(contextAddChecksum)
				checksum = util.Sha256Hash(checksum, contextAddChecksum)
			}

			logProcess.End()
		}
	}

	logboek.Context(ctx).Debug().LogF("Result checksum: %s\n", checksum)
	logboek.Context(ctx).Debug().LogOptionalLn()

	return checksum, nil
}

func (s *DockerfileStage) calculateFilesChecksumWithGit(ctx context.Context, giterminismManager giterminism_manager.Interface, wildcards []string, dockerfileLine string) (string, error) {
	if s.mainLsTreeResult == nil {
		logProcess := logboek.Context(ctx).Debug().LogProcess("ls-tree (%s)", s.dockerignorePathMatcher.String())
		logProcess.Start()
		result, err := giterminismManager.LocalGitRepo().LsTree(ctx, s.dockerignorePathMatcher, git_repo.LsTreeOptions{UseHeadCommit: true})
		if err != nil {
			if err.Error() == "entry not found" {
				logboek.Context(ctx).Debug().LogFWithCustomStyle(
					style.Get(style.FailName),
					"Entry %s is not found\n",
					s.dockerignorePathMatcher.BaseFilepath(),
				)
				logProcess.End()
				goto entryNotFoundInGitRepository
			}

			logProcess.Fail()
			return "", err
		} else {
			logProcess.End()
		}

		s.mainLsTreeResult = result
	}

entryNotFoundInGitRepository:
	wildcardsPathMatcher := path_matcher.NewSimplePathMatcher(s.dockerignorePathMatcher.BaseFilepath(), wildcards, false)

	var lsTreeResultChecksum string
	if s.mainLsTreeResult != nil {
		logProcess := logboek.Context(ctx).Debug().LogProcess("ls-tree (%s)", wildcardsPathMatcher.String())
		logProcess.Start()
		lsTreeResult, err := s.mainLsTreeResult.LsTree(ctx, wildcardsPathMatcher)
		if err != nil {
			logProcess.Fail()
			return "", err
		} else {
			logProcess.End()
		}

		if !lsTreeResult.IsEmpty() {
			logboek.Context(ctx).Debug().LogBlock("ls-tree result checksum (%s)", wildcardsPathMatcher.String()).Do(func() {
				lsTreeResultChecksum = lsTreeResult.Checksum(ctx)
				logboek.Context(ctx).Debug().LogOptionalLn()
				logboek.Context(ctx).Debug().LogLn(lsTreeResultChecksum)
			})
		}
	}

	if err := giterminismManager.Inspector().InspectBuildContextFiles(ctx, path_matcher.NewMultiPathMatcher(
		s.dockerignorePathMatcher,
		wildcardsPathMatcher,
		path_matcher.NewGitMappingPathMatcher("", []string{}, s.contextAddFileRelativeToProject(), true),
	)); err != nil {
		return "", err
	}

	return util.Sha256Hash(lsTreeResultChecksum), nil
}

func normalizeCopyAddSources(wildcards []string) []string {
	var result []string
	for _, wildcard := range wildcards {
		normalizedWildcard := path.Clean(wildcard)
		if normalizedWildcard == "/" {
			normalizedWildcard = "."
		} else if strings.HasPrefix(normalizedWildcard, "/") {
			normalizedWildcard = strings.TrimPrefix(normalizedWildcard, "/")
		}

		result = append(result, normalizedWildcard)
	}

	return result
}

func dockerfileStageDependenciesDebug() bool {
	return os.Getenv("WERF_DEBUG_DOCKERFILE_STAGE_DEPENDENCIES") == "1"
}
