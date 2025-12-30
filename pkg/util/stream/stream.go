package stream

import (
	"context"
	"io"

	"github.com/werf/logboek"
)

func PipeProducerConsumer(ctx context.Context, producer func(ctx context.Context, w io.Writer) error, consumer func(ctx context.Context, r io.Reader) error) error {
	pr, pw := io.Pipe()
	errCh := make(chan error, 1)

	go func() {
		defer close(errCh)

		err := producer(ctx, pw)
		if err != nil {
			if closeErr := pw.CloseWithError(err); closeErr != nil {
				logboek.Context(ctx).Warn().LogF("WARNING: PipeProducerConsumer.CloseWithError failed: %v\n", closeErr)
			}
			errCh <- err
			return
		}

		if closeErr := pw.Close(); closeErr != nil {
			logboek.Context(ctx).Warn().LogF("WARNING: PipeProducerConsumer.Close failed: %v\n", closeErr)
		}
		errCh <- nil
	}()

	consumerErr := consumer(ctx, pr)
	producerErr := <-errCh

	if consumerErr != nil {
		return consumerErr
	}

	return producerErr
}
