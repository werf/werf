package includes

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"gopkg.in/yaml.v3"

	"github.com/werf/common-go/pkg/util"
	"github.com/werf/logboek"
	"github.com/werf/werf/v2/pkg/git_repo"
)

type Config struct {
	Includes []includeConf `yaml:"includes"`
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

func NewConfig(ctx context.Context, fileReader GiterminismManagerFileReader, configRelPath string) (Config, error) {
	logboek.Context(ctx).Debug().LogF("Reading includes config from %q\n", configRelPath)
	exist, err := fileReader.IsIncludesConfigExistAnywhere(ctx, configRelPath)
	if err != nil {
		return Config{}, err
	}

	if !exist {
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

	for _, include := range config.Includes {
		if !oneOrNone([]bool{include.Branch != "", include.Commit != "", include.Tag != ""}) {
			return config, fmt.Errorf("specify only `branch: BRANCH` or `tag: TAG` or `commit: COMMIT` for include %s", include.Git)
		}
	}

	return config, nil
}

func getLockInfo(ctx context.Context, cfg Config, reomteRepos map[string]*git_repo.Remote, opts InitIncludesOptions) (*LockInfo, error) {
	var lockConf *lockConfig

	if opts.UseLatestVersion {
		cfg, err := createLockConfig(createLockConfigOptions{
			fileReader:     opts.FileReader,
			includesConfig: cfg,
			remoteRepos:    reomteRepos,
		})
		if err != nil {
			return nil, fmt.Errorf("create lock config: %w", err)
		}
		lockConf = &cfg
	}

	lockInfo, err := readLockInfo(ctx, opts.FileReader, opts.ConfigRelPath, lockConf)
	if err != nil {
		return nil, fmt.Errorf("unable to read include lock info: %w", err)
	}

	if len(lockInfo.includeToCommitMapper) == 0 {
		return nil, fmt.Errorf("no includes found in werf-includes.lock")
	}

	return lockInfo, nil
}

func readLockInfo(ctx context.Context, fileReader GiterminismManagerFileReader, configRelPath string, lockConf *lockConfig) (*LockInfo, error) {
	var config *lockConfig
	if lockConf != nil {
		config = lockConf
	} else {
		var err error
		config, err = parseLockConfig(ctx, fileReader, configRelPath)
		if err != nil {
			return nil, err
		}
	}

	lockInfo := &LockInfo{
		includeToCommitMapper: make(map[string]string),
	}

	for _, l := range config.IncludeLock {
		ref, err := l.Ref()
		if err != nil {
			return nil, fmt.Errorf("unable to get ref for include %s: %w", l.Git, err)
		}
		key := fmt.Sprintf("%s@%s", l.Git, ref)
		lockInfo.includeToCommitMapper[key] = l.Commit
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
	fileReader       GiterminismManagerFileReader
	includesConfig   Config
	includesLockPath string
	remoteRepos      map[string]*git_repo.Remote
}

func CreateOrUpdateLockConfig(ctx context.Context, opts createLockConfigOptions) error {
	exists, _ := util.FileExists(opts.includesLockPath)
	if exists {
		err := UpdateLockConfig(ctx, opts)
		if err != nil {
			return fmt.Errorf("error update includes lock file: %w", err)
		}
		logboek.Context(ctx).Info().LogF("Successfully updated %q file\n", opts.includesLockPath)
		return nil
	}
	err := CreateLockConfig(ctx, opts)
	if err != nil {
		return fmt.Errorf("error create includes lock file: %w", err)
	}
	logboek.Context(ctx).Info().LogF("Successfully created %q file\n", opts.includesLockPath)
	return nil
}

func CreateLockConfig(ctx context.Context, opts createLockConfigOptions) error {
	includesMap := make(map[string]bool)
	var lockConfs []includeLockConf
	for _, c := range opts.includesConfig.Includes {
		ref, err := c.Ref()
		if err != nil {
			return fmt.Errorf("get ref for include %s: %w", c.Git, err)
		}
		key := fmt.Sprintf("%s@%s", c.Git, ref)
		if !includesMap[key] {
			lockConfs = append(lockConfs, includeLockConf{
				Git:    c.Git,
				Branch: c.Branch,
				Tag:    c.Tag,
			})
			includesMap[key] = true
		}
	}

	return writeLockConfig(lockConfs, opts.includesLockPath, opts.remoteRepos)
}

func createLockConfig(opts createLockConfigOptions) (lockConfig, error) {
	includesMap := make(map[string]bool)
	var lockConfs []includeLockConf
	for _, c := range opts.includesConfig.Includes {
		ref, err := c.Ref()
		if err != nil {
			return lockConfig{}, fmt.Errorf("get ref for include %s: %w", c.Git, err)
		}
		key := fmt.Sprintf("%s@%s", c.Git, ref)
		if !includesMap[key] {
			lockConfs = append(lockConfs, includeLockConf{
				Git:    c.Git,
				Branch: c.Branch,
				Tag:    c.Tag,
			})
			includesMap[key] = true
		}
	}

	newLockConfig := lockConfig{
		IncludeLock: make([]includeLockConf, 0, len(lockConfs)),
	}

	for _, c := range lockConfs {
		updated, err := c.updateCommit(opts.remoteRepos)
		if err != nil {
			return lockConfig{}, err
		}
		newLockConfig.IncludeLock = append(newLockConfig.IncludeLock, *updated)
	}

	return newLockConfig, nil
}

func UpdateLockConfig(ctx context.Context, opts createLockConfigOptions) error {
	cfg, err := parseLockConfig(ctx, opts.fileReader, opts.includesLockPath)
	if err != nil {
		return fmt.Errorf("parse includes lock config: %w", err)
	}

	return writeLockConfig(cfg.IncludeLock, opts.includesLockPath, opts.remoteRepos)
}

func (l *LockInfo) GetCommit(git, ref string) (string, error) {
	key := fmt.Sprintf("%s@%s", git, ref)
	commit, ok := l.includeToCommitMapper[key]
	if !ok {
		return "", fmt.Errorf("lock config not found for %s", git)
	}
	return commit, nil
}

func (i *includeLockConf) Ref() (string, error) {
	return ref(i.Git, i.Tag, i.Branch, i.Commit)
}

func ref(git, tag, branch, commit string) (string, error) {
	switch {
	case tag != "":
		return tag, nil
	case branch != "":
		return branch, nil
	case commit != "":
		return commit, nil
	default:
		return "", fmt.Errorf("no ref specified for include %s", git)
	}
}

func (i *includeLockConf) getCommit(r *git.Repository) (*object.Commit, error) {
	return getCommit(r, i.Git, i.Tag, i.Branch, i.Commit)
}

func writeLockConfig(inputConfs []includeLockConf, configRelPath string, remoteRepos map[string]*git_repo.Remote) error {
	newLockConfig := lockConfig{
		IncludeLock: make([]includeLockConf, 0, len(inputConfs)),
	}

	for _, c := range inputConfs {
		updated, err := c.updateCommit(remoteRepos)
		if err != nil {
			return err
		}
		newLockConfig.IncludeLock = append(newLockConfig.IncludeLock, *updated)
	}

	outData, err := yaml.Marshal(newLockConfig)
	if err != nil {
		return fmt.Errorf("marshal new lock config: %w", err)
	}

	fp, _ := filepath.Abs(configRelPath)
	if err := os.WriteFile(fp, outData, os.ModePerm); err != nil {
		return fmt.Errorf("write new lock config: %w", err)
	}

	return nil
}

func (c *includeLockConf) updateCommit(remoteRepos map[string]*git_repo.Remote) (*includeLockConf, error) {
	repo, ok := remoteRepos[c.Git]
	if !ok || repo == nil {
		return nil, fmt.Errorf("remote repo %s not found", c.Git)
	}

	r, err := repo.PlainOpen()
	if err != nil {
		return nil, fmt.Errorf("plain open: %w", err)
	}

	commit, err := c.getCommit(r)
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

func oneOrNone(conditions []bool) bool {
	if len(conditions) == 0 {
		return true
	}

	exist := false
	for _, condition := range conditions {
		if condition {
			if exist {
				return false
			} else {
				exist = true
			}
		}
	}
	return true
}
