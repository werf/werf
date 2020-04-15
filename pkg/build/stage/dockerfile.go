package stage

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/moby/buildkit/frontend/dockerfile/instructions"
	"github.com/moby/buildkit/frontend/dockerfile/parser"
	"github.com/moby/buildkit/frontend/dockerfile/shell"

	"github.com/flant/logboek"

	"github.com/flant/werf/pkg/container_runtime"
	"github.com/flant/werf/pkg/git_repo"
	"github.com/flant/werf/pkg/git_repo/ls_tree"
	"github.com/flant/werf/pkg/git_repo/status"
	"github.com/flant/werf/pkg/path_matcher"
	"github.com/flant/werf/pkg/util"
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
		dockerStages:           dockerStages,
		dockerTargetStageIndex: dockerTargetStageIndex,
		dockerArgsHash:         dockerArgsHash,
	}
}

type DockerStages struct {
	dockerStages           []instructions.Stage
	dockerArgsHash         map[string]string
	dockerTargetStageIndex int
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

func (s *DockerfileStage) GetDependencies(_ Conveyor, _, _ container_runtime.ImageInterface) (string, error) {
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

		for _, cmd := range stage.Commands {
			cmdDependencies, cmdOnBuildDependencies, err := s.dockerfileInstructionDependencies(cmd)
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
					if err != nil || relatedStageIndex >= len(stagesDependencies) {
						return "", fmt.Errorf("COPY --from refers to nonexistent dockerfile stage %s", c.From)
					} else {
						stagesDependencies[ind] = append(stagesDependencies[ind], stagesDependencies[relatedStageIndex]...)
					}
				}
			}
		}
	}

	return util.Sha256Hash(stagesDependencies[s.dockerTargetStageIndex]...), nil
}

func (s *DockerfileStage) dockerfileInstructionDependencies(cmd interface{}) ([]string, []string, error) {
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

		checksum, err := s.calculateFilesChecksum(c.SourcesAndDest.Sources())
		if err != nil {
			return nil, nil, err
		}
		dependencies = append(dependencies, checksum)
	case *instructions.CopyCommand:
		dependencies = append(dependencies, c.String())
		if c.From == "" {
			checksum, err := s.calculateFilesChecksum(c.SourcesAndDest.Sources())
			if err != nil {
				return nil, nil, err
			}
			dependencies = append(dependencies, checksum)
		}
	case *instructions.OnbuildCommand:
		p, err := parser.Parse(bytes.NewReader([]byte(c.Expression)))
		if err != nil {
			return nil, nil, err
		}

		if len(p.AST.Children) != 1 {
			panic(fmt.Sprintf("unexpected condition: %s (%d children)", c.String(), len(p.AST.Children)))
		}

		instruction := p.AST.Children[0]
		cmd, err := instructions.ParseInstruction(instruction)
		if err != nil {
			return nil, nil, err
		}

		cDependencies, _, err := s.dockerfileInstructionDependencies(cmd)
		if err != nil {
			return nil, nil, err
		}

		dependencies = append(dependencies, c.String())
		onBuildDependencies = append(onBuildDependencies, cDependencies...)
	case dockerfileInstructionInterface:
		dependencies = append(dependencies, c.String())
	default:
		panic("runtime error")
	}

	return dependencies, onBuildDependencies, nil
}

