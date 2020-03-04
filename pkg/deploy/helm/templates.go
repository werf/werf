package helm

import (
	"bytes"
	"fmt"
	"path"
	"strings"
	"text/template"

	"gopkg.in/yaml.v2"

	"github.com/Masterminds/sprig"

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

	for _, t := range templates {
		if strings.ToLower(t.Kind) == strings.ToLower(kind) {
			resultTemplates = append(resultTemplates, t)
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

func GetTemplatesFromChart(chartPath, releaseName, namespace string, values []string, secretValues []map[string]interface{}, set, setString []string) (ChartTemplates, error) {
	rawTemplates, err := getRawTemplatesFromChart(chartPath, releaseName, namespace, values, secretValues, set, setString)
	if err != nil {
		return nil, err
	}

	chartTemplates, err := parseTemplates(rawTemplates)
	if err != nil {
		return nil, fmt.Errorf("unable to parse chart templates: %s", err)
	}

	return chartTemplates, nil
}

func getRawTemplatesFromChart(chartPath, releaseName, namespace string, values []string, secretValues []map[string]interface{}, set, setString []string) (string, error) {
	out := &bytes.Buffer{}

	renderOptions := RenderOptions{
		ShowNotes: false,
	}

	if err := Render(out, chartPath, releaseName, namespace, values, secretValues, set, setString, renderOptions); err != nil {
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
	err := yaml.UnmarshalStrict([]byte(rawTemplate), &t)
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
			err := yaml.UnmarshalStrict([]byte(manifestContent), &t)
			if err != nil {
				return nil, fmt.Errorf("parsing file %s failed: %s\n\n%s\n", fileName, err, util.NumerateLines(manifestContent, 1))
			}

			var h map[string]interface{}
			_ = yaml.UnmarshalStrict([]byte(manifestContent), &h)

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
	defaultEngine := engine.New()
	for _, name := range []string{"env", "expandenv"} {
		defaultEngine.FuncMap[name] = sprig.TxtFuncMap()[name]
	}

	return &WerfEngine{
		Engine:           defaultEngine,
		ExtraAnnotations: map[string]string{},
		ExtraLabels:      map[string]string{},
	}
}

func (e *WerfEngine) InitWerfEngineExtraTemplatesFunctions(decodedSecretFiles map[string]string) {
	e.AlterFuncMapHookFunc = func(t *template.Template, funcMap template.FuncMap) template.FuncMap {
		if _, err := t.Funcs(funcMap).Parse(werfEngineHelpers); err != nil {
			panic(fmt.Errorf("parse werf engine helpers failed: %s", err))
		}

		werfSecretFileFunc := func(secretRelativePath string) (string, error) {
			if path.IsAbs(secretRelativePath) {
				return "", fmt.Errorf("expected relative secret file path, given path %v", secretRelativePath)
			}

			decodedData, ok := decodedSecretFiles[secretRelativePath]

			if !ok {
				var secretFiles []string
				for key := range decodedSecretFiles {
					secretFiles = append(secretFiles, key)
				}

				return "", fmt.Errorf("secret file '%s' not found, you should use one of the following: '%s'", secretRelativePath, strings.Join(secretFiles, "', '"))
			}

			return decodedData, nil
		}

		funcMap["werf_secret_file"] = werfSecretFileFunc

		helmIncludeFunc := funcMap["include"].(func(name string, data interface{}) (string, error))
		werfIncludeFunc := func(name string, data interface{}) (string, error) {
			if name == "werf_secret_file" {
				var arg interface{}

				switch v := data.(type) {
				case []interface{}:
					if len(v) == 1 || len(v) == 2 {
						arg = v[0]
					} else {
						return "", fmt.Errorf("expected relative secret file path, given %v", v)
					}
				case interface{}:
					arg = v
				}

				argTyped, ok := arg.(string)
				if !ok {
					return "", fmt.Errorf("expected relative secret file path, given %v", arg)
				}

				if strings.HasPrefix(argTyped, "/") {
					legacyArgTyped := strings.TrimPrefix(argTyped, "/")
					if res, err := werfSecretFileFunc(legacyArgTyped); err == nil {
						return res, nil
					}
				}

				return werfSecretFileFunc(argTyped)
			}

			return helmIncludeFunc(name, data)
		}

		funcMap["include"] = werfIncludeFunc

		for _, name := range []string{
			"image",
			"image_id",
			"werf_container_image",
			"werf_container_env",
		} {
			boundedName := name
			funcMap[name] = func(data interface{}) (string, error) {
				return werfIncludeFunc(boundedName, data)
			}
		}

		return funcMap
	}
}

var werfEngineHelpers = `{{- define "_image" -}}
{{-   $context := index . 0 -}}
{{-   if not $context.Values.global.werf.is_nameless_image -}}
{{-     required "No image specified for template" nil -}}
{{-   end -}}
{{    $context.Values.global.werf.image.docker_image }}
{{- end -}}

{{- define "_image2" -}}
{{-   $name := index . 0 -}}
{{-   $context := index . 1 -}}
{{-   if $context.Values.global.werf.is_nameless_image -}}
{{-     required (printf "No image should be specified for template, got '%s'" $name) nil -}}
{{-   end -}}
{{    index (required (printf "Unknown image '%s' specified for template" $name) (pluck $name $context.Values.global.werf.image | first)) "docker_image" }}
{{- end -}}

{{- define "image" -}}
{{-   if eq (typeOf .) "chartutil.Values" -}}
{{-     $context := . -}}
{{      tuple $context | include "_image" }}
{{-   else if (ge (len .) 2) -}}
{{-     $name := index . 0 -}}
{{-     $context := index . 1 -}}
{{      tuple $name $context | include "_image2" }}
{{-   else -}}
{{-     $context := index . 0 -}}
{{      tuple $context | include "_image" }}
{{-   end -}}
{{- end -}}

{{- define "_werf_container__imagePullPolicy" -}}
{{-   $context := index . 0 -}}
{{-   if or $context.Values.global.werf.ci.is_branch $context.Values.global.werf.ci.is_custom -}}
imagePullPolicy: Always
{{-   end -}}
{{- end -}}

{{- define "_werf_container__image" -}}
{{-   $context := index . 0 -}}
image: {{ tuple $context | include "_image" }}
{{- end -}}

{{- define "_werf_container__image2" -}}
{{-   $name := index . 0 -}}
{{-   $context := index . 1 -}}
image: {{ tuple $name $context | include "_image2" }}
{{- end -}}

{{- define "werf_container_image" -}}
{{-   if eq (typeOf .) "chartutil.Values" -}}
{{-     $context := . -}}
{{      tuple $context | include "_werf_container__image" }}
{{      tuple $context | include "_werf_container__imagePullPolicy" }}
{{-   else if (ge (len .) 2) -}}
{{-     $name := index . 0 -}}
{{-     $context := index . 1 -}}
{{      tuple $name $context | include "_werf_container__image2" }}
{{      tuple $context | include "_werf_container__imagePullPolicy" }}
{{-   else -}}
{{-     $context := index . 0 -}}
{{      tuple $context | include "_werf_container__image" }}
{{      tuple $context | include "_werf_container__imagePullPolicy" }}
{{-   end -}}
{{- end -}}

{{- define "_image_id" -}}
{{-   $context := index . 0 -}}
{{-   if not $context.Values.global.werf.is_nameless_image -}}
{{-     required "No image specified for template" nil -}}
{{-   end -}}
{{    $context.Values.global.werf.image.docker_image_id }}
{{- end -}}

{{- define "_image_id2" -}}
{{-   $name := index . 0 -}}
{{-   $context := index . 1 -}}
{{-   if $context.Values.global.werf.is_nameless_image -}}
{{-     required (printf "No image should be specified for template, got '%s'" $name) nil -}}
{{-   end -}}
{{    index (required (printf "Unknown image '%s' specified for template" $name) (pluck $name $context.Values.global.werf.image | first)) "docker_image_id" }}
{{- end -}}

{{- define "image_id" -}}
{{-   if eq (typeOf .) "chartutil.Values" -}}
{{-     $context := . -}}
{{      tuple $context | include "_image_id" }}
{{-   else if (ge (len .) 2) -}}
{{-     $name := index . 0 -}}
{{-     $context := index . 1 -}}
{{      tuple $name $context | include "_image_id2" }}
{{-   else -}}
{{-     $context := index . 0 -}}
{{      tuple $context | include "_image_id" }}
{{-   end -}}
{{- end -}}

{{- define "_werf_container_env" -}}
{{-   $context := index . 0 -}}
{{-   if or $context.Values.global.werf.ci.is_branch $context.Values.global.werf.ci.is_custom -}}
- name: DOCKER_IMAGE_ID
  value: {{ tuple $context | include "_image_id" }}
{{-   end -}}
{{- end -}}

{{- define "_werf_container_env2" -}}
{{-   $name := index . 0 -}}
{{-   $context := index . 1 -}}
{{-   if or $context.Values.global.werf.ci.is_branch $context.Values.global.werf.ci.is_custom -}}
- name: DOCKER_IMAGE_ID
  value: {{ tuple $name $context | include "_image_id2" }}
{{-   end -}}
{{- end -}}

{{- define "werf_container_env" -}}
{{-   if eq (typeOf .) "chartutil.Values" -}}
{{-     $context := . -}}
{{      tuple $context | include "_werf_container_env" }}
{{-   else if (ge (len .) 2) -}}
{{-     $name := index . 0 -}}
{{-     $context := index . 1 -}}
{{      tuple $name $context | include "_werf_container_env2" }}
{{-   else -}}
{{-     $context := index . 0 -}}
{{      tuple $context | include "_werf_container_env" }}
{{-   end -}}
{{- end -}}
`

func WerfTemplateEngineWithExtraAnnotationsAndLabels(extraAnnotations, extraLabels map[string]string, f func() error) error {
	WerfTemplateEngine.ExtraAnnotations = extraAnnotations
	WerfTemplateEngine.ExtraLabels = extraLabels
	err := f()
	WerfTemplateEngine.ExtraAnnotations = map[string]string{}
	WerfTemplateEngine.ExtraLabels = map[string]string{}

	return err
}
