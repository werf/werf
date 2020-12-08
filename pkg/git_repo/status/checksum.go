package status

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"hash"
	"sort"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/utils/merkletrie/filesystem"
	mindex "github.com/go-git/go-git/v5/utils/merkletrie/index"
	"github.com/go-git/go-git/v5/utils/merkletrie/noder"

	"github.com/werf/logboek"

	"github.com/werf/werf/pkg/util"
)

type ChecksumOptions struct {
	FilterOptions
}

type FilterOptions struct {
	OnlyStaged   bool
	ExceptStaged bool
}

func (r *Result) Checksum(ctx context.Context, options ChecksumOptions) (string, error) {
	if r.IsEmpty(options.FilterOptions) {
		return "", nil
	}

	h := sha256.New()

	fileStatusPathListChecksum, err := r.calculateFileStatusPathListChecksum(ctx, options)
	if err != nil {
		return "", err
	}

	if fileStatusPathListChecksum != "" {
		h.Write([]byte(fileStatusPathListChecksum))
	}

	submoduleResultsChecksum, err := r.calculateSubmoduleResultsChecksum(ctx, options)
	if err != nil {
		return "", err
	}

	if submoduleResultsChecksum != "" {
		h.Write([]byte(submoduleResultsChecksum))
	}

	return fmt.Sprintf("%x", h.Sum(nil)), nil
}

func (r *Result) calculateFileStatusPathListChecksum(ctx context.Context, options ChecksumOptions) (string, error) {
	fileStatusPathList := r.filteredFilePathList(options.FilterOptions)
	sort.Strings(fileStatusPathList)

	if len(fileStatusPathList) == 0 {
		return "", nil
	}

	var n noder.Noder
	if options.OnlyStaged {
		i, err := r.repository.Storer.Index()
		if err != nil {
			return "", fmt.Errorf("unable to get git repository index: %s", err)
		}

		n = mindex.NewRootNode(i)
	} else {
		w, err := r.repository.Worktree()
		if err != nil {
			return "", fmt.Errorf("unable to get git repository worktree: %s", err)
		}

		submodulesStatus, err := getSubmodulesStatus(w)
		if err != nil {
			return "", fmt.Errorf("unable to get git submodules status: %s", err)
		}

		n = filesystem.NewRootNode(w.Filesystem, submodulesStatus)
	}

	h := sha256.New()
	if err := r.calculateChecksumForNoderDir(ctx, h, n, fileStatusPathList); err != nil {
		return "", fmt.Errorf("unable to calculate checksum: %s", err)
	}

	return fmt.Sprintf("%x", h.Sum(nil)), nil
}

func (r *Result) calculateChecksumForNoderDir(ctx context.Context, h hash.Hash, n noder.Noder, fileStatusPathList []string) error {
	children, err := n.Children()
	if err != nil {
		return err
	}

	fileStatusPathProcessed := map[string]bool{}
	for _, child := range children {
		var fileStatusPathListToGoThrough []string

		for _, fileStatusPath := range fileStatusPathList {
			if !child.IsDir() && child.String() == fileStatusPath {
				fileStatusPathProcessed[fileStatusPath] = true
				h.Write(child.Hash())

				fileStatus := r.fileStatusList[fileStatusPath]
				logboek.Context(ctx).Debug().LogF("Entry was processed: %s\n", fileStatusPath)
				logboek.Context(ctx).Debug().LogF("  worktree %s staging %s result %v\n", fileStatusMapping[fileStatus.Worktree], fileStatusMapping[fileStatus.Staging], hex.EncodeToString(child.Hash()[:]))

				continue
			} else if util.IsSubpathOfBasePath(child.String(), fileStatusPath) {
				fileStatusPathProcessed[fileStatusPath] = true
				fileStatusPathListToGoThrough = append(fileStatusPathListToGoThrough, fileStatusPath)
			}
		}

		if child.IsDir() && len(fileStatusPathListToGoThrough) != 0 {
			if err := r.calculateChecksumForNoderDir(ctx, h, child, fileStatusPathListToGoThrough); err != nil {
				return err
			}
		}
	}

	for _, fileStatusPath := range fileStatusPathList {
		if !fileStatusPathProcessed[fileStatusPath] {
			h.Write([]byte(fileStatusPath))
			h.Write([]byte(plumbing.ZeroHash.String()))

			fileStatus := r.fileStatusList[fileStatusPath]
			logboek.Context(ctx).Debug().LogF("Entry was processed: %s\n", fileStatusPath)
			logboek.Context(ctx).Debug().LogF("  worktree %s staging %s result %v\n", fileStatusMapping[fileStatus.Worktree], fileStatusMapping[fileStatus.Staging], plumbing.ZeroHash.String())
		}
	}

	return nil
}

func getSubmodulesStatus(w *git.Worktree) (map[string]plumbing.Hash, error) {
	o := map[string]plumbing.Hash{}

	sub, err := w.Submodules()
	if err != nil {
		return nil, err
	}

	status, err := sub.Status()
	if err != nil {
		return nil, err
	}

	for _, s := range status {
		if s.Current.IsZero() {
			o[s.Path] = s.Expected
			continue
		}

		o[s.Path] = s.Current
	}

	return o, nil
}

func (r *Result) calculateSubmoduleResultsChecksum(ctx context.Context, options ChecksumOptions) (string, error) {
	h := sha256.New()
	isEmpty := true

	sort.Slice(r.submoduleResults, func(i, j int) bool {
		return r.submoduleResults[i].repositoryFullFilepath < r.submoduleResults[j].repositoryFullFilepath
	})

	for _, sr := range r.submoduleResults {
		logboek.Context(ctx).Debug().LogOptionalLn()
		if err := logboek.Context(ctx).Debug().LogBlock("submodule %s", sr.repositoryFullFilepath).DoError(func() error {
			var srChecksumArgs []string

			srChecksumArgs = append(srChecksumArgs, sr.repositoryFullFilepath)

			if sr.isNotInitialized {
				srChecksumArgs = append(srChecksumArgs, "isNotInitialized")
				return nil
			} else {
				if sr.isNotClean {
					srChecksumArgs = append(srChecksumArgs, "isNotClean")
					srChecksumArgs = append(srChecksumArgs, sr.currentCommit)
				}

				srChecksum, err := sr.Checksum(ctx, options)
				if err != nil {
					return err
				}

				if srChecksum != "" {
					srChecksumArgs = append(srChecksumArgs, srChecksum)
				}
			}

			logboek.Context(ctx).Debug().LogF("Args was added: %v\n", srChecksumArgs)
			h.Write([]byte(strings.Join(srChecksumArgs, "🐜")))
			isEmpty = false

			return nil
		}); err != nil {
			return "", fmt.Errorf("submodule %s checksum failed: %s", sr.repositoryFullFilepath, err)
		}
	}

	if isEmpty {
		return "", nil
	} else {
		return fmt.Sprintf("%x", h.Sum(nil)), nil
	}
}
