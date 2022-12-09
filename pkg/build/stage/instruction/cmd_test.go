package instruction

import (
	"context"
	"fmt"

	"github.com/moby/buildkit/frontend/dockerfile/instructions"
	. "github.com/onsi/gomega"

	"github.com/werf/werf/pkg/build/stage"
	"github.com/werf/werf/pkg/dockerfile"
)

var _ = DescribeTable("CMD digest",
	func(data *TestData) {
		ctx := context.Background()

		digest, err := data.Stage.GetDependencies(ctx, data.Conveyor, data.ContainerBackend, nil, data.StageImage, data.BuildContext)
		Expect(err).To(Succeed())

		fmt.Printf("calculated digest: %s\n", digest)
		fmt.Printf("expected digest: %s\n", data.ExpectedDigest)

		Expect(digest).To(Equal(data.ExpectedDigest))
	},

	Entry("CMD basic", NewTestData(
		NewCmd(
			dockerfile.NewDockerfileStageInstruction(
				&instructions.CmdCommand{ShellDependantCmdLine: instructions.ShellDependantCmdLine{CmdLine: []string{"/bin/bash", "-lec", "while true ; do date ; sleep 1 ; done"}, PrependShell: false}},
				dockerfile.DockerfileStageInstructionOptions{},
			),
			nil, false,
			&stage.BaseStageOptions{
				ImageName:   "example-image",
				ProjectName: "example-project",
			},
		),
		"6c176ed20cfc341dc35111c329a50d33a672556432524e30fe77c794fd19a41f",
		TestDataOptions{
			Files: []*FileData{
				{Name: "src/main/java/worker/Worker.java", Data: []byte(`package worker;`)},
				{Name: "src/Worker/Program.cs", Data: []byte(`namespace Worker {}`)},
			},
		},
	)),

	Entry("CMD with shell", NewTestData(
		NewCmd(
			dockerfile.NewDockerfileStageInstruction(
				&instructions.CmdCommand{ShellDependantCmdLine: instructions.ShellDependantCmdLine{CmdLine: []string{"/bin/bash", "-lec", "while true ; do date ; sleep 1 ; done"}, PrependShell: true}},
				dockerfile.DockerfileStageInstructionOptions{},
			),
			nil, false,
			&stage.BaseStageOptions{
				ImageName:   "example-image",
				ProjectName: "example-project",
			},
		),
		"1ba4c1cdfa5bce43a107e540d20c7d2655fcaeeba20e6b1f4feac2284a06f27a",
		TestDataOptions{
			Files: []*FileData{
				{Name: "src/main/java/worker/Worker.java", Data: []byte(`package worker;`)},
				{Name: "src/Worker/Program.cs", Data: []byte(`namespace Worker {}`)},
			},
		},
	)),

	Entry("CMD with changed context", NewTestData(
		NewCmd(
			dockerfile.NewDockerfileStageInstruction(
				&instructions.CmdCommand{ShellDependantCmdLine: instructions.ShellDependantCmdLine{CmdLine: []string{"/bin/bash", "-lec", "while true ; do date ; sleep 1 ; done"}, PrependShell: true}},
				dockerfile.DockerfileStageInstructionOptions{},
			),
			nil, false,
			&stage.BaseStageOptions{
				ImageName:   "example-image",
				ProjectName: "example-project",
			},
		),
		"1ba4c1cdfa5bce43a107e540d20c7d2655fcaeeba20e6b1f4feac2284a06f27a",
		TestDataOptions{
			Files: []*FileData{
				{Name: "src/main/java/worker/Worker.java", Data: []byte(`package worker2;`)},
				{Name: "src/Worker/Program.cs", Data: []byte(`namespace Worker2 {}`)},
			},
		},
	)),
)
