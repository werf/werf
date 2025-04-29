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
	"gopkg.in/yaml.v3"
)

const (
	defaultIncludesConfigFileName = "werf-includes.yaml"
)

type GiterminismManagerFileReader interface {
	IsIncludesConfigExistAnywhere(ctx context.Context, relPath string) (bool, error)
	ReadIncludesConfig(ctx context.Context, relPath string) ([]byte, error)
}

type Config struct {
	Includes []includeConf `json:"includes"`
}

type includeConf struct {
	Name         string   `yaml:"name"`
	Git          string   `yaml:"git"`
	Branch       string   `yaml:"branch"`
	Tag          string   `yaml:"tag"`
	Commit       string   `yaml:"commit"`
	Add          string   `yaml:"add,omitempty"`
	To           string   `yaml:"to,omitempty"`
	IncludePaths []string `yaml:"includePaths"`
	ExcludePaths []string `yaml:"excludePaths"`
}

type Include struct {
	Name    string
	repo    *git_repo.Remote
	commit  *object.Commit
	objects map[string]string
}

func GetWerfIncludesConfigRelPath(path string) string {
	if path == "" {
		return defaultIncludesConfigFileName
	}
	return filepath.ToSlash(path)
}

func Init(ctx context.Context, fileReader GiterminismManagerFileReader, configRelPath string) ([]*Include, error) {
	config, err := NewConfig(ctx, fileReader, GetWerfIncludesConfigRelPath(configRelPath))
	if err != nil {
		return nil, fmt.Errorf("unable to initialize includes: %w", err)
	}

	includes, err := GetIncludes(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("unable to get includes: %w", err)
	}

	for _, include := range includes {
		fmt.Println(&include)
	}

	return includes, nil
}

func NewConfig(ctx context.Context, fileReader GiterminismManagerFileReader, configRelPath string) (Config, error) {
	config := Config{}
	logboek.Context(ctx).Debug().LogF("Reading includes config from %q\n", configRelPath)
	exist, err := fileReader.IsIncludesConfigExistAnywhere(ctx, configRelPath)
	if err != nil {
		return config, err
	}

	if !exist {
		return config, nil
	}

	data, err := fileReader.ReadIncludesConfig(ctx, configRelPath)
	if err != nil {
		return config, err
	}

	if err := yaml.Unmarshal(data, &config); err != nil {
		return config, fmt.Errorf("the includes config validation failed: %w", err)
	}

	return config, err
}

func GetIncludes(ctx context.Context, cfg Config) ([]*Include, error) {
	repoChache := make(map[string]*git_repo.Remote)
	includes := []*Include{}
	for _, i := range cfg.Includes {
		var remoteRepo *git_repo.Remote
		if remote, ok := repoChache[i.Git]; !ok {
			repo, err := git_repo.OpenRemoteRepo(i.Name, i.Git)
			if err != nil {
				return nil, err
			}
			if err := repo.CloneAndFetch(ctx); err != nil {
				return nil, fmt.Errorf("unable to clone %s repository: %w", i.Git, err)
			}
			repoChache[repo.Url] = repo
			remoteRepo = repo
		} else {
			remoteRepo = remote
		}
		r, err := remoteRepo.PlainOpen()
		if err != nil {
			return nil, fmt.Errorf("failed to open repository: %w", err)
		}
		commit, err := getCommit(r, &i)
		if err != nil {
			return nil, fmt.Errorf("failed to get commit: %w", err)
		}

		tree, err := commit.Tree()
		if err != nil {
			return nil, fmt.Errorf("failed to get tree: %w", err)
		}

		pm := path_matcher.NewPathMatcher(path_matcher.PathMatcherOptions{
			BasePath:     i.Add,
			IncludeGlobs: i.IncludePaths,
			ExcludeGlobs: i.ExcludePaths,
		})

		matchedMap := map[string]string{}
		err = tree.Files().ForEach(func(f *object.File) error {
			if pm.IsPathMatched(f.Name) {
				relPath := strings.TrimPrefix(f.Name, filepath.Clean(i.Add))
				relPath = strings.TrimPrefix(relPath, string(filepath.Separator))
				newPath := path.Join(i.To, relPath)
				newPath = strings.TrimPrefix(newPath, string(filepath.Separator))

				matchedMap[newPath] = f.Name
			}
			return nil
		})
		if err != nil {
			return nil, fmt.Errorf("failed to iterate over files: %w", err)
		}

		include := &Include{
			Name:    i.Name,
			repo:    remoteRepo,
			commit:  commit,
			objects: matchedMap,
		}

		includes = append(includes, include)
	}
	return includes, nil
}

func (i *Include) WalkObjects(fn func(toPath string, origPath string) error) error {
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

	data, err := i.repo.ReadCommitFile(ctx, i.commit.Hash.String(), filePath)
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

func IsDirExists(ctx context.Context, includes []*Include, relPath string) bool {
	for _, i := range includes {
		if i == nil {
			continue
		}
		exists, _ := i.repo.IsCommitDirectoryExist(ctx, i.commit.Hash.String(), relPath)
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
		exists, _ := i.repo.IsCommitFileExist(ctx, i.commit.Hash.String(), relPath)
		if exists {
			return true
		}
	}
	return false
}

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
				logboek.Context(ctx).Debug().LogF("Found config file %q in %q\n", cfgPath, include.repo.Url)
				return cfgPath, data, nil
			}
		}
	}
	return "", nil, fmt.Errorf("unable to find config file in includes")
}

func getCommit(r *git.Repository, i *includeConf) (*object.Commit, error) {
	switch {
	case i.Commit != "":
		return commitRef(r, i.Commit)
	case i.Tag != "":
		return tagRef(r, i.Tag)
	case i.Branch != "":
		return branchRef(r, i.Branch)
	default:
		return nil, fmt.Errorf("no commit, tag or branch specified for include %s", i.Name)
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
