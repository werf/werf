package parallel

type WorkerError struct {
	ID  int
	err error
}

func (e WorkerError) Unwrap() error {
	return e.err
}

func (e WorkerError) Error() string {
	return e.err.Error()
}

func NewWorkerError(id int, err error) *WorkerError {
	return &WorkerError{
		ID:  id,
		err: err,
	}
}
