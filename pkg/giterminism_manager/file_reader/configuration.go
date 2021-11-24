package file_reader

import (
	"context"
	"path/filepath"

	"github.com/werf/logboek"
	"github.com/werf/logboek/pkg/types"
	"github.com/werf/werf/pkg/path_matcher"
	"github.com/werf/werf/pkg/util"
)

// WalkConfigurationFilesWithGlob reads the configuration files taking into account the giterminism config.
// The result paths are relative to the passed directory, the method does reverse resolving for symlinks.
func (r FileReader) WalkConfigurationFilesWithGlob(ctx context.Context, dir, glob string, acceptedFilePathMatcher path_matcher.PathMatcher, handleFileFunc func(relativeToDirNotResolvedPath string, data []byte, err error) error) (err error) {
	logboek.Context(ctx).Debug().
		LogBlock("WalkConfigurationFilesWithGlob %q %q", dir, glob).
		Options(func(options types.LogBlockOptionsInterface) {
			if !debug() {
				options.Mute()
			}
		}).
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
			data, err := r.ReadAndCheckConfigurationFile(ctx, relPath, acceptedFilePathMatcher.IsPathMatched)
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
		data, err := r.ReadAndCheckConfigurationFile(ctx, relPath, acceptedFilePathMatcher.IsPathMatched)
		err = handleFileFunc(relToDirPath, data, err)
		if err != nil {
			switch err.(type) {
			case UntrackedFilesError:
				relPathListWithUntrackedFiles = append(relPathListWithUntrackedFiles, relPath)
				continue
			case UncommittedFilesError:
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
func (r FileReader) ReadAndCheckConfigurationFile(ctx context.Context, relPath string, isFileAcceptedCheckFunc func(relPath string) bool) (data []byte, err error) {
	logboek.Context(ctx).Debug().
		LogBlock("ReadAndCheckConfigurationFile %q", relPath).
		Options(func(options types.LogBlockOptionsInterface) {
			if !debug() {
				options.Mute()
			}
		}).
		Do(func() {
			data, err = r.readAndCheckConfigurationFile(ctx, relPath, isFileAcceptedCheckFunc)

			if debug() {
				logboek.Context(ctx).Debug().LogF("dataLength: %v\nerr: %q\n", len(data), err)
			}
		})

	return
}

func (r FileReader) readAndCheckConfigurationFile(ctx context.Context, relPath string, isFileAcceptedCheckFunc func(relPath string) bool) ([]byte, error) {
	if err := r.CheckConfigurationFileExistenceAndAcceptance(ctx, relPath, isFileAcceptedCheckFunc); err != nil {
		return nil, err
	}

	return r.ReadConfigurationFile(ctx, relPath, isFileAcceptedCheckFunc)
}

// ReadConfigurationFile does ReadFile or ReadCommitFile depending on the giterminism config.
func (r FileReader) ReadConfigurationFile(ctx context.Context, relPath string, isFileAcceptedCheckFunc func(relPath string) bool) (data []byte, err error) {
	logboek.Context(ctx).Debug().
		LogBlock("ReadConfigurationFile %q", relPath).
		Options(func(options types.LogBlockOptionsInterface) {
			if !debug() {
				options.Mute()
			}
		}).
		Do(func() {
			data, err = r.readConfigurationFile(ctx, relPath, isFileAcceptedCheckFunc)

			if debug() {
				logboek.Context(ctx).Debug().LogF("dataLength: %v\nerr: %q\n", len(data), err)
			}
		})

	return
}

func (r FileReader) readConfigurationFile(ctx context.Context, relPath string, isFileAcceptedCheckFunc func(relPath string) bool) ([]byte, error) {
	shouldFileBeReadFromFS, err := r.ShouldFileBeRead(ctx, relPath, isFileAcceptedCheckFunc)
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
func (r FileReader) CheckConfigurationFileExistenceAndAcceptance(ctx context.Context, relPath string, isFileAcceptedCheckFunc func(relPath string) bool) (err error) {
	logboek.Context(ctx).Debug().
		LogBlock("CheckConfigurationFileExistenceAndAcceptance %q", relPath).
		Options(func(options types.LogBlockOptionsInterface) {
			if !debug() {
				options.Mute()
			}
		}).
		Do(func() {
			err = r.checkConfigurationFileExistenceAndAcceptance(ctx, relPath, isFileAcceptedCheckFunc)

			if debug() {
				logboek.Context(ctx).Debug().LogF("err: %q\n", err)
			}
		})

	return
}

func (r FileReader) checkConfigurationFileExistenceAndAcceptance(ctx context.Context, relPath string, isFileAcceptedCheckFunc func(relPath string) bool) error {
	shouldFileBeReadFromFS, err := r.ShouldFileBeRead(ctx, relPath, isFileAcceptedCheckFunc)
	if err != nil {
		return err
	}

	if shouldFileBeReadFromFS {
		return r.CheckFileExistenceAndAcceptance(ctx, relPath, isFileAcceptedCheckFunc)
	}

	return r.CheckCommitFileExistenceAndLocalChanges(ctx, relPath)
}

// IsConfigurationFileExistAnywhere returns true if the configuration file exists in the project directory or in the project repository.
func (r FileReader) IsConfigurationFileExistAnywhere(ctx context.Context, relPath string) (exist bool, err error) {
	logboek.Context(ctx).Debug().
		LogBlock("IsConfigurationFileExistAnywhere %q", relPath).
		Options(func(options types.LogBlockOptionsInterface) {
			if !debug() {
				options.Mute()
			}
		}).
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

// IsConfigurationFileExist checks the configuration file existence taking into account the giterminism config.
// The method does not check acceptance for each symlink target if the configuration file is symlink.
func (r FileReader) IsConfigurationFileExist(ctx context.Context, relPath string, isFileAcceptedCheckFunc func(relPath string) bool) (exist bool, err error) {
	logboek.Context(ctx).Debug().
		LogBlock("IsConfigurationFileExist %q", relPath).
		Options(func(options types.LogBlockOptionsInterface) {
			if !debug() {
				options.Mute()
			}
		}).
		Do(func() {
			exist, err = r.isConfigurationFileExist(ctx, relPath, isFileAcceptedCheckFunc)

			if debug() {
				logboek.Context(ctx).Debug().LogF("exist: %v\nerr: %q\n", exist, err)
			}
		})

	return
}

func (r FileReader) isConfigurationFileExist(ctx context.Context, relPath string, isFileAcceptedCheckFunc func(relPath string) bool) (bool, error) {
	shouldFileBeReadFromFS, err := r.ShouldFileBeRead(ctx, relPath, isFileAcceptedCheckFunc)
	if err != nil {
		return false, err
	}

	if shouldFileBeReadFromFS {
		return r.IsRegularFileExist(ctx, relPath)
	} else {
		return r.IsCommitFileExist(ctx, relPath)
	}
}
