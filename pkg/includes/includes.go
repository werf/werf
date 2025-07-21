package includes

import (
	"context"
	"fmt"
	"path"
	"path/filepath"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"

	"github.com/werf/logboek"
	"github.com/werf/werf/v2/pkg/git_repo"
	"github.com/werf/werf/v2/pkg/path_matcher"
	"github.com/werf/werf/v2/pkg/true_git"
)

const (
	defaultIncludesConfigFileName     = "werf-includes.yaml"
	defaultIncludesLockConfigFileName = "werf-includes.lock"
)

type GiterminismManagerFileReader interface {
	IsIncludesConfigExistAnywhere(ctx context.Context, relPath string) (bool, error)
	ReadIncludesConfig(ctx context.Context, relPath string) ([]byte, error)
	ReadIncludesLockFile(ctx context.Context, relPath string) ([]byte, error)
}

type GitRepository interface {
	GetName() string
	ReadCommitFile(ctx context.Context, commit, path string) (data []byte, err error)
	IsCommitDirectoryExist(ctx context.Context, commit, path string) (exist bool, err error)
	IsCommitFileExist(ctx context.Context, commit, path string) (exist bool, err error)
}

type Include struct {
	repo       GitRepository
	commitHash string
	// `objects` is a map of destination path to original path
	// where the file was found in the remote repository
	// e.g. /path/to/file.txt (desired mount path) -> /path/to/remote/file.txt (original path in remote repo)
	// This is used to read the file from the remote repository
	objects map[string]string
}

func GetWerfIncludesConfigRelPath(path string) string {
	if path == "" {
		return defaultIncludesConfigFileName
	}
	return filepath.ToSlash(path)
}

func GetWerfIncludesLockConfigRelPath(path string) string {
	if path == "" {
		return defaultIncludesLockConfigFileName
	}
	return filepath.ToSlash(path)
}

type InitIncludesOptions struct {
	FileReader             GiterminismManagerFileReader
	ConfigRelPath          string
	LockFileRelPath        string
	CreateOrUpdateLockFile bool
	UseLatestVersion       bool
}

