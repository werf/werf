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
	protectionReason string
}

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

func (s *managedStageDescSet) MarkStageDescAsProtected(stageDesc *image.StageDesc, protectionReason string) {
	_, ok := s.stageDescMetaMap[stageDesc]
	if !ok {
		s.stageDescMetaMap[stageDesc] = &stageMeta{}
	}

	// If the stage is already protected, do not change the protection reason.
	if s.stageDescMetaMap[stageDesc].isProtected {
		return
	}

	s.stageDescMetaMap[stageDesc].isProtected = true
	s.stageDescMetaMap[stageDesc].protectionReason = protectionReason
}

func (s *managedStageDescSet) GetProtectedStageDescSet() image.StageDescSet {
	stageDescSet := image.NewStageDescSet()
	for _, set := range s.GetProtectedStageDescSetByReason() {
		stageDescSet = stageDescSet.Union(set)
	}

	return stageDescSet
}

func (s *managedStageDescSet) GetProtectedStageDescSetByReason() map[string]image.StageDescSet {
	stageDescSetByReason := make(map[string]image.StageDescSet)
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
