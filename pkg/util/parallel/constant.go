package parallel

const (
	CtxBackgroundTaskIDKey = "background_task_id"

	// CtxTaskStartOrderKey holds the 0-based position of the task in the
	// order tasks actually started, which is also the order their output is
	// printed in.
	CtxTaskStartOrderKey = "task_start_order"
)
