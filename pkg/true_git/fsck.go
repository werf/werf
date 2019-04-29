package true_git

import (
	"fmt"
	"os/exec"
	"strings"
)

type FsckResult struct {
	UnreachableCommits []string
}

type FsckOptions struct {
	Unreachable bool
	NoReflogs   bool
	Strict      bool
	Full        bool
}

// Fsck gives 'git fsck' output result
func Fsck(repoDir string, opts FsckOptions) (FsckResult, error) {
	gitArgs := []string{"--git-dir", repoDir, "fsck"}

	if opts.Unreachable {
		gitArgs = append(gitArgs, "--unreachable")
	}
	if opts.NoReflogs {
		gitArgs = append(gitArgs, "--no-reflogs")
	}
	if opts.Strict {
		gitArgs = append(gitArgs, "--strict")
	}
	if opts.Full {
		gitArgs = append(gitArgs, "--full")
	}

	cmd := exec.Command("git", gitArgs...)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return FsckResult{}, fmt.Errorf("'git fsck' failed: %s:\n%s", err, output)
	}

	lines := strings.Split(string(output), "\n")

	res := FsckResult{}

	for _, line := range lines {
		if strings.HasPrefix(line, "unreachable commit ") {
			fields := strings.Fields(line)
			if len(fields) != 3 {
				return FsckResult{}, fmt.Errorf("unexpected 'git fsck' output line: %#v", line)
			}

			commit := fields[2]
			res.UnreachableCommits = append(res.UnreachableCommits, commit)
		}
	}

	return res, nil
}
