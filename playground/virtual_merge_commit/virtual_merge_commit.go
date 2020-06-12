package main

import (
	"fmt"
	"os"

	"github.com/werf/werf/pkg/true_git"

	"github.com/werf/werf/pkg/werf"

	"github.com/werf/werf/pkg/git_repo"
)

func do(v1FromCommit, v1IntoCommit, v2FromCommit, v2IntoCommit string) error {
	if gitRepo, err := git_repo.OpenLocalRepo("virtual_merge_commit", "."); err != nil {
		return err
	} else {
		if v1Commit, err := gitRepo.CreateVirtualMergeCommit(v1FromCommit, v1IntoCommit); err != nil {
			return fmt.Errorf("unable to create virtual merge commit of %s and %s: %s", v1FromCommit, v1IntoCommit, err)
		} else if v2Commit, err := gitRepo.CreateVirtualMergeCommit(v2FromCommit, v2IntoCommit); err != nil {
			return fmt.Errorf("unable to create virtual merge commit of %s and %s: %s", v2FromCommit, v2IntoCommit, err)
		} else {
			fmt.Printf("v1Commit: %s\nv2Commit: %s\n", v1Commit, v2Commit)

			if patch, err := gitRepo.CreatePatch(git_repo.PatchOptions{
				FromCommit:            v1Commit,
				ToCommit:              v2Commit,
				WithEntireFileContext: true,
				WithBinary:            true,
			}); err != nil {
				return fmt.Errorf("unable to create patch between %s and %s: %s", v1Commit, v2Commit, err)
			} else {
				fmt.Printf("THE PATCH:\n%s\n", patch)
			}
		}
	}

	return nil
}

func main() {
	fmt.Printf("Virtual merge commit test BEGIN\n")

	if err := werf.Init("", ""); err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %s\n", err)
		os.Exit(1)
	}

	if err := true_git.Init(true_git.Options{LiveGitOutput: true}); err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %s\n", err)
		os.Exit(1)
	}

	if err := do(os.Args[1], os.Args[2], os.Args[3], os.Args[4]); err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %s\n", err)
		os.Exit(1)
	}

	fmt.Printf("Virtual merge commit test SUCCEEDED\n")
}
