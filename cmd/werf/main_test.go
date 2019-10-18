// +build integration_coverage

package main

import (
	"os"
	"testing"
)

func TestRunMain(t *testing.T) {
	main()
	discardStdOut()
}

// ignore test summary
// PASS
// coverage: 6.6% of statements in ./...
func discardStdOut() {
	w, err := os.Open(os.DevNull)
	if err != nil {
		panic(err)
	}

	os.Stdout = w
}
