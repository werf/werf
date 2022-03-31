package file_reader

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/werf/logboek"
	"github.com/werf/logboek/pkg/types"
	"github.com/werf/werf/pkg/giterminism_manager/errors"
	"github.com/werf/werf/pkg/path_matcher"
	"github.com/werf/werf/pkg/util"
)

func (r FileReader) projectRelativePathToAbsolutePath(relPath string) string {
	return filepath.Join(r.sharedOptions.ProjectDir(), relPath)
}

func (r FileReader) absolutePathToProjectDirRelativePath(absPath string) string {
	return util.GetRelativeToBaseFilepath(r.sharedOptions.ProjectDir(), absPath)
}

// ListFilesWithGlob returns the list of files by the glob, follows symlinks.
// The result paths are relative to the passed directory, the method does reverse resolving for symlinks.
func (r FileReader) ListFilesWithGlob(ctx context.Context, relDir, glob string, skipFileFunc func(ctx context.Context, r FileReader, existingRelPath string) (bool, error)) (files []string, err error) {
	logboek.Context(ctx).Debug().
		LogBlock("ListFilesWithGlob %q %q", relDir, glob).
		Options(func(options types.LogBlockOptionsInterface) {
			if !debug() {
				options.Mute()
			}
		}).
		Do(func() {
			files, err = r.listFilesWithGlob(ctx, relDir, glob, skipFileFunc)

			if debug() {
				var logFiles string
				if len(files) != 0 {
					logFiles = fmt.Sprintf("\n - %s", strings.Join(files, "\n - "))
				}
				logboek.Context(ctx).Debug().LogF("files: %v\nerr: %q\n", logFiles, err)
			}
		})

	return
}

func (r FileReader) listFilesWithGlob(ctx context.Context, relDir, glob string, skipFileFunc func(ctx context.Context, r FileReader, existingRelPath string) (bool, error)) ([]string, error) {
	var prefixWithoutPatterns string
	prefixWithoutPatterns, glob = util.GlobPrefixWithoutPatterns(glob)
	relDirOrFileWithGlobPart := filepath.Join(relDir, prefixWithoutPatterns)

	pathMatcher := path_matcher.NewPathMatcher(path_matcher.PathMatcherOptions{
		BasePath:     relDirOrFileWithGlobPart,
		IncludeGlobs: []string{glob},
	})
	if debug() {
		logboek.Context(ctx).Debug().LogLn("pathMatcher:", pathMatcher.String())
	}

	var result []string
	fileFunc := func(notResolvedPath string) error {
		if pathMatcher.IsPathMatched(notResolvedPath) {
			result = append(result, util.GetRelativeToBaseFilepath(relDir, notResolvedPath))
		}

		return nil
	}

	isRegularFile, err := r.isRegularFileExist(ctx, relDirOrFileWithGlobPart)
	if err != nil {
		return nil, err
	}

	if isRegularFile {
		skip, err := skipFileFunc(ctx, r, relDirOrFileWithGlobPart)
		if err != nil {
			return nil, err
		}

		if skip {
			return nil, nil
		}

		if err := fileFunc(relDirOrFileWithGlobPart); err != nil {
			return nil, err
		}

		return result, nil
	}

	err = r.walkFilesWithPathMatcher(ctx, relDirOrFileWithGlobPart, pathMatcher, skipFileFunc, fileFunc)
	return result, err
}

