// +build integration_coverage

package main

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"bou.ke/monkey"
)

var exitCode int

func TestMain(m *testing.M) {
	m.Run()
	os.Exit(exitCode)
}

func TestRunMain(t *testing.T) {
	// catch os.Exit
	fakeOsExit := func(code int) {
		exitCode = code
		panic(fmt.Sprintf("exit code %d", code))
	}
	patch := monkey.Patch(os.Exit, fakeOsExit)
	defer patch.Unpatch()

	// catch and ignore fakeOsExit panic
	defer func() {
		if r := recover(); r != nil {
			if strings.HasPrefix(fmt.Sprint(r), "exit code") {
				return
			}

			panic(r)
		}
	}()

	// ignore test options
	oldArgs := os.Args
	var newArgs []string
	for _, arg := range os.Args {
		if strings.HasPrefix(arg, "-test.") {
			continue
		}
		newArgs = append(newArgs, arg)
	}
	os.Args = newArgs
	defer func() {
		os.Args = oldArgs
	}()

	// ignore test summary
	// PASS
	// coverage: 6.6% of statements in ./...
	defer discardStdOut()

	main()
}

func discardStdOut() {
	w, err := os.Open(os.DevNull)
	if err != nil {
		panic(err)
	}

	os.Stdout = w
}
