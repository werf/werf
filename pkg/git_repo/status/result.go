package status

import (
	"context"
	"crypto/sha256"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/go-git/go-git/v5"

	"github.com/werf/logboek"
	"github.com/werf/logboek/pkg/style"

	"github.com/werf/werf/pkg/path_matcher"
	"github.com/werf/werf/pkg/util"
)

type Result struct {
	repository             *git.Repository
	repositoryAbsFilepath  string
	repositoryFullFilepath string
	fileStatusList         git.Status
	submoduleResults       []*SubmoduleResult
}

type SubmoduleResult struct {
	*Result
	isNotInitialized bool
	isNotClean       bool
	currentCommit    string
}

func (r *Result) Status(ctx context.Context, pathMatcher path_matcher.PathMatcher) (*Result, error) {
	res := &Result{
		repository:             r.repository,
		repositoryAbsFilepath:  r.repositoryAbsFilepath,
		repositoryFullFilepath: r.repositoryFullFilepath,
		fileStatusList:         git.Status{},
		submoduleResults:       []*SubmoduleResult{},
	}

	for fileStatusPath, fileStatus := range r.fileStatusList {
		fileStatusFilepath := filepath.FromSlash(fileStatusPath)
		fileStatusFullFilepath := filepath.Join(r.repositoryFullFilepath, fileStatusFilepath)

		if pathMatcher.MatchPath(fileStatusFullFilepath) {
			res.fileStatusList[fileStatusPath] = fileStatus

			if debugProcess() {
				logboek.Context(ctx).Debug().LogF(
					"File was added:         %s (worktree: %s, staging: %s)\n",
					fileStatusFullFilepath,
					fileStatusMapping[rune(fileStatus.Worktree)],
					fileStatusMapping[rune(fileStatus.Staging)],
				)
			}
		}
	}

	for _, submoduleResult := range r.submoduleResults {
		isMatched, shouldGoThrough := pathMatcher.ProcessDirOrSubmodulePath(submoduleResult.repositoryFullFilepath)
		if isMatched || shouldGoThrough {
			if debugProcess() {
				logboek.Context(ctx).Debug().LogF("Submodule was checking: %s\n", submoduleResult.repositoryFullFilepath)
			}

			if submoduleResult.isNotInitialized {
				res.submoduleResults = append(res.submoduleResults, submoduleResult)

				if debugProcess() {
					logboek.Context(ctx).Debug().LogFWithCustomStyle(
						style.Get(style.FailName),
						"Submodule is not initialized: path %s will be added to checksum\n",
						submoduleResult.repositoryFullFilepath,
					)
				}
				continue
			}

			if submoduleResult.isNotClean {
				if debugProcess() {
					logboek.Context(ctx).Debug().LogFWithCustomStyle(
						style.Get(style.FailName),
						"Submodule is not clean: current commit %s will be added to checksum\n",
						submoduleResult.currentCommit,
					)
				}
			}

			newResult, err := submoduleResult.Status(ctx, pathMatcher)
			if err != nil {
				return nil, err
			}

			newSubmoduleResult := &SubmoduleResult{
				Result:           newResult,
				isNotInitialized: false,
				isNotClean:       submoduleResult.isNotClean,
				currentCommit:    submoduleResult.currentCommit,
			}

			if !newSubmoduleResult.isEmpty() {
				res.submoduleResults = append(res.submoduleResults, newSubmoduleResult)
			}
		}
	}

	return res, nil
}

func (r *Result) FilePathList() ([]string, error) {
	if r.IsEmpty() {
		return []string{}, nil
	}

	var result []string
	for fileStatusPath := range r.fileStatusList {
		result = append(result, fileStatusPath)
	}

	return result, nil
}

