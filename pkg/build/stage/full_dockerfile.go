package stage

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/moby/buildkit/frontend/dockerfile/instructions"
	"github.com/moby/buildkit/frontend/dockerfile/parser"
	"github.com/moby/buildkit/frontend/dockerfile/shell"

	"github.com/werf/logboek"
	"github.com/werf/werf/pkg/config"
	"github.com/werf/werf/pkg/container_backend"
	"github.com/werf/werf/pkg/container_backend/stage_builder"
	"github.com/werf/werf/pkg/context_manager"
	"github.com/werf/werf/pkg/docker_registry"
	"github.com/werf/werf/pkg/dockerfile"
	"github.com/werf/werf/pkg/git_repo"
	"github.com/werf/werf/pkg/giterminism_manager"
	"github.com/werf/werf/pkg/image"
	"github.com/werf/werf/pkg/path_matcher"
	"github.com/werf/werf/pkg/util"
)

var ErrInvalidBaseImage = errors.New("invalid base image")

func IsErrInvalidBaseImage(err error) bool {
	return err != nil && errors.Is(err, ErrInvalidBaseImage)
}

func GenerateFullDockerfileStage(dockerRunArgs *DockerRunArgs, dockerStages *DockerStages, contextChecksum *ContextChecksum, baseStageOptions *BaseStageOptions, dependencies []*config.Dependency) *FullDockerfileStage {
	return newFullDockerfileStage(dockerRunArgs, dockerStages, contextChecksum, baseStageOptions, dependencies)
}

func newFullDockerfileStage(dockerRunArgs *DockerRunArgs, dockerStages *DockerStages, contextChecksum *ContextChecksum, baseStageOptions *BaseStageOptions, dependencies []*config.Dependency) *FullDockerfileStage {
	s := &FullDockerfileStage{}
	s.DockerRunArgs = dockerRunArgs
	s.DockerStages = dockerStages
	s.ContextChecksum = contextChecksum
	s.BaseStage = NewBaseStage(Dockerfile, baseStageOptions)
	s.dependencies = dependencies

	return s
}

type FullDockerfileStage struct {
	dependencies []*config.Dependency

	*DockerRunArgs
	*DockerStages
	*ContextChecksum
	*BaseStage
}

func NewDockerRunArgs(dockerfile []byte, dockerfilePath, target, context string, contextAddFiles []string, buildArgs map[string]interface{}, addHost []string, network, ssh string) *DockerRunArgs {
	return &DockerRunArgs{
		dockerfile:      dockerfile,
		dockerfilePath:  dockerfilePath,
		target:          target,
		context:         context,
		contextAddFiles: contextAddFiles,
		buildArgs:       buildArgs,
		addHost:         addHost,
		network:         network,
		ssh:             ssh,
	}
}

type DockerRunArgs struct {
	dockerfile      []byte
	dockerfilePath  string
	target          string
	context         string
	contextAddFiles []string
	buildArgs       map[string]interface{}
	addHost         []string
	network         string
	ssh             string
}

func (d *DockerRunArgs) contextRelativeToGitWorkTree(giterminismManager giterminism_manager.Interface) string {
	return filepath.Join(giterminismManager.RelativeToGitProjectDir(), d.context)
}

func (d *DockerRunArgs) contextAddFilesRelativeToGitWorkTree(giterminismManager giterminism_manager.Interface) []string {
	var result []string
	for _, addFile := range d.contextAddFiles {
		result = append(result, filepath.Join(d.contextRelativeToGitWorkTree(giterminismManager), addFile))
	}

	return result
}

type DockerStages struct {
	dockerStages           []instructions.Stage
	dockerTargetStageIndex int
	dockerBuildArgsHash    map[string]string
	dockerMetaArgs         []instructions.ArgCommand
	dockerStageArgsHash    map[int]map[string]string
	dockerStageEnvs        map[int]map[string]string

	imageOnBuildInstructions map[string][]string
}

func NewDockerStages(dockerStages []instructions.Stage, dockerBuildArgsHash map[string]string, dockerMetaArgs []instructions.ArgCommand, dockerTargetStageIndex int) *DockerStages {
	ds := &DockerStages{
		dockerStages:             dockerStages,
		dockerTargetStageIndex:   dockerTargetStageIndex,
		dockerBuildArgsHash:      dockerBuildArgsHash,
		dockerMetaArgs:           dockerMetaArgs,
		dockerStageArgsHash:      map[int]map[string]string{},
		dockerStageEnvs:          map[int]map[string]string{},
		imageOnBuildInstructions: map[string][]string{},
	}

	return ds
}

