package file_reader

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	pathPkg "path"
	"path/filepath"
	"strings"

	"github.com/bmatcuk/doublestar"

	"github.com/werf/werf/pkg/util"
)

// ListFilesWithGlob returns the list of files by the glob, follows symlinks.
// The result paths are relative to the passed directory, the method does reverse resolving for symlinks.
func (r FileReader) ListFilesWithGlob(ctx context.Context, relDir, glob string) ([]string, error) {
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

	absDirPath := filepath.Join(r.sharedOptions.ProjectDir(), resolvedDir)
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

		notResolvedPath := strings.Replace(path, resolvedDir, relDir, 1)
		notResolvedRelPath := util.GetRelativeToBaseFilepath(r.sharedOptions.ProjectDir(), notResolvedPath)
		if f.Mode()&os.ModeSymlink == os.ModeSymlink {
			link, err := os.Readlink(path)
			if err != nil {
				return fmt.Errorf("unable to read symlink %q: %s", path, err)
			}

			resolvedLink := link
			if !filepath.IsAbs(link) {
				resolvedLink = filepath.Join(filepath.Dir(path), link)
			}

			if !util.IsSubpathOfBasePath(r.sharedOptions.ProjectDir(), resolvedLink) {
				return configurationFileNotFoundInProjectDirectoryErr
			}

			lstat, err := os.Lstat(resolvedLink)
			if err != nil {
				return fmt.Errorf("lstat %q failed: %s", resolvedLink, err)
			}

			fmt.Println(lstat.IsDir(), resolvedLink, notResolvedRelPath)

			if lstat.IsDir() {
				return r.walkFiles(ctx, notResolvedRelPath, fileFunc)
			}

			return fileFunc(notResolvedRelPath)
		}

		return fileFunc(notResolvedRelPath)
	})
}

// IsRegularFileExistAndAccepted returns true if the file exists locally and every file path resolve is accepted by the giterminism config.
func (r FileReader) IsRegularFileExistAndAccepted(ctx context.Context, relPath string, isFileAcceptedCheckFunc func(relPath string) (bool, error)) (bool, error) {
	exist, err := r.IsRegularFileExist(ctx, relPath)
	if err != nil {
		return false, err
	}

	if exist {
		accepted, err := r.IsFilePathAccepted(ctx, relPath, isFileAcceptedCheckFunc)
		if err != nil {
			return false, err
		}

		if accepted {
			return true, nil
		}
	}

	return false, nil
}

// IsFilePathAccepted returns true if every file path resolve is accepted by the giterminism config.
func (r FileReader) IsFilePathAccepted(ctx context.Context, relPath string, isFileAcceptedCheckFunc func(relPath string) (bool, error)) (accepted bool, err error) {
	var notAcceptedFilePathErr = errors.New("skip not accepted file path")

	if _, err := r.ResolveAndCheckFilePath(ctx, relPath, func(resolvedPath string) error {
		if r.sharedOptions.LooseGiterminism() {
			return nil
		}

		accepted, err := isFileAcceptedCheckFunc(resolvedPath)
		if err != nil {
			return err
		}

		if !accepted {
			return notAcceptedFilePathErr
		}

		return nil
	}); err != nil {
		if err == notAcceptedFilePathErr || err == configurationFileNotFoundInProjectDirectoryErr {
			return false, nil
		}

		return false, err
	}

	return true, nil
}

// ReadFile returns the project file data.
func (r FileReader) ReadFile(ctx context.Context, relPath string) ([]byte, error) {
	absPath := filepath.Join(r.sharedOptions.ProjectDir(), relPath)
	data, err := ioutil.ReadFile(absPath)
	if err != nil {
		return nil, fmt.Errorf("unable to read file %q: %s", absPath, err)
	}

	return data, nil
}

// isDirectoryExist resolves symlinks and returns true if the resolved file is a directory.
func (r FileReader) IsDirectoryExist(ctx context.Context, relPath string) (bool, error) {
	resolvedPath, err := r.ResolveFilePath(ctx, relPath)
	if err != nil {
		if err == configurationFileNotFoundInProjectDirectoryErr {
			return false, nil
		}

		return false, fmt.Errorf("unable to resolve file path %q: %s", relPath, err)
	}

	absPath := filepath.Join(r.sharedOptions.ProjectDir(), resolvedPath)
	exist, err := util.DirExists(absPath)
	if err != nil {
		return false, fmt.Errorf("unable to check existence of directory %q: %s", absPath, err)
	}

	return exist, nil
}

// IsRegularFileExist resolves symlinks and returns true if the resolved file is a regular file.
func (r FileReader) IsRegularFileExist(ctx context.Context, relPath string) (bool, error) {
	resolvedPath, err := r.ResolveFilePath(ctx, relPath)
	if err != nil {
		if err == configurationFileNotFoundInProjectDirectoryErr {
			return false, nil
		}

		return false, fmt.Errorf("unable to resolve file path %q: %s", relPath, err)
	}

	absPath := filepath.Join(r.sharedOptions.ProjectDir(), resolvedPath)
	exist, err := util.RegularFileExists(absPath)
	if err != nil {
		return false, fmt.Errorf("unable to check existence of file %q: %s", absPath, err)
	}

	return exist, nil
}

// ResolveAndCheckFilePath resolves the path and run checkFunc for every file path resolve.
func (r FileReader) ResolveAndCheckFilePath(ctx context.Context, relPath string, checkFunc func(resolvedPath string) error) (resolvedPath string, err error) {
	if err := checkFunc(relPath); err != nil {
		return "", err
	}

	resolvedPath, err = r.resolveFilePath(ctx, relPath, 0, checkFunc)
	if err != nil {
		return "", err
	}

	if resolvedPath != relPath {
		if err := checkFunc(resolvedPath); err != nil {
			return "", err
		}
	}

	return resolvedPath, nil
}

func (r FileReader) ResolveFilePath(ctx context.Context, relPath string) (resolvedPath string, err error) {
	return r.resolveFilePath(ctx, relPath, 0, nil)
}

var configurationFileNotFoundInProjectDirectoryErr = fmt.Errorf("the configutation file not found in the project directory")

func (r FileReader) resolveFilePath(ctx context.Context, relPath string, depth int, checkFunc func(resolvedPath string) error) (string, error) {
	if depth > 1000 {
		return "", fmt.Errorf("too many levels of symbolic links")
	}
	depth++

	pathParts := util.SplitFilepath(relPath)
	pathPartsLen := len(pathParts)

	var resolvedPath string
	for ind := 0; ind < pathPartsLen; ind++ {
		pathToResolve := filepath.Join(resolvedPath, pathParts[ind])
		absPathToResolve := filepath.Join(r.sharedOptions.ProjectDir(), pathToResolve)

		lstat, err := os.Lstat(absPathToResolve)
		if err != nil {
			if os.IsNotExist(err) || util.IsNotADirectoryError(err) {
				return "", configurationFileNotFoundInProjectDirectoryErr
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

			if !util.IsSubpathOfBasePath(r.sharedOptions.ProjectDir(), resolvedLink) {
				return "", configurationFileNotFoundInProjectDirectoryErr
			}

			resolvedRelLink := util.GetRelativeToBaseFilepath(r.sharedOptions.ProjectDir(), resolvedLink)
			if checkFunc != nil {
				if err := checkFunc(resolvedRelLink); err != nil {
					return "", err
				}
			}

			resolvedTarget, err := r.resolveFilePath(ctx, resolvedRelLink, depth, checkFunc)
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
