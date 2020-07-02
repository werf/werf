package cleaning

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"

	"github.com/werf/logboek"

	"github.com/werf/werf/pkg/config"
)

type referenceToScan struct {
	*plumbing.Reference
	ModifiedAt time.Time
	HeadCommit *object.Commit
	referenceScanOptions
}

type referenceScanOptions struct {
	scanDepthLimit   int
	imageDepthToKeep int
}

func getReferencesToScan(gitRepository *git.Repository, policies []*config.MetaCleanupPolicy) ([]*referenceToScan, error) {
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

		var scanDepthLimit int
		var modifiedAt time.Time
		var refCommit *object.Commit
		if !n.IsTag() {
			scanDepthLimit = -1 // unlimited

			refHash := reference.Hash()
			if refHash == plumbing.ZeroHash {
				return nil
			}

			refCommit, err = gitRepository.CommitObject(refHash)
			if err != nil {
				return fmt.Errorf("reference %s: commit object %s failed: %s", n.Short(), refHash.String(), err)
			}

			modifiedAt = refCommit.Committer.When
		} else {
			scanDepthLimit = 1

			revHash, err := gitRepository.ResolveRevision(plumbing.Revision(n.Short()))
			if err != nil {
				return fmt.Errorf("resolve revision %s failed: %s", n.Short(), err)
			}

			refCommit, err = gitRepository.CommitObject(*revHash)
			if err != nil {
				return fmt.Errorf("reference %s: commit object %s failed: %s", n.Short(), revHash.String(), err)
			}

			tagObject, err := gitRepository.TagObject(reference.Hash())
			if err == nil {
				modifiedAt = tagObject.Tagger.When
			} else if err == plumbing.ErrObjectNotFound { // lightweight tag
				modifiedAt = refCommit.Committer.When
			} else {
				return fmt.Errorf("tag object %s failed: %s", reference.Hash(), err)
			}
		}

		refs = append(refs, &referenceToScan{
			Reference:  reference,
			ModifiedAt: modifiedAt,
			HeadCommit: refCommit,
			referenceScanOptions: referenceScanOptions{
				scanDepthLimit: scanDepthLimit,
			},
		})

		return nil
	}); err != nil {
		return nil, err
	}

	// Split branches and tags references
	var branchesRefs, tagsRefs []*referenceToScan
	for _, ref := range refs {
		if ref.Name().IsTag() {
			tagsRefs = append(tagsRefs, ref)
		} else {
			branchesRefs = append(branchesRefs, ref)
		}
	}

	// Apply user or default policies
	if len(policies) == 0 {
		tagLastDefault := 10
		tagImageDepthToKeep := 1
		policies = append(policies, &config.MetaCleanupPolicy{
			TagRegexp: regexp.MustCompile(".*"),
			RefsToKeepImagesIn: &config.RefsToKeepImagesIn{
				Last:     &tagLastDefault,
				Operator: config.OrOperator,
			},
			ImageDepthToKeep: &tagImageDepthToKeep,
		})

		branchLastDefault := 10
		branchModifiedInDefault := time.Hour * 24 * 7
		branchImageDepthToKeepDefault := 2
		policies = append(policies, &config.MetaCleanupPolicy{
			BranchRegexp: regexp.MustCompile(".*"),
			RefsToKeepImagesIn: &config.RefsToKeepImagesIn{
				Last:       &branchLastDefault,
				ModifiedIn: &branchModifiedInDefault,
				Operator:   config.AndOperator,
			},
			ImageDepthToKeep: &branchImageDepthToKeepDefault,
		})

		mainBranchImageDepthToKeepDefault := 10
		policies = append(policies, &config.MetaCleanupPolicy{
			BranchRegexp:     regexp.MustCompile("^master|staging|production$"),
			ImageDepthToKeep: &mainBranchImageDepthToKeepDefault,
		})
	}

	var resultTagsRefs, resultBranchesRefs []*referenceToScan
	for _, policy := range policies {
		var policyRefs []*referenceToScan

		if policy.BranchRegexp != nil {
			policyRefs = selectBranchReferencesByRegexp(branchesRefs, policy.BranchRegexp)
			policyRefs = applyCleanupPolicy(policyRefs, policy)
			resultBranchesRefs = mergeReferences(resultBranchesRefs, policyRefs)
		} else if policy.TagRegexp != nil {
			policyRefs = selectTagReferencesByRegexp(tagsRefs, policy.TagRegexp)
			policyRefs = applyCleanupPolicy(policyRefs, policy)
			resultTagsRefs = mergeReferences(resultTagsRefs, policyRefs)
		}

		_ = logboek.Default.LogBlock(policy.String(), logboek.LevelLogBlockOptions{}, func() error {
			for _, ref := range policyRefs {
				logboek.Default.LogLnDetails(ref.Name().Short())
			}

			return nil
		})
	}

	// Sort by ModifiedAt
	sort.Slice(resultBranchesRefs, func(i, j int) bool {
		return resultBranchesRefs[i].ModifiedAt.After(resultBranchesRefs[j].ModifiedAt)
	})
	sort.Slice(resultTagsRefs, func(i, j int) bool {
		return resultTagsRefs[i].ModifiedAt.After(resultTagsRefs[j].ModifiedAt)
	})

	// Unite tags and branches references
	result := append(resultBranchesRefs, resultTagsRefs...)

	return result, nil
}

