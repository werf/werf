package instruction_test

import (
	"bytes"
	"fmt"

	"github.com/moby/buildkit/frontend/dockerfile/instructions"
	"github.com/moby/buildkit/frontend/dockerfile/parser"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/werf/werf/v2/pkg/build/stage"
	"github.com/werf/werf/v2/pkg/build/stage/instruction"
)

func parseRunCommandAI(dockerfileText string) *instructions.RunCommand {
	p, err := parser.Parse(bytes.NewReader([]byte(dockerfileText)))
	Expect(err).To(Succeed())

	dockerStages, _, err := instructions.Parse(p.AST)
	Expect(err).To(Succeed())
	Expect(dockerStages).NotTo(BeEmpty())

	for _, cmd := range dockerStages[len(dockerStages)-1].Commands {
		if run, ok := cmd.(*instructions.RunCommand); ok {
			return run
		}
	}

	Fail("no RUN command found in dockerfile")
	return nil
}

func newRunStageAI(runCommand *instructions.RunCommand, dependencyStages []string) *instruction.Run {
	return instruction.NewRun(
		NewDockerfileStageInstructionWithDependencyStages(runCommand, dependencyStages),
		nil, false,
		&stage.BaseStageOptions{ImageName: "example-image", ProjectName: "example-project"},
		nil, "",
	)
}

var _ = Describe("TestAI_ RUN mount from stage resolution", func() {
	const resolvedOsImage = "ghcr.io/werf/instruction-test:a71052baf9c6ace8171e59a2ae5ea1aede3fb89aa95d160ec354b205-1661868399091"

	It("TestAI_ resolves --mount from=<stage> to the built werf stage image in the backend instruction", func(ctx SpecContext) {
		stg := newRunStageAI(
			parseRunCommandAI("FROM alpine AS os\nRUN --mount=type=bind,from=os,source=/apk,target=/apk true\n"),
			[]string{"os"},
		)

		conveyor := stage.NewConveyorStub(
			stage.NewGiterminismManagerStub(stage.NewLocalGitRepoStub("test"), stage.NewGiterminismInspectorStub()),
			map[string]string{"/stage/os": resolvedOsImage},
			nil, nil,
		)

		Expect(stg.ExpandDependencies(ctx, conveyor, map[string]string{})).To(Succeed())

		mounts := instruction.ExportRunMounts(stg)
		Expect(mounts).To(HaveLen(1))
		Expect(mounts[0].From).To(Equal(resolvedOsImage))
	})

	It("TestAI_ leaves external --mount from=<image> references unchanged", func(ctx SpecContext) {
		stg := newRunStageAI(
			parseRunCommandAI("FROM alpine\nRUN --mount=type=bind,from=alpine:3.19,source=/etc,target=/etc true\n"),
			nil,
		)

		conveyor := stage.NewConveyorStub(
			stage.NewGiterminismManagerStub(stage.NewLocalGitRepoStub("test"), stage.NewGiterminismInspectorStub()),
			nil, nil, nil,
		)

		Expect(stg.ExpandDependencies(ctx, conveyor, map[string]string{})).To(Succeed())

		mounts := instruction.ExportRunMounts(stg)
		Expect(mounts).To(HaveLen(1))
		Expect(mounts[0].From).To(Equal("alpine:3.19"))
	})

	digestFor := func(ctx SpecContext, resolvedImage string) string {
		stg := newRunStageAI(
			parseRunCommandAI("FROM alpine AS os\nRUN --mount=type=bind,from=os,source=/apk,target=/apk true\n"),
			[]string{"os"},
		)

		conveyor := stage.NewConveyorStub(
			stage.NewGiterminismManagerStub(stage.NewLocalGitRepoStub("test"), stage.NewGiterminismInspectorStub()),
			map[string]string{"/stage/os": resolvedImage},
			nil, nil,
		)

		Expect(stg.ExpandDependencies(ctx, conveyor, map[string]string{})).To(Succeed())

		digest, err := stg.GetDependencies(ctx, conveyor, stage.NewContainerBackendStub(), nil, nil, nil)
		Expect(err).To(Succeed())
		return digest
	}

	It("TestAI_ digest reflects the resolved stage image and is stable for identical inputs", func(ctx SpecContext) {
		digest1 := digestFor(ctx, resolvedOsImage)
		digest2 := digestFor(ctx, resolvedOsImage)
		Expect(digest1).To(Equal(digest2))

		changedDigest := digestFor(ctx, "ghcr.io/werf/instruction-test:4930d562bfbee9c931413c826137d49eff6a2e7d39519c1c9488a747-1655913653892")
		fmt.Printf("digest: %s, changedDigest: %s\n", digest1, changedDigest)
		Expect(changedDigest).NotTo(Equal(digest1))
	})
})
