package git_repo

import (
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"

	"github.com/go-git/go-git/v5/plumbing"
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
	PlainValue string `yaml:"plainValue"`
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
	} else if cfg.Password.Src != "" {
		f, err := filepath.Abs(cfg.Password.Src)
		if err != nil {
			return nil, fmt.Errorf("unable to get secret file absolute path: %w", err)
		}
		b, err := os.ReadFile(f)
		if err != nil {
			return nil, fmt.Errorf("unable to read secret file: %w", err)
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
