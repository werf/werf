package git_repo

import (
	"encoding/hex"
	"fmt"
	"os"

	"github.com/go-git/go-git/v5/plumbing"

	"github.com/werf/common-go/pkg/util"
)

func newHash(s string) (plumbing.Hash, error) {
	var h plumbing.Hash

	b, err := hex.DecodeString(s)
	if err != nil {
		return h, err
	}

	copy(h[:], b)
	return h, nil
}

type BasicAuthCredentials struct {
	Username string
	Password PasswordSource
}

type PasswordSource struct {
	Env        string `yaml:"env"`
	Src        string `yaml:"src"`
	PlainValue string `yaml:"value"`
}

func BasicAuthCredentialsHelper(cfg *BasicAuthCredentials) (*BasicAuth, error) {
	if cfg == nil {
		return nil, nil
	}

	if !oneOrNone([]bool{cfg.Password.Env != "", cfg.Password.Src != "", cfg.Password.PlainValue != ""}) {
		return nil, fmt.Errorf("password source type could be ONLY `env`, `src` or `value`")
	}

	password := ""

	if cfg.Password.Env != "" {
		password = os.Getenv(cfg.Password.Env)
		if len(password) == 0 {
			return nil, fmt.Errorf("environment variable %q is not set", cfg.Password.Env)
		}
	} else if cfg.Password.Src != "" {
		absPath, err := util.ExpandPath(cfg.Password.Src)
		if err != nil {
			return nil, fmt.Errorf("error load secret from src: %w", err)
		}

		if exists, _ := util.FileExists(absPath); !exists {
			return nil, fmt.Errorf("error load secret from src: path %s doesn't exist", absPath)
		}

		b, err := os.ReadFile(absPath)
		if err != nil {
			return nil, fmt.Errorf("error reading secret from file %s: %w", absPath, err)
		}

		password = string(b)
	} else if cfg.Password.PlainValue != "" {
		password = cfg.Password.PlainValue
	} else {
		return nil, fmt.Errorf("no secret specified")
	}

	return &BasicAuth{
		Username: cfg.Username,
		Password: password,
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
