package instruction

import (
	"context"
	"fmt"
	"strings"

	"github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/werf/werf/pkg/build/stage"
	"github.com/werf/werf/pkg/container_backend"
	"github.com/werf/werf/pkg/dockerfile"
	dockerfile_instruction "github.com/werf/werf/pkg/dockerfile/instruction"
	"github.com/werf/werf/pkg/util"
)

var (
	Entry         = ginkgo.Entry
	DescribeTable = ginkgo.DescribeTable
)

var _ = DescribeTable("calculating digest and configuring builder",
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
		"bf369f0c1d046108f1a7314b086cefe109b8edc01878033eccf41f1ff47e2c73",
		[]*FileData{
			{Name: "src/main/java/worker/Worker.java", Data: []byte(`package worker;`)},
			{Name: "src/Worker/Program.cs", Data: []byte(`namespace Worker {}`)},
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
		"2a9fc5cdddb5443ff19bcc6366a3ad42f2ac8e86a3779dadf0ac34857b3928cb",
		[]*FileData{
			{Name: "src/main/java/worker/Worker.java", Data: []byte(`package worker;`)},
			{Name: "src/Worker/Program.cs", Data: []byte(`namespace Worker {}`)},
			{Name: "pom.xml", Data: []byte(`<?xml version="1.0" encoding="UTF-8"?>`)},
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
		"e2c17b3d62eb61470f16544a4ea42b66b37326dc5f10d61b7a9ebab04e53e7a7",
		[]*FileData{
			{Name: "src/main/java/worker/Worker.java", Data: []byte(`package worker;`)},
			{Name: "src/Worker/Program.cs", Data: []byte(`namespace Worker {}`)},
			{Name: "pom.xml", Data: []byte(`<?xml version="1.0" encoding="UTF-8"?>`)},
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
		"cc635522d684dd86345d3a074722ecaf82fd33c33912c6667caea62177ba2540",
		[]*FileData{
			{Name: "src/main/java/worker/Worker.java", Data: []byte(`package worker;`)},
			{Name: "src/Worker/Program.cs", Data: []byte(`namespace Worker {}`)},
			{Name: "pom.xml", Data: []byte(`<?xml version="1.0" encoding="UTF-8"?>`)},
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
		"15add63c912f20138ed0eb4a2bb906bc3c3fa70d0d0ac26c7871363484a84c94",
		[]*FileData{
			{Name: "src/main/java/worker/Worker.java", Data: []byte(`package worker2;`)},
			{Name: "src/Worker/Program.cs", Data: []byte(`namespace Worker2 {}`)},
			{Name: "pom.xml", Data: []byte(`<?xml version="1.0" encoding="UTF-8"?>`)},
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
		"15b2f0b0d9b569a55f0af7c8326e2d5d8792f5a7df1f8fd4e612034f74a015e1",
		[]*FileData{
			{Name: "src/main/java/worker/Worker.java", Data: []byte(`package worker2;`)},
			{Name: "src/Worker/Program.cs", Data: []byte(`namespace Worker2 {}`)},
			{Name: "pom.xml", Data: []byte(`<?xml version="1.0" encoding="UTF-8"?>`)},
		},
	)),
)

type BuildContextStub struct {
	container_backend.BuildContextArchiver

	Files []*FileData
}

type FileData struct {
	Name string
	Data []byte
}

func NewBuildContextStub(files []*FileData) *BuildContextStub {
	return &BuildContextStub{Files: files}
}

func (buildContext *BuildContextStub) CalculatePathsChecksum(ctx context.Context, paths []string) (string, error) {
	var args []string

	for _, p := range paths {
		for _, f := range buildContext.Files {
			if f.Name == p {
				args = append(args, string(f.Data))
				break
			}
		}

		for _, f := range buildContext.Files {
			if strings.HasPrefix(f.Name, p) {
				args = append(args, string(f.Data))
				break
			}
		}
	}

	return util.Sha256Hash(args...), nil
}
