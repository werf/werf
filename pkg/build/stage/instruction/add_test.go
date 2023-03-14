package instruction

import (
	"context"
	"fmt"

	"github.com/moby/buildkit/frontend/dockerfile/instructions"
	. "github.com/onsi/gomega"

	"github.com/werf/werf/pkg/build/stage"
	"github.com/werf/werf/pkg/dockerfile"
)

var _ = DescribeTable("ADD digest",
	func(data *TestData) {
		ctx := context.Background()

		digest, err := data.Stage.GetDependencies(ctx, data.Conveyor, data.ContainerBackend, nil, data.StageImage, data.BuildContext)
		Expect(err).To(Succeed())

		fmt.Printf("calculated digest: %s\n", digest)
		fmt.Printf("expected digest: %s\n", data.ExpectedDigest)

		Expect(digest).To(Equal(data.ExpectedDigest))
	},

	Entry("ADD basic", NewTestData(
		NewAdd(
			dockerfile.NewDockerfileStageInstruction(
				&instructions.AddCommand{
					SourcesAndDest: instructions.SourcesAndDest{
						DestPath:    "/app",
						SourcePaths: []string{"src"},
					},
					Chown: "1000:1000",
					Chmod: "",
				},
				dockerfile.DockerfileStageInstructionOptions{},
			),
			nil, false,
			&stage.BaseStageOptions{
				ImageName:   "example-image",
				ProjectName: "example-project",
			},
		),
		"79d3642e997030deb225e0414f0c2d0e3c6681cc036899aaba62895f7e2ac4e3",
		TestDataOptions{
			Files: []*FileData{
				{Name: "src/main/java/worker/Worker.java", Data: []byte(`package worker;`)},
				{Name: "src/Worker/Program.cs", Data: []byte(`namespace Worker {}`)},
			},
		},
	)),

	Entry("ADD with changed chown", NewTestData(
		NewAdd(
			dockerfile.NewDockerfileStageInstruction(
				&instructions.AddCommand{
					SourcesAndDest: instructions.SourcesAndDest{
						DestPath:    "/app",
						SourcePaths: []string{"src"},
					},
					Chown: "1000:1001",
					Chmod: "",
				},
				dockerfile.DockerfileStageInstructionOptions{},
			),
			nil, false,
			&stage.BaseStageOptions{
				ImageName:   "example-image",
				ProjectName: "example-project",
			},
		),
		"ccfda65ec4fa8667abe7a54b047a98158d122b39e224929dbb7c33ea466e7f5a",
		TestDataOptions{
			Files: []*FileData{
				{Name: "src/main/java/worker/Worker.java", Data: []byte(`package worker;`)},
				{Name: "src/Worker/Program.cs", Data: []byte(`namespace Worker {}`)},
				{Name: "pom.xml", Data: []byte(`<?xml version="1.0" encoding="UTF-8"?>`)},
			},
		},
	)),

	Entry("ADD with changed chmod", NewTestData(
		NewAdd(
			dockerfile.NewDockerfileStageInstruction(
				&instructions.AddCommand{
					SourcesAndDest: instructions.SourcesAndDest{
						DestPath:    "/app",
						SourcePaths: []string{"src"},
					},
					Chown: "1000:1001",
					Chmod: "0777",
				},
				dockerfile.DockerfileStageInstructionOptions{},
			),
			nil, false,
			&stage.BaseStageOptions{
				ImageName:   "example-image",
				ProjectName: "example-project",
			},
		),
		"19e86e1145aecd554d7d492add3c6c6ed51b022279f059f6c8c54a4eac9d07f0",
		TestDataOptions{
			Files: []*FileData{
				{Name: "src/main/java/worker/Worker.java", Data: []byte(`package worker;`)},
				{Name: "src/Worker/Program.cs", Data: []byte(`namespace Worker {}`)},
				{Name: "pom.xml", Data: []byte(`<?xml version="1.0" encoding="UTF-8"?>`)},
			},
		},
	)),

	Entry("ADD with changed sources paths", NewTestData(
		NewAdd(
			dockerfile.NewDockerfileStageInstruction(
				&instructions.AddCommand{
					SourcesAndDest: instructions.SourcesAndDest{
						DestPath:    "/app",
						SourcePaths: []string{"src", "pom.xml"},
					},
					Chown: "1000:1001",
					Chmod: "0777",
				},
				dockerfile.DockerfileStageInstructionOptions{},
			),
			nil, false,
			&stage.BaseStageOptions{
				ImageName:   "example-image",
				ProjectName: "example-project",
			},
		),
		"4d17db1e6926bbe9e4f7b70d18bb055b7735f6bfb6db35452524bde561e8b95f",
		TestDataOptions{
			Files: []*FileData{
				{Name: "src/main/java/worker/Worker.java", Data: []byte(`package worker;`)},
				{Name: "src/Worker/Program.cs", Data: []byte(`namespace Worker {}`)},
				{Name: "pom.xml", Data: []byte(`<?xml version="1.0" encoding="UTF-8"?>`)},
			},
		},
	)),

	Entry("ADD with changed source files", NewTestData(
		NewAdd(
			dockerfile.NewDockerfileStageInstruction(
				&instructions.AddCommand{
					SourcesAndDest: instructions.SourcesAndDest{
						DestPath:    "/app",
						SourcePaths: []string{"src", "pom.xml"},
					},
					Chown: "1000:1001",
					Chmod: "0777",
				},
				dockerfile.DockerfileStageInstructionOptions{},
			),
			nil, false,
			&stage.BaseStageOptions{
				ImageName:   "example-image",
				ProjectName: "example-project",
			},
		),
		"372ed0cb6fff0a58e087fa8bf19e1f62a146d3983ca4510c496c452db8a7080e",
		TestDataOptions{
			Files: []*FileData{
				{Name: "src/main/java/worker/Worker.java", Data: []byte(`package worker2;`)},
				{Name: "src/Worker/Program.cs", Data: []byte(`namespace Worker2 {}`)},
				{Name: "pom.xml", Data: []byte(`<?xml version="1.0" encoding="UTF-8"?>`)},
			},
		},
	)),

	Entry("ADD with changed destination path", NewTestData(
		NewAdd(
			dockerfile.NewDockerfileStageInstruction(
				&instructions.AddCommand{
					SourcesAndDest: instructions.SourcesAndDest{
						DestPath:    "/app2",
						SourcePaths: []string{"src", "pom.xml"},
					},
					Chown: "1000:1001",
					Chmod: "0777",
				},
				dockerfile.DockerfileStageInstructionOptions{},
			),
			nil, false,
			&stage.BaseStageOptions{
				ImageName:   "example-image",
				ProjectName: "example-project",
			},
		),
		"825c66ecb926ed7897fc99f7686ed4fc2a7f8133d6a66860c8755772764d0293",
		TestDataOptions{
			Files: []*FileData{
				{Name: "src/main/java/worker/Worker.java", Data: []byte(`package worker2;`)},
				{Name: "src/Worker/Program.cs", Data: []byte(`namespace Worker2 {}`)},
				{Name: "pom.xml", Data: []byte(`<?xml version="1.0" encoding="UTF-8"?>`)},
			},
		},
	)),
)
