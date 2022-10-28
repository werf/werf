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
		NewCopy("COPY",
			dockerfile.NewDockerfileStageInstruction(
				&instructions.CopyCommand{
					SourcesAndDest: []string{"src/", "doc/", "/app"},
				},
			),
			nil, false,
			&stage.BaseStageOptions{
				ImageName:   "example-image",
				ProjectName: "example-project",
			},
		),
		"35bf3e72310b96b9fc7861ca705b0093935d3d830388dd3ebc47e89dad68151a",
		TestDataOptions{
			Files: []*FileData{
				{Name: "src/main/java/worker/Worker.java", Data: []byte(`package worker;`)},
				{Name: "src/Worker/Program.cs", Data: []byte(`namespace Worker {}`)},
				{Name: "doc/README.md", Data: []byte(`# README.md`)},
			},
		},
	)),

	Entry("COPY with changed context files", NewTestData(
		NewCopy("COPY",
			dockerfile.NewDockerfileStageInstruction(
				&instructions.CopyCommand{
					SourcesAndDest: []string{"src/", "doc/", "/app"},
				},
			),
			nil, false,
			&stage.BaseStageOptions{
				ImageName:   "example-image",
				ProjectName: "example-project",
			},
		),
		"2df11ae4260a97665e30f69eea0c86057e5a277ade1ad273af1b3a8c85b6a651",
		TestDataOptions{
			Files: []*FileData{
				{Name: "src/main/java/worker/Worker.java", Data: []byte(`package worker;`)},
				{Name: "src/Worker/Program.cs", Data: []byte(`namespace Worker {}`)},
				{Name: "doc/README.md", Data: []byte(`# Documentation`)},
			},
		},
	)),

	Entry("COPY from stage", NewTestData(
		NewCopy("COPY",
			NewDockerfileStageInstructionWithDependencyStages(
				&instructions.CopyCommand{
					From:           "base",
					SourcesAndDest: []string{"src/", "doc/", "/app"},
				},
				[]string{"base"},
			),
			nil, false,
			&stage.BaseStageOptions{
				ImageName:   "example-image",
				ProjectName: "example-project",
			},
		),
		"bb6495521017b98b1a6ebd5c24ae8881e2565cd64a09a825ffe7f5208da857e8",
		TestDataOptions{
			LastStageImageNameByWerfImage: map[string]string{
				"stage/base": "ghcr.io/werf/instruction-test:a71052baf9c6ace8171e59a2ae5ea1aede3fb89aa95d160ec354b205-1661868399091",
			},
		},
	)),

	Entry("COPY from changed stage", NewTestData(
		NewCopy("COPY",
			NewDockerfileStageInstructionWithDependencyStages(
				&instructions.CopyCommand{
					From:           "base",
					SourcesAndDest: []string{"src/", "doc/", "/app"},
				},
				[]string{"base"},
			),
			nil, false,
			&stage.BaseStageOptions{
				ImageName:   "example-image",
				ProjectName: "example-project",
			},
		),
		"60554221561909f478206d108bed14b781c6e642dee87a53b8ef6a961372d887",
		TestDataOptions{
			LastStageImageNameByWerfImage: map[string]string{
				"stage/base": "ghcr.io/werf/instruction-test:4930d562bfbee9c931413c826137d49eff6a2e7d39519c1c9488a747-1655913653892",
			},
		},
	)),

	Entry("COPY from same stage, with changed context", NewTestData(
		NewCopy("COPY",
			NewDockerfileStageInstructionWithDependencyStages(
				&instructions.CopyCommand{
					From:           "base",
					SourcesAndDest: []string{"src/", "doc/", "/app"},
				},
				[]string{"base"},
			),
			nil, false,
			&stage.BaseStageOptions{
				ImageName:   "example-image",
				ProjectName: "example-project",
			},
		),
		"60554221561909f478206d108bed14b781c6e642dee87a53b8ef6a961372d887",
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
		NewCopy("COPY",
			NewDockerfileStageInstructionWithDependencyStages(
				&instructions.CopyCommand{
					From:           "base",
					SourcesAndDest: []string{"src/", "doc/", "/app2"},
				},
				[]string{"base"},
			),
			nil, false,
			&stage.BaseStageOptions{
				ImageName:   "example-image",
				ProjectName: "example-project",
			},
		),
		"987f2b58c05b5b53fa523ac6c25169d9bbf29826f450d541fec0ed69354dbfe2",
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
		NewCopy("COPY",
			NewDockerfileStageInstructionWithDependencyStages(
				&instructions.CopyCommand{
					From:           "base",
					SourcesAndDest: []string{"src/", "doc/", "/app2"},
					Chown:          "1000:1000",
					Chmod:          "0777",
				},
				[]string{"base"},
			),
			nil, false,
			&stage.BaseStageOptions{
				ImageName:   "example-image",
				ProjectName: "example-project",
			},
		),
		"7cb785382f1e347e16b87f397ed7ea06b000dc04a7de74b47f583e7cf85db2b5",
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
