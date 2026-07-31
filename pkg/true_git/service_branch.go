package true_git

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/Masterminds/semver"
)

const (
	devIndexFileName     = "dev_index"
	devIndexBaseFileName = "dev_index_base"
)

type SyncSourceWorktreeWithServiceBranchOptions struct {
	ServiceBranch   string
	GlobExcludeList []string
}

func SyncSourceWorktreeWithServiceBranch(ctx context.Context, gitDir, sourceWorktreeDir, worktreeCacheDir, commit string, opts SyncSourceWorktreeWithServiceBranchOptions) (string, error) {
	var resultCommit string
	if err := withWorkTreeCacheLock(ctx, worktreeCacheDir, func() error {
		var err error
		if gitDir, err = filepath.Abs(gitDir); err != nil {
			return fmt.Errorf("bad git dir %s: %w", gitDir, err)
		}

		if worktreeCacheDir, err = filepath.Abs(worktreeCacheDir); err != nil {
			return fmt.Errorf("bad work tree cache dir %s: %w", worktreeCacheDir, err)
		}

		serviceWorktreeDir, err := prepareWorkTree(ctx, gitDir, worktreeCacheDir, commit, true)
		if err != nil {
			return fmt.Errorf("unable to prepare worktree for commit %v: %w", commit, err)
		}

		currentCommitPath := filepath.Join(worktreeCacheDir, "current_commit")
		if err := os.RemoveAll(currentCommitPath); err != nil {
			return fmt.Errorf("unable to remove %s: %w", currentCommitPath, err)
		}

		resultCommit, err = syncWorktreeWithServiceWorktreeBranch(ctx, sourceWorktreeDir, serviceWorktreeDir, worktreeCacheDir, commit, opts.ServiceBranch, opts.GlobExcludeList)
		if err != nil {
			return fmt.Errorf("unable to sync worktree with service branch %q: %w", opts.ServiceBranch, err)
		}

		return nil
	}); err != nil {
		return "", err
	}

	return resultCommit, nil
}

func syncWorktreeWithServiceWorktreeBranch(ctx context.Context, sourceWorktreeDir, serviceWorktreeDir, worktreeCacheDir, sourceCommit, branchName string, globExcludeList []string) (string, error) {
	if err := prepareAndCheckoutServiceBranch(ctx, serviceWorktreeDir, sourceCommit, branchName); err != nil {
		return "", fmt.Errorf("unable to get or prepare service branch head commit: %w", err)
	}

	serviceBranchHeadCommit, err := GetLastBranchCommitSHA(ctx, serviceWorktreeDir, branchName)
	if err != nil {
		return "", fmt.Errorf("unable to get service worktree commit SHA: %w", err)
	}

	conversionSignature, err := computeConversionConfigSignature(ctx, sourceWorktreeDir)
	if err != nil {
		return "", fmt.Errorf("unable to compute conversion config signature: %w", err)
	}

	treeSHA, err := prepareDevIndexAndWriteTree(ctx, sourceWorktreeDir, serviceWorktreeDir, worktreeCacheDir, sourceCommit, serviceBranchHeadCommit, conversionSignature, globExcludeList)
	if err != nil {
		// The persistent index may be corrupt or left inconsistent by an interrupted run;
		// discard it and rebuild once from a clean read-tree seed before surfacing the error.
		if rmErr := removeDevIndexFiles(worktreeCacheDir); rmErr != nil {
			return "", fmt.Errorf("unable to discard dev-index after failure (%v): %w", err, rmErr)
		}
		treeSHA, err = prepareDevIndexAndWriteTree(ctx, sourceWorktreeDir, serviceWorktreeDir, worktreeCacheDir, sourceCommit, serviceBranchHeadCommit, conversionSignature, globExcludeList)
		if err != nil {
			return "", fmt.Errorf("unable to rebuild dev-index after failure: %w", err)
		}
	}

	serviceBranchHeadTree, err := resolveTreeSHA(ctx, serviceWorktreeDir, serviceBranchHeadCommit)
	if err != nil {
		return "", fmt.Errorf("unable to resolve service branch head tree: %w", err)
	}

	if treeSHA == serviceBranchHeadTree {
		if err := writeDevIndexBase(worktreeCacheDir, serviceBranchHeadCommit, conversionSignature); err != nil {
			return "", fmt.Errorf("unable to write dev-index base marker: %w", err)
		}
		return serviceBranchHeadCommit, nil
	}

	newCommit, err := commitTreeToServiceBranch(ctx, serviceWorktreeDir, branchName, treeSHA, serviceBranchHeadCommit)
	if err != nil {
		return "", fmt.Errorf("unable to commit new changes in service branch: %w", err)
	}

	if err := writeDevIndexBase(worktreeCacheDir, newCommit, conversionSignature); err != nil {
		return "", fmt.Errorf("unable to write dev-index base marker: %w", err)
	}

	return newCommit, nil
}

