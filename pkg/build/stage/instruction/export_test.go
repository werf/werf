package instruction

import (
	"github.com/moby/buildkit/frontend/dockerfile/instructions"
)

func ExportCopyCommands(stg *Copy) (backend, source instructions.CopyCommand) {
	return stg.backendInstruction.CopyCommand, *stg.instruction.Data
}

func ExportAddCommands(stg *Add) (backend, source instructions.AddCommand) {
	return stg.backendInstruction.AddCommand, *stg.instruction.Data
}

func ExportWorkdirCommands(stg *Workdir) (backend, source instructions.WorkdirCommand) {
	return stg.backendInstruction.WorkdirCommand, *stg.instruction.Data
}

func ExportUserCommands(stg *User) (backend, source instructions.UserCommand) {
	return stg.backendInstruction.UserCommand, *stg.instruction.Data
}

func ExportStopSignalCommands(stg *StopSignal) (backend, source instructions.StopSignalCommand) {
	return stg.backendInstruction.StopSignalCommand, *stg.instruction.Data
}

func ExportRunMounts(stg *Run) []*instructions.Mount {
	return stg.backendInstruction.GetMounts()
}
