package helm

import (
	"bytes"
	"fmt"
	"strings"

	"gopkg.in/yaml.v2"

	"k8s.io/helm/pkg/chartutil"
	"k8s.io/helm/pkg/engine"
	"k8s.io/helm/pkg/proto/hapi/chart"
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
	Version  string `yaml:"apiVersion,omitempty"`
	Kind     string `yaml:"kind,omitempty"`
	Metadata struct {
		Name        string                 `yaml:"name,omitempty"`
		Namespace   string                 `yaml:"namespace,omitempty"`
		Annotations map[string]string      `yaml:"annotations,omitempty"`
		Labels      map[string]string      `yaml:"labels,omitempty"`
		UID         string                 `yaml:"uid,omitempty"`
		OtherFields map[string]interface{} `yaml:",inline"`
	} `yaml:"metadata,omitempty"`
	Status      string                 `yaml:"status,omitempty"`
	OtherFields map[string]interface{} `yaml:",inline"`
}

func (t Template) Namespace(namespace string) string {
	if t.Metadata.Namespace != "" {
		return t.Metadata.Namespace
	}

	return namespace
}

func (t Template) IsEmpty() bool {
	if t.Version == "" || t.Kind == "" {
		return true
	}

	return false
}

func GetTemplatesFromReleaseRevision(releaseName string, revision int32) (ChartTemplates, error) {
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

func GetTemplatesFromChart(chartPath, releaseName, namespace string, values, set, setString []string) (ChartTemplates, error) {
	rawTemplates, err := getRawTemplatesFromChart(chartPath, releaseName, namespace, values, set, setString)
	if err != nil {
		return nil, err
	}

	chartTemplates, err := parseTemplates(rawTemplates)
	if err != nil {
		return nil, fmt.Errorf("unable to parse chart templates: %s", err)
	}

	return chartTemplates, nil
}

func getRawTemplatesFromChart(chartPath, releaseName, namespace string, values, set, setString []string) (string, error) {
	out := &bytes.Buffer{}

	renderOptions := RenderOptions{
		ShowNotes: false,
	}

	if err := Render(out, chartPath, releaseName, namespace, values, set, setString, renderOptions); err != nil {
		return "", err
	}

	return out.String(), nil
}

func getRawTemplatesFromRevision(releaseName string, revision int32) (string, error) {
	var result string
	resp, err := releaseContent(releaseName, releaseContentOptions{Version: revision})
	if err != nil {
		return "", err
	}

	for _, hook := range resp.Release.Hooks {
		result += fmt.Sprintf("---\n# %s\n%s\n", hook.Name, hook.Manifest)
	}

	result += "\n"
	result += resp.Release.Manifest

	return result, nil
}

func parseTemplates(rawTemplates string) (ChartTemplates, error) {
	var templates ChartTemplates

	for _, doc := range releaseutil.SplitManifests(rawTemplates) {
		t, err := parseTemplate(doc)
		if err != nil {
			return nil, err
		}

		if t.Metadata.Name != "" {
			templates = append(templates, t)
		}
	}

	return templates, nil
}

func parseTemplate(rawTemplate string) (Template, error) {
	var t Template
	err := yaml.Unmarshal([]byte(rawTemplate), &t)
	if err != nil {
		return Template{}, fmt.Errorf("%s\n\n%s\n", err, util.NumerateLines(rawTemplate, 1))
	}

	return t, nil
}

type WerfEngine struct {
	*engine.Engine

	ExtraAnnotations map[string]string
	ExtraLabels      map[string]string
}

func (e *WerfEngine) Render(chrt *chart.Chart, values chartutil.Values) (map[string]string, error) {
	templates, err := e.Engine.Render(chrt, values)
	if err != nil {
		return nil, err
	}

	for fileName, fileContent := range templates {
		if fileContent == "" {
			continue
		}

		if strings.HasSuffix(fileName, "/NOTES.txt") {
			continue
		}

		var resultManifests []string
		for _, manifestContent := range releaseutil.SplitManifests(fileContent) {
			var t Template
			err := yaml.Unmarshal([]byte(manifestContent), &t)
			if err != nil {
				return nil, fmt.Errorf("parsing file %s failed: %s\n\n%s\n", fileName, err, util.NumerateLines(manifestContent, 1))
			}

			var h map[string]interface{}
			_ = yaml.Unmarshal([]byte(manifestContent), &h)

			manifestContentIsEmpty := len(h) == 0
			if manifestContentIsEmpty {
				continue
			}

			var resultManifestContent string
			if t.IsEmpty() {
				resultManifestContent = manifestContent
			} else {
				if len(t.Metadata.Annotations) == 0 {
					t.Metadata.Annotations = map[string]string{}
				}

				for annoName, annoValue := range e.ExtraAnnotations {
					t.Metadata.Annotations[annoName] = annoValue
				}

				if len(t.Metadata.Labels) == 0 {
					t.Metadata.Labels = map[string]string{}
				}

				for labelName, labelValue := range e.ExtraLabels {
					t.Metadata.Labels[labelName] = labelValue
				}

				res, err := yaml.Marshal(t)
				if err != nil {
					return nil, err
				}

				resultManifestContent = string(res)
			}

			resultManifests = append(resultManifests, resultManifestContent)
		}

		templates[fileName] = strings.Join(resultManifests, "\n---\n")
	}

	return templates, nil
}

func NewWerfEngine() *WerfEngine {
	return &WerfEngine{
		Engine:           engine.New(),
		ExtraAnnotations: map[string]string{},
		ExtraLabels:      map[string]string{},
	}
}

func WithExtra(extraAnnotations, extraLabels map[string]string, f func() error) error {
	WerfTemplateEngine.ExtraAnnotations = extraAnnotations
	WerfTemplateEngine.ExtraLabels = extraLabels
	err := f()
	WerfTemplateEngine.ExtraAnnotations = map[string]string{}
	WerfTemplateEngine.ExtraLabels = map[string]string{}

	return err
}
