package true_git

import (
	"archive/tar"
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"

	"github.com/go-git/go-git/v5/plumbing/filemode"

	"github.com/werf/common-go/pkg/util"
	"github.com/werf/logboek"
	"github.com/werf/werf/v2/pkg/git_repo/repo_handle"
	"github.com/werf/werf/v2/pkg/path_matcher"
	"github.com/werf/werf/v2/pkg/tmp_manager"
	"github.com/werf/werf/v2/pkg/true_git/ls_tree"
)

type ArchiveOptions struct {
	// ContentChecksum identifies selected Git files and checkout inputs independently of the commit and matcher.
	ContentChecksum string
	Commit          string
	PathScope       string // Determines the directory that will get into the result (similar to <pathspec> in the git commands).
	PathMatcher     path_matcher.PathMatcher
	FileRenames     map[string]string // Files to rename during archiving. Git repo relative paths of original files as keys, new filenames (without base path) as values.
	Owner           string
	Group           string
}

// TODO: 1.3 add git mapping type (dir, file, ...) to gitArchive stage digest
func (opts ArchiveOptions) ID() string {
	if opts.ContentChecksum != "" {
		args := []string{"dockerfile-context-v3", opts.ContentChecksum, opts.PathScope, opts.Owner, opts.Group}
		for _, path := range slices.Sorted(maps.Keys(opts.FileRenames)) {
			args = append(args, path, opts.FileRenames[path])
		}
		return util.Sha256Hash(args...)
	}

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
		)...,
	)
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
	return os.Getenv("WERF_DEBUG_TRUE_GIT_ARCHIVE") == "1" || os.Getenv("WERF_TRUE_GIT_DEBUG_ARCHIVE") == "1"
}

