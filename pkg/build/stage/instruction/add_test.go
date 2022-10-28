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
		NewAdd("ADD",
			dockerfile.NewDockerfileStageInstruction(
				&instructions.AddCommand{SourcesAndDest: []string{"src", "/app"}, Chown: "1000:1000", Chmod: ""},
			),
			nil, false,
			&stage.BaseStageOptions{
				ImageName:   "example-image",
				ProjectName: "example-project",
			},
		),
		"88c31da85ac26ae35d29462c6dc309c2a02997c0de92b4ccee7db2e41be17187",
		TestDataOptions{
			Files: []*FileData{
				{Name: "src/main/java/worker/Worker.java", Data: []byte(`package worker;`)},
				{Name: "src/Worker/Program.cs", Data: []byte(`namespace Worker {}`)},
			},
		},
	)),

	Entry("ADD with changed chown", NewTestData(
		NewAdd("ADD",
			dockerfile.NewDockerfileStageInstruction(
				&instructions.AddCommand{SourcesAndDest: []string{"src", "/app"}, Chown: "1000:1001", Chmod: ""},
			),
			nil, false,
			&stage.BaseStageOptions{
				ImageName:   "example-image",
				ProjectName: "example-project",
			},
		),
		"846ef29e994224dd84bf0a5de47b0b3255c8681b8178e8da5611b21547cd182b",
		TestDataOptions{
			Files: []*FileData{
				{Name: "src/main/java/worker/Worker.java", Data: []byte(`package worker;`)},
				{Name: "src/Worker/Program.cs", Data: []byte(`namespace Worker {}`)},
				{Name: "pom.xml", Data: []byte(`<?xml version="1.0" encoding="UTF-8"?>`)},
			},
		},
	)),

	Entry("ADD with changed chmod", NewTestData(
		NewAdd("ADD",
			dockerfile.NewDockerfileStageInstruction(
				&instructions.AddCommand{SourcesAndDest: []string{"src", "/app"}, Chown: "1000:1001", Chmod: "0777"},
			),
			nil, false,
			&stage.BaseStageOptions{
				ImageName:   "example-image",
				ProjectName: "example-project",
			},
		),
		"cef21e87710631a08edeb176a9487f81ae20171c22ec4537a3dc8fbc67aca868",
		TestDataOptions{
			Files: []*FileData{
				{Name: "src/main/java/worker/Worker.java", Data: []byte(`package worker;`)},
				{Name: "src/Worker/Program.cs", Data: []byte(`namespace Worker {}`)},
				{Name: "pom.xml", Data: []byte(`<?xml version="1.0" encoding="UTF-8"?>`)},
			},
		},
	)),

	Entry("ADD with changed sources paths", NewTestData(
		NewAdd("ADD",
			dockerfile.NewDockerfileStageInstruction(
				&instructions.AddCommand{SourcesAndDest: []string{"src", "pom.xml", "/app"}, Chown: "1000:1001", Chmod: "0777"},
			),
			nil, false,
			&stage.BaseStageOptions{
				ImageName:   "example-image",
				ProjectName: "example-project",
			},
		),
		"97f3f8a240902d73ec9a209f6c8368047b56d9247bdf9da88a40ac5dba925209",
		TestDataOptions{
			Files: []*FileData{
				{Name: "src/main/java/worker/Worker.java", Data: []byte(`package worker;`)},
				{Name: "src/Worker/Program.cs", Data: []byte(`namespace Worker {}`)},
				{Name: "pom.xml", Data: []byte(`<?xml version="1.0" encoding="UTF-8"?>`)},
			},
		},
	)),

	Entry("ADD with changed source files", NewTestData(
		NewAdd("ADD",
			dockerfile.NewDockerfileStageInstruction(
				&instructions.AddCommand{SourcesAndDest: []string{"src", "pom.xml", "/app"}, Chown: "1000:1001", Chmod: "0777"},
			),
			nil, false,
			&stage.BaseStageOptions{
				ImageName:   "example-image",
				ProjectName: "example-project",
			},
		),
		"60178e0b174bd1bce1cd29f8132ea84cc7212773b6fce9fad3ddff842d5cf2e0",
		TestDataOptions{
			Files: []*FileData{
				{Name: "src/main/java/worker/Worker.java", Data: []byte(`package worker2;`)},
				{Name: "src/Worker/Program.cs", Data: []byte(`namespace Worker2 {}`)},
				{Name: "pom.xml", Data: []byte(`<?xml version="1.0" encoding="UTF-8"?>`)},
			},
		},
	)),

	Entry("ADD with changed destination path", NewTestData(
		NewAdd("ADD",
			dockerfile.NewDockerfileStageInstruction(
				&instructions.AddCommand{SourcesAndDest: []string{"src", "pom.xml", "/app2"}, Chown: "1000:1001", Chmod: "0777"},
			),
			nil, false,
			&stage.BaseStageOptions{
				ImageName:   "example-image",
				ProjectName: "example-project",
			},
		),
		"c1f03d5701951fe9c5836957c753c9486f22e14b2d9291780ae70f288e531e1c",
		TestDataOptions{
			Files: []*FileData{
				{Name: "src/main/java/worker/Worker.java", Data: []byte(`package worker2;`)},
				{Name: "src/Worker/Program.cs", Data: []byte(`namespace Worker2 {}`)},
				{Name: "pom.xml", Data: []byte(`<?xml version="1.0" encoding="UTF-8"?>`)},
			},
		},
	)),
)
