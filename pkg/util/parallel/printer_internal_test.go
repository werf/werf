package parallel

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/werf/logboek"
	"github.com/werf/werf/v2/pkg/werf"
)

var _ = It("Printer removes a task's temp file as soon as its block is printed, not at the end of the run", func() {
	Expect(werf.Init(GinkgoT().TempDir(), "")).To(Succeed())

	ctx := logboek.NewContext(context.Background(), logboek.NewLogger(io.Discard, io.Discard))
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	var w *Worker
	taskFunc := func(ctx context.Context, taskId int) error {
		switch taskId {
		case 0:
			logboek.Context(ctx).LogLn("a")
			return nil
		case 1:
			// B runs while A is already printed and B itself is still open:
			// A's file must be gone by the time B finishes, B's must not.
			pathOfA := w.outputs[0].path
			for {
				_, err := os.Stat(pathOfA)
				if errors.Is(err, os.ErrNotExist) {
					break
				}
				select {
				case <-ctx.Done():
					return fmt.Errorf("A's temp file %q still exists while B runs: %w", pathOfA, ctx.Err())
				case <-time.After(10 * time.Millisecond):
				}
			}
			_, err := os.Stat(w.outputs[1].path)
			Expect(err).To(Succeed(), "B's own file must still be there while B runs")
			return nil
		default:
			return fmt.Errorf("unexpected task %d", taskId)
		}
	}

	err := runWorkers(ctx, 1, DoTasksOptions{}, taskFunc, func(workerCtx context.Context, worker *Worker, runTask func(taskId int) error) error {
		w = worker
		for taskId := 0; taskId < 2; taskId++ {
			if err := runTask(taskId); err != nil {
				return err
			}
		}
		return nil
	})
	Expect(err).To(Succeed())

	for _, out := range w.outputs {
		_, err := os.Stat(out.path)
		Expect(err).To(MatchError(os.ErrNotExist), "no temp file survives the run")
	}
})
