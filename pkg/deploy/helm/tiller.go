package helm

import (
	"errors"
	"fmt"
	"strings"
	"time"

	corev1 "k8s.io/api/core/v1"
	kubeErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/validation"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/helm/pkg/chartutil"
	"k8s.io/helm/pkg/helm"
	helm_env "k8s.io/helm/pkg/helm/environment"
	"k8s.io/helm/pkg/kube"
	"k8s.io/helm/pkg/proto/hapi/chart"
	"k8s.io/helm/pkg/proto/hapi/services"
	"k8s.io/helm/pkg/renderutil"
	"k8s.io/helm/pkg/storage"
	"k8s.io/helm/pkg/storage/driver"
	"k8s.io/helm/pkg/tiller"
	tiller_env "k8s.io/helm/pkg/tiller/environment"

	"github.com/flant/logboek"
)

var (
	tillerReleaseServer = &tiller.ReleaseServer{}
	tillerSettings      = tiller_env.New()
	helmSettings        helm_env.EnvSettings
	resourcesWaiter     *ResourcesWaiter
	releaseLogMessages  []string

	WerfTemplateEngine     = NewWerfEngine()
	WerfTemplateEngineName = "werfGoTpl"

	defaultTimeout           = int64((24 * time.Hour).Seconds())
	defaultReleaseHistoryMax = int32(256)

	DefaultReleaseStorageNamespace = "kube-system"

	ConfigMapStorage = "configmap"
	SecretStorage    = "secret"

	ErrNoSuccessfullyDeployedReleaseRevisionFound = errors.New("no DEPLOYED release revision found")
)

type InitOptions struct {
	KubeConfig                  string
	KubeContext                 string
	HelmReleaseStorageNamespace string
	HelmReleaseStorageType      string

	WithoutKube bool
}

func Init(options InitOptions) error {
	if options.WithoutKube {
		tillerSettings.EngineYard[WerfTemplateEngineName] = WerfTemplateEngine

		tillerReleaseServer = tiller.NewReleaseServer(tillerSettings, nil, false)
		tillerReleaseServer.Log = func(f string, args ...interface{}) {
			msg := fmt.Sprintf(fmt.Sprintf("Release server: %s", f), args...)
			releaseLogMessages = append(releaseLogMessages, msg)
		}

		return nil
	}

	helmSettings.KubeConfig = options.KubeConfig
	helmSettings.KubeContext = options.KubeContext
	helmSettings.TillerNamespace = options.HelmReleaseStorageNamespace

	configFlags := genericclioptions.NewConfigFlags(true)
	configFlags.Context = &helmSettings.KubeContext
	configFlags.KubeConfig = &helmSettings.KubeConfig
	configFlags.Namespace = &options.HelmReleaseStorageNamespace

	kubeClient := kube.New(configFlags)
	kubeClient.Log = func(f string, args ...interface{}) {
		msg := fmt.Sprintf(fmt.Sprintf("Kube client: %s", f), args...)
		releaseLogMessages = append(releaseLogMessages, msg)
	}

	resourcesWaiter = &ResourcesWaiter{Client: kubeClient}
	kubeClient.SetResourcesWaiter(resourcesWaiter)

	tillerSettings.KubeClient = kubeClient
	tillerSettings.EngineYard[WerfTemplateEngineName] = WerfTemplateEngine

	clientset, err := kubeClient.KubernetesClientSet()
	if err != nil {
		return err
	}

	if _, err := clientset.CoreV1().Namespaces().Get(options.HelmReleaseStorageNamespace, metav1.GetOptions{}); err != nil {
		if kubeErrors.IsNotFound(err) {
			if _, err := clientset.CoreV1().Namespaces().Create(&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: options.HelmReleaseStorageNamespace}}); err != nil {
				return fmt.Errorf("unable to create helm release storage namespace '%s': %s", options.HelmReleaseStorageNamespace, err)
			}
			logboek.LogInfoF("Created helm release storage namespace '%s'\n", options.HelmReleaseStorageNamespace)
		} else {
			return fmt.Errorf("unable to initialize helm release storage in namespace '%s': %s", options.HelmReleaseStorageNamespace, err)
		}
	}

	switch options.HelmReleaseStorageType {
	case ConfigMapStorage:
		cfgmaps := driver.NewConfigMaps(clientset.CoreV1().ConfigMaps(options.HelmReleaseStorageNamespace))
		cfgmaps.Log = func(f string, args ...interface{}) {
			msg := fmt.Sprintf(fmt.Sprintf("ConfigMaps release storage driver: %s", f), args...)
			releaseLogMessages = append(releaseLogMessages, msg)
		}
		tillerSettings.Releases = storage.Init(cfgmaps)
		tillerSettings.Releases.Log = func(f string, args ...interface{}) {
			msg := fmt.Sprintf(fmt.Sprintf("Release storage: %s", f), args...)
			releaseLogMessages = append(releaseLogMessages, msg)
		}
	case SecretStorage:
		secrets := driver.NewSecrets(clientset.CoreV1().Secrets(options.HelmReleaseStorageNamespace))
		secrets.Log = func(f string, args ...interface{}) {
			msg := fmt.Sprintf(fmt.Sprintf("Secrets release storage driver: %s", f), args...)
			releaseLogMessages = append(releaseLogMessages, msg)
		}
		tillerSettings.Releases = storage.Init(secrets)
		tillerSettings.Releases.Log = func(f string, args ...interface{}) {
			msg := fmt.Sprintf(fmt.Sprintf("Release storage: %s", f), args...)
			releaseLogMessages = append(releaseLogMessages, msg)
		}
	default:
		return fmt.Errorf("unknown helm release storage type '%s'", options.HelmReleaseStorageType)
	}

	tillerReleaseServer = tiller.NewReleaseServer(tillerSettings, clientset, false)
	tillerReleaseServer.Log = func(f string, args ...interface{}) {
		msg := fmt.Sprintf(fmt.Sprintf("Release server: %s", f), args...)
		releaseLogMessages = append(releaseLogMessages, msg)
	}

	return nil
}

