package exec_test

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	exec2 "github.com/werf/werf/v2/pkg/werf/exec"
)

var _ = DescribeTable("cancel",
	func(
		ctx0 context.Context,
		cmdExecutionTimeout time.Duration,
		cmdCancellationWaitDelay time.Duration,
		scriptParamSleepTimeout time.Duration,
		scriptParamExitCode int,
		expectedScriptExitCode int,
	) {
		workingDir, err := os.Getwd()
		Expect(err).To(Succeed())

		script := filepath.Join(workingDir, "testdata/script.sh")

		ctx, _ := context.WithTimeout(ctx0, cmdExecutionTimeout)

		cmd := exec2.CommandContextCancellation(ctx, script, scriptParamSleepTimeout.String(), strconv.Itoa(scriptParamExitCode))
		cmd.WaitDelay = cmdCancellationWaitDelay

		err = cmd.Run()
		Expect(err).NotTo(Succeed())

		Expect(errors.Is(ctx.Err(), context.DeadlineExceeded)).To(BeTrue())

		var targetErr *exec.ExitError
		if errors.As(err, &targetErr) {
			Expect(targetErr.Success()).To(BeFalse())
			Expect(targetErr.ExitCode()).To(Equal(expectedScriptExitCode))
		} else {
			Expect(scriptParamExitCode).To(Equal(expectedScriptExitCode))
		}
	},
	Entry(
		"command should exit with 0 after cancellation by timeout=1s",
		time.Second,
		time.Minute,
		time.Duration(0),
		0,
		0,
	),
	Entry(
		"command should exit with 143 after cancellation by timeout=1s",
		time.Second,
		time.Minute,
		time.Duration(0),
		143,
		143,
	),
	Entry(
		"command should exit with -1 after cancellation by timeout=1s, waitDelay=1s and sleepTimeout=2s",
		time.Second,
		time.Second,
		time.Second*3,
		255,
		-1,
	),
)
