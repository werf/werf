package true_git

import (
	"bytes"
	"context"
	"encoding/hex"
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/filemode"
	fdiff "github.com/go-git/go-git/v5/plumbing/format/diff"
	"github.com/go-git/go-git/v5/plumbing/object"

	"github.com/werf/common-go/pkg/util"
)

const entireFileContextLines = 1 << 30

func GoGitPatch(ctx context.Context, out io.Writer, gitDir, workTreeDir string, withSubmodules bool, opts PatchOptions) (*PatchDescriptor, error) {
	if out == nil {
		out = io.Discard
	}

	absGitDir, err := filepath.Abs(gitDir)
	if err != nil {
		return nil, fmt.Errorf("bad git dir %s: %w", gitDir, err)
	}

	repository, err := PlainOpenWithOptions(absGitDir, &PlainOpenOptions{EnableDotGitCommonDir: true})
	if err != nil {
		return nil, fmt.Errorf("cannot open git repo %q: %w", absGitDir, err)
	}

	if opts.ToCommit == "" {
		return nil, fmt.Errorf("to commit is required")
	}

	toHash, err := parseCommitHash(opts.ToCommit)
	if err != nil {
		return nil, fmt.Errorf("bad `to` commit hash %q: %w", opts.ToCommit, err)
	}
	if toHash.IsZero() {
		return nil, fmt.Errorf("bad `to` commit hash %q", opts.ToCommit)
	}

	toCommit, err := repository.CommitObject(toHash)
	if err != nil {
		return nil, fmt.Errorf("bad `to` commit %q: %w", opts.ToCommit, err)
	}

	toTree, err := toCommit.Tree()
	if err != nil {
		return nil, fmt.Errorf("get `to` commit tree %q: %w", opts.ToCommit, err)
	}

	var fromTree *object.Tree
	if opts.FromCommit != "" {
		fromHash, err := parseCommitHash(opts.FromCommit)
		if err != nil {
			return nil, fmt.Errorf("bad `from` commit hash %q: %w", opts.FromCommit, err)
		}

		if !fromHash.IsZero() {
			fromCommit, err := repository.CommitObject(fromHash)
			if err != nil {
				return nil, fmt.Errorf("bad `from` commit %q: %w", opts.FromCommit, err)
			}

			fromTree, err = fromCommit.Tree()
			if err != nil {
				return nil, fmt.Errorf("get `from` commit tree %q: %w", opts.FromCommit, err)
			}
		}
	}

	changes, err := object.DiffTreeWithOptions(ctx, fromTree, toTree, &object.DiffTreeOptions{DetectRenames: false})
	if err != nil {
		return nil, fmt.Errorf("diff commits %q..%q: %w", opts.FromCommit, opts.ToCommit, err)
	}

	parser := makeDiffParser(out, opts.PathScope, opts.PathMatcher, opts.FileRenames)

	var filePatches []fdiff.FilePatch
	if withSubmodules {
		if workTreeDir == "" {
			return nil, fmt.Errorf("work tree dir is required to handle submodules")
		}

		filePatches, err = applyChangesWithSubmodules(ctx, parser, workTreeDir, changes, "", opts)
		if err != nil {
			return nil, err
		}
	} else {
		filePatches, err = applyChangesPatch(ctx, parser, changes, "", opts)
		if err != nil {
			return nil, err
		}
	}

	binaryPaths := parser.BinaryPaths
	if opts.WithBinary {
		binaryPaths = appendBinaryPaths(binaryPaths, filePatches, opts)
	}

	desc := &PatchDescriptor{
		Paths:         parser.Paths,
		BinaryPaths:   binaryPaths,
		PathsToRemove: parser.PathsToRemove,
	}

	return desc, nil
}

func applyChangesWithSubmodules(ctx context.Context, parser *diffParser, workTreeDir string, changes object.Changes, prefix string, opts PatchOptions) ([]fdiff.FilePatch, error) {
	regularChanges, submoduleChanges := splitSubmoduleChanges(changes)

	filePatches, err := applyChangesPatch(ctx, parser, regularChanges, prefix, opts)
	if err != nil {
		return nil, err
	}

	for _, change := range submoduleChanges {
		submodulePath := changePath(change)
		if submodulePath == "" {
			return nil, fmt.Errorf("submodule change has empty path")
		}

		submodulePrefix := joinSubmodulePath(prefix, submodulePath)
		submoduleWorkTreeDir := filepath.Join(workTreeDir, submodulePath)
		submoduleRepo, err := PlainOpenWithOptions(submoduleWorkTreeDir, &PlainOpenOptions{EnableDotGitCommonDir: true})
		if err != nil {
			return nil, fmt.Errorf("open submodule repo %q: %w", submodulePrefix, err)
		}

		fromTree, err := commitTreeFromHash(submoduleRepo, change.From.TreeEntry.Hash)
		if err != nil {
			return nil, fmt.Errorf("get submodule %q from tree: %w", submodulePrefix, err)
		}

		toTree, err := commitTreeFromHash(submoduleRepo, change.To.TreeEntry.Hash)
		if err != nil {
			return nil, fmt.Errorf("get submodule %q to tree: %w", submodulePrefix, err)
		}

		subChanges, err := object.DiffTreeWithOptions(ctx, fromTree, toTree, &object.DiffTreeOptions{DetectRenames: false})
		if err != nil {
			return nil, fmt.Errorf("diff submodule %q: %w", submodulePrefix, err)
		}

		subFilePatches, err := applyChangesWithSubmodules(ctx, parser, submoduleWorkTreeDir, subChanges, submodulePrefix, opts)
		if err != nil {
			return nil, err
		}
		filePatches = append(filePatches, subFilePatches...)
	}

	return filePatches, nil
}

