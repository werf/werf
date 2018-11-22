package deploy

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/flant/dapp/pkg/docker_registry"
	"github.com/flant/dapp/pkg/git_repo"
	"github.com/flant/dapp/pkg/secret"
	"github.com/flant/dapp/pkg/slug"
	"github.com/flant/kubedog/pkg/kube"
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
	namespace := opts.Namespace
	if namespace == "" {
		namespace = kube.DefaultNamespace
	}
	namespace = slug.Slug(namespace)

	if debug() {
		fmt.Printf("Deploy options: %#v\n", opts)
		fmt.Printf("Namespace: %s\n", namespace)
	}

	var s secret.Secret

	isSecretsExists := false
	if _, err := os.Stat(filepath.Join(opts.ProjectDir, ProjectSecretDir)); !os.IsNotExist(err) {
		isSecretsExists = true
	}
	if _, err := os.Stat(filepath.Join(opts.ProjectDir, ProjectDefaultSecretValuesFile)); !os.IsNotExist(err) {
		isSecretsExists = true
	}
	if len(opts.SecretValues) > 0 {
		isSecretsExists = true
	}
	if isSecretsExists {
		var err error
		s, err = GetSecret(opts.ProjectDir)
		if err != nil {
			return fmt.Errorf("cannot get project secret: %s", err)
		}
	}

	dappChart, err := GenerateDappChart(opts.ProjectDir, s)
	if err != nil {
		return err
	}
	if debug() {
		// Do not remove tmp chart in debug
		fmt.Printf("Generated dapp chart: %#v\n", dappChart)
	} else {
		defer os.RemoveAll(dappChart.ChartDir)
	}

	for _, path := range opts.Values {
		err = dappChart.SetValuesFile(path)
		if err != nil {
			return err
		}
	}

	for _, path := range opts.SecretValues {
		err = dappChart.SetSecretValuesFile(path, s)
		if err != nil {
			return err
		}
	}

	for _, set := range opts.Set {
		err = dappChart.SetValuesSet(set)
		if err != nil {
			return err
		}
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

	err = dappChart.SetValues(serviceValues)
	if err != nil {
		return err
	}

	return dappChart.Deploy(releaseName, namespace, HelmChartOptions{KubeContext: opts.KubeContext, Timeout: opts.Timeout})
}

func debug() bool {
	return os.Getenv("DAPP_DEPLOY_DEBUG") == "1"
}
