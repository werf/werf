package true_git

import (
	"archive/tar"
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/go-git/go-git/v5/plumbing/filemode"

	"github.com/werf/logboek"

	"github.com/werf/werf/pkg/git_repo/repo_handle"
	"github.com/werf/werf/pkg/path_matcher"
	"github.com/werf/werf/pkg/true_git/ls_tree"
	"github.com/werf/werf/pkg/util"
)

type ArchiveOptions struct {
	Commit string

	// the PathScope option determines the directory or file that will get into the result (similar to <pathspec> in the git commands)
	PathScope string

	PathMatcher path_matcher.PathMatcher
}

func (opts ArchiveOptions) ID() string {
	return util.Sha256Hash(
		opts.Commit,
		opts.PathScope,
		opts.PathMatcher.ID(),
	)
}

type ArchiveDescriptor struct {
	Type    ArchiveType
	IsEmpty bool
}

type ArchiveType string

const (
	FileArchive      ArchiveType = "file"
	DirectoryArchive ArchiveType = "directory"
)

func ArchiveWithSubmodules(ctx context.Context, out io.Writer, gitDir, workTreeCacheDir string, opts ArchiveOptions) (*ArchiveDescriptor, error) {
	var res *ArchiveDescriptor

	err := withWorkTreeCacheLock(ctx, workTreeCacheDir, func() error {
		writeArchiveRes, err := writeArchive(ctx, out, gitDir, workTreeCacheDir, true, opts)
		res = writeArchiveRes
		return err
	})

	return res, err
}

func Archive(ctx context.Context, out io.Writer, gitDir, workTreeCacheDir string, opts ArchiveOptions) (*ArchiveDescriptor, error) {
	var res *ArchiveDescriptor

	err := withWorkTreeCacheLock(ctx, workTreeCacheDir, func() error {
		writeArchiveRes, err := writeArchive(ctx, out, gitDir, workTreeCacheDir, false, opts)
		res = writeArchiveRes
		return err
	})

	return res, err
}

func debugArchive() bool {
	return os.Getenv("WERF_TRUE_GIT_DEBUG_ARCHIVE") == "1"
}

