package main

import (
	"fmt"
	"os"

	"github.com/flant/werf/pkg/true_git"
)

func main() {
	res, err := true_git.Fsck(os.Args[1], true_git.FsckOptions{Unreachable: true, NoReflogs: true, Full: true, Strict: true})

	if err != nil {
		fmt.Printf("ERROR! %s\n", err)
	}

	fmt.Printf("RESULT: %#v\n", res)
}