func (s *DockerfileStage) PrepareImage(c Conveyor, prevBuiltImage, img container_runtime.ImageInterface) error {
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

func (s *DockerfileStage) calculateFilesChecksum(wildcards []string) (string, error) {
	var checksum string
	var err error

	logProcessMsg := fmt.Sprintf("Calculating files checksum (%v)", wildcards)
	logboek.Debug.LogProcessStart(logProcessMsg, logboek.LevelLogProcessStartOptions{})
	if s.localGitRepo != nil {
		checksum, err = s.calculateFilesChecksumWithGit(wildcards)
	} else {
		projectFilesPaths, err := s.getProjectFilesByWildcards(wildcards)
		if err != nil {
			return "", err
		}

		checksum, err = s.calculateProjectFilesChecksum(projectFilesPaths)
	}

	if err != nil {
		logboek.Debug.LogProcessFail(logboek.LevelLogProcessFailOptions{})
		return "", err
	}

	logboek.Debug.LogProcessEnd(logboek.LevelLogProcessEndOptions{})

	logboek.Debug.LogF("Result checksum: %s\n", checksum)
	logboek.Debug.LogOptionalLn()

	return checksum, nil
}

func (s *DockerfileStage) calculateFilesChecksumWithGit(wildcards []string) (string, error) {
	if s.mainLsTreeResult == nil {
		processMsg := fmt.Sprintf("ls-tree (%s)", s.dockerignorePathMatcher.String())
		logboek.Debug.LogProcessStart(processMsg, logboek.LevelLogProcessStartOptions{})
		result, err := s.localGitRepo.LsTree(s.dockerignorePathMatcher)
		if err != nil {
			if err.Error() == "entry not found" {
				logboek.Debug.LogFWithCustomStyle(
					logboek.StyleByName(logboek.FailStyleName),
					"Entry %s is not found\n",
					s.dockerignorePathMatcher.BaseFilepath(),
				)
				logboek.Debug.LogProcessEnd(logboek.LevelLogProcessEndOptions{})
				goto entryNotFoundInGitRepository
			}

			logboek.Debug.LogProcessFail(logboek.LevelLogProcessFailOptions{})
			return "", err
		}
		logboek.Debug.LogProcessEnd(logboek.LevelLogProcessEndOptions{})

		s.mainLsTreeResult = result
	}

entryNotFoundInGitRepository:
	wildcardsPathMatcher := path_matcher.NewSimplePathMatcher(s.dockerignorePathMatcher.BaseFilepath(), wildcards, false)

	var lsTreeResultChecksum string
	if s.mainLsTreeResult != nil {
		blockMsg := fmt.Sprintf("ls-tree (%s)", wildcardsPathMatcher.String())
		logboek.Debug.LogProcessStart(blockMsg, logboek.LevelLogProcessStartOptions{})
		lsTreeResult, err := s.mainLsTreeResult.LsTree(wildcardsPathMatcher)
		if err != nil {
			logboek.Debug.LogProcessFail(logboek.LevelLogProcessFailOptions{})
			return "", err
		}
		logboek.Debug.LogProcessEnd(logboek.LevelLogProcessEndOptions{})

		if !lsTreeResult.IsEmpty() {
			blockMsg = fmt.Sprintf("ls-tree result checksum (%s)", wildcardsPathMatcher.String())
			_ = logboek.Debug.LogBlock(blockMsg, logboek.LevelLogBlockOptions{}, func() error {
				lsTreeResultChecksum = lsTreeResult.Checksum()
				logboek.Debug.LogOptionalLn()
				logboek.Debug.LogLn(lsTreeResultChecksum)

				return nil
			})
		}
	}

	if s.mainStatusResult == nil {
		processMsg := fmt.Sprintf("status (%s)", s.dockerignorePathMatcher.String())
		logboek.Debug.LogProcessStart(processMsg, logboek.LevelLogProcessStartOptions{})
		result, err := s.localGitRepo.Status(s.dockerignorePathMatcher)
		if err != nil {
			logboek.Debug.LogProcessFail(logboek.LevelLogProcessFailOptions{})
			return "", err
		}
		logboek.Debug.LogProcessEnd(logboek.LevelLogProcessEndOptions{})

		s.mainStatusResult = result
	}

	blockMsg := fmt.Sprintf("status (%s)", wildcardsPathMatcher.String())
	logboek.Debug.LogProcessStart(blockMsg, logboek.LevelLogProcessStartOptions{})
	statusResult, err := s.mainStatusResult.Status(wildcardsPathMatcher)
	if err != nil {
		logboek.Debug.LogProcessFail(logboek.LevelLogProcessFailOptions{})
		return "", err
	}
	logboek.Debug.LogProcessEnd(logboek.LevelLogProcessEndOptions{})

	var statusResultChecksum string
	if !statusResult.IsEmpty() {
		blockMsg = fmt.Sprintf("Status result checksum (%s)", wildcardsPathMatcher.String())
		if err := logboek.Debug.LogBlock(blockMsg, logboek.LevelLogBlockOptions{}, func() error {
			statusResultChecksum, err = statusResult.Checksum()
			if err != nil {
				return err
			}

			logboek.Debug.LogOptionalLn()
			logboek.Debug.LogLn(statusResultChecksum)
			return nil
		}); err != nil {
			return "", fmt.Errorf("status result checksum failed: %s", err)
		}
	}

	blockMsg = fmt.Sprintf("ignored files by .gitignore files checksum (%s)", s.dockerignorePathMatcher.String())
	logboek.Debug.LogProcessStart(blockMsg, logboek.LevelLogProcessStartOptions{})
	gitIgnoredFilesChecksum, err := s.calculateGitIgnoredFilesChecksum(wildcards)
	if err != nil {
		logboek.Debug.LogProcessFail(logboek.LevelLogProcessFailOptions{})
		return "", err
	}
	if gitIgnoredFilesChecksum != "" {
		logboek.Debug.LogOptionalLn()
		logboek.Debug.LogLn(gitIgnoredFilesChecksum)
	}
	logboek.Debug.LogProcessEnd(logboek.LevelLogProcessEndOptions{})

	var resultChecksum string
	if gitIgnoredFilesChecksum == "" { // TODO: legacy till v1.2
		resultChecksum = util.Sha256Hash(lsTreeResultChecksum, statusResultChecksum)
	} else {
		resultChecksum = util.Sha256Hash(lsTreeResultChecksum, statusResultChecksum, gitIgnoredFilesChecksum)
	}

	return resultChecksum, nil
}

func (s *DockerfileStage) calculateGitIgnoredFilesChecksum(wildcards []string) (string, error) {
	projectFilesPaths, err := s.getProjectFilesByWildcards(wildcards)
	if err != nil {
		return "", err
	}

	if len(projectFilesPaths) == 0 {
		return "", nil
	}

	result, err := s.localGitRepo.CheckIgnore(projectFilesPaths)
	if err != nil {
		return "", err
	}

	return s.calculateProjectFilesChecksum(result.IgnoredFilesPaths())
}

func (s *DockerfileStage) getProjectFilesByWildcards(wildcards []string) ([]string, error) {
	var paths []string

	for _, wildcard := range wildcards {
		contextWildcard := filepath.Join(s.context, wildcard)

		relContextWildcard, err := filepath.Rel(s.projectPath, contextWildcard)
		if err != nil || relContextWildcard == ".." || strings.HasPrefix(relContextWildcard, ".."+string(os.PathSeparator)) {
			logboek.Warn.LogF("Outside the build context wildcard %s is not supported and skipped\n", wildcard)
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

func (s *DockerfileStage) calculateProjectFilesChecksum(paths []string) (checksum string, err error) {
	var dependencies []string

	sort.Strings(paths)
	paths = uniquePaths(paths)

	for _, path := range paths {
		relPath, err := filepath.Rel(s.projectPath, path)
		if err != nil || relPath == "." || relPath == ".." || strings.HasPrefix(relPath, ".."+string(os.PathSeparator)) {
			panic(fmt.Sprintf("unexpected condition: project (%s) file (%s)", s.projectPath, path))
		}

		dependencies = append(dependencies, relPath)
		logboek.Debug.LogF("File %s was added:\n", relPath)

		stat, err := os.Lstat(path)
		if err != nil {
			return "", fmt.Errorf("os stat %s failed: %s", path, err)
		}

		dependencies = append(dependencies, stat.Mode().String())
		logboek.Debug.LogF("  mode: %s\n", stat.Mode().String())

		if stat.Mode()&os.ModeSymlink != 0 {
			linkTo, err := os.Readlink(path)
			if err != nil {
				return "", fmt.Errorf("read link %s failed: %s", path, err)
			}

			dependencies = append(dependencies, linkTo)
			logboek.Debug.LogF("  linkTo: %s\n", linkTo)
		} else {
			data, err := ioutil.ReadFile(path)
			if err != nil {
				return "", fmt.Errorf("read file %s failed: %s", path, err)
			}

			dataHash := util.Sha256Hash(string(data))
			dependencies = append(dependencies, dataHash)
			logboek.Debug.LogF("  content hash: %s\n", dataHash)
		}
	}

	if len(dependencies) != 0 {
		checksum = util.Sha256Hash(dependencies...)
	}

	return checksum, nil
}

func uniquePaths(paths []string) []string {
	var result []string
	keys := make(map[string]bool)

	for _, path := range paths {
		if _, exist := keys[path]; !exist {
			keys[path] = true
			result = append(result, path)
		}
	}

	return result
}
