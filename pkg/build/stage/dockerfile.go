package stage

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/moby/buildkit/frontend/dockerfile/instructions"
	"github.com/moby/buildkit/frontend/dockerfile/parser"
	"github.com/moby/buildkit/frontend/dockerfile/shell"

	"github.com/werf/logboek"
	"github.com/werf/logboek/pkg/style"

	"github.com/werf/werf/pkg/container_runtime"
	"github.com/werf/werf/pkg/docker_registry"
	"github.com/werf/werf/pkg/git_repo"
	"github.com/werf/werf/pkg/git_repo/status"
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

func NewDockerRunArgs(dockerfilePath, target, context string, buildArgs map[string]interface{}, addHost []string) *DockerRunArgs {
	return &DockerRunArgs{
		dockerfilePath: dockerfilePath,
		target:         target,
		context:        context,
		buildArgs:      buildArgs,
		addHost:        addHost,
	}
}

type DockerRunArgs struct {
	dockerfilePath string
	target         string
	context        string
	buildArgs      map[string]interface{}
	addHost        []string
}

func NewDockerStages(dockerStages []instructions.Stage, dockerArgsHash map[string]string, dockerTargetStageIndex int) *DockerStages {
	return &DockerStages{
		dockerStages:             dockerStages,
		dockerTargetStageIndex:   dockerTargetStageIndex,
		dockerArgsHash:           dockerArgsHash,
		imageOnBuildInstructions: map[string][]string{},
	}
}

type DockerStages struct {
	dockerStages           []instructions.Stage
	dockerArgsHash         map[string]string
	dockerTargetStageIndex int

	imageOnBuildInstructions map[string][]string
}

func NewContextChecksum(projectPath string, dockerignorePathMatcher *path_matcher.DockerfileIgnorePathMatcher, localGitRepo *git_repo.Local) *ContextChecksum {
	return &ContextChecksum{
		projectPath:             projectPath,
		dockerignorePathMatcher: dockerignorePathMatcher,
		localGitRepo:            localGitRepo,
	}
}

type ContextChecksum struct {
	projectPath             string
	localGitRepo            *git_repo.Local
	dockerignorePathMatcher *path_matcher.DockerfileIgnorePathMatcher

	mainLsTreeResult *ls_tree.Result
	mainStatusResult *status.Result
}

type dockerfileInstructionInterface interface {
	String() string
	Name() string
}

func (s *DockerfileStage) FetchDependencies(ctx context.Context, _ Conveyor, cr container_runtime.ContainerRuntime) error {
	containerRuntime := cr.(*container_runtime.LocalDockerServerRuntime)

	var dockerMetaArgsString []string
	for key, value := range s.dockerArgsHash {
		dockerMetaArgsString = append(dockerMetaArgsString, fmt.Sprintf("%s=%s", key, value))
	}

	shlex := shell.NewLex(parser.DefaultEscapeToken)

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

		resolvedBaseName, err := shlex.ProcessWord(stage.BaseName, dockerMetaArgsString)
		if err != nil {
			return err
		}

		_, ok := s.imageOnBuildInstructions[resolvedBaseName]
		if ok {
			continue
		}

		inspect, err := containerRuntime.GetImageInspect(ctx, resolvedBaseName)
		if err != nil {
			return err
		} else if inspect != nil {
			s.imageOnBuildInstructions[resolvedBaseName] = inspect.Config.OnBuild
			continue
		}

		configFile, err := docker_registry.API().GetRepoImageConfigFile(ctx, resolvedBaseName)
		if err != nil {
			return fmt.Errorf("get repo image %s config file failed: %s", resolvedBaseName, err)
		} else {
			s.imageOnBuildInstructions[resolvedBaseName] = configFile.Config.OnBuild
		}
	}

	return nil
}

