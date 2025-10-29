package go_task

import (
	"context"
	"fmt"
	"os"
	"runtime"
	"time"

	"github.com/go-task/task/v3"
	"github.com/go-task/task/v3/args"
	"github.com/go-task/task/v3/taskfile/ast"
)

const (
	varOutputDir           = "outputDir"
	varRaceDetectorEnabled = "raceDetectorEnabled"
	varExtraGoBuildArgs    = "extraGoBuildArgs"
)

const (
	cmdBuildDev = "build"
)

type Taskfile struct {
	TaskfileName    string
	ProjectBasePath string
}

type VerbosityOptions struct {
	VerboseOutput bool
	Silent        bool
}

type Task struct {
	Taskfile *Taskfile
	TaskName string
	Vars     []string
	VerbosityOptions
}

type BuildTaskOpts struct {
	// OutputDir should be an absolute path
	OutputDir string
	// Builds binary with race detector enabled
	RaceDetectorEnabled bool
	ExtraGoBuildArgs    string
	VerbosityOptions
}

// Inits a new Taskfile struct
func NewTaskfile(taskfileName, projectBasePath string) *Taskfile {
	return &Taskfile{
		TaskfileName:    taskfileName,
		ProjectBasePath: projectBasePath,
	}
}

// Builds the werf binary in default dev mode. Equivalent to `task build`
func (tf *Taskfile) BuildDevTask(ctx context.Context, opts BuildTaskOpts) (string, error) {
	var vars []string
	outBinaryPath := tf.ProjectBasePath
	if opts.OutputDir != "" {
		vars = append(vars, buildVar(varOutputDir, opts.OutputDir))
		outBinaryPath = opts.OutputDir
	}
	if opts.RaceDetectorEnabled {
		vars = append(vars, buildVar(varRaceDetectorEnabled, "true"))
	}
	if opts.ExtraGoBuildArgs != "" {
		vars = append(vars, buildVar(varExtraGoBuildArgs, opts.ExtraGoBuildArgs))
	}
	task := &Task{
		TaskName:         cmdBuildDev,
		Vars:             vars,
		Taskfile:         tf,
		VerbosityOptions: VerbosityOptions{VerboseOutput: opts.VerboseOutput, Silent: opts.Silent},
	}
	if err := task.Execute(ctx); err != nil {
		return "", fmt.Errorf("error executing task %s: %s", task.TaskName, err)
	}
	return defaultBuildPath(outBinaryPath), nil
}

// Executes the task
func (t *Task) Execute(ctx context.Context) error {
	e := t.newExecutor()
	if err := e.Setup(); err != nil {
		return err
	}
	_, v := args.Parse(t.Vars...)
	if err := e.Run(ctx, &ast.Call{
		Task: t.TaskName,
		Vars: v,
	}); err != nil {
		return err
	}
	return nil
}

func (t *Task) newExecutor() *task.Executor {
	return &task.Executor{
		Dir:        t.Taskfile.ProjectBasePath,
		Entrypoint: fmt.Sprintf("%s/%s", t.Taskfile.ProjectBasePath, t.Taskfile.TaskfileName),
		Verbose:    t.VerboseOutput,
		Silent:     t.Silent,
		AssumeYes:  true,
		Stdin:      os.Stdin,
		Stdout:     os.Stdout,
		Stderr:     os.Stderr,
		Timeout:    time.Second * 5,
	}
}

func defaultBuildPath(basepath string) string {
	switch runtime.GOOS {
	case "windows":
		return basepath + "\\bin\\werf.exe"
	default:
		return basepath + "/bin/werf"
	}
}

func buildVar(key, value string) string {
	return fmt.Sprintf("%s=%s", key, value)
}
