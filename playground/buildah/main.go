package main

import (
	"context"
	"fmt"
	"os"

	"github.com/containers/storage/pkg/reexec"
	"github.com/containers/storage/pkg/unshare"
	"github.com/sirupsen/logrus"
	"github.com/werf/werf/pkg/buildah"
)

func init() {
	logrus.SetLevel(logrus.TraceLevel)

	unshare.MaybeReexecUsingUserNamespace(false)

	if err := buildah.Init(); err != nil {
		panic(err.Error())
	}
}

func main() {
	if reexec.Init() {
		return
	}

	if err := do(); err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %s\n", err)
	}
}

func do() error {
	return buildah.Run(context.Background(), "build-container", []string{"ls"}, buildah.NewRunInputOptions())
}
