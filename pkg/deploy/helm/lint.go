package helm

import (
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"k8s.io/helm/pkg/lint/rules"
	"k8s.io/helm/pkg/lint/support"
)

type LintOptions struct {
	Strict bool
}

func Lint(out io.Writer, chartPath, namespace string, values []string, secretValues []map[string]interface{}, set, setString []string, opts LintOptions) error {
	var lowestTolerance int
	if opts.Strict {
		lowestTolerance = support.WarningSev
	} else {
		lowestTolerance = support.ErrorSev
	}

	var total int
	var failures int
	if linter, err := lintChart(chartPath, namespace, values, secretValues, set, setString); err != nil {
		fmt.Fprintln(out, "==> Skipping", chartPath)
		fmt.Fprintln(out, err)
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

func lintChart(chartPath string, namespace string, values []string, secretValues []map[string]interface{}, set, setString []string) (support.Linter, error) {
	linter := support.Linter{}

	// Using abs path to get directory context
	chartDir, _ := filepath.Abs(chartPath)

	linter.ChartDir = chartDir

	rules.Values(&linter)
	templatesRules(&linter, chartPath, namespace, values, secretValues, set, setString)

	return linter, nil
}

func templatesRules(linter *support.Linter, chartPath, namespace string, values []string, secretValues []map[string]interface{}, set, setString []string) {
	templates, err := GetTemplatesFromChart(chartPath, "RELEASE_NAME", namespace, values, secretValues, set, setString)
	linter.RunLinterRule(support.ErrorSev, chartPath, err)

	for _, template := range templates {
		metadataName := template.Metadata.Name
		kind := strings.ToLower(template.Kind)

		_, err := prepareMultitrackSpec(metadataName, kind, template.Namespace(namespace), template.Metadata.Annotations, 1)
		if err != nil {
			linter.RunLinterRule(support.WarningSev, "templates/", err)
		}

	templateAnnotationsLoop:
		for annoName := range template.Metadata.Annotations {
			if strings.HasPrefix(annoName, "werf.io/") {
				for _, supportedAnnoName := range werfAnnoList {
					if annoName == supportedAnnoName {
						continue templateAnnotationsLoop
					}
				}

				for _, supportedAnnoPrefix := range werfAnnoPrefixList {
					if strings.HasPrefix(annoName, supportedAnnoPrefix) {
						continue templateAnnotationsLoop
					}
				}

				err := fmt.Errorf("%s/%s with unknown werf annotation %s", kind, metadataName, annoName)
				linter.RunLinterRule(support.WarningSev, "templates/", err)
			}
		}
	}
}