func (r *Result) Checksum(ctx context.Context) (string, error) {
	if r.IsEmpty() {
		return "", nil
	}

	h := sha256.New()

	var fileStatusPathList []string
	for fileStatusPath := range r.fileStatusList {
		fileStatusPathList = append(fileStatusPathList, fileStatusPath)
	}

	fileModeAndDataFunc := func(fileStatusAbsFilepath string) (string, string, error) {
		stat, err := os.Lstat(fileStatusAbsFilepath)
		if err != nil {
			return "", "", fmt.Errorf("os stat %s failed: %s", fileStatusAbsFilepath, err)
		}

		dataH := sha256.New()
		if stat.Mode()&os.ModeSymlink != 0 {
			linkTo, err := os.Readlink(fileStatusAbsFilepath)
			if err != nil {
				return "", "", fmt.Errorf("os read link %s failed: %s", fileStatusAbsFilepath, err)
			}

			dataH.Write([]byte(linkTo))
		} else {
			data, err := ioutil.ReadFile(fileStatusAbsFilepath)
			if err != nil {
				return "", "", fmt.Errorf("os read file %s failed: %s", fileStatusAbsFilepath, err)
			}

			dataH.Write(data)
		}

		return stat.Mode().String(), fmt.Sprintf("%x", dataH.Sum(nil)), nil
	}

	sort.Strings(fileStatusPathList)

	for _, fileStatusPath := range fileStatusPathList {
		fileStatus := r.fileStatusList[fileStatusPath]
		fileStatusFilepath := filepath.FromSlash(fileStatusPath)
		fileStatusFullFilepath := filepath.Join(r.repositoryFullFilepath, fileStatusFilepath)
		fileStatusAbsFilepath := filepath.Join(r.repositoryAbsFilepath, fileStatusFilepath)

		var modeAndFileDataShouldBeAdded bool
		var fileStatusToAdd git.StatusCode
		var extraToAdd string

		switch fileStatus.Staging {
		case git.Untracked:
			if fileStatus.Worktree == git.Untracked {
				modeAndFileDataShouldBeAdded = true
				fileStatusToAdd = git.Untracked
			} else {
				panic(fmt.Sprintf("unexpected condition (path %s worktree %s, staging %s)", fileStatusAbsFilepath, fileStatusMapping[rune(fileStatus.Worktree)], fileStatusMapping[rune(fileStatus.Staging)]))
			}
		case git.Unmodified:
			switch fileStatus.Worktree {
			case git.Modified:
				modeAndFileDataShouldBeAdded = true
				fileStatusToAdd = git.Modified
			case git.Deleted:
				fileStatusToAdd = git.Deleted
			default:
				panic(fmt.Sprintf("unexpected condition (path %s worktree %s, staging %s)", fileStatusAbsFilepath, fileStatusMapping[rune(fileStatus.Worktree)], fileStatusMapping[rune(fileStatus.Staging)]))
			}
		case git.Added, git.Modified:
			switch fileStatus.Worktree {
			case git.Unmodified, git.Modified:
				modeAndFileDataShouldBeAdded = true
				fileStatusToAdd = git.Added
			case git.Deleted:
				continue
			default:
				panic(fmt.Sprintf("unexpected condition (path %s worktree %s, staging %s)", fileStatusAbsFilepath, fileStatusMapping[rune(fileStatus.Worktree)], fileStatusMapping[rune(fileStatus.Staging)]))
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
				panic(fmt.Sprintf("unexpected condition (path %s worktree %s, staging %s)", fileStatusAbsFilepath, fileStatusMapping[rune(fileStatus.Worktree)], fileStatusMapping[rune(fileStatus.Staging)]))
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
				panic(fmt.Sprintf("unexpected condition (path %s worktree %s, staging %s)", fileStatusAbsFilepath, fileStatusMapping[rune(fileStatus.Worktree)], fileStatusMapping[rune(fileStatus.Staging)]))
			}
		case git.Deleted:
			switch fileStatus.Worktree {
			case git.Untracked:
				modeAndFileDataShouldBeAdded = true
				fileStatusToAdd = git.Untracked
			case git.Unmodified:
				fileStatusToAdd = git.Deleted
			default:
				panic(fmt.Sprintf("unexpected condition (path %s worktree %s, staging %s)", fileStatusAbsFilepath, fileStatusMapping[rune(fileStatus.Worktree)], fileStatusMapping[rune(fileStatus.Staging)]))
			}
		case git.UpdatedButUnmerged:
			exist, err := util.FileExists(fileStatusAbsFilepath)
			if err != nil {
				panic(err)
			}

			fileStatusToAdd = git.UpdatedButUnmerged
			if exist {
				modeAndFileDataShouldBeAdded = true
			}
		default:
			panic(fmt.Sprintf("unexpected condition (path %s worktree %s, staging %s)", fileStatusAbsFilepath, fileStatusMapping[rune(fileStatus.Worktree)], fileStatusMapping[rune(fileStatus.Staging)]))
		}

		fileStatusFullPath := filepath.ToSlash(fileStatusFullFilepath)

		var args []string
		args = append(args, fileStatusFullPath)

		if extraToAdd != "" {
			args = append(args, extraToAdd)
		}

		fileStatusName := fileStatusMapping[rune(fileStatusToAdd)]
		if fileStatusName != "" {
			args = append(args, fileStatusName)
		}

		if modeAndFileDataShouldBeAdded {
			mode, data, err := fileModeAndDataFunc(fileStatusAbsFilepath)
			if err != nil {
				return "", err
			}

			args = append(args, mode, data)
		}

		logboek.Context(ctx).Debug().LogF("Args was added: %v\n", args)
		logboek.Context(ctx).Debug().LogF("  worktree %s  staging %s  result %s\n", fileStatusMapping[rune(fileStatus.Worktree)], fileStatusMapping[rune(fileStatus.Staging)], fileStatusMapping[rune(fileStatusToAdd)])
		h.Write([]byte(strings.Join(args, "🐜")))
	}

	sort.Slice(r.submoduleResults, func(i, j int) bool {
		return r.submoduleResults[i].repositoryFullFilepath < r.submoduleResults[j].repositoryFullFilepath
	})

	for _, sr := range r.submoduleResults {
		logboek.Context(ctx).Debug().LogOptionalLn()
		if err := logboek.Context(ctx).Debug().LogBlock("submodule %s", sr.repositoryFullFilepath).DoError(func() error {
			var srChecksumArgs []string

			srChecksumArgs = append(srChecksumArgs, sr.repositoryFullFilepath)

			if sr.isNotInitialized {
				srChecksumArgs = append(srChecksumArgs, "isNotInitialized")
				return nil
			} else {
				if sr.isNotClean {
					srChecksumArgs = append(srChecksumArgs, "isNotClean")
					srChecksumArgs = append(srChecksumArgs, sr.currentCommit)
				}

				srChecksum, err := sr.Checksum(ctx)
				if err != nil {
					return err
				}

				if srChecksum != "" {
					srChecksumArgs = append(srChecksumArgs, srChecksum)
				}
			}

			logboek.Context(ctx).Debug().LogF("Args was added: %v\n", srChecksumArgs)
			h.Write([]byte(strings.Join(srChecksumArgs, "🐜")))

			return nil
		}); err != nil {
			return "", fmt.Errorf("submodule %s checksum failed: %s", sr.repositoryFullFilepath, err)
		}
	}

	return fmt.Sprintf("%x", h.Sum(nil)), nil
}

func (r *Result) IsEmpty() bool {
	return len(r.fileStatusList) == 0 && len(r.submoduleResults) == 0
}

func (sr *SubmoduleResult) isEmpty() bool {
	return sr.Result.IsEmpty() && !sr.isNotClean && !sr.isNotInitialized
}