type releaseContentOptions struct {
	Version int32
}

func releaseContent(releaseName string, opts releaseContentOptions) (*services.GetReleaseContentResponse, error) {
	ctx := helm.NewContext()
	req := &services.GetReleaseContentRequest{
		Name:    releaseName,
		Version: opts.Version,
	}

	return tillerReleaseServer.GetReleaseContent(ctx, req)
}

type releaseHistoryOptions struct {
	Max int32
}

func releaseHistory(releaseName string, opts releaseHistoryOptions) (*services.GetHistoryResponse, error) {
	max := opts.Max
	if opts.Max == 0 {
		max = defaultReleaseHistoryMax
	}

	ctx := helm.NewContext()
	req := &services.GetHistoryRequest{
		Name: releaseName,
		Max:  max,
	}

	return tillerReleaseServer.GetHistory(ctx, req)
}

type releaseStatusOptions struct {
	Version int32
}

func releaseStatus(releaseName string, opts releaseStatusOptions) (*services.GetReleaseStatusResponse, error) {
	releaseLogMessages = nil
	defer func() { releaseLogMessages = nil }()

	ctx := helm.NewContext()
	req := &services.GetReleaseStatusRequest{
		Name:    releaseName,
		Version: opts.Version,
	}

	res, err := tillerReleaseServer.GetReleaseStatus(ctx, req)
	if err != nil {
		for _, msg := range releaseLogMessages {
			logboek.LogInfoF("%s\n", msg)
		}
	}
	return res, err
}

type releaseDeleteOptions struct {
	Purge   bool
	Timeout int64
}

func releaseDelete(releaseName string, opts releaseDeleteOptions) error {
	releaseLogMessages = nil
	defer func() { releaseLogMessages = nil }()

	timeout := opts.Timeout
	if opts.Timeout == 0 {
		timeout = defaultTimeout
	}

	ctx := helm.NewContext()
	req := &services.UninstallReleaseRequest{
		Name:    releaseName,
		Purge:   opts.Purge,
		Timeout: timeout,
	}

	_, err := tillerReleaseServer.UninstallRelease(ctx, req)
	if err != nil {
		for _, msg := range releaseLogMessages {
			logboek.LogInfoF("%s\n", msg)
		}
		return err
	}

	return nil
}

func isReleaseNotFoundError(err error) bool {
	return strings.HasSuffix(err.Error(), "not found")
}

type ReleaseInstallOptions struct {
	releaseInstallOptions

	Debug bool
}

func ReleaseInstall(chartPath, releaseName, namespace string, values, set, setString []string, opts ReleaseInstallOptions) error {
	rawVals, err := vals(values, set, setString, []string{}, "", "", "")
	if err != nil {
		return err
	}

	if msgs := validation.IsDNS1123Subdomain(releaseName); releaseName != "" && len(msgs) > 0 {
		return fmt.Errorf("release name %s is invalid: %s", releaseName, strings.Join(msgs, ";"))
	}

	// Check chart requirements to make sure all dependencies are present in /charts
	loadedChart, err := chartutil.Load(chartPath)
	if err != nil {
		return err
	}

	if req, err := chartutil.LoadRequirements(loadedChart); err == nil {
		// If checkDependencies returns an error, we have unfulfilled dependencies.
		// As of Helm 2.4.0, this is treated as a stopping condition:
		// https://github.com/kubernetes/helm/issues/2209
		if err := renderutil.CheckDependencies(loadedChart, req); err != nil {
			return err
		}
	} else if err != chartutil.ErrRequirementsNotFound {
		return fmt.Errorf("cannot load requirements: %v", err)
	}

	_, err = releaseInstall(loadedChart, releaseName, namespace, &chart.Config{Raw: string(rawVals)}, opts.releaseInstallOptions)
	if err != nil {
		return err
	}

	return nil
}

type ReleaseUpdateOptions struct {
	releaseUpdateOptions

	Debug bool
}

