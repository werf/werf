package logging

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

func backgroundOutput(werfServiceDir string) (io.Writer, io.Writer, error) {
	outFileName := backgroundOutputFilename(werfServiceDir)
	errFileName := backgroundErrorFilename(werfServiceDir)

	outFile, err := openFile(outFileName)
	if err != nil {
		return nil, nil, fmt.Errorf("unable to open %q: %w", outFileName, err)
	}

	errFile, err := openFile(errFileName)
	if err != nil {
		return nil, nil, fmt.Errorf("unable to open %q: %w", outFileName, err)
	}

	// The opened files live until the end of the program.
	// So, Go's will close them automatically using
	//	a) runtime.SetFinalizer() up to go@v1.23 or
	//	b) runtime.AddCleanup() from go@v1.24
	return outFile, errFile, nil
}

func backgroundOutputFilename(werfServiceDir string) string {
	return filepath.Join(werfServiceDir, "background_output.log")
}

func backgroundErrorFilename(werfServiceDir string) string {
	return filepath.Join(werfServiceDir, "background_error.log")
}

func openFile(name string) (*os.File, error) {
	return os.OpenFile(name, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0o644)
}
