package true_git

import (
	"archive/tar"
	"bytes"
	"context"
	"fmt"
	"io"
	"path/filepath"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/filemode"

	"github.com/werf/common-go/pkg/util"
	"github.com/werf/werf/v2/pkg/git_repo/repo_handle"
	"github.com/werf/werf/v2/pkg/true_git/ls_tree"
)

func GoGitArchive(ctx context.Context, out io.Writer, gitDir string, opts ArchiveOptions) error {
	if out == nil {
		out = io.Discard
	}

	absGitDir, err := filepath.Abs(gitDir)
	if err != nil {
		return fmt.Errorf("bad git dir %s: %w", gitDir, err)
	}

	repository, err := PlainOpenWithOptions(absGitDir, &PlainOpenOptions{EnableDotGitCommonDir: true})
	if err != nil {
		return fmt.Errorf("cannot open git repo %q: %w", absGitDir, err)
	}

	repoHandle, err := repo_handle.NewHandle(repository)
	if err != nil {
		return err
	}

	return goGitWriteArchive(ctx, out, repoHandle, repository, opts)
}

func GoGitArchiveWithSubmodules(ctx context.Context, out io.Writer, gitDir, workTreeDir string, opts ArchiveOptions) error {
	if out == nil {
		out = io.Discard
	}

	absGitDir, err := filepath.Abs(gitDir)
	if err != nil {
		return fmt.Errorf("bad git dir %s: %w", gitDir, err)
	}

	absWorkTreeDir, err := filepath.Abs(workTreeDir)
	if err != nil {
		return fmt.Errorf("bad work tree dir %s: %w", workTreeDir, err)
	}

	repository, err := PlainOpenWithOptions(absGitDir, &PlainOpenOptions{EnableDotGitCommonDir: true})
	if err != nil {
		return fmt.Errorf("cannot open git repo %q: %w", absGitDir, err)
	}

	commitHash, err := parseCommitHash(opts.Commit)
	if err != nil {
		return fmt.Errorf("bad commit hash %q: %w", opts.Commit, err)
	}
	if commitHash.IsZero() {
		return fmt.Errorf("bad commit hash %q", opts.Commit)
	}

	repoHandle, err := repo_handle.NewHandle(repository, repo_handle.NewHandleOptions{CommitHash: commitHash, WorkTreeDir: absWorkTreeDir})
	if err != nil {
		return err
	}

	return goGitWriteArchive(ctx, out, repoHandle, repository, opts)
}

func goGitWriteArchive(ctx context.Context, out io.Writer, repoHandle repo_handle.Handle, repository *git.Repository, opts ArchiveOptions) error {
	commitHash, err := parseCommitHash(opts.Commit)
	if err != nil {
		return fmt.Errorf("bad commit hash %q: %w", opts.Commit, err)
	}
	if commitHash.IsZero() {
		return fmt.Errorf("bad commit hash %q", opts.Commit)
	}

	commitObj, err := repository.CommitObject(commitHash)
	if err != nil {
		return fmt.Errorf("get commit %s: %w", commitHash, err)
	}
	commitTime := commitObj.Author.When

	result, err := ls_tree.LsTree(ctx, repoHandle, opts.Commit, ls_tree.LsTreeOptions{
		PathScope:   opts.PathScope,
		PathMatcher: opts.PathMatcher,
		AllFiles:    true,
	})
	if err != nil {
		return err
	}
	if result.IsEmpty() {
		return fmt.Errorf("lstree result is empty when writing tar archive. PathScope: %q. PathMatcher configuration: %q", opts.PathScope, opts.PathMatcher)
	}

	tw := tar.NewWriter(out)
	creadedDirEntries := make(map[string]bool)
	if err := result.Walk(func(lsTreeEntry *ls_tree.LsTreeEntry) error {
		if lsTreeEntry.FullFilepath == "" {
			return nil
		}

		var tarEntryName string
		if renameToFileName, willRename := opts.FileRenames[filepath.ToSlash(filepath.Clean(lsTreeEntry.FullFilepath))]; willRename {
			tarEntryName = renameToFileName
		} else {
			tarEntryName = filepath.ToSlash(util.GetRelativeToBaseFilepath(opts.PathScope, lsTreeEntry.FullFilepath))
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

				header := &tar.Header{
					Format:     tar.FormatGNU,
					Name:       p,
					Typeflag:   tar.TypeDir,
					Mode:       0o775,
					ModTime:    commitTime,
					AccessTime: commitTime,
					ChangeTime: commitTime,
				}
				applyOwnership(header, opts)

				if err := tw.WriteHeader(header); err != nil {
					return fmt.Errorf("unable to write tar header for dir %q: %w", p, err)
				}

				creadedDirEntries[p] = true
			}
		}

		content, err := result.LsTreeEntryContent(repoHandle, lsTreeEntry.FullFilepath)
		if err != nil {
			return err
		}

		switch lsTreeEntry.Mode {
		case filemode.Regular, filemode.Executable, filemode.Deprecated:
			header := &tar.Header{
				Format:     tar.FormatGNU,
				Name:       tarEntryName,
				Mode:       int64(lsTreeEntry.Mode),
				Size:       int64(len(content)),
				ModTime:    commitTime,
				AccessTime: commitTime,
				ChangeTime: commitTime,
			}
			applyOwnership(header, opts)

			if err := tw.WriteHeader(header); err != nil {
				return fmt.Errorf("unable to write tar header for file %q: %w", tarEntryName, err)
			}

			if _, err := io.Copy(tw, bytes.NewReader(content)); err != nil {
				return fmt.Errorf("unable to write data to tar archive from file %s: %w", tarEntryName, err)
			}
		case filemode.Symlink:
			linkname := string(bytes.TrimSpace(content))
			header := &tar.Header{
				Format:     tar.FormatGNU,
				Typeflag:   tar.TypeSymlink,
				Name:       tarEntryName,
				Linkname:   linkname,
				Mode:       int64(lsTreeEntry.Mode),
				Size:       int64(len(linkname)),
				ModTime:    commitTime,
				AccessTime: commitTime,
				ChangeTime: commitTime,
			}
			applyOwnership(header, opts)

			if err := tw.WriteHeader(header); err != nil {
				return fmt.Errorf("unable to write tar symlink header for file %s: %w", tarEntryName, err)
			}
		default:
			panic(fmt.Sprintf("unexpected git file mode %s", lsTreeEntry.Mode.String()))
		}

		return nil
	}); err != nil {
		return err
	}

	if err := tw.Close(); err != nil {
		return fmt.Errorf("cannot write tar archive: %w", err)
	}

	return nil
}
