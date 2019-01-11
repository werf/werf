package common

import (
	"bytes"
	"fmt"
	"os"
	"text/template"

	"github.com/Masterminds/sprig"
	"github.com/flant/dapp/pkg/config"
	"github.com/flant/dapp/pkg/slug"
)

func GetHelmRelease(releaseOption string, environmentOption string, dappfile *config.Dappfile) (string, error) {
	if releaseOption != "" {
		err := slug.ValidateHelmRelease(releaseOption)
		if err != nil {
			return "", fmt.Errorf("bad Helm release specified '%s': %s", releaseOption, err)
		}
		return releaseOption, nil
	}

	releaseTemplate := dappfile.Meta.DeployTemplates.HelmRelease
	if releaseTemplate == "" {
		releaseTemplate = "[[ project ]]-[[ environment ]]"
	}

	renderedRelease, err := renderDeployParamTemplate("release", releaseTemplate, environmentOption, dappfile)
	if err != nil {
		return "", fmt.Errorf("cannot render Helm release name by template '%s': %s", releaseTemplate, err)
	}

	if dappfile.Meta.DeployTemplates.HelmReleaseSlug {
		return slug.HelmRelease(renderedRelease), nil
	}

	err = slug.ValidateHelmRelease(renderedRelease)
	if err != nil {
		return "", fmt.Errorf("bad Helm release '%s' rendered by template '%s': %s", renderedRelease, releaseTemplate, err)
	}

	return renderedRelease, nil
}

func GetKubernetesNamespace(namespaceOption string, environmentOption string, dappfile *config.Dappfile) (string, error) {
	if namespaceOption != "" {
		err := slug.ValidateKubernetesNamespace(namespaceOption)
		if err != nil {
			return "", fmt.Errorf("bad Kubernetes namespace specified '%s': %s", namespaceOption, err)
		}
		return namespaceOption, nil
	}

	namespaceTemplate := dappfile.Meta.DeployTemplates.KubernetesNamespace
	if namespaceTemplate == "" {
		namespaceTemplate = "[[ project ]]-[[ environment ]]"
	}

	renderedNamespace, err := renderDeployParamTemplate("namespace", namespaceTemplate, environmentOption, dappfile)
	if err != nil {
		return "", fmt.Errorf("cannot render Kubernetes namespace by template '%s': %s", namespaceTemplate, err)
	}

	if dappfile.Meta.DeployTemplates.KubernetesNamespaceSlug {
		return slug.KubernetesNamespace(renderedNamespace), nil
	}

	err = slug.ValidateKubernetesNamespace(renderedNamespace)
	if err != nil {
		return "", fmt.Errorf("bad Kubernetes namespace '%s' rendered by template '%s': %s", renderedNamespace, namespaceTemplate, err)
	}

	return renderedNamespace, nil
}

func renderDeployParamTemplate(templateName, templateText string, environmentOption string, dappfile *config.Dappfile) (string, error) {
	tmpl := template.New(templateName).Delims("[[", "]]")

	funcMap := sprig.TxtFuncMap()

	funcMap["project"] = func() string {
		return dappfile.Meta.Project
	}

	funcMap["environment"] = func() (string, error) {
		environment := os.Getenv("CI_ENVIRONMENT_SLUG")

		if environment == "" {
			environment = environmentOption
		}

		if environment == "" {
			return "", fmt.Errorf("--environment option or CI_ENVIRONMENT_SLUG variable required to construct name by template '%s'", templateText)
		}

		return environment, nil
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
