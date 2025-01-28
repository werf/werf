package stage_manager

import (
	"github.com/werf/werf/v2/pkg/image"
)

type managedStageDescSet struct {
	stageDescSet     image.StageDescSet
	stageDescMetaMap map[*image.StageDesc]*stageMeta
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
	ProtectionReasonAncestor                    = newProtectionReason("ancestor")
	ProtectionReasonNotFoundInRepo              = newProtectionReason("not found in repo")
)

func newManagedStageDescSet(set image.StageDescSet) managedStageDescSet {
	return managedStageDescSet{
		stageDescSet:     set,
		stageDescMetaMap: map[*image.StageDesc]*stageMeta{},
	}
}

func (s *managedStageDescSet) StageDescSet() image.StageDescSet {
	return s.stageDescSet.Clone()
}

func (s *managedStageDescSet) DifferenceInPlace(stageDescSet image.StageDescSet) {
	s.stageDescSet = s.stageDescSet.Difference(stageDescSet)
}

func (s *managedStageDescSet) MarkStageDescAsProtected(stageDesc *image.StageDesc, reason *protectionReason, forceReason bool) {
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
	for stageDesc := range s.stageDescSet.Iter() {
		if stageDesc.StageID.String() == stageID {
			return stageDesc
		}
	}

	return nil
}
