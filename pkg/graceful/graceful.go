package graceful

import (
	"errors"
	"fmt"
	"runtime"

	"github.com/werf/werf/v2/pkg/logging"
)

type panicError struct {
	error error
	code  int
}

func (pe *panicError) Unwrap() error {
	return pe.error
}

func newPanicError(message string, code int) *panicError {
	return &panicError{
		error: errors.New(message),
		code:  code,
	}
}

var (
	panicChan = make(chan *panicError, 1)
	fatal     = logging.Fatal
)

func Panic(message string, exitCode int) {
	panic(newPanicError(message, exitCode))
}

func Shutdown() {
	r := recover()

	var pe *panicError

	if len(panicChan) > 0 {
		pe = <-panicChan
	}

	if r != nil {
		switch v := r.(type) {
		case *panicError:
			pe = v
		default:
			pe = newPanicError(fmt.Sprintf("%v", r), 1)
		}
	}

	if pe == nil {
		return
	}

	fatal(pe.error.Error(), pe.code)
}

func PanicGoroutine(message string, exitCode int) {
	panicChan <- newPanicError(message, exitCode)
	close(panicChan)
	runtime.Goexit()
}
