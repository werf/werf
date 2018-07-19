package dapp

import (
	"os"
	"path/filepath"
)

var (
	HomeDir = filepath.Join(os.Getenv("HOME"), ".dapp")
)
