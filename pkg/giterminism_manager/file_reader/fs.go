package file_reader

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	pathPkg "path"
	"path/filepath"
	"strings"

	"github.com/bmatcuk/doublestar"

	"github.com/werf/logboek"
	"github.com/werf/logboek/pkg/types"
	"github.com/werf/werf/pkg/util"
)

func (r FileReader) toProjectDirAbsolutePath(relPath string) string {
	return filepath.Join(r.sharedOptions.ProjectDir(), relPath)
}

func (r FileReader) toProjectDirRelativePath(absPath string) string {
	return util.GetRelativeToBaseFilepath(r.sharedOptions.ProjectDir(), absPath)
}

// ListFilesWithGlob returns the list of files by the glob, follows symlinks.
// The result paths are relative to the passed directory, the method does reverse resolving for symlinks.
func (r FileReader) ListFilesWithGlob(ctx context.Context, relDir, glob string) (files []string, err error) {
	logboek.Context(ctx).Debug().
		LogBlock("ListFilesWithGlob %q %q", relDir, glob).
		Options(func(options types.LogBlockOptionsInterface) {
			if !debug() {
				options.Mute()
			}
		}).
		Do(func() {
			files, err = r.listFilesWithGlob(ctx, relDir, glob)

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

func (r FileReader) listFilesWithGlob(ctx context.Context, relDir, glob string) ([]string, error) {
	var result []string

	glob = filepath.FromSlash(glob)
	matchFunc := func(path string) (bool, error) {
		for _, glob := range []string{
			glob,
			pathPkg.Join(glob, "**", "*"),
		} {
			matched, err := doublestar.PathMatch(glob, path)
			if err != nil {
				return false, err
			}

			if matched {
				return true, nil
			}
		}

		return false, nil
	}

	err := r.walkFiles(ctx, relDir, func(notResolvedPath string) error {
		matched, err := matchFunc(notResolvedPath)
		if err != nil {
			return err
		}

		if debug() {
			logboek.Context(ctx).Debug().LogF("-- %q %q\n", notResolvedPath, matched)
		}

		if matched {
			result = append(result, notResolvedPath)
		}

		return nil
	})

	return result, err
}

func (r FileReader) walkFiles(ctx context.Context, relDir string, fileFunc func(notResolvedPath string) error) error {
	exist, err := r.IsDirectoryExist(ctx, relDir)
	if err != nil {
		return err
	}

	if !exist {
		return nil
	}

	resolvedDir, err := r.ResolveFilePath(ctx, relDir)
	if err != nil {
		return fmt.Errorf("unable to resolve file path %q: %s", relDir, err)
	}

	absDirPath := r.toProjectDirAbsolutePath(resolvedDir)
	return filepath.Walk(absDirPath, func(path string, f os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if f.IsDir() {
			if filepath.Base(path) == ".git" {
				return filepath.SkipDir
			}

			return nil
		}

		if debug() {
			logboek.Context(ctx).Debug().LogF("-- path: %q symlink: %v\n", path, f.Mode()&os.ModeSymlink == os.ModeSymlink)
		}

		resolvedRelPath := r.toProjectDirRelativePath(path)
		notResolvedRelPath := strings.Replace(resolvedRelPath, resolvedDir, relDir, 1)
		if f.Mode()&os.ModeSymlink == os.ModeSymlink {
			link, err := os.Readlink(path)
			if err != nil {
				return fmt.Errorf("unable to read symlink %q: %s", path, err)
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
				if err := r.walkFiles(ctx, notResolvedRelPath, fileFunc); err != nil {
					return fmt.Errorf("symlink %q resolve failed: %s", resolvedRelPath, err)
				}

				return nil
			}

			return fileFunc(notResolvedRelPath)
		}

		return fileFunc(notResolvedRelPath)
	})
}

// CheckFileExistenceAndAcceptance returns nil if the resolved file exists and is fully accepted by the giterminism config (each symlink target must be accepted if the file path accepted)
func (r FileReader) CheckFileExistenceAndAcceptance(ctx context.Context, relPath string, isFileAcceptedCheckFunc func(relPath string) (bool, error)) (err error) {
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

func (r FileReader) checkFileExistenceAndAcceptance(ctx context.Context, relPath string, isFileAcceptedCheckFunc func(relPath string) (bool, error)) error {
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

	accepted, err := isFileAcceptedCheckFunc(relPath)
	if err != nil {
		return err
	}

	if !accepted {
		return FileNotAcceptedError{fmt.Errorf("the file %q not accepted by giterminism config", relPath)}
	}

	resolvedPath, err := r.ResolveAndCheckFilePath(ctx, relPath, func(resolvedRelPath string) error {
		accepted, err := isFileAcceptedCheckFunc(resolvedRelPath)
		if err != nil {
			return err
		}

		if !accepted {
			return fmt.Errorf("the link target %q should be also accepted by giterminism config", resolvedRelPath)
		}

		return nil
	})
	if err != nil {
		return r.NewSymlinkResolveFailedError(relPath, err)
	}

	if resolvedPath != relPath {
		accepted, err := isFileAcceptedCheckFunc(relPath)
		if err != nil {
			return err
		}

		if !accepted {
			return r.NewSymlinkResolveFailedError(relPath, fmt.Errorf("the link target %q should be also accepted by giterminism config", resolvedPath))
		}
	}

	return nil
}

// ShouldFileBeRead return true if not resolved path accepted by giterminism config.
func (r FileReader) ShouldFileBeRead(ctx context.Context, relPath string, isFileAcceptedCheckFunc func(relPath string) (bool, error)) (should bool, err error) {
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

func (r FileReader) shouldFileBeRead(relPath string, isFileAcceptedCheckFunc func(relPath string) (bool, error)) (bool, error) {
	if r.sharedOptions.LooseGiterminism() {
		return true, nil
	}

	accepted, err := isFileAcceptedCheckFunc(relPath)
	if err != nil {
		return false, err
	}

	return accepted, nil
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
	absPath := r.toProjectDirAbsolutePath(relPath)
	data, err := ioutil.ReadFile(absPath)
	if err != nil {
		return nil, fmt.Errorf("unable to read file %q: %s", absPath, err)
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

		return false, fmt.Errorf("unable to resolve file path %q: %s", relPath, err)
	}

	absPath := r.toProjectDirAbsolutePath(resolvedPath)
	exist, err := util.DirExists(absPath)
	if err != nil {
		return false, fmt.Errorf("unable to check existence of directory %q: %s", absPath, err)
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

		return false, fmt.Errorf("unable to resolve file path %q: %s", relPath, err)
	}

	absPath := r.toProjectDirAbsolutePath(resolvedPath)
	exist, err := util.RegularFileExists(absPath)
	if err != nil {
		return false, fmt.Errorf("unable to check existence of file %q: %s", absPath, err)
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
		absPathToResolve := r.toProjectDirAbsolutePath(pathToResolve)

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

			return "", fmt.Errorf("unable to access file %q: %s", absPathToResolve, err)
		}

		switch {
		case lstat.Mode()&os.ModeSymlink == os.ModeSymlink:
			link, err := os.Readlink(absPathToResolve)
			if err != nil {
				return "", fmt.Errorf("unable to read symlink %q: %s", link, err)
			}

			resolvedLink := link
			if !filepath.IsAbs(link) {
				resolvedLink = filepath.Join(filepath.Dir(absPathToResolve), link)
			}

			if !r.isSubpathOfWorkTreeDir(resolvedLink) {
				return "", r.NewFileNotFoundInProjectDirectoryError(resolvedLink)
			}

			resolvedRelLink := r.toProjectDirRelativePath(resolvedLink)
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
