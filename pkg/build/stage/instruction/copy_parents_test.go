package instruction_test

import (
	"bytes"

	"github.com/moby/buildkit/frontend/dockerfile/instructions"
	"github.com/moby/buildkit/frontend/dockerfile/parser"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/werf/werf/v2/pkg/build/stage"
	"github.com/werf/werf/v2/pkg/build/stage/instruction"
)

func newCopyStage(dockerfileText string) *instruction.Copy {
	p, err := parser.Parse(bytes.NewReader([]byte(dockerfileText)))
	Expect(err).To(Succeed())

	dockerStages, _, err := instructions.Parse(p.AST, nil)
	Expect(err).To(Succeed())
	Expect(dockerStages).NotTo(BeEmpty())

	var copyCommand *instructions.CopyCommand
	for _, cmd := range dockerStages[len(dockerStages)-1].Commands {
		if c, ok := cmd.(*instructions.CopyCommand); ok {
			copyCommand = c
		}
	}
	Expect(copyCommand).NotTo(BeNil())

	return instruction.NewCopy(
		NewDockerfileStageInstructionWithDependencyStages(copyCommand, nil),
		nil, false,
		&stage.BaseStageOptions{ImageName: "example-image", ProjectName: "example-project"},
	)
}

var _ = Describe("Test COPY --parents", func() {
	const dockerfileWithParents = "FROM alpine\nCOPY --parents ./aaa/ ./bbb/ /target\n"
	const dockerfileWithoutParents = "FROM alpine\nCOPY ./aaa/ ./bbb/ /target\n"

	It("Test propagates the parents flag to the backend instruction", func(ctx SpecContext) {
		stg := newCopyStage(dockerfileWithParents)
		data := NewTestData(stg, "", TestDataOptions{})

		Expect(stg.ExpandInstruction(data.Conveyor, map[string]string{})).To(Succeed())

		backend, source := instruction.ExportCopyCommands(stg)
		Expect(source.Parents).To(BeTrue())
		Expect(backend.Parents).To(BeTrue())
	})

	It("Test changes the stage digest", func(ctx SpecContext) {
		files := []*FileData{{Name: "aaa/x", Data: []byte(`x`)}, {Name: "bbb/y", Data: []byte(`y`)}}

		withParents := NewTestData(newCopyStage(dockerfileWithParents), "", TestDataOptions{Files: files})
		digestWithParents, err := withParents.Stage.GetDependencies(ctx, withParents.Conveyor, withParents.ContainerBackend, nil, withParents.StageImage, withParents.BuildContext)
		Expect(err).To(Succeed())

		withoutParents := NewTestData(newCopyStage(dockerfileWithoutParents), "", TestDataOptions{Files: files})
		digestWithoutParents, err := withoutParents.Stage.GetDependencies(ctx, withoutParents.Conveyor, withoutParents.ContainerBackend, nil, withoutParents.StageImage, withoutParents.BuildContext)
		Expect(err).To(Succeed())

		Expect(digestWithParents).NotTo(Equal(digestWithoutParents))
	})
})