func ReleaseUpdate(chartPath, releaseName string, values, set, setString []string, opts ReleaseUpdateOptions) error {
	rawVals, err := vals(values, set, setString, []string{}, "", "", "")
	if err != nil {
		return err
	}

	// Check chart requirements to make sure all dependencies are present in /charts
	loadedChart, err := chartutil.Load(chartPath)
	if err == nil {
		if req, err := chartutil.LoadRequirements(loadedChart); err == nil {
			if err := renderutil.CheckDependencies(loadedChart, req); err != nil {
				return err
			}
		} else if err != chartutil.ErrRequirementsNotFound {
			return fmt.Errorf("cannot load requirements: %v", err)
		}
	} else {
		return err
	}

	_, err = releaseUpdate(loadedChart, releaseName, &chart.Config{Raw: string(rawVals)}, opts.releaseUpdateOptions)
	if err != nil {
		return err
	}

	return nil
}

type ReleaseRollbackOptions struct {
	releaseRollbackOptions
}

func ReleaseRollback(releaseName string, revision int32, opts ReleaseRollbackOptions) error {
	if _, err := releaseRollback(releaseName, revision, opts.releaseRollbackOptions); err != nil {
		return err
	}

	return nil
}

type releaseInstallOptions struct {
	Timeout int64
	Wait    bool
	DryRun  bool
}

func releaseInstall(chart *chart.Chart, releaseName, namespace string, values *chart.Config, opts releaseInstallOptions) (*services.InstallReleaseResponse, error) {
	releaseLogMessages = nil
	defer func() { releaseLogMessages = nil }()

	timeout := opts.Timeout
	if opts.Timeout == 0 {
		timeout = defaultTimeout
	}

	err := chartutil.ProcessRequirementsEnabled(chart, values)
	if err != nil {
		return nil, err
	}

	err = chartutil.ProcessRequirementsImportValues(chart)
	if err != nil {
		return nil, err
	}

	ctx := helm.NewContext()
	req := &services.InstallReleaseRequest{
		Chart:     chart,
		Name:      releaseName,
		Namespace: namespace,
		Values:    values,
		Wait:      opts.Wait,
		DryRun:    opts.DryRun,
		Timeout:   timeout,
	}

	resp, err := tillerReleaseServer.InstallRelease(ctx, req)
	if err != nil {
		displayReleaseLogMessages()
		return nil, err
	}

	return resp, nil
}

type releaseUpdateOptions struct {
	Timeout       int64
	CleanupOnFail bool
	Wait          bool
	DryRun        bool
}

func releaseUpdate(chart *chart.Chart, releaseName string, values *chart.Config, opts releaseUpdateOptions) (*services.UpdateReleaseResponse, error) {
	releaseLogMessages = nil
	defer func() { releaseLogMessages = nil }()

	timeout := opts.Timeout
	if opts.Timeout == 0 {
		timeout = defaultTimeout
	}

	err := chartutil.ProcessRequirementsEnabled(chart, values)
	if err != nil {
		return nil, err
	}

	err = chartutil.ProcessRequirementsImportValues(chart)
	if err != nil {
		return nil, err
	}

	ctx := helm.NewContext()
	req := &services.UpdateReleaseRequest{
		Chart:         chart,
		Name:          releaseName,
		Values:        values,
		Timeout:       timeout,
		CleanupOnFail: opts.CleanupOnFail,
		Wait:          opts.Wait,
		DryRun:        opts.DryRun,
	}

	resp, err := tillerReleaseServer.UpdateRelease(ctx, req)
	if err != nil {
		displayReleaseLogMessages()
		return nil, err
	}

	return resp, nil
}

type releaseRollbackOptions struct {
	Timeout       int64
	CleanupOnFail bool
	Wait          bool
	DryRun        bool
}

func releaseRollback(releaseName string, revision int32, opts releaseRollbackOptions) (*services.RollbackReleaseResponse, error) {
	releaseLogMessages = nil
	defer func() { releaseLogMessages = nil }()

	timeout := opts.Timeout
	if opts.Timeout == 0 {
		timeout = defaultTimeout
	}

	ctx := helm.NewContext()
	req := &services.RollbackReleaseRequest{
		Name:          releaseName,
		Version:       revision,
		Timeout:       timeout,
		CleanupOnFail: opts.CleanupOnFail,
		Wait:          opts.Wait,
		DryRun:        opts.DryRun,
	}

	resp, err := tillerReleaseServer.RollbackRelease(ctx, req)
	if err != nil {
		displayReleaseLogMessages()
		return nil, err
	}

	return resp, nil
}

func displayReleaseLogMessages() {
	logboek.LogBlock("Debug info", logboek.LogBlockOptions{}, func() {
		for _, msg := range releaseLogMessages {
			_ = logboek.WithFittedStreamsOutputOn(func() error {
				_, _ = logboek.OutF("%s\n", logboek.ColorizeInfo(msg))
				return nil
			})
		}
	})
}
