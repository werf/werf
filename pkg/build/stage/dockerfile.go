package stage

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/moby/buildkit/frontend/dockerfile/instructions"
	"github.com/moby/buildkit/frontend/dockerfile/parser"
	"github.com/moby/buildkit/frontend/dockerfile/shell"

	"github.com/flant/werf/pkg/git_repo"
	"github.com/flant/werf/pkg/git_repo/ls_tree"
	"github.com/flant/werf/pkg/git_repo/status"
	"github.com/flant/werf/pkg/image"
	"github.com/flant/werf/pkg/path_matcher"
	"github.com/flant/werf/pkg/util"

	"github.com/flant/logboek"
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

func (s *DockerfileStage) GetDependencies(_ Conveyor, _, _ image.ImageInterface) (string, error) {
	var dockerMetaArgsString []string
	for key, value := range s.dockerArgsHash {
		dockerMetaArgsString = append(dockerMetaArgsString, fmt.Sprintf("%s=%s", key, value))
	}

	shlex := shell.NewLex(parser.DefaultEscapeToken)

	var stagesDependencies [][]string
	for _, stage := range s.dockerStages {
		var dependencies []string

		dependencies = append(dependencies, s.addHost...)

		resolvedBaseName, err := shlex.ProcessWord(stage.BaseName, dockerMetaArgsString)
		if err != nil {
			return "", err
		}

		dependencies = append(dependencies, resolvedBaseName)

		for _, cmd := range stage.Commands {
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
					return "", err
				}
				dependencies = append(dependencies, checksum)
			case *instructions.CopyCommand:
				dependencies = append(dependencies, c.String())
				if c.From == "" {
					checksum, err := s.calculateFilesChecksum(c.SourcesAndDest.Sources())
					if err != nil {
						return "", err
					}
					dependencies = append(dependencies, checksum)
				}
			case dockerfileInstructionInterface:
				dependencies = append(dependencies, c.String())
			default:
				panic("runtime error")
			}
		}

		stagesDependencies = append(stagesDependencies, dependencies)
	}

	for ind, stage := range s.dockerStages {
		for relatedStageIndex, relatedStage := range s.dockerStages {
			if ind == relatedStageIndex {
				continue
			}

			if stage.BaseName == relatedStage.Name {
				stagesDependencies[ind] = append(stagesDependencies[ind], stagesDependencies[relatedStageIndex]...)
			}
		}

		for _, cmd := range stage.Commands {
			switch c := cmd.(type) {
			case *instructions.CopyCommand:
				if c.From != "" {
					relatedStageIndex, err := strconv.Atoi(c.From)
					if err == nil && relatedStageIndex < len(stagesDependencies) {
						stagesDependencies[ind] = append(stagesDependencies[ind], stagesDependencies[relatedStageIndex]...)
					} else {
						logboek.LogWarnF("WARNING: COPY --from with unexistent stage %s detected\n", c.From)
					}
				}
			}
		}
	}

	return util.Sha256Hash(stagesDependencies[s.dockerTargetStageIndex]...), nil
}

func (s *DockerfileStage) PrepareImage(c Conveyor, prevBuiltImage, img image.ImageInterface) error {
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
		checksum, err = s.calculateFilesChecksumWithLsTree(wildcards)
	} else {
		checksum, err = s.calculateFilesChecksumWithFilesRead(wildcards)
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

func (s *DockerfileStage) calculateFilesChecksumWithLsTree(wildcards []string) (string, error) {
	if s.mainLsTreeResult == nil {
		processMsg := fmt.Sprintf("ls-tree (%s)", s.dockerignorePathMatcher.String())
		logboek.Debug.LogProcessStart(processMsg, logboek.LevelLogProcessStartOptions{})
		result, err := s.localGitRepo.LsTree(s.dockerignorePathMatcher)
		if err != nil {
			if err.Error() == "entry not found" {
				logboek.Debug.LogFWithCustomStyle(
					logboek.StyleByName(logboek.FailStyleName),
					"Entry %s is not found\n",
					s.dockerignorePathMatcher.BasePath(),
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
	wildcardsPathMatcher := path_matcher.NewSimplePathMatcher(s.dockerignorePathMatcher.BasePath(), wildcards, false)

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
				logboek.Debug.LogLn()
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
		_ = logboek.Debug.LogBlock(blockMsg, logboek.LevelLogBlockOptions{}, func() error {
			statusResultChecksum = statusResult.Checksum()
			logboek.Debug.LogLn()
			logboek.Debug.LogLn(statusResultChecksum)
			return nil
		})
	}

	resultChecksum := util.Sha256Hash(lsTreeResultChecksum, statusResultChecksum)

	return resultChecksum, nil
}

func (s *DockerfileStage) calculateFilesChecksumWithFilesRead(wildcards []string) (string, error) {
	var dependencies []string

	for _, wildcard := range wildcards {
		contextWildcard := filepath.Join(s.context, wildcard)

		matches, err := filepath.Glob(contextWildcard)
		if err != nil {
			return "", fmt.Errorf("glob %s failed: %s", contextWildcard, err)
		}

		var fileList []string
		for _, match := range matches {
			matchFileList, err := getAllFiles(match)
			if err != nil {
				return "", fmt.Errorf("walk %s failed: %s", match, err)
			}

			fileList = append(fileList, matchFileList...)
		}

		var finalFileList []string
		for _, filePath := range fileList {
			relFilePath, err := filepath.Rel(s.projectPath, filePath)
			if err != nil {
				panic(fmt.Sprintf("unexpected condition: %s", err))
			} else if strings.HasPrefix(relFilePath, "."+string(os.PathSeparator)) || strings.HasPrefix(relFilePath, ".."+string(os.PathSeparator)) {
				panic(fmt.Sprintf("unexpected condition: %s", relFilePath))
			}

			if s.dockerignorePathMatcher.MatchPath(relFilePath) {
				finalFileList = append(finalFileList, filePath)
			}
		}

		for _, filePath := range finalFileList {
			data, err := ioutil.ReadFile(filePath)
			if err != nil {
				return "", fmt.Errorf("read file %s failed: %s", filePath, err)
			}

			dependencies = append(dependencies, string(data))
			logboek.Debug.LogF("File was added: %s\n", strings.TrimPrefix(filePath, s.projectPath+string(os.PathSeparator)))
		}
	}

	checksum := util.Sha256Hash(dependencies...)

	return checksum, nil
}

func getAllFiles(target string) ([]string, error) {
	var fileList []string
	err := filepath.Walk(target, func(path string, f os.FileInfo, err error) error {
		if f.IsDir() {
			return nil
		}

		if f.Mode()&os.ModeSymlink != 0 {
			linkTo, err := os.Readlink(path)
			if err != nil {
				return err
			}

			linkFilePath := filepath.Join(filepath.Dir(path), linkTo)
			exist, err := util.FileExists(linkFilePath)
			if err != nil {
				return err
			} else if !exist {
				return nil
			} else {
				lfinfo, err := os.Stat(linkFilePath)
				if err != nil {
					return err
				}

				if lfinfo.IsDir() {
					// infinite loop detector
					if target == linkFilePath {
						return nil
					}

					lfileList, err := getAllFiles(linkFilePath)
					if err != nil {
						return err
					}

					fileList = append(fileList, lfileList...)
				} else {
					fileList = append(fileList, linkFilePath)
				}

				return nil
			}
		}

		fileList = append(fileList, path)
		return err
	})

	return fileList, err
}