func (r FileReader) walkFilesWithPathMatcher(ctx context.Context, relDir string, pathMatcher path_matcher.PathMatcher, skipFileFunc func(ctx context.Context, r FileReader, existingRelPath string) (bool, error), fileFunc func(notResolvedPath string) error) error {
	if !pathMatcher.IsDirOrSubmodulePathMatched(relDir) {
		return nil
	}

	exist, err := r.IsDirectoryExist(ctx, relDir)
	if err != nil {
		return err
	}

	if !exist {
		return nil
	}

	skipDir, err := skipFileFunc(ctx, r, relDir)
	if err != nil {
		return err
	}

	if skipDir {
		return nil
	}

	resolvedDir, err := r.ResolveFilePath(ctx, relDir)
	if err != nil {
		return fmt.Errorf("unable to resolve file path %q: %w", relDir, err)
	}

	absDirPath := r.projectRelativePathToAbsolutePath(resolvedDir)
	return filepath.Walk(absDirPath, func(path string, f os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		resolvedRelPath := r.absolutePathToProjectDirRelativePath(path)
		notResolvedRelPath := strings.Replace(resolvedRelPath, resolvedDir, relDir, 1)
		for _, shouldSkipFileFunc := range []func(context.Context, FileReader, string, string) (bool, error){
			// check requires not resolved parts in path to correctly process symlinks
			func(_ context.Context, _ FileReader, resolvedRelPath, notResolvedRelPath string) (bool, error) {
				return !pathMatcher.IsDirOrSubmodulePathMatched(notResolvedRelPath), nil
			},
			// skip check expects file path
			func(ctx context.Context, r FileReader, resolvedRelPath, notResolvedRelPath string) (bool, error) {
				return skipFileFunc(ctx, r, resolvedRelPath)
			},
		} {
			shouldSkip, err := shouldSkipFileFunc(ctx, r, resolvedRelPath, notResolvedRelPath)
			if err != nil {
				return err
			}

			if shouldSkip {
				if f.IsDir() {
					return filepath.SkipDir
				} else {
					return nil
				}
			}
		}

		if f.IsDir() {
			return nil
		}

		if f.Mode()&os.ModeSymlink == os.ModeSymlink {
			link, err := os.Readlink(path)
			if err != nil {
				return fmt.Errorf("unable to read symlink %q: %w", path, err)
			}

			resolvedLink := link
			if !filepath.IsAbs(link) {
				resolvedLink = filepath.Join(filepath.Dir(path), link)
			}

			if !r.isSubpathOfWorkTreeDir(resolvedLink) {
				return r.NewFileNotFoundInProjectDirectoryError(resolvedLink)
			}

			lstat, err := os.Lstat(resolvedLink)
			if err != nil {
				return err
			}

			if lstat.IsDir() {
				if err := r.walkFilesWithPathMatcher(ctx, notResolvedRelPath, pathMatcher, skipFileFunc, fileFunc); err != nil {
					return fmt.Errorf("symlink %q resolve failed: %w", resolvedRelPath, err)
				}

				return nil
			}

			return fileFunc(notResolvedRelPath)
		}

		return fileFunc(notResolvedRelPath)
	})
}

func (r FileReader) SkipFileFunc(acceptedFilePathMatcher path_matcher.PathMatcher) func(ctx context.Context, r FileReader, existingRelPath string) (bool, error) {
	return func(ctx context.Context, r FileReader, existingRelPath string) (skip bool, err error) {
		logboek.Context(ctx).Debug().
			LogBlock("SkipFile %q", existingRelPath).
			Options(func(options types.LogBlockOptionsInterface) {
				if !debug() {
					options.Mute()
				}
			}).
			Do(func() {
				skip, err = r.skipFileFunc(acceptedFilePathMatcher)(ctx, r, existingRelPath)

				if debug() {
					logboek.Context(ctx).Debug().LogF("skip: %v\nerr: %q\n", skip, err)
				}
			})

		return
	}
}

