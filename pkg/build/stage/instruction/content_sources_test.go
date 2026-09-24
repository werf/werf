package instruction_test

import (
	"github.com/moby/buildkit/frontend/dockerfile/instructions"
	"github.com/moby/buildkit/frontend/dockerfile/parser"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/werf/werf/v3/pkg/build/stage"
	"github.com/werf/werf/v3/pkg/build/stage/instruction"
	"github.com/werf/werf/v3/pkg/dockerfile"
	"github.com/werf/werf/v3/pkg/dockerfile/frontend"
)

var _ = DescribeTable("content source paths with inherited environment variables",
	func(ctx SpecContext, newStage func() stage.Interface) {
		dataA := NewTestData(newStage(), "", TestDataOptions{Files: []*FileData{
			{Name: "outside", Data: []byte("first")},
		}})
		digestA, err := dataA.Stage.GetContentDependencies(ctx, dataA.Conveyor, dataA.BuildContext)
		Expect(err).To(Succeed())

		dataB := NewTestData(newStage(), "", TestDataOptions{Files: []*FileData{
			{Name: "outside", Data: []byte("second")},
		}})
		digestB, err := dataB.Stage.GetContentDependencies(ctx, dataB.Conveyor, dataB.BuildContext)
		Expect(err).To(Succeed())
		Expect(digestB).NotTo(Equal(digestA))
	},
	Entry("COPY checksums the whole context", func() stage.Interface {
		return instruction.NewCopy(
			dockerfile.NewDockerfileStageInstruction(
				&instructions.CopyCommand{SourcesAndDest: instructions.SourcesAndDest{SourcePaths: []string{"$SRC/file"}, DestPath: "/payload"}},
				dockerfile.DockerfileStageInstructionOptions{ExpanderFactory: frontend.NewShlexExpanderFactory(parser.DefaultEscapeToken)},
			),
			nil,
			false,
			&stage.BaseStageOptions{},
		)
	}),
	Entry("ADD checksums the whole context", func() stage.Interface {
		return instruction.NewAdd(
			dockerfile.NewDockerfileStageInstruction(
				&instructions.AddCommand{SourcesAndDest: instructions.SourcesAndDest{SourcePaths: []string{"$SRC/file"}, DestPath: "/payload"}},
				dockerfile.DockerfileStageInstructionOptions{ExpanderFactory: frontend.NewShlexExpanderFactory(parser.DefaultEscapeToken)},
			),
			nil,
			false,
			&stage.BaseStageOptions{},
		)
	}),
)
