//
// Package graceful implements graceful shutdown functionality.
//
// For main goroutine it uses panic + recover mechanism.
// For child goroutines it uses the terminationErrChan + runtime.Goexit().
// Both mechanisms ensure deferred calls to be called.
//
// Example:
//
//	import graceful
//
// 	func main () {
// 		defer graceful.Shutdown() // place "defer Shutdown()" on the top of app to handle termination
//
//  	graceful.Terminate("msg", 1) // use Terminate() to terminate the app from main goroutine
//
//  	go func () {
//			graceful.TerminateGoroutine("msg", 1) // use TerminateGoroutine() to terminate the app from child goroutine
//		}()
// 	}
//

package graceful

import (
	"errors"
	"fmt"
	"runtime"

	"github.com/werf/werf/v2/pkg/logging"
)

type terminationError struct {
	error error
	code  int
}

func (pe *terminationError) Unwrap() error {
	return pe.error
}

func newTerminationError(message string, code int) *terminationError {
	return &terminationError{
		error: errors.New(message),
		code:  code,
	}
}

var (
	terminationErrChan = make(chan *terminationError, 1)
	fatal              = logging.Fatal
)

// Shutdown handles termination from main and child goroutines.
func Shutdown() {
	r := recover()

	var pe *terminationError

	// handle termination err from main goroutine
	if r != nil {
		switch v := r.(type) {
		case *terminationError:
			pe = v
		default:
			pe = newTerminationError(fmt.Sprintf("%v", r), 1)
		}
	}

	// handle termination err from child goroutine
	if len(terminationErrChan) > 0 {
		pe = <-terminationErrChan
	}

	if pe == nil {
		return
	}

	fatal(pe.error.Error(), pe.code)
}

// Terminate terminates the program from main goroutine.
func Terminate(message string, exitCode int) {
	panic(newTerminationError(message, exitCode))
}

// TerminateGoroutine terminates the program from child goroutines.
func TerminateGoroutine(message string, exitCode int) {
	terminationErrChan <- newTerminationError(message, exitCode)
	// don't close the chan to prevent "writing to the closed channel" panic in case of multiple goroutines
	// close(terminationErrChan)
	runtime.Goexit()
}
