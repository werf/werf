package main

import (
	"os"

	"github.com/flant/dapp/pkg/git"
)

func main() {
	p, err := os.Create("/tmp/myarchive.tar")
	if err != nil {
		panic(err)
	}

	_, err = git.Archive(p, ".", git.ArchiveOptions{
		Commit:         "26e3e9191fb40050e254b48085be41a38036e19a",
		WithSubmodules: true,
		PathFilter:     git.PathFilter{},
	})
	if err != nil {
		panic(err)
	}

	err = p.Close()
	if err != nil {
		panic(err)
	}
}
