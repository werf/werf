package instruction

import "github.com/werf/werf/pkg/build/stage"

const (
	InstructionEnv         stage.StageName = "ENV"
	InstructionCopy        stage.StageName = "COPY"
	InstructionAdd         stage.StageName = "ADD"
	InstructionRun         stage.StageName = "RUN"
	InstructionEntrypoint  stage.StageName = "ENTRYPOINT"
	InstructionCmd         stage.StageName = "CMD"
	InstructionUser        stage.StageName = "USER"
	InstructionWorkdir     stage.StageName = "WORKDIR"
	InstructionExpose      stage.StageName = "EXPOSE"
	InstructionVolume      stage.StageName = "VOLUME"
	InstructionOnBuild     stage.StageName = "ONBUILD"
	InstructionStopSignal  stage.StageName = "STOPSIGNAL"
	InstructionShell       stage.StageName = "SHELL"
	InstructionHealthcheck stage.StageName = "HEALTHCHECK"
	InstructionLabel       stage.StageName = "LABEL"
)
