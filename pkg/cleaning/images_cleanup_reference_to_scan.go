package cleaning

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/werf/logboek"
)

type referenceToScan struct {
	*plumbing.Reference
	HeadCommit *object.Commit
	referenceScanOptions
}

type referenceScanOptions struct {
	scanDepthLimit   int
	imageDepthToKeep int
}

func (opts *referenceScanOptions) scanDepthLimitLogString() string {
	if opts.scanDepthLimit == 0 {
		return "∞"
	}

	return fmt.Sprint(opts.scanDepthLimit)
}

func getReferencesToScan(gitRepository *git.Repository) ([]*referenceToScan, error) {
	rs, err := gitRepository.References()
	if err != nil {
		return nil, fmt.Errorf("get repository references failed: %s", err)
	}

	var refs []*referenceToScan
	if err := rs.ForEach(func(reference *plumbing.Reference) error {
		n := reference.Name()

		// Get all remote branches and tags
		if !(n.IsRemote() || n.IsTag()) {
			return nil
		}

		// Use only origin upstream
		if n.IsRemote() && !strings.HasPrefix(n.Short(), "origin/") {
			return nil
		}

		refHash := reference.Hash()
		if n.IsTag() {
			revHash, err := gitRepository.ResolveRevision(plumbing.Revision(n.Short()))
			if err != nil {
				return fmt.Errorf("resolve revision %s failed: %s", n.Short(), err)
			}

			refHash = *revHash
		}

		if refHash == plumbing.ZeroHash {
			return nil
		}

		refCommit, err := gitRepository.CommitObject(refHash)
		if err != nil {
			return fmt.Errorf("reference %s: commit object %s failed: %s", n.Short(), refHash.String(), err)
		}

		refs = append(refs, &referenceToScan{
			Reference:  reference,
			HeadCommit: refCommit,
		})

		return nil
	}); err != nil {
		return nil, err
	}

	// Sort by committer when
	sort.Slice(refs, func(i, j int) bool {
		return refs[i].HeadCommit.Committer.When.After(refs[j].HeadCommit.Committer.When)
	})

	// Split branches and tags references
	var branchesRefs, tagsRefs []*referenceToScan
	for _, ref := range refs {
		if ref.Name().IsTag() {
			tagsRefs = append(tagsRefs, ref)
		} else {
			branchesRefs = append(branchesRefs, ref)
		}
	}

	// TODO: Skip by regexps

	// Filter by modifiedIn
	//branchesRefs = skipByPeriod(branchesRefs, period)
	//tagsRefs = skipByPeriod(tagsRefs, period)

	// Filter by last
	branchesRefs = filterReferencesByLast(branchesRefs, 15)
	tagsRefs = filterReferencesByLast(tagsRefs, 10)

	// TODO: Set depth / reached commits limit

	// Set defaults
	for _, tagRef := range tagsRefs {
		tagRef.referenceScanOptions.scanDepthLimit = 1
		tagRef.referenceScanOptions.imageDepthToKeep = 1
	}

	for _, branchRef := range branchesRefs {
		branchRef.referenceScanOptions.imageDepthToKeep = 5
	}

	// Unite tags and branches references
	result := append(branchesRefs, tagsRefs...)

	return result, nil
}

func filterReferencesByModifiedIn(refs []*referenceToScan, modifiedIn time.Duration) (result []*referenceToScan) {
	for _, ref := range refs {
		if ref.HeadCommit.Committer.When.Before(time.Now().Add(-modifiedIn)) {
			logboek.LogF("Reference %s filtered by the modifiedIn parameter\n", ref.Name().Short())
			continue
		}

		result = append(result, ref)
	}

	return
}

func filterReferencesByLast(refs []*referenceToScan, last int) []*referenceToScan {
	if len(refs) < last {
		return refs
	}

	for _, ref := range refs[last:] {
		logboek.Debug.LogF("Reference %s filtered by the last parameter\n", ref.Name().Short())
	}

	return refs[:last]
}
