package cleaning

import (
	"fmt"
	"sort"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"

	"github.com/werf/logboek"
)

func scanReferencesHistory(gitRepository *git.Repository, refs []*referenceToScan, expectedContentSignatureCommitHashes map[string][]plumbing.Hash) ([]string, error) {
	var reachedContentSignatureList []string
	var stopCommitHashes []plumbing.Hash

	for i := len(refs) - 1; i >= 0; i-- {
		ref := refs[i]

		logProcessMessage := fmt.Sprintf("Reference %s (scanDepthLimit: %s, imageDepthToKeep: %d)", ref.Name().Short(), ref.referenceScanOptions.scanDepthLimitLogString(), ref.referenceScanOptions.imageDepthToKeep)
		var refReachedContentSignatureList []string
		var refStopCommitHashes []plumbing.Hash
		var err error

		if err := logboek.Info.LogProcess(logProcessMessage, logboek.LevelLogProcessOptions{}, func() error {
			refReachedContentSignatureList, refStopCommitHashes, err = scanReferenceHistory(gitRepository, ref, expectedContentSignatureCommitHashes, stopCommitHashes)
			if err != nil {
				return fmt.Errorf("scan reference %s history failed: %s", ref.Name().Short(), err)
			}

			stopCommitHashes = append(stopCommitHashes, refStopCommitHashes...)

		outerLoop:
			for _, c1 := range refReachedContentSignatureList {
				for _, c2 := range reachedContentSignatureList {
					if c1 == c2 {
						continue outerLoop
					}
				}

				reachedContentSignatureList = append(reachedContentSignatureList, c1)
			}

			return nil
		}); err != nil {
			return nil, err
		}

		if len(reachedContentSignatureList) == len(expectedContentSignatureCommitHashes) {
			logboek.Info.LogLn("Scanning stopped due to all expected commit hashes were reached")
			break
		}
	}

	return reachedContentSignatureList, nil
}

type commitHistoryScanner struct {
	gitRepository                        *git.Repository
	expectedContentSignatureCommitHashes map[string][]plumbing.Hash
	reachedContentSignatureCommitHashes  map[string][]plumbing.Hash
	reachedCommitHashes                  []plumbing.Hash
	stopCommitHashes                     []plumbing.Hash
	isAlreadyScannedCommit               map[plumbing.Hash]bool
	scanDepth                            int

	referenceScanOptions
}

func (s *commitHistoryScanner) reachedContentSignatureList() []string {
	var reachedContentSignatureList []string
	for contentSignarure, commitHashes := range s.reachedContentSignatureCommitHashes {
		if len(commitHashes) != 0 {
			reachedContentSignatureList = append(reachedContentSignatureList, contentSignarure)
		}
	}

	return reachedContentSignatureList
}

func scanReferenceHistory(gitRepository *git.Repository, ref *referenceToScan, expectedContentSignatureCommitHashes map[string][]plumbing.Hash, stopCommitHashes []plumbing.Hash) ([]string, []plumbing.Hash, error) {
	s := &commitHistoryScanner{
		gitRepository:                        gitRepository,
		expectedContentSignatureCommitHashes: expectedContentSignatureCommitHashes,
		reachedContentSignatureCommitHashes:  map[string][]plumbing.Hash{},
		stopCommitHashes:                     stopCommitHashes,

		referenceScanOptions:   ref.referenceScanOptions,
		isAlreadyScannedCommit: map[plumbing.Hash]bool{},
	}

	if err := s.scanCommitHistory(ref.HeadCommit.Hash); err != nil {
		return nil, nil, fmt.Errorf("scan commit %s history failed: %s", ref.HeadCommit.Hash.String(), err)
	}

	if s.referenceScanOptions.imageDepthToKeep != 0 {
		if len(s.reachedContentSignatureList()) == s.referenceScanOptions.imageDepthToKeep {
			return s.reachedContentSignatureList(), s.stopCommitHashes, nil
		} else if len(s.reachedContentSignatureList()) > s.referenceScanOptions.imageDepthToKeep {
			logboek.Info.LogLn("Reached more commit hashes than expected %d/%d", len(s.reachedContentSignatureList()), s.referenceScanOptions.imageDepthToKeep)

			latestCommitContentSignature := s.latestCommitContentSignature()
			var latestCommitList []*object.Commit
			for latestCommit := range latestCommitContentSignature {
				latestCommitList = append(latestCommitList, latestCommit)
			}

			sort.Slice(latestCommitList, func(i, j int) bool {
				return latestCommitList[i].Committer.When.After(latestCommitList[j].Committer.When)
			})

			var reachedContentSignatureList []string
			var skippedContentSignatureList []string
			for ind, latestCommit := range latestCommitList {
				contentSignature := latestCommitContentSignature[latestCommit]
				if ind < s.referenceScanOptions.imageDepthToKeep {
					reachedContentSignatureList = append(reachedContentSignatureList, contentSignature)
				} else {
					skippedContentSignatureList = append(skippedContentSignatureList, contentSignature)
				}
			}

			_ = logboek.Info.LogBlock("Skipped oldest content signatures by associated commits", logboek.LevelLogBlockOptions{}, func() error {
				for _, contentSignature := range skippedContentSignatureList {
					logboek.Info.LogLn(contentSignature)
				}
				return nil
			})

			return reachedContentSignatureList, s.stopCommitHashes, nil
		}
	}

	if s.referenceScanOptions.scanDepthLimit != 0 {
		if len(s.reachedContentSignatureList()) == len(s.expectedContentSignatureCommitHashes) {
			s.stopCommitHashes = append(s.stopCommitHashes, s.reachedCommitHashes[len(s.reachedCommitHashes)-1])
		} else {
			return s.reachedContentSignatureList(), s.stopCommitHashes, nil
		}
	} else if len(s.reachedContentSignatureList()) != 0 {
		s.stopCommitHashes = append(s.stopCommitHashes, s.reachedCommitHashes[len(s.reachedCommitHashes)-1])
	} else {
		s.stopCommitHashes = append(s.stopCommitHashes, ref.HeadCommit.Hash)
	}
	logboek.Debug.LogF("Stop commit %s added\n", s.stopCommitHashes[len(s.stopCommitHashes)-1].String())

	return s.reachedContentSignatureList(), s.stopCommitHashes, nil
}

