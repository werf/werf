package stage_manager

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/werf/werf/v2/pkg/image"
)

func newTestStageDesc(digest string, creationTs int64) *image.StageDesc {
	stageID := image.NewStageID(digest, creationTs)
	return &image.StageDesc{StageID: stageID, Info: &image.Info{Tag: stageID.String()}}
}

func TestGetFinalProtectedStageDescSetByReason(t *testing.T) {
	manager := NewManager()

	finalStageDesc := newTestStageDesc("finaldigest", 1749456960043)
	manager.finalManagedStageDescSet = newManagedStageDescSet(image.NewStageDescSet(finalStageDesc))

	nonFinalStageDesc := newTestStageDesc("nonfinaldigest", 1749390012345)
	manager.managedStageDescSet = newManagedStageDescSet(image.NewStageDescSet(nonFinalStageDesc))

	manager.MarkFinalStageDescAsProtected(finalStageDesc, ProtectionReasonKubernetesBasedPolicy, false)
	manager.MarkStageDescAsProtected(nonFinalStageDesc, ProtectionReasonGitPolicy, false)

	finalByReason := manager.GetFinalProtectedStageDescSetByReason()

	require.Len(t, finalByReason, 1)
	require.Contains(t, finalByReason, ProtectionReasonKubernetesBasedPolicy)
	assert.Equal(t, []*image.StageDesc{finalStageDesc}, finalByReason[ProtectionReasonKubernetesBasedPolicy].ToSlice())
	assert.NotContains(t, finalByReason, ProtectionReasonGitPolicy)

	nonFinalByReason := manager.GetProtectedStageDescSetByReason()

	require.Len(t, nonFinalByReason, 1)
	require.Contains(t, nonFinalByReason, ProtectionReasonGitPolicy)
	assert.Equal(t, []*image.StageDesc{nonFinalStageDesc}, nonFinalByReason[ProtectionReasonGitPolicy].ToSlice())
}

func TestGetFinalProtectedStageDescSetByReasonEmptyWithoutProtection(t *testing.T) {
	manager := NewManager()
	manager.finalManagedStageDescSet = newManagedStageDescSet(image.NewStageDescSet(newTestStageDesc("finaldigest", 1749456960043)))

	assert.Empty(t, manager.GetFinalProtectedStageDescSetByReason())
}
