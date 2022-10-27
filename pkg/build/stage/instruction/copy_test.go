package instruction

import (
	"context"
	"fmt"

	. "github.com/onsi/gomega"

	"github.com/werf/werf/pkg/build/stage"
	"github.com/werf/werf/pkg/dockerfile"
	dockerfile_instruction "github.com/werf/werf/pkg/dockerfile/instruction"
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
		NewCopy("COPY", dockerfile.NewDockerfileStageInstruction(
			dockerfile_instruction.NewCopy("", "", []string{"src/", "doc/"}, "/app", "", ""),
		), nil, false,
			&stage.BaseStageOptions{
				ImageName:   "example-image",
				ProjectName: "example-project",
			},
		),
		"5d6051dce3ede19b81baaafd27adc2ed27c10a3ade3c81f520043afe3cb0d4f6",
		TestDataOptions{
			Files: []*FileData{
				{Name: "src/main/java/worker/Worker.java", Data: []byte(`package worker;`)},
				{Name: "src/Worker/Program.cs", Data: []byte(`namespace Worker {}`)},
				{Name: "doc/README.md", Data: []byte(`# README.md`)},
			},
		},
	)),

	Entry("COPY with changed context files", NewTestData(
		NewCopy("COPY", dockerfile.NewDockerfileStageInstruction(
			dockerfile_instruction.NewCopy("", "", []string{"src/", "doc/"}, "/app", "", ""),
		), nil, false,
			&stage.BaseStageOptions{
				ImageName:   "example-image",
				ProjectName: "example-project",
			},
		),
		"2b30d18482d64e88c1afad2e5fc1da1b252663b955cb23548e5ecc164aa9baec",
		TestDataOptions{
			Files: []*FileData{
				{Name: "src/main/java/worker/Worker.java", Data: []byte(`package worker;`)},
				{Name: "src/Worker/Program.cs", Data: []byte(`namespace Worker {}`)},
				{Name: "doc/README.md", Data: []byte(`# Documentation`)},
			},
		},
	)),

	Entry("COPY from stage", NewTestData(
		NewCopy("COPY", NewDockerfileStageInstructionWithDependencyStages(
			dockerfile_instruction.NewCopy("", "base", []string{"src/", "doc/"}, "/app", "", ""),
			[]string{"base"},
		), nil, false,
			&stage.BaseStageOptions{
				ImageName:   "example-image",
				ProjectName: "example-project",
			},
		),
		"fa91138526b31f450e59e1924c4da74572af67b5de1b65c65d7a680b55035281",
		TestDataOptions{
			LastStageImageNameByWerfImage: map[string]string{
				"stage/base": "ghcr.io/werf/instruction-test:a71052baf9c6ace8171e59a2ae5ea1aede3fb89aa95d160ec354b205-1661868399091",
			},
		},
	)),

	Entry("COPY from changed stage", NewTestData(
		NewCopy("COPY", NewDockerfileStageInstructionWithDependencyStages(
			dockerfile_instruction.NewCopy("", "base", []string{"src/", "doc/"}, "/app", "", ""),
			[]string{"base"},
		), nil, false,
			&stage.BaseStageOptions{
				ImageName:   "example-image",
				ProjectName: "example-project",
			},
		),
		"195e34303f2c13b3e0374e9fe06db2542232028f202f1b03c9d50933c85f4ae0",
		TestDataOptions{
			LastStageImageNameByWerfImage: map[string]string{
				"stage/base": "ghcr.io/werf/instruction-test:4930d562bfbee9c931413c826137d49eff6a2e7d39519c1c9488a747-1655913653892",
			},
		},
	)),

	Entry("COPY from same stage, with changed context", NewTestData(
		NewCopy("COPY", NewDockerfileStageInstructionWithDependencyStages(
			dockerfile_instruction.NewCopy("", "base", []string{"src/", "doc/"}, "/app", "", ""),
			[]string{"base"},
		), nil, false,
			&stage.BaseStageOptions{
				ImageName:   "example-image",
				ProjectName: "example-project",
			},
		),
		"195e34303f2c13b3e0374e9fe06db2542232028f202f1b03c9d50933c85f4ae0",
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
		NewCopy("COPY", NewDockerfileStageInstructionWithDependencyStages(
			dockerfile_instruction.NewCopy("", "base", []string{"src/", "doc/"}, "/app2", "", ""),
			[]string{"base"},
		), nil, false,
			&stage.BaseStageOptions{
				ImageName:   "example-image",
				ProjectName: "example-project",
			},
		),
		"8339ee5efad1ef83edebbce532f8889a30589a3280dcbdb2aa6b951df2f8fc07",
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
		NewCopy("COPY", NewDockerfileStageInstructionWithDependencyStages(
			dockerfile_instruction.NewCopy("", "base", []string{"src/", "doc/"}, "/app2", "1000:1000", "0777"),
			[]string{"base"},
		), nil, false,
			&stage.BaseStageOptions{
				ImageName:   "example-image",
				ProjectName: "example-project",
			},
		),
		"dfc1bfa6afbb5ca271f6236553e9dd09ffbcfe2092732f0c9026babeba8c8303",
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
