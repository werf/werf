package common

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"text/template"

	"github.com/Masterminds/sprig"

	"github.com/flant/werf/pkg/config"
	"github.com/flant/werf/pkg/deploy/helm"
	"github.com/flant/werf/pkg/slug"
)

func GetHelmRelease(releaseOption string, environmentOption string, werfConfig *config.WerfConfig) (string, error) {
	if releaseOption != "" {
		err := slug.ValidateHelmRelease(releaseOption)
		if err != nil {
			return "", fmt.Errorf("bad Helm release specified '%s': %s", releaseOption, err)
		}
		return releaseOption, nil
	}

	releaseTemplate := werfConfig.Meta.DeployTemplates.HelmRelease
	if releaseTemplate == "" {
		releaseTemplate = "[[ project ]]-[[ env ]]"
	}

	renderedRelease, err := renderDeployParamTemplate("release", releaseTemplate, environmentOption, werfConfig)
	if err != nil {
		return "", fmt.Errorf("cannot render Helm release name by template '%s': %s", releaseTemplate, err)
	}

	if renderedRelease == "" {
		return "", fmt.Errorf("Helm release rendered by template '%s' is empty: release name cannot be empty", releaseTemplate)
	}

	if werfConfig.Meta.DeployTemplates.HelmReleaseSlug {
		return slug.HelmRelease(renderedRelease), nil
	}

	err = slug.ValidateHelmRelease(renderedRelease)
	if err != nil {
		return "", fmt.Errorf("bad Helm release '%s' rendered by template '%s': %s", renderedRelease, releaseTemplate, err)
	}

	return renderedRelease, nil
}

func GetKubernetesNamespace(namespaceOption string, environmentOption string, werfConfig *config.WerfConfig) (string, error) {
	if namespaceOption != "" {
		err := slug.ValidateKubernetesNamespace(namespaceOption)
		if err != nil {
			return "", fmt.Errorf("bad Kubernetes namespace specified '%s': %s", namespaceOption, err)
		}
		return namespaceOption, nil
	}

	namespaceTemplate := werfConfig.Meta.DeployTemplates.Namespace
	if namespaceTemplate == "" {
		namespaceTemplate = "[[ project ]]-[[ env ]]"
	}

	renderedNamespace, err := renderDeployParamTemplate("namespace", namespaceTemplate, environmentOption, werfConfig)
	if err != nil {
		return "", fmt.Errorf("cannot render Kubernetes namespace by template '%s': %s", namespaceTemplate, err)
	}

	if renderedNamespace == "" {
		return "", fmt.Errorf("Kubernetes namespace rendered by template '%s' is empty: namespace cannot be empty", namespaceTemplate)
	}

	if werfConfig.Meta.DeployTemplates.NamespaceSlug {
		return slug.KubernetesNamespace(renderedNamespace), nil
	}

	err = slug.ValidateKubernetesNamespace(renderedNamespace)
	if err != nil {
		return "", fmt.Errorf("bad Kubernetes namespace '%s' rendered by template '%s': %s", renderedNamespace, namespaceTemplate, err)
	}

	return renderedNamespace, nil
}

func GetHelmReleaseStorageType(helmReleaseStorageType string) (string, error) {
	switch helmReleaseStorageType {
	case helm.ConfigMapStorage, helm.SecretStorage:
		return helmReleaseStorageType, nil
	default:
		return "", fmt.Errorf("bad --helm-release-storage-type value '%s'. Use one of '%s' or '%s'", helmReleaseStorageType, helm.ConfigMapStorage, helm.SecretStorage)
	}
}

func GetUserExtraAnnotations(cmdData *CmdData) (map[string]string, error) {
	extraAnnotations := map[string]string{}
	var addAnnotations []string

	if *cmdData.AddAnnotations != nil {
		addAnnotations = append(addAnnotations, *cmdData.AddAnnotations...)
	}

	for _, keyValue := range os.Environ() {
		parts := strings.SplitN(keyValue, "=", 2)
		if strings.HasPrefix(parts[0], "WERF_ADD_ANNOTATION") {
			addAnnotations = append(addAnnotations, parts[1])
		}
	}

	for _, addAnnotation := range addAnnotations {
		parts := strings.Split(addAnnotation, "=")
		if len(parts) != 2 {
			return nil, fmt.Errorf("bad --add-annotation value %s", addAnnotation)
		}

		extraAnnotations[parts[0]] = parts[1]
	}

	return extraAnnotations, nil
}

func GetUserExtraLabels(cmdData *CmdData) (map[string]string, error) {
	extraLabels := map[string]string{}
	var addLabels []string

	if *cmdData.AddLabels != nil {
		addLabels = append(addLabels, *cmdData.AddLabels...)
	}

	for _, keyValue := range os.Environ() {
		parts := strings.SplitN(keyValue, "=", 2)
		if strings.HasPrefix(parts[0], "WERF_ADD_LABEL") {
			addLabels = append(addLabels, parts[1])
		}
	}

	for _, addLabel := range addLabels {
		parts := strings.Split(addLabel, "=")
		if len(parts) != 2 {
			return nil, fmt.Errorf("bad --add-label value %s", addLabel)
		}

		extraLabels[parts[0]] = parts[1]
	}

	return extraLabels, nil
}

func renderDeployParamTemplate(templateName, templateText string, environmentOption string, werfConfig *config.WerfConfig) (string, error) {
	tmpl := template.New(templateName).Delims("[[", "]]")

	funcMap := sprig.TxtFuncMap()

	funcMap["project"] = func() string {
		return werfConfig.Meta.Project
	}

	funcMap["env"] = func() (string, error) {
		if environmentOption == "" {
			return "", fmt.Errorf("--env option or $WERF_ENV variable required to construct name by template '%s'", templateText)
		}

		return environmentOption, nil
	}

	tmpl = tmpl.Funcs(template.FuncMap(funcMap))

	tmpl, err := tmpl.Parse(templateText)
	if err != nil {
		return "", fmt.Errorf("bad template: %s", err)
	}

	buf := bytes.NewBuffer(nil)
	if err := tmpl.ExecuteTemplate(buf, templateName, nil); err != nil {
		return "", err
	}

	return buf.String(), nil
}