func selectBranchReferencesByRegexp(branchesRefs []*referenceToScan, regexp *regexp.Regexp) []*referenceToScan {
	var result []*referenceToScan

	for _, branchRef := range branchesRefs {
		refShortNameWithoutRemote := strings.SplitN(branchRef.Name().Short(), "/", 2)[1]
		if regexp.MatchString(refShortNameWithoutRemote) {
			result = append(result, branchRef)
		}
	}

	return result
}

func selectTagReferencesByRegexp(tagsRefs []*referenceToScan, regexp *regexp.Regexp) []*referenceToScan {
	var result []*referenceToScan

	for _, tagRef := range tagsRefs {
		if regexp.MatchString(tagRef.Name().Short()) {
			result = append(result, tagRef)
		}
	}

	return result
}

func applyCleanupPolicy(refs []*referenceToScan, policy *config.MetaCleanupPolicy) []*referenceToScan {
	if policy.RefsToKeepImagesIn != nil {
		refs = applyRefsToKeepImagesInPolicy(refs, policy.RefsToKeepImagesIn)
	}

	if policy.ImageDepthToKeep != nil {
		applyImageDepthToKeepPolicy(refs, *policy.ImageDepthToKeep)
	}

	return refs
}

func applyRefsToKeepImagesInPolicy(policyTagsRefs []*referenceToScan, refsToKeepImagesIn *config.RefsToKeepImagesIn) []*referenceToScan {
	var policyModifiedInRefs []*referenceToScan
	if refsToKeepImagesIn.ModifiedIn != nil {
		policyModifiedInRefs = filterReferencesByModifiedIn(policyTagsRefs, *refsToKeepImagesIn.ModifiedIn)
	}

	var policyLastRefs []*referenceToScan
	if refsToKeepImagesIn.Last != nil {
		policyLastRefs = filterReferencesByLast(policyTagsRefs, *refsToKeepImagesIn.Last)
	}

	var policyRefs []*referenceToScan
	if refsToKeepImagesIn.Operator == config.AndOperator {
		policyRefs = referencesAnd(policyModifiedInRefs, policyLastRefs)
	} else {
		policyRefs = referencesOr(policyModifiedInRefs, policyLastRefs)
	}

	return policyRefs
}

func applyImageDepthToKeepPolicy(policyBranchesRefs []*referenceToScan, imageDepthToKeep int) {
	for _, ref := range policyBranchesRefs {
		if ref.imageDepthToKeep < imageDepthToKeep {
			ref.imageDepthToKeep = imageDepthToKeep
		}
	}
}

func referencesOr(refs1 []*referenceToScan, refs2 []*referenceToScan) []*referenceToScan {
	return mergeReferences(refs1, refs2)
}

func referencesAnd(refs1 []*referenceToScan, refs2 []*referenceToScan) []*referenceToScan {
	var result []*referenceToScan

outerLoop:
	for _, ref1 := range refs1 {
		for _, ref2 := range refs2 {
			if ref1 == ref2 {
				result = append(result, ref1)
				continue outerLoop
			}
		}
	}

	return result
}

func filterReferencesByModifiedIn(refs []*referenceToScan, modifiedIn time.Duration) (result []*referenceToScan) {
	for _, ref := range refs {
		if ref.ModifiedAt.Before(time.Now().Add(-modifiedIn)) {
			continue
		}

		result = append(result, ref)
	}

	return
}

func filterReferencesByLast(refs []*referenceToScan, last int) []*referenceToScan {
	sort.Slice(refs, func(i, j int) bool {
		return refs[i].ModifiedAt.After(refs[j].ModifiedAt)
	})

	if len(refs) < last {
		return refs
	}

	return refs[:last]
}

func mergeReferences(refs1 []*referenceToScan, refs2 []*referenceToScan) []*referenceToScan {
	result := refs2[:]

outerLoop:
	for _, ref1 := range refs1 {
		for _, ref2 := range refs2 {
			if ref1 == ref2 {
				continue outerLoop
			}
		}

		result = append(result, ref1)
	}

	return result
}
