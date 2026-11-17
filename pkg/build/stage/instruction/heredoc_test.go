package instruction_test

import (
	"github.com/moby/buildkit/frontend/dockerfile/instructions"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/werf/werf/v3/pkg/build/stage"
	"github.com/werf/werf/v3/pkg/build/stage/instruction"
	"github.com/werf/werf/v3/pkg/dockerfile"
	"github.com/werf/werf/v3/pkg/dockerfile/frontend"
)

func heredocSourcesAndDest(data string) instructions.SourcesAndDest {
	return instructions.SourcesAndDest{
		DestPath:       "/install.sh",
		SourceContents: []instructions.SourceContent{{Path: "EOF", Data: data}},
	}
}

func newHeredocCopy(data string) stage.Interface {
	return instruction.NewCopy(
		dockerfile.NewDockerfileStageInstruction(
			&instructions.CopyCommand{SourcesAndDest: heredocSourcesAndDest(data)},
			dockerfile.DockerfileStageInstructionOptions{ExpanderFactory: frontend.NewShlexExpanderFactory('\\')},
		),
		nil, false,
		&stage.BaseStageOptions{ImageName: "test", ProjectName: "test"},
	)
}

func newHeredocAdd(data string) stage.Interface {
	return instruction.NewAdd(
		dockerfile.NewDockerfileStageInstruction(
			&instructions.AddCommand{SourcesAndDest: heredocSourcesAndDest(data)},
			dockerfile.DockerfileStageInstructionOptions{ExpanderFactory: frontend.NewShlexExpanderFactory('\\')},
		),
		nil, false,
		&stage.BaseStageOptions{ImageName: "test", ProjectName: "test"},
	)
}

var _ = DescribeTable("heredoc-only instruction digest",
	func(ctx SpecContext, newStage func(data string) stage.Interface) {
		first := NewTestData(newStage("echo first\n"), "", TestDataOptions{})
		second := NewTestData(newStage("echo second\n"), "", TestDataOptions{})

		firstDigest, err := first.Stage.GetDependencies(ctx, first.Conveyor, first.ContainerBackend, nil, first.StageImage, first.BuildContext)
		Expect(err).To(Succeed())
		secondDigest, err := second.Stage.GetDependencies(ctx, second.Conveyor, second.ContainerBackend, nil, second.StageImage, second.BuildContext)
		Expect(err).To(Succeed())
		Expect(secondDigest).NotTo(Equal(firstDigest))

		firstContentDigest, err := first.Stage.GetContentDependencies(ctx, first.Conveyor, first.BuildContext)
		Expect(err).To(Succeed())
		secondContentDigest, err := second.Stage.GetContentDependencies(ctx, second.Conveyor, second.BuildContext)
		Expect(err).To(Succeed())
		Expect(secondContentDigest).NotTo(Equal(firstContentDigest))
	},
	Entry("COPY", newHeredocCopy),
	Entry("ADD", newHeredocAdd),
)

var _ = Describe("heredoc expansion", func() {
	It("expands variables only in heredocs with unquoted delimiters and keeps quotes", func(ctx SpecContext) {
		sourceContents := []instructions.SourceContent{
			{Path: "EXPANDED", Data: "${TEST_VAR} \"quoted\" 'single' \\$KEEP\n", Expand: true},
			{Path: "RAW", Data: "${TEST_VAR}\n"},
		}
		stg := instruction.NewCopy(
			dockerfile.NewDockerfileStageInstruction(
				&instructions.CopyCommand{SourcesAndDest: instructions.SourcesAndDest{DestPath: "/dst/", SourceContents: sourceContents}},
				dockerfile.DockerfileStageInstructionOptions{ExpanderFactory: frontend.NewShlexExpanderFactory('\\')},
			),
			nil, false,
			&stage.BaseStageOptions{ImageName: "test", ProjectName: "test"},
		)
		conveyor := stage.NewConveyorStub(
			stage.NewGiterminismManagerStub(stage.NewLocalGitRepoStub("test"), stage.NewGiterminismInspectorStub()),
			nil, nil,
		)

		Expect(stg.ExpandDependencies(ctx, conveyor, map[string]string{"TEST_VAR": "resolved"})).To(Succeed())

		backend, source := instruction.ExportCopyCommands(stg)
		Expect(backend.SourceContents).To(Equal([]instructions.SourceContent{
			{Path: "EXPANDED", Data: "resolved \"quoted\" 'single' $KEEP\n", Expand: true},
			{Path: "RAW", Data: "${TEST_VAR}\n"},
		}))
		Expect(source.SourceContents).To(Equal([]instructions.SourceContent{
			{Path: "EXPANDED", Data: "${TEST_VAR} \"quoted\" 'single' \\$KEEP\n", Expand: true},
			{Path: "RAW", Data: "${TEST_VAR}\n"},
		}))
	})

	It("escapes with a backslash regardless of the Dockerfile escape token", func(ctx SpecContext) {
		stg := instruction.NewCopy(
			dockerfile.NewDockerfileStageInstruction(
				&instructions.CopyCommand{SourcesAndDest: instructions.SourcesAndDest{
					DestPath:       "/dst",
					SourceContents: []instructions.SourceContent{{Path: "EOF", Data: "\\$KEEP `${TEST_VAR}`\n", Expand: true}},
				}},
				dockerfile.DockerfileStageInstructionOptions{ExpanderFactory: frontend.NewShlexExpanderFactory('`')},
			),
			nil, false,
			&stage.BaseStageOptions{ImageName: "test", ProjectName: "test"},
		)
		data := NewTestData(stg, "", TestDataOptions{})

		Expect(stg.ExpandDependencies(ctx, data.Conveyor, map[string]string{"TEST_VAR": "resolved"})).To(Succeed())

		backend, _ := instruction.ExportCopyCommands(stg)
		Expect(backend.SourceContents).To(Equal([]instructions.SourceContent{{Path: "EOF", Data: "$KEEP `resolved`\n", Expand: true}}))
	})

	DescribeTable("digest follows the expanded heredoc content",
		func(ctx SpecContext, expand, expectDifferentDigests bool) {
			digest := func(value string) string {
				stg := instruction.NewCopy(
					dockerfile.NewDockerfileStageInstruction(
						&instructions.CopyCommand{SourcesAndDest: instructions.SourcesAndDest{
							DestPath:       "/dst",
							SourceContents: []instructions.SourceContent{{Path: "EOF", Data: "${TEST_VAR}\n", Expand: expand}},
						}},
						dockerfile.DockerfileStageInstructionOptions{ExpanderFactory: frontend.NewShlexExpanderFactory('\\')},
					),
					nil, false,
					&stage.BaseStageOptions{ImageName: "test", ProjectName: "test"},
				)
				data := NewTestData(stg, "", TestDataOptions{})
				Expect(stg.ExpandDependencies(ctx, data.Conveyor, map[string]string{"TEST_VAR": value})).To(Succeed())
				result, err := stg.GetDependencies(ctx, data.Conveyor, data.ContainerBackend, nil, data.StageImage, data.BuildContext)
				Expect(err).To(Succeed())
				return result
			}

			if expectDifferentDigests {
				Expect(digest("first")).NotTo(Equal(digest("second")))
			} else {
				Expect(digest("first")).To(Equal(digest("second")))
			}
		},
		Entry("unquoted delimiter", true, true),
		Entry("quoted delimiter", false, false),
	)
})
