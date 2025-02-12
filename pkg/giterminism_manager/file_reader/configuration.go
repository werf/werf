package file_reader

import (
	"context"
	"errors"
	"path/filepath"

	"github.com/werf/common-go/pkg/util"
	"github.com/werf/logboek"
	"github.com/werf/werf/v2/pkg/path_matcher"
)

// WalkConfigurationFilesWithGlob reads the configuration files taking into account the giterminism config.
// The result paths are relative to the passed directory, the method does reverse resolving for symlinks.
func (r FileReader) WalkConfigurationFilesWithGlob(ctx context.Context, dir, glob string, acceptedFilePathMatcher path_matcher.PathMatcher, handleFileFunc func(relativeToDirNotResolvedPath string, data []byte, err error) error) (err error) {
	logboek.Context(ctx).Debug().
		LogBlock("WalkConfigurationFilesWithGlob %q %q", dir, glob).
		Options(applyDebugToLogboek).
		Do(func() {
			err = r.walkConfigurationFilesWithGlob(ctx, dir, glob, acceptedFilePathMatcher, handleFileFunc)

			if debug() {
				logboek.Context(ctx).Debug().LogF("err: %q\n", err)
			}
		})

	return
}

func (r FileReader) walkConfigurationFilesWithGlob(ctx context.Context, dir, glob string, acceptedFilePathMatcher path_matcher.PathMatcher, handleFileFunc func(relativeToDirNotResolvedPath string, data []byte, err error) error) (err error) {
	relToDirFilePathListFromFS, err := r.ListFilesWithGlob(ctx, dir, glob, r.SkipFileFunc(acceptedFilePathMatcher))
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
