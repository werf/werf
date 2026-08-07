package stage

import (
	"context"
	"testing"

	"github.com/onsi/gomega"
	"github.com/stretchr/testify/require"

	"github.com/werf/werf/v2/pkg/config"
)

func TestAI_GetContentDependencies_NoPanicOnNilDelegation(t *testing.T) {
	gomega.RegisterTestingT(t)
	ctx := context.Background()
	conveyor := NewConveyorStub(
		NewGiterminismManagerStub(NewLocalGitRepoStub("9d8059842b6fde712c58315ca0ab4713d90761c0"), NewGiterminismInspectorStub()),
		map[string]string{},
		map[string]string{},
	)
	opts := &BaseStageOptions{TargetPlatform: "linux/amd64", ImageName: "example-image", ProjectName: "example-project"}

	dockerfile := []byte("FROM base\nRUN true\n")
	dockerStages, dockerMetaArgs := testDockerfileToDockerStages(dockerfile)

	stages := map[string]Interface{
		"dependencies":    &DependenciesStage{BaseStage: NewBaseStage(DependenciesBeforeInstall, opts)},
		"from":            &FromStage{BaseStage: NewBaseStage(From, opts)},
		"image_spec":      newImageSpecStage(&config.ImageSpec{}, opts),
		"full_dockerfile": newTestFullDockerfileStage(dockerfile, "", nil, dockerStages, dockerMetaArgs, nil, ""),
	}

	for name, s := range stages {
		t.Run(name, func(t *testing.T) {
			require.NotPanics(t, func() {
				_, _ = s.GetContentDependencies(ctx, conveyor, nil)
			})
		})
	}
}
