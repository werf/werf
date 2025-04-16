package exec_test

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/cenkalti/backoff/v5"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/samber/lo"
)

var _ = Describe("detach", func() {
	t := GinkgoT()

	if runtime.GOOS != "linux" && runtime.GOOS != "darwin" {
		t.Skip("Skipping test on non-Unix OS")
	}

	var progBinary string

	BeforeEach(func(ctx SpecContext) {
		cwd, err := os.Getwd()
		Expect(err).To(Succeed())

		progSrc := filepath.Join(cwd, "testdata/prog.go")
		progBinary = filepath.Join(t.TempDir(), fmt.Sprintf("prog_%s_%s", runtime.GOOS, runtime.GOARCH))

		cmd := exec.CommandContext(ctx, "go", "build", "-o", progBinary, progSrc)
		Expect(cmd.Run()).To(Succeed())

		Expect(os.Chmod(progBinary, 0o755)).To(Succeed())
	})

	It("should start new detached process from binary", func(ctx SpecContext) {
		cmd := exec.CommandContext(ctx, progBinary)
		cmd.Dir = t.TempDir()
		cmd.Env = append(cmd.Env, fmt.Sprintf("WERF_ORIGINAL_EXECUTABLE=%v", progBinary))
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		Expect(cmd.Run()).To(Succeed())

		proc, err := findProcessByCommand(ctx, progBinary)
		Expect(err).To(Succeed())
		Expect(proc.Kill()).To(Succeed())
	})
})

func findProcessByCommand(ctx context.Context, command string) (*os.Process, error) {
	line, err := findProcessLineByCommand(ctx, command)
	Expect(err).To(Succeed())

	// line example: " 166938 /tmp/ginkgo2394386441/prog_linux_amd64"
	lineTrimmed := strings.TrimLeft(line, " ")
	sepIdx := strings.Index(lineTrimmed, " ")
	Expect(sepIdx).NotTo(Equal(-1))

	pidStr := lineTrimmed[:sepIdx]
	pid, err := strconv.Atoi(pidStr)
	Expect(err).To(Succeed())

	p, err := os.FindProcess(pid)
	Expect(err).To(Succeed())

	return p, nil
}

func findProcessLineByCommand(ctx context.Context, command string) (string, error) {
	b := backoff.NewConstantBackOff(time.Millisecond * 10)

	operation := func() (string, error) {
		cmd := exec.CommandContext(ctx, "ps", "-eo", "pid,cmd")
		outBytes, err := cmd.Output()
		Expect(err).To(Succeed())

		outLines := strings.Split(string(outBytes), "\n")

		line, ok := lo.Find(outLines, func(item string) bool {
			return strings.HasSuffix(item, command)
		})

		if !ok {
			return "", errors.New("line not found")
		}

		return line, err
	}

	return backoff.Retry(context.TODO(), operation,
		backoff.WithBackOff(b), backoff.WithMaxElapsedTime(time.Second*30))
}
