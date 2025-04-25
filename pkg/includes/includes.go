package includes

import (
	"context"
	"fmt"
	"path"
	"path/filepath"
	"strings"

	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/werf/logboek"
	"github.com/werf/werf/v2/pkg/git_repo"
	"github.com/werf/werf/v2/pkg/path_matcher"
	"gopkg.in/yaml.v3"
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
	Ref          string   `yaml:"ref"`
	Add          string   `yaml:"add,omitempty"`
	To           string   `yaml:"to,omitempty"`
	IncludePaths []string `yaml:"includePaths"`
	ExcludePaths []string `yaml:"excludePaths"`
}

type Include struct {
	Name    string `yaml:"name"`
	repo    *git_repo.Remote
	commit  *object.Commit
	objects map[string]string
}

func GetWerfIncludesConfigRelPath(path string) string {
	if path == "" {
		return "werf-includes.yaml"
	}
	return filepath.ToSlash(path)
}

func NewConfig(ctx context.Context, fileReader GiterminismManagerFileReader, configRelPath string) (Config, error) {
	config := Config{}
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

func GetIncludes(cfg Config) ([]*Include, error) {
	ctx := context.Background()

	inc := []*Include{}
	for _, i := range cfg.Includes {
		repo, err := git_repo.OpenRemoteRepo(i.Name, i.Git)
		if err != nil {
			return nil, err
		}

		if err := logboek.Context(ctx).Info().LogProcess(fmt.Sprintf("Refreshing %s repository", i.Name)).
			DoError(func() error {
				return repo.CloneAndFetch(ctx)
			}); err != nil {
			return nil, fmt.Errorf("unable to clone %s repository: %w", i.Name, err)
		}

		pm := path_matcher.NewPathMatcher(path_matcher.PathMatcherOptions{
			BasePath:     i.Add,
			IncludeGlobs: i.IncludePaths,
			ExcludeGlobs: i.ExcludePaths,
		})

		r, _ := repo.PlainOpen()

		ref, err := r.Reference(plumbing.NewBranchReferenceName("main"), true)
		if err != nil {
			return nil, fmt.Errorf("failed to resolve ref: %w", err)
		}

		commit, err := r.CommitObject(ref.Hash())
		if err != nil {
			return nil, fmt.Errorf("failed to get commit object: %w", err)
		}

		tree, err := commit.Tree()
		if err != nil {
			return nil, fmt.Errorf("failed to get tree: %w", err)
		}

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

		fmt.Println("matchedMap", matchedMap)
		include := &Include{
			Name:    i.Name,
			repo:    repo,
			commit:  commit,
			objects: matchedMap,
		}

		inc = append(inc, include)
	}
	return inc, nil
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
	fmt.Println("GetFile", relPath)
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