func writeArchive(ctx context.Context, out io.Writer, gitDir, workTreeCacheDir string, withSubmodules bool, opts ArchiveOptions) error {
	var err error

	gitDir, err = filepath.Abs(gitDir)
	if err != nil {
		return fmt.Errorf("bad git dir %s: %w", gitDir, err)
	}

	workTreeCacheDir, err = filepath.Abs(workTreeCacheDir)
	if err != nil {
		return fmt.Errorf("bad work tree cache dir %s: %w", workTreeCacheDir, err)
	}

	workTreeDir, err := prepareWorkTree(ctx, gitDir, workTreeCacheDir, opts.Commit, withSubmodules)
	if err != nil {
		return fmt.Errorf("cannot prepare work tree in cache %s for commit %s: %w", workTreeCacheDir, opts.Commit, err)
	}

	repository, err := GitOpenWithCustomWorktreeDir(gitDir, workTreeDir)
	if err != nil {
		return fmt.Errorf("git open failed: %w", err)
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

	if opts.ContentChecksum != "" {
		tmpDir, err := tmp_manager.TempDir(ctx, "git-context-")
		if err != nil {
			return fmt.Errorf("create Git context export directory: %w", err)
		}
		defer func() {
			if err := os.RemoveAll(tmpDir); err != nil {
				logboek.Context(ctx).Warn().LogF("Remove Git context export directory %q: %s\n", tmpDir, err)
			}
		}()
		absTmpDir, err := filepath.Abs(tmpDir)
		if err != nil {
			return fmt.Errorf("resolve Git context export directory: %w", err)
		}
		tmpDir = absTmpDir
		exportDir := filepath.Join(tmpDir, "files")
		if err := os.Mkdir(exportDir, 0o700); err != nil {
			return fmt.Errorf("create Git context files directory: %w", err)
		}

		var paths, entries []string
		if err := result.Walk(func(entry *ls_tree.LsTreeEntry) error {
			path := filepath.ToSlash(entry.FullFilepath)
			paths = append(paths, path)
			entries = append(entries, fmt.Sprintf("%o %s\t%s\x00", entry.Mode, entry.Hash.String(), path))
			return nil
		}); err != nil {
			return fmt.Errorf("collect Git context export paths: %w", err)
		}
		cmdOpts := &GitCmdOptions{RepoDir: workTreeDir, Env: []string{
			"GIT_ATTR_SOURCE=" + opts.Commit,
			"GIT_INDEX_FILE=" + filepath.Join(tmpDir, "index"),
		}}
		indexCmd := NewGitCmd(ctx, cmdOpts, "-c", "core.hooksPath="+os.DevNull, "update-index", "-z", "--index-info")
		indexCmd.Stdin = strings.NewReader(strings.Join(entries, ""))
		if err := indexCmd.Run(ctx); err != nil {
			return fmt.Errorf("prepare Git context export index: %w", err)
		}
		cmd := NewGitCmd(ctx, cmdOpts, "-c", "core.hooksPath="+os.DevNull, "checkout-index", "--prefix="+filepath.ToSlash(exportDir)+"/", "-z", "--stdin")
		cmd.Stdin = strings.NewReader(strings.Join(paths, "\x00") + "\x00")
		if err := cmd.Run(ctx); err != nil {
			return fmt.Errorf("export Git context files: %w", err)
		}
		workTreeDir = exportDir
	}

	logProcess = logboek.Context(ctx).Debug().LogProcess("ls-tree result walk (%s)", opts.PathMatcher.String())
	logProcess.Start()

	creadedDirEntries := make(map[string]bool)

	if err := result.Walk(func(lsTreeEntry *ls_tree.LsTreeEntry) error {
		if debugArchive() {
			logboek.Context(ctx).Debug().LogF("ls-tree entry %s\n", lsTreeEntry.FullFilepath)
		}

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
			return fmt.Errorf("lstat %q failed: %w", absFilepath, err)
		}

		dirEntry := filepath.Dir(tarEntryName)
		if dirEntry != "." {
			var p string
			for _, pathPart := range util.SplitFilepath(dirEntry) {
				if p == "" {
					p = pathPart
				} else {
					p = filepath.Join(p, pathPart)
				}

				if creadedDirEntries[p] {
					continue
				}

				if debugArchive() {
					fmt.Printf("[writeArchive] creating tar archive dir entry %q ...\n", p)
				}

				header := &tar.Header{
					Format:     tar.FormatGNU,
					Name:       p,
					Typeflag:   tar.TypeDir,
					Mode:       0o775,
					ModTime:    info.ModTime(),
					AccessTime: info.ModTime(),
					ChangeTime: info.ModTime(),
				}
				applyOwnership(header, opts)

				if err := tw.WriteHeader(header); err != nil {
					return fmt.Errorf("unable to write tar header for dir %q: %w", p, err)
				}

				creadedDirEntries[p] = true
			}
		}

		switch gitFileMode {
		case filemode.Regular, filemode.Executable, filemode.Deprecated:
			header := &tar.Header{
				Format:     tar.FormatGNU,
				Name:       tarEntryName,
				Mode:       int64(gitFileMode),
				Size:       info.Size(),
				ModTime:    info.ModTime(),
				AccessTime: info.ModTime(),
				ChangeTime: info.ModTime(),
			}

			applyOwnership(header, opts)

			err = tw.WriteHeader(header)
			if err != nil {
				return fmt.Errorf("unable to write tar header for file %q: %w", tarEntryName, err)
			}

			f, err := os.Open(absFilepath)
			if err != nil {
				return fmt.Errorf("unable to open file %s: %w", absFilepath, err)
			}

			_, err = io.Copy(tw, f)
			if err != nil {
				return fmt.Errorf("unable to write data to tar archive from file %s: %w", absFilepath, err)
			}

			err = f.Close()
			if err != nil {
				return fmt.Errorf("error closing file %s: %w", absFilepath, err)
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
					return fmt.Errorf("cannot read symlink %s: %w", absFilepath, err)
				}
			} else {
				data, err := ioutil.ReadFile(absFilepath)
				if err != nil {
					return fmt.Errorf("cannot read file %s: %w", absFilepath, err)
				}

				linkname = string(bytes.TrimSpace(data))
			}

			header := &tar.Header{
				Format:     tar.FormatGNU,
				Typeflag:   tar.TypeSymlink,
				Name:       tarEntryName,
				Linkname:   linkname,
				Mode:       int64(gitFileMode),
				Size:       info.Size(),
				ModTime:    info.ModTime(),
				AccessTime: info.ModTime(),
				ChangeTime: info.ModTime(),
			}

			applyOwnership(header, opts)

			err = tw.WriteHeader(header)
			if err != nil {
				return fmt.Errorf("unable to write tar symlink header for file %s: %w", tarEntryName, err)
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
		return fmt.Errorf("cannot write tar archive: %w", err)
	}

	return nil
}

func applyOwnership(header *tar.Header, opts ArchiveOptions) {
	if opts.Owner != "" {
		if uid, err := strconv.Atoi(opts.Owner); err == nil {
			header.Uid = uid
			header.Uname = ""
		} else {
			header.Uname = opts.Owner
		}
	}

	if opts.Group != "" {
		if gid, err := strconv.Atoi(opts.Group); err == nil {
			header.Gid = gid
			header.Gname = ""
		} else {
			header.Gname = opts.Group
		}
	}
}
