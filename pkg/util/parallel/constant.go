package parallel

import "context"

const (
	CtxBackgroundTaskIDKey = "background_task_id"

	// CtxTaskStartOrderKey holds the 0-based position of the task in the
	// order tasks actually started, which is also the order their output is
	// printed in.
	CtxTaskStartOrderKey = "task_start_order"
)

// TaskStartOrder returns the task's start-order position from a context
// passed to a TaskFunc, and whether ctx belongs to a parallel task at all.
// Use it, not the task ID, for "N/Total" progress labels: tasks are printed
// in start order, so any other numbering comes out scattered in the log.
func TaskStartOrder(ctx context.Context) (int, bool) {
	order, ok := ctx.Value(CtxTaskStartOrderKey).(int)
	return order, ok
}
