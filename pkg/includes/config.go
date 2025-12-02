package includes

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"gopkg.in/yaml.v3"

	"github.com/werf/logboek"
	"github.com/werf/werf/v2/pkg/git_repo"
)

type Config struct {
	Includes []includeConf `yaml:"includes"`
}

type includeConf struct {
	Git          string                         `yaml:"git"`
	BasicAuth    *git_repo.BasicAuthCredentials `yaml:"basicAuth,omitempty"`
	Branch       string                         `yaml:"branch"`
	Tag          string                         `yaml:"tag"`
	Commit       string                         `yaml:"commit"`
	Add          string                         `yaml:"add,omitempty"`
	To           string                         `yaml:"to,omitempty"`
	IncludePaths []string                       `yaml:"includePaths"`
	ExcludePaths []string                       `yaml:"excludePaths"`
}

func (i *includeConf) Ref() (string, error) {
	return ref(i.Git, i.Commit, i.Tag, i.Branch)
}

type LockInfo struct {
	includeToCommitMapper map[string]string
}

type lockConfig struct {
	IncludeLock []includeLockConf `yaml:"includes"`
}

type includeLockConf struct {
	Git    string `yaml:"git"`
	Branch string `yaml:"branch,omitempty"`
	Tag    string `yaml:"tag,omitempty"`
	Commit string `yaml:"commit"`
}

func NewConfig(ctx context.Context, fileReader GiterminismManagerFileReader, configRelPath string, createLockConfig bool) (Config, error) {
	logboek.Context(ctx).Debug().LogF("Reading includes config from %q\n", configRelPath)
	exist, err := fileReader.IsIncludesConfigExistAnywhere(ctx, configRelPath)
	if err != nil {
		return Config{}, err
	}

	if !exist {
		if createLockConfig {
			return Config{}, fmt.Errorf("includes config file %q not found", configRelPath)
		}
		return Config{}, nil
	}

	includesConfig, err := parseConfig(ctx, fileReader, configRelPath)
	if err != nil {
		return Config{}, err
	}

	return includesConfig, err
}

func parseConfig(ctx context.Context, fileReader GiterminismManagerFileReader, configRelPath string) (Config, error) {
	data, err := fileReader.ReadIncludesConfig(ctx, configRelPath)
	if err != nil {
		return Config{}, err
	}
	config := Config{}
	if err := yaml.Unmarshal(data, &config); err != nil {
		return config, fmt.Errorf("the includes config validation failed: %w", err)
	}

	if err := validate(config); err != nil {
		return config, fmt.Errorf("includes config validation failed: %w\n\n%s", err, string(data))
	}

	return config, nil
}

func validate(config Config) error {
	for _, include := range config.Includes {
		if include.Git == "" {
			return fmt.Errorf("`git` field is required")
		}

		if include.BasicAuth != nil {
			if include.BasicAuth.Username == "" {
				return fmt.Errorf("username should be specified when using git basic auth")
			}

			if !exactlyOne([]bool{include.BasicAuth.Password.Env != "", include.BasicAuth.Password.Src != "", include.BasicAuth.Password.PlainValue != ""}) {
				err := fmt.Errorf("include %s: specify only env or src or plain as basic auth password source", include.BasicAuth)
				return err
			}
		}

		if include.Add == "" {
			return fmt.Errorf("include %s: `add` field is required", include.Git)
		}
		if !strings.HasPrefix(include.Add, "/") {
			return fmt.Errorf("include %s: `add` must be an absolute path relative to the repository root", include.Git)
		}
		if include.To == "" {
			return fmt.Errorf("include %s: `to` field is required", include.Git)
		}
		if !strings.HasPrefix(include.To, "/") {
			return fmt.Errorf("include %s: `to` must be an absolute path relative to the repository root", include.Git)
		}

		for _, path := range include.IncludePaths {
			if strings.HasPrefix(path, "/") {
				return fmt.Errorf("include %s: `includePaths` must be relative paths to the repository root", include.Git)
			}
		}

		for _, path := range include.ExcludePaths {
			if strings.HasPrefix(path, "/") {
				return fmt.Errorf("include %s: `excludePaths` must be relative paths to the repository root", include.Git)
			}
		}

		if !exactlyOne([]bool{include.Branch != "", include.Commit != "", include.Tag != ""}) {
			err := fmt.Errorf("include %s: specify only `branch` or `tag` or `commit`", include.Git)
			return err
		}
	}

	return nil
}

type getLockInfoOptions struct {
	includesConfig         Config
	createOrUpdateLockFile bool
	useLatestVersion       bool
	remoteRepos            *gitRepositoriesWithCache
	lockConfig             *lockConfig
}

func getLockInfo(opts getLockInfoOptions) (*LockInfo, error) {
	var lockConf *lockConfig

	if opts.useLatestVersion {
		cfg, err := createLockConfig(createLockConfigOptions{
			includesConfig: opts.includesConfig,
			remoteRepos:    opts.remoteRepos,
		})
		if err != nil {
			return nil, fmt.Errorf("create lock config: %w", err)
		}
		lockConf = &cfg
	} else {
		lockConf = opts.lockConfig
	}

	lockInfo, err := readLockInfo(lockConf)
	if err != nil {
		return nil, fmt.Errorf("unable to read include lock info: %w", err)
	}

	if len(lockInfo.includeToCommitMapper) == 0 {
		return nil, fmt.Errorf("no includes found in werf-includes.lock")
	}

	return lockInfo, nil
}