func (r FileReader) skipFileFunc(acceptedFilePathMatcher path_matcher.PathMatcher) func(ctx context.Context, r FileReader, existingRelPath string) (bool, error) {
	return func(ctx context.Context, r FileReader, existingRelPath string) (bool, error) {
		if filepath.Base(existingRelPath) == ".git" {
			return true, nil
		}

		if r.sharedOptions.LooseGiterminism() {
			return false, nil
		}

		pathsToCheck := []string{existingRelPath}
		resolvedFilePath, err := r.ResolveFilePath(ctx, existingRelPath)
		if err != nil {
			return false, err
		}

		if existingRelPath != resolvedFilePath {
			pathsToCheck = append(pathsToCheck, resolvedFilePath)
		}

		var modified bool
		for _, relPath := range pathsToCheck {
			/* The accepted file should be read from fs */
			if acceptedFilePathMatcher.IsDirOrSubmodulePathMatched(relPath) {
				return false, nil
			}

			/* The file with changes in worktree/index should not be skipped */
			modified, err = r.IsFileModifiedLocally(ctx, relPath)
			if err != nil {
				return false, err
			}

			if modified {
				break
			}
		}

		if modified {
			return false, nil
		}

		return true, nil
	}
}

// CheckFileExistenceAndAcceptance returns nil if the resolved file exists and is fully accepted by the giterminism config (each symlink target must be accepted if the file path accepted)
func (r FileReader) CheckFileExistenceAndAcceptance(ctx context.Context, relPath string, isFileAcceptedCheckFunc func(relPath string) bool) (err error) {
	logboek.Context(ctx).Debug().
		LogBlock("CheckFileExistenceAndAcceptance %q", relPath).
		Options(func(options types.LogBlockOptionsInterface) {
			if !debug() {
				options.Mute()
			}
		}).
		Do(func() {
			err = r.checkFileExistenceAndAcceptance(ctx, relPath, isFileAcceptedCheckFunc)

			if debug() {
				logboek.Context(ctx).Debug().LogF("err: %q\n", err)
			}
		})

	return
}

func (r FileReader) checkFileExistenceAndAcceptance(ctx context.Context, relPath string, isFileAcceptedCheckFunc func(relPath string) bool) error {
	if r.sharedOptions.LooseGiterminism() {
		exist, err := r.IsRegularFileExist(ctx, relPath)
		if err != nil {
			return err
		}

		if !exist {
			return r.NewFileNotFoundInProjectDirectoryError(relPath)
		}

		return nil
	}

	if !isFileAcceptedCheckFunc(relPath) {
		return FileNotAcceptedError{fmt.Errorf("the file %q not accepted by giterminism config", relPath)}
	}

	if err := func() error {
		notAcceptedError := func(resolvedPath string) error {
			return errors.NewError(fmt.Sprintf("the link target %q should be also accepted by giterminism config", resolvedPath))
		}

		resolvedPath, err := r.ResolveAndCheckFilePath(ctx, relPath, func(resolvedRelPath string) error {
			if !isFileAcceptedCheckFunc(resolvedRelPath) {
				return notAcceptedError(resolvedRelPath)
			}

			return nil
		})
		if err != nil {
			return err
		}

		if resolvedPath != relPath {
			if !isFileAcceptedCheckFunc(relPath) {
				return notAcceptedError(resolvedPath)
			}
		}

		return nil
	}(); err != nil {
		return fmt.Errorf("accepted file %q check failed: %w", relPath, err)
	}

	return nil
}

// ShouldFileBeRead return true if not resolved path accepted by giterminism config.
func (r FileReader) ShouldFileBeRead(ctx context.Context, relPath string, isFileAcceptedCheckFunc func(relPath string) bool) (should bool, err error) {
	logboek.Context(ctx).Debug().
		LogBlock("ShouldFileBeRead %q", relPath).
		Options(func(options types.LogBlockOptionsInterface) {
			if !debug() {
				options.Mute()
			}
		}).
		Do(func() {
			should, err = r.shouldFileBeRead(relPath, isFileAcceptedCheckFunc)

			if debug() {
				logboek.Context(ctx).Debug().LogF("should: %v\nerr: %q\n", should, err)
			}
		})

	return
}

func (r FileReader) shouldFileBeRead(relPath string, isFileAcceptedCheckFunc func(relPath string) bool) (bool, error) {
	if r.sharedOptions.LooseGiterminism() {
		return true, nil
	}

	return isFileAcceptedCheckFunc(relPath), nil
}

