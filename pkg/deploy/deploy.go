package deploy

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/flant/dapp/pkg/config"
	"github.com/flant/dapp/pkg/docker_registry"
	"github.com/flant/dapp/pkg/git_repo"
)

type DeployOptions struct {
	Values          []string
	SecretValues    []string
	Set             []string
	SetString       []string
	Timeout         time.Duration
	WithoutRegistry bool
}

type DimgInfoGetterStub struct {
	Name     string
	ImageTag string
	Repo     string
}

func (d *DimgInfoGetterStub) IsNameless() bool {
	return d.Name == ""
}

func (d *DimgInfoGetterStub) GetName() string {
	return d.Name
}

func (d *DimgInfoGetterStub) GetImageName() string {
	if d.Name == "" {
		return fmt.Sprintf("%s:%s", d.Repo, d.ImageTag)
	}
	return fmt.Sprintf("%s/%s:%s", d.Repo, d.Name, d.ImageTag)
}

func (d *DimgInfoGetterStub) GetImageId() (string, error) {
	return docker_registry.ImageId(d.GetImageName())
}

type DimgInfo struct {
	Config          *config.Dimg
	WithoutRegistry bool
	Repo            string
	Tag             string
}

func (d *DimgInfo) IsNameless() bool {
	return d.Config.Name == ""
}

func (d *DimgInfo) GetName() string {
	return d.Config.Name
}

func (d *DimgInfo) GetImageName() string {
	if d.Config.Name == "" {
		return fmt.Sprintf("%s:%s", d.Repo, d.Tag)
	}
	return fmt.Sprintf("%s/%s:%s", d.Repo, d.Config.Name, d.Tag)
}

func (d *DimgInfo) GetImageId() (string, error) {
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

func RunDeploy(projectName, projectDir, releaseName, namespace, kubeContext, repo, tag string, dappfile *config.Dappfile, opts DeployOptions) error {
	if debug() {
		fmt.Printf("Deploy options: %#v\n", opts)
		fmt.Printf("Namespace: %s\n", namespace)
	}

	m, err := getSafeSecretManager(projectDir, opts.SecretValues)
	if err != nil {
		return fmt.Errorf("cannot get project secret: %s", err)
	}

	localGit := &git_repo.Local{Path: projectDir, GitDir: filepath.Join(projectDir, ".git")}

	var images []DimgInfoGetter
	for _, dimg := range dappfile.Dimgs {
		d := &DimgInfo{Config: dimg, WithoutRegistry: opts.WithoutRegistry, Repo: repo, Tag: tag}
		images = append(images, d)
	}

	serviceValues, err := GetServiceValues(projectName, repo, namespace, tag, localGit, images, ServiceValuesOptions{})
	if err != nil {
		return fmt.Errorf("error creating service values: %s", err)
	}

	dappChart, err := getDappChart(projectDir, m, opts.Values, opts.SecretValues, opts.Set, opts.SetString, serviceValues)
	if err != nil {
		return err
	}
	if !debug() {
		// Do not remove tmp chart in debug
		defer os.RemoveAll(dappChart.ChartDir)
	}

	return dappChart.Deploy(releaseName, namespace, HelmChartOptions{CommonHelmOptions: CommonHelmOptions{KubeContext: kubeContext}, Timeout: opts.Timeout})
}
