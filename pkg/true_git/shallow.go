package true_git

import (
	"context"
	"fmt"
	"strings"
)

func InitBareRepoWithOrigin(ctx context.Context, gitDir, originUrl string) error {
	initCmd := NewGitCmd(ctx, nil, "init", "--bare", gitDir)
	if err := initCmd.Run(ctx); err != nil {
		return fmt.Errorf("git init bare repo command failed: %w", err)
	}

	remoteAddCmd := NewGitCmd(ctx, &GitCmdOptions{RepoDir: gitDir}, "remote", "add", "origin", originUrl)
	if err := remoteAddCmd.Run(ctx); err != nil {
		return fmt.Errorf("git remote add command failed: %w", err)
	}

	return nil
}

type ShallowFetchOptions struct {
	Env []string
}

func ShallowFetch(ctx context.Context, gitDir string, refSpecs []string, opts ShallowFetchOptions) error {
	args := append([]string{"fetch", "--depth=1", "origin"}, refSpecs...)

	fetchCmd := NewGitCmd(ctx, &GitCmdOptions{RepoDir: gitDir, Env: opts.Env}, args...)
	if err := fetchCmd.Run(ctx); err != nil {
		return fmt.Errorf("git shallow fetch command failed: %w", err)
	}

	return nil
}

func UpdateRef(ctx context.Context, gitDir, ref, value string) error {
	updateRefCmd := NewGitCmd(ctx, &GitCmdOptions{RepoDir: gitDir}, "update-ref", ref, value)
	if err := updateRefCmd.Run(ctx); err != nil {
		return fmt.Errorf("git update-ref command failed: %w", err)
	}

	return nil
}

type LsRemoteTagsOptions struct {
	Env []string
}

// RemoteTagRef describes a remote tag: ObjectSHA is what refs/tags/<name>
// points to; PeeledSHA is the fully peeled commit SHA from the "^{}" line
// (empty for lightweight tags, whose ObjectSHA is already peeled).
type RemoteTagRef struct {
	ObjectSHA string
	PeeledSHA string
}

func LsRemoteTags(ctx context.Context, url string, opts LsRemoteTagsOptions) (map[string]RemoteTagRef, error) {
	lsRemoteCmd := NewGitCmd(ctx, &GitCmdOptions{Env: opts.Env}, "ls-remote", "--tags", url)
	if err := lsRemoteCmd.Run(ctx); err != nil {
		return nil, fmt.Errorf("git ls-remote command failed: %w", err)
	}

	return parseLsRemoteTagsOutput(lsRemoteCmd.OutBuf.String())
}

func parseLsRemoteTagsOutput(out string) (map[string]RemoteTagRef, error) {
	res := map[string]RemoteTagRef{}

	for _, line := range strings.Split(out, "\n") {
		if line == "" {
			continue
		}

		parts := strings.SplitN(line, "\t", 2)
		if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
			return nil, fmt.Errorf("malformed git ls-remote output line %q", line)
		}

		sha, refName := parts[0], parts[1]

		if !strings.HasPrefix(refName, "refs/tags/") {
			continue
		}
		tagName := strings.TrimPrefix(refName, "refs/tags/")

		if peeledTagName, isPeelLine := strings.CutSuffix(tagName, "^{}"); isPeelLine {
			tagRef := res[peeledTagName]
			tagRef.PeeledSHA = sha
			res[peeledTagName] = tagRef
		} else {
			tagRef := res[tagName]
			tagRef.ObjectSHA = sha
			res[tagName] = tagRef
		}
	}

	return res, nil
}
