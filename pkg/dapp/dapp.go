package dapp

import (
	"os"
	"path/filepath"
)

const (
	Version = "1.0.0-alpha.1"
)

var (
	tmpDir, homeDir string
)

func GetHomeDir() string {
	if homeDir == "" {
		panic("bug: init required!")
	}

	return homeDir
}

func GetTmpDir() string {
	if tmpDir == "" {
		panic("bug: init required!")
	}

	return tmpDir
}

func Init(tmpDirOption, homeDirOption string) error {
	if val, ok := os.LookupEnv("DAPP_TMP"); ok {
		tmpDir = val
	} else if tmpDirOption != "" {
		tmpDir = tmpDirOption
	} else {
		tmpDir = os.TempDir()
	}

	if val, ok := os.LookupEnv("DAPP_HOME"); ok {
		homeDir = val
	} else if homeDirOption != "" {
		homeDir = homeDirOption
	} else {
		homeDir = filepath.Join(os.Getenv("HOME"), ".dapp")
	}

	return nil
}
