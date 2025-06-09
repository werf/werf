package stage_manager

import (
	"sync"

	"github.com/werf/werf/v2/pkg/image"
)

type managedStageDescSet struct {
	stageDescSet       image.StageDescSet
	stageDescByStageID map[string]*image.StageDesc // To avoid iterating over the stageDescSet accessing the stageDesc by ID.
	stageDescMetaMap   map[*image.StageDesc]*stageMeta
	mu                 sync.Mutex
}

type stageMeta struct {
	isProtected      bool
	protectionReason *protectionReason
}

type protectionReason struct {
	description string
}

func (r *protectionReason) String() string {
	return r.description
}

func newProtectionReason(desc string) *protectionReason {
	return &protectionReason{description: desc}
}

func (r *protectionReason) SetDescription(desc string) {
	r.description = desc
}

var (
	ProtectionReasonKubernetesBasedPolicy       = newProtectionReason("used in Kubernetes")
	ProtectionReasonGitPolicy                   = newProtectionReason("git policy")
	ProtectionReasonBuiltWithinLastNHoursPolicy = newProtectionReason("built within last N hours")
	ProtectionReasonImportSource                = newProtectionReason("import source")
	ProtectionReasonDependencySource            = newProtectionReason("dependency source")
	ProtectionReasonAncestor                    = newProtectionReason("ancestor")
	ProtectionReasonImageIndexPlatform          = newProtectionReason("image index platform")
	ProtectionReasonNotFoundInRepo              = newProtectionReason("not found in repo")
	ProtectionReasonKeepList                    = newProtectionReason("keep list")
)

func newManagedStageDescSet(set image.StageDescSet) *managedStageDescSet {
	s := &managedStageDescSet{
		stageDescSet:       set,
		stageDescMetaMap:   map[*image.StageDesc]*stageMeta{},
		stageDescByStageID: map[string]*image.StageDesc{},
	}

	for stageDesc := range set.Iter() {
		s.stageDescByStageID[stageDesc.StageID.String()] = stageDesc
	}

	return s
}

func (s *managedStageDescSet) StageDescSet() image.StageDescSet {
	return s.stageDescSet.Clone()
}

func (s *managedStageDescSet) DifferenceInPlace(stageDescSet image.StageDescSet) {
	s.stageDescSet = s.stageDescSet.Difference(stageDescSet)
	for stageDesc := range stageDescSet.Iter() {
		delete(s.stageDescByStageID, stageDesc.StageID.String())
	}
}

func (s *managedStageDescSet) MarkStageDescAsProtected(stageDesc *image.StageDesc, reason *protectionReason, forceReason bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, ok := s.stageDescMetaMap[stageDesc]
	if !ok {
		s.stageDescMetaMap[stageDesc] = &stageMeta{}
	}

	// If the stage is already protected, do not change the protection reason.
	if s.stageDescMetaMap[stageDesc].isProtected && !forceReason {
		return
	}

	s.stageDescMetaMap[stageDesc].isProtected = true
	s.stageDescMetaMap[stageDesc].protectionReason = reason
}

func (s *managedStageDescSet) GetProtectedStageDescSet() image.StageDescSet {
	stageDescSet := image.NewStageDescSet()
	for _, set := range s.GetProtectedStageDescSetByReason() {
		stageDescSet = stageDescSet.Union(set)
	}

	return stageDescSet
}

func (s *managedStageDescSet) GetProtectedStageDescSetByReason() map[*protectionReason]image.StageDescSet {
	stageDescSetByReason := make(map[*protectionReason]image.StageDescSet)
	for stageDesc, meta := range s.stageDescMetaMap {
		if !meta.isProtected {
			continue
		}

		_, ok := stageDescSetByReason[meta.protectionReason]
		if !ok {
			stageDescSetByReason[meta.protectionReason] = image.NewStageDescSet()
		}

		stageDescSetByReason[meta.protectionReason].Add(stageDesc)
	}

	return stageDescSetByReason
}

func (s *managedStageDescSet) GetStageDescByStageID(stageID string) *image.StageDesc {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.stageDescByStageID[stageID]
}
