package helm

import (
	"fmt"
	"strings"

	"gopkg.in/yaml.v2"
	"k8s.io/helm/pkg/releaseutil"

	"github.com/flant/werf/pkg/util"
)

type ChartTemplates []Template

func (templates ChartTemplates) Pods() []Template {
	return templates.ByKind("Pod")
}

func (templates ChartTemplates) Jobs() []Template {
	return templates.ByKind("Job")
}

func (templates ChartTemplates) Deployments() []Template {
	return templates.ByKind("Deployment")
}

func (templates ChartTemplates) StatefulSets() []Template {
	return templates.ByKind("StatefulSet")
}

func (templates ChartTemplates) DaemonSets() []Template {
	return templates.ByKind("DaemonSet")
}

func (templates ChartTemplates) ByKind(kind string) []Template {
	var resultTemplates []Template

	for _, template := range templates {
		if strings.ToLower(template.Kind) == strings.ToLower(kind) {
			resultTemplates = append(resultTemplates, template)
		}
	}

	return resultTemplates
}

type Template struct {
	Version  string `yaml:"apiVersion"`
	Kind     string `yaml:"kind,omitempty"`
	Metadata struct {
		Name        string            `yaml:"name"`
		Namespace   string            `yaml:"namespace"`
		Annotations map[string]string `yaml:"annotations"`
		UID         string            `yaml:"uid"`
	} `yaml:"metadata,omitempty"`
	Status string `yaml:"status,omitempty"`
}

func (t Template) Namespace(namespace string) string {
	if t.Metadata.Namespace != "" {
		return t.Metadata.Namespace
	}

	return namespace
}

func GetTemplatesFromRevision(releaseName string, revision int) (ChartTemplates, error) {
	rawTemplates, err := getRawTemplatesFromRevision(releaseName, revision)
	if err != nil {
		return nil, err
	}

	chartTemplates, err := parseTemplates(rawTemplates)
	if err != nil {
		return nil, fmt.Errorf("unable to parse revision templates: %s", err)
	}

	return chartTemplates, nil
}

func GetTemplatesFromChart(chartPath, releaseName string, set, setString, values []string) (ChartTemplates, error) {
	rawTemplates, err := getRawTemplatesFromChart(chartPath, releaseName, set, setString, values)
	if err != nil {
		return nil, err
	}

	chartTemplates, err := parseTemplates(rawTemplates)
	if err != nil {
		return nil, fmt.Errorf("unable to parse chart templates: %s", err)
	}

	return chartTemplates, nil
}

func getRawTemplatesFromChart(chartPath, releaseName string, set, setString, values []string) (string, error) {
	args := []string{"template", chartPath, "--name", releaseName}
	for _, s := range set {
		args = append(args, "--set", s)
	}
	for _, s := range setString {
		args = append(args, "--set-string", s)
	}
	for _, v := range values {
		args = append(args, "--values", v)
	}

	stdout, stderr, err := HelmCmd(args...)
	if err != nil {
		return "", FormatHelmCmdError(stdout, stderr, err)
	}

	return stdout, nil
}

func getRawTemplatesFromRevision(releaseName string, revision int) (string, error) {
	hooksOutput, stderr, err := HelmCmd("get", "hooks", releaseName, "--revision", fmt.Sprintf("%d", revision))
	if err != nil {
		return "", FormatHelmCmdError(hooksOutput, stderr, err)
	}

	manifestOutput, stderr, err := HelmCmd("get", "manifest", releaseName, "--revision", fmt.Sprintf("%d", revision))
	if err != nil {
		return "", FormatHelmCmdError(manifestOutput, stderr, err)
	}

	return strings.Join([]string{hooksOutput, manifestOutput}, "\n"), nil
}

func parseTemplates(rawTemplates string) (ChartTemplates, error) {
	var templates ChartTemplates

	for _, doc := range releaseutil.SplitManifests(rawTemplates) {
		var t Template
		err := yaml.Unmarshal([]byte(doc), &t)
		if err != nil {
			return nil, fmt.Errorf("%s\n\n%s\n", err, util.NumerateLines(doc, 1))
		}

		if t.Metadata.Name != "" {
			templates = append(templates, t)
		}
	}

	return templates, nil
}