// ReadFile returns the project file data.
func (r FileReader) ReadFile(ctx context.Context, relPath string) (data []byte, err error) {
	logboek.Context(ctx).Debug().
		LogBlock("ReadFile %q", relPath).
		Options(func(options types.LogBlockOptionsInterface) {
			if !debug() {
				options.Mute()
			}
		}).
		Do(func() {
			data, err = r.readFile(relPath)

			if debug() {
				logboek.Context(ctx).Debug().LogF("dataLength: %v\nerr: %q\n", len(data), err)
			}
		})

	return
}

func (r FileReader) readFile(relPath string) ([]byte, error) {
	absPath := r.projectRelativePathToAbsolutePath(relPath)
	data, err := ioutil.ReadFile(absPath)
	if err != nil {
		return nil, fmt.Errorf("unable to read file %q: %w", absPath, err)
	}

	return data, nil
}

// IsDirectoryExist resolves symlinks and returns true if the resolved file is a directory.
func (r FileReader) IsDirectoryExist(ctx context.Context, relPath string) (exist bool, err error) {
	logboek.Context(ctx).Debug().
		LogBlock("IsDirectoryExist %q", relPath).
		Options(func(options types.LogBlockOptionsInterface) {
			if !debug() {
				options.Mute()
			}
		}).
		Do(func() {
			exist, err = r.isDirectoryExist(ctx, relPath)

			if debug() {
				logboek.Context(ctx).Debug().LogF("exist: %v\nerr: %q\n", exist, err)
			}
		})

	return
}

func (r FileReader) isDirectoryExist(ctx context.Context, relPath string) (bool, error) {
	resolvedPath, err := r.ResolveFilePath(ctx, relPath)
	if err != nil {
		if IsFileNotFoundInProjectDirectoryError(err) {
			return false, nil
		}

		return false, fmt.Errorf("unable to resolve file path %q: %w", relPath, err)
	}

	absPath := r.projectRelativePathToAbsolutePath(resolvedPath)
	exist, err := util.DirExists(absPath)
	if err != nil {
		return false, fmt.Errorf("unable to check existence of directory %q: %w", absPath, err)
	}

	return exist, nil
}

// IsRegularFileExist resolves symlinks and returns true if the resolved file is a regular file.
func (r FileReader) IsRegularFileExist(ctx context.Context, relPath string) (exist bool, err error) {
	logboek.Context(ctx).Debug().
		LogBlock("IsRegularFileExist %q", relPath).
		Options(func(options types.LogBlockOptionsInterface) {
			if !debug() {
				options.Mute()
			}
		}).
		Do(func() {
			exist, err = r.isRegularFileExist(ctx, relPath)

			if debug() {
				logboek.Context(ctx).Debug().LogF("exist: %v\nerr: %q\n", exist, err)
			}
		})

	return
}

func (r FileReader) isRegularFileExist(ctx context.Context, relPath string) (bool, error) {
	resolvedPath, err := r.ResolveFilePath(ctx, relPath)
	if err != nil {
		if IsFileNotFoundInProjectDirectoryError(err) {
			return false, nil
		}

		return false, fmt.Errorf("unable to resolve file path %q: %w", relPath, err)
	}

	absPath := r.projectRelativePathToAbsolutePath(resolvedPath)
	exist, err := util.RegularFileExists(absPath)
	if err != nil {
		return false, fmt.Errorf("unable to check existence of file %q: %w", absPath, err)
	}

	return exist, nil
}

