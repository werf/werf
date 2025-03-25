package background

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/werf/werf/v2/pkg/werf"
)

type CloseFunc func()

// Output returns writer for output in background mode. The writer must be closed using closeOutput func.
func Output() (io.Writer, CloseFunc) {
	fileName := filepath.Join(werf.GetServiceDir(), fmt.Sprintf("background_output_%d.log", os.Getpid()))

	out, err := os.OpenFile(fileName, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0o644)
	if err != nil {
		panic(fmt.Errorf("unable to open background output file %q: %w", fileName, err))
	}

	closeFn := func() {
		err := out.Close()
		if err != nil {
			panic(err)
		}
	}

	return out, closeFn
}