func readLockInfo(lockConf *lockConfig) (*LockInfo, error) {
	lockInfo := &LockInfo{
		includeToCommitMapper: make(map[string]string),
	}

	for _, l := range lockConf.IncludeLock {
		ref, err := l.Ref()
		if err != nil {
			return nil, fmt.Errorf("unable to get ref for include %s: %w", l.Git, err)
		}
		lockInfo.includeToCommitMapper[lockId(l.Git, ref)] = l.Commit
	}

	return lockInfo, nil
}

func parseLockConfig(ctx context.Context, fileReader GiterminismManagerFileReader, configRelPath string) (*lockConfig, error) {
	config := lockConfig{}
	data, err := fileReader.ReadIncludesLockFile(ctx, configRelPath)
	if err != nil {
		return nil, err
	}

	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("the includes lock config validation failed: %w", err)
	}
	return &config, nil
}

type createLockConfigOptions struct {
	projectDir       string
	includesConfig   Config
	includesLockPath string
	remoteRepos      *gitRepositoriesWithCache
}

func CreateOrUpdateLockConfig(ctx context.Context, opts createLockConfigOptions) error {
	err := CreateLockConfig(ctx, opts)
	if err != nil {
		return fmt.Errorf("error create includes lock file: %w", err)
	}
	logboek.Context(ctx).Info().LogF("Successfully created %q file\n", opts.includesLockPath)
	return nil
}

func CreateLockConfig(_ context.Context, opts createLockConfigOptions) error {
	locksConf, err := createLockConfig(opts)
	if err != nil {
		return fmt.Errorf("create lock config: %w", err)
	}

	includesLockPathAbs := filepath.Join(opts.projectDir, opts.includesLockPath)
	return writeLockConfig(locksConf, includesLockPathAbs)
}

func createLockConfig(opts createLockConfigOptions) (lockConfig, error) {
	includesMap := make(map[string]bool)
	var lockConfs []includeLockConf
	for _, c := range opts.includesConfig.Includes {
		ref, err := c.Ref()
		if err != nil {
			return lockConfig{}, fmt.Errorf("get ref for include %s: %w", c.Git, err)
		}
		lockId := lockId(c.Git, ref)
		if !includesMap[lockId] {
			lockConfs = append(lockConfs, includeLockConf{
				Git:    c.Git,
				Branch: c.Branch,
				Tag:    c.Tag,
				Commit: c.Commit,
			})
			includesMap[lockId] = true
		}
	}

	newLockConfig, err := newLockConfig(lockConfs, opts.remoteRepos)
	if err != nil {
		return lockConfig{}, fmt.Errorf("unable to update lock config: %w", err)
	}

	return newLockConfig, nil
}

func newLockConfig(cfg []includeLockConf, remoteRepos *gitRepositoriesWithCache) (lockConfig, error) {
	newLockConfig := lockConfig{
		IncludeLock: make([]includeLockConf, 0, len(cfg)),
	}

	for _, c := range cfg {
		updated, err := c.updateCommit(remoteRepos)
		if err != nil {
			return newLockConfig, err
		}
		newLockConfig.IncludeLock = append(newLockConfig.IncludeLock, *updated)
	}

	return newLockConfig, nil
}

func (l *LockInfo) GetCommit(git, ref string) (string, error) {
	commit, ok := l.includeToCommitMapper[lockId(git, ref)]
	if !ok {
		return "", fmt.Errorf("lock config not found for %s.\n\nUpdate lock file using `werf includes update` command", git)
	}
	return commit, nil
}

func (i *includeLockConf) Ref() (string, error) {
	return ref(i.Git, i.Tag, i.Branch, i.Commit)
}

func (i *includeLockConf) getCommit(r *git.Repository) (*object.Commit, error) {
	return getCommit(r, i.Git, i.Tag, i.Branch, i.Commit)
}

func writeLockConfig(inputConfs lockConfig, configAbsPath string) error {
	outData, err := yaml.Marshal(inputConfs)
	if err != nil {
		return fmt.Errorf("marshal new lock config: %w", err)
	}

	if err := os.WriteFile(configAbsPath, outData, os.ModePerm); err != nil {
		return fmt.Errorf("write new lock config: %w", err)
	}

	return nil
}

func (c *includeLockConf) updateCommit(remoteRepos *gitRepositoriesWithCache) (*includeLockConf, error) {
	r, err := remoteRepos.getRepository(c.Git)
	if err != nil {
		return nil, err
	}

	repo, err := r.repo.PlainOpen()
	if err != nil {
		return nil, fmt.Errorf("plain open: %w", err)
	}

	commit, err := c.getCommit(repo)
	if err != nil {
		return nil, fmt.Errorf("get commit: %w", err)
	}

	return &includeLockConf{
		Git:    c.Git,
		Branch: c.Branch,
		Tag:    c.Tag,
		Commit: commit.Hash.String(),
	}, nil
}

func exactlyOne(conditions []bool) bool {
	count := 0
	for _, c := range conditions {
		if c {
			count++
		}
	}
	return count == 1
}

func lockId(git, ref string) string {
	return fmt.Sprintf("%s@%s", git, ref)
}
