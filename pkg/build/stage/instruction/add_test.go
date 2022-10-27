package instruction

import (
	"context"
	"fmt"

	. "github.com/onsi/gomega"

	"github.com/werf/werf/pkg/build/stage"
	"github.com/werf/werf/pkg/dockerfile"
	dockerfile_instruction "github.com/werf/werf/pkg/dockerfile/instruction"
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
		NewAdd("ADD", dockerfile.NewDockerfileStageInstruction(
			dockerfile_instruction.NewAdd("", []string{"src"}, "/app", "1000:1000", "")), nil, false,
			&stage.BaseStageOptions{
				ImageName:   "example-image",
				ProjectName: "example-project",
			},
		),
		"febdd3676850b1b28161dc2a518495ba8e413b1aab3f28c7121c44c9da7d1212",
		TestDataOptions{
			Files: []*FileData{
				{Name: "src/main/java/worker/Worker.java", Data: []byte(`package worker;`)},
				{Name: "src/Worker/Program.cs", Data: []byte(`namespace Worker {}`)},
			},
		},
	)),

	Entry("ADD with changed chown", NewTestData(
		NewAdd("ADD", dockerfile.NewDockerfileStageInstruction(
			dockerfile_instruction.NewAdd("", []string{"src"}, "/app", "1000:1001", "")), nil, false,
			&stage.BaseStageOptions{
				ImageName:   "example-image",
				ProjectName: "example-project",
			},
		),
		"7f2ce5be335c61b5117446a3e508af27af1217bab68319a690f0f8d4653bed21",
		TestDataOptions{
			Files: []*FileData{
				{Name: "src/main/java/worker/Worker.java", Data: []byte(`package worker;`)},
				{Name: "src/Worker/Program.cs", Data: []byte(`namespace Worker {}`)},
				{Name: "pom.xml", Data: []byte(`<?xml version="1.0" encoding="UTF-8"?>`)},
			},
		},
	)),

	Entry("ADD with changed chmod", NewTestData(
		NewAdd("ADD", dockerfile.NewDockerfileStageInstruction(
			dockerfile_instruction.NewAdd("", []string{"src"}, "/app", "1000:1001", "0777")), nil, false,
			&stage.BaseStageOptions{
				ImageName:   "example-image",
				ProjectName: "example-project",
			},
		),
		"f37bc012a10364919e30f5c6ac1ed9de8d5d32590ecb57564c449d1261439488",
		TestDataOptions{
			Files: []*FileData{
				{Name: "src/main/java/worker/Worker.java", Data: []byte(`package worker;`)},
				{Name: "src/Worker/Program.cs", Data: []byte(`namespace Worker {}`)},
				{Name: "pom.xml", Data: []byte(`<?xml version="1.0" encoding="UTF-8"?>`)},
			},
		},
	)),

	Entry("ADD with changed sources paths", NewTestData(
		NewAdd("ADD", dockerfile.NewDockerfileStageInstruction(
			dockerfile_instruction.NewAdd("", []string{"src", "pom.xml"}, "/app", "1000:1000", "0777")), nil, false,
			&stage.BaseStageOptions{
				ImageName:   "example-image",
				ProjectName: "example-project",
			},
		),
		"4a66595bf3f692d892cc1f5132a21939f681b5e38c59831b0359ba69889f1392",
		TestDataOptions{
			Files: []*FileData{
				{Name: "src/main/java/worker/Worker.java", Data: []byte(`package worker;`)},
				{Name: "src/Worker/Program.cs", Data: []byte(`namespace Worker {}`)},
				{Name: "pom.xml", Data: []byte(`<?xml version="1.0" encoding="UTF-8"?>`)},
			},
		},
	)),

	Entry("ADD with changed source files", NewTestData(
		NewAdd("ADD", dockerfile.NewDockerfileStageInstruction(
			dockerfile_instruction.NewAdd("", []string{"src", "pom.xml"}, "/app", "1000:1000", "0777")), nil, false,
			&stage.BaseStageOptions{
				ImageName:   "example-image",
				ProjectName: "example-project",
			},
		),
		"c52c2a9ff2aa7054ab1f5c57717775c4e00a3b80d618b8a2aacefeb48b582ac0",
		TestDataOptions{
			Files: []*FileData{
				{Name: "src/main/java/worker/Worker.java", Data: []byte(`package worker2;`)},
				{Name: "src/Worker/Program.cs", Data: []byte(`namespace Worker2 {}`)},
				{Name: "pom.xml", Data: []byte(`<?xml version="1.0" encoding="UTF-8"?>`)},
			},
		},
	)),

	Entry("ADD with changed destination path", NewTestData(
		NewAdd("ADD", dockerfile.NewDockerfileStageInstruction(
			dockerfile_instruction.NewAdd("", []string{"src", "pom.xml"}, "/app2", "1000:1000", "0777")), nil, false,
			&stage.BaseStageOptions{
				ImageName:   "example-image",
				ProjectName: "example-project",
			},
		),
		"a067cbedb63b837bc9583487b27ad2f5dcc9d744df03414b9b1436e90e8c514e",
		TestDataOptions{
			Files: []*FileData{
				{Name: "src/main/java/worker/Worker.java", Data: []byte(`package worker2;`)},
				{Name: "src/Worker/Program.cs", Data: []byte(`namespace Worker2 {}`)},
				{Name: "pom.xml", Data: []byte(`<?xml version="1.0" encoding="UTF-8"?>`)},
			},
		},
	)),
)
