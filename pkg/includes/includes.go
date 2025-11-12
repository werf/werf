package includes

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"

	"github.com/werf/logboek"
	"github.com/werf/werf/v2/pkg/path_matcher"
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

type Include struct {
	repo       GitRepository
	commitHash string
	// `objects` is a map of destination path to original path
	// where the file was found in the remote repository
	// e.g. /path/to/file.txt (desired mount path) -> /path/to/remote/file.txt (original path in remote repo)
	// This is used to read the file from the remote repository
	objects map[string]string
}

func GetWerfIncludesConfigRelPath() string {
	return defaultIncludesConfigFileName
}

func GetWerfIncludesLockConfigRelPath() string {
	return defaultIncludesLockConfigFileName
}

type InitIncludesOptions struct {
	FileReader             GiterminismManagerFileReader
	ProjectDir             string
	CreateOrUpdateLockFile bool
	UseLatestVersion       bool
}

func Init(ctx context.Context, opts InitIncludesOptions) ([]*Include, error) {
	config, err := NewConfig(ctx, opts.FileReader, GetWerfIncludesConfigRelPath(), opts.CreateOrUpdateLockFile)
	if err != nil {
		return nil, fmt.Errorf("unable to initialize includes: %w", err)
	}

	if len(config.Includes) > 0 {
		lockFileProjectRelPath := filepath.Join(opts.ProjectDir, GetWerfIncludesLockConfigRelPath())
		lockFilePath := GetWerfIncludesLockConfigRelPath()
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
				includesLockPath: lockFileProjectRelPath,
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

func GetIncludes(ctx context.Context, cfg Config, lockInfo *LockInfo, remoteRepos *gitRepositoriesWithCache) ([]*Include, error) {
	includes := []*Include{}
	err := logboek.Context(ctx).Default().LogBlock("Initializing includes").DoError(func() error {
		for i := len(cfg.Includes) - 1; i >= 0; i-- {
			// Reverse order to prioritize the last include in the list
			inc := cfg.Includes[i]
			r, err := remoteRepos.getRepository(inc.Git)
			if err != nil {
				return fmt.Errorf("unable to find remote repository %s: %w", inc.Git, err)
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

				repo, err := r.repo.PlainOpen()
				if err != nil {
					return fmt.Errorf("failed to open repository: %w", err)
				}

				commit, err := repo.CommitObject(plumbing.NewHash(commitFromLockInfo))
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
					repo:       r.repo,
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