// ResolveAndCheckFilePath resolves the path and run checkFunc for every file path resolve.
func (r FileReader) ResolveAndCheckFilePath(ctx context.Context, relPath string, checkFunc func(resolvedPath string) error) (resolvedPath string, err error) {
	logboek.Context(ctx).Debug().
		LogBlock("ResolveAndCheckFilePath %q", relPath).
		Options(func(options types.LogBlockOptionsInterface) {
			if !debug() {
				options.Mute()
			}
		}).
		Do(func() {
			checkWithDebugFunc := func(resolvedPath string) error {
				return logboek.Context(ctx).Debug().
					LogBlock("-- check %q", resolvedPath).
					Options(func(options types.LogBlockOptionsInterface) {
						if !debug() {
							options.Mute()
						}
					}).
					DoError(func() error {
						err := checkFunc(resolvedPath)

						if debug() {
							logboek.Context(ctx).Debug().LogF("err: %q\n", err)
						}

						return err
					})
			}

			resolvedPath, err = r.resolveAndCheckFilePath(ctx, relPath, checkWithDebugFunc)

			if debug() {
				logboek.Context(ctx).Debug().LogF("resolvedPath: %q\nerr: %q\n", resolvedPath, err)
			}
		})

	return
}

func (r FileReader) resolveAndCheckFilePath(ctx context.Context, relPath string, checkSymlinkTargetFunc func(resolvedPath string) error) (resolvedPath string, err error) {
	return r.resolveFilePath(ctx, relPath, 0, checkSymlinkTargetFunc)
}

func (r FileReader) ResolveFilePath(ctx context.Context, relPath string) (resolvedPath string, err error) {
	logboek.Context(ctx).Debug().
		LogBlock("ResolveFilePath %q", relPath).
		Options(func(options types.LogBlockOptionsInterface) {
			if !debug() {
				options.Mute()
			}
		}).
		Do(func() {
			resolvedPath, err = r.resolveFilePath(ctx, relPath, 0, nil)

			if debug() {
				logboek.Context(ctx).Debug().LogF("resolvedPath: %q\nerr: %q\n", resolvedPath, err)
			}
		})

	return
}

func (r FileReader) resolveFilePath(ctx context.Context, relPath string, depth int, checkSymlinkTargetFunc func(resolvedPath string) error) (string, error) {
	if depth > 1000 {
		return "", fmt.Errorf("too many levels of symbolic links")
	}
	depth++

	pathParts := util.SplitFilepath(relPath)
	pathPartsLen := len(pathParts)

	var resolvedPath string
	for ind := 0; ind < pathPartsLen; ind++ {
		pathToResolve := filepath.Join(resolvedPath, pathParts[ind])
		absPathToResolve := r.projectRelativePathToAbsolutePath(pathToResolve)

		lstat, err := os.Lstat(absPathToResolve)

		if debug() {
			var logStat string
			if lstat != nil {
				logStat = lstat.Mode().Perm().String()
			}
			logboek.Context(ctx).Debug().LogF("-- [%d] %q %q %q\n", ind, pathToResolve, logStat, err)
		}

		if err != nil {
			if os.IsNotExist(err) || util.IsNotADirectoryError(err) {
				return "", r.NewFileNotFoundInProjectDirectoryError(pathToResolve)
			}

			return "", fmt.Errorf("unable to access file %q: %w", absPathToResolve, err)
		}

		switch {
		case lstat.Mode()&os.ModeSymlink == os.ModeSymlink:
			link, err := os.Readlink(absPathToResolve)
			if err != nil {
				return "", fmt.Errorf("unable to read symlink %q: %w", link, err)
			}

			resolvedLink := link
			if !filepath.IsAbs(link) {
				resolvedLink = filepath.Join(filepath.Dir(absPathToResolve), link)
			}

			if !r.isSubpathOfWorkTreeDir(resolvedLink) {
				return "", r.NewFileNotFoundInProjectDirectoryError(resolvedLink)
			}

			resolvedRelLink := r.absolutePathToProjectDirRelativePath(resolvedLink)
			if checkSymlinkTargetFunc != nil {
				if err := checkSymlinkTargetFunc(resolvedRelLink); err != nil {
					return "", err
				}
			}

			resolvedTarget, err := r.resolveFilePath(ctx, resolvedRelLink, depth, checkSymlinkTargetFunc)
			if err != nil {
				return "", err
			}

			resolvedPath = resolvedTarget
		default:
			resolvedPath = pathToResolve
		}
	}

	return resolvedPath, nil
}
