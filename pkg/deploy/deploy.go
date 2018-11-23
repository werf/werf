package deploy

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/flant/dapp/pkg/docker_registry"
	"github.com/flant/dapp/pkg/git_repo"
)

type DeployOptions struct {
	ProjectName     string
	ProjectDir      string
	Namespace       string
	Repo            string
	Values          []string
	SecretValues    []string
	Set             []string
	Timeout         time.Duration
	KubeContext     string
	ImageTag        string
	WithoutRegistry bool

	// TODO: remove this after full port to go
	Dimgs []*DimgInfoGetterStub
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

// RunDeploy runs deploy of dapp chart
func RunDeploy(releaseName string, opts DeployOptions) error {
	namespace := getNamespace(opts.Namespace)

	if debug() {
		fmt.Printf("Deploy options: %#v\n", opts)
		fmt.Printf("Namespace: %s\n", namespace)
	}

	s, err := getOptionalSecret(opts.ProjectDir, opts.SecretValues)
	if err != nil {
		return fmt.Errorf("cannot get project secret: %s", err)
	}

	localGit := &git_repo.Local{Path: opts.ProjectDir, GitDir: filepath.Join(opts.ProjectDir, ".git")}

	images := []DimgInfoGetter{}
	for _, dimg := range opts.Dimgs {
		if debug() {
			fmt.Printf("DimgInfoGetterStub: %#v\n", dimg)
		}
		images = append(images, dimg)
	}

	serviceValues, err := GetServiceValues(opts.ProjectName, opts.Repo, namespace, opts.ImageTag, localGit, images, ServiceValuesOptions{
		WithoutRegistry: opts.WithoutRegistry,
	})
	if err != nil {
		return fmt.Errorf("error creating service values: %s", err)
	}

	dappChart, err := getDappChart(opts.ProjectDir, s, opts.Values, opts.SecretValues, opts.Set, serviceValues)
	if err != nil {
		return err
	}
	if !debug() {
		// Do not remove tmp chart in debug
		defer os.RemoveAll(dappChart.ChartDir)
	}

	return dappChart.Deploy(releaseName, namespace, HelmChartOptions{CommonHelmOptions: CommonHelmOptions{KubeContext: opts.KubeContext}, Timeout: opts.Timeout})
}
