package file_reader

import (
	"context"
	"errors"
	"path/filepath"
	"strings"

	"github.com/samber/lo"

	"github.com/werf/common-go/pkg/util"
	"github.com/werf/logboek"
	"github.com/werf/werf/v2/pkg/path_matcher"
)

type WalkConfigurationFilesWithGlobOptions struct {
	// SkipRelativeToDirPathFunc excludes a path before it is read or checked against the
	// giterminism config, and excludes a directory along with its whole subtree.
	SkipRelativeToDirPathFunc func(relativeToDirPath string, isDir bool) bool
}

// WalkConfigurationFilesWithGlob reads the configuration files taking into account the giterminism config.
// The result paths are relative to the passed directory, the method does reverse resolving for symlinks.
func (r FileReader) WalkConfigurationFilesWithGlob(ctx context.Context, dir, glob string, acceptedFilePathMatcher path_matcher.PathMatcher, handleFileFunc func(relativeToDirNotResolvedPath string, data []byte, err error) error, opts WalkConfigurationFilesWithGlobOptions) (err error) {
	logboek.Context(ctx).Debug().
		LogBlock("WalkConfigurationFilesWithGlob %q %q", dir, glob).
		Options(applyDebugToLogboek).
		Do(func() {
			err = r.walkConfigurationFilesWithGlob(ctx, dir, glob, acceptedFilePathMatcher, handleFileFunc, opts)

			if debug() {
				logboek.Context(ctx).Debug().LogF("err: %q\n", err)
			}
		})

	return err
}

func (r FileReader) walkConfigurationFilesWithGlob(ctx context.Context, dir, glob string, acceptedFilePathMatcher path_matcher.PathMatcher, handleFileFunc func(relativeToDirNotResolvedPath string, data []byte, err error) error, opts WalkConfigurationFilesWithGlobOptions) (err error) {
	skipFileFunc := r.SkipFileFunc(acceptedFilePathMatcher)
	if opts.SkipRelativeToDirPathFunc != nil {
		skipFileFunc = r.skipConfigurationPathFunc(dir, opts.SkipRelativeToDirPathFunc, skipFileFunc)
	}

	relToDirFilePathListFromFS, err := r.ListFilesWithGlob(ctx, dir, glob, skipFileFunc)
	if err != nil {
		return err
	}

	if r.sharedOptions.LooseGiterminism() {
		for _, relToDirPath := range relToDirFilePathListFromFS {
			relPath := filepath.Join(dir, relToDirPath)
			data, err := r.ReadAndCheckConfigurationFile(ctx, relPath, acceptedFilePathMatcher.IsPathMatched, func(path string) (bool, error) {
				return r.IsRegularFileExist(ctx, path)
			})
			if err := handleFileFunc(relToDirPath, data, err); err != nil {
				return err
			}
		}

		return nil
	}

	relToDirFilePathListFromCommit, err := r.ListCommitFilesWithGlob(ctx, dir, glob)
	if err != nil {
		return err
	}

	relToDirPathList := util.AddNewStringsToStringArray(relToDirFilePathListFromFS, relToDirFilePathListFromCommit...)

	// The commit file list is flat and does not go through skipFileFunc, so it is filtered
	// separately, and an excluded directory has to be recognized through its children.
	if opts.SkipRelativeToDirPathFunc != nil {
		relToDirPathList = lo.Reject(relToDirPathList, func(relToDirPath string, _ int) bool {
			return skipRelativeToDirPathWithParents(filepath.ToSlash(relToDirPath), opts.SkipRelativeToDirPathFunc)
		})
	}

	var relPathListWithUncommittedFiles []string
	var relPathListWithUntrackedFiles []string
	for _, relToDirPath := range relToDirPathList {
		relPath := filepath.Join(dir, relToDirPath)
		data, err := r.ReadAndCheckConfigurationFile(ctx, relPath, acceptedFilePathMatcher.IsPathMatched, func(path string) (bool, error) {
			return r.IsRegularFileExist(ctx, path)
		})
		err = handleFileFunc(relToDirPath, data, err)
		if err != nil {
			switch {
			case errors.As(err, &UntrackedFilesError{}):
				relPathListWithUntrackedFiles = append(relPathListWithUntrackedFiles, relPath)
				continue
			case errors.As(err, &UncommittedFilesError{}):
				relPathListWithUncommittedFiles = append(relPathListWithUncommittedFiles, relPath)
				continue
			}

			return err
		}
	}

	if len(relPathListWithUntrackedFiles) != 0 {
		return r.NewUntrackedFilesError(relPathListWithUntrackedFiles...)
	}

	if len(relPathListWithUncommittedFiles) != 0 {
		return r.NewUncommittedFilesError(relPathListWithUncommittedFiles...)
	}

	return nil
}

// skipRelativeToDirPathWithParents applies the skip predicate to a file path and to each of its
// parent directories, which is how a directory exclusion reaches the files under it in a flat list.
func skipRelativeToDirPathWithParents(relativeToDirPath string, skipRelativeToDirPathFunc func(relativeToDirPath string, isDir bool) bool) bool {
	if skipRelativeToDirPathFunc(relativeToDirPath, false) {
		return true
	}

	parts := strings.Split(relativeToDirPath, "/")
	for i := 1; i < len(parts); i++ {
		if skipRelativeToDirPathFunc(strings.Join(parts[:i], "/"), true) {
			return true
		}
	}

	return false
}

