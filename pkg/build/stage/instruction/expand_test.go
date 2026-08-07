package instruction_test

import (
	"github.com/moby/buildkit/frontend/dockerfile/instructions"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/werf/werf/v2/pkg/build/stage"
	"github.com/werf/werf/v2/pkg/build/stage/instruction"
	"github.com/werf/werf/v2/pkg/dockerfile"
	"github.com/werf/werf/v2/pkg/dockerfile/frontend"
)

type expandTestEntry struct {
	stg     stage.Interface
	compare func()
}

var _ = DescribeTable("backend instruction sync after env expansion",
	func(ctx SpecContext, entry expandTestEntry) {
		conveyor := stage.NewConveyorStub(
			stage.NewGiterminismManagerStub(stage.NewLocalGitRepoStub("test"), stage.NewGiterminismInspectorStub()),
			nil, nil,
		)
		Expect(entry.stg.ExpandDependencies(ctx, conveyor, map[string]string{"TEST_VAR": "resolved"})).To(Succeed())
		entry.compare()
	},

	Entry("COPY", func() expandTestEntry {
		stg := instruction.NewCopy(
			dockerfile.NewDockerfileStageInstruction(
				&instructions.CopyCommand{
					SourcesAndDest: instructions.SourcesAndDest{
						DestPath:    "${TEST_VAR}/dest",
						SourcePaths: []string{"src/"},
					},
					Chown: "${TEST_VAR}:group",
				},
				dockerfile.DockerfileStageInstructionOptions{
					ExpanderFactory: frontend.NewShlexExpanderFactory('\\'),
				},
			),
			nil, false,
			&stage.BaseStageOptions{ImageName: "test", ProjectName: "test"},
		)
		return expandTestEntry{
			stg: stg,
			compare: func() {
				backend, source := instruction.ExportCopyCommands(stg)
				Expect(backend).To(Equal(source))
			},
		}
	}()),

	Entry("ADD", func() expandTestEntry {
		stg := instruction.NewAdd(
			dockerfile.NewDockerfileStageInstruction(
				&instructions.AddCommand{
					SourcesAndDest: instructions.SourcesAndDest{
						DestPath:    "${TEST_VAR}/dest",
						SourcePaths: []string{"src/"},
					},
					Chown: "${TEST_VAR}:group",
				},
				dockerfile.DockerfileStageInstructionOptions{
					ExpanderFactory: frontend.NewShlexExpanderFactory('\\'),
				},
			),
			nil, false,
			&stage.BaseStageOptions{ImageName: "test", ProjectName: "test"},
		)
		return expandTestEntry{
			stg: stg,
			compare: func() {
				backend, source := instruction.ExportAddCommands(stg)
				Expect(backend).To(Equal(source))
			},
		}
	}()),

	Entry("WORKDIR", func() expandTestEntry {
		stg := instruction.NewWorkdir(
			dockerfile.NewDockerfileStageInstruction(
				&instructions.WorkdirCommand{Path: "${TEST_VAR}/workdir"},
				dockerfile.DockerfileStageInstructionOptions{
					ExpanderFactory: frontend.NewShlexExpanderFactory('\\'),
				},
			),
			nil, false,
			&stage.BaseStageOptions{ImageName: "test", ProjectName: "test"},
		)
		return expandTestEntry{
			stg: stg,
			compare: func() {
				backend, source := instruction.ExportWorkdirCommands(stg)
				Expect(backend).To(Equal(source))
			},
		}
	}()),

	Entry("USER", func() expandTestEntry {
		stg := instruction.NewUser(
			dockerfile.NewDockerfileStageInstruction(
				&instructions.UserCommand{User: "${TEST_VAR}"},
				dockerfile.DockerfileStageInstructionOptions{
					ExpanderFactory: frontend.NewShlexExpanderFactory('\\'),
				},
			),
			nil, false,
			&stage.BaseStageOptions{ImageName: "test", ProjectName: "test"},
		)
		return expandTestEntry{
			stg: stg,
			compare: func() {
				backend, source := instruction.ExportUserCommands(stg)
				Expect(backend).To(Equal(source))
			},
		}
	}()),

	Entry("STOPSIGNAL", func() expandTestEntry {
		stg := instruction.NewStopSignal(
			dockerfile.NewDockerfileStageInstruction(
				&instructions.StopSignalCommand{Signal: "${TEST_VAR}"},
				dockerfile.DockerfileStageInstructionOptions{
					ExpanderFactory: frontend.NewShlexExpanderFactory('\\'),
				},
			),
			nil, false,
			&stage.BaseStageOptions{ImageName: "test", ProjectName: "test"},
		)
		return expandTestEntry{
			stg: stg,
			compare: func() {
				backend, source := instruction.ExportStopSignalCommands(stg)
				Expect(backend).To(Equal(source))
			},
		}
	}()),
)

type negativeTestEntry struct {
	stg    stage.Interface
	verify func()
}