func devIndexFilePath(worktreeCacheDir string) string {
	return filepath.Join(worktreeCacheDir, devIndexFileName)
}

func devIndexBaseFilePath(worktreeCacheDir string) string {
	return filepath.Join(worktreeCacheDir, devIndexBaseFileName)
}

func removeDevIndexFiles(worktreeCacheDir string) error {
	if err := os.RemoveAll(devIndexFilePath(worktreeCacheDir)); err != nil {
		return fmt.Errorf("unable to remove %s: %w", devIndexFilePath(worktreeCacheDir), err)
	}
	if err := os.RemoveAll(devIndexBaseFilePath(worktreeCacheDir)); err != nil {
		return fmt.Errorf("unable to remove %s: %w", devIndexBaseFilePath(worktreeCacheDir), err)
	}
	return nil
}

func devIndexBaseSignature(serviceBranchHeadCommit, conversionSignature string) string {
	sum := sha256.Sum256([]byte(serviceBranchHeadCommit + "\n" + conversionSignature))
	return hex.EncodeToString(sum[:])
}

func writeDevIndexBase(worktreeCacheDir, serviceBranchHeadCommit, conversionSignature string) error {
	sig := devIndexBaseSignature(serviceBranchHeadCommit, conversionSignature)
	return os.WriteFile(devIndexBaseFilePath(worktreeCacheDir), []byte(sig+"\n"), 0o644)
}

func devIndexBaseMatches(worktreeCacheDir, expected string) bool {
	data, err := os.ReadFile(devIndexBaseFilePath(worktreeCacheDir))
	if err != nil {
		return false
	}
	return strings.TrimSpace(string(data)) == expected
}

// prepareDevIndexAndWriteTree captures the source worktree into the persistent dev-index and
// returns the resulting tree SHA. The index is (re)seeded from serviceBranchHeadCommit only when
// its base marker is absent/mismatched (e.g. the service-branch head moved or the add-time
// conversion config changed), so unchanged files keep matching the source worktree's stat cache
// and are not re-read or re-hashed.
func prepareDevIndexAndWriteTree(ctx context.Context, sourceWorktreeDir, serviceWorktreeDir, worktreeCacheDir, sourceCommit, serviceBranchHeadCommit, conversionSignature string, globExcludeList []string) (string, error) {
	devIndexFile := devIndexFilePath(worktreeCacheDir)
	expectedBase := devIndexBaseSignature(serviceBranchHeadCommit, conversionSignature)

	needSeed := true
	if _, err := os.Stat(devIndexFile); err == nil && devIndexBaseMatches(worktreeCacheDir, expectedBase) {
		needSeed = false
	}
	if needSeed {
		if err := seedDevIndex(ctx, serviceWorktreeDir, devIndexFile, serviceBranchHeadCommit); err != nil {
			return "", fmt.Errorf("unable to seed dev-index from %s: %w", serviceBranchHeadCommit, err)
		}
	}

	if err := revertExcludedChangesInDevIndex(ctx, serviceWorktreeDir, devIndexFile, sourceCommit, serviceBranchHeadCommit, globExcludeList); err != nil {
		return "", fmt.Errorf("unable to revert excluded changes in dev-index: %w", err)
	}

	if err := runGitAddCmd(ctx, sourceWorktreeDir, serviceWorktreeDir, devIndexFile, globExcludeList); err != nil {
		return "", fmt.Errorf("unable to add source worktree changes to dev-index: %w", err)
	}

	treeSHA, err := writeDevIndexTree(ctx, serviceWorktreeDir, devIndexFile)
	if err != nil {
		return "", fmt.Errorf("unable to write tree from dev-index: %w", err)
	}

	return treeSHA, nil
}