// skipConfigurationPathFunc adapts a directory-relative skip predicate to the project-relative
// paths the walk operates on, so an excluded directory is skipped along with its whole subtree.
func (r FileReader) skipConfigurationPathFunc(dir string, skipRelativeToDirPathFunc func(relativeToDirPath string, isDir bool) bool, skipFileFunc func(ctx context.Context, r FileReader, existingRelPath string) (bool, error)) func(ctx context.Context, r FileReader, existingRelPath string) (bool, error) {
	return func(ctx context.Context, r FileReader, existingRelPath string) (bool, error) {
		relativeToDirPath := filepath.ToSlash(util.GetRelativeToBaseFilepath(dir, existingRelPath))
		if relativeToDirPath != "" && relativeToDirPath != "." {
			isDir, err := r.IsDirectoryExist(ctx, existingRelPath)
			if err != nil {
				return false, err
			}

			if skipRelativeToDirPathFunc(relativeToDirPath, isDir) {
				return true, nil
			}
		}

		return skipFileFunc(ctx, r, existingRelPath)
	}
}

// ReadAndCheckConfigurationFile does CheckConfigurationFileExistenceAndAcceptance and ReadConfigurationFile.
func (r FileReader) ReadAndCheckConfigurationFile(ctx context.Context, relPath string, isPathMatched matchPathFunc, isFileExist testFileFunc) (data []byte, err error) {
	logboek.Context(ctx).Debug().
		LogBlock("ReadAndCheckConfigurationFile %q", relPath).
		Options(applyDebugToLogboek).
		Do(func() {
			data, err = r.readAndCheckConfigurationFile(ctx, relPath, isPathMatched, isFileExist)

			if debug() {
				logboek.Context(ctx).Debug().LogF("dataLength: %v\nerr: %q\n", len(data), err)
			}
		})

	return
}

func (r FileReader) readAndCheckConfigurationFile(ctx context.Context, relPath string, isPathMatched matchPathFunc, isFileExist testFileFunc) ([]byte, error) {
	if _, err := r.CheckConfigurationFileExistenceAndAcceptance(ctx, relPath, isPathMatched, isFileExist); err != nil {
		return nil, err
	}

	return r.ReadConfigurationFile(ctx, relPath, isPathMatched, isFileExist)
}

// ReadConfigurationFile does ReadFile or ReadCommitFile depending on the giterminism config.
func (r FileReader) ReadConfigurationFile(ctx context.Context, relPath string, isPathMatched matchPathFunc, isFileExist testFileFunc) (data []byte, err error) {
	logboek.Context(ctx).Debug().
		LogBlock("ReadConfigurationFile %q", relPath).
		Options(applyDebugToLogboek).
		Do(func() {
			data, err = r.readConfigurationFile(ctx, relPath, isPathMatched, isFileExist)

			if debug() {
				logboek.Context(ctx).Debug().LogF("dataLength: %v\nerr: %q\n", len(data), err)
			}
		})

	return
}

func (r FileReader) readConfigurationFile(ctx context.Context, relPath string, isPathMatched matchPathFunc, isFileExist testFileFunc) ([]byte, error) {
	shouldFileBeReadFromFS, err := r.ShouldFileBeRead(ctx, relPath, isPathMatched, isFileExist)
	if err != nil {
		return nil, err
	}

	if shouldFileBeReadFromFS {
		return r.ReadFile(ctx, relPath)
	} else {
		return r.ReadCommitFile(ctx, relPath)
	}
}

// CheckConfigurationFileExistenceAndAcceptance does CheckFileExistenceAndAcceptance or CheckCommitFileExistenceAndLocalChanges depending on the giterminism config.
func (r FileReader) CheckConfigurationFileExistenceAndAcceptance(ctx context.Context, relPath string, isPathMatched matchPathFunc, isFileExist testFileFunc) (ok bool, err error) {
	logboek.Context(ctx).Debug().
		LogBlock("CheckConfigurationFileExistenceAndAcceptance %q", relPath).
		Options(applyDebugToLogboek).
		Do(func() {
			ok, err = r.checkConfigurationFileExistenceAndAcceptance(ctx, relPath, isPathMatched, isFileExist)

			if debug() {
				logboek.Context(ctx).Debug().LogF("err: %q\n", err)
			}
		})

	return
}

func (r FileReader) checkConfigurationFileExistenceAndAcceptance(ctx context.Context, relPath string, isPathMatched matchPathFunc, isFileExist testFileFunc) (bool, error) {
	shouldFileBeReadFromFS, err := r.ShouldFileBeRead(ctx, relPath, isPathMatched, isFileExist)
	if err != nil {
		return false, err
	}

	if shouldFileBeReadFromFS {
		err = r.CheckFileExistenceAndAcceptance(ctx, relPath, isPathMatched, isFileExist)
		if err != nil {
			return false, err
		}

		return isFileExist(relPath)
	}

	err = r.CheckCommitFileExistenceAndLocalChanges(ctx, relPath)
	if err != nil {
		return false, err
	}

	return isFileExist(relPath)
}

// IsConfigurationFileExistAnywhere returns true if the configuration file exists in the project directory or in the project repository.
func (r FileReader) IsConfigurationFileExistAnywhere(ctx context.Context, relPath string) (exist bool, err error) {
	logboek.Context(ctx).Debug().
		LogBlock("IsConfigurationFileExistAnywhere %q", relPath).
		Options(applyDebugToLogboek).
		Do(func() {
			exist, err = r.isConfigurationFileExistAnywhere(ctx, relPath)

			if debug() {
				logboek.Context(ctx).Debug().LogF("exist: %v\nerr: %q\n", exist, err)
			}
		})

	return
}

func (r FileReader) isConfigurationFileExistAnywhere(ctx context.Context, relPath string) (bool, error) {
	exist, err := r.IsRegularFileExist(ctx, relPath)
	if err != nil {
		return false, err
	}

	if exist {
		return true, nil
	}

	if r.sharedOptions.LooseGiterminism() {
		return false, nil
	} else {
		return r.IsCommitFileExist(ctx, relPath)
	}
}