var _ = DescribeTable("backend instruction contains expanded values after env expansion",
	func(ctx SpecContext, entry negativeTestEntry) {
		conveyor := stage.NewConveyorStub(
			stage.NewGiterminismManagerStub(stage.NewLocalGitRepoStub("test"), stage.NewGiterminismInspectorStub()),
			nil, nil,
		)
		Expect(entry.stg.ExpandDependencies(ctx, conveyor, map[string]string{"TEST_VAR": "resolved"})).To(Succeed())
		entry.verify()
	},

	Entry("COPY", func() negativeTestEntry {
		stg := instruction.NewCopy(
			dockerfile.NewDockerfileStageInstruction(
				&instructions.CopyCommand{
					SourcesAndDest: instructions.SourcesAndDest{
						DestPath:    "${TEST_VAR}/dest",
						SourcePaths: []string{"src/"},
					},
					Chown: "${TEST_VAR}:group",
				},
				dockerfile.DockerfileStageInstructionOptions{
					ExpanderFactory: frontend.NewShlexExpanderFactory('\\'),
				},
			),
			nil, false,
			&stage.BaseStageOptions{ImageName: "test", ProjectName: "test"},
		)
		return negativeTestEntry{
			stg: stg,
			verify: func() {
				backend, _ := instruction.ExportCopyCommands(stg)
				Expect(backend.DestPath).To(Equal("resolved/dest"))
				Expect(backend.Chown).To(Equal("resolved:group"))
				Expect(backend.DestPath).NotTo(ContainSubstring("${"))
				Expect(backend.Chown).NotTo(ContainSubstring("${"))
			},
		}
	}()),

	Entry("ADD", func() negativeTestEntry {
		stg := instruction.NewAdd(
			dockerfile.NewDockerfileStageInstruction(
				&instructions.AddCommand{
					SourcesAndDest: instructions.SourcesAndDest{
						DestPath:    "${TEST_VAR}/dest",
						SourcePaths: []string{"src/"},
					},
					Chown: "${TEST_VAR}:group",
				},
				dockerfile.DockerfileStageInstructionOptions{
					ExpanderFactory: frontend.NewShlexExpanderFactory('\\'),
				},
			),
			nil, false,
			&stage.BaseStageOptions{ImageName: "test", ProjectName: "test"},
		)
		return negativeTestEntry{
			stg: stg,
			verify: func() {
				backend, _ := instruction.ExportAddCommands(stg)
				Expect(backend.DestPath).To(Equal("resolved/dest"))
				Expect(backend.Chown).To(Equal("resolved:group"))
				Expect(backend.DestPath).NotTo(ContainSubstring("${"))
				Expect(backend.Chown).NotTo(ContainSubstring("${"))
			},
		}
	}()),

	Entry("WORKDIR", func() negativeTestEntry {
		stg := instruction.NewWorkdir(
			dockerfile.NewDockerfileStageInstruction(
				&instructions.WorkdirCommand{Path: "${TEST_VAR}/workdir"},
				dockerfile.DockerfileStageInstructionOptions{
					ExpanderFactory: frontend.NewShlexExpanderFactory('\\'),
				},
			),
			nil, false,
			&stage.BaseStageOptions{ImageName: "test", ProjectName: "test"},
		)
		return negativeTestEntry{
			stg: stg,
			verify: func() {
				backend, _ := instruction.ExportWorkdirCommands(stg)
				Expect(backend.Path).To(Equal("resolved/workdir"))
				Expect(backend.Path).NotTo(ContainSubstring("${"))
			},
		}
	}()),

	Entry("USER", func() negativeTestEntry {
		stg := instruction.NewUser(
			dockerfile.NewDockerfileStageInstruction(
				&instructions.UserCommand{User: "${TEST_VAR}"},
				dockerfile.DockerfileStageInstructionOptions{
					ExpanderFactory: frontend.NewShlexExpanderFactory('\\'),
				},
			),
			nil, false,
			&stage.BaseStageOptions{ImageName: "test", ProjectName: "test"},
		)
		return negativeTestEntry{
			stg: stg,
			verify: func() {
				backend, _ := instruction.ExportUserCommands(stg)
				Expect(backend.User).To(Equal("resolved"))
				Expect(backend.User).NotTo(ContainSubstring("${"))
			},
		}
	}()),

	Entry("STOPSIGNAL", func() negativeTestEntry {
		stg := instruction.NewStopSignal(
			dockerfile.NewDockerfileStageInstruction(
				&instructions.StopSignalCommand{Signal: "${TEST_VAR}"},
				dockerfile.DockerfileStageInstructionOptions{
					ExpanderFactory: frontend.NewShlexExpanderFactory('\\'),
				},
			),
			nil, false,
			&stage.BaseStageOptions{ImageName: "test", ProjectName: "test"},
		)
		return negativeTestEntry{
			stg: stg,
			verify: func() {
				backend, _ := instruction.ExportStopSignalCommands(stg)
				Expect(backend.Signal).To(Equal("resolved"))
				Expect(backend.Signal).NotTo(ContainSubstring("${"))
			},
		}
	}()),
)
