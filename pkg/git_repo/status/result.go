package status

import (
	"crypto/sha256"
	"fmt"
	"github.com/flant/werf/pkg/path_matcher"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"gopkg.in/src-d/go-git.v4"

	"github.com/flant/logboek"

	"github.com/flant/werf/pkg/util"
)

type Result struct {
	repository         *git.Repository
	repositoryFilepath string
	fileStatusList     git.Status
	submoduleResults   []*SubmoduleResult
}

type SubmoduleResult struct {
	*Result
	relToBaseRepositorySubmoduleFilepath string
	isNotInitialized                     bool
	isNotClean                           bool
	currentCommit                        string
}

func (r *Result) Status(pathMatcher path_matcher.PathMatcher) (*Result, error) {
	res := &Result{
		repository:         r.repository,
		repositoryFilepath: r.repositoryFilepath,
		fileStatusList:     git.Status{},
		submoduleResults:   []*SubmoduleResult{},
	}

	for fileFilepath, fileStatus := range r.fileStatusList {
		if pathMatcher.MatchPath(fileFilepath) {
			res.fileStatusList[fileFilepath] = fileStatus

			if debugProcess() {
				logboek.Debug.LogF("File was added:         %s (worktree: %s, staging: %s)\n", fileFilepath,
					fileStatusMapping[rune(fileStatus.Worktree)], fileStatusMapping[rune(fileStatus.Staging)])
			}
		}
	}

	for _, submoduleResult := range r.submoduleResults {
		isMatched, shouldGoThrough := pathMatcher.ProcessDirOrSubmodulePath(submoduleResult.relToBaseRepositorySubmoduleFilepath)
		if isMatched || shouldGoThrough {
			if debugProcess() {
				logboek.Debug.LogF("Submodule was checking: %s\n", submoduleResult.relToBaseRepositorySubmoduleFilepath)
			}

			if submoduleResult.isNotInitialized {
				res.submoduleResults = append(res.submoduleResults, submoduleResult)

				if debugProcess() {
					logboek.Debug.LogFWithCustomStyle(
						logboek.StyleByName(logboek.FailStyleName),
						"Submodule is not initialized: path %s will be added to checksum\n",
						submoduleResult.relToBaseRepositorySubmoduleFilepath,
					)
				}
				continue
			}

			if submoduleResult.isNotClean {
				if debugProcess() {
					logboek.Debug.LogFWithCustomStyle(
						logboek.StyleByName(logboek.FailStyleName),
						"Submodule is not clean: current commit %s will be added to checksum\n",
						submoduleResult.currentCommit,
					)
				}
			}

			sResult, err := submoduleResult.Status(pathMatcher)
			if err != nil {
				return nil, err
			}

			sSubmoduleResult := &SubmoduleResult{
				Result:                               sResult,
				relToBaseRepositorySubmoduleFilepath: submoduleResult.relToBaseRepositorySubmoduleFilepath,
				isNotInitialized:                     false,
				isNotClean:                           submoduleResult.isNotClean,
				currentCommit:                        submoduleResult.currentCommit,
			}

			if !sSubmoduleResult.isEmpty() {
				res.submoduleResults = append(res.submoduleResults, sSubmoduleResult)
			}
		}
	}

	return res, nil
}

