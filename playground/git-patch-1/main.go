package main

import (
	"fmt"
	"os"

	"github.com/flant/dapp/pkg/true_git"
)

func main() {
	err := true_git.Init()
	if err != nil {
		panic(err)
	}

	f, err := os.Create("my-patch.diff")
	if err != nil {
		panic(err)
	}

	p, err := true_git.PatchWithSubmodules(f, os.Args[1], "my-work-tree", true_git.PatchOptions{
		FromCommit: os.Args[2],
		ToCommit:   os.Args[3],
		PathFilter: true_git.PathFilter{},
	})
	if err != nil {
		panic(err)
	}

	fmt.Printf("Patch is-empty=%v has-binary=%v\n", p.IsEmpty, p.HasBinary)
}