func writeArchive(ctx context.Context, out io.Writer, gitDir, workTreeCacheDir string, withSubmodules bool, opts ArchiveOptions) (*ArchiveDescriptor, error) {
	var err error

	gitDir, err = filepath.Abs(gitDir)
	if err != nil {
		return nil, fmt.Errorf("bad git dir %s: %s", gitDir, err)
	}

	workTreeCacheDir, err = filepath.Abs(workTreeCacheDir)
	if err != nil {
		return nil, fmt.Errorf("bad work tree cache dir %s: %s", workTreeCacheDir, err)
	}

	workTreeDir, err := prepareWorkTree(ctx, gitDir, workTreeCacheDir, opts.Commit, withSubmodules)
	if err != nil {
		return nil, fmt.Errorf("cannot prepare work tree in cache %s for commit %s: %s", workTreeCacheDir, opts.Commit, err)
	}

	repository, err := GitOpenWithCustomWorktreeDir(gitDir, workTreeDir)
	if err != nil {
		return nil, fmt.Errorf("git open failed: %s", err)
	}

	desc := &ArchiveDescriptor{
		IsEmpty: true,
	}

	absBasePath := filepath.Join(workTreeDir, opts.PathScope)
	exist, err := util.FileExists(absBasePath)
	if err != nil {
		return nil, fmt.Errorf("file exists %s failed: %s", absBasePath, err)
	}

	if !exist {
		return nil, fmt.Errorf("base path %s entry not found repo", opts.PathScope)
	}

	info, err := os.Lstat(absBasePath)
	if err != nil {
		return nil, fmt.Errorf("lstat %s failed: %s", absBasePath, err)
	}

	if info.IsDir() {
		desc.Type = DirectoryArchive

		if debugArchive() {
			logboek.Context(ctx).Debug().LogF("Found BasePath %s directory: directory archive type\n", absBasePath)
		}
	} else {
		desc.Type = FileArchive

		if debugArchive() {
			logboek.Context(ctx).Debug().LogF("Found BasePath %s file: file archive\n", absBasePath)
		}
	}

	repoHandle, err := repo_handle.NewHandle(repository)
	if err != nil {
		return nil, err
	}

	tw := tar.NewWriter(out)
	logProcess := logboek.Context(ctx).Debug().LogProcess("ls-tree (%s)", opts.PathMatcher.String())
	logProcess.Start()
	result, err := ls_tree.LsTree(ctx, repoHandle, opts.Commit, ls_tree.LsTreeOptions{
		PathScope:   opts.PathScope,
		PathMatcher: opts.PathMatcher,
		AllFiles:    true,
	})
	if err != nil {
		logProcess.Fail()
		return nil, err
	}
	logProcess.End()

	logProcess = logboek.Context(ctx).Debug().LogProcess("ls-tree result walk (%s)", opts.PathMatcher.String())
	logProcess.Start()
	if err := result.Walk(func(lsTreeEntry *ls_tree.LsTreeEntry) error {
		logboek.Context(ctx).Debug().LogF("ls-tree entry %s\n", lsTreeEntry.FullFilepath)

		desc.IsEmpty = false

		gitFileMode := lsTreeEntry.Mode
		absFilepath := filepath.Join(workTreeDir, lsTreeEntry.FullFilepath)

		var relToBasePathFilepath string
		if filepath.FromSlash(opts.PathScope) == lsTreeEntry.FullFilepath {
			// lsTreeEntry.FullFilepath is always a path to a file, not a directory.
			// Thus if opts.PathScope is equal lsTreeEntry.FullFilepath, then opts.PathScope is a path to a file.
			// Use file name in this case by convention.
			relToBasePathFilepath = filepath.Base(lsTreeEntry.FullFilepath)
		} else {
			relToBasePathFilepath = util.GetRelativeToBaseFilepath(opts.PathScope, lsTreeEntry.FullFilepath)
		}

		tarEntryName := filepath.ToSlash(relToBasePathFilepath)
		info, err := os.Lstat(absFilepath)
		if err != nil {
			return fmt.Errorf("lstat %s failed: %s", absFilepath, err)
		}

		switch gitFileMode {
		case filemode.Regular, filemode.Executable, filemode.Deprecated:
			err = tw.WriteHeader(&tar.Header{
				Format:     tar.FormatGNU,
				Name:       tarEntryName,
				Mode:       int64(gitFileMode),
				Size:       info.Size(),
				ModTime:    info.ModTime(),
				AccessTime: info.ModTime(),
				ChangeTime: info.ModTime(),
			})
			if err != nil {
				return fmt.Errorf("unable to write tar header for file %s: %s", tarEntryName, err)
			}

			f, err := os.Open(absFilepath)
			if err != nil {
				return fmt.Errorf("unable to open file %s: %s", absFilepath, err)
			}

			_, err = io.Copy(tw, f)
			if err != nil {
				return fmt.Errorf("unable to write data to tar archive from file %s: %s", absFilepath, err)
			}

			err = f.Close()
			if err != nil {
				return fmt.Errorf("error closing file %s: %s", absFilepath, err)
			}

			if debugArchive() {
				logboek.Context(ctx).Debug().LogF("Added archive file %q\n", relToBasePathFilepath)
			}
		case filemode.Symlink:
			isSymlink := info.Mode()&os.ModeSymlink != 0

			var linkname string
			if isSymlink {
				linkname, err = os.Readlink(absFilepath)
				if err != nil {
					return fmt.Errorf("cannot read symlink %s: %s", absFilepath, err)
				}
			} else {
				data, err := ioutil.ReadFile(absFilepath)
				if err != nil {
					return fmt.Errorf("cannot read file %s: %s", absFilepath, err)
				}

				linkname = string(bytes.TrimSpace(data))
			}

			err = tw.WriteHeader(&tar.Header{
				Format:     tar.FormatGNU,
				Typeflag:   tar.TypeSymlink,
				Name:       tarEntryName,
				Linkname:   linkname,
				Mode:       int64(gitFileMode),
				Size:       info.Size(),
				ModTime:    info.ModTime(),
				AccessTime: info.ModTime(),
				ChangeTime: info.ModTime(),
			})
			if err != nil {
				return fmt.Errorf("unable to write tar symlink header for file %s: %s", tarEntryName, err)
			}

			if debugArchive() {
				logboek.Context(ctx).Debug().LogF("Added archive symlink %s -> %s\n", relToBasePathFilepath, linkname)
			}

			return nil
		default:
			panic(fmt.Sprintf("unexpected git file mode %s", gitFileMode.String()))
		}

		return nil
	}); err != nil {
		logProcess.Fail()
		return nil, err
	}
	logProcess.End()

	err = tw.Close()
	if err != nil {
		return nil, fmt.Errorf("cannot write tar archive: %s", err)
	}

	return desc, nil
}
