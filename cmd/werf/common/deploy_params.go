package common

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"text/template"
	"time"

	"helm.sh/helm/v3/pkg/cli"

	"github.com/werf/werf/pkg/deploy/werf_chart"

	"helm.sh/helm/v3/pkg/chart/loader"

	"github.com/werf/werf/pkg/git_repo"

	"github.com/Masterminds/sprig"

	"github.com/werf/werf/pkg/config"
	"github.com/werf/werf/pkg/image"
	"github.com/werf/werf/pkg/slug"
)

func GetHelmRelease(releaseOption string, environmentOption string, werfConfig *config.WerfConfig) (string, error) {
	if releaseOption != "" {
		err := slug.ValidateHelmRelease(releaseOption)
		if err != nil {
			return "", fmt.Errorf("bad Helm release specified '%s': %s", releaseOption, err)
		}
		return releaseOption, nil
	}

	var releaseTemplate string
	if werfConfig.Meta.MetaDeploy.HelmRelease != nil {
		releaseTemplate = *werfConfig.Meta.MetaDeploy.HelmRelease
	} else if environmentOption == "" {
		releaseTemplate = "[[ project ]]"
	} else {
		releaseTemplate = "[[ project ]]-[[ env ]]"
	}

	renderedRelease, err := renderDeployParamTemplate("release", releaseTemplate, environmentOption, werfConfig)
	if err != nil {
		return "", fmt.Errorf("cannot render Helm release name by template '%s': %s", releaseTemplate, err)
	}

	if renderedRelease == "" {
		return "", fmt.Errorf("Helm release rendered by template '%s' is empty: release name cannot be empty", releaseTemplate)
	}

	var helmReleaseSlug bool
	if werfConfig.Meta.MetaDeploy.HelmReleaseSlug != nil {
		helmReleaseSlug = *werfConfig.Meta.MetaDeploy.HelmReleaseSlug
	} else {
		helmReleaseSlug = true
	}

	if helmReleaseSlug {
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

	var namespaceTemplate string
	if werfConfig.Meta.MetaDeploy.Namespace != nil {
		namespaceTemplate = *werfConfig.Meta.MetaDeploy.Namespace
	} else if environmentOption == "" {
		namespaceTemplate = "[[ project ]]"
	} else {
		namespaceTemplate = "[[ project ]]-[[ env ]]"
	}

	renderedNamespace, err := renderDeployParamTemplate("namespace", namespaceTemplate, environmentOption, werfConfig)
	if err != nil {
		return "", fmt.Errorf("cannot render Kubernetes namespace by template '%s': %s", namespaceTemplate, err)
	}

	if renderedNamespace == "" {
		return "", fmt.Errorf("Kubernetes namespace rendered by template '%s' is empty: namespace cannot be empty", namespaceTemplate)
	}

	var namespaceSlug bool
	if werfConfig.Meta.MetaDeploy.NamespaceSlug != nil {
		namespaceSlug = *werfConfig.Meta.MetaDeploy.NamespaceSlug
	} else {
		namespaceSlug = true
	}

	if namespaceSlug {
		return slug.KubernetesNamespace(renderedNamespace), nil
	}

	err = slug.ValidateKubernetesNamespace(renderedNamespace)
	if err != nil {
		return "", fmt.Errorf("bad Kubernetes namespace '%s' rendered by template '%s': %s", renderedNamespace, namespaceTemplate, err)
	}

	return renderedNamespace, nil
}

func GetStatusProgressPeriod(cmdData *CmdData) time.Duration {
	return time.Second * time.Duration(*cmdData.StatusProgressPeriodSeconds)
}

func GetHooksStatusProgressPeriod(cmdData *CmdData) time.Duration {
	return time.Second * time.Duration(*cmdData.HooksStatusProgressPeriodSeconds)
}

func GetUserExtraAnnotations(cmdData *CmdData) (map[string]string, error) {
	extraAnnotationMap := map[string]string{}
	var addAnnotations []string

	if *cmdData.AddAnnotations != nil {
		addAnnotations = append(addAnnotations, *cmdData.AddAnnotations...)
	}

	for _, addAnnotation := range addAnnotations {
		parts := strings.Split(addAnnotation, "=")
		if len(parts) != 2 {
			return nil, fmt.Errorf("bad --add-annotation value %s", addAnnotation)
		}

		extraAnnotationMap[parts[0]] = parts[1]
	}

	return extraAnnotationMap, nil
}

func GetUserExtraLabels(cmdData *CmdData) (map[string]string, error) {
	extraLabelMap := map[string]string{}
	var addLabels []string

	if *cmdData.AddLabels != nil {
		addLabels = append(addLabels, *cmdData.AddLabels...)
	}

	for _, addLabel := range addLabels {
		parts := strings.Split(addLabel, "=")
		if len(parts) != 2 {
			return nil, fmt.Errorf("bad --add-label value %s", addLabel)
		}

		extraLabelMap[parts[0]] = parts[1]
	}

	return extraLabelMap, nil
}

func renderDeployParamTemplate(templateName, templateText string, environmentOption string, werfConfig *config.WerfConfig) (string, error) {
	tmpl := template.New(templateName).Delims("[[", "]]")

	funcMap := sprig.TxtFuncMap()

	funcMap["project"] = func() string {
		return werfConfig.Meta.Project
	}

	funcMap["env"] = func() (string, error) {
		return environmentOption, nil
	}

	tmpl = tmpl.Funcs(funcMap)

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

func StubImageInfoGetters(werfConfig *config.WerfConfig) (list []*image.InfoGetter) {
	var imagesNames []string
	for _, imageConfig := range werfConfig.StapelImages {
		imagesNames = append(imagesNames, imageConfig.Name)
	}
	for _, imageConfig := range werfConfig.ImagesFromDockerfile {
		imagesNames = append(imagesNames, imageConfig.Name)
	}

	for _, imageName := range imagesNames {
		list = append(list, image.NewInfoGetter(imageName, fmt.Sprintf("%s:%s", StubRepoAddress, StubTag), StubTag))
	}

	return list
}

func MakeChartDirLoadFunc(ctx context.Context, localGitRepo *git_repo.Local, projectDir string, disableDeterminism bool) func(dir string) ([]*loader.BufferedFile, error) {
	if disableDeterminism || localGitRepo == nil {
		return nil
	}

	return func(dir string) ([]*loader.BufferedFile, error) {
		return werf_chart.LoadFilesFromGit(ctx, localGitRepo, projectDir, dir)
	}
}

func MakeLocateChartFunc(ctx context.Context, localGitRepo *git_repo.Local, projectDir string, disableDeterminism bool) func(name string, settings *cli.EnvSettings) (string, error) {
	if disableDeterminism || localGitRepo == nil {
		return nil
	}

	return func(name string, settings *cli.EnvSettings) (string, error) {
		commit, err := localGitRepo.HeadCommit(ctx)
		if err != nil {
			return "", fmt.Errorf("unable to get local repo head commit: %s", err)
		}

		if exist, err := localGitRepo.IsFileExists(commit, name); err != nil {
			return "", fmt.Errorf("unable to check existance of chart path %q in the local git repo commit %s: %s", name, commit, err)
		} else if !exist {
			return "", fmt.Errorf("chart path %q not found in the local git repo commit %s", name, commit)
		} else {
			return name, nil
		}
	}
}