func (s *commitHistoryScanner) scanCommitHistory(commitHash plumbing.Hash) error {
	var currentIteration, nextIteration []plumbing.Hash

	currentIteration = append(currentIteration, commitHash)
	for {
		s.scanDepth++

		for _, commitHash := range currentIteration {
			if s.isAlreadyScannedCommit[commitHash] {
				continue
			}

			if s.isStopCommitHash(commitHash) {
				logboek.Debug.LogF("Stop scanning commit history %s due to stop commit reached\n", commitHash.String())
				continue
			}

			commitParents, err := s.handleCommitHash(commitHash)
			if err != nil {
				return err
			}

			if s.scanDepth == s.referenceScanOptions.scanDepthLimit {
				logboek.Debug.LogF("Stop scanning commit history %s due to scanDepthLimit (%d)\n", commitHash.String(), s.referenceScanOptions.scanDepthLimit)
				continue
			}

			if len(s.expectedContentSignatureCommitHashes) == len(s.reachedContentSignatureCommitHashes) {
				logboek.Debug.LogF("Stop scanning due to all expected content signatures reached\n", commitHash.String())
				break
			}

			s.isAlreadyScannedCommit[commitHash] = true
			nextIteration = append(nextIteration, commitParents...)
		}

		if len(nextIteration) == 0 {
			break
		}

		currentIteration = nextIteration
		nextIteration = []plumbing.Hash{}
	}

	return nil
}

func (s *commitHistoryScanner) handleCommitHash(commitHash plumbing.Hash) ([]plumbing.Hash, error) {
outerLoop:
	for contentSignature, commitHashes := range s.expectedContentSignatureCommitHashes {
		for _, c := range commitHashes {
			if c == commitHash {
				for _, reachedC := range s.reachedCommitHashes {
					if reachedC == c {
						break outerLoop
					}
				}
				s.reachedCommitHashes = append(s.reachedCommitHashes, c)

				reachedCommitHashes, ok := s.reachedContentSignatureCommitHashes[contentSignature]
				if !ok {
					reachedCommitHashes = []plumbing.Hash{}
				}
				reachedCommitHashes = append(reachedCommitHashes, c)
				s.reachedContentSignatureCommitHashes[contentSignature] = reachedCommitHashes

				if !ok {
					logboek.Info.LogF(
						"Expected content signature %s was reached on commit %s (scanDepthLimit: %d/%s, imageDepthToKeep: %d/%d)\n",
						contentSignature,
						commitHash.String(),
						s.scanDepth,
						s.referenceScanOptions.scanDepthLimitLogString(),
						len(s.reachedContentSignatureCommitHashes),
						len(s.expectedContentSignatureCommitHashes),
					)
				} else {
					logboek.Info.LogF(
						"Expected content signature %s was reached again on another commit %s (scanDepthLimit: %d/%s, imageDepthToKeep: %d/%d)\n",
						contentSignature,
						commitHash.String(),
						s.scanDepth,
						s.referenceScanOptions.scanDepthLimitLogString(),
						len(s.reachedContentSignatureCommitHashes),
						len(s.expectedContentSignatureCommitHashes),
					)
				}

				break outerLoop
			}
		}
	}

	co, err := s.gitRepository.CommitObject(commitHash)
	if err != nil {
		return nil, err
	}

	return co.ParentHashes, nil
}

func (s *commitHistoryScanner) isStopCommitHash(commitHash plumbing.Hash) bool {
	for _, c := range s.stopCommitHashes {
		if commitHash == c {
			return true
		}
	}

	return false
}

func (s *commitHistoryScanner) latestCommitContentSignature() map[*object.Commit]string {
	contentSignatureLatestCommit := map[*object.Commit]string{}
	for contentSignature, commitHashes := range s.reachedContentSignatureCommitHashes {
		var latestCommit *object.Commit
		for _, commitHash := range commitHashes {
			commit, err := s.gitRepository.CommitObject(commitHash)
			if err != nil {
				panic("unexpected condition")
			}

			if latestCommit == nil || commit.Committer.When.After(latestCommit.Committer.When) {
				latestCommit = commit
			}
		}

		if latestCommit != nil {
			contentSignatureLatestCommit[latestCommit] = contentSignature
		}
	}

	return contentSignatureLatestCommit
}
