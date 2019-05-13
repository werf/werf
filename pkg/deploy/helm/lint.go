package helm

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/Masterminds/semver"
	"github.com/asaskevich/govalidator"

	"k8s.io/helm/pkg/chartutil"
	"k8s.io/helm/pkg/lint/rules"
	"k8s.io/helm/pkg/lint/support"
	"k8s.io/helm/pkg/proto/hapi/chart"
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

	var total int
	var failures int
	if linter, err := lintChart(chartPath, namespace, values, set, setString, opts.Strict); err != nil {
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

func lintChart(chartPath string, namespace string, values, set, setString []string, strict bool) (support.Linter, error) {
	linter := support.Linter{}

	rvals, err := vals(values, set, setString, []string{}, "", "", "")
	if err != nil {
		return linter, err
	}

	// Guard: Error out of this is not a chart.
	if _, err := os.Stat(filepath.Join(chartPath, "Chart.yaml")); err != nil {
		return linter, errors.New("no chart found for linting (missing Chart.yaml)")
	}

	// Using abs path to get directory context
	chartDir, _ := filepath.Abs(chartPath)

	linter.ChartDir = chartDir

	lintChartfileRules(&linter)
	rules.Values(&linter)
	rules.Templates(&linter, rvals, namespace, strict)
	templatesRules(&linter, chartPath, namespace, values, set, setString)

	return linter, nil
}

func templatesRules(linter *support.Linter, chartPath, namespace string, values, set, setString []string) {
	supportedWerfAnnotations := []string{
		TrackAnnoName,
		FailModeAnnoName,
		AllowFailuresCountAnnoName,
		LogWatchRegexAnnoName,
		ShowLogsUntilAnnoName,
		SkipLogsForContainersAnnoName,
		ShowLogsOnlyForContainers,
	}

	templates, _ := GetTemplatesFromChart(chartPath, "RELEASE_NAME", namespace, values, set, setString)

	for _, template := range templates {
		metadataName := template.Metadata.Name
		kind := strings.ToLower(template.Kind)

		_, err := prepareMultitrackSpec(metadataName, kind, template.Namespace(namespace), template.Metadata.Annotations)
		if err != nil {
			linter.RunLinterRule(support.WarningSev, "templates/", err)
		}

	templateAnnotationsLoop:
		for annoName := range template.Metadata.Annotations {
			if strings.HasPrefix(annoName, "werf.io/") {
				for _, supportedAnnoName := range supportedWerfAnnotations {
					if annoName == supportedAnnoName {
						continue templateAnnotationsLoop
					}
				}

				if strings.HasPrefix(annoName, LogWatchRegexForAnnoPrefix) {
					continue templateAnnotationsLoop
				}

				err := fmt.Errorf("%s/%s with unknown werf annotation %s", kind, metadataName, annoName)
				linter.RunLinterRule(support.WarningSev, "templates/", err)
			}
		}
	}
}

func lintChartfileRules(linter *support.Linter) {
	chartFileName := "Chart.yaml"
	chartPath := filepath.Join(linter.ChartDir, chartFileName)

	linter.RunLinterRule(support.ErrorSev, chartFileName, validateChartYamlNotDirectory(chartPath))

	chartFile, err := chartutil.LoadChartfile(chartPath)
	validChartFile := linter.RunLinterRule(support.ErrorSev, chartFileName, validateChartYamlFormat(err))

	// Guard clause. Following linter rules require a parseable ChartFile
	if !validChartFile {
		return
	}

	linter.RunLinterRule(support.ErrorSev, chartFileName, validateChartNamePresence(chartFile))
	linter.RunLinterRule(support.WarningSev, chartFileName, validateChartNameFormat(chartFile))
	linter.RunLinterRule(support.ErrorSev, chartFileName, validateChartNameDirMatch(linter.ChartDir, chartFile))

	// Chart metadata
	linter.RunLinterRule(support.ErrorSev, chartFileName, validateChartVersion(chartFile))
	linter.RunLinterRule(support.ErrorSev, chartFileName, validateChartEngine(chartFile))
	linter.RunLinterRule(support.ErrorSev, chartFileName, validateChartMaintainer(chartFile))
	linter.RunLinterRule(support.ErrorSev, chartFileName, validateChartSources(chartFile))
	//linter.RunLinterRule(support.InfoSev, chartFileName, validateChartIconPresence(chartFile))
	//linter.RunLinterRule(support.ErrorSev, chartFileName, validateChartIconURL(chartFile))
}

func validateChartYamlNotDirectory(chartPath string) error {
	fi, err := os.Stat(chartPath)

	if err == nil && fi.IsDir() {
		return errors.New("should be a file, not a directory")
	}
	return nil
}

func validateChartYamlFormat(chartFileError error) error {
	if chartFileError != nil {
		return fmt.Errorf("unable to parse YAML\n\t%s", chartFileError.Error())
	}
	return nil
}

func validateChartNamePresence(cf *chart.Metadata) error {
	if cf.Name == "" {
		return errors.New("name is required")
	}
	return nil
}

func validateChartNameFormat(cf *chart.Metadata) error {
	if strings.Contains(cf.Name, ".") {
		return errors.New("name should be lower case letters and numbers. Words may be separated with dashes")
	}
	return nil
}

func validateChartNameDirMatch(chartDir string, cf *chart.Metadata) error {
	if cf.Name != filepath.Base(chartDir) {
		return fmt.Errorf("directory name (%s) and chart name (%s) must be the same", filepath.Base(chartDir), cf.Name)
	}
	return nil
}

func validateChartVersion(cf *chart.Metadata) error {
	if cf.Version == "" {
		return errors.New("version is required")
	}

	version, err := semver.NewVersion(cf.Version)

	if err != nil {
		return fmt.Errorf("version '%s' is not a valid SemVer", cf.Version)
	}

	c, err := semver.NewConstraint("> 0")
	if err != nil {
		return err
	}
	valid, msg := c.Validate(version)

	if !valid && len(msg) > 0 {
		return fmt.Errorf("version %v", msg[0])
	}

	return nil
}

func validateChartEngine(cf *chart.Metadata) error {
	if cf.Engine == "" {
		return nil
	}

	if cf.Engine == WerfTemplateEngineName {
		return nil
	}

	keys := make([]string, 0, len(chart.Metadata_Engine_value))
	for engine := range chart.Metadata_Engine_value {
		str := strings.ToLower(engine)

		if str == "unknown" {
			continue
		}

		if str == cf.Engine {
			return nil
		}

		keys = append(keys, str)
	}

	return fmt.Errorf("engine '%v' not valid. Valid options are %v", cf.Engine, keys)
}

func validateChartMaintainer(cf *chart.Metadata) error {
	for _, maintainer := range cf.Maintainers {
		if maintainer.Name == "" {
			return errors.New("each maintainer requires a name")
		} else if maintainer.Email != "" && !govalidator.IsEmail(maintainer.Email) {
			return fmt.Errorf("invalid email '%s' for maintainer '%s'", maintainer.Email, maintainer.Name)
		} else if maintainer.Url != "" && !govalidator.IsURL(maintainer.Url) {
			return fmt.Errorf("invalid url '%s' for maintainer '%s'", maintainer.Url, maintainer.Name)
		}
	}
	return nil
}

func validateChartSources(cf *chart.Metadata) error {
	for _, source := range cf.Sources {
		if source == "" || !govalidator.IsRequestURL(source) {
			return fmt.Errorf("invalid source URL '%s'", source)
		}
	}
	return nil
}