func seedDevIndex(ctx context.Context, serviceWorktreeDir, devIndexFile, serviceBranchHeadCommit string) error {
	readTreeCmd := NewGitCmd(ctx, &GitCmdOptions{RepoDir: serviceWorktreeDir, Env: []string{"GIT_INDEX_FILE=" + devIndexFile}}, "read-tree", serviceBranchHeadCommit)
	if err := readTreeCmd.Run(ctx); err != nil {
		return fmt.Errorf("git read-tree command failed: %w", err)
	}
	return nil
}

func writeDevIndexTree(ctx context.Context, serviceWorktreeDir, devIndexFile string) (string, error) {
	writeTreeCmd := NewGitCmd(ctx, &GitCmdOptions{RepoDir: serviceWorktreeDir, Env: []string{"GIT_INDEX_FILE=" + devIndexFile}}, "write-tree")
	if err := writeTreeCmd.Run(ctx); err != nil {
		return "", fmt.Errorf("git write-tree command failed: %w", err)
	}
	return strings.TrimSpace(writeTreeCmd.OutBuf.String()), nil
}

func resolveTreeSHA(ctx context.Context, serviceWorktreeDir, commit string) (string, error) {
	revParseCmd := NewGitCmd(ctx, &GitCmdOptions{RepoDir: serviceWorktreeDir}, "rev-parse", commit+"^{tree}")
	if err := revParseCmd.Run(ctx); err != nil {
		return "", fmt.Errorf("git rev-parse tree command failed: %w", err)
	}
	return strings.TrimSpace(revParseCmd.OutBuf.String()), nil
}

// computeConversionConfigSignature captures everything that affects how `git add` converts working
// tree content into blobs, so a change forces a full reseed instead of silently keeping stale blobs
// for files whose own content and stat did not change. Read from git in the source worktree's own
// config context so global filters (e.g. git-lfs's ~/.gitconfig filter.lfs) are included.
func computeConversionConfigSignature(ctx context.Context, sourceWorktreeDir string) (string, error) {
	h := sha256.New()

	filterConfig, err := gitConfigFilterValues(ctx, sourceWorktreeDir)
	if err != nil {
		return "", err
	}
	fmt.Fprintf(h, "filter-config\x00%s\x00", filterConfig)

	attrFiles, err := trackedGitattributesFiles(ctx, sourceWorktreeDir)
	if err != nil {
		return "", err
	}
	for _, rel := range attrFiles {
		content, err := os.ReadFile(filepath.Join(sourceWorktreeDir, rel))
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				continue
			}
			return "", fmt.Errorf("unable to read %s: %w", rel, err)
		}
		fmt.Fprintf(h, "attr\x00%s\x00", rel)
		h.Write(content)
	}

	for _, extra := range extraAttributesFilePaths(ctx, sourceWorktreeDir) {
		content, err := os.ReadFile(extra)
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				continue
			}
			return "", fmt.Errorf("unable to read %s: %w", extra, err)
		}
		fmt.Fprintf(h, "extra-attr\x00%s\x00", extra)
		h.Write(content)
	}

	return hex.EncodeToString(h.Sum(nil)), nil
}

func gitConfigFilterValues(ctx context.Context, repoDir string) (string, error) {
	configCmd := NewGitCmd(ctx, &GitCmdOptions{RepoDir: repoDir}, "config", "--get-regexp", "^filter\\.")
	if err := configCmd.Run(ctx); err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) && exitErr.ExitCode() == 1 {
			return "", nil
		}
		return "", fmt.Errorf("git config command failed: %w", err)
	}
	return configCmd.OutBuf.String(), nil
}

