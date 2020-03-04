package main

import (
	"os"

	"github.com/flant/werf/pkg/path_matcher"
	"github.com/flant/werf/pkg/true_git"
)

func main() {
	err := true_git.Init(true_git.Options{})
	if err != nil {
		panic(err)
	}

	p, err := os.Create("my-archive.tar")
	if err != nil {
		panic(err)
	}

	_, err = true_git.ArchiveWithSubmodules(p, os.Args[1], "my-work-tree", true_git.ArchiveOptions{
		Commit:      os.Args[2],
		PathMatcher: &path_matcher.GitMappingPathMatcher{},
	})
	if err != nil {
		panic(err)
	}

	err = p.Close()
	if err != nil {
		panic(err)
	}
}
