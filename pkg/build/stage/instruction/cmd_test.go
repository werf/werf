package instruction

import (
	"context"
	"fmt"

	. "github.com/onsi/gomega"

	"github.com/werf/werf/pkg/build/stage"
	"github.com/werf/werf/pkg/dockerfile"
	dockerfile_instruction "github.com/werf/werf/pkg/dockerfile/instruction"
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
		NewCmd("CMD", dockerfile.NewDockerfileStageInstruction(
			dockerfile_instruction.NewCmd("", []string{"/bin/bash", "-lec", "while true ; do date ; sleep 1 ; done"}, false)), nil, false,
			&stage.BaseStageOptions{
				ImageName:   "example-image",
				ProjectName: "example-project",
			},
		),
		"0aa7a4cb09a46c09f5e3f66ebf96ca45162aea10747bad4dd92d269272eede4a",
		TestDataOptions{
			Files: []*FileData{
				{Name: "src/main/java/worker/Worker.java", Data: []byte(`package worker;`)},
				{Name: "src/Worker/Program.cs", Data: []byte(`namespace Worker {}`)},
			},
		},
	)),

	Entry("CMD with shell", NewTestData(
		NewCmd("CMD", dockerfile.NewDockerfileStageInstruction(
			dockerfile_instruction.NewCmd("", []string{"/bin/bash", "-lec", "while true ; do date ; sleep 1 ; done"}, true)), nil, false,
			&stage.BaseStageOptions{
				ImageName:   "example-image",
				ProjectName: "example-project",
			},
		),
		"f0c14536daf15ca56863306fdc6763f5b00b2a28f24a7ffb38207d11d28d2adb",
		TestDataOptions{
			Files: []*FileData{
				{Name: "src/main/java/worker/Worker.java", Data: []byte(`package worker;`)},
				{Name: "src/Worker/Program.cs", Data: []byte(`namespace Worker {}`)},
			},
		},
	)),

	Entry("CMD with changed context", NewTestData(
		NewCmd("CMD", dockerfile.NewDockerfileStageInstruction(
			dockerfile_instruction.NewCmd("", []string{"/bin/bash", "-lec", "while true ; do date ; sleep 1 ; done"}, true)), nil, false,
			&stage.BaseStageOptions{
				ImageName:   "example-image",
				ProjectName: "example-project",
			},
		),
		"f0c14536daf15ca56863306fdc6763f5b00b2a28f24a7ffb38207d11d28d2adb",
		TestDataOptions{
			Files: []*FileData{
				{Name: "src/main/java/worker/Worker.java", Data: []byte(`package worker2;`)},
				{Name: "src/Worker/Program.cs", Data: []byte(`namespace Worker2 {}`)},
			},
		},
	)),
)