func trackedGitattributesFiles(ctx context.Context, sourceWorktreeDir string) ([]string, error) {
	lsCmd := NewGitCmd(ctx, &GitCmdOptions{RepoDir: sourceWorktreeDir}, "ls-files", "-z", "--", ".gitattributes", ":(glob)**/.gitattributes")
	if err := lsCmd.Run(ctx); err != nil {
		return nil, fmt.Errorf("git ls-files command failed: %w", err)
	}
	var files []string
	for _, p := range strings.Split(lsCmd.OutBuf.String(), "\x00") {
		if p != "" {
			files = append(files, p)
		}
	}
	sort.Strings(files)
	return files, nil
}

func extraAttributesFilePaths(ctx context.Context, sourceWorktreeDir string) []string {
	var paths []string

	gitPathCmd := NewGitCmd(ctx, &GitCmdOptions{RepoDir: sourceWorktreeDir}, "rev-parse", "--git-path", "info/attributes")
	if err := gitPathCmd.Run(ctx); err == nil {
		if p := strings.TrimSpace(gitPathCmd.OutBuf.String()); p != "" {
			if !filepath.IsAbs(p) {
				p = filepath.Join(sourceWorktreeDir, p)
			}
			paths = append(paths, p)
		}
	}

	attrFileCmd := NewGitCmd(ctx, &GitCmdOptions{RepoDir: sourceWorktreeDir}, "config", "--get", "core.attributesFile")
	if err := attrFileCmd.Run(ctx); err == nil {
		if p := strings.TrimSpace(attrFileCmd.OutBuf.String()); p != "" {
			paths = append(paths, expandTilde(p))
		}
	}

	return paths
}

func expandTilde(path string) string {
	if path == "~" || strings.HasPrefix(path, "~/") {
		if home, err := os.UserHomeDir(); err == nil {
			return filepath.Join(home, strings.TrimPrefix(path[1:], "/"))
		}
	}
	return path
}

func prepareAndCheckoutServiceBranch(ctx context.Context, serviceWorktreeDir, sourceCommit, branchName string) error {
	branchListCmd := NewGitCmd(ctx, &GitCmdOptions{RepoDir: serviceWorktreeDir}, "branch", "--list", branchName)
	if err := branchListCmd.Run(ctx); err != nil {
		return fmt.Errorf("git branch list command failed: %w", err)
	}

	if branchListCmd.OutBuf.Len() == 0 {
		checkoutCmd := NewGitCmd(ctx, &GitCmdOptions{RepoDir: serviceWorktreeDir}, "checkout", "-b", branchName, sourceCommit)
		if err := checkoutCmd.Run(ctx); err != nil {
			return fmt.Errorf("git checkout command failed: %w", err)
		}

		return nil
	}

	checkoutCmd := NewGitCmd(ctx, &GitCmdOptions{RepoDir: serviceWorktreeDir}, "checkout", branchName)
	if err := checkoutCmd.Run(ctx); err != nil {
		return fmt.Errorf("git checkout command failed: %w", err)
	}

	isSourceCommitInServiceBranch, err := IsAncestor(ctx, sourceCommit, branchName, serviceWorktreeDir)
	if err != nil {
		return fmt.Errorf("unable to detect whether sourceCommit %q is in service branch: %w", sourceCommit, err)
	}
	if isSourceCommitInServiceBranch {
		return nil
	}

	mergeCmd := NewGitCmd(
		ctx, &GitCmdOptions{RepoDir: serviceWorktreeDir},
		"-c", "user.email=werf@werf.io", "-c", "user.name=werf",
		"merge", "--no-verify", "--no-edit", "--no-ff", "--allow-unrelated-histories", "-s", "ours", sourceCommit,
	)
	if err = mergeCmd.Run(ctx); err != nil {
		return fmt.Errorf("git merge of source commit %q into service branch %q failed: %w\nNOTE: To continue you can remove the service branch %q with \"git branch -D %s\", but we would also ask you to report this issue to https://github.com/werf/werf/issues", sourceCommit, branchName, err, branchName, branchName)
	}

	return nil
}

