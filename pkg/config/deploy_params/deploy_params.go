package deploy_params

import (
	"bytes"
	"fmt"
	"text/template"

	"github.com/Masterminds/sprig/v3"

	"github.com/werf/werf/pkg/config"
	"github.com/werf/werf/pkg/slug"
)

func GetHelmRelease(releaseOption, environmentOption, namespace string, werfConfig *config.WerfConfig) (string, error) {
	if releaseOption != "" {
		err := slug.ValidateHelmRelease(releaseOption)
		if err != nil {
			return "", fmt.Errorf("bad Helm release specified %q: %w", releaseOption, err)
		}
		return releaseOption, nil
	}

	var releaseTemplate string
	switch {
	case werfConfig.Meta.Deploy.HelmRelease != nil:
		releaseTemplate = *werfConfig.Meta.Deploy.HelmRelease
	case environmentOption == "":
		releaseTemplate = "[[ project ]]"
	default:
		releaseTemplate = "[[ project ]]-[[ env ]]"
	}

	renderedRelease, err := renderDeployParamTemplate("release", releaseTemplate, environmentOption, map[string]string{
		"namespace": namespace,
	}, werfConfig)
	if err != nil {
		return "", fmt.Errorf("cannot render Helm release name by template %q: %w", releaseTemplate, err)
	}

	if renderedRelease == "" {
		return "", fmt.Errorf("Helm release rendered by template %q is empty: release name cannot be empty", releaseTemplate)
	}

	var helmReleaseSlug bool
	if werfConfig.Meta.Deploy.HelmReleaseSlug != nil {
		helmReleaseSlug = *werfConfig.Meta.Deploy.HelmReleaseSlug
	} else {
		helmReleaseSlug = true
	}

	if helmReleaseSlug {
		return slug.HelmRelease(renderedRelease), nil
	}

	err = slug.ValidateHelmRelease(renderedRelease)
	if err != nil {
		return "", fmt.Errorf("bad Helm release %q rendered by template %q: %w", renderedRelease, releaseTemplate, err)
	}

	return renderedRelease, nil
}

func GetKubernetesNamespace(namespaceOption, environmentOption string, werfConfig *config.WerfConfig) (string, error) {
	if namespaceOption != "" {
		err := slug.ValidateKubernetesNamespace(namespaceOption)
		if err != nil {
			return "", fmt.Errorf("bad Kubernetes namespace specified %q: %w", namespaceOption, err)
		}
		return namespaceOption, nil
	}

	var namespaceTemplate string
	switch {
	case werfConfig.Meta.Deploy.Namespace != nil:
		namespaceTemplate = *werfConfig.Meta.Deploy.Namespace
	case environmentOption == "":
		namespaceTemplate = "[[ project ]]"
	default:
		namespaceTemplate = "[[ project ]]-[[ env ]]"
	}

	renderedNamespace, err := renderDeployParamTemplate("namespace", namespaceTemplate, environmentOption, nil, werfConfig)
	if err != nil {
		return "", fmt.Errorf("cannot render Kubernetes namespace by template %q: %w", namespaceTemplate, err)
	}

	if renderedNamespace == "" {
		return "", fmt.Errorf("Kubernetes namespace rendered by template %q is empty: namespace cannot be empty", namespaceTemplate)
	}

	var namespaceSlug bool
	if werfConfig.Meta.Deploy.NamespaceSlug != nil {
		namespaceSlug = *werfConfig.Meta.Deploy.NamespaceSlug
	} else {
		namespaceSlug = true
	}

	if namespaceSlug {
		return slug.KubernetesNamespace(renderedNamespace), nil
	}

	err = slug.ValidateKubernetesNamespace(renderedNamespace)
	if err != nil {
		return "", fmt.Errorf("bad Kubernetes namespace %q rendered by template %q: %w", renderedNamespace, namespaceTemplate, err)
	}

	return renderedNamespace, nil
}

func renderDeployParamTemplate(templateName, templateText, environmentOption string, extraData map[string]string, werfConfig *config.WerfConfig) (string, error) {
	tmpl := template.New(templateName).Delims("[[", "]]")

	funcMap := sprig.TxtFuncMap()
	delete(funcMap, "env")
	delete(funcMap, "expandenv")

	funcMap["project"] = func() string {
		return werfConfig.Meta.Project
	}

	funcMap["env"] = func() (string, error) {
		return environmentOption, nil
	}

	for k, v := range extraData {
		getValue := func(value string) func() string {
			return func() string {
				return value
			}
		}

		funcMap[k] = getValue(v)
	}

	tmpl = tmpl.Funcs(funcMap)

	tmpl, err := tmpl.Parse(templateText)
	if err != nil {
		return "", fmt.Errorf("bad template: %w", err)
	}

	buf := bytes.NewBuffer(nil)
	if err := tmpl.ExecuteTemplate(buf, templateName, nil); err != nil {
		return "", err
	}

	return buf.String(), nil
}
