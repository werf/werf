package main

import (
	"os"

	"github.com/flant/dapp/pkg/git"
)

func main() {
	err := git.Init()
	if err != nil {
		panic(err)
	}

	// var buf bytes.Buffer

	err = git.Diff(os.Stdout, ".", git.DiffOptions{
		// WithSubmodules: true,
		FromCommit: "37223be14696e38eeeac3512f6648c7741a34619",
		ToCommit:   "721e19e4e938fc44c5ad09e558d0115deb64dc33",
		PathFilter: git.PathFilter{},
	})
	if err != nil {
		panic(err)
	}
}
