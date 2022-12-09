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
					SourcesAndDest: []string{"src/", "doc/", "/app"},
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
					SourcesAndDest: []string{"src/", "doc/", "/app"},
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
		"f49fa51e16c8cf0728f1cb9bd2be873555c5825d00cfac406057e6357d9900ed",
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
		"a1374babfa54a99e2efa0dc16e0c267395e2adfa6ee66d177ba9813b1745f0fa",
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
		"a1374babfa54a99e2efa0dc16e0c267395e2adfa6ee66d177ba9813b1745f0fa",
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
		"834afb3164923c905c25b6238e1366e533d7b65c0e8c1011163131399e200d36",
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
		"919c20233b0060d5726cf23c99c199f6055d0db050bea6f95d29ca3796397912",
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