func applyChangesPatch(ctx context.Context, parser *diffParser, changes object.Changes, prefix string, opts PatchOptions) ([]fdiff.FilePatch, error) {
	if len(changes) == 0 {
		return nil, nil
	}

	patch, err := prefixedChanges(changes, prefix).PatchContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("build patch: %w", err)
	}

	var patchBuf bytes.Buffer
	if opts.WithEntireFileContext {
		encoder := fdiff.NewUnifiedEncoder(&patchBuf, entireFileContextLines)
		if err := encoder.Encode(patch); err != nil {
			return nil, fmt.Errorf("encode patch: %w", err)
		}
	} else {
		if err := patch.Encode(&patchBuf); err != nil {
			return nil, fmt.Errorf("encode patch: %w", err)
		}
	}

	if patchBuf.Len() > 0 {
		if err := parser.HandleStdout(patchBuf.Bytes()); err != nil {
			return nil, err
		}
	}

	return patch.FilePatches(), nil
}

func splitSubmoduleChanges(changes object.Changes) (object.Changes, []*object.Change) {
	var regularChanges object.Changes
	var submoduleChanges []*object.Change
	for _, change := range changes {
		if isSubmoduleChange(change) {
			submoduleChanges = append(submoduleChanges, change)
			continue
		}
		regularChanges = append(regularChanges, change)
	}

	return regularChanges, submoduleChanges
}

func isSubmoduleChange(change *object.Change) bool {
	return change.From.TreeEntry.Mode == filemode.Submodule || change.To.TreeEntry.Mode == filemode.Submodule
}

func prefixedChanges(changes object.Changes, prefix string) object.Changes {
	if prefix == "" {
		return changes
	}

	prefixed := make(object.Changes, 0, len(changes))
	for _, change := range changes {
		prefixed = append(prefixed, prefixedChange(change, prefix))
	}

	return prefixed
}

func prefixedChange(change *object.Change, prefix string) *object.Change {
	if prefix == "" {
		return change
	}

	prefixed := *change
	prefixed.From = prefixedChangeEntry(change.From, prefix)
	prefixed.To = prefixedChangeEntry(change.To, prefix)
	return &prefixed
}

func prefixedChangeEntry(entry object.ChangeEntry, prefix string) object.ChangeEntry {
	if entry.Name == "" {
		return entry
	}

	entry.Name = joinSubmodulePath(prefix, entry.Name)
	return entry
}

func changePath(change *object.Change) string {
	if change.From.Name != "" {
		return change.From.Name
	}

	return change.To.Name
}

func joinSubmodulePath(prefix, path string) string {
	if prefix == "" {
		return filepath.ToSlash(path)
	}

	return filepath.ToSlash(filepath.Join(prefix, path))
}

func commitTreeFromHash(repository *git.Repository, hash plumbing.Hash) (*object.Tree, error) {
	if hash.IsZero() {
		return nil, nil
	}

	commit, err := repository.CommitObject(hash)
	if err != nil {
		return nil, fmt.Errorf("get commit %s: %w", hash, err)
	}

	tree, err := commit.Tree()
	if err != nil {
		return nil, fmt.Errorf("get commit tree %s: %w", hash, err)
	}

	return tree, nil
}

func parseCommitHash(value string) (plumbing.Hash, error) {
	var hash plumbing.Hash
	if value == "" {
		return hash, nil
	}

	decoded, err := hex.DecodeString(value)
	if err != nil {
		return hash, err
	}

	copy(hash[:], decoded)
	return hash, nil
}

func appendBinaryPaths(paths []string, filePatches []fdiff.FilePatch, opts PatchOptions) []string {
	for _, filePatch := range filePatches {
		if !filePatch.IsBinary() {
			continue
		}

		from, to := filePatch.Files()
		if from != nil {
			paths = appendBinaryPath(paths, from.Path(), opts)
		}
		if to != nil {
			paths = appendBinaryPath(paths, to.Path(), opts)
		}
	}

	return paths
}

func appendBinaryPath(paths []string, path string, opts PatchOptions) []string {
	if !opts.PathMatcher.IsPathMatched(path) {
		return paths
	}

	path = applyFileRenames(path, opts.PathScope, opts.FileRenames)
	path = trimFileBasePath(path, opts.PathScope)

	return appendUnique(paths, path)
}

func applyFileRenames(path, pathScope string, fileRenames map[string]string) string {
	if renamedFileName, willRename := fileRenames[path]; willRename {
		return filepath.ToSlash(filepath.Join(pathScope, renamedFileName))
	}
	return path
}

func trimFileBasePath(path, pathScope string) string {
	newPath := filepath.ToSlash(util.GetRelativeToBaseFilepath(filepath.FromSlash(pathScope), filepath.FromSlash(path)))
	return strings.TrimRight(newPath, "\t")
}
