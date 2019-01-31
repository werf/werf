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
	Name       string
	ImageTag   string
	ImagesRepo string
}

func (d *ImageInfoGetterStub) IsNameless() bool {
	return d.Name == ""
}

func (d *ImageInfoGetterStub) GetName() string {
	return d.Name
}

func (d *ImageInfoGetterStub) GetImageName() string {
	if d.Name == "" {
		return fmt.Sprintf("%s:%s", d.ImagesRepo, d.ImageTag)
	}
	return fmt.Sprintf("%s/%s:%s", d.ImagesRepo, d.Name, d.ImageTag)
}

func (d *ImageInfoGetterStub) GetImageId() (string, error) {
	return docker_registry.ImageId(d.GetImageName())
}

type ImageInfo struct {
	Config          *config.Image
	WithoutRegistry bool
	ImagesRepo      string
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
		return fmt.Sprintf("%s:%s", d.ImagesRepo, d.Tag)
	}
	return fmt.Sprintf("%s/%s:%s", d.ImagesRepo, d.Config.Name, d.Tag)
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

func RunDeploy(projectDir, imagesRepo, tag, release, namespace string, werfConfig *config.WerfConfig, opts DeployOptions) error {
	if debug() {
		fmt.Printf("Deploy options: %#v\n", opts)
	}

	m, err := GetSafeSecretManager(projectDir, opts.SecretValues)
	if err != nil {
		return fmt.Errorf("cannot get project secret: %s", err)
	}

	localGit := &git_repo.Local{Path: projectDir, GitDir: filepath.Join(projectDir, ".git")}

	images := GetImagesInfoGetters(werfConfig.Images, imagesRepo, tag, opts.WithoutRegistry)

	serviceValues, err := GetServiceValues(werfConfig.Meta.Project, imagesRepo, namespace, tag, localGit, images, ServiceValuesOptions{})
	if err != nil {
		return fmt.Errorf("error creating service values: %s", err)
	}

	werfChart, err := PrepareWerfChart(GetTmpWerfChartPath(werfConfig.Meta.Project), werfConfig.Meta.Project, projectDir, m, opts.Values, opts.SecretValues, opts.Set, opts.SetString, serviceValues)
	if err != nil {
		return err
	}
	if !debug() {
		// Do not remove tmp chart in debug
		defer os.RemoveAll(werfChart.ChartDir)
	}

	return werfChart.Deploy(release, namespace, HelmChartOptions{CommonHelmOptions: CommonHelmOptions{KubeContext: opts.KubeContext}, Timeout: opts.Timeout})
}
