package helm

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/flant/helm/pkg/chartutil"
	"github.com/flant/helm/pkg/lint"
	"github.com/flant/helm/pkg/lint/support"
)

type LintOptions struct {
	Strict bool
}

func Lint(out io.Writer, chartPath, namespace string, values, set, setString []string, opts LintOptions) error {
	var lowestTolerance int
	if opts.Strict {
		lowestTolerance = support.WarningSev
	} else {
		lowestTolerance = support.ErrorSev
	}

	rvals, err := vals(values, set, setString, []string{}, "", "", "")
	if err != nil {
		return err
	}

	var total int
	var failures int
	if linter, err := lintChart(chartPath, rvals, namespace, opts.Strict); err != nil {
		fmt.Fprintln(out, "==> Skipping", chartPath)
		fmt.Fprintln(out, err)
		if err == errors.New("no chart found for linting (missing Chart.yaml)") {
			failures = failures + 1
		}
	} else {
		fmt.Fprintln(out, "==> Linting", chartPath)

		if len(linter.Messages) == 0 {
			fmt.Fprintln(out, "Lint OK")
		}

		for _, msg := range linter.Messages {
			fmt.Fprintln(out, msg)
		}

		total = total + 1
		if linter.HighestSeverity >= lowestTolerance {
			failures = failures + 1
		}
	}
	fmt.Fprintln(out)

	msg := fmt.Sprintf("%d chart(s) linted", total)
	if failures > 0 {
		return fmt.Errorf("%s, %d chart(s) failed", msg, failures)
	}

	fmt.Fprintf(out, "%s, no failures\n", msg)

	return nil
}

func lintChart(path string, vals []byte, namespace string, strict bool) (support.Linter, error) {
	var chartPath string
	linter := support.Linter{}

	if strings.HasSuffix(path, ".tgz") {
		tempDir, err := ioutil.TempDir("", "helm-lint")
		if err != nil {
			return linter, err
		}
		defer os.RemoveAll(tempDir)

		file, err := os.Open(path)
		if err != nil {
			return linter, err
		}
		defer file.Close()

		if err = chartutil.Expand(tempDir, file); err != nil {
			return linter, err
		}

		lastHyphenIndex := strings.LastIndex(filepath.Base(path), "-")
		if lastHyphenIndex <= 0 {
			return linter, fmt.Errorf("unable to parse chart archive %q, missing '-'", filepath.Base(path))
		}
		base := filepath.Base(path)[:lastHyphenIndex]
		chartPath = filepath.Join(tempDir, base)
	} else {
		chartPath = path
	}

	// Guard: Error out of this is not a chart.
	if _, err := os.Stat(filepath.Join(chartPath, "Chart.yaml")); err != nil {
		return linter, errors.New("no chart found for linting (missing Chart.yaml)")
	}

	return lint.All(chartPath, vals, namespace, strict), nil
}