func revertExcludedChangesInDevIndex(ctx context.Context, serviceWorktreeDir, devIndexFile, sourceCommit, serviceBranchHeadCommit string, globExcludeList []string) error {
	if len(globExcludeList) == 0 || serviceBranchHeadCommit == sourceCommit {
		return nil
	}

	gitDiffArgs := []string{
		"-c", "diff.renames=false",
		"-c", "core.quotePath=false",
		"diff",
		"--binary",
		serviceBranchHeadCommit, sourceCommit,
		"--",
	}
	gitDiffArgs = append(gitDiffArgs, globExcludeList...)

	diffCmd := NewGitCmd(ctx, &GitCmdOptions{RepoDir: serviceWorktreeDir}, gitDiffArgs...)
	if err := diffCmd.Run(ctx); err != nil {
		return fmt.Errorf("git diff command failed: %w", err)
	}

	if diffCmd.OutBuf.Len() == 0 {
		return nil
	}

	applyCmd := NewGitCmd(ctx, &GitCmdOptions{RepoDir: serviceWorktreeDir, Env: []string{"GIT_INDEX_FILE=" + devIndexFile}}, "apply", "--binary", "--cached")
	applyCmd.Stdin = diffCmd.OutBuf
	if err := applyCmd.Run(ctx); err != nil {
		return fmt.Errorf("git apply command failed: %w", err)
	}

	return nil
}

func runGitAddCmd(ctx context.Context, sourceWorktreeDir, serviceWorktreeDir, devIndexFile string, globExcludeList []string) error {
	gitAddArgs := []string{
		"--work-tree",
		sourceWorktreeDir,
		"add",
	}

	var pathSpecList []string
	{
		pathSpecList = append(pathSpecList, ":.")
		for _, glob := range globExcludeList {
			pathSpecList = append(pathSpecList, ":!"+glob)
		}
	}

	var pathSpecFileBuf *bytes.Buffer
	if gitVersion.LessThan(semver.MustParse("2.25.0")) {
		gitAddArgs = append(gitAddArgs, "--")
		gitAddArgs = append(gitAddArgs, pathSpecList...)
	} else {
		gitAddArgs = append(gitAddArgs, "--pathspec-from-file=-", "--pathspec-file-nul")
		pathSpecFileBuf = bytes.NewBufferString(strings.Join(pathSpecList, "\000"))
	}

	addCmd := NewGitCmd(ctx, &GitCmdOptions{RepoDir: serviceWorktreeDir, Env: []string{"GIT_INDEX_FILE=" + devIndexFile}}, gitAddArgs...)
	if pathSpecFileBuf != nil {
		addCmd.Stdin = pathSpecFileBuf
	}
	if err := addCmd.Run(ctx); err != nil {
		return err
	}

	return nil
}

// commitTreeToServiceBranch creates the service commit with plumbing (commit-tree + update-ref)
// rather than `git commit`: `git commit` would refresh the index against the service worktree,
// poisoning the source-worktree stat cache the dev-index relies on, and could run repo hooks.
func commitTreeToServiceBranch(ctx context.Context, serviceWorktreeDir, branchName, treeSHA, parentCommit string) (string, error) {
	commitTreeCmd := NewGitCmd(
		ctx, &GitCmdOptions{RepoDir: serviceWorktreeDir},
		"-c", "user.email=werf@werf.io", "-c", "user.name=werf",
		"commit-tree", treeSHA, "-p", parentCommit, "-m", time.Now().String(),
	)
	if err := commitTreeCmd.Run(ctx); err != nil {
		return "", fmt.Errorf("git commit-tree command failed: %w", err)
	}
	serviceNewCommit := strings.TrimSpace(commitTreeCmd.OutBuf.String())

	updateRefCmd := NewGitCmd(ctx, &GitCmdOptions{RepoDir: serviceWorktreeDir}, "update-ref", "refs/heads/"+branchName, serviceNewCommit)
	if err := updateRefCmd.Run(ctx); err != nil {
		return "", fmt.Errorf("git update-ref command failed: %w", err)
	}

	checkoutCmd := NewGitCmd(ctx, &GitCmdOptions{RepoDir: serviceWorktreeDir}, "checkout", "--force", "--detach", serviceNewCommit)
	if err := checkoutCmd.Run(ctx); err != nil {
		return "", fmt.Errorf("git checkout command failed: %w", err)
	}

	return serviceNewCommit, nil
}
