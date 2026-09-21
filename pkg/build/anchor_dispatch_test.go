package build

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	buildImage "github.com/werf/werf/v2/pkg/build/image"
	"github.com/werf/werf/v2/pkg/build/stage"
	"github.com/werf/werf/v2/pkg/config"
	imagePkg "github.com/werf/werf/v2/pkg/image"
)

type anchorStub struct {
	*stage.BaseStage
	selectFn func(context.Context, stage.Conveyor, imagePkg.StageDescSet) (*imagePkg.StageDesc, error)
}

func (s *anchorStub) SelectSuitableStageDesc(ctx context.Context, c stage.Conveyor, set imagePkg.StageDescSet) (*imagePkg.StageDesc, error) {
	if s.selectFn != nil {
		return s.selectFn(ctx, c, set)
	}
	return s.BaseStage.SelectSuitableStageDesc(ctx, c, set)
}

func TestAnchorSelector_HonoursParentTsToggle(t *testing.T) {
	phase := &BuildPhase{StagesIterator: NewStagesIterator(nil)}

	anchor := &anchorStub{BaseStage: stage.NewBaseStage(stage.ImageSpec, &stage.BaseStageOptions{})}
	anchor.SetContentAnchor(true)
	require.Equal(t, int64(0), phase.getPrevNonEmptyStageCreationTsForStage(anchor),
		"anchor stage must resolve with parentStageCreationTs=0")

	nonAnchor := &anchorStub{BaseStage: stage.NewBaseStage(stage.Setup, &stage.BaseStageOptions{})}
	require.Equal(t, int64(0), phase.getPrevNonEmptyStageCreationTsForStage(nonAnchor),
		"non-anchor branch resolves through the iterator (nil PrevNonEmptyStage → 0)")
}

func TestCollectHolisticInputs_FoldsInDependencyAnchorDigests(t *testing.T) {
	img, err := buildImage.NewImage(context.Background(), "linux/amd64", "app", buildImage.NoBaseImage, buildImage.ImageOptions{IsFinal: true})
	require.NoError(t, err)

	dep, err := buildImage.NewImage(context.Background(), "linux/amd64", "base", buildImage.NoBaseImage, buildImage.ImageOptions{})
	require.NoError(t, err)

	dep.SetAnchorDigest("anchor-digest-v1")
	before, err := collectHolisticInputs(context.Background(), img, []*buildImage.Image{dep}, nil, nil)
	require.NoError(t, err)

	dep.SetAnchorDigest("anchor-digest-v2")
	after, err := collectHolisticInputs(context.Background(), img, []*buildImage.Image{dep}, nil, nil)
	require.NoError(t, err)

	require.NotEqual(t, before, after, "content of a dependency image must reach the anchor digest of its dependent")

	dep.SetAnchorDigest("")
	_, err = collectHolisticInputs(context.Background(), img, []*buildImage.Image{dep}, nil, nil)
	require.ErrorContains(t, err, "no content-based digest",
		"a dependency without a digest would silently drop out of the anchor")
}

func TestCollectHolisticInputs_FoldsInDockerfileDependencyArgs(t *testing.T) {
	newDockerfileImage := func(importType config.DependencyImportType) *buildImage.Image {
		img, err := buildImage.NewImage(context.Background(), "linux/amd64", "app", buildImage.NoBaseImage, buildImage.ImageOptions{
			IsFinal:           true,
			IsDockerfileImage: true,
			DockerfileImageConfig: &config.ImageFromDockerfile{
				Staged: true,
				Dependencies: []*config.Dependency{
					{From: "base", Imports: []*config.DependencyImport{{Type: importType, TargetBuildArg: "BASE_IMAGE"}}},
				},
			},
		})
		require.NoError(t, err)
		return img
	}

	dep, err := buildImage.NewImage(context.Background(), "linux/amd64", "base", buildImage.NoBaseImage, buildImage.ImageOptions{})
	require.NoError(t, err)
	dep.SetAnchorDigest("anchor-digest")

	byImageName, err := collectHolisticInputs(context.Background(), newDockerfileImage(config.ImageNameImport), []*buildImage.Image{dep}, nil, nil)
	require.NoError(t, err)

	byImageRepo, err := collectHolisticInputs(context.Background(), newDockerfileImage(config.ImageRepoImport), []*buildImage.Image{dep}, nil, nil)
	require.NoError(t, err)

	require.NotEqual(t, byImageName, byImageRepo,
		"the import type decides what FROM ${BASE_IMAGE} resolves to, so it must reach the anchor digest")
}
