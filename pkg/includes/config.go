package includes

import (
	"context"
	"fmt"

	"github.com/werf/logboek"
	"gopkg.in/yaml.v3"
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
	Branch string `yaml:"branch"`
	Tag    string `yaml:"tag"`
	Commit string `yaml:"commit"`
}

func NewConfig(ctx context.Context, fileReader GiterminismManagerFileReader, configRelPath, lockConfigRelPath string) (Config, error) {
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

	for i, include := range config.Includes {
		config.Includes[i].Name = include.GetName()
	}

	return config, nil
}

func parseLockConfig(ctx context.Context, fileReader GiterminismManagerFileReader, configRelPath string) (*LockInfo, error) {
	config := lockConfig{}
	data, err := fileReader.ReadIncludesLockFile(ctx, configRelPath)
	if err != nil {
		return nil, err
	}

	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("the includes lock config validation failed: %w", err)
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

func (l *LockInfo) CheckVersion(git, ref, commit string) error {
	lockCommit, ok := l.includeToCommitMapper[fmt.Sprintf("%s@%s", git, ref)]
	if !ok {
		return fmt.Errorf("lock config not found for %s", git)
	}

	if lockCommit != commit {
		return fmt.Errorf("commit mismatch for %s: expected %s, got %s", git, lockCommit, commit)
	}

	return nil
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
