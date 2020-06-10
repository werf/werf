package check_ignore

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/go-git/go-git/v5"

	"github.com/werf/logboek"
)

func CheckIgnore(repository *git.Repository, absRepositoryFilepath string, absFilepathsToCheck []string) (*Result, error) {
	return checkIgnore(repository, absRepositoryFilepath, absFilepathsToCheck)
}

func checkIgnore(repository *git.Repository, absRepositoryFilepath string, absFilepathsToCheck []string) (*Result, error) {
	worktree, err := repository.Worktree()
	if err != nil {
		return nil, err
	}

	submodules, err := worktree.Submodules()
	if err != nil {
		return nil, err
	}

	var repositoryAbsFilepathsToCheck []string
	absFilepathsToCheckBySubmodule := map[string][]string{}

mainLoop:
	for _, absFilepathToCheck := range absFilepathsToCheck {
		for _, submodule := range submodules {
			submoduleFilepath := filepath.FromSlash(submodule.Config().Path)
			submoduleAbsFilepath := filepath.Join(absRepositoryFilepath, submoduleFilepath)

			if strings.HasPrefix(absFilepathToCheck, submoduleAbsFilepath+string(os.PathSeparator)) {
				submoduleAbsFilepathsToCheck, ok := absFilepathsToCheckBySubmodule[submoduleFilepath]
				if !ok {
					submoduleAbsFilepathsToCheck = []string{}
				}

				submoduleAbsFilepathsToCheck = append(submoduleAbsFilepathsToCheck, absFilepathToCheck)
				absFilepathsToCheckBySubmodule[submoduleFilepath] = submoduleAbsFilepathsToCheck

				continue mainLoop
			}
		}

		repositoryAbsFilepathsToCheck = append(repositoryAbsFilepathsToCheck, absFilepathToCheck)
	}

	ignoredAbsFilepaths, err := getRepositoryIgnoredAbsFilepaths(absRepositoryFilepath, repositoryAbsFilepathsToCheck)
	if err != nil {
		return nil, err
	}

	result := &Result{
		repository:            repository,
		repositoryAbsFilepath: absRepositoryFilepath,
		ignoredAbsFilepaths:   ignoredAbsFilepaths,
	}

	for _, submodule := range submodules {
		submoduleFilepath := filepath.FromSlash(submodule.Config().Path)
		submoduleAbsFilepath := filepath.Join(absRepositoryFilepath, submoduleFilepath)

		submoduleAbsFilepathsToCheck, ok := absFilepathsToCheckBySubmodule[submoduleFilepath]
		if !ok || len(submoduleAbsFilepathsToCheck) == 0 {
			continue
		}

		submoduleRepository, err := submodule.Repository()
		if err != nil {
			if err == git.ErrSubmoduleNotInitialized {
				logboek.Debug.LogFWithCustomStyle(
					logboek.StyleByName(logboek.FailStyleName),
					"Submodule %s is not initialized: the following paths will not be counted:\n%s",
					submoduleFilepath,
					strings.Join(submoduleAbsFilepathsToCheck, "\n"),
				)

				continue
			}

			return nil, fmt.Errorf("getting submodule repository failed (%s): %s", submoduleFilepath, err)
		}

		submoduleResult, err := checkIgnore(submoduleRepository, submoduleAbsFilepath, submoduleAbsFilepathsToCheck)
		if err != nil {
			return nil, err
		}

		result.submoduleResults = append(result.submoduleResults, &SubmoduleResult{Result: submoduleResult})
	}

	return result, nil
}

func getRepositoryIgnoredAbsFilepaths(repositoryAbsFilepath string, absFilepathsToCheck []string) ([]string, error) {
	if len(absFilepathsToCheck) == 0 {
		return []string{}, nil
	}

	toStdinString := strings.Join(absFilepathsToCheck, "\n")

	var b bytes.Buffer
	b.Write([]byte(toStdinString))

	command := "git"
	commandArgs := []string{"-C", repositoryAbsFilepath, "check-ignore", "--stdin"}
	commandString := strings.Join(append([]string{command}, commandArgs...), " ")

	cmd := exec.Command(command, commandArgs...)
	cmd.Stdin = &b

	if debugProcess() {
		logboek.Debug.LogLn("command:", commandString)
		logboek.Debug.LogLn("stdin:  ", toStdinString)
	}

	output, err := cmd.CombinedOutput()

	if debugProcess() {
		logboek.Debug.LogLn("output:\n", string(output))
	}

	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			if s, ok := exitErr.Sys().(syscall.WaitStatus); ok {
				exitCode := s.ExitStatus()
				if exitCode == 1 { // None of the provided paths are ignored
					if debugProcess() {
						logboek.Debug.LogLn("None of the provided paths are ignored")
					}

					return []string{}, nil
				}
			}
		}

		panic(fmt.Sprintf("\nerr: %s\ncommand: %s\noutput:\n%s", err, commandString, string(output)))
	}

	var ignoredPaths []string
	scanner := bufio.NewScanner(bytes.NewReader(output))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		line = strings.Trim(line, "\"")
		if line == "" {
			continue
		}

		ignoredPaths = append(ignoredPaths, line)
	}

	return ignoredPaths, nil
}

func debugProcess() bool {
	return os.Getenv("WERF_DEBUG_CHECK_IGNORE") == "1"
}
