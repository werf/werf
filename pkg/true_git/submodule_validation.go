package true_git

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/filemode"
	"github.com/go-git/go-git/v5/plumbing/object"
)

type SubmoduleValidationResult struct {
	Valid  bool
	Errors []SubmoduleValidationError
}

type SubmoduleValidationError struct {
	SubmodulePath  string
	ExpectedCommit string
	ActualCommit   string
	Initialized    bool
	Message        string
}

func ValidateSubmoduleState(ctx context.Context, repository *git.Repository, commitHash plumbing.Hash, workTreeDir string) (*SubmoduleValidationResult, error) {
	result := &SubmoduleValidationResult{Valid: true}

	commit, err := repository.CommitObject(commitHash)
	if err != nil {
		return nil, fmt.Errorf("get commit object %s: %w", commitHash, err)
	}

	tree, err := commit.Tree()
	if err != nil {
		return nil, fmt.Errorf("get commit tree %s: %w", commitHash, err)
	}

	gitmodulesFile, err := tree.File(".gitmodules")
	if err != nil {
		if errors.Is(err, object.ErrFileNotFound) {
			return result, nil
		}
		return nil, fmt.Errorf("read .gitmodules entry: %w", err)
	}

	gitmodulesContent, err := gitmodulesFile.Contents()
	if err != nil {
		return nil, fmt.Errorf("read .gitmodules: %w", err)
	}

	modules := config.NewModules()
	if err := modules.Unmarshal([]byte(gitmodulesContent)); err != nil {
		return nil, fmt.Errorf("parse .gitmodules: %w", err)
	}

	for _, module := range modules.Submodules {
		subResult, err := validateSubmodule(ctx, tree, module, workTreeDir)
		if err != nil {
			return nil, err
		}
		if !subResult.Valid {
			result.Valid = false
			result.Errors = append(result.Errors, subResult.Errors...)
		}
	}

	return result, nil
}

func validateSubmodule(ctx context.Context, tree *object.Tree, module *config.Submodule, workTreeDir string) (*SubmoduleValidationResult, error) {
	entry, err := tree.FindEntry(module.Path)
	if err != nil {
		return nil, fmt.Errorf("find submodule entry %q: %w", module.Path, err)
	}

	if entry.Mode != filemode.Submodule {
		return nil, fmt.Errorf("submodule %q has invalid mode %s", module.Path, entry.Mode.String())
	}

	result := &SubmoduleValidationResult{Valid: true}
	expected := entry.Hash
	submoduleDir := filepath.Join(workTreeDir, module.Path)

	if _, err := os.Stat(submoduleDir); err != nil {
		if os.IsNotExist(err) {
			result.Errors = append(result.Errors, newUninitializedError(module.Path, expected))
			result.Valid = false
			return result, nil
		}
		return nil, fmt.Errorf("check submodule dir %q: %w", submoduleDir, err)
	}

	subRepo, err := git.PlainOpen(submoduleDir)
	if err != nil {
		result.Errors = append(result.Errors, newUninitializedError(module.Path, expected))
		result.Valid = false
		return result, nil
	}

	headRef, err := subRepo.Head()
	if err != nil {
		return nil, fmt.Errorf("get submodule head %q: %w", module.Path, err)
	}

	actual := headRef.Hash()
	if actual != expected {
		result.Errors = append(result.Errors, newWrongCommitError(module.Path, expected, actual))
		result.Valid = false
		return result, nil
	}

	nestedResult, err := ValidateSubmoduleState(ctx, subRepo, expected, submoduleDir)
	if err != nil {
		return nil, err
	}
	if !nestedResult.Valid {
		result.Valid = false
		result.Errors = append(result.Errors, nestedResult.Errors...)
	}

	return result, nil
}

func newUninitializedError(path string, expected plumbing.Hash) SubmoduleValidationError {
	return SubmoduleValidationError{
		SubmodulePath:  path,
		ExpectedCommit: expected.String(),
		ActualCommit:   "",
		Initialized:    false,
		Message:        fmt.Sprintf("submodule %q is not initialized. Run: git submodule update --init --recursive", path),
	}
}

func newWrongCommitError(path string, expected, actual plumbing.Hash) SubmoduleValidationError {
	return SubmoduleValidationError{
		SubmodulePath:  path,
		ExpectedCommit: expected.String(),
		ActualCommit:   actual.String(),
		Initialized:    true,
		Message:        fmt.Sprintf("submodule %q: expected commit %s, got %s. Run: git submodule update --init --recursive", path, expected, actual),
	}
}
