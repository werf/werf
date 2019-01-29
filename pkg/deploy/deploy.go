package deploy

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/flant/werf/pkg/config"
	"github.com/flant/werf/pkg/docker_registry"
	"github.com/flant/werf/pkg/git_repo"
)

type DeployOptions struct {
	Values          []string
	SecretValues    []string
	Set             []string
	SetString       []string
	Timeout         time.Duration
	WithoutRegistry bool

	Release     string
	Namespace   string
	Environment string
	KubeContext string
}

type ImageInfoGetterStub struct {
	Name     string
	ImageTag string
	Repo     string
}

func (d *ImageInfoGetterStub) IsNameless() bool {
	return d.Name == ""
}

func (d *ImageInfoGetterStub) GetName() string {
	return d.Name
}

func (d *ImageInfoGetterStub) GetImageName() string {
	if d.Name == "" {
		return fmt.Sprintf("%s:%s", d.Repo, d.ImageTag)
	}
	return fmt.Sprintf("%s/%s:%s", d.Repo, d.Name, d.ImageTag)
}

func (d *ImageInfoGetterStub) GetImageId() (string, error) {
	return docker_registry.ImageId(d.GetImageName())
}

type ImageInfo struct {
	Config          *config.Image
	WithoutRegistry bool
	Repo            string
	Tag             string
}

func (d *ImageInfo) IsNameless() bool {
	return d.Config.Name == ""
}

func (d *ImageInfo) GetName() string {
	return d.Config.Name
}

func (d *ImageInfo) GetImageName() string {
	if d.Config.Name == "" {
		return fmt.Sprintf("%s:%s", d.Repo, d.Tag)
	}
	return fmt.Sprintf("%s/%s:%s", d.Repo, d.Config.Name, d.Tag)
}

func (d *ImageInfo) GetImageId() (string, error) {
	if d.WithoutRegistry {
		return "", nil
	}

	imageName := d.GetImageName()

	res, err := docker_registry.ImageId(imageName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR getting image %s id: %s\n", imageName, err)
		return "", nil
	}

	return res, nil
}

func RunDeploy(projectDir, repo, tag, release, namespace string, werfConfig *config.WerfConfig, opts DeployOptions) error {
	if debug() {
		fmt.Printf("Deploy options: %#v\n", opts)
	}

	fmt.Printf("Using Helm release name: %s\n", release)
	fmt.Printf("Using Kubernetes namespace: %s\n", namespace)

	m, err := getSafeSecretManager(projectDir, opts.SecretValues)
	if err != nil {
		return fmt.Errorf("cannot get project secret: %s", err)
	}

	localGit := &git_repo.Local{Path: projectDir, GitDir: filepath.Join(projectDir, ".git")}

	var images []ImageInfoGetter
	for _, image := range werfConfig.Images {
		d := &ImageInfo{Config: image, WithoutRegistry: opts.WithoutRegistry, Repo: repo, Tag: tag}
		images = append(images, d)
	}

	serviceValues, err := GetServiceValues(werfConfig.Meta.Project, repo, namespace, tag, localGit, images, ServiceValuesOptions{})
	if err != nil {
		return fmt.Errorf("error creating service values: %s", err)
	}

	werfChart, err := getWerfChart(werfConfig.Meta.Project, projectDir, m, opts.Values, opts.SecretValues, opts.Set, opts.SetString, serviceValues)
	if err != nil {
		return err
	}
	if !debug() {
		// Do not remove tmp chart in debug
		defer os.RemoveAll(werfChart.ChartDir)
	}

	return werfChart.Deploy(release, namespace, HelmChartOptions{CommonHelmOptions: CommonHelmOptions{KubeContext: opts.KubeContext}, Timeout: opts.Timeout})
}
