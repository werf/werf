package main

import (
	"fmt"
	"path/filepath"

	"github.com/flant/werf/pkg/git_repo"
)

func main() {
	repoPath, err := filepath.Abs(".")
	if err != nil {
		panic(err)
	}
	repo := git_repo.Local{Path: repoPath}

	for _, c := range []string{"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaadsfXXZVN<CXNV<MVCXNCX<NVNVCX<MNV324a", "48926fbcc6735b4b5d7e54cef16fbb4ea9c705fa", "18768775b70ad50f1cedf63c506bf2b382050720"} {
		isCommitExist, err := repo.IsCommitExists(c)
		fmt.Printf("%v -> IsCommitExist=%v, err=%v\n", c, isCommitExist, err)
	}
	fmt.Println()

	tags, err := repo.TagsList()
	if err != nil {
		panic(err)
	}

	fmt.Printf("Tags:\n")
	for _, t := range tags {
		fmt.Printf(" * %s\n", t)
	}
	fmt.Println()

	branches, err := repo.RemoteBranchesList()
	if err != nil {
		panic(err)
	}

	fmt.Printf("Branches:\n")
	for _, b := range branches {
		fmt.Printf(" * %s\n", b)
	}
}
