package true_git

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/Masterminds/semver"
	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-billy/v5/osfs"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/cache"
	"github.com/go-git/go-git/v5/storage/filesystem"
	"github.com/go-git/go-git/v5/storage/filesystem/dotgit"
	"github.com/go-git/go-git/v5/utils/ioutil"

	"github.com/werf/common-go/pkg/util"
)

func GitOpenWithCustomWorktreeDir(gitDir, worktreeDir string) (*git.Repository, error) {
	return PlainOpenWithOptions(worktreeDir, &PlainOpenOptions{EnableDotGitCommonDir: true})
}

type BasicAuth struct {
	Username string
	Password string
}

type FetchOptions struct {
	All                  bool
	TagsOnly             bool
	Prune                bool
	PruneTags            bool
	Unshallow            bool
	UpdateHeadOk         bool
	RefSpecs             map[string][]string
	BasicAuthCredentials *BasicAuth
}

func IsShallowFileChangedSinceWeReadIt(err error) bool {
	return err != nil && strings.Contains(err.Error(), "shallow file has changed since we read it")
}

func Fetch(ctx context.Context, path string, options FetchOptions) error {
	commandArgs := []string{"fetch"}

	if options.Unshallow {
		commandArgs = append(commandArgs, "--unshallow")
	}

	if options.All {
		commandArgs = append(commandArgs, "--all")
	}

	if options.TagsOnly {
		commandArgs = append(commandArgs, "--tags")
	}

	if options.UpdateHeadOk {
		commandArgs = append(commandArgs, "--update-head-ok")
	}

	if options.Prune || options.PruneTags {
		commandArgs = append(commandArgs, "--prune")

		if options.PruneTags && !gitVersion.LessThan(semver.MustParse("2.17.0")) {
			commandArgs = append(commandArgs, "--prune-tags")
		}
	}

	for remote, refSpec := range options.RefSpecs {
		remoteRefSpecs := append([]string{remote}, refSpec...)
		commandArgs = append(commandArgs, remoteRefSpecs...)
	}

	gitCmd := NewGitCmd(ctx, &GitCmdOptions{RepoDir: path}, commandArgs...)

	return gitCmd.Run(ctx)
}

func GetLastBranchCommitSHA(ctx context.Context, repoPath, branch string) (string, error) {
	revParseCmd := NewGitCmd(ctx, &GitCmdOptions{RepoDir: repoPath}, "rev-parse", branch)
	if err := revParseCmd.Run(ctx); err != nil {
		return "", fmt.Errorf("git rev parse branch command failed: %w", err)
	}

	return strings.TrimSpace(revParseCmd.OutBuf.String()), nil
}

func IsShallowClone(ctx context.Context, path string) (bool, error) {
	if gitVersion.LessThan(semver.MustParse("2.15.0")) {
		exist, err := util.FileExists(filepath.Join(path, ".git", "shallow"))
		if err != nil {
			return false, err
		}

		return exist, nil
	}

	checkShallowCmd := NewGitCmd(ctx, &GitCmdOptions{RepoDir: path}, "rev-parse", "--is-shallow-repository")
	if err := checkShallowCmd.Run(ctx); err != nil {
		return false, fmt.Errorf("git shallow repository check command failed: %w", err)
	}

	return strings.TrimSpace(checkShallowCmd.OutBuf.String()) == "true", nil
}

type PlainOpenOptions struct {
	// DetectDotGit defines whether parent directories should be
	// walked until a .git directory or file is found.
	DetectDotGit bool
	// Enable .git/commondir support (see https://git-scm.com/docs/gitrepository-layout#Documentation/gitrepository-layout.txt).
	// NOTE: This option will only work with the filesystem storage.
	EnableDotGitCommonDir bool
}

func PlainOpenWithOptions(path string, o *PlainOpenOptions) (*git.Repository, error) {
	dot, wt, err := dotGitToOSFilesystems(path, o.DetectDotGit)
	if err != nil {
		return nil, err
	}

	if _, err := dot.Stat(""); err != nil {
		if os.IsNotExist(err) {
			return nil, git.ErrRepositoryNotExists
		}

		return nil, err
	}

	var repositoryFs billy.Filesystem

	if o.EnableDotGitCommonDir {
		dotGitCommon, err := dotGitCommonDirectory(dot)
		if err != nil {
			return nil, err
		}
		repositoryFs = dotgit.NewRepositoryFilesystem(dot, dotGitCommon)
	} else {
		repositoryFs = dot
	}

	s := filesystem.NewStorageWithOptions(repositoryFs, cache.NewObjectLRUDefault(), filesystem.Options{
		AlternatesFS: osfs.New("/", osfs.WithBoundOS()),
	})

	return git.Open(s, wt)
}

func dotGitToOSFilesystems(path string, detect bool) (dot, wt billy.Filesystem, err error) {
	path, err = util.ExpandPath(path)
	if err != nil {
		return nil, nil, err
	}

	var fs billy.Filesystem
	var fi os.FileInfo
	for {
		fs = osfs.New(path)

		pathinfo, err := fs.Stat("/")
		if !os.IsNotExist(err) {
			if pathinfo == nil {
				return nil, nil, err
			}
			if !pathinfo.IsDir() && detect {
				fs = osfs.New(filepath.Dir(path))
			}
		}

		fi, err = fs.Stat(git.GitDirName)
		if err == nil {
			// no error; stop
			break
		}
		if !os.IsNotExist(err) {
			// unknown error; stop
			return nil, nil, err
		}
		if detect {
			// try its parent as long as we haven't reached
			// the root dir
			if dir := filepath.Dir(path); dir != path {
				path = dir
				continue
			}
		}
		// not detecting via parent dirs and the dir does not exist;
		// stop
		return fs, nil, nil
	}

	if fi.IsDir() {
		dot, err = fs.Chroot(git.GitDirName)
		return dot, fs, err
	}

	dot, err = dotGitFileToOSFilesystem(path, fs)
	if err != nil {
		return nil, nil, err
	}

	return dot, fs, nil
}

func dotGitFileToOSFilesystem(path string, fs billy.Filesystem) (bfs billy.Filesystem, err error) {
	f, err := fs.Open(git.GitDirName)
	if err != nil {
		return nil, err
	}
	defer ioutil.CheckClose(f, &err)

	b, err := io.ReadAll(f)
	if err != nil {
		return nil, err
	}

	line := string(b)
	const prefix = "gitdir: "
	if !strings.HasPrefix(line, prefix) {
		return nil, fmt.Errorf(".git file has no %s prefix", prefix)
	}

	gitdir := strings.Split(line[len(prefix):], "\n")[0]
	gitdir = strings.TrimSpace(gitdir)
	if filepath.IsAbs(gitdir) {
		return osfs.New(gitdir), nil
	}

	return osfs.New(fs.Join(path, gitdir)), nil
}

func dotGitCommonDirectory(fs billy.Filesystem) (commonDir billy.Filesystem, err error) {
	f, err := fs.Open("commondir")
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	b, err := io.ReadAll(f)
	if err != nil {
		return nil, err
	}
	if len(b) > 0 {
		path := strings.TrimSpace(string(b))
		if filepath.IsAbs(path) {
			commonDir = osfs.New(path)
		} else {
			commonDir = osfs.New(filepath.Join(fs.Root(), path))
		}
		if _, err := commonDir.Stat(""); err != nil {
			if os.IsNotExist(err) {
				return nil, git.ErrRepositoryIncomplete
			}

			return nil, err
		}
	}

	return commonDir, nil
}
