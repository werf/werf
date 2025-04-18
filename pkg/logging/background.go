package logging

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
)

func BackgroundStreams(werfServiceDir string) (io.Writer, io.Writer, error) {
	outFileName := backgroundOutputFilename(werfServiceDir)
	errFileName := backgroundErrorFilename(werfServiceDir)

	fileFlag := os.O_TRUNC | os.O_WRONLY | os.O_CREATE

	outFile, err := openFile(outFileName, fileFlag)
	if err != nil {
		return nil, nil, fmt.Errorf("unable to open %q: %w", outFileName, err)
	}

	errFile, err := openFile(errFileName, fileFlag)
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

func openFile(name string, flag int) (*os.File, error) {
	return os.OpenFile(name, flag, 0o644)
}

func BackgroundWarning(werfServiceDir string) (bool, string, error) {
	errFilename := backgroundErrorFilename(werfServiceDir)

	file, err := openFile(errFilename, os.O_RDONLY)
	if errors.Is(err, fs.ErrNotExist) {
		return false, "", nil
	} else if err != nil {
		return false, "", err
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		return false, "", err
	}

	if stat.Size() == 0 {
		return false, "", nil
	}

	warning := fmt.Sprintf(`Recent running of "werf host cleanup" in background mode was ended with errors.
Please, check these errors in %s file and remove its file after.
`, errFilename)

	return true, warning, nil
}
