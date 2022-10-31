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
		NewCmd("CMD",
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
		"c52718afd8fe6f79c1039464791418dbca4ac242a48fc1dae0494880fa858c56",
		TestDataOptions{
			Files: []*FileData{
				{Name: "src/main/java/worker/Worker.java", Data: []byte(`package worker;`)},
				{Name: "src/Worker/Program.cs", Data: []byte(`namespace Worker {}`)},
			},
		},
	)),

	Entry("CMD with shell", NewTestData(
		NewCmd("CMD",
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
		"a763588e431e8f3f3ff846898525730ec3d19535328608d9c7b0301345d101f4",
		TestDataOptions{
			Files: []*FileData{
				{Name: "src/main/java/worker/Worker.java", Data: []byte(`package worker;`)},
				{Name: "src/Worker/Program.cs", Data: []byte(`namespace Worker {}`)},
			},
		},
	)),

	Entry("CMD with changed context", NewTestData(
		NewCmd("CMD",
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
		"a763588e431e8f3f3ff846898525730ec3d19535328608d9c7b0301345d101f4",
		TestDataOptions{
			Files: []*FileData{
				{Name: "src/main/java/worker/Worker.java", Data: []byte(`package worker2;`)},
				{Name: "src/Worker/Program.cs", Data: []byte(`namespace Worker2 {}`)},
			},
		},
	)),
)
