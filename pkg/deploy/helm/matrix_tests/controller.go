package matrix_tests

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"

	"gopkg.in/yaml.v3"
	"k8s.io/helm/pkg/releaseutil"

	"github.com/flant/werf/pkg/config"
	"github.com/flant/werf/pkg/deploy"
	"github.com/flant/werf/pkg/deploy/helm"
	"github.com/flant/werf/pkg/deploy/werf_chart"
)

var (
	ValuesConfigFilename = "matrix_test.yaml"
)

type ModuleController struct {
	valuesDir  string
	projectDir string
	werfConfig *config.WerfConfig
	werfChart  *werf_chart.WerfChart
	opts       deploy.RenderOptions
}

func NewModuleController(valuesDir, projectDir string, werfConfig *config.WerfConfig, opts deploy.RenderOptions) (ModuleController, error) {
	m, err := deploy.GetSafeSecretManager(projectDir, opts.SecretValues, opts.IgnoreSecretKey)
	if err != nil {
		return ModuleController{}, err
	}

	images := deploy.GetImagesInfoGetters(werfConfig.StapelImages, werfConfig.ImagesFromDockerfile, opts.ImagesRepoManager, opts.Tag, opts.WithoutImagesRepo)

	serviceValues, err := deploy.GetServiceValues(werfConfig.Meta.Project, opts.ImagesRepoManager, opts.Namespace, opts.Tag, opts.TagStrategy, images, deploy.ServiceValuesOptions{Env: opts.Env})
	if err != nil {
		return ModuleController{}, err
	}

	projectChartDir := filepath.Join(projectDir, werf_chart.ProjectHelmChartDirName)
	werfChart, err := deploy.PrepareWerfChart(werfConfig.Meta.Project, projectChartDir, opts.Env, m, opts.SecretValues, serviceValues)
	if err != nil {
		return ModuleController{}, err
	}

	helm.WerfTemplateEngine.InitWerfEngineExtraTemplatesFunctions(werfChart.DecodedSecretFilesData)
	deploy.PatchLoadChartfile(werfChart.Name)

	return ModuleController{
		valuesDir:  valuesDir,
		werfChart:  werfChart,
		projectDir: projectDir,
		werfConfig: werfConfig,
		opts:       opts,
	}, nil
}

func (c *ModuleController) Run() error {
	var files []string
	err := filepath.Walk(c.valuesDir, func(path string, info os.FileInfo, err error) error {
		if c.valuesDir == path {
			return nil
		}
		files = append(files, path)
		return nil
	})
	if err != nil {
		return err
	}

	var wg sync.WaitGroup
	waitCh := make(chan struct{})
	errorCh := make(chan error, len(files))

	go func() {
		for index, file := range files {
			wg.Add(1)
			go func(index int, file string) {
				defer wg.Done()

				objectStore := NewUnstructuredObjectStore()
				fileContent, err := ioutil.ReadFile(file)
				if err != nil {
					errorCh <- fmt.Errorf("test #%v failed: %s\n", index, err)
					return
				}

				buf := bytes.NewBuffer([]byte{})
				err = c.RunRender(buf, file)
				if err != nil {
					errorCh <- testsError(index, err, string(fileContent), "")
					return
				}

				err, doc := fillObjectStore(objectStore, buf.Bytes())
				if err != nil {
					errorCh <- testsError(index, err, string(fileContent), doc)
					return
				}

				err = ApplyLintRules(objectStore)
				if err != nil {
					errorCh <- testsError(index, fmt.Errorf("lint rule failed: %v", err), string(fileContent), doc)
					return
				}
			}(index+1, file)
		}
		wg.Wait()
		close(waitCh)
	}()

	select {
	case <-waitCh:
		fmt.Printf("\n---\tSuccess: %v test cases passed!\t---\n\n", len(files))
		return nil
	case err := <-errorCh:
		return err
	}
}

func (c *ModuleController) RunRender(out io.Writer, values string) error {
	combinedValues := append(c.opts.Values, values)
	return helm.WerfTemplateEngineWithExtraAnnotationsAndLabels(c.werfChart.ExtraAnnotations, c.werfChart.ExtraLabels, func() error {
		return helm.Render(
			out,
			c.werfChart.ChartDir,
			c.opts.ReleaseName,
			c.opts.Namespace,
			append(c.werfChart.Values, combinedValues...),
			c.werfChart.SecretValues,
			append(c.werfChart.Set, c.opts.Set...),
			append(c.werfChart.SetString, c.opts.SetString...),
			helm.RenderOptions{ShowNotes: false})
	})
}

func testsError(index int, errorHeader error, generatedValues, doc string) error {
	return fmt.Errorf("test #%v failed: %s\n\n-----\n%s\n\n-----\n%s\n", index, errorHeader, generatedValues, doc)
}

func fillObjectStore(objectStore UnstructuredObjectStore, out []byte) (error, string) {
	for _, doc := range releaseutil.SplitManifests(string(out)) {
		var node map[string]interface{}
		err := yaml.Unmarshal([]byte(doc), &node)
		if err != nil {
			return err, doc
		}

		if node == nil {
			continue
		}

		err = objectStore.Put(node)
		if err != nil {
			return fmt.Errorf("helm chart object already exists: %v", err), doc
		}
	}
	return nil, ""
}