func (ds *DockerStages) resolveDockerMetaArgs(resolvedDependenciesArgsHash map[string]string) (map[string]string, error) {
	resolved := map[string]string{}

	for _, argInstruction := range ds.dockerMetaArgs {
		for _, keyValuePairOptional := range argInstruction.Args {
			key := keyValuePairOptional.Key

			var value string
			if keyValuePairOptional.Value != nil {
				value = *keyValuePairOptional.Value
			}

			resolvedKey, resolvedValue, err := ds.resolveDockerMetaArg(key, value, resolved, resolvedDependenciesArgsHash)
			if err != nil {
				return nil, fmt.Errorf("unable to resolve docker meta arg: %w", err)
			}

			resolved[resolvedKey] = resolvedValue
		}
	}

	return resolved, nil
}

// addDockerMetaArg function sets --build-arg value or resolved meta ARG value
func (ds *DockerStages) resolveDockerMetaArg(key, value string, resolvedDockerMetaArgsHash, resolvedDependenciesArgsHash map[string]string) (string, string, error) {
	resolvedKey, err := ds.ShlexProcessWordWithMetaArgs(key, resolvedDockerMetaArgsHash)
	if err != nil {
		return "", "", err
	}

	var resolvedValue string

	dependencyArgValue, ok := resolvedDependenciesArgsHash[resolvedKey]
	if ok {
		resolvedValue = dependencyArgValue
	} else {
		if buildArgValue, ok := ds.dockerBuildArgsHash[resolvedKey]; ok {
			resolvedValue = buildArgValue
		} else {
			rValue, err := ds.ShlexProcessWordWithMetaArgs(value, resolvedDockerMetaArgsHash)
			if err != nil {
				return "", "", err
			}

			resolvedValue = rValue
		}
	}

	return resolvedKey, resolvedValue, err
}

