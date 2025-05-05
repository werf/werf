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
	Name   string `yaml:"name"`
	Commit string `yaml:"commit"`
}

type NewConfigOptions struct {
	configRelPath     string
	lockConfigRelPath string
	ignoreLockfile    bool
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
		lockInfo.includeToCommitMapper[l.Name] = l.Commit
	}

	return lockInfo, nil
}

func (l *LockInfo) CheckVersion(includeName, commitHash string) error {
	if commit, ok := l.includeToCommitMapper[includeName]; ok {
		if commit != commitHash {
			return fmt.Errorf("include %q commit hash %q does not match the lock file commit hash %q", includeName, commitHash, commit)
		}
		return nil
	}
	return fmt.Errorf("include %q commit hash %q not found in the lock file", includeName, commitHash)
}