func (r *Result) Checksum() string {
	if r.IsEmpty() {
		return ""
	}

	h := sha256.New()

	var fileFilepaths []string
	for fileFilepath := range r.fileStatusList {
		fileFilepaths = append(fileFilepaths, fileFilepath)
	}

	fileModeAndDataFunc := func(path string) (string, string) {
		absPath := filepath.Join(r.repositoryFilepath, path)

		stat, err := os.Lstat(absPath)
		if err != nil {
			panic(err)
		}

		dataH := sha256.New()
		data, err := ioutil.ReadFile(absPath)
		dataH.Write(data)

		return stat.Mode().String(), fmt.Sprintf("%x", dataH.Sum(nil))
	}

	sort.Strings(fileFilepaths)
	for _, fileFilepath := range fileFilepaths {
		fileStatus := r.fileStatusList[fileFilepath]

		var modeAndFileDataShouldBeAdded bool
		var fileStatusToAdd git.StatusCode
		var extraToAdd string

		switch fileStatus.Staging {
		case git.Untracked:
			if fileStatus.Worktree == git.Untracked {
				modeAndFileDataShouldBeAdded = true
				fileStatusToAdd = git.Untracked
			} else {
				panic(fmt.Sprintf("unexpected condition (path %s worktree %s, staging %s)", fileFilepath, fileStatusMapping[rune(fileStatus.Worktree)], fileStatusMapping[rune(fileStatus.Staging)]))
			}
		case git.Unmodified:
			switch fileStatus.Worktree {
			case git.Modified:
				modeAndFileDataShouldBeAdded = true
				fileStatusToAdd = git.Modified
			case git.Deleted:
				fileStatusToAdd = git.Deleted
			default:
				panic(fmt.Sprintf("unexpected condition (path %s worktree %s, staging %s)", fileFilepath, fileStatusMapping[rune(fileStatus.Worktree)], fileStatusMapping[rune(fileStatus.Staging)]))
			}
		case git.Added:
			switch fileStatus.Worktree {
			case git.Unmodified, git.Modified:
				modeAndFileDataShouldBeAdded = true
				fileStatusToAdd = git.Added
			case git.Deleted:
				continue
			default:
				panic(fmt.Sprintf("unexpected condition (path %s worktree %s, staging %s)", fileFilepath, fileStatusMapping[rune(fileStatus.Worktree)], fileStatusMapping[rune(fileStatus.Staging)]))
			}
		case git.Renamed:
			switch fileStatus.Worktree {
			case git.Unmodified:
				extraToAdd = fileStatus.Extra
				fileStatusToAdd = git.Renamed
			case git.Modified:
				fileStatusToAdd = git.Modified
			case git.Deleted:
				fileStatusToAdd = git.Deleted
			default:
				panic(fmt.Sprintf("unexpected condition (path %s worktree %s, staging %s)", fileFilepath, fileStatusMapping[rune(fileStatus.Worktree)], fileStatusMapping[rune(fileStatus.Staging)]))
			}
		case git.Copied:
			switch fileStatus.Worktree {
			case git.Unmodified:
				extraToAdd = fileStatus.Extra
				fileStatusToAdd = git.Copied
			case git.Modified:
				fileStatusToAdd = git.Modified
			case git.Deleted:
				fileStatusToAdd = git.Deleted
			default:
				panic(fmt.Sprintf("unexpected condition (path %s worktree %s, staging %s)", fileFilepath, fileStatusMapping[rune(fileStatus.Worktree)], fileStatusMapping[rune(fileStatus.Staging)]))
			}
		case git.Deleted:
			switch fileStatus.Worktree {
			case git.Untracked:
				modeAndFileDataShouldBeAdded = true
				fileStatusToAdd = git.Untracked
			case git.Unmodified:
				fileStatusToAdd = git.Deleted
			default:
				panic(fmt.Sprintf("unexpected condition (path %s worktree %s, staging %s)", fileFilepath, fileStatusMapping[rune(fileStatus.Worktree)], fileStatusMapping[rune(fileStatus.Staging)]))
			}
		case git.UpdatedButUnmerged:
			exist, err := util.FileExists(fileFilepath)
			if err != nil {
				panic(err)
			}

			fileStatusToAdd = git.UpdatedButUnmerged
			if exist {
				modeAndFileDataShouldBeAdded = true
			}
		default:
			panic(fmt.Sprintf("unexpected condition (path %s worktree %s, staging %s)", fileFilepath, fileStatusMapping[rune(fileStatus.Worktree)], fileStatusMapping[rune(fileStatus.Staging)]))
		}

		var args []string
		args = append(args, filepath.ToSlash(fileFilepath))

		if extraToAdd != "" {
			args = append(args, extraToAdd)
		}

		fileStatusName := fileStatusMapping[rune(fileStatusToAdd)]
		if fileStatusName != "" {
			args = append(args, fileStatusName)
		}

		if modeAndFileDataShouldBeAdded {
			mode, data := fileModeAndDataFunc(fileFilepath)
			args = append(args, mode, data)
		}

		logboek.Debug.LogF("Args was added: %v\n", args)
		logboek.Debug.LogF("  worktree %s  staging %s  result %s\n", fileStatusMapping[rune(fileStatus.Worktree)], fileStatusMapping[rune(fileStatus.Staging)], fileStatusMapping[rune(fileStatusToAdd)])
		h.Write([]byte(strings.Join(args, "🐜")))
	}

	sort.Slice(r.submoduleResults, func(i, j int) bool {
		return r.submoduleResults[i].relToBaseRepositorySubmoduleFilepath < r.submoduleResults[j].relToBaseRepositorySubmoduleFilepath
	})

	for _, sr := range r.submoduleResults {
		logBlockMsg := fmt.Sprintf("submodule %s", sr.relToBaseRepositorySubmoduleFilepath)
		_ = logboek.Debug.LogBlock(logBlockMsg, logboek.LevelLogBlockOptions{}, func() error {
			var srChecksumArgs []string

			srChecksumArgs = append(srChecksumArgs, sr.relToBaseRepositorySubmoduleFilepath)

			if sr.isNotInitialized {
				srChecksumArgs = append(srChecksumArgs, "isNotInitialized")
				return nil
			} else {
				if sr.isNotClean {
					srChecksumArgs = append(srChecksumArgs, "isNotClean")
					srChecksumArgs = append(srChecksumArgs, sr.currentCommit)
				}

				srChecksum := sr.Checksum()
				if srChecksum != "" {
					srChecksumArgs = append(srChecksumArgs, srChecksum)
				}
			}

			logboek.Debug.LogF("Args was added: %v\n", srChecksumArgs)
			h.Write([]byte(strings.Join(srChecksumArgs, "🐜")))

			return nil
		})
	}

	return fmt.Sprintf("%x", h.Sum(nil))
}

func (r *Result) IsEmpty() bool {
	return len(r.fileStatusList) == 0 && len(r.submoduleResults) == 0
}

func (sr *SubmoduleResult) isEmpty() bool {
	return sr.Result.IsEmpty() && !sr.isNotClean && !sr.isNotInitialized
}