// resolveDockerStageArg function sets dependency arg value, or --build-arg value, or resolved dockerfile stage ARG value, or resolved meta ARG value (if stage ARG value is empty)
func (ds *DockerStages) resolveDockerStageArg(dockerStageID int, key, value string, resolvedDockerMetaArgsHash, resolvedDependenciesArgsHash map[string]string) (string, string, error) {
	resolvedKey, err := ds.ShlexProcessWordWithStageArgsAndEnvs(dockerStageID, key)
	if err != nil {
		return "", "", err
	}

	var resolvedValue string
	dependencyArgValue, ok := resolvedDependenciesArgsHash[resolvedKey]
	if ok {
		resolvedValue = dependencyArgValue
	} else {
		buildArgValue, ok := ds.dockerBuildArgsHash[resolvedKey]
		switch {
		case ok:
			resolvedValue = buildArgValue
		case value == "":
			resolvedValue = resolvedDockerMetaArgsHash[resolvedKey]
		default:
			rValue, err := ds.ShlexProcessWordWithStageArgsAndEnvs(dockerStageID, value)
			if err != nil {
				return "", "", err
			}
			resolvedValue = rValue
		}
	}

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

func (ds *DockerStages) ShlexProcessWordWithMetaArgs(value string, resolvedDockerMetaArgsHash map[string]string) (string, error) {
	return shlexProcessWord(value, toArgsArray(resolvedDockerMetaArgsHash))
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

func NewContextChecksum(dockerignorePathMatcher path_matcher.PathMatcher) *ContextChecksum {
	return &ContextChecksum{
		dockerignorePathMatcher: dockerignorePathMatcher,
	}
}

type ContextChecksum struct {
	dockerignorePathMatcher path_matcher.PathMatcher
}

type dockerfileInstructionInterface interface {
	String() string
	Name() string
}

func (s *FullDockerfileStage) FetchDependencies(ctx context.Context, c Conveyor, containerBackend container_backend.ContainerBackend, dockerRegistry docker_registry.GenericApiInterface) error {
	resolvedDependenciesArgsHash := ResolveDependenciesArgs(s.targetPlatform, s.dependencies, c)

	resolvedDockerMetaArgsHash, err := s.resolveDockerMetaArgs(resolvedDependenciesArgsHash)
	if err != nil {
		return fmt.Errorf("unable to resolve docker meta args: %w", err)
	}

outerLoop:
	for ind, stage := range s.dockerStages {
		resolvedBaseName, err := s.ShlexProcessWordWithMetaArgs(stage.BaseName, resolvedDockerMetaArgsHash)
		if err != nil {
			return err
		}

		if resolvedBaseName == "" {
			return ErrInvalidBaseImage
		}

		for relatedStageIndex, relatedStage := range s.dockerStages {
			if ind == relatedStageIndex {
				continue
			}

			if resolvedBaseName == relatedStage.Name {
				continue outerLoop
			}
		}

		_, ok := s.imageOnBuildInstructions[resolvedBaseName]
		if ok || resolvedBaseName == "scratch" {
			continue
		}

		getBaseImageOnBuildLocally := func() ([]string, error) {
			info, err := containerBackend.GetImageInfo(ctx, resolvedBaseName, container_backend.GetImageInfoOpts{})
			if err != nil {
				return nil, err
			}

			if info == nil {
				return nil, errImageNotExistLocally
			}

			return info.OnBuild, nil
		}

		getBaseImageOnBuildRemotely := func() ([]string, error) {
			configFile, err := dockerRegistry.GetRepoImageConfigFile(ctx, resolvedBaseName)
			if err != nil {
				return nil, fmt.Errorf("get repo image %q config file failed: %w", resolvedBaseName, err)
			}

			return configFile.Config.OnBuild, nil
		}

		var onBuild []string
		if onBuild, err = getBaseImageOnBuildLocally(); err != nil && err != errImageNotExistLocally {
			return err
		} else if err == errImageNotExistLocally {
			var getRemotelyErr error
			if onBuild, getRemotelyErr = getBaseImageOnBuildRemotely(); getRemotelyErr != nil {
				if isUnsupportedMediaTypeError(getRemotelyErr) {
					logboek.Context(ctx).Warn().LogF("WARNING: Could not get base image manifest from local docker and from docker registry: %s\n", getRemotelyErr)
					logboek.Context(ctx).Warn().LogLn("WARNING: The base image pulling is necessary for calculating digest of image correctly\n")
					if err := logboek.Context(ctx).Default().LogProcess("Pulling base image %s", resolvedBaseName).DoError(func() error {
						return containerBackend.Pull(ctx, resolvedBaseName, container_backend.PullOpts{})
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

var errImageNotExistLocally = errors.New("IMAGE_NOT_EXIST_LOCALLY")

func (s *FullDockerfileStage) GetDependencies(ctx context.Context, c Conveyor, cb container_backend.ContainerBackend, prevImage, prevBuiltImage *StageImage, buildContextArchive container_backend.BuildContextArchiver) (string, error) {
	resolvedDependenciesArgsHash := ResolveDependenciesArgs(s.targetPlatform, s.dependencies, c)

	resolvedDockerMetaArgsHash, err := s.resolveDockerMetaArgs(resolvedDependenciesArgsHash)
	if err != nil {
		return "", fmt.Errorf("unable to resolve docker meta args: %w", err)
	}

	var stagesDependencies [][]string
	var stagesOnBuildDependencies [][]string

	for ind, stage := range s.dockerStages {
		var dependencies []string
		var onBuildDependencies []string

		dependencies = append(dependencies, s.addHost...)

		resolvedBaseName, err := s.ShlexProcessWordWithMetaArgs(stage.BaseName, resolvedDockerMetaArgsHash)
		if err != nil {
			return "", err
		}

		dependencies = append(dependencies, resolvedBaseName)

		onBuildInstructions, ok := s.imageOnBuildInstructions[resolvedBaseName]
		if ok {
			for _, instruction := range onBuildInstructions {
				_, iOnBuildDependencies, err := s.dockerfileOnBuildInstructionDependencies(ctx, c.GiterminismManager(), resolvedDockerMetaArgsHash, resolvedDependenciesArgsHash, ind, instruction, true)
				if err != nil {
					return "", err
				}

				dependencies = append(dependencies, iOnBuildDependencies...)
			}
		}

		for _, cmd := range stage.Commands {
			cmdDependencies, cmdOnBuildDependencies, err := s.dockerfileInstructionDependencies(ctx, c.GiterminismManager(), resolvedDockerMetaArgsHash, resolvedDependenciesArgsHash, ind, cmd, false, false)
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
			switch typedCmd := cmd.(type) {
			case *instructions.CopyCommand:
				relatedStageIndex, err := strconv.Atoi(typedCmd.From)
				if err == nil && relatedStageIndex < len(stagesDependencies) {
					stagesDependencies[ind] = append(stagesDependencies[ind], stagesDependencies[relatedStageIndex]...)
				}

			case *instructions.RunCommand:
				for _, mount := range instructions.GetMounts(typedCmd) {
					relatedStageIndex, err := strconv.Atoi(mount.From)
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

func (s *FullDockerfileStage) HasPrevStage() bool {
	return false
}

func (s *FullDockerfileStage) IsStapelStage() bool {
	return false
}

func (s *FullDockerfileStage) UsesBuildContext() bool {
	return true
}

func (s *FullDockerfileStage) dockerfileInstructionDependencies(ctx context.Context, giterminismManager giterminism_manager.Interface, resolvedDockerMetaArgsHash, resolvedDependenciesArgsHash map[string]string, dockerStageID int, cmd interface{}, isOnbuildInstruction, isBaseImageOnbuildInstruction bool) ([]string, []string, error) {
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
			resolvedKey, resolvedValue, err = s.resolveDockerStageArg(dockerStageID, key, value, resolvedDockerMetaArgsHash, resolvedDependenciesArgsHash)
			if err != nil {
				return "", "", err
			}

			s.DockerStageArgsHash(dockerStageID)[resolvedKey] = resolvedValue
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
		for _, keyValuePairOptional := range c.Args {
			key := keyValuePairOptional.Key

			var value string
			if keyValuePairOptional.Value != nil {
				value = *keyValuePairOptional.Value
			}

			resolvedKey, resolvedValue, err := processArgFunc(key, value)
			if err != nil {
				return nil, nil, err
			}

			dependencies = append(dependencies, fmt.Sprintf("ARG %s=%s", resolvedKey, resolvedValue))
		}
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

		resolvedSources, err := resolveSourcesFunc(c.SourcePaths)
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
			resolvedSources, err := resolveSourcesFunc(c.SourcePaths)
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
		cDependencies, cOnBuildDependencies, err := s.dockerfileOnBuildInstructionDependencies(ctx, giterminismManager, resolvedDockerMetaArgsHash, resolvedDependenciesArgsHash, dockerStageID, c.Expression, false)
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

func (s *FullDockerfileStage) dockerfileOnBuildInstructionDependencies(ctx context.Context, giterminismManager giterminism_manager.Interface, resolvedDockerMetaArgsHash, resolvedDependenciesArgsHash map[string]string, dockerStageID int, expression string, isBaseImageOnbuildInstruction bool) ([]string, []string, error) {
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

	onBuildDependencies, _, err := s.dockerfileInstructionDependencies(ctx, giterminismManager, resolvedDockerMetaArgsHash, resolvedDependenciesArgsHash, dockerStageID, cmd, true, isBaseImageOnbuildInstruction)
	if err != nil {
		return nil, nil, err
	}

	return []string{expression}, onBuildDependencies, nil
}

func (s *FullDockerfileStage) PrepareImage(ctx context.Context, c Conveyor, cb container_backend.ContainerBackend, prevBuiltImage, stageImage *StageImage, buildContextArchive container_backend.BuildContextArchiver) error {
	if err := s.SetupDockerImageBuilder(stageImage.Builder.DockerfileBuilder(), c); err != nil {
		return err
	}

	stageImage.Builder.DockerfileBuilder().SetBuildContextArchive(buildContextArchive)

	stageImage.Builder.DockerfileBuilder().AppendLabels(fmt.Sprintf("%s=%s", image.WerfProjectRepoCommitLabel, c.GiterminismManager().HeadCommit()))

	if c.GiterminismManager().Dev() {
		stageImage.Builder.DockerfileBuilder().AppendLabels(fmt.Sprintf("%s=true", image.WerfDevLabel))
	}

	return nil
}

func (s *FullDockerfileStage) SetupDockerImageBuilder(b stage_builder.DockerfileBuilderInterface, c Conveyor) error {
	b.SetDockerfile(s.dockerfile)
	b.SetDockerfileCtxRelPath(s.dockerfilePath)

	if s.target != "" {
		b.SetTarget(s.target)
	}

	if len(s.buildArgs) > 0 {
		for key, value := range s.buildArgs {
			b.AppendBuildArgs(fmt.Sprintf("%s=%v", key, value))
		}
	}

	resolvedDependenciesArgsHash := ResolveDependenciesArgs(s.targetPlatform, s.dependencies, c)
	if len(resolvedDependenciesArgsHash) > 0 {
		for key, value := range resolvedDependenciesArgsHash {
			b.AppendBuildArgs(fmt.Sprintf("%s=%v", key, value))
		}
	}

	if len(s.addHost) > 0 {
		b.AppendAddHost(s.addHost...)
	}

	if s.network != "" {
		b.SetNetwork(s.network)
	}

	if s.ssh != "" {
		b.SetSSH(s.ssh)
	}

	return nil
}

func (s *FullDockerfileStage) calculateFilesChecksum(ctx context.Context, giterminismManager giterminism_manager.Interface, wildcards []string, dockerfileLine string) (string, error) {
	var checksum string
	var err error

	normalizedWildcards := dockerfile.NormalizeCopyAddSourcesForPathMatcher(wildcards)

	logProcess := logboek.Context(ctx).Debug().LogProcess("Calculating files checksum (%v) from local git repo", normalizedWildcards)
	logProcess.Start()

	checksum, err = s.calculateFilesChecksumWithGit(ctx, giterminismManager, normalizedWildcards, dockerfileLine)
	if err != nil {
		logProcess.Fail()
		return "", err
	} else {
		logProcess.End()
	}

	if len(s.contextAddFiles) > 0 {
		logProcess = logboek.Context(ctx).Debug().LogProcess("Calculating contextAddFiles checksum")
		logProcess.Start()

		wildcardsPathMatcher := path_matcher.NewPathMatcher(path_matcher.PathMatcherOptions{
			BasePath:     s.contextRelativeToGitWorkTree(giterminismManager),
			IncludeGlobs: wildcards,
		})
		if contextAddChecksum, err := context_manager.ContextAddFilesChecksum(ctx, giterminismManager.ProjectDir(), s.context, s.contextAddFiles, wildcardsPathMatcher); err != nil {
			logProcess.Fail()
			return "", fmt.Errorf("unable to calculate checksum for contextAddFiles files list: %w", err)
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

func (s *FullDockerfileStage) calculateFilesChecksumWithGit(ctx context.Context, giterminismManager giterminism_manager.Interface, wildcards []string, dockerfileLine string) (string, error) {
	contextPathRelativeToGitWorkTree := s.contextRelativeToGitWorkTree(giterminismManager)
	wildcardsPathMatcher := path_matcher.NewPathMatcher(path_matcher.PathMatcherOptions{
		BasePath:     contextPathRelativeToGitWorkTree,
		IncludeGlobs: wildcards,
	})
	lsTreeResultChecksum, err := giterminismManager.LocalGitRepo().GetOrCreateChecksum(ctx, git_repo.ChecksumOptions{
		LsTreeOptions: git_repo.LsTreeOptions{
			PathScope: contextPathRelativeToGitWorkTree,
			PathMatcher: path_matcher.NewMultiPathMatcher(
				s.dockerignorePathMatcher,
				wildcardsPathMatcher,
			),
			AllFiles: false,
		},
		Commit: giterminismManager.HeadCommit(),
	})
	if err != nil {
		return "", err
	}

	pathMatcher := path_matcher.NewPathMatcher(path_matcher.PathMatcherOptions{
		ExcludeGlobs: s.contextAddFilesRelativeToGitWorkTree(giterminismManager),
		Matchers: []path_matcher.PathMatcher{
			wildcardsPathMatcher,
			s.dockerignorePathMatcher,
		},
	})

	if err := giterminismManager.Inspector().InspectBuildContextFiles(ctx, pathMatcher); err != nil {
		return "", err
	}

	return util.Sha256Hash(lsTreeResultChecksum), nil
}

func dockerfileStageDependenciesDebug() bool {
	return os.Getenv("WERF_DEBUG_DOCKERFILE_STAGE_DEPENDENCIES") == "1"
}
