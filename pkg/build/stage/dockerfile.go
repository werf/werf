package stage

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"

	"github.com/docker/docker/pkg/fileutils"
	"github.com/moby/buildkit/frontend/dockerfile/instructions"
	"github.com/moby/buildkit/frontend/dockerfile/parser"
	"github.com/moby/buildkit/frontend/dockerfile/shell"

	"github.com/flant/werf/pkg/image"
	"github.com/flant/werf/pkg/util"

	"github.com/flant/logboek"
)

func GenerateDockerfileStage(dockerfilePath, target, context string, dockerignorePatternMatcher *fileutils.PatternMatcher, buildArgs map[string]interface{}, addHost []string, dockerStages []instructions.Stage, dockerArgsHash map[string]string, dockerTargetStageIndex int, baseStageOptions *NewBaseStageOptions) *DockerfileStage {
	return newDockerfileStage(dockerfilePath, target, context, dockerignorePatternMatcher, buildArgs, addHost, dockerStages, dockerArgsHash, dockerTargetStageIndex, baseStageOptions)
}

func newDockerfileStage(dockerfilePath, target, context string, dockerignorePatternMatcher *fileutils.PatternMatcher, buildArgs map[string]interface{}, addHost []string, dockerStages []instructions.Stage, dockerArgsHash map[string]string, dockerTargetStageIndex int, baseStageOptions *NewBaseStageOptions) *DockerfileStage {
	s := &DockerfileStage{}
	s.dockerfilePath = dockerfilePath
	s.target = target
	s.context = context
	s.dockerignorePatternMatcher = dockerignorePatternMatcher
	s.buildArgs = buildArgs
	s.addHost = addHost

	s.dockerStages = dockerStages
	s.dockerArgsHash = dockerArgsHash
	s.dockerTargetStageIndex = dockerTargetStageIndex

	s.BaseStage = newBaseStage(Dockerfile, baseStageOptions)

	return s
}

type DockerfileStage struct {
	dockerfilePath string
	target         string
	context        string
	buildArgs      map[string]interface{}
	addHost        []string

	dockerStages           []instructions.Stage
	dockerArgsHash         map[string]string
	dockerTargetStageIndex int

	dockerignorePatternMatcher *fileutils.PatternMatcher

	*BaseStage
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

				hashSum, err := s.calculateFilesHashsum(c.SourcesAndDest.Sources())
				if err != nil {
					return "", err
				}
				dependencies = append(dependencies, hashSum)
			case *instructions.CopyCommand:
				dependencies = append(dependencies, c.String())
				if c.From == "" {
					hashSum, err := s.calculateFilesHashsum(c.SourcesAndDest.Sources())
					if err != nil {
						return "", err
					}
					dependencies = append(dependencies, hashSum)
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

func (s *DockerfileStage) calculateFilesHashsum(wildcards []string) (string, error) {
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
			ignore, err := s.dockerignorePatternMatcher.Matches(filePath)
			if err != nil {
				return "", err
			}

			if !ignore {
				finalFileList = append(finalFileList, filePath)
			}
		}

		for _, file := range finalFileList {
			data, err := ioutil.ReadFile(file)
			if err != nil {
				return "", fmt.Errorf("read file %s failed: %s", file, err)
			}

			dependencies = append(dependencies, string(data))
		}
	}

	return util.Sha256Hash(dependencies...), nil
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