func (s *DockerfileStage) GetDependencies(ctx context.Context, _ Conveyor, _, _ container_runtime.ImageInterface) (string, error) {
	var dockerMetaArgsString []string
	for key, value := range s.dockerArgsHash {
		dockerMetaArgsString = append(dockerMetaArgsString, fmt.Sprintf("%s=%s", key, value))
	}

	shlex := shell.NewLex(parser.DefaultEscapeToken)

	var stagesDependencies [][]string
	var stagesOnBuildDependencies [][]string

	for _, stage := range s.dockerStages {
		var dependencies []string
		var onBuildDependencies []string

		dependencies = append(dependencies, s.addHost...)

		resolvedBaseName, err := shlex.ProcessWord(stage.BaseName, dockerMetaArgsString)
		if err != nil {
			return "", err
		}

		dependencies = append(dependencies, resolvedBaseName)

		onBuildInstructions, ok := s.imageOnBuildInstructions[resolvedBaseName]
		if ok {
			for _, instruction := range onBuildInstructions {
				_, iOnBuildDependencies, err := s.dockerfileOnBuildInstructionDependencies(ctx, instruction)
				if err != nil {
					return "", err
				}

				dependencies = append(dependencies, iOnBuildDependencies...)
			}
		}

		for _, cmd := range stage.Commands {
			cmdDependencies, cmdOnBuildDependencies, err := s.dockerfileInstructionDependencies(ctx, cmd)
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

	return util.Sha256Hash(stagesDependencies[s.dockerTargetStageIndex]...), nil
}

func (s *DockerfileStage) dockerfileInstructionDependencies(ctx context.Context, cmd interface{}) ([]string, []string, error) {
	var dependencies []string
	var onBuildDependencies []string

	switch c := cmd.(type) {
	case *instructions.ArgCommand:
		dependencies = append(dependencies, c.String())
		if argValue, exist := s.dockerArgsHash[c.Key]; exist {
			dependencies = append(dependencies, argValue)
		}
	case *instructions.AddCommand:
		dependencies = append(dependencies, c.String())

		checksum, err := s.calculateFilesChecksum(ctx, c.SourcesAndDest.Sources())
		if err != nil {
			return nil, nil, err
		}
		dependencies = append(dependencies, checksum)
	case *instructions.CopyCommand:
		dependencies = append(dependencies, c.String())
		if c.From == "" {
			checksum, err := s.calculateFilesChecksum(ctx, c.SourcesAndDest.Sources())
			if err != nil {
				return nil, nil, err
			}
			dependencies = append(dependencies, checksum)
		}
	case *instructions.OnbuildCommand:
		cDependencies, cOnBuildDependencies, err := s.dockerfileOnBuildInstructionDependencies(ctx, c.Expression)
		if err != nil {
			return nil, nil, err
		}

		dependencies = append(dependencies, cDependencies...)
		onBuildDependencies = append(onBuildDependencies, cOnBuildDependencies...)
	case dockerfileInstructionInterface:
		dependencies = append(dependencies, c.String())
	default:
		panic("runtime error")
	}

	return dependencies, onBuildDependencies, nil
}

func (s *DockerfileStage) dockerfileOnBuildInstructionDependencies(ctx context.Context, expression string) ([]string, []string, error) {
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

	onBuildDependencies, _, err := s.dockerfileInstructionDependencies(ctx, cmd)
	if err != nil {
		return nil, nil, err
	}

	return []string{expression}, onBuildDependencies, nil
}

func (s *DockerfileStage) PrepareImage(_ context.Context, c Conveyor, prevBuiltImage, img container_runtime.ImageInterface) error {
	img.DockerfileImageBuilder().AppendBuildArgs(s.DockerBuildArgs()...)
	return nil
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

	result = append(result, s.context)

	return result
}

func (s *DockerfileStage) calculateFilesChecksum(ctx context.Context, wildcards []string) (string, error) {
	var checksum string
	var err error

	normalizedWildcards := normalizeCopyAddSources(wildcards)

	logProcess := logboek.Context(ctx).Debug().LogProcess("Calculating files checksum (%v)", normalizedWildcards)
	logProcess.Start()
	if s.localGitRepo != nil {
		checksum, err = s.calculateFilesChecksumWithGit(ctx, normalizedWildcards)
	} else {
		projectFilesPaths, err := s.getProjectFilesByWildcards(ctx, normalizedWildcards)
		if err != nil {
			return "", err
		}

		checksum, err = s.calculateProjectFilesChecksum(ctx, projectFilesPaths)
	}

	if err != nil {
		logProcess.Fail()
		return "", err
	} else {
		logProcess.End()
	}

	logboek.Context(ctx).Debug().LogF("Result checksum: %s\n", checksum)
	logboek.Context(ctx).Debug().LogOptionalLn()

	return checksum, nil
}

func (s *DockerfileStage) calculateFilesChecksumWithGit(ctx context.Context, wildcards []string) (string, error) {
	if s.mainLsTreeResult == nil {
		logProcess := logboek.Context(ctx).Debug().LogProcess("ls-tree (%s)", s.dockerignorePathMatcher.String())
		logProcess.Start()
		result, err := s.localGitRepo.LsTree(ctx, s.dockerignorePathMatcher, git_repo.LsTreeOptions{UseHeadCommit: true})
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

	if s.mainStatusResult == nil {
		logProcess := logboek.Context(ctx).Debug().LogProcess("status (%s)", s.dockerignorePathMatcher.String())
		logProcess.Start()
		result, err := s.localGitRepo.Status(ctx, s.dockerignorePathMatcher)
		if err != nil {
			logProcess.Fail()
			return "", err
		} else {
			logProcess.End()
		}

		s.mainStatusResult = result
	}

	logProcess := logboek.Context(ctx).Debug().LogProcess("status (%s)", wildcardsPathMatcher.String())
	logProcess.Start()
	statusResult, err := s.mainStatusResult.Status(ctx, wildcardsPathMatcher)
	if err != nil {
		logProcess.Fail()
		return "", err
	} else {
		logProcess.End()
	}

	var statusResultChecksum string
	if !statusResult.IsEmpty() {
		if err := logboek.Context(ctx).Debug().LogBlock("Status result checksum (%s)", wildcardsPathMatcher.String()).
			DoError(func() error {
				statusResultChecksum, err = statusResult.Checksum(ctx)
				if err != nil {
					return err
				}

				logboek.Context(ctx).Debug().LogOptionalLn()
				logboek.Context(ctx).Debug().LogLn(statusResultChecksum)
				return nil
			}); err != nil {
			return "", fmt.Errorf("status result checksum failed: %s", err)
		}
	}

	logProcess = logboek.Context(ctx).Debug().LogProcess("ignored files by .gitignore files checksum (%s)", s.dockerignorePathMatcher.String())
	logProcess.Start()
	gitIgnoredFilesChecksum, err := s.calculateGitIgnoredFilesChecksum(ctx, wildcards)
	if err != nil {
		logProcess.Fail()
		return "", err
	} else {
		if gitIgnoredFilesChecksum != "" {
			logboek.Context(ctx).Debug().LogOptionalLn()
			logboek.Context(ctx).Debug().LogLn(gitIgnoredFilesChecksum)
		}

		logProcess.End()
	}

	var resultChecksum string
	if gitIgnoredFilesChecksum == "" { // TODO: legacy till v1.2
		resultChecksum = util.Sha256Hash(lsTreeResultChecksum, statusResultChecksum)
	} else {
		resultChecksum = util.Sha256Hash(lsTreeResultChecksum, statusResultChecksum, gitIgnoredFilesChecksum)
	}

	return resultChecksum, nil
}

func (s *DockerfileStage) calculateGitIgnoredFilesChecksum(ctx context.Context, wildcards []string) (string, error) {
	projectFilesPaths, err := s.getProjectFilesByWildcards(ctx, wildcards)
	if err != nil {
		return "", err
	}

	if len(projectFilesPaths) == 0 {
		return "", nil
	}

	result, err := s.localGitRepo.CheckIgnore(ctx, projectFilesPaths)
	if err != nil {
		return "", err
	}

	return s.calculateProjectFilesChecksum(ctx, result.IgnoredFilesPaths())
}

func (s *DockerfileStage) getProjectFilesByWildcards(ctx context.Context, wildcards []string) ([]string, error) {
	var paths []string

	for _, wildcard := range wildcards {
		contextWildcard := filepath.Join(s.context, wildcard)

		relContextWildcard, err := filepath.Rel(s.projectPath, contextWildcard)
		if err != nil || relContextWildcard == ".." || strings.HasPrefix(relContextWildcard, ".."+string(os.PathSeparator)) {
			logboek.Context(ctx).Warn().LogF("Outside the build context wildcard %s is not supported and skipped\n", wildcard)
			continue
		}

		matches, err := filepath.Glob(contextWildcard)
		if err != nil {
			return nil, fmt.Errorf("glob %s failed: %s", contextWildcard, err)
		}

		for _, match := range matches {
			err := filepath.Walk(match, func(path string, f os.FileInfo, err error) error {
				if err != nil {
					return err
				}

				if f.IsDir() {
					return nil
				}

				relPath, err := filepath.Rel(s.projectPath, path)
				if err != nil || relPath == "." || relPath == ".." || strings.HasPrefix(relPath, ".."+string(os.PathSeparator)) {
					panic(fmt.Sprintf("unexpected condition: project (%s) file (%s)", s.projectPath, path))
				}

				if s.dockerignorePathMatcher.MatchPath(relPath) {
					paths = append(paths, path)
				}

				return nil
			})

			if err != nil {
				return nil, fmt.Errorf("filepath walk failed: %s", err)
			}
		}
	}

	return paths, nil
}

func (s *DockerfileStage) calculateProjectFilesChecksum(ctx context.Context, paths []string) (checksum string, err error) {
	var dependencies []string

	sort.Strings(paths)
	paths = uniquePaths(paths)

	for _, path := range paths {
		relPath, err := filepath.Rel(s.projectPath, path)
		if err != nil || relPath == "." || relPath == ".." || strings.HasPrefix(relPath, ".."+string(os.PathSeparator)) {
			panic(fmt.Sprintf("unexpected condition: project (%s) file (%s)", s.projectPath, path))
		}

		dependencies = append(dependencies, relPath)
		logboek.Context(ctx).Debug().LogF("File %s was added:\n", relPath)

		stat, err := os.Lstat(path)
		if err != nil {
			return "", fmt.Errorf("os stat %s failed: %s", path, err)
		}

		dependencies = append(dependencies, stat.Mode().String())
		logboek.Context(ctx).Debug().LogF("  mode: %s\n", stat.Mode().String())

		if stat.Mode()&os.ModeSymlink != 0 {
			linkTo, err := os.Readlink(path)
			if err != nil {
				return "", fmt.Errorf("read link %s failed: %s", path, err)
			}

			dependencies = append(dependencies, linkTo)
			logboek.Context(ctx).Debug().LogF("  linkTo: %s\n", linkTo)
		} else {
			data, err := ioutil.ReadFile(path)
			if err != nil {
				return "", fmt.Errorf("read file %s failed: %s", path, err)
			}

			dataHash := util.Sha256Hash(string(data))
			dependencies = append(dependencies, dataHash)
			logboek.Context(ctx).Debug().LogF("  content hash: %s\n", dataHash)
		}
	}

	if len(dependencies) != 0 {
		checksum = util.Sha256Hash(dependencies...)
	}

	return checksum, nil
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

func uniquePaths(paths []string) []string {
	var result []string
	keys := make(map[string]bool)

	for _, p := range paths {
		if _, exist := keys[p]; !exist {
			keys[p] = true
			result = append(result, p)
		}
	}

	return result
}
