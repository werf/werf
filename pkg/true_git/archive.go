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
	Commit      string
	PathScope   string // Determines the directory that will get into the result (similar to <pathspec> in the git commands).
	PathMatcher path_matcher.PathMatcher
	FileRenames map[string]string // Files to rename during archiving. Git repo relative paths of original files as keys, new filenames (without base path) as values.
}

// TODO: 1.3 add git mapping type (dir, file, ...) to gitArchive stage digest
func (opts ArchiveOptions) ID() string {
	var renamedOldFilePaths, renamedNewFileNames []string
	for renamedOldFilePath, renamedNewFileName := range opts.FileRenames {
		renamedOldFilePaths = append(renamedOldFilePaths, renamedOldFilePath)
		renamedNewFileNames = append(renamedNewFileNames, renamedNewFileName)
	}

	return util.Sha256Hash(
		append(
			append(renamedOldFilePaths, renamedNewFileNames...),
			opts.Commit,
			opts.PathScope,
			opts.PathMatcher.ID(),
		)...
	)
}

type ArchiveDescriptor struct {
	IsEmpty bool
}

func ArchiveWithSubmodules(ctx context.Context, out io.Writer, gitDir, workTreeCacheDir string, opts ArchiveOptions) error {
	return withWorkTreeCacheLock(ctx, workTreeCacheDir, func() error {
		return writeArchive(ctx, out, gitDir, workTreeCacheDir, true, opts)
	})
}

func Archive(ctx context.Context, out io.Writer, gitDir, workTreeCacheDir string, opts ArchiveOptions) error {
	return withWorkTreeCacheLock(ctx, workTreeCacheDir, func() error {
		return writeArchive(ctx, out, gitDir, workTreeCacheDir, false, opts)
	})
}

func debugArchive() bool {
	return os.Getenv("WERF_TRUE_GIT_DEBUG_ARCHIVE") == "1"
}

func writeArchive(ctx context.Context, out io.Writer, gitDir, workTreeCacheDir string, withSubmodules bool, opts ArchiveOptions) error {
	var err error

	gitDir, err = filepath.Abs(gitDir)
	if err != nil {
		return fmt.Errorf("bad git dir %s: %s", gitDir, err)
	}

	workTreeCacheDir, err = filepath.Abs(workTreeCacheDir)
	if err != nil {
		return fmt.Errorf("bad work tree cache dir %s: %s", workTreeCacheDir, err)
	}

	workTreeDir, err := prepareWorkTree(ctx, gitDir, workTreeCacheDir, opts.Commit, withSubmodules)
	if err != nil {
		return fmt.Errorf("cannot prepare work tree in cache %s for commit %s: %s", workTreeCacheDir, opts.Commit, err)
	}

	repository, err := GitOpenWithCustomWorktreeDir(gitDir, workTreeDir)
	if err != nil {
		return fmt.Errorf("git open failed: %s", err)
	}

	repoHandle, err := repo_handle.NewHandle(repository)
	if err != nil {
		return err
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
		return err
	}
	if result.IsEmpty() {
		logProcess.Fail()
		return fmt.Errorf("lstree result is empty when writing tar archive. PathScope: %q. PathMatcher configuration: %q", opts.PathScope, opts.PathMatcher)
	}
	logProcess.End()

	logProcess = logboek.Context(ctx).Debug().LogProcess("ls-tree result walk (%s)", opts.PathMatcher.String())
	logProcess.Start()
	if err := result.Walk(func(lsTreeEntry *ls_tree.LsTreeEntry) error {
		logboek.Context(ctx).Debug().LogF("ls-tree entry %s\n", lsTreeEntry.FullFilepath)

		gitFileMode := lsTreeEntry.Mode
		absFilepath := filepath.Join(workTreeDir, lsTreeEntry.FullFilepath)
	
		var tarEntryName string
		if renameToFileName, willRename := opts.FileRenames[filepath.ToSlash(filepath.Clean(lsTreeEntry.FullFilepath))]; willRename {
			tarEntryName = renameToFileName
		} else {
			tarEntryName = filepath.ToSlash(util.GetRelativeToBaseFilepath(opts.PathScope, lsTreeEntry.FullFilepath))
		}

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
				logboek.Context(ctx).Debug().LogF("Added archive file %q\n", tarEntryName)
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
				logboek.Context(ctx).Debug().LogF("Added archive symlink %s -> %s\n", tarEntryName, linkname)
			}

			return nil
		default:
			panic(fmt.Sprintf("unexpected git file mode %s", gitFileMode.String()))
		}

		return nil
	}); err != nil {
		logProcess.Fail()
		return err
	}
	logProcess.End()

	err = tw.Close()
	if err != nil {
		return fmt.Errorf("cannot write tar archive: %s", err)
	}

	return nil
}
