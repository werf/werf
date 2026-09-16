package parallel

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"strings"
	"sync"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/werf/logboek"
	"github.com/werf/werf/v2/pkg/werf"
)

type lockedBuffer struct {
	mu  sync.Mutex
	buf bytes.Buffer
}

func (b *lockedBuffer) Write(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.Write(p)
}

func (b *lockedBuffer) String() string {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.String()
}

// The docker cli is created once per worker and bound to Worker.OutStream();
// this drives the real runWorkers wiring and writes through that stream the
// way the cli would, from inside two consecutive tasks on one worker.
var _ = Describe("worker-level stream relay", func() {
	It("formats relayed output with the current task's logger and drops what a finished task left behind", func() {
		Expect(werf.Init(GinkgoT().TempDir(), "")).To(Succeed())

		sink := &lockedBuffer{}
		ctx := logboek.NewContext(context.Background(), logboek.NewLogger(sink, sink))

		var cli io.Writer
		taskFunc := func(ctx context.Context, taskId int) error {
			switch taskId {
			case 0:
				return logboek.Context(ctx).LogProcess("A block").DoError(func() error {
					if _, err := cli.Write([]byte("cli line in A\n")); err != nil {
						return err
					}
					_, err := cli.Write([]byte("A partial"))
					return err
				})
			case 1:
				_, err := cli.Write([]byte("cli line in B\n"))
				return err
			default:
				return fmt.Errorf("unexpected task %d", taskId)
			}
		}

		err := runWorkers(ctx, 1, DoTasksOptions{}, taskFunc, func(workerCtx context.Context, worker *Worker, runTask func(taskId int) error) error {
			cli = worker.OutStream()
			for taskId := 0; taskId < 2; taskId++ {
				if err := runTask(taskId); err != nil {
					return err
				}
			}
			return nil
		})
		Expect(err).To(Succeed())

		got := sink.String()
		Expect(got).To(ContainSubstring("┌ A block\n│ cli line in A\n"), "cli output must carry the indentation of the task's own log block")
		Expect(got).To(HaveSuffix("\ncli line in B\n"), "cli output in the next task starts on its own line, unindented, and nothing trails it")

		aBlock, _, found := strings.Cut(got, "\ncli line in B\n")
		Expect(found).To(BeTrue())
		Expect(aBlock).To(ContainSubstring("A partial"), "an unterminated line the cli left in A is rendered inside A's own block (logboek flushes it with the block end), never in B")
	})

	DescribeTable("terminates a line the cli left unfinished at the end of a task, outside any log block",
		func(pick func(worker *Worker) io.Writer) {
			Expect(werf.Init(GinkgoT().TempDir(), "")).To(Succeed())

			sink := &lockedBuffer{}
			ctx := logboek.NewContext(context.Background(), logboek.NewLogger(sink, sink))

			var cli io.Writer
			taskFunc := func(ctx context.Context, taskId int) error {
				switch taskId {
				case 0:
					_, err := cli.Write([]byte("A partial"))
					return err
				case 1:
					_, err := cli.Write([]byte("B line\n"))
					return err
				default:
					return fmt.Errorf("unexpected task %d", taskId)
				}
			}

			err := runWorkers(ctx, 1, DoTasksOptions{}, taskFunc, func(workerCtx context.Context, worker *Worker, runTask func(taskId int) error) error {
				cli = pick(worker)
				for taskId := 0; taskId < 2; taskId++ {
					if err := runTask(taskId); err != nil {
						return err
					}
				}
				return nil
			})
			Expect(err).To(Succeed())

			Expect(sink.String()).To(Equal("A partial\n\nB line\n"), "the unfinished line is emitted with A, on its own line, and B starts fresh")
		},
		Entry("stdout", func(worker *Worker) io.Writer { return worker.OutStream() }),
		Entry("stderr", func(worker *Worker) io.Writer { return worker.ErrStream() }),
	)
})

var _ = It("task loggers keep the width a directly derived sub-logger would have had", func() {
	// Task loggers are cloned from a per-worker template that was itself
	// cloned from the caller's logger; each clone subtracts the caller's
	// indentation from the width, so without correction a task logger would
	// wrap two indents early.
	Expect(werf.Init(GinkgoT().TempDir(), "")).To(Succeed())

	sink := &lockedBuffer{}
	parent := logboek.NewLogger(sink, sink)
	parent.Streams().SetWidth(80)
	ctx := logboek.NewContext(context.Background(), parent)

	var taskWidth int
	parent.Streams().DoWithIndent(func() {
		expected := parent.Streams().ContentWidth()
		Expect(expected).To(BeNumerically("<", 80), "the indent must actually cost width for this test to mean anything")

		err := runWorkers(ctx, 1, DoTasksOptions{}, func(ctx context.Context, taskId int) error {
			taskWidth = logboek.Context(ctx).Streams().Width()
			return nil
		}, func(workerCtx context.Context, worker *Worker, runTask func(taskId int) error) error {
			return runTask(0)
		})
		Expect(err).To(Succeed())
		Expect(taskWidth).To(Equal(expected))
	})
})
