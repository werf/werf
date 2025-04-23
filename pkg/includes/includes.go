package includes

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

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
	Includes []Include `json:"includes"`
}

type Include struct {
	Name         string   `yaml:"name"`
	Git          string   `yaml:"git"`
	Ref          string   `yaml:"ref"`
	Add          string   `yaml:"add,omitempty"`
	To           string   `yaml:"to,omitempty"`
	IncludePaths []string `yaml:"includePaths"`
	ExcludePaths []string `yaml:"excludePaths"`
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

func GetIncludes(cfg Config) error {
	ctx := context.Background()
	for _, i := range cfg.Includes {
		repo, err := git_repo.OpenRemoteRepo(i.Name, i.Git)
		if err != nil {
			return err
		}

		if err := logboek.Context(ctx).Info().LogProcess(fmt.Sprintf("Refreshing %s repository", i.Name)).
			DoError(func() error {
				return repo.CloneAndFetch(ctx)
			}); err != nil {
			return err
		}

		fileRenames, err := i.getFileRenames(ctx, i.Ref, repo)
		if err != nil {
			return fmt.Errorf("unable to make git archive options: %w", err)
		}

		pathScope, err := i.getPathScope(ctx, i.Ref, repo)
		if err != nil {
			return fmt.Errorf("unable to make git archive options: %w", err)
		}

		pm := path_matcher.NewPathMatcher(path_matcher.PathMatcherOptions{
			BasePath:     i.Add,
			IncludeGlobs: i.IncludePaths,
			ExcludeGlobs: i.ExcludePaths,
		})
		archive, err := repo.GetOrCreateArchive(ctx, git_repo.ArchiveOptions{
			PathScope:   pathScope,
			PathMatcher: pm,
			Commit:      i.Ref,
			FileRenames: fileRenames,
		})
		if err != nil {
			return fmt.Errorf("unable to create git archive for commit %s with path scope %s: %w", i.Ref, pathScope, err)
		}

		wd, _ := os.Getwd()
		err = extractTar(archive.GetFilePath(), wd+i.To)
		if err != nil {
			return fmt.Errorf("unable to extract tar archive %s to %s: %w", archive.GetFilePath(), wd+i.To, err)
		}
	}
	return nil
}