func Init(ctx context.Context, opts InitIncludesOptions) ([]*Include, error) {
	config, err := NewConfig(ctx, opts.FileReader, GetWerfIncludesConfigRelPath(opts.ConfigRelPath), opts.CreateOrUpdateLockFile)
	if err != nil {
		return nil, fmt.Errorf("unable to initialize includes: %w", err)
	}

	if len(config.Includes) > 0 {

		lockFilePath := GetWerfIncludesLockConfigRelPath(opts.LockFileRelPath)

		var lockConfig *lockConfig
		if !opts.CreateOrUpdateLockFile && !opts.UseLatestVersion {
			lockConfig, err = parseLockConfig(ctx, opts.FileReader, lockFilePath)
			if err != nil {
				return nil, err
			}
		}

		remoteRepos, err := initRemoteRepos(ctx, config)
		if err != nil {
			return nil, fmt.Errorf("unable to initialize remote repositories: %w", err)
		}

		if opts.CreateOrUpdateLockFile {
			if err := CreateOrUpdateLockConfig(ctx, createLockConfigOptions{
				fileReader:       opts.FileReader,
				includesConfig:   config,
				includesLockPath: lockFilePath,
				remoteRepos:      remoteRepos,
			}); err != nil {
				return nil, fmt.Errorf("create or update werf-includes.lock: %w", err)
			}
			return nil, nil
		}

		lockInfo, err := getLockInfo(getLockInfoOptions{
			includesConfig:         config,
			fileReader:             opts.FileReader,
			createOrUpdateLockFile: opts.CreateOrUpdateLockFile,
			useLatestVersion:       opts.UseLatestVersion,
			remoteRepos:            remoteRepos,
			lockConfig:             lockConfig,
		})
		if err != nil {
			return nil, err
		}

		includes, err := GetIncludes(ctx, config, lockInfo, remoteRepos)
		if err != nil {
			return nil, fmt.Errorf("unable to get includes: %w", err)
		}
		return includes, nil
	}

	return []*Include{}, nil
}
func initRemoteRepos(ctx context.Context, cfg Config) (map[string]*git_repo.Remote, error) {
	repoCache := make(map[string]*git_repo.Remote)

	err := logboek.Context(ctx).Default().LogBlock("Initializing remote repositories").DoError(func() error {
		for _, i := range cfg.Includes {
			if _, ok := repoCache[i.Git]; !ok {
				basicAuth, err := git_repo.BasicAuthCredentialsHelper(&git_repo.BasicAuthCredentials{
					Username: i.BasicAuth.Username,
					Password: i.BasicAuth.Password,
				})
				if err != nil {
					return fmt.Errorf("unable to get basic auth for repository %s: %w", i.Git, err)
				}
				repo, err := git_repo.OpenRemoteRepo(i.Git, i.Git, basicAuth)
				if err != nil {
					return fmt.Errorf("unable to open remote repository %s: %w", i.Git, err)
				}

				isCloned, err := repo.Clone(ctx)
				if err != nil {
					return fmt.Errorf("unable to clone %s repository: %w", i.Git, err)
				}

				repoCache[i.Git] = repo

				if isCloned {
					continue
				}

				err = logboek.Context(ctx).Default().LogProcess(fmt.Sprintf("Syncing origin branches and tags for: %s", i.Git)).DoError(func() error {
					fetchOptions := true_git.FetchOptions{
						Prune:     true,
						PruneTags: true,
						RefSpecs: map[string][]string{
							"origin": {
								"+refs/heads/*:refs/heads/*",
								"+refs/tags/*:refs/tags/*",
							},
						},
						UpdateHeadOk: true,
					}

					if err := true_git.Fetch(ctx, repo.GetClonePath(), fetchOptions); err != nil {
						return fmt.Errorf("fetch failed: %w", err)
					}

					return nil
				})
				if err != nil {
					return fmt.Errorf("unable to sync origin branches and tags for %s: %w", i.Git, err)
				}
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return repoCache, nil
}

func GetIncludes(ctx context.Context, cfg Config, lockInfo *LockInfo, remoteRepos map[string]*git_repo.Remote) ([]*Include, error) {
	includes := []*Include{}
	err := logboek.Context(ctx).Default().LogBlock("Initializing includes").DoError(func() error {
		for i := len(cfg.Includes) - 1; i >= 0; i-- {
			// Reverse order to prioritize the last include in the list
			inc := cfg.Includes[i]
			remoteRepo, ok := remoteRepos[inc.Git]
			if !ok || remoteRepo == nil {
				return fmt.Errorf("unable to find remote repository %s", inc.Git)
			}

			ref, err := inc.Ref()
			if err != nil {
				return err
			}

			err = logboek.Context(ctx).Default().LogProcess(fmt.Sprintf("Processing include %s with ref %s", inc.Git, ref)).DoError(func() error {
				commitFromLockInfo, err := lockInfo.GetCommit(inc.Git, ref)
				if err != nil {
					return fmt.Errorf("unable to get commit from lock info: %w", err)
				}

				r, err := remoteRepo.PlainOpen()
				if err != nil {
					return fmt.Errorf("failed to open repository: %w", err)
				}

				commit, err := r.CommitObject(plumbing.NewHash(commitFromLockInfo))
				if err != nil {
					return fmt.Errorf("failed to get commit object: %w", err)
				}

				tree, err := commit.Tree()
				if err != nil {
					return fmt.Errorf("failed to get tree: %w", err)
				}

				pm := path_matcher.NewPathMatcher(path_matcher.PathMatcherOptions{
					BasePath:     inc.Add,
					IncludeGlobs: inc.IncludePaths,
					ExcludeGlobs: inc.ExcludePaths,
				})

				logboek.Context(ctx).Debug().LogF("Using path matcher: basePath=%s, includeGlobs=%v, excludeGlobs=%v\n", inc.Add, inc.IncludePaths, inc.ExcludePaths)

				matchedMap := map[string]string{}
				err = tree.Files().ForEach(func(f *object.File) error {
					if pm.IsPathMatched(f.Name) {
						newPath := prepareRelPath(f.Name, inc.Add, inc.To)
						matchedMap[newPath] = f.Name
					}
					return nil
				})
				if err != nil {
					return fmt.Errorf("failed to iterate over files: %w", err)
				}

				if len(matchedMap) == 0 {
					return fmt.Errorf("no files matched for include %s with ref %s", inc.Git, ref)
				}

				include := &Include{
					repo:       remoteRepo,
					commitHash: commit.Hash.String(),
					objects:    matchedMap,
				}

				includes = append(includes, include)

				logboek.Context(ctx).Debug().LogF("Include initialized: repo: %s commit: %s\n", include.repo, include.commitHash)
				return nil
			})
			if err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return includes, nil
}

func (i *Include) GetName() string {
	if i.repo == nil {
		return ""
	}
	return i.repo.GetName()
}

func (i *Include) WalkObjects(fn func(toPath, origPath string) error) error {
	for toPath, origPath := range i.objects {
		if err := fn(toPath, origPath); err != nil {
			return err
		}
	}
	return nil
}

func (i *Include) GetFile(ctx context.Context, relPath string) ([]byte, error) {
	filePath, ok := i.objects[relPath]
	if !ok {
		return nil, fmt.Errorf("file not found in include: %s", relPath)
	}

	data, err := i.repo.ReadCommitFile(ctx, i.commitHash, filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read commit file: %w", err)
	}
	return data, nil
}

func (i *Include) GetFilesByGlob(ctx context.Context, pattern string) (map[string][]byte, error) {
	result := make(map[string][]byte)

	pm := path_matcher.NewPathMatcher(path_matcher.PathMatcherOptions{
		IncludeGlobs: []string{pattern},
	})

	for relPath := range i.objects {
		if pm.IsPathMatched(relPath) {
			data, err := i.GetFile(ctx, relPath)
			if err != nil {
				return nil, fmt.Errorf("failed to read file %q: %w", relPath, err)
			}
			result[relPath] = data
		}
	}

	return result, nil
}

func ListFilesByGlobs(ctx context.Context, includes []*Include, globs, sources []string) map[string]*Include {
	var filterSources bool = len(sources) > 0
	result := make(map[string]*Include)

	for _, i := range includes {
		if filterSources && !sliceContainsSubstring(i.GetName(), sources) {
			continue
		}
		list := i.ListFilesByGlobs(ctx, globs)
		for _, l := range list {
			if _, ok := result[l]; !ok {
				result[l] = i
			}
		}
	}

	return result
}

func sliceContainsSubstring(s string, substrings []string) bool {
	for _, sub := range substrings {
		if strings.Contains(s, sub) {
			return true
		}
	}
	return false
}

func (i *Include) ListFilesByGlobs(ctx context.Context, patterns []string) []string {
	result := make([]string, 0, len(i.objects))

	pm := path_matcher.NewPathMatcher(path_matcher.PathMatcherOptions{
		IncludeGlobs: patterns,
	})

	for relPath := range i.objects {
		if pm.IsPathMatched(relPath) {
			result = append(result, relPath)
		}
	}
	return result
}

func IsDirExists(ctx context.Context, includes []*Include, relPath string) bool {
	for _, i := range includes {
		if i == nil {
			continue
		}
		exists, _ := i.repo.IsCommitDirectoryExist(ctx, i.commitHash, relPath)
		if exists {
			return true
		}
	}
	return false
}

func IsFileExists(ctx context.Context, includes []*Include, relPath string) bool {
	for _, i := range includes {
		if i == nil {
			continue
		}
		exists, _ := i.repo.IsCommitFileExist(ctx, i.commitHash, relPath)
		if exists {
			return true
		}
	}
	return false
}

var ErrConfigFileNotFound = fmt.Errorf("config file not found")

func FindWerfConfig(ctx context.Context, includes []*Include, cfgPaths []string) (string, []byte, error) {
	for _, include := range includes {
		if include == nil {
			continue
		}
		for _, cfgPath := range cfgPaths {
			if _, ok := include.objects[cfgPath]; ok {
				data, err := include.GetFile(ctx, cfgPath)
				if err != nil {
					return "", nil, fmt.Errorf("unable to read config file %q: %w", cfgPath, err)
				}
				logboek.Context(ctx).Debug().LogF("Found config file %q in %q\n", cfgPath, include.repo.GetName())
				return cfgPath, data, nil
			}
		}
	}
	return "", nil, ErrConfigFileNotFound
}

func (i *includeConf) Ref() (string, error) {
	return ref(i.Git, i.Commit, i.Tag, i.Branch)
}

func getCommit(r *git.Repository, git, tag, branch, commit string) (*object.Commit, error) {
	switch {
	case commit != "":
		return commitRef(r, commit)
	case tag != "":
		return tagRef(r, tag)
	case branch != "":
		return branchRef(r, branch)
	default:
		return nil, fmt.Errorf("no commit, tag or branch specified for include %s", git)
	}
}

func commitRef(r *git.Repository, commit string) (*object.Commit, error) {
	rev := plumbing.Revision(commit)
	h, err := r.ResolveRevision(rev)
	if err != nil {
		return nil, fmt.Errorf("cannot resolve commit %s: %w", commit, err)
	}
	return r.CommitObject(*h)
}

func tagRef(r *git.Repository, tag string) (*object.Commit, error) {
	tagRef, err := r.Reference(plumbing.NewTagReferenceName(tag), true)
	if err != nil {
		return nil, fmt.Errorf("failed to get tag %s: %w", tag, err)
	}
	return commitRef(r, tagRef.Hash().String())
}

func branchRef(r *git.Repository, branch string) (*object.Commit, error) {
	branchRef, err := r.Reference(plumbing.NewBranchReferenceName(branch), true)
	if err != nil {
		return nil, fmt.Errorf("failed to get branch %s: %w", branch, err)
	}
	return commitRef(r, branchRef.Hash().String())
}

func prepareRelPath(fileName, add, to string) string {
	addClean := strings.TrimPrefix(filepath.Clean(add), string(filepath.Separator))
	relPath := strings.TrimPrefix(fileName, addClean)
	relPath = strings.TrimPrefix(relPath, string(filepath.Separator))
	newPath := path.Join(to, relPath)
	newPath = strings.TrimPrefix(newPath, string(filepath.Separator))
	return newPath
}
