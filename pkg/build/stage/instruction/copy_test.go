package instruction

import (
	"context"
	"fmt"

	"github.com/moby/buildkit/frontend/dockerfile/instructions"
	. "github.com/onsi/gomega"

	"github.com/werf/werf/pkg/build/stage"
	"github.com/werf/werf/pkg/dockerfile"
)

var _ = DescribeTable("COPY digest",
	func(data *TestData) {
		ctx := context.Background()

		digest, err := data.Stage.GetDependencies(ctx, data.Conveyor, data.ContainerBackend, nil, data.StageImage, data.BuildContext)
		Expect(err).To(Succeed())

		fmt.Printf("calculated digest: %s\n", digest)
		fmt.Printf("expected digest: %s\n", data.ExpectedDigest)

		Expect(digest).To(Equal(data.ExpectedDigest))
	},

	Entry("COPY basic", NewTestData(
		NewCopy(
			dockerfile.NewDockerfileStageInstruction(
				&instructions.CopyCommand{
					SourcesAndDest: instructions.SourcesAndDest{
						DestPath:    "/app",
						SourcePaths: []string{"src/", "doc/"},
					},
				},
				dockerfile.DockerfileStageInstructionOptions{},
			),
			nil, false,
			&stage.BaseStageOptions{
				ImageName:   "example-image",
				ProjectName: "example-project",
			},
		),
		"b37df6b6878045c3bf0e9972e4618d6157dfa06bec990a83e0a1881ec12621b1",
		TestDataOptions{
			Files: []*FileData{
				{Name: "src/main/java/worker/Worker.java", Data: []byte(`package worker;`)},
				{Name: "src/Worker/Program.cs", Data: []byte(`namespace Worker {}`)},
				{Name: "doc/README.md", Data: []byte(`# README.md`)},
			},
		},
	)),

	Entry("COPY with changed context files", NewTestData(
		NewCopy(
			dockerfile.NewDockerfileStageInstruction(
				&instructions.CopyCommand{
					SourcesAndDest: instructions.SourcesAndDest{
						DestPath:    "/app",
						SourcePaths: []string{"src/", "doc/"},
					},
				},
				dockerfile.DockerfileStageInstructionOptions{},
			),
			nil, false,
			&stage.BaseStageOptions{
				ImageName:   "example-image",
				ProjectName: "example-project",
			},
		),
		"88158fb775544cccd8bd1a345a89bda08434620dda37efe3867df077007602e1",
		TestDataOptions{
			Files: []*FileData{
				{Name: "src/main/java/worker/Worker.java", Data: []byte(`package worker;`)},
				{Name: "src/Worker/Program.cs", Data: []byte(`namespace Worker {}`)},
				{Name: "doc/README.md", Data: []byte(`# Documentation`)},
			},
		},
	)),

	Entry("COPY from stage", NewTestData(
		NewCopy(
			NewDockerfileStageInstructionWithDependencyStages(
				&instructions.CopyCommand{
					From: "base",
					SourcesAndDest: instructions.SourcesAndDest{
						DestPath:    "/app",
						SourcePaths: []string{"src/", "doc/"},
					},
				},
				[]string{"base"},
			),
			nil, false,
			&stage.BaseStageOptions{
				ImageName:   "example-image",
				ProjectName: "example-project",
			},
		),
		"c65e7b2d1c3a7614106b6eb061a6bf4712cf833817ce4e3b79d8af8de22c2ba0",
		TestDataOptions{
			LastStageImageNameByWerfImage: map[string]string{
				"stage/base": "ghcr.io/werf/instruction-test:a71052baf9c6ace8171e59a2ae5ea1aede3fb89aa95d160ec354b205-1661868399091",
			},
		},
	)),

	Entry("COPY from changed stage", NewTestData(
		NewCopy(
			NewDockerfileStageInstructionWithDependencyStages(
				&instructions.CopyCommand{
					From: "base",
					SourcesAndDest: instructions.SourcesAndDest{
						DestPath:    "/app",
						SourcePaths: []string{"src/", "doc/"},
					},
				},
				[]string{"base"},
			),
			nil, false,
			&stage.BaseStageOptions{
				ImageName:   "example-image",
				ProjectName: "example-project",
			},
		),
		"c65e7b2d1c3a7614106b6eb061a6bf4712cf833817ce4e3b79d8af8de22c2ba0",
		TestDataOptions{
			LastStageImageNameByWerfImage: map[string]string{
				"stage/base": "ghcr.io/werf/instruction-test:4930d562bfbee9c931413c826137d49eff6a2e7d39519c1c9488a747-1655913653892",
			},
		},
	)),

	Entry("COPY from same stage, with changed context", NewTestData(
		NewCopy(
			NewDockerfileStageInstructionWithDependencyStages(
				&instructions.CopyCommand{
					From: "base",
					SourcesAndDest: instructions.SourcesAndDest{
						DestPath:    "/app",
						SourcePaths: []string{"src/", "doc/"},
					},
				},
				[]string{"base"},
			),
			nil, false,
			&stage.BaseStageOptions{
				ImageName:   "example-image",
				ProjectName: "example-project",
			},
		),
		"c65e7b2d1c3a7614106b6eb061a6bf4712cf833817ce4e3b79d8af8de22c2ba0",
		TestDataOptions{
			LastStageImageNameByWerfImage: map[string]string{
				"stage/base": "ghcr.io/werf/instruction-test:4930d562bfbee9c931413c826137d49eff6a2e7d39519c1c9488a747-1655913653892",
			},
			Files: []*FileData{
				{Name: "src/main/java/worker/Worker.java", Data: []byte(`package worker;`)},
				{Name: "src/Worker/Program.cs", Data: []byte(`namespace Worker {}`)},
				{Name: "doc/README.md", Data: []byte(`# Documentation`)},
			},
		},
	)),

	Entry("COPY from same stage, with changed destination", NewTestData(
		NewCopy(
			NewDockerfileStageInstructionWithDependencyStages(
				&instructions.CopyCommand{
					From: "base",
					SourcesAndDest: instructions.SourcesAndDest{
						DestPath:    "/app2",
						SourcePaths: []string{"src/", "doc/"},
					},
				},
				[]string{"base"},
			),
			nil, false,
			&stage.BaseStageOptions{
				ImageName:   "example-image",
				ProjectName: "example-project",
			},
		),
		"d1898898971186720898055e0682e356ff9e53178b6ac7c629f255113c6b17a0",
		TestDataOptions{
			LastStageImageNameByWerfImage: map[string]string{
				"stage/base": "ghcr.io/werf/instruction-test:4930d562bfbee9c931413c826137d49eff6a2e7d39519c1c9488a747-1655913653892",
			},
			Files: []*FileData{
				{Name: "src/main/java/worker/Worker.java", Data: []byte(`package worker;`)},
				{Name: "src/Worker/Program.cs", Data: []byte(`namespace Worker {}`)},
				{Name: "doc/README.md", Data: []byte(`# Documentation`)},
			},
		},
	)),

	Entry("COPY from same stage, with changed owner and modes", NewTestData(
		NewCopy(
			NewDockerfileStageInstructionWithDependencyStages(
				&instructions.CopyCommand{
					From: "base",
					SourcesAndDest: instructions.SourcesAndDest{
						DestPath:    "/app2",
						SourcePaths: []string{"src/", "doc/"},
					},
					Chown: "1000:1000",
					Chmod: "0777",
				},
				[]string{"base"},
			),
			nil, false,
			&stage.BaseStageOptions{
				ImageName:   "example-image",
				ProjectName: "example-project",
			},
		),
		"18a8898734bef3802598b82bff90e2fd8fe37e5b4806f269553e70f017374e35",
		TestDataOptions{
			LastStageImageNameByWerfImage: map[string]string{
				"stage/base": "ghcr.io/werf/instruction-test:4930d562bfbee9c931413c826137d49eff6a2e7d39519c1c9488a747-1655913653892",
			},
			Files: []*FileData{
				{Name: "src/main/java/worker/Worker.java", Data: []byte(`package worker;`)},
				{Name: "src/Worker/Program.cs", Data: []byte(`namespace Worker {}`)},
				{Name: "doc/README.md", Data: []byte(`# Documentation`)},
			},
		},
	)),
)
